package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDetectFiles(t *testing.T) {
	tmpDir := t.TempDir()
	fmt.Printf("Moving into %v ...\n", tmpDir)
	os.Chdir(tmpDir)

	// Arrange
	wantInFiles := []string{"foo.txt", filepath.Join("bar", "baz.xyz"), filepath.Join("bar", "xyz.abc")}
	for _, f := range wantInFiles {
		createDirAndFile(f)
	}

	wantOutFiles := []string{"out.png"}

	stringsToCheck := []string{
		"foo.txt",
		"out.png",
		">",
		"|",
		filepath.Join("bar", "baz.xyz"),
		filepath.Join("bar", "xyz.abc"),
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
