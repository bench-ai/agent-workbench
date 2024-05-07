package command

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testRequest struct {
	Mode string
}

func (t *testRequest) Validate(messageSlice []MessageInterface) error {
	_ = messageSlice
	return nil
}

func (t *testRequest) Request(messages []MessageInterface, ctx context.Context) (error, *ChatCompletion) {

	_ = messages
	_ = ctx

	rateErr := GetRateLimitError("exceeded rate")
	stanErr := GetStandardError("generic error")

	switch t.Mode {
	case rateErr.Mode():
		return &rateErr, nil
	case stanErr.Mode():
		return &stanErr, nil
	case "wait":
		time.Sleep(4 * time.Second)
		return nil, &ChatCompletion{}
	default:
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

	tr := testRequest{"rate-limit"}

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

	tr.Mode = "wait"

	err, _ = timestamp(messageSlice, llms)

	if !errors.Is(err, &BackoffError{}) {
		t.Error("did not receive backoff error for timeout")
	}

	tr.Mode = "standard"

	err, _ = timestamp(messageSlice, llms)

	if err == nil {
		t.Error("request did not fail on improper request")
	}

	tr.Mode = "pass"

	err, _ = timestamp(messageSlice, llms)

	if err != nil {
		t.Error("failed on successful request")
	}
}
