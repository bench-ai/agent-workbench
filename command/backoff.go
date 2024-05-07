package command

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"
)

// LLM interface that implements request function
type LLM interface {
	Validate(messageSlice []MessageInterface) error
	Request(messages []MessageInterface, ctx context.Context) (error, *ChatCompletion)
}

type LLMError struct {
	mode    string
	Message string
}

func (l *LLMError) Error() string {
	return fmt.Sprintf("%s: %s. \n", l.mode, l.Message)
}

func (l *LLMError) Mode() string {
	return l.mode
}

func GetRateLimitError(message string) LLMError {
	return LLMError{
		mode:    "rate-limit",
		Message: message,
	}
}

func GetStandardError(message string) LLMError {
	return LLMError{
		mode:    "standard",
		Message: message,
	}
}

type BackoffError struct{}

func (b *BackoffError) Error() string {
	return "all backoff attempts failed"
}

// ExponentialBackoff
/*
Executes an Api Request to the LLM, switches the LLM based on execution time and availability
*/
func ExponentialBackoff(
	llmSlice []LLM,
	chatRequest *[]MessageInterface,
	tryLimit int16,
	requestWaitTime *int16) (*ChatCompletion, error) {

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

				var llmError *LLMError

				if errors.As(result.err, &llmError) {
					if (llmError.Mode() == "rate-limit") && (i != (tryLimit - 1)) {
						log.Println("rate limit hit, sleeping...")
						time.Sleep(time.Second * time.Duration(math.Pow(2.0, exp)))
						exp++
					} else if llmError.Mode() == "standard" {
						return nil, errors.New(llmError.Message)
					}
				} else {
					return nil, errors.New("llm does not return LLMError")
				}
			}
		}
	}

	return nil, &BackoffError{}
}
