#!/usr/bin/env bash
# Git pre-commit hook: a commit made on main would have to be moved to a branch before it
# could reach main through a squash-merged PR.
set -euo pipefail

protected_branch="main"

# A detached HEAD, as during a rebase, has no branch to protect.
current_branch="$(git symbolic-ref --short -q HEAD || true)"
if [[ "$current_branch" == "$protected_branch" ]]; then
	echo "Refusing to commit to main. Create a branch first: git switch -c <branch>." >&2
	exit 1
fi
