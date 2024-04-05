package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	flag.NewFlagSet("run", flag.ExitOnError)
	flag.NewFlagSet("htmlize", flag.ExitOnError)
	flag.Parse()

	usage := `
Usage:
sci run <command>
sci htmlize <html-file>
`

	if len(os.Args) < 2 {
		fmt.Println("ERROR: No command supplied")
		fmt.Println(usage)
		os.Exit(2)
		return
	}

	switch os.Args[1] {
	case "run":
		cmdStr := strings.Join(os.Args[2:], " ")
		executeCommand(cmdStr)
	case "htmlize":
		fmt.Println("You ran htmlize")
	default:
		fmt.Println("ERROR: Expected run or htmlize")
	}
}

func executeCommand(cmdStr string) {
	cmdParts := strings.Split(cmdStr, " ")
	cmdBase := cmdParts[0]
	cmdArgs := cmdParts[1:]

	inFiles, outFiles := detectFiles(cmdArgs)
	fmt.Printf("Infiles: %v, Outfiles: %v\n", inFiles, outFiles)

	cmd := exec.Command(cmdBase, cmdArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	errMsg := fmts("Could not run command: %s", cmdStr)
	checkMsg(err, errMsg)
}

func detectFiles(strs []string) ([]string, []string) {
	inFiles := []string{}
	outFiles := []string{}
	for _, str := range strs {
		if _, err := os.Stat(str); os.IsNotExist(err) {
		} else {
			auditPath := str + ".au"
			if _, err := os.Stat(auditPath); os.IsNotExist(err) {
				inFiles = append(inFiles, str)
			} else {
				outFiles = append(outFiles, str)
			}
		}
	}
	return inFiles, outFiles
}

func checkMsg(err error, message string) {
	if err != nil {
		fmt.Println(message)
		fmt.Println(err)
		os.Exit(126)
	}
}

func fmts(s string, v ...interface{}) string {
	return fmt.Sprintf(s, v...)
}
