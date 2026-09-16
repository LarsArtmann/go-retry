#!/usr/bin/env bash
# Validates every CHANGELOG compare-link tag pair against git tags.
# [0.1.0] links to releases/tag (no prior tag); everything else compares
# vPREV...vNEW, and both tags must exist locally.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

fail=0

while IFS= read -r line; do
	label=$(printf '%s' "$line" | sed -E 's/^\[([^]]+)\]:.*/\1/')
	url=$(printf '%s' "$line" | sed -E 's/^[^:]+: *(.*)$/\1/')

	case "$url" in
	*/releases/tag/*)
		tag=${url##*/}
		if ! git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
			echo "MISSING tag $tag (for [$label])"
			fail=1
		fi
		;;
	*/compare/*)
		pair=${url##*compare/}
		prev=${pair%%...*}
		next=${pair##*...}
		for tag in "$prev" "$next"; do
			if [ "$tag" = "HEAD" ]; then
				continue
			fi
			if ! git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
				echo "MISSING tag $tag (for [$label])"
				fail=1
			fi
		done
		;;
	HEAD)
		: # [Unreleased] compares against HEAD; nothing to verify
		;;
	*)
		echo "UNRECOGNIZED link for [$label]: $url"
		fail=1
		;;
	esac
done < <(sed -n '/^\[/s/^/[OK] /p' CHANGELOG.md | sed 's/^\[OK\] //' | grep -E '^\[[^]]+\]: *https')

if [ "$fail" -eq 0 ]; then
	echo "compare links: all tag pairs resolve"
fi

exit "$fail"
