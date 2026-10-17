package bfl_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/psyb0t/go-bfl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAPIKey = "EXAMPLE-DO-NOT-USE"

func newTestClient(
	t *testing.T,
	handler http.HandlerFunc,
) *bfl.ClientWithResponses {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := bfl.NewClientWithResponses(
		srv.URL,
		bfl.WithRequestEditorFn(
			func(_ context.Context, req *http.Request) error {
				req.Header.Set("x-key", testAPIKey)

				return nil
			},
		),
	)
	require.NoError(t, err)

	return client
}

func TestSubmitImage(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/flux-pro-1.1", r.URL.Path)
		assert.Equal(t, testAPIKey, r.Header.Get("x-key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body bfl.FluxPro11Inputs

		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "A small cabin beside a lake", *body.Prompt)
		assert.Nil(t, body.Seed)

		w.Header().Set("Content-Type", "application/json")
		// A disconnected test client cannot affect these response assertions.
		_, _ = w.Write([]byte(`{
			"id":"task-1",
			"polling_url":"https://api.bfl.ai/v1/get_result?id=task-1",
			"cost":1.23456789,"input_mp":null,"output_mp":0.75
		}`))
	})

	prompt := "A small cabin beside a lake"
	res, err := client.FluxPro11V1FluxPro11PostWithResponse(
		t.Context(),
		bfl.FluxPro11Inputs{Prompt: &prompt},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)

	task, err := res.JSON200.AsAsyncResponse()
	require.NoError(t, err)
	assert.Equal(t, "task-1", task.Id)
	assert.Contains(t, task.PollingUrl, "/v1/get_result?id=task-1")
	require.NotNil(t, task.Cost)
	assert.InDelta(t, 1.23456789, *task.Cost, 0)
	assert.Nil(t, task.InputMp)
}

func TestPollResult(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/get_result", r.URL.Path)
		assert.Equal(t, "task-1 &?", r.URL.Query().Get("id"))
		assert.Equal(t, testAPIKey, r.Header.Get("x-key"))
		assert.Len(t, r.URL.Query(), 1)

		w.Header().Set("Content-Type", "application/json")
		// The response is validated through the client below.
		_, _ = w.Write([]byte(`{
			"id":"task-1 &?","status":"Ready","cost":0.125,
			"progress":1,"details":null,"preview":null,
			"result":{"sample":"https://example.invalid/image.png"}
		}`))
	})

	res, err := client.GetResultV1GetResultGetWithResponse(
		t.Context(),
		&bfl.GetResultV1GetResultGetParams{Id: "task-1 &?"},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)

	result, err := res.JSON200.AsSettledCostResultResponse()
	require.NoError(t, err)
	assert.Equal(t, bfl.StatusResponseReady, result.Status)
	assert.InDelta(t, 0.125, result.Cost, 0)
	assert.Equal(t, map[string]any{
		"sample": "https://example.invalid/image.png",
	}, result.Result)
	assert.Nil(t, result.Details)
}

func TestCredits(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/credits", r.URL.Path)
		assert.Equal(t, testAPIKey, r.Header.Get("x-key"))

		w.Header().Set("Content-Type", "application/json")
		// The decoded value below proves the test response was received.
		_, _ = w.Write([]byte(`{"credits":123.456789}`))
	})

	res, err := client.GetCreditsV1CreditsGetWithResponse(t.Context())
	require.NoError(t, err)
	require.NotNil(t, res.JSON200)
	assert.InDelta(t, 123.456789, res.JSON200.Credits, 0)
}

func TestHTTPFailures(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		status int
	}{
		{name: "unauthorized", status: http.StatusUnauthorized},
		{name: "rate limited", status: http.StatusTooManyRequests},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
			}

			client := newTestClient(t, handler)
			res, err := client.GetCreditsV1CreditsGetWithResponse(t.Context())
			require.NoError(t, err)
			assert.Equal(t, tc.status, res.StatusCode())
			assert.Nil(t, res.JSON200)
		})
	}
}

func TestCanceledRequest(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("canceled request reached server")
	})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := client.GetCreditsV1CreditsGetWithResponse(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestVideoKeyframes(t *testing.T) {
	t.Parallel()

	var keyframes bfl.Flux3VideoI2VInputs_Keyframes

	require.NoError(t, keyframes.FromFlux3VideoI2VInputsKeyframes3(
		bfl.Flux3VideoI2VInputsKeyframes3{
			{0.0, "https://example.invalid/start.png"},
			{3.5, "https://example.invalid/end.png"},
		},
	))

	var body bfl.Flux3VideoInputsBody

	require.NoError(t, body.FromFlux3VideoI2VInputs(
		bfl.Flux3VideoI2VInputs{
			Mode:      bfl.Flux3VideoI2VInputsModeI2v,
			Prompt:    "A moving camera",
			Keyframes: keyframes,
		},
	))
	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"mode":"i2v","prompt":"A moving camera",
		"keyframes":[[0,"https://example.invalid/start.png"],
		[3.5,"https://example.invalid/end.png"]]
	}`, string(encoded))
}
