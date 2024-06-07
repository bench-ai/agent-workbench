package llm

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

// model
/*
interface that implements request function
*/
type model interface {
	Validate(messageSlice []messageInterface) error
	Request(messages []messageInterface, ctx context.Context) (error, *ChatCompletion)
}

// modelError
/*
error class representing an error produced in the chat completion generation process
*/
type modelError struct {
	mode    string
	Message string
}

/*
Error

the error as a string
*/
func (l *modelError) Error() string {
	return fmt.Sprintf("%s: %s. \n", l.mode, l.Message)
}

/*
Mode

the reason behind the error
*/
func (l *modelError) Mode() string {
	return l.mode
}

/*
getRateLimitError

get a modelError object with mode rate-limit
*/
func getRateLimitError(message string) modelError {
	return modelError{
		mode:    "rate-limit",
		Message: message,
	}
}

/*
getStandardError

get a modelError object with mode standard
*/
func getStandardError(message string) modelError {
	return modelError{
		mode:    "standard",
		Message: message,
	}
}

/*
backoffError

Error produced when all backoff attempts are exhausted
*/
type backoffError struct{}

/*
Error

the error as a string
*/
func (b *backoffError) Error() string {
	return "all backoff attempts failed"
}

// exponentialBackoff
/*
Executes an Api Request to the model, switches the model based on execution time and availability
*/
func exponentialBackoff(
	llmSlice []model,
	chatRequest *[]messageInterface,
	tryLimit int16,
	requestWaitTime *int32) (*ChatCompletion, error) {

	/*
		how to calculate the max time the exponential backoff function will run for:
		max_time = (len(llmSlice) * requestWaitTime * tryLimit) + ((2^n for n in range(0, tryLimit-1)) * len(llmSlice))
	*/

	type request struct {
		comp *ChatCompletion
		err  error
	}

	for _, llm := range llmSlice {
		exp := 2.0 // initial sleep duration

		for i := range tryLimit {

			ctxTimeout, cancel := context.WithTimeout(
				context.Background(),
				time.Millisecond*time.Duration(*requestWaitTime))

			ch := make(chan request)

			go func() {
				chatErr, chatCompletion := llm.Request(*chatRequest, ctxTimeout)
				ch <- request{
					chatCompletion,
					chatErr,
				}
			}()

			select {
			case <-ctxTimeout.Done():
				cancel()
				<-ch
			case result := <-ch:
				cancel()
				if result.err == nil {
					return result.comp, nil
				}

				var llmError *modelError

				if errors.As(result.err, &llmError) {
					if (llmError.Mode() == "rate-limit") && (i != (tryLimit - 1)) {
						time.Sleep(time.Second * time.Duration(math.Pow(2.0, exp)))
						exp++
					} else if llmError.Mode() == "standard" {
						return nil, errors.New(llmError.Message)
					}
				} else {
					return nil, errors.New("llm does not return modelError")
				}
			}
		}
	}

	return nil, &backoffError{}
}
