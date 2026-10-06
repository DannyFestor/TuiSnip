#!/usr/bin/env bash
# Pre-command guard: the shell command arrives on stdin, never argv, so heredocs and quotes
# arrive unmangled. Exits 0 to pass or 2 to deny, with the reason on stderr.
#
# This is a guardrail against mistakes, not a security boundary: a command nested in
# `sh -c '...'` is not inspected.
set -euo pipefail

# shellcheck source=scripts/agent/lib.sh
source "$(dirname "$0")/lib.sh"

# Shell syntax is ASCII, so bytes are enough, and under a UTF-8 locale macOS awk fails on
# the partial character flatten_quotes takes when it walks a multibyte character byte by byte.
export LC_ALL=C

readonly PROTECTED_BRANCH="main"
readonly USE_EDIT_TOOLS="use the edit tools instead, so the post-edit format and vet run"
readonly INTERPRETERS_PATTERN='^(python[0-9.]*|node|ruby|perl|php|deno|bun)$'
# Standard-stream writes are output, not file writes.
readonly STREAM_WRITES_PATTERN='std(out|err)\.write'
# lefthook reads these to switch itself off, skip jobs, or load another config.
readonly LEFTHOOK_OVERRIDE_PATTERN='^(LEFTHOOK|LEFTHOOK_EXCLUDE|LEFTHOOK_CONFIG)='
readonly FILE_WRITE_CALLS_PATTERN='write|rename|unlink|remove\(|rmtree|shutil\.|copyFile|truncate|open\([^)]*['"'"'"]\+?[wax>]'
readonly ASSIGNMENT_PATTERN='^[A-Za-z_][A-Za-z0-9_]*='
# These run the program named after them, so the word after them is not the program.
readonly WRAPPERS_PATTERN='^(env|command|exec|sudo|nohup|nice|time|xargs)$'
readonly REMOTE_PATH_PATTERN='^[^/]*:'
readonly GIT_APPLY_READ_ONLY_PATTERN='^--(check|stat|numstat|summary)$'
readonly PERL_RECORD_SEPARATOR_PATTERN='^([xX][0-9A-Fa-f]*|[0-7]*)'
readonly OCTAL_DIGITS_PATTERN='^[0-7]*'
readonly PERL_UNICODE_FEATURES_PATTERN='^[0-9IOEioSADLa]*'

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
# so a commit message like "a -> b" is not taken for a redirect. A quoted span left empty
# becomes a placeholder, so `-m "" -n` still reads -n as a switch rather than as -m's value.
flatten_quotes() {
	awk '
		BEGIN { keep = "[A-Za-z0-9_$/{}.:@%+=,~*-]"; placeholder = "_" }
		{
			out = ""
			for (i = 1; i <= length($0); i++) {
				c = substr($0, i, 1)
				if (quote == "" && c == "\\") { out = out c substr($0, i + 1, 1); i++; continue }
				if (quote == "" && (c == "\"" || c == "'"'"'")) { quote = c; opened_at = length(out); continue }
				if (quote != "" && c == quote) {
					if (length(out) == opened_at) out = out placeholder
					quote = ""
					continue
				}
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

is_under() {
	local path="$1"
	local dir="$2"

	[[ "$path" == "$dir" || "$path" == "$dir/"* ]]
}

# A relative path resolves against the working directory, as it does for the command. The git
# dir belongs to no checkout, but a write there can still replace the git hooks.
is_inside_clone() {
	local path

	path="$(canonical_path "$1")"
	checkout_root "$path" >/dev/null || is_under "$path" "$(clone_git_dir)"
}

check_write_destination() {
	local program="$1"
	local destination="$2"

	if is_allowed_write_target "$destination" || ! is_inside_clone "$destination"; then
		return
	fi
	deny "$program must not write into the repo: $USE_EDIT_TOOLS."
}

last_operand() {
	local last=""
	local arg

	for arg in "$@"; do
		if [[ "$arg" != -* ]]; then
			last="$arg"
		fi
	done
	echo "$last"
}

target_directory() {
	local previous=""
	local arg

	for arg in "$@"; do
		case "$previous" in
		-t | --target-directory) echo "$arg" && return ;;
		esac
		case "$arg" in
		--target-directory=*) echo "${arg#*=}" && return ;;
		-t?*) echo "${arg#-t}" && return ;;
		esac
		previous="$arg"
	done
}

# cp, mv, install, and ln write to the -t directory when one is given, else to the last operand.
check_copy() {
	local program="$1"
	shift
	local destination

	destination="$(target_directory "$@")"
	if [[ -z "$destination" ]]; then
		destination="$(last_operand "$@")"
	fi
	if [[ -n "$destination" ]]; then
		check_write_destination "$program" "$destination"
	fi
}

check_rsync() {
	local destination

	destination="$(last_operand "$@")"
	if [[ -n "$destination" && ! "$destination" =~ $REMOTE_PATH_PATTERN ]]; then
		check_write_destination rsync "$destination"
	fi
}

check_dd() {
	local arg

	for arg in "$@"; do
		if [[ "$arg" == of=* ]]; then
			check_write_destination dd "${arg#of=}"
		fi
	done
}

# patch writes to the -o file when one is given, else beside the files it patches, which sit
# under the -d directory or the working directory.
check_patch() {
	local output=""
	local directory="."
	local previous=""
	local arg

	for arg in "$@"; do
		case "$arg" in
		--dry-run) return ;;
		--output=*) output="${arg#*=}" ;;
		--directory=*) directory="${arg#*=}" ;;
		esac
		case "$previous" in
		-o) output="$arg" ;;
		-d) directory="$arg" ;;
		esac
		previous="$arg"
	done
	check_write_destination patch "${output:-$directory}"
}

is_command_prefix() {
	local word="$1"

	[[ "$word" =~ $ASSIGNMENT_PATTERN || "$(basename -- "$word")" =~ $WRAPPERS_PATTERN || "$word" == -* ]]
}

program_index() {
	local -a words=("$@")
	local i

	for ((i = 0; i < ${#words[@]}; i++)); do
		if ! is_command_prefix "${words[i]}"; then
			echo "$i"
			return
		fi
	done
}

# Only the program in command position counts, so `git mv` and `go install` pass.
check_copy_programs() {
	local -a words=("$@")
	local i program

	i="$(program_index "$@")"
	if [[ -z "$i" ]]; then
		return
	fi
	program="$(basename -- "${words[i]}")"
	case "$program" in
	cp | mv | install | ln) check_copy "$program" "${words[@]:i+1}" ;;
	rsync) check_rsync "${words[@]:i+1}" ;;
	dd) check_dd "${words[@]:i+1}" ;;
	patch) check_patch "${words[@]:i+1}" ;;
	esac
}

is_short_flag_cluster() {
	[[ "$1" == -[!-]* ]]
}

argument_matching() {
	local pattern="$1"
	local rest="$2"

	[[ "$rest" =~ $pattern ]]
	printf '%s' "${BASH_REMATCH[0]}"
}

# Switches that take a value from the rest of their cluster, or from the next word when the
# cluster ends with them. sed -l is left out because BSD sed takes no argument for it, and
# reading -li as -l -i errs on the side of denying.
value_switches() {
	case "$1" in
	git-commit) printf '%s' mFCct ;;
	git-push) printf '%s' o ;;
	git-clean) printf '%s' e ;;
	sed) printf '%s' ef ;;
	perl) printf '%s' eE ;;
	esac
}

# These take a value only from the rest of their cluster, so a bare one leaves the next word
# alone.
perl_attached_argument() {
	case "$1" in
	0) argument_matching "$PERL_RECORD_SEPARATOR_PATTERN" "$2" ;;
	l) argument_matching "$OCTAL_DIGITS_PATTERN" "$2" ;;
	C) argument_matching "$PERL_UNICODE_FEATURES_PATTERN" "$2" ;;
	d | D | F | I | m | M | x) printf '%s' "$2" ;;
	esac
}

attached_argument() {
	local program="$1"
	local switch="$2"
	local rest="$3"

	if [[ "$(value_switches "$program")" == *"$switch"* ]]; then
		printf '%s' "$rest"
	elif [[ "$program" == perl ]]; then
		perl_attached_argument "$switch" "$rest"
	fi
}

# Walks the cluster switch by switch, skipping each switch's argument, so -0pi is caught and
# the i in -Mstrict is not. Succeeds when the last switch takes the next word as its value.
cluster_switches() {
	local program="$1"
	local cluster="${2#-}"
	local switch argument

	while [[ -n "$cluster" ]]; do
		switch="${cluster:0:1}"
		printf '%s' "$switch"
		cluster="${cluster:1}"
		argument="$(attached_argument "$program" "$switch" "$cluster")"
		cluster="${cluster:${#argument}}"
	done
	[[ -z "$argument" && "$(value_switches "$program")" == *"$switch"* ]]
}

short_switches() {
	local program="$1"
	shift
	local next_is_value=0
	local arg

	for arg in "$@"; do
		if ((next_is_value)); then
			next_is_value=0
		elif is_short_flag_cluster "$arg" && cluster_switches "$program" "$arg"; then
			next_is_value=1
		fi
	done
}

# Reads clusters through the program's value table, so the n in -m"Add login" is not -n.
has_short_switch() {
	local program="$1"
	local switches="$2"
	shift 2

	[[ "$(short_switches "$program" "$@")" == *["$switches"]* ]]
}

has_arg_matching() {
	local pattern="$1"
	shift
	local arg

	for arg in "$@"; do
		# shellcheck disable=SC2053 # the pattern is a glob on purpose
		if [[ "$arg" == $pattern ]]; then
			return 0
		fi
	done
	return 1
}

# BSD sed edits in place with -I as well as -i.
is_in_place_edit() {
	local program="$1"
	shift

	case "$program" in
	sed) has_arg_matching '--in-place*' "$@" || has_short_switch sed iI "$@" ;;
	perl) has_short_switch perl i "$@" ;;
	esac
}

check_in_place_edit() {
	if is_in_place_edit "$@"; then
		deny "$1 must not edit files in place: $USE_EDIT_TOOLS."
	fi
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

check_lefthook_override() {
	if [[ "$1" =~ $LEFTHOOK_OVERRIDE_PATTERN ]]; then
		deny "Git hooks must not be skipped: fix what the hook reports instead of setting ${BASH_REMATCH[1]}."
	fi
}

check_commit() {
	if has_short_switch git-commit n "$@"; then
		deny "Git hooks must not be skipped: fix what the hook reports instead of passing commit -n."
	fi
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

deny_force_push() {
	deny "Force-pushing is not allowed: push new commits instead, or ask the user to force-push."
}

is_force_push_arg() {
	case "$1" in
	--force | --force-with-lease* | +*) return 0 ;;
	esac
	return 1
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

	if has_short_switch git-push f "$@"; then
		deny_force_push
	fi
	for arg in "$@"; do
		if is_force_push_arg "$arg"; then
			deny_force_push
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
	if has_short_switch git-clean f "$@" || has_arg_matching --force "$@"; then
		deny_destructive "clean -f"
	fi
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

# Even --cached writes, because it changes what the next commit holds without the edit hooks.
check_apply() {
	local read_only=0
	local arg

	for arg in "$@"; do
		case "$arg" in
		--apply) read_only=0 && break ;;
		esac
		if [[ "$arg" =~ $GIT_APPLY_READ_ONLY_PATTERN ]]; then
			read_only=1
		fi
	done
	if ((!read_only)); then
		deny "git apply must not write into the repo: $USE_EDIT_TOOLS."
	fi
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
	apply) check_apply "$@" ;;
	esac
}

check_git_option_value() {
	if [[ "$(lowercase "$1")" == core.hookspath* ]]; then
		deny "core.hooksPath must not be overridden: it would disable the repo's git hooks."
	fi
}

# A directory that doesn't exist yet leaves the command where it is, so it is judged there.
enter_directory() {
	if [[ -d "$1" ]]; then
		cd "$1"
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
		-C)
			enter_directory "${2:-}"
			shift 2 || shift
			;;
		--git-dir | --work-tree | --namespace) shift 2 || shift ;;
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
	# The subshell keeps a git -C directory from moving the commands after it.
	git) (check_git "$@") || exit "$?" ;;
	sed | perl) check_in_place_edit "$program" "$@" ;;
	awk | gawk) check_awk "$@" ;;
	tee) check_tee "$@" ;;
	esac
}

# The commands after a cd run where it took them, which can be a linked worktree (#149).
follow_cd() {
	local -a words=("$@")
	local i

	i="$(program_index "$@")"
	if [[ -n "$i" && "${words[i]}" == cd ]]; then
		enter_directory "$(last_operand "${words[@]:i+1}")"
	fi
}

check_command() {
	local command="$1"
	local -a words
	local i

	check_redirects "$command"
	read -r -a words <<<"$command"
	for ((i = 0; i < ${#words[@]}; i++)); do
		check_lefthook_override "${words[i]}"
		check_program "$(basename -- "${words[i]}")" "${words[@]:i+1}"
	done
	check_copy_programs ${words[@]+"${words[@]}"}
	follow_cd ${words[@]+"${words[@]}"}
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
