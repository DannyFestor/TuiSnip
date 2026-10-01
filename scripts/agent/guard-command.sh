#!/usr/bin/env bash
# Pre-command guard: the shell command arrives on stdin, never argv, so heredocs and quotes
# arrive unmangled. Exits 0 to pass or 2 to deny, with the reason on stderr.
#
# This is a guardrail against mistakes, not a security boundary: a command nested in
# `sh -c '...'` is not inspected.
set -euo pipefail

# shellcheck source=scripts/agent/lib.sh
source "$(dirname "$0")/lib.sh"

readonly PROTECTED_BRANCH="main"
readonly USE_EDIT_TOOLS="use the edit tools instead, so the post-edit format and vet run"
readonly INTERPRETERS_PATTERN='^(python[0-9.]*|node|ruby|perl|php|deno|bun)$'
# Standard-stream writes are output, not file writes.
readonly STREAM_WRITES_PATTERN='std(out|err)\.write'
readonly FILE_WRITE_CALLS_PATTERN='write|rename|unlink|remove\(|rmtree|shutil\.|copyFile|truncate|open\([^)]*['"'"'"]\+?[wax>]'

deny() {
	echo "$1" >&2
	exit "$EXIT_DENY"
}

# Drops heredoc bodies, so text fed to a command is not mistaken for shell syntax.
strip_heredoc_bodies() {
	awk '
		in_body {
			line = $0
			if (strip_tabs) sub(/^\t+/, "", line)
			if (line == delimiter) in_body = 0
			next
		}
		{
			print
			probe = $0
			gsub(/<<</, "", probe)
			if (match(probe, /<<-?[ \t]*["'"'"']?[A-Za-z_][A-Za-z0-9_]*/)) {
				delimiter = substr(probe, RSTART, RLENGTH)
				strip_tabs = (delimiter ~ /^<<-/)
				gsub(/^<<-?[ \t]*["'"'"']?/, "", delimiter)
				in_body = 1
			}
		}
	'
}

# Removes the quotes and, inside them, every character that could read as shell syntax,
# so a commit message like "a -> b" is not taken for a redirect.
flatten_quotes() {
	awk '
		BEGIN { keep = "[A-Za-z0-9_$/{}.:@%+=,~*-]" }
		{
			out = ""
			for (i = 1; i <= length($0); i++) {
				c = substr($0, i, 1)
				if (quote == "" && c == "\\") { out = out c substr($0, i + 1, 1); i++; continue }
				if (quote == "" && (c == "\"" || c == "'"'"'")) { quote = c; continue }
				if (quote != "" && c == quote) { quote = ""; continue }
				if (quote == "\"" && c == "\\") { i++; c = substr($0, i, 1) }
				if (quote != "" && c !~ keep) continue
				out = out c
			}
			print out
		}
	'
}

# Prints one simple command per line. File-descriptor duplication such as 2>&1 is dropped
# first, so its & does not split the command.
split_commands() {
	sed -E \
		-e 's/[0-9]*>&[0-9-]+//g' \
		-e 's/&>/>/g' \
		-e 's/>&/> /g' \
		-e 's/\|&/|/g' |
		tr ';&|()`' '\n'
}

is_allowed_write_target() {
	local target="$1"

	# shellcheck disable=SC2016 # matches the unexpanded variable name the agent wrote
	case "$target" in
	/dev/null | /dev/stdout | /dev/stderr | /tmp/* | '$TMPDIR'* | '${TMPDIR}'*) return 0 ;;
	esac
	[[ -n "${TMPDIR:-}" && "$target" == "$TMPDIR"* ]]
}

check_redirects() {
	local command="$1"
	local rest="$command"
	local redirect_pattern='>+[[:space:]]*([^[:space:]<>]+)(.*)'

	while [[ "$rest" =~ $redirect_pattern ]]; do
		if ! is_allowed_write_target "${BASH_REMATCH[1]}"; then
			deny "Shell redirects may write only to /dev/null, /tmp, or \$TMPDIR: $USE_EDIT_TOOLS."
		fi
		rest="${BASH_REMATCH[2]}"
	done
}

check_tee() {
	local arg

	for arg in "$@"; do
		if [[ "$arg" == -* ]]; then
			continue
		fi
		if ! is_allowed_write_target "$arg"; then
			deny "tee may write only to /dev/null, /tmp, or \$TMPDIR: $USE_EDIT_TOOLS."
		fi
	done
}

check_in_place_edit() {
	local program="$1"
	local in_place_pattern="$2"
	shift 2
	local arg

	for arg in "$@"; do
		if [[ "$arg" =~ $in_place_pattern ]]; then
			deny "$program must not edit files in place: $USE_EDIT_TOOLS."
		fi
	done
}

check_awk() {
	local previous=""
	local arg

	for arg in "$@"; do
		if [[ "$previous" == "-i" && "$arg" == "inplace" ]]; then
			deny "awk must not edit files in place: $USE_EDIT_TOOLS."
		fi
		previous="$arg"
	done
}

is_whole_tree_path() {
	case "$1" in
	. | ./ | :/ | '*' | ':/*') return 0 ;;
	esac
	return 1
}

is_short_flag_cluster_with() {
	local flag="$1"
	local arg="$2"

	[[ "$arg" =~ ^-[A-Za-z]*${flag}[A-Za-z]*$ ]]
}

lowercase() {
	tr '[:upper:]' '[:lower:]' <<<"$1"
}

check_hook_bypass_flag() {
	local arg

	for arg in "$@"; do
		if [[ "$arg" == "--no-verify" ]]; then
			deny "Git hooks must not be skipped: fix what the hook reports instead of passing --no-verify."
		fi
	done
}

check_commit() {
	local takes_value=0
	local arg

	for arg in "$@"; do
		if ((takes_value)); then
			takes_value=0
			continue
		fi
		case "$arg" in
		-m | -F | -C | -c | -t) takes_value=1 ;;
		--*) ;;
		-*) if is_short_flag_cluster_with n "$arg"; then
			deny "Git hooks must not be skipped: fix what the hook reports instead of passing commit -n."
		fi ;;
		esac
	done
}

check_config() {
	local reads=0
	local sets_hooks_path=0
	local arg

	for arg in "$@"; do
		case "$(lowercase "$arg")" in
		--get* | get | --list | -l) reads=1 ;;
		core.hookspath*) sets_hooks_path=1 ;;
		esac
	done
	if ((sets_hooks_path && !reads)); then
		deny "core.hooksPath must not be changed: it would disable the repo's git hooks."
	fi
}

current_branch() {
	git symbolic-ref --short -q HEAD || true
}

is_force_push_arg() {
	local arg="$1"

	case "$arg" in
	--force | --force-with-lease*) return 0 ;;
	--*) return 1 ;;
	+*) return 0 ;;
	-*) is_short_flag_cluster_with f "$arg" ;;
	*) return 1 ;;
	esac
}

is_protected_refspec() {
	case "$1" in
	--all | --mirror | "$PROTECTED_BRANCH" | *:"$PROTECTED_BRANCH" | *refs/heads/"$PROTECTED_BRANCH") return 0 ;;
	esac
	return 1
}

check_push() {
	local positional=0
	local pushes_head=0
	local arg

	for arg in "$@"; do
		if is_force_push_arg "$arg"; then
			deny "Force-pushing is not allowed: push new commits instead, or ask the user to force-push."
		fi
		if is_protected_refspec "$arg"; then
			deny "Pushing to $PROTECTED_BRANCH is not allowed: push a branch and open a PR."
		fi
		if [[ "$arg" == HEAD ]]; then
			pushes_head=1
		fi
		if [[ "$arg" != -* ]]; then
			positional=$((positional + 1))
		fi
	done
	if [[ "$(current_branch)" == "$PROTECTED_BRANCH" ]] && ((positional < 2 || pushes_head)); then
		deny "Pushing to $PROTECTED_BRANCH is not allowed: create a branch, push it, and open a PR."
	fi
}

deny_destructive() {
	deny "\`git $1\` discards work: ask the user to run it."
}

check_reset() {
	local arg

	for arg in "$@"; do
		if [[ "$arg" == "--hard" ]]; then
			deny_destructive "reset --hard"
		fi
	done
}

check_clean() {
	local arg

	for arg in "$@"; do
		if [[ "$arg" == "--force" ]] || is_short_flag_cluster_with f "$arg"; then
			deny_destructive "clean -f"
		fi
	done
}

check_checkout() {
	local arg

	for arg in "$@"; do
		if is_whole_tree_path "$arg"; then
			deny_destructive "checkout -- $arg"
		fi
	done
}

check_restore() {
	local staged_only=0
	local whole_tree=""
	local arg

	for arg in "$@"; do
		case "$arg" in
		--staged | -S) staged_only=1 ;;
		--worktree | -W) staged_only=0 && break ;;
		esac
		if is_whole_tree_path "$arg"; then
			whole_tree="$arg"
		fi
	done
	if [[ -n "$whole_tree" ]] && ((!staged_only)); then
		deny_destructive "restore $whole_tree"
	fi
}

check_stash() {
	case "${1:-}" in
	drop | clear) deny_destructive "stash $1" ;;
	esac
}

check_git_subcommand() {
	local subcommand="$1"
	shift

	case "$subcommand" in
	commit) check_commit "$@" ;;
	config) check_config "$@" ;;
	push) check_push "$@" ;;
	reset) check_reset "$@" ;;
	clean) check_clean "$@" ;;
	checkout) check_checkout "$@" ;;
	restore) check_restore "$@" ;;
	stash) check_stash "$@" ;;
	esac
}

check_git_option_value() {
	if [[ "$(lowercase "$1")" == core.hookspath* ]]; then
		deny "core.hooksPath must not be overridden: it would disable the repo's git hooks."
	fi
}

check_git() {
	check_hook_bypass_flag "$@"
	while (($# > 0)); do
		case "$1" in
		-c)
			check_git_option_value "${2:-}"
			shift 2 || shift
			;;
		-C | --git-dir | --work-tree | --namespace) shift 2 || shift ;;
		-*) shift ;;
		*)
			check_git_subcommand "$@"
			return
			;;
		esac
	done
}

check_program() {
	local program="$1"
	shift

	case "$program" in
	git) check_git "$@" ;;
	sed) check_in_place_edit sed '^(-[nErsuz]*i|--in-place)' "$@" ;;
	perl) check_in_place_edit perl '^-[nplaswWtTuUcvhE]*i' "$@" ;;
	awk | gawk) check_awk "$@" ;;
	tee) check_tee "$@" ;;
	esac
}

check_command() {
	local command="$1"
	local -a words
	local i

	check_redirects "$command"
	read -r -a words <<<"$command"
	for ((i = 0; i < ${#words[@]}; i++)); do
		check_program "$(basename -- "${words[i]}")" "${words[@]:i+1}"
	done
}

uses_interpreter() {
	local commands="$1"
	local -a words
	local command word

	while IFS= read -r command; do
		read -r -a words <<<"$command"
		for word in ${words[@]+"${words[@]}"}; do
			if [[ "$(basename -- "$word")" =~ $INTERPRETERS_PATTERN ]]; then
				return 0
			fi
		done
	done <<<"$commands"
	return 1
}

# Inspects the raw text, heredoc bodies included, because that is where the script is.
check_interpreter_writes() {
	local raw="$1"
	local commands="$2"

	if ! uses_interpreter "$commands"; then
		return
	fi
	if sed -E "s/$STREAM_WRITES_PATTERN//g" <<<"$raw" | grep -Eq "$FILE_WRITE_CALLS_PATTERN"; then
		deny "Interpreters must not write files: $USE_EDIT_TOOLS."
	fi
}

main() {
	local raw commands command

	raw="$(cat)"
	commands="$(strip_heredoc_bodies <<<"$raw" | flatten_quotes | split_commands)"
	check_interpreter_writes "$raw" "$commands"
	while IFS= read -r command; do
		check_command "$command"
	done <<<"$commands"
	exit "$EXIT_PASS"
}

main "$@"
