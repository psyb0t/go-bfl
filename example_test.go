package bfl_test

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/psyb0t/go-bfl"
)

// This README example submits a paid task, so it only compiles during tests.
//
//nolint:testableexamples // needs a real API key and submits a paid task
func Example() {
	const (
		requestTimeout = 30 * time.Second
		taskTimeout    = 2 * time.Minute
		pollInterval   = time.Second
	)

	auth := bfl.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set("x-key", os.Getenv("BFL_API_KEY"))

			return nil
		},
	)

	client, err := bfl.NewClientWithResponses(
		"https://api.bfl.ai",
		auth,
		bfl.WithHTTPClient(&http.Client{Timeout: requestTimeout}),
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
	defer cancel()

	prompt := "A small cabin beside a lake"

	submitted, err := client.FluxPro11V1FluxPro11PostWithResponse(
		ctx,
		bfl.FluxPro11Inputs{Prompt: &prompt},
	)
	if err != nil {
		panic(err)
	}

	if submitted.JSON200 == nil {
		panic(submitted.Status())
	}

	task, err := submitted.JSON200.AsAsyncResponse()
	if err != nil {
		panic(err)
	}

	if task.Id == "" || task.PollingUrl == "" {
		panic("submission did not return a polling task")
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			panic(ctx.Err())
		case <-ticker.C:
		}

		polled, err := client.GetResultV1GetResultGetWithResponse(
			ctx,
			&bfl.GetResultV1GetResultGetParams{Id: task.Id},
		)
		if err != nil {
			panic(err)
		}

		if polled.JSON200 == nil {
			panic(polled.Status())
		}

		result, err := polled.JSON200.AsResultResponse()
		if err != nil {
			panic(err)
		}

		switch result.Status {
		case bfl.StatusResponsePending, bfl.StatusResponseReasoning,
			bfl.StatusResponseGenerating:
			continue
		case bfl.StatusResponseReady:
			_ = result.Result // Untyped JSON containing the generated output.

			return
		case bfl.StatusResponseError, bfl.StatusResponseTaskNotFound,
			bfl.StatusResponseRequestModerated,
			bfl.StatusResponseContentModerated,
			bfl.StatusResponseNoCardGenerated:
			panic(result.Status)
		default:
			panic("unknown task status")
		}
	}
}
