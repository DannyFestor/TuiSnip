#!/usr/bin/env bash
# Pre-edit comment guard: guard-go-comments.sh [--without-ask] <session> <path> <replaced-text-file>,
# with the edit's new text on stdin. Agents add comments that restate the code, so the first
# submission of a new comment is denied for a self-check and a resubmission asks the user.
# --without-ask is for systems that cannot ask from a hook: there the resubmission passes.
# Exits 0 to pass, 2 to deny, or 3 to ask, with the message on stderr.
set -euo pipefail

# shellcheck source=scripts/agent/lib.sh
source "$(dirname "$0")/lib.sh"

readonly GENERATED_HEADER_PATTERN='^// Code generated .* DO NOT EDIT\.$'
readonly EXEMPT_COMMENT_PATTERN='^//([a-z0-9]+:[a-z0-9]|nolint)|^// (Unordered output|Output):|^// ENUM\('
readonly APPROVALS_FILE_NAME="comment-approvals"

# Prints each comment in the text, trimmed: full-line // and /* comments, and trailing //
# comments found by the first " //" after code. Literals are dropped first, so a "//" inside a
# string is not taken for a comment.
comment_lines() {
	awk '
		/^[[:space:]]*(\/\/|\/\*)/ { sub(/^[[:space:]]+/, ""); print; next }
		{
			code = $0
			gsub(/"([^"\\]|\\.)*"|`[^`]*`|'\''([^'\''\\]|\\.)*'\''/, "", code)
			if (match(code, /[^[:space:]][[:space:]]+\/\//)) { print substr(code, RSTART + RLENGTH - 2) }
		}
	'
}

added_comments() {
	local replaced_file="$1"
	local new_text="$2"

	comm -13 <(comment_lines <"$replaced_file" | sort) <(comment_lines <<<"$new_text" | sort) |
		grep -Ev "$EXEMPT_COMMENT_PATTERN" || true
}

is_generated() {
	local path="$1"
	local new_text="$2"

	grep -Eq "$GENERATED_HEADER_PATTERN" "$path" 2>/dev/null ||
		grep -Eq "$GENERATED_HEADER_PATTERN" <<<"$new_text"
}

is_exempt_file() {
	local relative_path="$1"

	[[ "$relative_path" != *.go || "$(basename "$relative_path")" == "doc.go" ]]
}

comments_fingerprint() {
	local relative_path="$1"
	local comments="$2"

	printf '%s\n%s\n' "$relative_path" "$comments" | git hash-object --stdin
}

listed_comments() {
	local relative_path="$1"
	local comments="$2"

	echo "New comment in $relative_path:"
	awk '{ print "  " $0 }' <<<"$comments"
}

deny_for_self_check() {
	local retry_hint="$1"

	echo "Delete it if it says what the code does (docs/standards/code.md#comments)."
	echo "If it says why, $retry_hint"
}

decide() {
	local ask="$1"
	local approvals_file="$2"
	local fingerprint="$3"

	if ! grep -qxF "$fingerprint" "$approvals_file" 2>/dev/null; then
		echo "$fingerprint" >>"$approvals_file"
		if [[ "$ask" == true ]]; then
			deny_for_self_check "submit it again and the user will be asked."
		else
			deny_for_self_check "submit it again."
		fi
		return "$EXIT_DENY"
	fi
	if [[ "$ask" == true ]]; then
		echo "The agent submitted it again after a self-check. Approve it only if it says why."
		return "$EXIT_ASK"
	fi
	return "$EXIT_PASS"
}

main() {
	local ask=true

	if [[ "$1" == "--without-ask" ]]; then
		ask=false
		shift
	fi

	local session="$1"
	local path="$2"
	local replaced_file="$3"
	local relative_path new_text comments state_dir status=0 decision

	relative_path="$(repo_relative_path "$path")"
	new_text="$(cat)"
	if [[ -z "$relative_path" ]] || is_exempt_file "$relative_path" || is_generated "$path" "$new_text"; then
		exit "$EXIT_PASS"
	fi
	comments="$(added_comments "$replaced_file" "$new_text")"
	if [[ -z "$comments" ]]; then
		exit "$EXIT_PASS"
	fi
	state_dir="$(session_state_dir "$session")"
	mkdir -p "$state_dir"
	decision="$(decide "$ask" "$state_dir/$APPROVALS_FILE_NAME" \
		"$(comments_fingerprint "$relative_path" "$comments")")" || status=$?
	if ((status != EXIT_PASS)); then
		{
			listed_comments "$relative_path" "$comments"
			echo "$decision"
		} >&2
	fi
	exit "$status"
}

main "$@"
