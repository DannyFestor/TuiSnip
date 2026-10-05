#!/usr/bin/env bash
# Reads Go package paths on stdin and prints, sorted, those not inside another listed package.
# A tool that recurses into the directory it's given, like gremlins, already covers the rest.
#
# Usage: changed-packages.sh | outermost-packages.sh
set -euo pipefail

# Sorting puts every package after the packages it's inside, so each is checked against
# all of its listed ancestors.
LC_ALL=C sort -u | awk '
	{
		for (root in outermost) {
			if (index($0, root "/") == 1) {
				next
			}
		}
		outermost[$0]
		print
	}
'
