#!/usr/bin/env bash
# Git pre-push hook: changes reach main only through a squash-merged PR.
set -euo pipefail

protected_ref="refs/heads/main"

while read -r _local_ref _local_oid remote_ref _remote_oid; do
	if [[ "$remote_ref" == "$protected_ref" ]]; then
		echo "Refusing to push to main. Push a branch and open a PR." >&2
		exit 1
	fi
done
