#!/usr/bin/env bash
# Post-edit check: check-go-file.sh <path>. Formats a Go file and vets its package.
# Exits 0 to pass or 2 with the findings on stderr. Tests don't run here, so the red phase
# of TDD isn't reported as a failure.
set -euo pipefail

# shellcheck source=scripts/agent/lib.sh
source "$(dirname "$0")/lib.sh"

# The feature and e2e packages hold only tagged files; without the tags vet finds no package.
readonly BUILD_TAGS="feature,e2e"

report_failure() {
	local step="$1"
	local output="$2"

	echo "$step failed:" >&2
	echo "$output" >&2
	exit "$EXIT_DENY"
}

run_step() {
	local step="$1"
	shift
	local output

	if ! output="$("$@" 2>&1)"; then
		report_failure "$step" "$output"
	fi
}

main() {
	local relative_path

	relative_path="$(repo_relative_path "$1")"
	if [[ "$relative_path" != *.go || ! -f "$1" ]]; then
		exit "$EXIT_PASS"
	fi
	cd "$(repo_root)"
	run_step "golangci-lint fmt" golangci-lint fmt "$relative_path"
	run_step "go vet" go vet -tags "$BUILD_TAGS" "./$(dirname "$relative_path")"
	exit "$EXIT_PASS"
}

main "$@"
