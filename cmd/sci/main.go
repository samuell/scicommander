package main

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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
	COLRESET    = "\033[0m"
	COLGREEN    = "\033[0;32m"
	COLYELLOW   = "\033[0;33m"
	COLBRGREEN  = "\033[1;32m"
	COLBRBLUE   = "\033[1;34m"
	COLBRYELLOW = "\033[1;33m"
	COLGREY     = "\033[1;30m"
	VERSION     = "0.5.0"
)

func init() {
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Println(`Usage:
sci run <command>
sci to-html <audit-file>
sci shell`)
	}
}

func main() {
	flag.NewFlagSet("run", flag.ExitOnError)
	flag.NewFlagSet("to-html", flag.ExitOnError)
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
	case "to-html":
		auditPath := strings.Join(os.Args[2:], " ")
		toHtml(auditPath)
	case "version":
		fmt.Printf("SciCommander %s\n", VERSION)
	case "shell":
		runShell()
	default:
		fmt.Println("ERROR: Expected help, run, to-html, version or shell")
		os.Exit(2)
	}
}

func executeCommand(cmdStr string) {
	cmdParts := strings.Split(cmdStr, " ")
	cmdArgs := cmdParts[1:]

	inFiles, _ := detectFiles(cmdArgs)

	filesBefore, err := filepath.Glob("./*")
	checkMsg(err, "Could not glob folder before executing command!")

	// Execute the command
	timeBefore := time.Now()

	cmd := exec.Command("bash", "-c", cmdStr)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Printf(COLGREEN+" ->"+COLRESET+" %s\n", cmdStr)
	err = cmd.Run()

	timeAfter := time.Now()
	commandDuration := timeAfter.Sub(timeBefore)
	errMsg := f("Could not run command: %s\nSTDERR: %s\nSTDOUT: %s", cmdStr, cmd.Stderr, cmd.Stdout)
	checkMsg(err, errMsg)

	filesAfter, err := filepath.Glob("./*")
	checkMsg(err, "Could not glob folder after executing command!")

	newFiles := []string{}
	for _, file := range filesAfter {
		if !slices.Contains(filesBefore, file) {
			newFiles = append(newFiles, file)
		}
	}

	for _, newFile := range newFiles {
		newAuditFile := newFile + ".au"
		auditInfo := NewAuditInfo(cmdStr, inFiles, newFiles)
		auditInfo.Tags.StartTime = timeBefore
		auditInfo.Tags.EndTime = timeAfter
		auditInfo.Tags.Duration = commandDuration

		auditJson, jsonErr := json.MarshalIndent(auditInfo, "", "    ")
		checkMsg(jsonErr, "Could not marshall JSON")

		writeErr := ioutil.WriteFile(newAuditFile, auditJson, 0644)
		checkMsg(writeErr, f("Could not write audit file: %v", newAuditFile))
	}
}

func detectFiles(strs []string) ([]string, []string) {
	inFiles := []string{}
	outFiles := []string{}

	filtered := []string{}
	nonPaths := []string{">", "|", ">>", ">>>", "<", "<<", "<<<"}
	for _, s := range strs {
		if !slices.Contains(nonPaths, s) {
			filtered = append(filtered, s)
		}
	}

	for _, str := range filtered {
		if _, err := os.Stat(str); os.IsNotExist(err) {
			outFiles = append(outFiles, str)
		} else {
			auditPath := str + ".au"
			if _, err := os.Stat(auditPath); os.IsNotExist(err) {
				inFiles = append(inFiles, str)
			} else {
				inFiles = append(inFiles, str)
			}
		}
	}
	return inFiles, outFiles
}

func toHtml(auditPath string) {
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

	cmd := exec.Command("bash", "-c", "open "+htmlPath)
	err = cmd.Run()
	checkMsg(err, "Could not run command: "+cmd.String())
}

func runShell() {
	exe, lookErr := exec.LookPath("bash")
	checkMsg(lookErr, "Could not find command \"bash\"")

	tempScriptPath := ".scishell.bash"

	shellCode := `#!/bin/bash -l
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
echo " Welcome to the SciCommander shell!"
echo " Commands executed here will be logged for provenance."
echo " (Exit with Ctrl+C)"
echo;
echo " See also the other sci sub-commands:"
echo " ` + COLBRGREEN + `>` + COLRESET + ` sci help"
echo " ` + COLBRGREEN + `>` + COLRESET + ` sci run <command>"
echo " ` + COLBRGREEN + `>` + COLRESET + ` sci to-html <file.au>"
echo " ` + COLBRGREEN + `>` + COLRESET + ` sci shell (default)"
echo " (These can be executed both outside or inside the shell)"
echo;

# Handle fake-prompt
history -r .scishell.hist
while true; do
	dirstr="[$(basename $(pwd))]"
	read -ep "${dirstr} ` + COLBRGREEN + `sci>` + COLRESET + ` " CMD
	history -s "$CMD"
    if [[ $CMD == $'\04' ]]; then
        exit
	elif [[ "true" == $((echo $CMD | grep -Eq "^\!.*") && echo true || echo false) ]]; then
		echo "` + COLYELLOW + `Executing outside scicommander: [${CMD:1}]` + COLRESET + `"
		bash -c "${CMD:1}";
	elif [[ "true" == $((echo $CMD | grep -Eq "^>.*") && echo true || echo false) ]]; then
		sci run "${CMD:1}";
	elif [[ "true" == $((echo $CMD | grep -Eq "^(ls|ll|pwd|lltr|git|tig|tree|t|vim|emacs|nano|history)\>.*") && echo true || echo false) ]]; then
		echo "` + COLYELLOW + `Executing outside scicommander: [$CMD]` + COLRESET + `"
        bash -c "$CMD"
	elif [[ "true" == $((echo $CMD | grep -Eq ".*\<(less|more|bat)\>.*") && echo true || echo false) ]]; then
		echo "` + COLYELLOW + `Executing outside scicommander: [$CMD]` + COLRESET + `"
        bash -c "$CMD"
	elif [[ "true" == $((echo $CMD | grep -Eq "^(cd|c)\>.*") && echo true || echo false) ]]; then
		echo "` + COLYELLOW + `Executing outside scicommander: [$CMD]` + COLRESET + `"
		$CMD;
    elif [[ $CMD == "" ]]; then
		echo "(Exit with Ctrl+C)"
	elif [[ $CMD =~ ^(help|to-html|run|version|shell) ]]; then
		sci $CMD;
    else
        sci run "$CMD"
    fi
	history -a .scishell.hist
done;
echo "Exited SciCommander Shell"
`
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
		fullInputPath := filepath.Join(baseDir, inputPath+".au")
		inputAuditInfos := getInputAuditInfos(fullInputPath, baseDir)
		for inputPath, inputAuditInfo := range inputAuditInfos {
			auditInfos[inputPath] = inputAuditInfo
		}
	}
	return auditInfos
}

func generateDot(auditInfos []AuditInfo) string {
	cmdNodes, fileNodes, edges := generateGraph(auditInfos)

	dot := "DIGRAPH G {\n"
	dot += "  node [shape=box, style=filled, fillcolor=lightgrey, fontname=monospace, penwidth=0, fontsize=11, pad=0];\n"
	for _, node := range cmdNodes {
		dot += f("  \"%s\" [fillcolor=\"#CCE2F1\"]\n", node)
	}
	for _, node := range fileNodes {
		dot += f("  \"%s\" [fillcolor=\"#FFEEC8\"]\n", node)
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
	for _, auditInfo := range auditInfos {
		commandStr := ""
		if len(auditInfo.Executors) > 0 {
			commandStr = strings.ReplaceAll(strings.Join(auditInfo.Executors[0].Command, " "), "\"", "\\\"")
		}
		cmdNodesSet[commandStr] = nil
		if len(auditInfo.Inputs) > 0 {
			for _, input := range auditInfo.Inputs {
				edge := StringTuple{string(input), commandStr}
				edgesSet[edge.String()] = edge
				fileNodesSet[string(input)] = nil
			}
		}
		cmdNodesSet[commandStr] = nil
		if len(auditInfo.Outputs) > 0 {
			for _, output := range auditInfo.Outputs {
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
	max-width: 960px;
	margin: 0 auto;
	box-shadow: 2px 2px 10px #ccc;
	padding: 1em;\n
}
hr {
	border: 2px solid #efefef;
}
table {
	borders: none;
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

func f(s string, v ...interface{}) string {
	return fmt.Sprintf(s, v...)
}

func out(s string, v ...interface{}) {
	fmt.Printf(s+"\n", v...)
}
