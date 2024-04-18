package APIs

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type APIExecutor struct {
	// necessary info
}

func (e *APIExecutor) Init(timeout *int16) *APIExecutor {

	return nil
}

// AccessAPI connects to the OpenAI API using the provided API key and makes a request for text generation.
func (e *APIExecutor) AccessAPI(apiKey string) error {
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

type OpenAiApiAccess struct {
	ApiKey string `json:"api_key"`
}

func (o *OpenAiApiAccess) Validate() error {
	if o.ApiKey == "" {
		return errors.New("api_key must be set")
	}
	return nil
}

func (o *OpenAiApiAccess) doTask(a *APIExecutor) {
	//check input type
	err := a.AccessAPI(o.ApiKey)
	if err != nil {
		return
	}
}
