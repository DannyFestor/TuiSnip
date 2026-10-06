#!/bin/sh
printf 'discarded' > "$1"
printf '%s' "$1" > "$RECORD_PATH"
exit 1
