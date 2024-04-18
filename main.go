package main

import (
	"agent/APIs"
	"agent/browser"
	"agent/command"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

type Settings struct {
	Timeout  *int16 `json:"timeout"`
	Headless bool   `json:"headless"`
}

type Command struct {
	CommandName string                 `json:"command_name"`
	Params      map[string]interface{} `json:"params"`
}

type Operation struct {
	Type        string    `json:"type"`
	Settings    Settings  `json:"settings"`
	CommandList []Command `json:"command_list"`
}

func runBrowserCommands(settings Settings, commandList []Command) {
	var browserBuilder browser.Executor
	browserBuilder.Init(settings.Headless, settings.Timeout)

	for _, com := range commandList {
		addOperation(com, &browserBuilder)
	}

	browserBuilder.Execute()
}

// the following 3 structs are LLM operation structs

type LlmSettings struct {
	Timeout *int16 `json:"timeout"`
}

type LlmCommand struct {
	CommandName string                 `json:"command_name"`
	Params      map[string]interface{} `json:"params"`
}

type LlmOperation struct {
	Type        string       `json:"type"`
	Settings    LlmSettings  `json:"settings"`
	CommandList []LlmCommand `json:"command_list"`
}

func runLlmCommands(settings LlmSettings, commandList []LlmCommand) {
	var apiBuilder APIs.APIExecutor
	apiBuilder.Init(settings.Timeout) //need to update in apicalls.go

	for _, com := range commandList {
		addLlmOperation(com, &apiBuilder)
	}

	apiBuilder.Execute()
}

// Configuration struct creates a list called Operations that is of type interface, so it can be a list of either Operation or LlmOperation
type Configuration struct {
	Operations []interface{} `json:"operations"`
}

// OperationInterface isOperation LlmOperationInterface and isLlmOperation mark that Operation and LlmOperation implement an interface
type OperationInterface interface {
	isOperation()
}

func (o Operation) isOperation() {}

type LlmOperationInterface interface {
	isLlmOperation()
}

func (o LlmOperation) isLlmOperation() {}

type runner interface {
	init([]string) error
	run()
	getName() string
}

type runCommand struct {
	fs                 *flag.FlagSet
	configIsJsonString bool
}

func (r *runCommand) init(args []string) error {
	return r.fs.Parse(args)
}

func (r *runCommand) getName() string {
	return r.fs.Name()
}

// run
/**
The run command, checks if the user wishes to run their browser in headless mode, and whether they are pointing to
a file or passing raw json
*/
func (r *runCommand) run() {

	if err := os.RemoveAll("./resources"); err != nil {
		log.Fatal("cannot create resources directory due to: ", err)
	}

	configString := r.fs.Arg(0)

	if configString == "" {
		log.Fatal("invalid config argument")
	}

	var bytes []byte
	var err error

	if r.configIsJsonString {
		bytes = []byte(configString)
	} else {
		bytes, err = os.ReadFile(configString)
	}

	if err != nil {
		log.Fatalf("failed to read json file due to: %v", err)
	}

	var config Configuration

	err = json.Unmarshal(bytes, &config)

	if err != nil {
		log.Fatalf("failed to decode json file: %v", err)
	}

	for _, opInterface := range config.Operations {
		switch op := opInterface.(type) {
		case Operation:
			switch op.Type {
			case "browser":
				runBrowserCommands(op.Settings, op.CommandList)
			default:
				log.Fatalf("unknown operation type: %s", op.Type)
			}
		case LlmOperation:
			switch op.Type {
			case "llm":
				runLlmCommands(op.Settings, op.CommandList)
			default:
				log.Fatalf("unknown operation type: %s", op.Type)
			}
		default:
			log.Fatalf("unknown operation type")
		}
	}
}

func newRunCommand() *runCommand {
	rc := runCommand{
		fs: flag.NewFlagSet("run", flag.ExitOnError),
	}

	rc.fs.BoolVar(
		&rc.configIsJsonString,
		"j",
		false,
		"whether or not the string being provided is a json string")

	return &rc
}

type versionCommand struct {
	fs *flag.FlagSet
}

func (v *versionCommand) init(args []string) error {
	return v.fs.Parse(args)
}

func (v *versionCommand) run() {
	fmt.Println("Version 0.0.0")
}

func (v *versionCommand) getName() string {
	return v.fs.Name()
}

func newVersionCommand() *versionCommand {
	vc := versionCommand{
		fs: flag.NewFlagSet("version", flag.ExitOnError),
	}

	return &vc
}

// addOperation
/*
checks for if an operation exists and adds it to the execution queue
*/
func addOperation(com Command, builder *browser.Executor) {

	paramBytes, _ := json.Marshal(com.Params)
	var browserParams command.BrowserParams

	switch com.CommandName {
	case "open_web_page":
		browserParams = &command.OpenWebPage{}
	case "full_page_screenshot":
		browserParams = &command.FullPageScreenShot{}
	case "element_screenshot":
		browserParams = &command.ElementScreenshot{}
	case "collect_nodes":
		browserParams = &command.CollectNodes{}
	case "click":
		browserParams = &command.Click{}
	case "save_html":
		browserParams = &command.SaveHtml{}
	case "sleep":
		browserParams = &command.Sleep{}
	default:
		log.Fatalf("%s is not a supported browser command \n", com.CommandName)
	}

	if err := json.Unmarshal(paramBytes, browserParams); err != nil {
		log.Fatalf("failed to parse %s command \n", com.CommandName)
	}

	if err := browserParams.Validate(); err != nil {
		log.Fatalf("%v", err)
	}

	browserParams.AppendTask(builder)
}

func addLlmOperation(com LlmCommand, builder *APIs.APIExecutor) {

	/*paramBytes, _ := json.Marshal(com.Params)
	var browserParams command.BrowserParams

	switch com.CommandName {
	case "open_web_page":
		browserParams = &command.OpenWebPage{}
	case "full_page_screenshot":
		browserParams = &command.FullPageScreenShot{}
	case "element_screenshot":
		browserParams = &command.ElementScreenshot{}
	case "collect_nodes":
		browserParams = &command.CollectNodes{}
	case "click":
		browserParams = &command.Click{}
	case "save_html":
		browserParams = &command.SaveHtml{}
	case "sleep":
		browserParams = &command.Sleep{}
	default:
		log.Fatalf("%s is not a supported browser command \n", com.CommandName)
	}

	if err := json.Unmarshal(paramBytes, browserParams); err != nil {
		log.Fatalf("failed to parse %s command \n", com.CommandName)
	}

	if err := browserParams.Validate(); err != nil {
		log.Fatalf("%v", err)
	}

	browserParams.AppendTask(builder)*/
}

// root
/*
Checks for present subcommands and executes them
*/
func root(args []string) error {
	if len(args) < 1 {
		return errors.New("no command passed")
	}

	cmds := []runner{
		newRunCommand(),
		newVersionCommand(),
	}

	subcommand := os.Args[1]

	for _, cmd := range cmds {
		if cmd.getName() == subcommand {
			if err := cmd.init(os.Args[2:]); err == nil {
				cmd.run()
				return nil
			} else {
				return err
			}
		}
	}

	return fmt.Errorf("unknown command: %s", subcommand)
}

func main() {
	if err := root(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
