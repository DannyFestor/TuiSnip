#!/bin/sh
cat >"$0.stdin"
printf '%s' "$*" >"$0.args"
printf '%s' "${LC_CTYPE:-}" >"$0.lc_ctype"
