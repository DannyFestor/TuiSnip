#!/usr/bin/env bash
# Prints every path an OpenCode apply_patch touches, one per line, read from the patch on
# stdin. It lives here so the TypeScript plugin holds no logic.
set -euo pipefail

sed -n -E 's/^\*\*\* (Add File|Update File|Move to|Delete File): (.+)$/\2/p'
