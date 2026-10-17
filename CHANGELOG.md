# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.1, 2026-10-17

- Documentation only. The agent skill explains how to look up every endpoint, parameter and model with `go doc` or on pkg.go.dev, and declares its license.
- The skill no longer lists the API credentials under OpenClaw's `requires.env`, which kept OpenClaw from loading it until they were set. It requires only `go`; the credentials are still documented as what your program reads.
- The README points to the full API reference, and the examples in the README and the skill were checked against the current code.

## v0.1.0, 2026-10-17

- Added `bfl`, a REST client generated with oapi-codegen from Black Forest Labs' BFL API spec version 0.0.1 (OpenAPI 3.1.0), downloaded from https://api.bfl.ai/openapi.json. Covers all 30 operations for image and video generation, image and video tools, fine-tune management, credits and result polling.
- Added OpenAPI 3.0 conversion for nullable unions, constants and keyframe tuples, request and response tests, a submit-then-poll example, development tooling and the `go-bfl` agent skill.
