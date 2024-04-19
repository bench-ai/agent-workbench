package apifunctions

import (
	"context"
	"fmt"
)

// Action defines an interface for actions.
type Action interface {
	Do(ctx context.Context)
}

// FunctionAction defines an action based on a function.
type FunctionAction func(ctx context.Context)

// Do executes the action.
func (f FunctionAction) Do(ctx context.Context) {
	f(ctx)
}

// AccessAPI dummy logic.
// this function logic should be replaced with the AccessAPI logic that is commented out in the apicalls.go file
// and should be followed by the StartSession function as well as all associated structs and interfaces
func AccessAPI(s string) func(ctx context.Context) {
	return func(ctx context.Context) {
		fmt.Printf("Executing function AccessAPI with parameter: %s\n", s)
	}
}

// GptForTextAlternatives dummy logic.
func GptForTextAlternatives(s string) func(ctx context.Context) {
	return func(ctx context.Context) {
		fmt.Printf("Executing function GptForTextAlternatives with parameter: %s\n", s)
	}
}

// GptForCodeParsing dummy logic.
func GptForCodeParsing(s string) func(ctx context.Context) {
	return func(ctx context.Context) {
		fmt.Printf("Executing function GptForCodeParsing with parameter: %s\n", s)
	}
}

// GptForImage dummy logic.
func GptForImage(s string) func(ctx context.Context) {
	return func(ctx context.Context) {
		fmt.Printf("Executing function GptForImage with parameter: %s\n", s)
	}
}

// GptForWebpageAnalysis dummy logic.
func GptForWebpageAnalysis(s string) func(ctx context.Context) {
	return func(ctx context.Context) {
		fmt.Printf("Executing function GptForWebpageAnalysis with parameter: %s\n", s)
	}
}

// Tasks is a list of actions.
type Tasks []Action

// Do executes all the tasks in the list.
func (tasks Tasks) Do(ctx context.Context) {
	for _, task := range tasks {
		task.Do(ctx)
	}
}
