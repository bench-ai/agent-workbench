package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteErr(t *testing.T) {
	err, tempPath := tempDir()

	if err == nil {
		defer removeTemp(tempPath, t)
	} else {
		t.Fatal(err)
	}

	err = errors.New("test error")

	err = writeErr(tempPath, err)

	if err != nil {
		t.Fatal(err)
	}

	if bytes, err := os.ReadFile(filepath.Join(tempPath, "err.txt")); err == nil {
		text := string(bytes)

		if text != "test error" {
			t.Fatal("error not written to file")
		}
	} else {
		t.Fatal(err)
	}
}

func TestWriteSuccess(t *testing.T) {
	err, tempPath := tempDir()

	if err == nil {
		defer removeTemp(tempPath, t)
	} else {
		t.Fatal(err)
	}

	err = writeSuccess(tempPath)

	if err != nil {
		t.Fatal(err)
	}

	if _, err = os.ReadFile(filepath.Join(tempPath, "success.txt")); err != nil {
		t.Fatal(err)
	}
}

func TestEndSession(t *testing.T) {
	err, tempPath := tempDir()

	if err == nil {
		defer removeTemp(tempPath, t)
	} else {
		t.Fatal(err)
	}

	err = endSession(tempPath, errors.New("test error"))

	if err != nil {
		t.Fatal(err)
	}

	if bytes, err := os.ReadFile(filepath.Join(tempPath, "exit.txt")); err == nil {
		text := string(bytes)

		if text != "test error" {
			t.Fatal("error not written to file")
		}
	} else {
		t.Fatal(err)
	}
}

func TestCreateLiveFolders(t *testing.T) {
	err, tempDirectory := tempDir()

	if err == nil {
		defer removeTemp(tempDirectory, t)
	} else {
		t.Fatal(err)
	}

	if err := os.Setenv("BENCHAI_SAVEDIR", tempDirectory); err != nil {
		t.Fatal(err)
	}

	pth := createSessionDirectory("test-sess")

	createLiveFolders(pth)

	commandPath := filepath.Join(pth, "commands")
	responsePath := filepath.Join(pth, "responses")

	if _, err := os.Stat(commandPath); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(responsePath); err != nil {
		t.Fatal(err)
	}
}
