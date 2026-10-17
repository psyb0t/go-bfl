#!/bin/bash
# Downconverts OpenAPI 3.1 schemas for the generator's OpenAPI 3.0 reader.
# Nullable anyOf branches become nullable: true; single remaining branches
# are inlined to avoid undeclared single-union accessor types. String const
# becomes a one-value enum. Genuine multi-type unions remain intact.
# prefixItems tuples become arrays of untyped JSON values: OpenAPI 3.0 has
# no positional tuple schema. Keep minItems/maxItems and the wire payload.
# Normalize provider first-person wording in generated documentation.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

SPEC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/api"
readonly SPEC_DIR

on_error() {
	local exit_code="$1"
	local line="$2"

	printf '{"level":"ERROR","file":"%s","line":%d,"msg":"command failed","exit_code":%d}\n' \
		"${BASH_SOURCE[1]}" "$line" "$exit_code" >&2
}

log() {
	local level="$1"
	shift

	jq -nc \
		--arg time "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
		--arg level "$level" \
		--arg file "${BASH_SOURCE[1]##*/}" \
		--arg line "${BASH_LINENO[0]}" \
		--arg func "${FUNCNAME[1]:-main}" \
		--arg msg "$*" \
		'{time:$time,level:$level,file:$file,line:($line|tonumber),func:$func,msg:$msg}' >&2
}

main() {
	local tmp
	tmp="$(mktemp "${SPEC_DIR}/.openapi.fixed.XXXXXX")"
	# shellcheck disable=SC2064 # expand tmp while the local variable exists
	trap "rm -f -- '${tmp}'" EXIT

	jq '
	  walk(if type == "object" then
	    (if has("anyOf") and any(.anyOf[]; .type? == "null") then
	      .anyOf |= map(select(.type? != "null"))
	      | .nullable = true
	    else . end)
	    | (if has("anyOf") and (.anyOf | length) == 1 then
	        . + .anyOf[0] | del(.anyOf)
	      else . end)
	    | (if has("prefixItems") then .items = {} | del(.prefixItems)
	      else . end)
	    | (if has("const") then .enum = [.const] | del(.const)
	      else . end)
	    | (if (.description? | type) == "string" then
	        .description |= gsub("our latest"; "the latest")
	      else . end)
	  else . end)
	  | .openapi = "3.0.3"
	' "${SPEC_DIR}/openapi.json" >"$tmp"
	chmod 0644 "$tmp"
	mv -- "$tmp" "${SPEC_DIR}/openapi.fixed.json"
	log INFO "patched nullable unions and constants"
}

main "$@"
