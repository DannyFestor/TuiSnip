#!/bin/sh
signature=$(basename "$0")
while [ $# -gt 1 ]; do
	signature="$signature $1"
	shift
done
printf '%s' "$signature" > "$1"
if [ -n "$RECORD_PATH" ]; then
	printf '%s' "$1" > "$RECORD_PATH"
fi
