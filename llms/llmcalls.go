package llms

import (
	"agent/command"
	"context"
	"errors"

	"time"
)

/////////////////// equivalent to tools.go ///////////////////

type APIExecutor struct {
	// necessary info
	ctx     context.Context
	cancel  context.CancelFunc
	llmName string
	tasks   command.Tasks
}

func (a *APIExecutor) Init(max_token *int16, timeout *int16) *APIExecutor {
	a.ctx = context.Background()
	if timeout != nil {
		a.ctx, a.cancel = context.WithTimeout(a.ctx, time.Duration(*timeout)*time.Second)
	}
	return a
}

func (a *APIExecutor) gptForTextAlternatives(text string, prompt string, memory string) {
	a.tasks = append(a.tasks, command.FunctionAction(command.GptForTextAlternatives(text, prompt, memory)))
}
func (a *APIExecutor) gptForImage(image string, description string, prompt string) {
	a.tasks = append(a.tasks, command.FunctionAction(command.GptForImage(image, description, prompt)))
}

func (a *APIExecutor) gptForAudio(audio string, description string, prompt string) {
	a.tasks = append(a.tasks, command.FunctionAction(command.GptForAudio(audio, description, prompt)))
}

func (a *APIExecutor) gptForVideo(video string, description string, prompt string) {
	a.tasks = append(a.tasks, command.FunctionAction(command.GptForVideo(video, description, prompt)))
}

func (a *APIExecutor) Execute() error {
	defer a.cancel()
	if err := a.tasks.Do(a.ctx); err != nil {
		return err
	}
	return nil
}

///////////////////equivalent to browser.go///////////////////

type Parameters interface {
	Validate() error
}

type ApiParams interface {
	Parameters
	AppendTask(a *APIExecutor)
}

type TextToAnalyze struct {
	Text   string `json:"text"`
	Prompt string `json:"prompt"`
	Memory string `json:"memory"`
}

func (o *TextToAnalyze) Validate() error {
	if o.Text == "" {
		return errors.New("text must be set")
	}
	if o.Prompt == "" {
		return errors.New("prompt must be set")
	}
	return nil
}

func (o *TextToAnalyze) AppendTask(a *APIExecutor) {
	a.gptForTextAlternatives(o.Text, o.Prompt, o.Memory)
}

//gptForImage

type ImageToCheck struct {
	Image       string `json:"media_source"` //url to image or base64 encoding so still of type string
	Description string `json:"description"`  //not a mandatory entry
	Prompt      string `json:"prompt"`
}

func (o *ImageToCheck) Validate() error {
	if o.Image == "" {
		return errors.New("image source must be provided")
	}
	if o.Prompt == "" {
		return errors.New("prompt must be set")
	}
	//since description is not a mandatory entry, there is no check for it in this function
	return nil
}

func (o *ImageToCheck) AppendTask(a *APIExecutor) {
	a.gptForImage(o.Image, o.Description, o.Prompt)
}

//gptForAudio

type AudioToCheck struct {
	Audio       string `json:"media_source"` //url to audio so still of type string
	Description string `json:"description"`  //not a mandatory entry
	Prompt      string `json:"prompt"`
}

func (o *AudioToCheck) Validate() error {
	if o.Audio == "" {
		return errors.New("audio source must be provided")
	}
	if o.Prompt == "" {
		return errors.New("prompt must be set")
	}
	//since description is not a mandatory entry, there is no check for it in this function
	return nil
}

func (o *AudioToCheck) AppendTask(a *APIExecutor) {
	a.gptForAudio(o.Audio, o.Description, o.Prompt)
}

//gptForVideo

type VideoToCheck struct {
	Video       string `json:"media_source"` //url to video so still of type string
	Description string `json:"description"`  //not a mandatory entry
	Prompt      string `json:"prompt"`
}

func (o *VideoToCheck) Validate() error {
	if o.Video == "" {
		return errors.New("video source must be provided")
	}
	if o.Prompt == "" {
		return errors.New("prompt must be set")
	}
	//since description is not a mandatory entry, there is no check for it in this function
	return nil
}

func (o *VideoToCheck) AppendTask(a *APIExecutor) {
	a.gptForVideo(o.Video, o.Description, o.Prompt)
}
