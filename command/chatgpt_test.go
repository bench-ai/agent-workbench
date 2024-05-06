package command

import (
	"testing"
)

/**
TODO: Add unit test to helpers
*/

func TestValidateResponseFormat(t *testing.T) {

	failTable := []map[string]string{
		{
			"type":  "text",
			"other": "json_object",
		},
		{
			"type": "json_obje",
		},
		{
			"other": "json_obje",
		},
	}

	passTable := []map[string]string{
		{
			"type": "text",
		},
		{
			"type": "json_object",
		},
	}

	engine := getEngineMap()["gpt-3.5-turbo-1106"]

	for _, m := range failTable {
		if err := validateResponseFormat(m, engine); err == nil {
			t.Error("failed to detect invalid response format")
		}
	}

	for _, m := range passTable {
		if err := validateResponseFormat(m, engine); err != nil {
			t.Error("failed to detect valid response format")
		}
	}

	engine = getEngineMap()["gpt-3.5-turbo"]

	if err := validateResponseFormat(passTable[1], engine); err == nil {
		t.Error("response format failed to detect invalid engine")
	}
}

func TestValidateTools(t *testing.T) {

	func1 := ToolFunction{
		Name:       "test",
		Parameters: map[string]interface{}{"test": "test"},
	}

	t1 := Tool{
		Type:     "function",
		Function: func1,
	}

	engine := getEngineMap()["gpt-3.5-turbo-1106"]
	if err := t1.validateTools(engine); err != nil {
		t.Error("rejected function that was not invalid")
	}

	t1.Type = "f"

	if err := t1.validateTools(engine); err == nil {
		t.Error("failed to reject invalid tool type")
	}

	engine = getEngineMap()["gpt-3.5-turbo"]

	if err := t1.validateTools(engine); err == nil {
		t.Error("failed to reject invalid engine that has no function calling capabilities")
	}

}

func TestValidateMessages(t *testing.T) {

	prompt := "this is a test"
	content := GPTMultiModalContent{
		Type: "text",
		Text: &prompt,
	}

	m1 := GPTMultiModalCompliantMessage{
		Role: "user",
		Content: []GPTMultiModalContent{
			content,
		},
	}

	m3 := GptAssistantMessage{
		Content: &prompt,
		Role:    "assistant",
	}

	m4 := GPTStandardMessage{
		Content: prompt,
		Role:    "system",
	}

	mList := []MessageInterface{
		&m1,
		&m3,
		&m4,
	}

	engine := getEngineMap()["gpt-3.5-turbo-0125"]

	if err := validateMessages(engine, mList); err == nil {
		t.Error("failed to enforce multimodal requirements")
	}

	engine = getEngineMap()["gpt-4-turbo-2024-04-09"]

	if err := validateMessages(engine, mList); err != nil {
		t.Errorf("rejected proper engine, %v", err)
	}
}

func TestValidateToolChoice(t *testing.T) {

	toolChoicesStringPass := "auto"

	toolChoicesStringFail := "other"

	passData := map[string]interface{}{
		"type": "function",
		"function": map[string]string{
			"name": "my_func",
		},
	}

	failData := map[string]interface{}{
		"type": "function",
		"function": map[string]string{
			"nm": "my_func",
		},
	}

	if err := validateToolChoice(toolChoicesStringPass); err != nil {
		t.Errorf("rejected valid tool choice string, %v", err)
	}

	if err := validateToolChoice(toolChoicesStringFail); err == nil {
		t.Errorf("accpeted invalid tool choice, %v", err)
	}

	if err := validateToolChoice(passData); err != nil {
		t.Errorf("rejected valid tool choice string, %v", err)
	}

	if err := validateToolChoice(failData); err == nil {
		t.Errorf("accpeted invalid tool choice, %v", err)
	}
}
