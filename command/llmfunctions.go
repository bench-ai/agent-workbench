package command

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

// LLM interface that implements request function
type LLM interface {
	Validate(messageSlice []MessageInterface) error
	Request(messages []MessageInterface, ctx context.Context) (*GptRequestError, *ChatCompletion)
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

// exponential backoff function, parameters are subject to change as this just takes in a string for chat request, but
// it should account for if the request is multimodal and has a description or if the request is just text
func ExponentialBackoff(
	llmSlice []LLM,
	chatRequest *[]MessageInterface,
	tryLimit int16,
	requestWaitTime *int16) (*ChatCompletion, error) {

	for _, llm := range llmSlice {
		coolDown := 2.0 // initial sleep duration

		for range tryLimit {
			responseChan := make(chan *ChatCompletion)
			errChan := make(chan *GptRequestError)

			var ctx context.Context
			var cancel context.CancelFunc

			if requestWaitTime != nil {
				ctx, cancel = context.WithTimeout(context.Background(), time.Second*time.Duration(*requestWaitTime))
			} else {
				ctx, cancel = context.WithCancel(context.Background())
			}

			go func() {
				defer cancel()
				chatErr, chatCompletion := llm.Request(*chatRequest, ctx)

				fmt.Println(chatErr)

				if chatErr != nil {
					errChan <- chatErr
				}

				if chatCompletion != nil {
					responseChan <- chatCompletion
				}
			}()

			select {
			case <-ctx.Done():
				break
			case err := <-errChan:
				if err.StatusCode == 429 {
					time.Sleep(time.Second * time.Duration(math.Pow(coolDown, 2.0)))
					coolDown++
				} else {
					break
				}
			case comp := <-responseChan:
				return comp, nil
			}
		}
	}
	return nil, errors.New("all attempts failed")
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
