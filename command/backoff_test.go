package command

import (
	"errors"
	"testing"
	"time"
)

type testRequest struct {
	Mode int
}

func (t *testRequest) Validate(messageSlice []MessageInterface) error {
	_ = messageSlice
	return nil
}

func (t *testRequest) Request(messages []MessageInterface, waitTime *int16) (*GptRequestError, *ChatCompletion) {

	_ = messages
	_ = waitTime

	if t.Mode == 429 {
		return &GptRequestError{
			StatusCode: 429,
		}, nil
	} else if t.Mode == 1 {
		time.Sleep(time.Second * 2)
		return &GptRequestError{
			StatusCode: 1,
		}, nil
	} else if t.Mode == 0 {
		return &GptRequestError{
			StatusCode: 0,
			Error: GptError{
				Message: "test error",
			},
		}, nil
	} else {
		return nil, &ChatCompletion{}
	}
}

func timestamp(m []MessageInterface, llms []LLM) (error, int64) {

	wt := int16(1)
	start := time.Now()
	_, err := ExponentialBackoff(llms, &m, 2, &wt)
	end := time.Now()

	return err, end.Sub(start).Milliseconds()
}

func TestExponentialBackoff(t *testing.T) {

	tr := testRequest{429}

	message := GPTStandardMessage{
		Role:    "user",
		Content: "test",
	}

	messageSlice := []MessageInterface{
		&message,
	}

	llms := []LLM{&tr}

	err, ts := timestamp(messageSlice, llms)

	if ts < 4000 {
		t.Error("failed to sleep for a minimum of 4s after hitting 409")
	}

	if !errors.Is(err, &BackoffError{}) {
		t.Error("failed instead of sleeping for a 409 request")
	}

	tr.Mode = 1

	err, ts = timestamp(messageSlice, llms)

	if ts < 4000 {
		t.Error("failed to sleep for a minimum of 4s after failing with status 1")
	}

	if !errors.Is(err, &BackoffError{}) {
		t.Error("failed instead of sleeping for a 409 request")
	}

	tr.Mode = 0

	err, _ = timestamp(messageSlice, llms)

	if err == nil {
		t.Error("request did not fail on improper request")
	}

	tr.Mode = 108

	err, _ = timestamp(messageSlice, llms)

	if err != nil {
		t.Error("failed on successful request")
	}
}
