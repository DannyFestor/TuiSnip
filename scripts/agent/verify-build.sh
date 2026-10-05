#!/usr/bin/env bash
# Stop gate: verify-build.sh <session-id>. Builds, layer-checks, lints, and unit-tests the Go
# packages changed since origin/main, uncommitted changes included.
# Exits 0 to let the agent stop, or 2 with the failure on stderr to send it back to work.
set -euo pipefail

# shellcheck source=scripts/agent/lib.sh
source "$(dirname "$0")/lib.sh"

readonly BASE_REF="origin/main"
# Past this many consecutive blocks the agent is stuck; let it stop and hand over to the user.
readonly MAX_BLOCKS=3
readonly OUTPUT_TAIL_LINES=80
# Feature and e2e packages hold only tagged files and are left to pre-push and CI.
readonly TAGGED_TEST_PACKAGES='^\./test(/|$)'

changed_packages() {
	"$SCRIPTS_DIR/changed-packages.sh" --uncommitted
}

untracked_go_files_digest() {
	local file

	git ls-files --others --exclude-standard -- '*.go' | while read -r file; do
		echo "$file $(git hash-object -- "$file")"
	done
}

go_fingerprint() {
	{
		git diff "$(git merge-base "$BASE_REF" HEAD)" -- '*.go' go.mod go.sum
		untracked_go_files_digest
	} | git hash-object --stdin
}

run_step() {
	local step="$1"
	shift
	local output

	if ! output="$("$@" 2>&1)"; then
		echo "$step failed:"
		tail -n "$OUTPUT_TAIL_LINES" <<<"$output"
		return 1
	fi
}

unit_test_packages() {
	grep -Ev "$TAGGED_TEST_PACKAGES" <<<"$1" || true
}

run_checks() {
	local packages="$1"
	local -a lint_packages test_packages

	read -r -a lint_packages <<<"$(tr '\n' ' ' <<<"$packages")"
	read -r -a test_packages <<<"$(unit_test_packages "$packages" | tr '\n' ' ')"
	run_step "go build" go build ./... &&
		run_step "go-arch-lint" go-arch-lint check &&
		run_step "golangci-lint" golangci-lint run --allow-serial-runners "${lint_packages[@]}" &&
		if ((${#test_packages[@]} > 0)); then
			run_step "go test" go test -short -race "${test_packages[@]}"
		fi
}

read_state() {
	local file="$1"

	if [[ -f "$file" ]]; then
		cat "$file"
	fi
}

block_or_release() {
	local blocks_file="$1"
	local output="$2"
	local blocks

	blocks=$(($(read_state "$blocks_file" || true) + 1))
	echo "$blocks" >"$blocks_file"
	if ((blocks > MAX_BLOCKS)); then
		echo "The Stop gate has failed $MAX_BLOCKS times in a row, so the agent may stop. Still failing:" >&2
		echo "$output" >&2
		exit "$EXIT_PASS"
	fi
	echo "$output" >&2
	exit "$EXIT_DENY"
}

main() {
	local session="$1"
	local state_dir packages fingerprint output

	cd "$(repo_root)"
	packages="$(changed_packages)"
	if [[ -z "$packages" ]]; then
		exit "$EXIT_PASS"
	fi
	state_dir="$(session_state_dir "$session")"
	mkdir -p "$state_dir"
	fingerprint="$(go_fingerprint)"
	if [[ "$fingerprint" == "$(read_state "$state_dir/passed-fingerprint")" ]]; then
		exit "$EXIT_PASS"
	fi
	if output="$(run_checks "$packages")"; then
		echo "$fingerprint" >"$state_dir/passed-fingerprint"
		rm -f "$state_dir/blocks"
		exit "$EXIT_PASS"
	fi
	block_or_release "$state_dir/blocks" "$output"
}

main "$@"
