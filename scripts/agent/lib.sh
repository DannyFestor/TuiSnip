#!/usr/bin/env bash
# Shared by the core scripts and adapters. Sourced, never run.
# shellcheck disable=SC2034 # the exit codes are read by the scripts that source this file

readonly EXIT_PASS=0
readonly EXIT_DENY=2
readonly EXIT_ASK=3

AGENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
SCRIPTS_DIR="$(dirname "$AGENT_DIR")"
readonly AGENT_DIR SCRIPTS_DIR

repo_root() {
	git rev-parse --show-toplevel
}

# The state directory sits in the common git dir so every worktree of a clone shares it.
session_state_dir() {
	local session="$1"
	local common_dir

	common_dir="$(cd "$(git rev-parse --git-common-dir)" && pwd -P)"
	echo "$common_dir/agent-hooks/$session"
}

# Resolves symlinks in the deepest existing ancestor, so a path to a file not yet created
# compares equal to the repo root that git reports.
canonical_path() {
	local path="$1"
	local dir="$path"
	local rest=""

	if [[ "$path" != /* ]]; then
		dir="$PWD/$path"
	fi
	while [[ ! -d "$dir" ]]; do
		rest="/$(basename "$dir")$rest"
		dir="$(dirname "$dir")"
	done
	echo "$(cd "$dir" && pwd -P)$rest"
}

# Prints the path relative to the repo root, or nothing when the path is outside the repo.
repo_relative_path() {
	local path root

	path="$(canonical_path "$1")"
	root="$(repo_root)"
	if [[ "$path" == "$root/"* ]]; then
		echo "${path#"$root/"}"
	fi
}
