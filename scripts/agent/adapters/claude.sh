#!/usr/bin/env bash
# Claude Code adapter: claude.sh <pre-edit|pre-bash|post-edit|stop>, with the hook's JSON on
# stdin. Claude Code already treats exit 2 as "block and show stderr to Claude", so a deny
# passes straight through; ask and the released Stop gate need JSON on stdout.
set -euo pipefail

# shellcheck source=scripts/agent/adapters/common.sh
source "$(dirname "$0")/common.sh"

pass_through_deny() {
	if ((CORE_STATUS == EXIT_DENY)); then
		echo "$CORE_MESSAGE" >&2
		exit "$EXIT_DENY"
	fi
}

ask() {
	jq -n --arg reason "$1" '{
		hookSpecificOutput: {
			hookEventName: "PreToolUse",
			permissionDecision: "ask",
			permissionDecisionReason: $reason
		}
	}'
}

collect_ask() {
	if ((CORE_STATUS == EXIT_ASK)); then
		ASK_REASONS+=("$CORE_MESSAGE")
	fi
}

guard_comments() {
	local path="$1"
	local content="$2"
	local session

	session="$(field '.session_id')"
	if [[ "$(field '.tool_name')" == "Write" ]]; then
		run_core "$AGENT_DIR/guard-go-comments.sh" "$session" "$path" "$(file_or_empty "$path")" <<<"$content"
		return
	fi
	run_core "$AGENT_DIR/guard-go-comments.sh" "$session" "$path" <(field '.tool_input.old_string') <<<"$content"
}

pre_edit() {
	local path content
	local -a ASK_REASONS=()

	path="$(field '.tool_input.file_path // .tool_input.notebook_path')"
	content="$(field '.tool_input.content // .tool_input.new_string // .tool_input.new_source')"
	run_core "$AGENT_DIR/guard-path.sh" "$path" <<<"$content"
	pass_through_deny
	collect_ask
	guard_comments "$path" "$content"
	pass_through_deny
	collect_ask
	if ((${#ASK_REASONS[@]} > 0)); then
		ask "$(printf '%s\n' "${ASK_REASONS[@]}")"
	fi
}

pre_bash() {
	run_core "$AGENT_DIR/guard-command.sh" <<<"$(field '.tool_input.command')"
	pass_through_deny
}

post_edit() {
	run_core "$AGENT_DIR/check-go-file.sh" "$(field '.tool_input.file_path')"
	pass_through_deny
}

stop() {
	run_core "$AGENT_DIR/verify-build.sh" "$(field '.session_id')"
	pass_through_deny
	if [[ -n "$CORE_MESSAGE" ]]; then
		jq -n --arg message "$CORE_MESSAGE" '{systemMessage: $message}'
	fi
}

main() {
	local event="$1"

	INPUT="$(cat)"
	# The cwd follows a worktree the agent entered; the script's own location does not.
	cd "$(field '.cwd')"
	case "$event" in
	pre-edit) pre_edit ;;
	pre-bash) pre_bash ;;
	post-edit) post_edit ;;
	stop) stop ;;
	*)
		echo "unknown event: $event" >&2
		exit "$EXIT_DENY"
		;;
	esac
}

main "$@"
