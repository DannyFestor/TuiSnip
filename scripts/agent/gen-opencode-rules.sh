#!/usr/bin/env bash
# Writes opencode.json from protected-paths. OpenCode matches permission.edit globs against
# the worktree-relative path with a wildcard where * crosses /, like guard-path.sh, so the
# globs copy over unchanged.
set -euo pipefail

# shellcheck source=scripts/agent/lib.sh
source "$(dirname "$0")/lib.sh"

OUTPUT="$(dirname "$SCRIPTS_DIR")/opencode.json"
readonly OUTPUT

# OpenCode applies the last matching rule and guard-path.sh the first, so the order is
# reversed to keep both agreeing when globs overlap.
edit_rules() {
	awk '
		$1 == "deny" || $1 == "ask" { rules[++count] = $1 "\t" $2 }
		END { for (i = count; i > 0; i--) print rules[i] }
	' "$AGENT_DIR/protected-paths"
}

main() {
	edit_rules | jq --raw-input --null-input --indent 2 '
		{
			"$schema": "https://opencode.ai/config.json",
			"formatter": false,
			"permission": {
				"edit": [inputs | split("\t") | {(.[1]): .[0]}] | add
			}
		}
	' >"$OUTPUT"
}

main "$@"
