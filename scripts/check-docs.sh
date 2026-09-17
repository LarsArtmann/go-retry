#!/usr/bin/env bash
# One command for the documentation battery: repo-doc guard tests, dprint
# format check, and CHANGELOG compare links. Doc edits skip the heavy Go
# suite, and the 2026-09-17 rendering catastrophe shipped exactly through
# that gap — so the doc gates must be this cheap, unskippable command.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

echo '== doc guard tests (strike rendering, cell code spans, archive verdicts, index)'
# -count=1 defeats the test cache: the fixtures are data files, which are
# not part of the cache key, so a cached PASS can hide a fresh regression.
go test -count=1 -run 'TestMarkdownStrikethroughSpansRender|TestMarkdownTableCellsCloseCodeSpans|TestArchivedReportItemsCarryVerdicts|TestStatusIndexCoversArchive' .

echo '== dprint (markdown/JSON/YAML format)'
nix run nixpkgs#dprint -- check

echo '== CHANGELOG compare links'
./scripts/check-compare-links.sh

echo 'docs battery: all green'
