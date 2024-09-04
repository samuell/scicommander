package main

import (
	"bytes"
	"cmp"
	"encoding/json"
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
		fmt.Println("ERROR: No command supplied")
		flag.Usage()
		os.Exit(2)
		return
	}

	switch os.Args[1] {
	case "run":
		cmdStr := strings.Join(os.Args[2:], " ")
		executeCommand(cmdStr)
	case "to-html":
		auditPath := strings.Join(os.Args[2:], " ")
		toHtml(auditPath)
	case "shell":
		runShell()
	default:
		fmt.Println("ERROR: Expected run or to-html")
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
}

func runShell() {
	exe, lookErr := exec.LookPath("bash")
	checkMsg(lookErr, "Could not find command \"bash\"")

	tempScriptPath := ".scishell.bash"

	shellCode := `#!/bin/bash -l
echo "Starting SciCommander Shell"
echo "(Exit by pressing Ctrl+C)"
echo "------------------------------------------------"
while true; do
    read -ep "sci> " cmd
    if [[ $cmd == $'\04' ]]; then
        exit
    elif [[ $cmd =~ (ls|ll|pwd|cd|vim|emacs|nano|less|more).* ]]; then
        $cmd
    elif [[ $cmd == "" ]]; then
        echo "Command was emtpy. Did you want to exit?"
        echo "Exit by pressing: Ctrl+C"
    else
        echo "Executing command via SciCommander: $cmd"
        sci run "$cmd"
    fi
done;
echo "Exited SciCommander Shell"
`
	wrtErr := os.WriteFile(tempScriptPath, []byte(shellCode), 0644)
	checkMsg(wrtErr, f("Could not write %s", tempScriptPath))

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
		auditInfos = append(auditInfos, auditInfo)
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
	nodes, edges := generateGraph(auditInfos)

	dot := "DIGRAPH G {\n"
	dot += "  node [shape=box, style=filled, fillcolor=lightgrey, fontname=monospace, penwidth=0];\n"
	for _, node := range nodes {
		dot += f("  \"%s\"\n", node)
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

func generateGraph(auditInfos []AuditInfo) (nodes []string, edges []StringTuple) {
	nodesSet := map[string]interface{}{}
	edgesSet := map[string]StringTuple{}
	for _, auditInfo := range auditInfos {
		commandStr := strings.ReplaceAll(strings.Join(auditInfo.Executors[0].Command, " "), "\"", "\\\"")
		nodesSet[commandStr] = nil
		if len(auditInfo.Inputs) > 0 {
			for _, input := range auditInfo.Inputs {
				edge := StringTuple{string(input), commandStr}
				edgesSet[edge.String()] = edge
				nodesSet[string(input)] = nil
			}
		}
		if len(auditInfo.Outputs) > 0 {
			for _, output := range auditInfo.Outputs {
				edge := StringTuple{commandStr, string(output)}
				edgesSet[edge.String()] = edge
				nodesSet[string(output)] = nil
			}
		}
	}
	for node, _ := range nodesSet {
		nodes = append(nodes, node)
	}
	for _, edge := range edgesSet {
		edges = append(edges, edge)
	}
	return
}

func unmarshalAuditInfo(auditPath string) AuditInfo {
	auditJson, readErr := ioutil.ReadFile(auditPath)
	checkMsg(readErr, f("Error reading file %s", auditPath))
	var auditInfo AuditInfo
	err := json.Unmarshal([]byte(auditJson), &auditInfo)
	checkMsg(err, "Failed to unmarshal JSON file "+auditPath)
	return auditInfo
}

func auditInfosToHTML(auditInfos []AuditInfo, outPath string, svgPath string) (html string) {
	svgRaw, err := os.ReadFile(svgPath)
	svg := string(svgRaw)
	checkMsg(err, "Could not read SVG file: "+svgPath)
	html = "<html>\n"
	html += "<head>\n"
	html += "<style>\n"
	html += `
body{
	font-family:monospace, courier new;
	max-width: 1024px;
	margin: 0 auto;
	box-shadow: 2px 2px 10px #ccc;
	padding: 1em;\n
}
hr {
	border: 2px solid #eee;
}
table {
	borders: none;
}
table td {
	padding: 1em;
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
		html += f("<tr><td>%s</td>"+
			"<td style='background: #efefef;'>%s</td>"+
			"<td>%d ms</tr>"+
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
