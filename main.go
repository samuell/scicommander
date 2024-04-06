package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
	//cmdBase := cmdParts[0]
	cmdArgs := cmdParts[1:]

	inFiles, outFiles := detectFiles(cmdArgs)
	out("Infiles: %v, Outfiles: %v", inFiles, outFiles)

	// Write shell script for each output file
	//cmdScript := "!/bin/bash\n" + cmdStr + "\n"
	//for _, outFile := range outFiles {
	//	os.WriteFile(outFile+".sh", []byte(cmdScript), 0744)
	//}

	filesBefore, err := filepath.Glob("./*")
	checkMsg(err, "Could not glob folder before executing command!")

	// Execute the command
	bashArgs := strings.Join(cmdParts, " ")
	out("Executing command: %v", bashArgs)
	cmd := exec.Command("bash", "-c", bashArgs)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	errMsg := fmts("Could not run command: %s\nSTDERR: %s\nSTDOUT: %s", cmdStr, cmd.Stderr, cmd.Stdout)
	checkMsg(err, errMsg)

	filesAfter, err := filepath.Glob("./*")
	checkMsg(err, "Could not glob folder after executing command!")

	newFiles := []string{}
	numFiles := len(filesBefore)
	for _, file := range filesAfter {
		if !slices.Contains(filesBefore, file) {
			newFiles = append(newFiles, file)
			fmt.Printf("New file found after checking against %d files: %v\n", numFiles, file)
		}
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

func out(s string, v ...interface{}) {
	fmt.Printf(s+"\n", v...)
}
