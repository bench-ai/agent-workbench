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
	"path/filepath"
)

type Credentials struct {
	Name   string `json:"Name"`
	APIKey string `json:"apiKey"`
}

type Workflow struct {
	WorkflowType string `json:"workflow_type"`
}
type Settings struct {
	Timeout     *int16                   `json:"timeout"`
	Headless    bool                     `json:"headless"`
	MaxToken    *int                     `json:"max_tokens"`
	Credentials []Credentials            `json:"credentials"`
	Workflow    Workflow                 `json:"workflow"`
	LLMSettings []map[string]interface{} `json:"llm_settings"`
	TryLimit    int16                    `json:"try_limit"`
}

type Command struct {
	CommandName string                 `json:"command_name,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
	MessageType string                 `json:"message_type"`
	Message     interface{}            `json:"message"`
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

func collectSettings(llmSettings map[string]interface{}, key string, required bool) interface{} {
	if val, ok := llmSettings[key]; ok {
		return val
	}

	if required {
		log.Fatalf(`setting: '%s' not found`, key)
	}

	return nil
}

// create an array of LLMs and calls exponential backoff on the array of messages built in addLlmOpperations
func runLlmCommands(settings Settings, commandList []Command) {

	var llmArray []command.LLM
	for _, item := range settings.LLMSettings {

		name, ok := item["name"]

		if !ok {
			log.Fatal("LLM setting name not found")
		}

		switch name {
		case "OpenAI":
			apiKey, ok := collectSettings(item, "api_key", true).(string)
			if !ok {
				log.Fatal("api_key must be a string")
			}

			model, ok := collectSettings(item, "model", true).(string)
			if !ok {
				log.Fatal("model must be a string")
			}

			temp := collectSettings(item, "temperature", false)
			var temperature float64
			if temp != nil {
				if temperature, ok = temp.(float64); !ok {
					log.Fatal("temperature must be a float")
				}
			}

			tempfix := float32(temperature)

			gpt := command.InitChatgpt(model, apiKey, settings.MaxToken, &tempfix)
			llmArray = append(llmArray, gpt)
		default:
			log.Fatalf("%s is not a supported llm \n", name)
		}
	}

	messageList := addLlmOperation(commandList)

	chat, err := command.ExponentialBackoff(llmArray, &messageList, settings.TryLimit, settings.Timeout)

	if err != nil {
		log.Fatalf("chat com is not a supported llm \n")
	}

	for _, sett := range settings.LLMSettings {
		delete(sett, "api_key")
	}

	type writeStruct struct {
		SettingsSlice []map[string]interface{} `json:"settings"`
		Completion    *command.ChatCompletion  `json:"completion"`
		MessageList   []Command                `json:"message_list"`
	}

	msg := command.ConvertChatCompletion(chat)
	commandList = append(commandList, Command{
		Message:     msg,
		MessageType: "assistant",
	})

	writeData := writeStruct{
		SettingsSlice: settings.LLMSettings,
		Completion:    chat,
		MessageList:   commandList,
	}

	if er := os.MkdirAll("./resources", os.ModePerm); !os.IsExist(er) && er != nil {
		log.Fatal("Could not create directory: " + "./resources")
	}

	b, err := json.MarshalIndent(writeData, "", "    ")

	if err != nil {
		log.Fatal("Could not marshall llm response")
	}

	err = os.WriteFile(filepath.Join("./resources", "completion.json"), b, 0666)

	if err != nil {
		log.Fatal("Could not write llm response")
	}
}

type Configuration struct {
	Operations []Operation `json:"operations"`
}

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

	for _, op := range config.Operations {
		switch op.Type {
		case "browser":
			runBrowserCommands(op.Settings, op.CommandList)
		case "llm":
			runLlmCommands(op.Settings, op.CommandList)
		default:
			log.Fatalf("unknown operation type: %s", op.Type)
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

// switch on message_type and builds an array of messages
func addLlmOperation(msgSlice []Command) []command.MessageInterface {

	var retSlice []command.MessageInterface
	for _, msg := range msgSlice {
		messageType := msg.MessageType

		var message command.MessageInterface

		messageByte, err := json.Marshal(msg.Message)

		if err != nil {
			log.Fatalf("could not marshal message due to: %v", err)
		}

		switch messageType {
		case "multimodal":
			message = &command.GPTMultiModalCompliantMessage{}
		case "standard":
			message = &command.GPTStandardMessage{}
		case "assistant":
			message = &command.GptAssistantMessage{}
		case "tool":
			message = &command.GptToolMessage{}
		default:
			log.Fatalf("%s is not a supported llm message type \n", messageType)
		}

		err = json.Unmarshal(messageByte, &message)

		if err != nil {
			log.Fatalf("could not read message due to: %v", err)
		}

		retSlice = append(retSlice, message)
	}

	return retSlice
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
