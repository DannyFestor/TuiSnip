#!/usr/bin/env bash
# Shared by the adapters. Sourced, never run. Expects the hook's JSON in INPUT.
# shellcheck disable=SC2034 # CORE_STATUS and CORE_MESSAGE are read by the adapters

# shellcheck source=scripts/agent/lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/../lib.sh"

field() {
	jq -r "$1 // empty" <<<"$INPUT"
}

# Runs a core script, keeping its exit code and stderr instead of letting set -e exit.
run_core() {
	CORE_STATUS=0
	CORE_MESSAGE="$("$@" 2>&1 >/dev/null)" || CORE_STATUS=$?
}
