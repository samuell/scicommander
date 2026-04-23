package main

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"
)

var (
	COLRESET                               = "\033[0m"
	COLGREEN                               = "\033[0;32m"
	COLYELLOW                              = "\033[0;33m"
	COLBRGREEN                             = "\033[1;32m"
	COLBRBLUE                              = "\033[1;34m"
	COLBRYELLOW                            = "\033[1;33m"
	COLGREY                                = "\033[1;30m"
	VERSION                                = "0.5.1"
	FILESVSTASKSFRAC_FOR_HORIZONTAL_LAYOUT = 5
)

func init() {
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Println(`Usage:
sci run <command>
sci tohtml <audit-file>
sci shell`)
	}
}

func main() {
	flag.NewFlagSet("run", flag.ExitOnError)
	flag.NewFlagSet("tohtml", flag.ExitOnError)
	flag.NewFlagSet("shell", flag.ExitOnError)
	flag.Parse()

	if len(os.Args) < 2 {
		// Run shell by default
		runShell()
	}

	switch os.Args[1] {
	case "help":
		flag.Usage()
	case "run":
		cmdStr := strings.Join(os.Args[2:], " ")
		executeCommand(cmdStr)
	case "tohtml":
		auditPath := strings.Join(os.Args[2:], " ")
		htmlPath := toHtml(auditPath)
		openHtmlFile(htmlPath)
	case "version":
		out("SciCommander %s", VERSION)
	case "shell":
		runShell()
	default:
		fmt.Println("ERROR: Expected help, run, tohtml, version or shell")
		os.Exit(2)
	}
}

func executeCommand(cmdStr string) {
	inFiles, existingOutFiles, _, _, _ := detectFiles(cmdStr)
	if len(existingOutFiles) > 0 {
		out(COLYELLOW+"[x] Skipping:"+COLRESET+" %s"+COLYELLOW+" (existing outputs)"+COLRESET, cmdStr)
		return
	}

	filesBefore := []string{}
	err := filepath.WalkDir(".", func(path string, dirEntry fs.DirEntry, err error) error {
		filesBefore = append(filesBefore, path)
		return err
	})
	checkMsg(err, "Could not walk folder structure before executing command!")

	// Execute the command
	timeBefore := time.Now()

	cmd := exec.Command("bash", "-c", cmdStr)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	out(COLBRGREEN+"[>] Executing:"+COLRESET+" %s", cmdStr)
	err = cmd.Run()

	timeAfter := time.Now()
	commandDuration := timeAfter.Sub(timeBefore)
	errMsg := f("Could not run command: %s\nSTDERR: %s\nSTDOUT: %s", cmdStr, cmd.Stderr, cmd.Stdout)
	checkMsg(err, errMsg)

	filesAfter := []string{}
	err = filepath.WalkDir(".", func(path string, dirEntry fs.DirEntry, err error) error {
		filesAfter = append(filesAfter, path)
		return err
	})
	checkMsg(err, "Could not walk folder structure after executing command!")

	// Only store files which did not exist before and is not a directory
	newPaths := []string{}
	for _, file := range filesAfter {
		if !slices.Contains(filesBefore, file) {
			newPaths = append(newPaths, file)
		}
	}

	for _, newFile := range newPaths {
		newAuditFile := newFile + ".au"

		auditInfo := NewAuditInfo(cmdStr, inFiles, newPaths)
		auditInfo.Tags.StartTime = timeBefore
		auditInfo.Tags.EndTime = timeAfter
		auditInfo.Tags.Duration = commandDuration

		auditJson, jsonErr := json.MarshalIndent(auditInfo, "", "    ")
		checkMsg(jsonErr, "Could not marshall JSON")

		writeErr := ioutil.WriteFile(newAuditFile, auditJson, 0644)
		checkMsg(writeErr, f("Could not write audit file: %v", newAuditFile))
	}
}

func detectFiles(cmdStr string) (inFiles []string, existingOutFiles []string, newOutFiles []string, existingOutDirs []string, newOutDirs []string) {
	cmdParts := strings.Split(cmdStr, " ")
	cmdArgs := cmdParts[1:]

	filtered := []string{}
	nonPaths := []string{">", "|", ">>", ">>>", "<", "<<", "<<<"}
	for _, ca := range cmdArgs {
		if !slices.Contains(nonPaths, ca) {
			filtered = append(filtered, ca)
		}
	}

	for _, cmdPart := range filtered {
		if stat, err := os.Stat(cmdPart); os.IsNotExist(err) {
			// If the file does not exist, treat as an (non-existent) output file (we don't know if it is an output file or dir)
			newOutFiles = append(newOutFiles, cmdPart)
		} else {
			// If the file does exist, check if it has an audit file
			auditPath := cmdPart + ".au"
			if _, err := os.Stat(auditPath); os.IsNotExist(err) {
				// If it lacks an audit file, treat as input file
				inFiles = append(inFiles, cmdPart)
			} else {
				// If it has an audit file, check if the command is the same
				if stat.IsDir() {
					existingOutDirs = append(existingOutDirs, cmdPart)
					// Walk the directory and check for any auditInfos with the same command as ours
					err := filepath.WalkDir(cmdPart, func(walkPath string, dirEntry fs.DirEntry, err error) error {
						walkPathStat, statErr := os.Stat(walkPath)
						checkMsg(statErr, f("Could not stat: %s", walkPath))
						if !walkPathStat.IsDir() {
							auPath := walkPath + ".au"
							if _, statErr := os.Stat(auPath); !os.IsNotExist(statErr) {
								detectedAuditInfo := unmarshalAuditInfo(auPath)
								detectedCommand := strings.Join(detectedAuditInfo.Executors[0].Command, " ")
								if detectedCommand == cmdStr {
									// If the audit info has the same command, detect as an existing outfile
									existingOutFiles = append(existingOutFiles, walkPath)
								}
							}
						} else {
							out("Skipping dir: %s", walkPath)
						}
						return err
					})
					checkMsg(err, f("Could not walk directory: %s", cmdPart))
				} else {
					detectedAuditInfo := unmarshalAuditInfo(auditPath)
					detectedCommand := strings.Join(detectedAuditInfo.Executors[0].Command, " ")
					if detectedCommand == cmdStr {
						// If the audit info has the same command, detect as an existing outfile
						existingOutFiles = append(existingOutFiles, cmdPart)
					}
					inFiles = append(inFiles, cmdPart)
				}
			}
		}
	}
	return
}

func toHtml(auditPath string) string {
	auditInfos := getAllUpstreamAuditInfos(auditPath)

	outPath := auditPath[0 : len(auditPath)-3]
	dotPath := auditPath + ".dot"
	svgPath := auditPath + ".svg"

	dotStr := generateDot(auditInfos)
	os.WriteFile(dotPath, []byte(dotStr), 0644)

	dotCmdStr := f("dot -Tsvg %s > %s", dotPath, svgPath)
	executeExternalCommand(dotCmdStr)

	html := auditInfosToHTML(auditInfos, outPath, svgPath)
	htmlPath := auditPath + ".html"
	writeErr := ioutil.WriteFile(htmlPath, []byte(html), 0644)
	checkMsg(writeErr, f("Error writing file %s", htmlPath))

	fmt.Printf("Wrote HTML file to %s\n", htmlPath)
	cwd, err := os.Getwd()
	checkMsg(err, "Could not get current working directory")
	fmt.Printf("file://%s/%s\n", cwd, htmlPath)
	fmt.Println("(You might right-click the path above to open in a browser)")

	return htmlPath
}

func openHtmlFile(htmlPath string) {
	cmd := exec.Command("bash", "-c", "open "+htmlPath)
	err := cmd.Run()
	checkMsg(err, "Could not run command: "+cmd.String())
}

func runShell() {
	exe, lookErr := exec.LookPath("bash")
	checkMsg(lookErr, "Could not find command \"bash\"")

	tempScriptPath := ".scishell.bash"

	shellCode := `#!/bin/bash -l
# Catch Ctrl+C so we don't accidentally kill SciCommander.
trap "echo \"WARNING: Ctrl+C is handled differently in SciCommander, and will not halt it. Exit with 'exit' instead.\"; echo \"Press ENTER to continue!\"" INT

# Add some short-hand functions
function ll() { ls -l; };
function lltr() { ls -ltr; };
function c() { cd $1; ls -l; echo; pwd; };
function t() { tig; };
export -f ll;
export -f lltr;
export -f c;
export -f t;

# Print logo
echo "` + COLBRGREEN + `  ___     _  ___                              _         ";
echo " / __| __(_)/ __|___ _ __  _ __  __ _ _ _  __| |___ _ _ ";
echo " \__ \/ _| | (__/ _ \ '  \| '  \/ _' | ' \/ _' / -_) '_|";
echo " |___/\__|_|\___\___/_|_|_|_|_|_\__,_|_||_\__,_\___|_|  ";
echo "` + COLBRBLUE + `>------------------------------------------------------>` + COLRESET + `"
echo;
echo " Welcome to the shell feature of SciCommander version ` + VERSION + ` (Exit with 'exit')"
echo " For more info and help see: https://github.com/samuell/scicommander"
echo;
echo " Run commands here just like normal. Most commands except some interactive"
echo " ones will then be executed via SciCommander and logged for provenance."
echo;
echo "` + COLBRBLUE + `>------------------------------------------------------>` + COLRESET + `"
echo;
echo " To force execution via SciCommander, prepend with ` + COLYELLOW + `>` + COLRESET + `:"
echo " ` + COLBRGREEN + `sci>` + COLRESET + ` ` + COLYELLOW + `>` + COLRESET + `ls -l > files.txt"
echo;
echo " To force execution outside SciCommander e.g. to show output on screen, prepend with ` + COLYELLOW + `!` + COLRESET + `:"
echo " ` + COLBRGREEN + `sci>` + COLRESET + ` ` + COLYELLOW + `!` + COLRESET + `cat files.txt"
echo;
echo " To create and open an HTML report from an audit file, run:"
echo " ` + COLBRGREEN + `sci>` + COLRESET + ` tohtml <files.txt.au>"
echo;
echo "` + COLBRBLUE + `>------------------------------------------------------>` + COLRESET + `"
echo;

# Handle fake-prompt
history -r .scishell.hist
while true; do
    dirstr="[$(basename $(pwd))]"
    read -ep "${dirstr} ` + COLBRGREEN + `sci>` + COLRESET + ` " CMD
    history -s "$CMD"
    if [[ $CMD == "" ]]; then
	    echo "(Exit SciCommander with 'exit')";
    elif [[ $CMD == "exit" ]]; then
        break;
    elif [[ $CMD == "sci" ]]; then
		echo "` + COLYELLOW + `Uh-oh! You can't run SciCommander shell inside SciCommander shell :-o` + COLRESET + `"
    elif [[ "true" == $((echo $CMD | grep -Eq "^\!.*") && echo true || echo false) ]]; then
        echo "` + COLYELLOW + `[!] Executing externally: ${CMD:1}` + COLRESET + `"
        bash -c "${CMD:1}";
    elif [[ "true" == $((echo $CMD | grep -Eq "^>.*") && echo true || echo false) ]]; then
        sci run "${CMD:1}";
    elif [[ "true" == $((echo $CMD | grep -Eq "^(ls|ll|pwd|lltr|rm|git|tig|tree|t|vim|emacs|nano|history|man)\>.*") && echo true || echo false) ]]; then
        echo "` + COLYELLOW + `Executing externally: [$CMD]` + COLRESET + `"
        bash -c "$CMD"
    elif [[ "true" == $((echo $CMD | grep -Eq ".*\<(less|more|bat)\>.*") && echo true || echo false) ]]; then
        echo "` + COLYELLOW + `Executing externally: [$CMD]` + COLRESET + `"
        bash -c "$CMD"
    elif [[ "true" == $((echo $CMD | grep -Eq "^(cd|c)\>.*") && echo true || echo false) ]]; then
        echo "` + COLYELLOW + `Executing externally: [$CMD]` + COLRESET + `"
        $CMD;
    elif [[ $CMD =~ ^(help|tohtml|run|version|shell) ]]; then
       sci $CMD;
    else
        sci run "$CMD"
    fi
    history -a .scishell.hist
done;
echo;
echo "` + COLYELLOW + `Exited SciCommander Shell` + COLRESET + `"
echo;
echo "You can start it with:"
echo "$ ` + COLBRGREEN + `sci` + COLRESET + `"
echo;`

	os.Remove(tempScriptPath)
	wrtErr := os.WriteFile(tempScriptPath, []byte(shellCode), 0644)
	checkMsg(wrtErr, f("Could not write %s", tempScriptPath))
	defer os.Remove(tempScriptPath)

	args := []string{"-i", tempScriptPath}

	env := os.Environ()
	execErr := syscall.Exec(exe, args, env)
	checkMsg(execErr, f("Could not execute command: %s %v", exe, args))
}

// getAllUpstreamAuditInfos takes a path to an audit info file and returns the
// auditinfo object it corresponds to, together with all auditinfo objects from
// upstream tasks
func getAllUpstreamAuditInfos(auditPath string) []AuditInfo {
	baseDir := filepath.Dir(auditPath)
	auditInfoMap := getInputAuditInfos(auditPath, baseDir)

	auditInfos := []AuditInfo{}
	for _, auditInfo := range auditInfoMap {
		if len(auditInfo.Executors) > 0 && !(auditInfo.Executors[0].Command[0] == "") {
			auditInfos = append(auditInfos, auditInfo)
		}
	}

	return auditInfos
}

func getInputAuditInfos(auditPath string, baseDir string) map[string]AuditInfo {
	auditInfos := map[string]AuditInfo{}
	auditInfo := unmarshalAuditInfo(auditPath)
	auditInfos[auditPath] = auditInfo

	// Recursively call this same method
	for _, inputPath := range auditInfo.Inputs {
		fullInputPath := inputPath + ".au"
		inputAuditInfos := getInputAuditInfos(fullInputPath, baseDir)
		for inputPath, inputAuditInfo := range inputAuditInfos {
			auditInfos[inputPath] = inputAuditInfo
		}
	}
	return auditInfos
}

func generateDot(auditInfos []AuditInfo) string {
	cmdNodes, fileNodes, edges := generateGraph(auditInfos)

	doHorizontalLayout := false
	if len(fileNodes) > FILESVSTASKSFRAC_FOR_HORIZONTAL_LAYOUT*len(cmdNodes) {
		doHorizontalLayout = true
	}

	dot := "DIGRAPH G {\n"
	if doHorizontalLayout {
		dot += "  rankdir=\"LR\";\n"
	}
	dot += "  node [shape=box, style=filled, fillcolor=lightgrey, fontname=monospace, penwidth=0, fontsize=11, pad=0];\n"
	for _, cmd := range cmdNodes {
		dot += f("  \"%s\" [fillcolor=\"#CCE2F1\"]\n", cmd)
	}
	for _, file := range fileNodes {
		dot += f("  \"%s\" [fillcolor=\"#FFEEC8\"]\n", file)
	}
	for _, edge := range edges {
		dot += f("  \"%s\" -> \"%s\"\n", edge.a, edge.b)
	}
	dot += "}"

	return dot
}

type StringTuple struct {
	a string
	b string
}

func (st StringTuple) String() string {
	return st.a + "," + st.b
}

func generateGraph(auditInfos []AuditInfo) (cmdNodes []string, fileNodes []string, edges []StringTuple) {
	cmdNodesSet := map[string]interface{}{}
	fileNodesSet := map[string]interface{}{}
	edgesSet := map[string]StringTuple{}

	filesCnt := 0
	for _, auditInfo := range auditInfos {
		filesCnt += len(auditInfo.Inputs) + len(auditInfo.Outputs)
	}

	foldPaths := true
	if filesCnt > FILESVSTASKSFRAC_FOR_HORIZONTAL_LAYOUT*len(auditInfos) {
		// Don't fold paths, since we will then run a horizontal layout anyway
		foldPaths = false
	}

	for _, auditInfo := range auditInfos {
		commandStr := ""
		if len(auditInfo.Executors) > 0 {
			commandStr = strings.ReplaceAll(strings.Join(auditInfo.Executors[0].Command, " "), "\"", "\\\"")
			commandStr = foldCommand(commandStr, "\\l", " ", "\\\\")
		}
		cmdNodesSet[commandStr] = nil
		if len(auditInfo.Inputs) > 0 {
			for _, input := range auditInfo.Inputs {
				if foldPaths {
					input = foldPath(input, "\\l", " ")
				}
				edge := StringTuple{string(input), commandStr}
				edgesSet[edge.String()] = edge
				fileNodesSet[string(input)] = nil
			}
		}
		cmdNodesSet[commandStr] = nil
		if len(auditInfo.Outputs) > 0 {
			for _, output := range auditInfo.Outputs {
				if foldPaths {
					output = foldPath(output, "\\l", " ")
				}
				edge := StringTuple{commandStr, string(output)}
				edgesSet[edge.String()] = edge
				fileNodesSet[string(output)] = nil
			}
		}
	}
	for node := range cmdNodesSet {
		cmdNodes = append(cmdNodes, node)
	}
	for node := range fileNodesSet {
		fileNodes = append(fileNodes, node)
	}
	for _, edge := range edgesSet {
		edges = append(edges, edge)
	}
	return
}

func unmarshalAuditInfo(auditPath string) AuditInfo {
	if _, statErr := os.Stat(auditPath); statErr == nil {
		var auditInfo AuditInfo
		auditJson, readErr := ioutil.ReadFile(auditPath)
		checkMsg(readErr, f("Error reading file %s", auditPath))
		err := json.Unmarshal([]byte(auditJson), &auditInfo)
		checkMsg(err, "Failed to unmarshal JSON file "+auditPath)
		return auditInfo
	} else if errors.Is(statErr, os.ErrNotExist) {
		return *NewAuditInfo("", nil, nil)
	} else {
		checkMsg(statErr, f("Error stat:ing file %s", auditPath))
		return *NewAuditInfo("", nil, nil)
	}
}

func auditInfosToHTML(auditInfos []AuditInfo, outPath string, svgPath string) (html string) {
	svgRaw, err := os.ReadFile(svgPath)
	svg := string(svgRaw)
	checkMsg(err, "Could not read SVG file: "+svgPath)
	html = "<html>\n"
	html += "<head>\n"
	html += "<style>\n"
	html += `
html {
	background: #efefef;
}
body{
	background: white;
	font-family:monospace, courier new;
	font-size: 9pt;
	width: 90%;
	margin: 0 auto;
	box-shadow: 2px 2px 10px #ccc;
	padding: 1em;
}
hr {
	border: 2px solid #efefef;
}
table {
	borders: none;
	width: 100%;
}
table td {
	padding: .4em 1em;
	text-align: left;
}
`
	html += "</style>\n"
	html += "</head>\n"
	html += "<body>\n"
	html += f("<h1>SciCommander Audit Report for %s<h1>\n", outPath)
	html += "<hr>\n"
	html += "<table>\n"
	html += "<tr><th>Start time</th><th>Command</th><th>Duration</th></tr>\n"

	slices.SortFunc(auditInfos, func(a, b AuditInfo) int {
		return cmp.Compare(a.Tags.StartTime.Format(time.RFC3339), b.Tags.StartTime.Format(time.RFC3339))
	})

	for _, auditInfo := range auditInfos {
		command := strings.Join(auditInfo.Executors[0].Command, " ")
		command = foldCommand(command, "<br>", "&nbsp;", "\\")
		html += f("<tr><td style=\"background: #E6F5FF;\">%s</td>"+
			"<td style=\"background: #CCE2F1;\">%s</td>"+
			"<td style=\"background: #E6F5FF;\">%d ms</tr>"+
			"\n", auditInfo.Tags.StartTime.Format(time.RFC3339), command, auditInfo.Tags.Duration.Milliseconds())
	}
	html += "</table>"
	html += "<hr>\n"
	html += svg + "\n"
	html += "<hr>\n"
	html += "</body>\n"
	html += "</html>\n"
	return html
}

func executeExternalCommand(cmd string) {
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	checkMsg(err, f("Command failed!\nCommand:\n%s\n\nOutput:\n%s", cmd, out))
}

func checkMsg(err error, message string) {
	if err != nil {
		fmt.Println(message)
		fmt.Println(err)
		os.Exit(126)
	}
}

func foldCommand(cmdStr string, newLineStr string, spaceStr string, backSlashStr string) (foldedCmd string) {
	redirections := []string{
		">",
		">>",
		"<",
		"<<",
		">|",
		"<>",
		"|",
		"|&",
		"2>",
		"2>>",
		"1>",
		"1>>",
		"&>",
		"&>>",
	}
	foldedCmd = ""
	foldNextLine := true
	for i, cp := range strings.Split(cmdStr, " ") {
		if string(cp[0]) == "-" || slices.Contains(redirections, cp) {
			foldedCmd += backSlashStr + " " + newLineStr + spaceStr + spaceStr + cp + " "
			foldNextLine = false
		} else {
			if foldNextLine && i > 0 && len(cp) > 16 {
				foldedCmd += backSlashStr + " " + newLineStr + spaceStr + spaceStr + cp + " "
			} else {
				foldedCmd += cp + " "
			}
			foldNextLine = true
		}
	}
	foldedCmd = foldedCmd + newLineStr
	return foldedCmd
}

func foldPath(path string, newLineStr string, spaceStr string) (foldedPath string) {
	foldedPath = ""
	for i, cp := range strings.Split(path, "/") {
		if i == 0 {
			foldedPath += cp
		} else {
			totSpace := ""
			for j := 1; j <= i+1; j++ {
				totSpace += spaceStr
			}
			foldedPath += "/" + newLineStr + totSpace + cp
		}
	}
	foldedPath += newLineStr
	return foldedPath
}

func f(s string, v ...interface{}) string {
	return fmt.Sprintf(s, v...)
}

func out(s string, v ...interface{}) {
	fmt.Printf(s+"\n", v...)
}
