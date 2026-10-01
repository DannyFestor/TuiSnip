#!/usr/bin/env bash
# Prints one side of an OpenCode apply_patch for one file: patch-sides.sh <path> <removed|added>,
# with the patch on stdin. The prefix is dropped, so the lines read as that file's text.
set -euo pipefail

main() {
	local path="$1"
	local side="$2"
	local prefix="+"

	if [[ "$side" == "removed" ]]; then
		prefix="-"
	fi
	awk -v path="$path" -v prefix="$prefix" '
		/^\*\*\* (Add|Update|Delete) File: / { current = substr($0, index($0, ": ") + 2); next }
		current == path && substr($0, 1, 1) == prefix { print substr($0, 2) }
	'
}

main "$@"
