package main

import (
	"agent/browser"
	"agent/command"
	"encoding/json"
	"log"
	"os"
)

type Operation struct {
	Type        string                 `json:"type"`
	CommandName string                 `json:"command_name"`
	Params      map[string]interface{} `json:"params"`
}

type Config struct {
	Operations []Operation `json:"operations"`
}

var config Config

func getBrowserSession(operation Operation, builder *browser.Executor) {

	paramBytes, _ := json.Marshal(operation.Params)
	var browserParams command.BrowserParams

	switch operation.CommandName {
	case "open_web_page":
		browserParams = &command.OpenWebPage{}
	case "full_page_screenshot":
		browserParams = &command.FullPageScreenShot{}
	case "element_screenshot":
		browserParams = &command.ElementScreenshot{}
	case "collect_nodes":
		browserParams = &command.CollectNodes{}
	default:
		log.Fatalf("%s is not a supported browser command \n", operation.CommandName)
	}

	if err := json.Unmarshal(paramBytes, browserParams); err != nil {
		log.Fatalf("failed to parse %s command \n", operation.CommandName)
	}

	if err := browserParams.Validate(); err != nil {
		log.Fatalf("%v", err)
	}

	browserParams.AppendTask(builder)
}

func init() {
	if err := os.RemoveAll("./resources"); err != nil {
		log.Fatal("cannot create resources directory due to: ", err)
	}

	err := os.MkdirAll("./resources/images", os.ModePerm)
	if !os.IsExist(err) && err != nil {
		log.Fatal("cannot create resources directory due to: ", err)
	}

	jsonPath := os.Args[1]

	bytes, err := os.ReadFile(jsonPath)

	if err != nil {
		log.Fatalf("failed to read json file due to: %v", err)
	}

	err = json.Unmarshal(bytes, &config)

	if err != nil {
		log.Fatalf("failed to decode json file: %v", err)
	}
}

func main() {
	var browserBuilder browser.Executor
	browserBuilder.Init(false)

	for _, opt := range config.Operations {
		getBrowserSession(opt, &browserBuilder)
	}

	browserBuilder.Execute()
}
