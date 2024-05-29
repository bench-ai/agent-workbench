package llm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"
)

// model interface that implements request function
type model interface {
	Validate(messageSlice []messageInterface) error
	Request(messages []messageInterface, ctx context.Context) (error, *ChatCompletion)
}

type modelError struct {
	mode    string
	Message string
}

func (l *modelError) Error() string {
	return fmt.Sprintf("%s: %s. \n", l.mode, l.Message)
}

func (l *modelError) Mode() string {
	return l.mode
}

func getRateLimitError(message string) modelError {
	return modelError{
		mode:    "rate-limit",
		Message: message,
	}
}

func getStandardError(message string) modelError {
	return modelError{
		mode:    "standard",
		Message: message,
	}
}

type backoffError struct{}

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
				log.Println("failed to complete request in allocated duration")
				<-ch
			case result := <-ch:
				cancel()
				if result.err == nil {
					return result.comp, nil
				}

				var llmError *modelError

				if errors.As(result.err, &llmError) {
					if (llmError.Mode() == "rate-limit") && (i != (tryLimit - 1)) {
						log.Println("rate limit hit, sleeping...")
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
