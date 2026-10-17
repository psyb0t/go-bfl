# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.0, 2026-10-17

- Added `bfl`, a REST client generated with oapi-codegen from Black Forest Labs' BFL API spec version 0.0.1 (OpenAPI 3.1.0), downloaded from https://api.bfl.ai/openapi.json. Covers all 30 operations for image and video generation, image and video tools, fine-tune management, credits and result polling.
- Added OpenAPI 3.0 conversion for nullable unions, constants and keyframe tuples, request and response tests, a submit-then-poll example, development tooling and the `go-bfl` agent skill.
