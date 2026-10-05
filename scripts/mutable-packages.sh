#!/usr/bin/env bash
# Reads Go package paths on stdin and prints those with non-test Go files under the build tags
# in MUTATION_TAGS. gremlins has nothing to mutate in the rest, and fails on a package whose
# files the tags all exclude, such as test/e2e.
#
# Usage: changed-packages.sh | MUTATION_TAGS=feature mutable-packages.sh
set -euo pipefail

has_code() {
	local package="$1"

	[[ -n "$(go list -e -tags "$MUTATION_TAGS" -f '{{if .GoFiles}}code{{end}}' "$package")" ]]
}

while read -r package; do
	if has_code "$package"; then
		echo "$package"
	fi
done
