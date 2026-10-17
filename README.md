# go-bfl

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/go-bfl.svg)](https://pkg.go.dev/github.com/psyb0t/go-bfl)
[![CI](https://github.com/psyb0t/go-bfl/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/go-bfl/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/go-bfl/badges/version.svg)](https://github.com/psyb0t/go-bfl/tags)
[![license](https://raw.githubusercontent.com/psyb0t/go-bfl/badges/license.svg)](LICENSE)
[![imported by](https://raw.githubusercontent.com/psyb0t/go-bfl/badges/importers.svg)](https://github.com/psyb0t/go-bfl/blob/badges/importers.md)

Unofficial Go client for the Black Forest Labs FLUX API, generated with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) from its [public OpenAPI document](https://api.bfl.ai/openapi.json). I am not affiliated with Black Forest Labs.

The package is `bfl`, in the repository root. Everything in `client.gen.go` is generated. Regenerate it rather than editing it by hand. All 30 operations are included: image and video generation, image and video tools, fine-tune management, credits and result polling.

## Contents

[Install](#install) · [Usage](#usage) · [Auth](#auth) · [Regenerating](#regenerating) · [Agent integrations](#agent-integrations) · [Development](#development) · [License](#license)

## Install

```sh
go get github.com/psyb0t/go-bfl
```

## Usage

Every endpoint has a typed method on `ClientWithResponses`. Submit a task, keep its ID, then poll until `Ready`. This example submits a paid image generation task with bounded polling:

```go
package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/psyb0t/go-bfl"
)

func main() {
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
```

The same code is in `example_test.go`, so `make test` checks that it compiles. The example polls on the configured API host using the task ID. Result data is untyped JSON; inspect it before using a download URL. The spec also returns `polling_url` for regional routing; validate its HTTPS host before sending credentials there.

Method names follow the spec's operation IDs. Some success bodies are unions: use `AsAsyncResponse()`, `AsAsyncWebhookResponse()`, `AsResultResponse()` or `AsSettledCostResultResponse()` as appropriate. These accessors decode without validating required fields. HTTP failures usually return nil `JSON200`; inspect `StatusCode()` and `Body`. Only transport and decoding failures use the Go error return. There are no automatic retries, polling or schema validation.

## Auth

The spec's `APIKeyHeader` scheme sends the API key in `x-key`. Read `BFL_API_KEY` from the environment and set it with `WithRequestEditorFn`, as shown above. The base URL is `https://api.bfl.ai`. `/v1/get_result` has no security requirement in the spec; a client-wide request editor still adds its header. Keep task IDs and signed result URLs private. Back off on 429 responses; avoid retrying a paid submission blindly.

## Regenerating

The spec is not committed. `make generate` downloads it into the gitignored `api/` directory. `gen.go` declares the steps:

- `scripts/fetch-spec.sh` downloads `https://api.bfl.ai/openapi.json` to `api/openapi.json`.
- `scripts/patch-spec.sh` writes `api/openapi.fixed.json` as OpenAPI 3.0.3. It converts nullable `anyOf` branches to `nullable: true`, inlines singleton unions, converts `const` to single-value enums and replaces `prefixItems` tuples with untyped array items. The generator cannot express positional tuple constraints; callers must validate keyframe pairs. Genuine unions remain intact.
- `go tool oapi-codegen --config=config.yaml api/openapi.fixed.json` generates every operation. The config maps `number` to `float64` for costs, credits and progress.

```sh
make generate
```

`.github/workflows/regen.yml` does the same every day in CI. When the new spec changes the client, it runs `make lint`, `make test` and `make sec`. If a check fails, Claude edits the code and the checks run again, up to three times, and a review step reads the result. A passing change is committed with a CHANGELOG entry and tagged as the next release: a patch version for fixes and additions, a minor version for breaking changes.

## Agent integrations

The [skill](.agents/skills/go-bfl) teaches a coding agent how to use go-bfl when it writes Go. It works in any agent that reads `.agents/skills/`, and installs natively in the clients below.

### Claude Code

```bash
claude plugin marketplace add psyb0t/agents
claude plugin install go-bfl@psyb0t
```

### Codex

```bash
codex plugin marketplace add psyb0t/agents
codex plugin add go-bfl@psyb0t
```

Codex also picks the skill up with no install in any checkout of this repo.

### OpenClaw

The skill is published to ClawHub on every release:

```bash
openclaw skills install @psyb0t/go-bfl
```

## Development

Every command runs in the `Dockerfile.dev` image. The host needs Docker and Make.

```sh
make help      # list targets
make dep       # go mod tidy + vendor
make generate  # download the spec and regenerate
make lint      # shellcheck, shfmt, go fix, golangci-lint
make test      # go test -race
make sec       # gitleaks, govulncheck, semgrep
```

## License

[MIT](LICENSE). Generated code derives from Black Forest Labs' published API specification. The specification itself is not redistributed. Its metadata does not declare a license or terms URL; this repository's MIT license does not grant rights to the provider's service, models or generated media.
