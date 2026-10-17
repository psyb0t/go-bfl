// Package bfl is an unofficial client for the Black Forest Labs API,
// generated from its public OpenAPI document.
package bfl

//go:generate bash scripts/fetch-spec.sh
//go:generate bash scripts/patch-spec.sh
//go:generate go tool oapi-codegen --config=config.yaml api/openapi.fixed.json
