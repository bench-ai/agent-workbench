package main

import (
	"agent/helper"
	"context"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type liveCommand struct {
}

func collectCommandFiles(sessionPath string, completedCommands helper.Set[string]) []string {

	pth := filepath.Join(sessionPath, "commands")
	dirSlice, err := os.ReadDir(pth)

	if err != nil {
		log.Fatal(err)
	}

	type fileCommand struct {
		modTime  time.Time
		filename string
	}

	newCommands := make([]fileCommand, 0, 3)

	for _, el := range dirSlice {
		if !el.IsDir() {
			info, err := el.Info()

			if err != nil {
				log.Fatal(err)
			}

			if !completedCommands.Has(el.Name()) {
				newCommands = append(newCommands, fileCommand{
					modTime:  info.ModTime(),
					filename: el.Name(),
				})
			}
		}
	}

	sort.Slice(newCommands, func(i, j int) bool {
		return newCommands[i].modTime.Unix() < newCommands[j].modTime.Unix()
	})

	retCommands := make([]string, len(newCommands))

	for index, com := range newCommands {
		retCommands[index] = com.filename
	}

	return retCommands
}

func getLiveSession(sessionPath string) chromedp.ActionFunc {
	return func(c context.Context) error {
		var commandSet helper.Set[string]
		commandSlice := collectCommandFiles(sessionPath, commandSet)

		return nil
	}
}
