package APIs

import (
	"context"
	"errors"
	//"log"
	"agent/apifunctions"
	"time"
)

/////////////////// equivalent to tools.go ///////////////////

type APIExecutor struct {
	// necessary info
	ctx     context.Context
	cancel  context.CancelFunc
	llmName string
	//the following lines are currently commented out in case the new implementation with apifunctions.go works
	//tasks   []interface{} //list of function calls. Currently implemented as a list if interfaces.
	// Once the implementation of all tasks is complete, it will be implemented as a list of functions
	Tasks apifunctions.Tasks
}

func (a *APIExecutor) Init(max_token *int16, timeout *int16) *APIExecutor {
	a.ctx = context.Background()
	if timeout != nil {
		a.ctx, a.cancel = context.WithTimeout(a.ctx, time.Duration(*timeout)*time.Second)
	}
	return a
}

func (a *APIExecutor) appendTask(action apifunctions.Action) {
	a.Tasks = append(a.Tasks, action)
}

func (a *APIExecutor) AccessAPI(apiKey string) {
	a.Tasks = append(a.Tasks, apifunctions.FunctionAction(apifunctions.AccessAPI(apiKey)))
}
func (a *APIExecutor) gptForTextAlternatives(s string) {
	a.Tasks = append(a.Tasks, apifunctions.FunctionAction(apifunctions.GptForTextAlternatives(s)))
}

func (a *APIExecutor) gptForCodeParsing(s string) {
	a.Tasks = append(a.Tasks, apifunctions.FunctionAction(apifunctions.GptForCodeParsing(s)))
}

func (a *APIExecutor) gptForImage(s string) {
	a.Tasks = append(a.Tasks, apifunctions.FunctionAction(apifunctions.GptForImage(s)))
}

func (a *APIExecutor) gptForWebpageAnalysis(s string) {
	a.Tasks = append(a.Tasks, apifunctions.FunctionAction(apifunctions.GptForWebpageAnalysis(s)))
}

//AccessAPI logic
/*
type apiKeyTransport struct {
	apiKey string
}

func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	return http.DefaultTransport.RoundTrip(req)
}

// AccessAPI connects to the OpenAI API using the provided API key and makes a request for text generation.
func (a *APIExecutor) AccessAPI(apiKey string) *http.Client {
	client := &http.Client{
		Transport: &apiKeyTransport{apiKey: apiKey},
	}
	return client
}

type SessionResponse struct {
	ID string `json:"id"`
}

// StartSession is a function that starts a session with the client and takes in the desired engine as a parameter.
// This function should be called by the function that use gpt
func StartSession(client *http.Client, engine string) (string, error) {
	ctx := context.Background()
	url := "https://api.openai.com/v1/sessions"

	// Create request body
	requestBody := map[string]string{"engine": engine}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Send POST request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse response
	var sessionResp SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return "", err
	}

	return sessionResp.ID, nil
}
*/
//functions to execute commands

/*
	func (a *APIExecutor) gptForTextAlternatives(s string) (string, error) {
		fmt.Print(s)
		return s, nil
	}

	func (a *APIExecutor) gptForCodeParsing(s string) (string, error) {
		fmt.Print(s)
		return s, nil
	}

	func (a *APIExecutor) gptForImage(s string) (string, error) {
		fmt.Print(s)
		return s, nil
	}

	func (a *APIExecutor) gptForWebpageAnalysis(s string) (string, error) {
		fmt.Print(s)
		return s, nil
	}
*/
func (a *APIExecutor) Execute() {
	defer a.cancel()
	//figure out how to do error handling
	a.Tasks.Do(a.ctx)
}

///////////////////equivalent to browser.go///////////////////

/* commands to implement are
AccessLLM
gptForTextAlternatives
gptForCodeParsing
gptForImage
gptForWebpageAnalysis*/

type Parameters interface {
	Validate() error
}

type ApiParams interface {
	Parameters
	AppendTask(a *APIExecutor)
}

//accessLLM

type ApiAccess struct {
	ApiKey string `json:"api_key"`
}

func (o *ApiAccess) Validate() error {
	if o.ApiKey == "" {
		return errors.New("api_key must be set")
	}
	return nil
}

func (o *ApiAccess) AppendTask(a *APIExecutor) {
	a.AccessAPI(o.ApiKey)
}

// gptForTextAlternatives

type TextToFix struct {
	Text   string `json:"text"`
	Prompt string `json:"prompt"`
}

func (o *TextToFix) Validate() error {
	if o.Text == "" {
		return errors.New("text must be set")
	}
	if o.Prompt == "" {
		return errors.New("prompt must be set")
	}
	return nil
}

func (o *TextToFix) AppendTask(a *APIExecutor) {
	a.gptForTextAlternatives(o.Text)
}

//gptForCodeParsing

type CodeToCheck struct {
	Code   string `json:"code"`
	Prompt string `json:"prompt"`
}

func (o *CodeToCheck) Validate() error {
	if o.Code == "" {
		return errors.New("code must be set")
	}
	if o.Prompt == "" {
		return errors.New("prompt must be set")
	}
	return nil
}

func (o *CodeToCheck) AppendTask(a *APIExecutor) {
	a.gptForCodeParsing(o.Code)
}

//gptForImage

type ImageToCheck struct {
	Image  string `json:"image"`
	Prompt string `json:"prompt"`
}

func (o *ImageToCheck) Validate() error {
	if o.Image == "" {
		return errors.New("image must be set")
	}
	if o.Prompt == "" {
		return errors.New("prompt must be set")
	}
	return nil
}

func (o *ImageToCheck) AppendTask(a *APIExecutor) {
	a.gptForImage(o.Image)
}

//gptForWebpageAnalysis

type WebpageToCheck struct {
	Webpage string `json:"webpage"`
	Prompt  string `json:"prompt"`
}

func (o *WebpageToCheck) Validate() error {
	if o.Webpage == "" {
		return errors.New("webpage must be set")
	}
	if o.Prompt == "" {
		return errors.New("prompt must be set")
	}
	return nil
}

func (o *WebpageToCheck) AppendTask(a *APIExecutor) {
	a.gptForWebpageAnalysis(o.Webpage)
}
