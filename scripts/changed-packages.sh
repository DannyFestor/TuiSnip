#!/usr/bin/env bash
# Prints the Go packages changed on this branch, measured from where it left origin/main,
# so local checks cover what the branch's PR will contain.
#
# Usage: changed-packages.sh [--uncommitted]
#   --uncommitted  also count staged, unstaged, and untracked changes, for checks that run
#                  before a commit exists (the agent Stop gate).
set -euo pipefail

base_ref="origin/main"

merge_base="$(git merge-base "$base_ref" HEAD)"

changed_go_files() {
	if [[ "${1:-}" == "--uncommitted" ]]; then
		git diff --name-only --diff-filter=d "$merge_base" -- '*.go'
		git ls-files --others --exclude-standard -- '*.go'
	else
		git diff --name-only --diff-filter=d "$merge_base" HEAD -- '*.go'
	fi
}

changed_go_files "$@" |
	while read -r file; do
		dirname "$file"
	done |
	sort -u |
	while read -r dir; do
		echo "./$dir"
	done
