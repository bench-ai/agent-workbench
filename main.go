package main

import (
	"agent/browser"
	"agent/command"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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

type runner interface {
	init([]string) error
	run()
	getName() string
}

type RunCommand struct {
	fs                 *flag.FlagSet
	headless           bool
	configIsJsonString bool
}

func (r *RunCommand) init(args []string) error {
	return r.fs.Parse(args)
}

func (r *RunCommand) getName() string {
	return r.fs.Name()
}

func (r *RunCommand) run() {

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

	var config Config

	err = json.Unmarshal(bytes, &config)

	if err != nil {
		log.Fatalf("failed to decode json file: %v", err)
	}

	var browserBuilder browser.Executor
	browserBuilder.Init(r.headless)

	for _, opt := range config.Operations {
		getBrowserSession(opt, &browserBuilder)
	}

	browserBuilder.Execute()
}

func newRunCommand() *RunCommand {
	rc := RunCommand{
		fs: flag.NewFlagSet("run", flag.ExitOnError),
	}

	rc.fs.BoolVar(&rc.headless, "h", false, "use headless mode")

	rc.fs.BoolVar(
		&rc.configIsJsonString,
		"j",
		false,
		"whether or not the string being provided is a json string")

	return &rc
}

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

func root(args []string) error {
	if len(args) < 1 {
		return errors.New("no command passed")
	}

	cmds := []runner{
		newRunCommand(),
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
