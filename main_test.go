package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDetectFiles(t *testing.T) {
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Arrange
	wantFiles := []string{"foo.txt", filepath.Join("bar/baz.xyz")}
	for _, fileToCreate := range wantFiles {
		baseDir := filepath.Dir(fileToCreate)
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			err := os.MkdirAll(baseDir, 0744)
			checkMsg(err, "Could not create dir: "+baseDir)
		}
		_, err := os.Create(filepath.Join(tmpDir, fileToCreate))
		checkMsg(err, "Could not create file: "+fileToCreate)
	}

	stringsToCheck := []string{"foo.txt", "foz.tsv", filepath.Join("bar/baz.xyz"), "baz.csv"}

	// Act
	haveFiles := detectFiles(stringsToCheck)

	// Assert
	if !reflect.DeepEqual(haveFiles, wantFiles) {
		t.Fatalf("Wanted %v but got %v\n", wantFiles, haveFiles)
	}
}
