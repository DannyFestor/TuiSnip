#!/usr/bin/env bash
# Pre-edit guard: guard-path.sh <path>, with the new content on stdin.
# Exits 0 to pass, 2 to deny, 3 to ask, with the reason on stderr.
set -euo pipefail

# shellcheck source=scripts/agent/lib.sh
source "$(dirname "$0")/lib.sh"

readonly PROTECTED_PATHS="$AGENT_DIR/protected-paths"

# Go's convention: the header must come before the first non-comment, non-blank text.
has_generated_header() {
	awk '
		/^\/\/ Code generated .* DO NOT EDIT\.$/ { found = 1; exit }
		/^[[:space:]]*$/ || /^\/\// { next }
		{ exit }
		END { exit !found }
	'
}

# Prints "<tier> <way forward>" for the first rule whose glob matches the path.
matching_rule() {
	local path="$1"
	local tier glob way_forward

	while read -r tier glob way_forward; do
		if [[ "$tier" != "deny" && "$tier" != "ask" ]]; then
			continue
		fi
		# shellcheck disable=SC2053 # the glob must stay unquoted to match as a pattern
		if [[ "$path" == $glob ]]; then
			echo "$tier $way_forward"
			return
		fi
	done <"$PROTECTED_PATHS"
}

check_rules() {
	local path="$1"
	local rule tier way_forward

	rule="$(matching_rule "$path")"
	if [[ -z "$rule" ]]; then
		return
	fi
	tier="${rule%% *}"
	way_forward="${rule#* }"
	if [[ "$tier" == "deny" ]]; then
		echo "$path is generated and must not be edited by hand: $way_forward." >&2
		exit "$EXIT_DENY"
	fi
	echo "$path is protected: $way_forward." >&2
	exit "$EXIT_ASK"
}

check_generated_header() {
	local path="$1"
	local new_content="$2"

	if [[ -f "$path" ]] && has_generated_header <"$path"; then
		echo "$path has a generated-code header and must not be edited by hand: change the generator's input and run \`make generate\`." >&2
		exit "$EXIT_DENY"
	fi
	if has_generated_header <<<"$new_content"; then
		echo "$path would carry a generated-code header, but only generators may write one: add a generator to \`make generate\` instead." >&2
		exit "$EXIT_DENY"
	fi
}

main() {
	local path="$1"
	local relative_path new_content

	new_content="$(cat)"
	relative_path="$(repo_relative_path "$path")"
	if [[ -z "$relative_path" ]]; then
		exit "$EXIT_PASS"
	fi
	check_rules "$relative_path"
	check_generated_header "$path" "$new_content"
	exit "$EXIT_PASS"
}

main "$@"
