#!/usr/bin/env bash
# Antigravity adapter: antigravity.sh <pre|post|pre-invocation|stop>, with the hook's JSON on
# stdin and a JSON decision on stdout. Best effort: exit codes and the working directory are
# undocumented, so every answer is JSON and the repo root comes from workspacePaths.
set -euo pipefail

# shellcheck source=scripts/agent/adapters/common.sh
source "$(dirname "$0")/common.sh"

readonly FINDINGS_FILE_NAME="findings"

# PostToolUse output is ignored, so findings wait here for the next PreInvocation.
findings_file() {
	echo "$(session_state_dir "$(field '.conversationId')")/$FINDINGS_FILE_NAME"
}

# "allow" would skip the user's permission prompt, so a clean call answers "ask".
answer_tool_call() {
	local decision="ask"

	if ((CORE_STATUS == EXIT_DENY)); then
		decision="deny"
	fi
	jq -n --arg decision "$decision" --arg reason "$CORE_MESSAGE" \
		'{decision: $decision} + (if $reason == "" then {} else {reason: $reason} end)'
}

new_content() {
	field '.toolCall.args.CodeContent // .toolCall.args.ReplacementContent // ([.toolCall.args.ReplacementChunks[]?.ReplacementContent] | join("\n"))'
}

pre() {
	if [[ "$(field '.toolCall.name')" == "run_command" ]]; then
		run_core "$AGENT_DIR/guard-command.sh" <<<"$(field '.toolCall.args.CommandLine')"
	else
		run_core "$AGENT_DIR/guard-path.sh" "$(field '.toolCall.args.TargetFile')" <<<"$(new_content)"
	fi
	answer_tool_call
}

post() {
	local file

	run_core "$AGENT_DIR/check-go-file.sh" "$(field '.toolCall.args.TargetFile')"
	if ((CORE_STATUS == EXIT_DENY)); then
		file="$(findings_file)"
		mkdir -p "$(dirname "$file")"
		echo "$CORE_MESSAGE" >>"$file"
	fi
	echo '{}'
}

pre_invocation() {
	local file

	file="$(findings_file)"
	if [[ ! -s "$file" ]]; then
		echo '{}'
		return
	fi
	jq -n --rawfile message "$file" '{injectSteps: [{ephemeralMessage: $message}]}'
	rm -f "$file"
}

stop() {
	run_core "$AGENT_DIR/verify-build.sh" "$(field '.conversationId')"
	if ((CORE_STATUS == EXIT_DENY)); then
		jq -n --arg reason "$CORE_MESSAGE" '{decision: "continue", reason: $reason}'
		return
	fi
	echo '{}'
}

main() {
	local event="$1"

	INPUT="$(cat)"
	cd "$(field '.workspacePaths[0]')"
	case "$event" in
	pre) pre ;;
	post) post ;;
	pre-invocation) pre_invocation ;;
	stop) stop ;;
	*)
		echo "unknown event: $event" >&2
		exit "$EXIT_DENY"
		;;
	esac
}

main "$@"
