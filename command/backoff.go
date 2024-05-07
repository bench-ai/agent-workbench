package command

import (
	"errors"
	"math"
	"time"
)

// LLM interface that implements request function
type LLM interface {
	Validate(messageSlice []MessageInterface) error
	Request(messages []MessageInterface, waitTime *int16) (*GptRequestError, *ChatCompletion)
}

// ExponentialBackoff
/*
Executes an Api Request to the LLM, switches the LLM based on execution time and availability
*/

type BackoffError struct{}

func (b *BackoffError) Error() string {
	return "all backoff attempts failed"
}

func ExponentialBackoff(
	llmSlice []LLM,
	chatRequest *[]MessageInterface,
	tryLimit int16,
	requestWaitTime *int16) (*ChatCompletion, error) {

	for _, llm := range llmSlice {
		exp := 2.0 // initial sleep duration

		for i := range tryLimit {
			chatErr, chatCompletion := llm.Request(*chatRequest, requestWaitTime)

			if chatErr != nil {
				switch chatErr.StatusCode {
				case 429:
					if i != (tryLimit - 1) {
						// Too many requests perform exponential backoff
						time.Sleep(time.Second * time.Duration(math.Pow(2.0, exp)))
						exp++
					}
				case 1:
					// Exceeded request time limit
					break
				default:
					// Standard Error return instantly
					return nil, errors.New(chatErr.Error.Message)
				}
			}

			if chatCompletion != nil {
				return chatCompletion, nil
			}
		}
	}

	return nil, &BackoffError{}
}
