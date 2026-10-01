#!/usr/bin/env bash
# OpenCode adapter: opencode.sh pre-edit, with {tool, sessionID, args} of a file tool call as
# JSON on stdin. Exits 2 with the message on stderr to block, so the plugin only has to throw.
# A plugin cannot ask the user, so the comment guard runs without its ask step.
set -euo pipefail

# shellcheck source=scripts/agent/adapters/common.sh
source "$(dirname "$0")/common.sh"

guard_comments() {
	local path="$1"
	local replaced_file="$2"

	run_core "$AGENT_DIR/guard-go-comments.sh" --without-ask "$(field '.sessionID')" "$path" "$replaced_file"
	if ((CORE_STATUS == EXIT_DENY)); then
		echo "$CORE_MESSAGE" >&2
		exit "$EXIT_DENY"
	fi
}

guard_patch() {
	local patch path

	patch="$(field '.args.patchText')"
	while read -r path; do
		guard_comments "$path" <("$AGENT_DIR/patch-sides.sh" "$path" removed <<<"$patch") \
			< <("$AGENT_DIR/patch-sides.sh" "$path" added <<<"$patch")
	done < <("$AGENT_DIR/patch-paths.sh" <<<"$patch")
}

pre_edit() {
	local path

	path="$(field '.args.filePath')"
	case "$(field '.tool')" in
	apply_patch) guard_patch ;;
	write) guard_comments "$path" "$(file_or_empty "$path")" <<<"$(field '.args.content')" ;;
	*) guard_comments "$path" <(field '.args.oldString') <<<"$(field '.args.newString')" ;;
	esac
}

main() {
	local event="$1"

	INPUT="$(cat)"
	case "$event" in
	pre-edit) pre_edit ;;
	*)
		echo "unknown event: $event" >&2
		exit "$EXIT_DENY"
		;;
	esac
}

main "$@"
