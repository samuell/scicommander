package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRunCommandWithDeepFolderStructure(t *testing.T) {
	tmpDir := t.TempDir()
	fmt.Printf("Moving into %v ...\n", tmpDir)
	os.Chdir(tmpDir)

	type testCase struct {
		commands           []string
		lastAuditInfo      string
		wantAuditCount     int
		wantCommandsInHTML []string
	}

	tests := []testCase{
		{
			commands: []string{
				"echo hello > hello.txt",
			},
			lastAuditInfo:  "hello.txt.au",
			wantAuditCount: 1,
		},
		{
			commands: []string{
				"echo ACGT > seq.txt",
				"mkdir -p o/ou/out && rev seq.txt > o/ou/out/rev.txt",
				"cat o/ou/out/rev.txt > rev.txt",
			},
			lastAuditInfo:  "rev.txt.au",
			wantAuditCount: 3,
		},
		{
			commands: []string{
				"echo ACGT > s.txt",
				"mkdir -p o/ou/out && rev s.txt > o/ou/out/r.txt",
				"cat o/ou/out/r.txt > r.txt",
			},
			lastAuditInfo:  "r.txt.au",
			wantAuditCount: 3,
		},
		{
			commands: []string{
				"cat rev.txt r.txt > rev-and-r.txt",
			},
			lastAuditInfo:  "rev-and-r.txt.au",
			wantAuditCount: 7,
			wantCommandsInHTML: []string{
				"echo ACGT > seq.txt",
				"mkdir -p o/ou/out && rev seq.txt > o/ou/out/rev.txt",
				"cat o/ou/out/rev.txt > rev.txt",
				"echo ACGT > s.txt",
				"mkdir -p o/ou/out && rev s.txt > o/ou/out/r.txt",
				"cat o/ou/out/r.txt > r.txt",
				"cat rev.txt r.txt > rev-and-r.txt",
			},
		},
	}

	for _, tc := range tests {
		for _, cmd := range tc.commands {
			executeCommand(cmd)
		}
		auditInfos := getAllUpstreamAuditInfos(tc.lastAuditInfo)
		haveAuditCount := len(auditInfos)
		if haveAuditCount != tc.wantAuditCount {
			t.Fatal(f("Wrong number of audit infos found! Expected %d but found %d", tc.wantAuditCount, haveAuditCount))
		}

		htmlPath := toHtml(tc.lastAuditInfo)
		html, err := ioutil.ReadFile(htmlPath)
		checkMsg(err, f("Could not read file %s", htmlPath))

		checkForCommands := tc.commands
		if tc.wantCommandsInHTML != nil {
			checkForCommands = tc.wantCommandsInHTML
		}

		for _, cmd := range checkForCommands {
			if !strings.Contains(string(html), cmd) {
				t.Fatal(f("Could not find command [%s] in html-file %s", cmd, htmlPath))
			}
		}
	}
}

func TestRunCommand(t *testing.T) {
	tmpDir := t.TempDir()
	fmt.Printf("Moving into %v ...\n", tmpDir)
	os.Chdir(tmpDir)

	type testCase struct {
		command      string
		wantOutFiles []string
	}

	tests := []testCase{
		{command: "mkdir out && echo ACGT > out/seq.txt", wantOutFiles: []string{"out/seq.txt"}},
		{command: "mkdir -p foo/bar/baz && echo ACGT > foo/bar/baz/seq.txt", wantOutFiles: []string{"foo/bar/baz/seq.txt"}},
	}

	for _, tc := range tests {
		executeCommand(tc.command)
		for _, wantedOutFile := range tc.wantOutFiles {
			if _, err := os.Stat(wantedOutFile); os.IsNotExist(err) {
				t.Fatal(f("Could not find wanted outfile %s after execution of [%s]", wantedOutFile, tc.command))
			}
			wantedAuditFile := wantedOutFile + ".au"
			if _, err := os.Stat(wantedAuditFile); os.IsNotExist(err) {
				t.Fatal(f("Could not find wanted audit file %s after execution of [%s]", wantedAuditFile, tc.command))
			}
		}
	}
}

func TestDetectFiles(t *testing.T) {
	tmpDir := t.TempDir()
	fmt.Printf("Moving into %v ...\n", tmpDir)
	os.Chdir(tmpDir)

	// Arrange
	wantInFiles := []string{
		"foo.txt",
		filepath.Join("bar", "baz.xyz"),
		filepath.Join("bar", "xyz.abc"),
	}
	for _, f := range wantInFiles {
		createDirAndFile(f)
	}

	wantOutFiles := []string{"out.png", filepath.Join("out", "someresult.tar.gz")}

	type testCase struct {
		command []string
	}

	stringsToCheck := []string{
		"foo.txt",
		"out.png",
		">",
		"|",
		filepath.Join("bar", "baz.xyz"),
		filepath.Join("bar", "xyz.abc"),
		filepath.Join("out", "someresult.tar.gz"),
	}

	// Act
	haveInFiles, haveOutFiles := detectFiles(stringsToCheck)

	// Assert
	if !reflect.DeepEqual(haveInFiles, wantInFiles) {
		t.Fatalf("Wanted infiles %v but got %v\n", wantInFiles, haveInFiles)
	}
	if !reflect.DeepEqual(haveOutFiles, wantOutFiles) {
		t.Fatalf("Wanted outfiles %v but got %v\n", wantOutFiles, haveOutFiles)
	}
}

func createDirAndFile(filePath string) {
	baseDir := filepath.Dir(filePath)
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		err := os.MkdirAll(baseDir, 0744)
		checkMsg(err, "Could not create dir: "+baseDir)
	}
	_, err := os.Create(filePath)
	checkMsg(err, "Could not create file: "+filePath)
}
