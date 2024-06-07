package main

import (
	"os"
	"path"
	"strings"
	"testing"
)

func tempDir() (error, string) {
	td, err := os.MkdirTemp("", "temp_dir")
	return err, td
}

func removeTemp(tempPath string, t *testing.T) {
	err := os.RemoveAll(tempPath)
	if err != nil {
		t.Fatal(err)
	}
}

func scanDir(startPath string, t *testing.T, suffix string) bool {
	fileInfo, err := os.ReadDir(startPath)

	if err != nil {
		t.Fatal(err)
	}

	for _, file := range fileInfo {
		if strings.HasSuffix(path.Join(startPath, file.Name()), suffix) {
			return true
		} else {
			if file.IsDir() {
				if scanDir(path.Join(startPath, file.Name()), t, suffix) {
					return true
				}
			}
		}
	}

	return false
}

func TestCreateSessionDirectory(t *testing.T) {
	err, tempDirect := tempDir()

	if err != nil {
		t.Fatal(err)
	}

	defer removeTemp(tempDirect, t)

	err = os.Setenv("BENCHAI_SAVEDIR", tempDirect)

	if err != nil {
		t.Fatal(err)
	}

	createSessionDirectory("test-sess")

	if !scanDir(tempDirect, t, "test-sess") {
		t.Fatal("session dir was not generated")
	}
}
