package llm

import (
	"encoding/json"
	"log"
)

func collectSettings(llmSettings map[string]interface{}, key string, required bool) interface{} {
	if val, ok := llmSettings[key]; ok {
		return val
	}

	if required {
		log.Fatalf(`setting: '%s' not found`, key)
	}

	return nil
}

// switch on message_type and builds an array of messages
func mapToMessage(messageType string, messageMap map[string]interface{}) messageInterface {
	var message messageInterface
	messageByte, err := json.Marshal(messageMap)

	if err != nil {
		log.Fatalf("could not marshal message due to: %v", err)
	}

	switch messageType {
	case "multimodal":
		message = &gptMultiModalCompliantMessage{}
	case "standard":
		message = &gptStandardMessage{}
	case "assistant":
		message = &gptAssistantMessage{}
	case "tool":
		message = &gptToolMessage{}
	default:
		log.Fatalf("%s is not a supported llm message type \n", messageType)
	}

	err = json.Unmarshal(messageByte, &message)

	if err != nil {
		log.Fatalf("could not read message due to: %v", err)
	}

	return message
}

func settingsMapToModel(settings map[string]interface{}, maxToken *int, tools *[]Tool, toolChoice interface{}) model {
	name, ok := settings["name"]

	if !ok {
		log.Fatal("modelType setting name not found")
	}

	err := validateToolChoice(toolChoice)
	if err != nil {
		log.Fatal(err)
	}
	var retModel model

	switch name {
	case "OpenAI":
		apiKey, ok := collectSettings(settings, "api_key", true).(string)
		if !ok {
			log.Fatal("api_key must be a string")
		}

		modelType, ok := collectSettings(settings, "model", true).(string)
		if !ok {
			log.Fatal("model must be a string")
		}

		temp := collectSettings(settings, "temperature", false)
		var temperature float64
		if temp != nil {
			if temperature, ok = temp.(float64); !ok {
				log.Fatal("temperature must be a float")
			}
		}

		tempfix := float32(temperature)

		retModel = initChatGpt(modelType, apiKey, maxToken, &tempfix, tools, toolChoice)
	default:
		log.Fatalf("%s is not a supported llm \n", name)
	}

	return retModel
}

func Execute(
	messageTypeSlice []string,
	messageMapSlice []map[string]interface{},
	settingsMapSlice []map[string]interface{},
	tools *[]Tool,
	toolChoice interface{},
	maxToken *int,
	tryLimit int16,
	requestWaitTime *int16) (*ChatCompletion, error) {

	if len(messageTypeSlice) != len(messageMapSlice) {
		log.Fatal("type slice and array slice lengths are not the same in llm.Execute")
	}

	messageSlice := make([]messageInterface, len(messageTypeSlice))
	modelSlice := make([]model, len(settingsMapSlice))

	for i := range messageSlice {
		messageSlice[i] = mapToMessage(messageTypeSlice[i], messageMapSlice[i])
	}

	for i := range modelSlice {
		modelSlice[i] = settingsMapToModel(settingsMapSlice[i], maxToken, tools, toolChoice)
	}

	return exponentialBackoff(modelSlice, &messageSlice, tryLimit, requestWaitTime)
}
