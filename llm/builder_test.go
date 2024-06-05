package llm

import (
	"errors"
	"testing"
)

func TestMapToMessage(t *testing.T) {

	messageMap := map[string]interface{}{
		"role":    "system",
		"content": "we will succeed",
	}

	err, _ := mapToMessage("unknown", messageMap)
	if err == nil {
		t.Error("processed command with invalid name")
	}

	messageMap = map[string]interface{}{
		"one": complex128(100),
	}

	err, _ = mapToMessage("standard", messageMap)

	if err == nil {
		t.Error(err)
	}

	messageMap = map[string]interface{}{
		"role": 10,
	}

	err, _ = mapToMessage("standard", messageMap)

	if err == nil {
		t.Error(err)
	}

	messageMap = map[string]interface{}{
		"role":    "system",
		"content": "we will succeed",
	}

	err, _ = mapToMessage("standard", messageMap)

	if err != nil {
		t.Fatal("could not marshall valid value")
	}
}

func TestSettingsMapToModel(t *testing.T) {
	testSettings := make(map[string]interface{}, 4)

	err, _ := settingsMapToModel(testSettings, nil, nil, nil)

	if err == nil {
		t.Error("successfully created model from a setting with no name")
	}

	testSettings["name"] = 10

	err, _ = settingsMapToModel(testSettings, nil, nil, nil)

	if err == nil {
		t.Error("successfully created model from a setting with non string name")
	}

	testSettings["name"] = "wormhole"

	err, _ = settingsMapToModel(testSettings, nil, nil, nil)

	if err == nil {
		t.Error("successfully created model from nonexistent name")
	}

	testSettings["name"] = "oPeNaI"
	testSettings["api_key"] = "test1243"
	testSettings["model"] = "gpt-3.5-turbo-0125"

	err, _ = settingsMapToModel(testSettings, nil, nil, nil)

	if err != nil {
		t.Error("could not crate valid model")
	}
}

func TestExecute(t *testing.T) {
	messageTypeSlice := []string{
		"assistant",
		"standard",
		"standard",
	}

	setting := map[string]interface{}{
		"name": "opentet",
	}

	messageMapSlice := []map[string]interface{}{
		{
			"role":    "assistant",
			"content": "bruh",
		},
		{
			"role":    "user",
			"content": "huh?",
		},
		{
			"role":    "data",
			"content": 12,
		},
	}

	_, err := Execute(messageTypeSlice, messageMapSlice, []map[string]interface{}{setting}, nil, nil, nil, 4, nil)

	if err == nil {
		t.Fatal("invalid message was not caught")
	}

	validMapSlice := messageMapSlice[:len(messageMapSlice)-1]
	validMessageSlice := messageTypeSlice[:len(messageMapSlice)-1]

	_, err = Execute(validMessageSlice, validMapSlice, []map[string]interface{}{setting}, nil, nil, nil, 4, nil)

	var llmError *backoffError

	if !errors.As(err, &llmError) {
		t.Fatal(err)
	}
}
