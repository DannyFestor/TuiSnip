#!/usr/bin/env bash
# go test -fuzz accepts exactly one package and one fuzz target per run, so each target gets its own run.
set -euo pipefail

fuzztime="${FUZZTIME:-30s}"

go list ./... | while read -r pkg; do
	go test -list '^Fuzz' "$pkg" | { grep '^Fuzz' || true; } | while read -r target; do
		go test -run='^$' -fuzz="^${target}\$" -fuzztime="$fuzztime" "$pkg"
	done
done
