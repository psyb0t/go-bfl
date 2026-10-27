# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.3, 2026-10-27

- The development image is based on Debian (`golang:1.26.10-trixie`) instead of Alpine. Alpine keeps only the latest build of each package, so the exact pins broke the image build whenever a package was updated. System packages now come from the snapshot.debian.org date the base image was built from, which keeps them fixed for a given base image digest. jq 1.8.2, shellcheck 0.11.0 and shfmt 3.13.1 keep their versions, and Semgrep runs in its own virtual environment.

## v0.1.2, 2026-10-22

- Added `Flux3ImageInputs.Width` and `Flux3ImageInputs.Height` for exact output size in pixels. Set both together; they override `Resolution`.
- Added the `Flux3ImageInputsResolutionMatchInput` constant (`"match_input"`) for `Flux3ImageInputs.Resolution`. It applies to edits only and keeps the size of the first reference image. `Valid()` accepts it.
- Added an optional `User` field, an opaque end user identifier, to `Flux3ImageInputs`, `Flux3ImageKleinInputs` and `FinetuneFluxUltraInput`.
- Updated the `Flux3ImageInputs.Resolution` doc comment: `1k` is the default.
- Raised the minimum Go version in `go.mod` from 1.26.9 to 1.26.10, which fixes the `crypto/x509` vulnerability GO-2026-6809 found by govulncheck. The dev image and the skill setup notes use 1.26.10 too.
- Updated the `ca-certificates` pin in `Dockerfile.dev` to `20261020-r0`.

## v0.1.1, 2026-10-17

- Documentation only. The agent skill explains how to look up every endpoint, parameter and model with `go doc` or on pkg.go.dev, and declares its license.
- The skill no longer lists the API credentials under OpenClaw's `requires.env`, which kept OpenClaw from loading it until they were set. It requires only `go`; the credentials are still documented as what your program reads.
- The README points to the full API reference, and the examples in the README and the skill were checked against the current code.

## v0.1.0, 2026-10-17

- Added `bfl`, a REST client generated with oapi-codegen from Black Forest Labs' BFL API spec version 0.0.1 (OpenAPI 3.1.0), downloaded from https://api.bfl.ai/openapi.json. Covers all 30 operations for image and video generation, image and video tools, fine-tune management, credits and result polling.
- Added OpenAPI 3.0 conversion for nullable unions, constants and keyframe tuples, request and response tests, a submit-then-poll example, development tooling and the `go-bfl` agent skill.
