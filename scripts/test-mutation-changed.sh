#!/usr/bin/env bash
# Mutation-tests the packages changed since origin/main. Uncommitted changes count, because
# gremlins mutates the working tree. A make recipe would lose a failed changed-packages.sh
# in the pipe, since make's shell has no pipefail.
#
# Usage: MUTATION_TAGS=feature test-mutation-changed.sh (make test-mutation-changed sets it)
set -euo pipefail

: "${MUTATION_TAGS:?set MUTATION_TAGS to the build tags gremlins runs with, or use make test-mutation-changed}"

scripts_dir="$(dirname "$0")"

report_for() {
	local package="$1"
	local slug="${package#./}"

	echo "mutation-report-${slug//\//-}.json"
}

"$scripts_dir/changed-packages.sh" --uncommitted |
	"$scripts_dir/mutable-packages.sh" |
	"$scripts_dir/outermost-packages.sh" |
	while read -r package; do
		# Without the redirect, the tests gremlins runs could read the remaining package list.
		make --no-print-directory test-mutation PKG="$package" MUTATION_REPORT="$(report_for "$package")" </dev/null
	done
