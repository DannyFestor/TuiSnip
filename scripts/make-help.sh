#!/usr/bin/env bash
set -euo pipefail

awk '
BEGIN { print "Usage: make [target]" }
/^##@ / { printf "\n%s\n", substr($0, 5); next }
/^[a-zA-Z0-9_-]+:.*## / {
	name = $0
	sub(/:.*/, "", name)
	description = $0
	sub(/.*## /, "", description)
	printf "  %-22s %s\n", name, description
}' "$@"
