package command

import (
	"context"
	"fmt"
)

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
func GptForTextAlternatives(text string, prompt string) func(ctx context.Context) error {
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
