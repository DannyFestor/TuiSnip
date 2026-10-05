#!/bin/sh
printf '%s' "$*" >"$0.args"
printf '%s' "${LC_CTYPE:-}" >"$0.lc_ctype"
cat "$0.clipboard"
