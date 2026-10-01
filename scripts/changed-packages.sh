#!/usr/bin/env bash
# Prints the Go packages changed on this branch, measured from where it left origin/main,
# so local checks cover what the branch's PR will contain.
set -euo pipefail

base_ref="origin/main"

merge_base="$(git merge-base "$base_ref" HEAD)"

git diff --name-only --diff-filter=d "$merge_base" HEAD -- '*.go' |
	while read -r file; do
		dirname "$file"
	done |
	sort -u |
	while read -r dir; do
		echo "./$dir"
	done
