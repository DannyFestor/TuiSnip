#!/usr/bin/env bash
# Shared by the core scripts and adapters. Sourced, never run.
# shellcheck disable=SC2034 # the exit codes are read by the scripts that source this file

readonly EXIT_PASS=0
readonly EXIT_DENY=2
readonly EXIT_ASK=3

AGENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
SCRIPTS_DIR="$(dirname "$AGENT_DIR")"
# guard-command.sh follows the cds in a command, so the clone is fixed where the script starts.
START_DIR="$PWD"
readonly AGENT_DIR SCRIPTS_DIR START_DIR

repo_root() {
	git rev-parse --show-toplevel
}

common_git_dir() {
	local dir="$1"

	(cd "$dir" && cd "$(git rev-parse --git-common-dir)" && pwd -P)
}

clone_git_dir() {
	common_git_dir "$START_DIR"
}

# The state directory sits in the common git dir so every worktree of a clone shares it.
# Session ids come from the agent, so anything but a safe file name character is replaced.
session_state_dir() {
	local session="$1"

	echo "$(common_git_dir .)/agent-hooks/$(printf '%s' "$session" | tr -c 'A-Za-z0-9_-' '_')"
}

existing_ancestor() {
	local dir="$1"

	while [[ ! -d "$dir" ]]; do
		dir="$(dirname "$dir")"
	done
	echo "$dir"
}

# The hook's working directory can be the main checkout while the agent works in a linked
# worktree (#149), so the root comes from the path.
checkout_root() {
	local dir

	dir="$(existing_ancestor "$1")"
	if [[ "$(common_git_dir "$dir" 2>/dev/null)" != "$(clone_git_dir)" ]]; then
		return 1
	fi
	git -C "$dir" rev-parse --show-toplevel 2>/dev/null
}

# Resolves symlinks in the deepest existing ancestor, so a path to a file not yet created
# compares equal to the checkout root that git reports.
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

# Prints the path relative to the root of the checkout that holds it, or nothing when no
# checkout of this clone holds it.
repo_relative_path() {
	local path root

	path="$(canonical_path "$1")"
	root="$(checkout_root "$path")" || return 0
	if [[ "$path" == "$root/"* ]]; then
		echo "${path#"$root/"}"
	fi
}
