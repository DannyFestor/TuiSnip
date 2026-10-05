#!/bin/sh
private_dir=$(find "$(dirname "$1")" -prune -perm 0700)
private_file=$(find "$1" -prune -perm 0600)
if [ -n "$private_dir" ] && [ -n "$private_file" ]; then
	printf 'private' > "$1"
else
	printf 'exposed' > "$1"
fi
