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
scicmdr run <command>
scicmdr htmlize <html-file>
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
	cmd := exec.Command(cmdBase, cmdArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	errMsg := fmts("Could not run command: %s", cmdStr)
	checkMsg(err, errMsg)
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
