package APIs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	//"log"
	"net/http"
	"time"
)

/////////////////// equivalent to tools.go ///////////////////

type APIExecutor struct {
	// necessary info
	ctx     context.Context
	cancel  context.CancelFunc
	llmName string
	task    []interface{} //tasks will be of different types so it is stored as a list of interfaces

}

func (a *APIExecutor) Init(max_token *int16, timeout *int16) *APIExecutor {
	if timeout != nil {
		a.ctx, a.cancel = context.WithTimeout(a.ctx, time.Duration(*timeout)*time.Second)
	}
	return a
}

// AccessAPI connects to the OpenAI API using the provided API key and makes a request for text generation.
func (a *APIExecutor) AccessAPI(apiKey string) error {
	// Prepare request body
	//Example code assuming text generation request
	requestBody := map[string]interface{}{
		"prompt":     "some prompt",
		"max_tokens": 500,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/engines/davinci/completions", bytes.NewBuffer(requestJSON))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("API call failed with status: " + resp.Status)
	}

	// Handle API response

	return nil
}

func (a *APIExecutor) Execute() {
	defer a.cancel()
}

func (a *APIExecutor) appendTask(action string) {
	a.task = append(a.task, action)
}

///////////////////equivalent to browser.go///////////////////

/* commands to implement are
accessLLM
gptForTextAlternative
gptForCodeParsing
gptForImage
gptForWebpageAnalysis
sleep*/

type Parameters interface {
	Validate() error
}

type ApiParams interface {
	Parameters
	AppendTask(a *APIExecutor)
}

//accessLLM

type OpenAiApiAccess struct {
	ApiKey string `json:"api_key"`
}

func (o *OpenAiApiAccess) Validate() error {
	if o.ApiKey == "" {
		return errors.New("api_key must be set")
	}
	return nil
}

func (o *OpenAiApiAccess) AppendTask(a *APIExecutor) {
	err := a.AccessAPI(o.ApiKey)
	if err != nil {
		return
	}
}

//gptForTextAlternatives
//gptForTextAlternative
//gptForCodeParsing
//gptForImage
//gptForWebpageAnalysis
//sleep
