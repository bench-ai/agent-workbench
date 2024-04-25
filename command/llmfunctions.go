package command

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// LLM interface that implements request function
type LLM struct {
	//TODO
}

// Action defines an interface for actions.
type Action interface {
	Do(ctx context.Context) error
}

// FunctionAction defines an action based on a function.
type FunctionAction func(ctx context.Context) error

// Do execute the action.
func (f FunctionAction) Do(ctx context.Context) error {
	return f(ctx)
}

// Tasks is a list of actions.
type Tasks []Action

// ChatRequest Implementation is where the api actually gets called. This should have different cases for each type of model
func (l *LLM) ChatRequest(llm *LLM, chatRequest []string) (string, error) {
	//TODO
	return "", nil
}

// exponential backoff function, parameters are subject to change as this just takes in a string for chat request, but
// it should account for if the request is multimodal and has a description or if the request is just text
func exponentialBackoff(llms []*LLM, chatRequest []string, tryLimit int, waitLimit int) (string, error) {

	for _, l := range llms {
		t := 2 * time.Second // initial sleep duration
		for x := 0; x < tryLimit; x++ {
			responseChan := make(chan string)
			errChan := make(chan error)

			go func() {
				response, err := l.ChatRequest(l, chatRequest)
				if err != nil {
					errChan <- err
				} else {
					responseChan <- response
				}
			}()

			select {
			case response := <-responseChan:
				return response, nil
			case _, ok := <-errChan:
				if !ok {
					continue
				}
				if x == tryLimit-1 {
					return "", errors.New("all attempts failed")
				}

			case <-time.After(time.Duration(int64(waitLimit)) * time.Second): // Break if time exceeds waitLimit minutes
				break
			}

			time.Sleep(t)
			t = time.Duration(int64(t) * int64(t)) // exponential backoff
		}
	}
	return "", errors.New("all attempts failed")
}

// Do execute all the tasks in the list.
func (tasks Tasks) Do(ctx context.Context) error {
	for _, task := range tasks {
		if err := task.Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

// GptForTextAlternatives dummy logic.
func GptForTextAlternatives(text string, prompt string, memory string) func(ctx context.Context) error {
	return func(ctx context.Context) error {

		fmt.Printf("Executing function GptForTextAlternatives with parameters: %s, %s\n", text, prompt)
		//simulated error
		//if the string is an error return error
		return nil
	}
}

// GptForImage dummy logic.
func GptForImage(image string, description string, prompt string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		fmt.Printf("Executing function GptForImage with parameter: %s, %s, %s\n", image, description, prompt)
		return nil
	}
}

// GptForAudio dummy logic.
func GptForAudio(audio string, description string, prompt string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		fmt.Printf("Executing function GptForImage with parameter: %s, %s, %s\n", audio, description, prompt)
		return nil
	}
}

// GptForVideo dummy logic.
func GptForVideo(video string, description string, prompt string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		fmt.Printf("Executing function GptForImage with parameter: %s, %s, %s\n", video, description, prompt)
		return nil
	}
}
