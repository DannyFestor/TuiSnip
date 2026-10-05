#!/usr/bin/env bash
# Mutation-tests the packages changed since origin/main. Uncommitted changes count, because
# gremlins mutates the working tree. A make recipe would lose a failed changed-packages.sh
# in the pipe, since make's shell has no pipefail.
set -euo pipefail

scripts_dir="$(dirname "$0")"

"$scripts_dir/changed-packages.sh" --uncommitted |
	"$scripts_dir/mutable-packages.sh" |
	"$scripts_dir/outermost-packages.sh" |
	while read -r package; do
		# Without the redirect, the tests gremlins runs could read the remaining package list.
		make --no-print-directory test-mutation PKG="$package" </dev/null
	done
