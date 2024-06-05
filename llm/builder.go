package llm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// mapToMessage
// converts an interface representing internal data of a message into a message
func mapToMessage(messageType string, messageMap map[string]interface{}) (error, messageInterface) {
	var message messageInterface
	messageByte, err := json.Marshal(messageMap)

	if err != nil {
		return err, nil
	}

	switch strings.ToLower(messageType) {
	case "multimodal":
		message = &gptMultiModalCompliantMessage{}
	case "standard":
		message = &gptStandardMessage{}
	case "assistant":
		message = &gptAssistantMessage{}
	case "tool":
		message = &gptToolMessage{}
	default:
		return fmt.Errorf("%s is not a supported llm message type \n", messageType), nil
	}

	err = json.Unmarshal(messageByte, &message)

	if err != nil {
		return fmt.Errorf("could not read message due to: %v", err), nil
	}

	return nil, message
}

// settingsMapToModel
// converts an interface representing unique settings of a model into a model
func settingsMapToModel(settings map[string]interface{}, maxToken *int, tools *[]Tool, toolChoice interface{}) (error, model) {
	name, ok := settings["name"]

	if !ok {
		return errors.New("parameter name was not found in llm settings, name distinguishes " +
			"what family of LLM's you are using, i.e: openai, llama"), nil
	}

	var settingName string

	if settingName, ok = name.(string); !ok {
		return errors.New("parameter name in llm settings must be of type string"), nil
	}

	var retModel model
	var initErr error

	switch strings.ToLower(settingName) {
	case "openai":

		jsonBytes, err := json.Marshal(settings)
		if err != nil {
			return err, nil
		}

		initErr, retModel = initChatGpt(jsonBytes, maxToken, tools, toolChoice)
	default:
		return fmt.Errorf("%s is not a supported llm \n", name), nil
	}

	return initErr, retModel
}

// Execute
/*
runs a chat completion on a llm with exponential backoff
*/
func Execute(
	messageTypeSlice []string,
	messageMapSlice []map[string]interface{},
	settingsMapSlice []map[string]interface{},
	tools *[]Tool,
	toolChoice interface{},
	maxToken *int,
	tryLimit int16,
	requestWaitTime *int32) (*ChatCompletion, error) {

	if len(messageTypeSlice) != len(messageMapSlice) {
		panic("type slice and array slice lengths are not the same in llm.Execute")
	}

	messageSlice := make([]messageInterface, len(messageTypeSlice))
	modelSlice := make([]model, 0, len(settingsMapSlice))

	for i := range messageSlice {
		err, mess := mapToMessage(messageTypeSlice[i], messageMapSlice[i])
		if err != nil {
			return nil, err
		}
		messageSlice[i] = mess
	}

	for i := range modelSlice {
		err, mod := settingsMapToModel(settingsMapSlice[i], maxToken, tools, toolChoice)
		if err != nil {
			modelSlice = append(modelSlice, mod)
		} else {
			fmt.Println(err)
		}
	}

	return exponentialBackoff(modelSlice, &messageSlice, tryLimit, requestWaitTime)
}
