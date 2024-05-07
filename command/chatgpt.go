package command

import (
	"agent/helper"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type MessageInterface interface {
	ValidateRole() bool
	GetType() string
	GetRole() string
}

func containsRole(roleSlice []string, role string) bool {
	contains := helper.Contains[string]
	return contains(roleSlice, role)
}

type Engine struct {
	ContextWindow   uint32
	HasJsonMode     bool
	Name            string
	Multimodal      bool
	FunctionCalling bool
}

func getEngineMap() map[string]Engine {
	gpt30125 := Engine{16385, false, "gpt-3.5-turbo-0125", false, false}
	gpt3turbo := Engine{16385, false, "gpt-3.5-turbo", false, false}
	gpt31106 := Engine{16385, true, "gpt-3.5-turbo-1106", false, true}

	gpt4lat := Engine{128000, true, "gpt-4-turbo-2024-04-09", true, true}
	gpt40125 := Engine{128000, false, "gpt-4-0125-preview", false, false}
	gpt41106 := Engine{128000, true, "gpt-4-1106-preview", false, true}

	return map[string]Engine{
		gpt30125.Name:  gpt30125,
		gpt3turbo.Name: gpt3turbo,
		gpt31106.Name:  gpt31106,
		gpt4lat.Name:   gpt4lat,
		gpt40125.Name:  gpt40125,
		gpt41106.Name:  gpt41106,
	}
}

func getEngineOptionList() string {

	engineString := ""
	for k := range getEngineMap() {
		engineString += k + ", "
	}

	return engineString[:len(engineString)-2]
}

type GPTStandardMessage struct {
	Role    string  `json:"role"`
	Content string  `json:"content"`
	Name    *string `json:"name,omitempty"`
}

func (g *GPTStandardMessage) ValidateRole() bool {
	return containsRole([]string{"system", "user"}, g.Role)
}

func (g *GPTStandardMessage) GetType() string {
	return "standard"
}

func (g *GPTStandardMessage) GetRole() string {
	return g.Role
}

type ImageUrl struct {
	Url    string  `json:"url"`
	Detail *string `json:"detail,omitempty"`
}

type GPTMultiModalContent struct {
	Type     string    `json:"type"`
	Text     *string   `json:"text,omitempty"`
	ImageUrl *ImageUrl `json:"image_url,omitempty"`
}

type GPTMultiModalCompliantMessage struct {
	Role    string                 `json:"role"`
	Content []GPTMultiModalContent `json:"content"`
	Name    *string                `json:"name,omitempty"`
}

func (g *GPTMultiModalCompliantMessage) ValidateRole() bool {
	return containsRole([]string{"user"}, g.Role)
}

func (g *GPTMultiModalCompliantMessage) GetType() string {
	return "multimodal"
}

func (g *GPTMultiModalCompliantMessage) GetRole() string {
	return g.Role
}

type ToolCall struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type GptAssistantMessage struct {
	Content   *string     `json:"content,omitempty"`
	Role      string      `json:"role,omitempty"`
	Name      *string     `json:"name,omitempty"`
	ToolCalls *[]ToolCall `json:"tool_calls,omitempty"`
}

func (g *GptAssistantMessage) ValidateRole() bool {
	return containsRole([]string{"assistant"}, g.Role)
}

func (g *GptAssistantMessage) GetType() string {
	return "assistant"
}

func (g *GptAssistantMessage) GetRole() string {
	return g.Role
}

type GptToolMessage struct {
	Role       string `json:"role"`
	Content    string `json:"Content"`
	ToolCallId string `json:"tool_call_id"`
}

func (g *GptToolMessage) ValidateRole() bool {
	return containsRole([]string{"tool"}, g.Role)
}

func (g *GptToolMessage) GetType() string {
	return "tool"
}

func (g *GptToolMessage) GetRole() string {
	return g.Role
}

type ToolFunction struct {
	Description *string                `json:"description,omitempty"`
	Name        string                 `json:"name"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

func (t *Tool) validateTools(engine Engine) error {
	validSlice := []string{
		"function",
	}

	containsString := helper.Contains[string]

	if !containsString(validSlice, t.Type) {
		return fmt.Errorf("tool can be only type function, found %s", t.Type)
	}

	if t.Type == "function" && !engine.FunctionCalling {
		return errors.New("engine is not function call capable")
	}

	return nil
}

type ChatGPT struct {
	Model            string             `json:"model"`
	FrequencyPenalty *float32           `json:"frequency_penalty,omitempty"`
	Messages         []MessageInterface `json:"messages"`
	LogitBias        *map[string]int    `json:"logit_bias,omitempty"`
	LogProbs         *bool              `json:"log_probs,omitempty"`
	TopLogprobs      *uint8             `json:"top_logprobs,omitempty"`
	MaxTokens        *int               `json:"max_tokens,omitempty"`
	N                *int               `json:"n,omitempty"`
	PresencePenalty  *float32           `json:"presence_penalty,omitempty"`
	ResponseFormat   *map[string]string `json:"response_format,omitempty"`
	Seed             *int               `json:"seed,omitempty"`
	Stop             *interface{}       `json:"stop,omitempty"`
	Stream           *bool              `json:"stream,omitempty"`
	Temperature      *float32           `json:"temperature,omitempty"`
	TopP             *float32           `json:"top_p,omitempty"`
	Tools            *[]Tool            `json:"tools,omitempty"`
	ToolChoice       interface{}        `json:"tool_choice"`
	key              string
}

func InitChatgpt(model, key string, maxTokens *int, temperature *float32) *ChatGPT {
	c := ChatGPT{
		Model:       model,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		key:         key,
	}

	return &c
}

type Message struct {
	Content   *string    `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Role      string     `json:"role"`
}

type LogprobContent struct {
	Token   string   `json:"token"`
	Logprob int32    `json:"logprob"`
	Bytes   *[]int32 `json:"bytes"`
}

type FullLogprobContent struct {
	LogprobContent
	TopLogprobs []LogprobContent `json:"top_logprobs"`
}

type Choice struct {
	FinishReason string              `json:"finish_reason"`
	Index        int32               `json:"index"`
	Message      Message             `json:"message"`
	Logprobs     *FullLogprobContent `json:"logprobs,omitempty"`
}

type ChatCompletion struct {
	Id                string   `json:"id"`
	Created           int64    `json:"created"`
	Choices           []Choice `json:"choices"`
	Model             string   `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Object            string   `json:"object"`
	Usage             struct {
		PromptTokens     int32 `json:"prompt_tokens"`
		CompletionTokens int32 `json:"completion_tokens"`
		TotalTokens      int32 `json:"total_tokens"`
	} `json:"usage"`
}

type GptError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}

type gptRequestError struct {
	StatusCode int      `json:"statusCode"`
	Error      GptError `json:"error"`
}

func validateResponseFormat(responseFormat map[string]string, engine Engine) error {

	validSlice := []string{
		"json_object", "text",
	}

	containsString := helper.Contains[string]

	if len(responseFormat) > 1 {
		return fmt.Errorf("response format musto only have one key, detected %d keys", len(responseFormat))
	}

	var v string
	var k string

	for k, v = range responseFormat {
		if k != "type" {
			return fmt.Errorf("response format must only have one key called type, detected key named %s", k)
		}

		if !containsString(validSlice, v) {
			return fmt.Errorf("%s is not a valid type", v)
		}
	}

	if v == "json_object" {
		if !engine.HasJsonMode {
			return errors.New("engine is not json mode capable")
		}
	}

	return nil
}

func validateToolChoice(toolChoice interface{}) error {

	if val, ok := toolChoice.(string); ok {
		if !(val == "auto" || val == "none") {
			return fmt.Errorf("tool choice must be auto or none found %s", val)
		} else {
			return nil
		}
	}

	val, ok := toolChoice.(map[string]interface{})

	if !ok {
		return errors.New("tool choice must be either a string or map")
	}

	valType, ok := val["type"]

	if !ok {
		return errors.New("tool choice map missing type key")
	}

	typeString, ok := valType.(string)

	if !ok {
		return errors.New("type must be a string")
	}

	optSlice := []string{
		"function",
	}

	if !helper.Contains[string](optSlice, typeString) {
		return fmt.Errorf("tool choice type must be function found %s", typeString)
	}

	functionInterface, ok := val[typeString]

	if !ok {
		return errors.New("tool choice is missing function definition object")
	}

	functionDefintion, ok := functionInterface.(map[string]string)

	if !ok {
		return errors.New("definition must be of type object")
	}

	if _, ok = functionDefintion["name"]; !ok {
		return errors.New("missing function name")
	}

	return nil
}

func validateMessages(engine Engine, messageSlice []MessageInterface) error {

	for _, mess := range messageSlice {
		if !mess.ValidateRole() {
			return fmt.Errorf("MessageInterface of type %s does not accept the role you provided", mess.GetType())
		}

		if mess.GetType() == "multimodal" && !engine.Multimodal {
			return fmt.Errorf("MessageInterface of type %s is not multimodal friendly", mess.GetType())
		}

		if mess.GetType() == "tool" && !engine.FunctionCalling {
			return fmt.Errorf("MessageInterface of type %s is not function friendly", mess.GetType())
		}
	}

	return nil
}

func (c *ChatGPT) Validate(messageSlice []MessageInterface) error {
	floatHelper := helper.IsBetween[float32]
	intHelper := helper.IsBetween[uint8]

	if c.FrequencyPenalty != nil && !floatHelper(-2.0, 2.0, *c.FrequencyPenalty, true, true) {
		return fmt.Errorf("frequency penalty must be between -2.0 and 2.0 got %f", *c.FrequencyPenalty)
	}

	if c.TopLogprobs != nil && !intHelper(0, 20, *c.TopLogprobs, true, true) {
		return fmt.Errorf("top log probs must be between 0 and 20 got %d", *c.TopLogprobs)
	}

	if c.PresencePenalty != nil && !floatHelper(-2.0, 2.0, *c.PresencePenalty, true, true) {
		return fmt.Errorf("presence penalty must be between -2.0 and 2.0 got %f", *c.PresencePenalty)
	}

	if c.Temperature != nil && !floatHelper(0.0, 2.0, *c.Temperature, true, true) {
		return fmt.Errorf("temperature must be between 0.0 and 2.0 got %f", *c.Temperature)
	}

	if c.TopP != nil && !floatHelper(0.0, 1.0, *c.TopP, true, true) {
		return fmt.Errorf("top p must be between 0.0 and 1.0 got %f", *c.TopP)
	}

	var engine Engine
	var ok bool

	if engine, ok = getEngineMap()[c.Model]; !ok {
		return fmt.Errorf(
			"gpt has no integrated engine named %s, available options are %s",
			c.Model,
			getEngineOptionList())
	}

	if c.ResponseFormat != nil {
		err := validateResponseFormat(*c.ResponseFormat, engine)

		if err != nil {
			return err
		}
	}

	if c.Tools != nil {
		for _, i := range *c.Tools {
			if err := i.validateTools(engine); err != nil {
				return err
			}
		}
	}

	if err := validateMessages(engine, messageSlice); err != nil {
		return err
	}

	if c.ToolChoice != nil {
		if err := validateToolChoice(c.ToolChoice); err != nil {
			return err
		}
	}

	return nil
}

/**
TODO
Add method to estimate context window For multimodal and regular requests
*/

func (c *ChatGPT) Request(messages []MessageInterface, ctx context.Context) (error, *ChatCompletion) {
	lastMessage := messages[len(messages)-1]
	var resp gptRequestError

	if lastMessage.GetRole() != "user" {
		err := GetStandardError("last message is not a user message")
		return &err, nil
	}

	if err := c.Validate(messages); err != nil {
		standardErr := GetStandardError(err.Error())
		return &standardErr, nil
	}

	url := "https://api.openai.com/v1/chat/completions"
	c.Messages = messages

	defer func() {
		c.Messages = []MessageInterface{}
	}()

	jsonBytes, err := json.Marshal(c)

	if err != nil {
		standardErr := GetStandardError(err.Error())
		return &standardErr, nil
	}

	var client http.Client

	reader := bytes.NewReader(jsonBytes)
	pRequest, err := http.NewRequestWithContext(ctx, "POST", url, reader)

	if err != nil {
		standardErr := GetStandardError(err.Error())
		return &standardErr, nil
	}

	pRequest.Header.Set("Content-Type", "application/json")
	pRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.key))

	pResponse, err := client.Do(pRequest)

	if err != nil {
		standardErr := GetStandardError(err.Error())
		return &standardErr, nil
	}

	defer func() {
		closeErr := pResponse.Body.Close()
		if closeErr != nil {
			log.Fatal(closeErr)
		}
	}()

	responseBytes, err := io.ReadAll(pResponse.Body)

	if err != nil {
		standardErr := GetStandardError(err.Error())
		return &standardErr, nil
	}

	if pResponse.StatusCode == 200 {
		var gptResp ChatCompletion

		if err = json.Unmarshal(responseBytes, &gptResp); err != nil {
			standardErr := GetStandardError(err.Error())
			return &standardErr, nil
		}

		return nil, &gptResp
	} else {
		if err = json.Unmarshal(responseBytes, &resp); err != nil {
			standardErr := GetStandardError(err.Error())
			return &standardErr, nil
		}

		if pResponse.StatusCode == 429 {
			rateLimitErr := GetRateLimitError(resp.Error.Message)
			return &rateLimitErr, nil
		}

		standardErr := GetStandardError(resp.Error.Message)
		return &standardErr, nil
	}
}

func ConvertChatCompletion(completion *ChatCompletion) *GptAssistantMessage {
	var message GptAssistantMessage
	message.Role = completion.Choices[0].Message.Role

	var choiceMessage Choice

	index := int32(0)
	for _, ch := range completion.Choices {
		if ch.Index >= index {
			choiceMessage = ch
		}
	}

	if choiceMessage.FinishReason == "tool_calls" {
		message.ToolCalls = &choiceMessage.Message.ToolCalls
	} else {
		message.Content = choiceMessage.Message.Content
	}

	return &message
}
