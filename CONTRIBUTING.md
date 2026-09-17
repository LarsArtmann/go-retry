# Contributing

Thanks for your interest in contributing to `go-retry`!

## Prerequisites

- **Go 1.26** or later (`go.mod` pins the toolchain version)
- **[golangci-lint](https://golangci-lint.run/)** v2.x (config is committed at
  [`.golangci.yml`](.golangci.yml))

There is no `justfile`, `Makefile`, or `flake.nix` — `go` and `golangci-lint`
are the only tools required.

## Development commands

```bash
go test ./... -race        # tests (always with -race; backoff uses math/rand/v2)
golangci-lint run ./...    # lint (uses the committed .golangci.yml)
go vet ./...               # vet
go -C tools vet ./...      # vet the nested tools module (invisible to root ./...)
./scripts/check-docs.sh    # doc battery: guard tests + dprint + compare-links
# Development tools (actionlint, govulncheck) are version-pinned in the
# nested tools module — install them once per bump, then run by name:
go -C tools install github.com/rhysd/actionlint/cmd/actionlint golang.org/x/vuln/cmd/govulncheck
actionlint -verbose        # workflow schema check
```

The workflow schema check also runs in CI (first step of the `lint` job,
built from the same `tools/go.mod` pin), so invalid workflow YAML fails fast
on the runner instead of surfacing as mysterious job failures.

### Session ritual (maintainers)

`AGENTS.md` ends with a **Session Ritual** checklist — the self-checks every
change set should pass before claiming done (gate order incl.
`golangci-lint config verify`, proving new guard tests fail on their drift,
hash verification for cited evidence, raw test summaries over filtered
tails). Read it once; run it always. Two rituals from it deserve emphasis:

- **Doc battery** — after any documentation edit, run
  [`scripts/check-docs.sh`](scripts/check-docs.sh); it chains the repo-doc
  guard tests (strikethrough rendering, archive verdicts, index consistency),
  the dprint format check, and
  [`scripts/check-compare-links.sh`](scripts/check-compare-links.sh), which
  fails on a malformed or poisoned compare link and is part of the release
  ritual.
- **Annotate as you land** — when a change completes an item from a plan or
  status report, the row gets its `done at <hash>` verdict **in the same
  change**, citing only the final hash of the work (never an intermediate
  one). Re-check `git status` immediately before `git add` — the auto-commit
  daemon races explicit commits, and the daemon may also reformat markdown
  (dprint realigns tables) mid-session: re-`view` a file before editing it
  after any formatter or daemon pass.

### Coverage

`reports/` is gitignored (`.gitignore`). Regenerate coverage locally:

```bash
go test ./... -race -coverprofile=reports/coverage.out \
  && go tool cover -func=reports/coverage.out
```

Current statement coverage is **100%**. CI enforces a **95% floor**
(`.github/workflows/ci.yml`, `coverage` job) — keep new code at or above
100% where practical; the floor exists so a hard-to-test edge never blocks
a fix.

### Fuzzing

`FuzzComputeDelayNeverPanics` hardens the delay computation against panics
and negative durations. Run a bounded campaign locally:

```bash
go test -run '^$' -fuzz=FuzzComputeDelayNeverPanics -fuzztime=5m ./...
```

The seven `f.Add` seeds are mirrored in
`testdata/fuzz/FuzzComputeDelayNeverPanics/` — keep seeds and corpus files in
sync (a plain `go test -run '^FuzzComputeDelayNeverPanics$'` exercises the
committed corpus without fuzzing). A daily scheduled workflow
(`.github/workflows/fuzz.yml`) runs a 30-minute campaign in CI; new crashers
land in the committed corpus, not just in seeds.

## Formatting

Formatting is enforced by the lint gate, not by discipline — a drifted file
fails `golangci-lint run` / CI:

- **Go** — the `.golangci.yml` `formatters` section enables `gci`, `goimports`,
  `gofumpt`, and `golines` (max line length 120). Run `golangci-lint run
  --fix` (or let the editor's golangci integration) to auto-format.
- **Markdown / JSON / YAML / Dockerfile** — [dprint](https://dprint.dev) with
  the committed `dprint.json` (the markdown plugin maintains existing line
  wrapping; `CHANGELOG.md` is excluded). Tabs for Go files, 2 spaces for
  YAML/JSON (`.editorconfig`). dprint is a Rust binary, so it is pinned where
  it runs in CI (the `dprint/check` step in `.github/workflows/ci.yml`, with
  an attestation-verified download) rather than in `tools/go.mod`:

  ```bash
  nix run nixpkgs#dprint -- check   # gate: fails on drift
  nix run nixpkgs#dprint -- fmt     # fix drift in place
  ```

  dprint's markdown plugin normalizes emphasis to `_underscores_` (not
  `*asterisks*`) and realigns table columns — when hand-edits introduce
  `*emphasis*` or ragged tables, `dprint fmt` rewrites them; that drift in a
  diff you didn't author is dprint at work, not vandalism.

## Lint policy

The committed [`.golangci.yml`](.golangci.yml) enables the standard default
linters plus ~100 extra ones (`gosec`, `mnd`, `exhaustruct_v5`, `errorlint`,
and many more — see the `enable` list). CI pins the same golangci-lint version
the repo develops against (the version pinned in `.github/workflows/ci.yml`). The following in-source `//nolint:`
markers are **deliberate** — do not "fix" them by removing the marker or
restructuring the code:

- `config.go` (`DefaultConfig`) — `//nolint:exhaustruct_v5`: `OnRetry` and
  `OnExhausted` are intentionally omitted (they are optional callbacks).
- `retry.go` (`computeDelay`) — `//nolint:gosec`: the jitter source uses a
  weak RNG on purpose; this is jitter, not security-sensitive randomness
  (`mnd` does not fire — the divisor `2` is in its ignored-numbers).
- `retry_test.go` (`TestDo_DoesNotRetryNonRetryableError`) —
  `//nolint:errorlint`: the whole point of the assertion is the identity
  comparison `err != rejection`, which `errorlint` would otherwise flag.

`mnd` and `exhaustruct_v5` are excluded from `*_test.go` (see
`.golangci.yml`), where partial struct literals and bare scalars are
legitimate.

## Testing conventions

- External test package (`package retry_test`) — exercise the public API only.
- Every test calls `t.Parallel()`.
- Counters use `sync/atomic` (`atomic.Int32`), not mutexes.
- Keep delays millisecond-scale (see the `fastConfig()` helper) so the suite
  stays fast; the two context-ending tests (cancel + deadline) are the only
  deliberate exceptions — they use `5s` delays so the context end fires
  during the wait.

See [`AGENTS.md`](AGENTS.md) for the deeper architectural context (the
`error-family` dependency, the no-CQRS/no-OTel boundary, control flow of `Do`).

## How to contribute

1. Fork the repository and create a feature branch.
2. Make your change with tests; keep coverage at 100%.
3. Ensure `go test ./... -race`, `golangci-lint run ./...`, and `go vet ./...`
   all pass (CI runs all three plus the 95% coverage floor).
4. Keep the [no-CQRS / no-OTel boundary](doc.go) — features needing CQRS message
   types or OpenTelemetry belong in `go-cqrs-lite/middleware/v4`, not here.
5. Submit a pull request.

## Reporting issues

Please use [GitHub Issues](https://github.com/larsartmann/go-retry/issues) to
report bugs or request features.

## CHANGELOG entries

User-visible changes (API, behavior, error messages, guarantees, CI-visible
tooling) get a `[Unreleased]` entry. Pure docs fixes (typos, reformatting)
and internal status/planning files do not. Doc changes that alter documented
behavior or guarantees (README snippets, godoc examples) **do** get an entry.

## Release notes

Bodies live **GitHub-only** (decided 2026-09-13): composed at release time
from the matching `CHANGELOG.md` section — a curated, user-focused summary,
not a copy. The shape that worked for v0.6.0/v0.6.1:

1. `# vX.Y.Z` heading.
2. One-line classification opener: what kind of release, "no API changes —
   everything from vX.Y.Z keeps working", and the exact `go get` command.
3. Sections by audience, most important first (`User-visible`, `Documentation`,
   `Guarantees pinned by new tests`, `CI & tooling`) — only the sections that
   have content.
4. `**Full changelog:**` compare link (`vPREV...vNEW`).

Pre-tag ritual additions to the `go-release` Phase 4 gate: local
`govulncheck ./...` (first exercised at v0.6.1: no vulnerabilities),
`go mod tidy` producing no diff at the tag, and
`./scripts/check-compare-links.sh` after any CHANGELOG link edit
(proven to fail on a poisoned link before it counted as a guard).

Post-push verification ritual: proxy `go list -m -versions` shows the tag;
clean-room `go get` in a scratch module; the canonical
`pkg.go.dev/github.com/larsartmann/go-retry` page renders the new version
(the `/fetch` trigger URL may 404 while indexing — the page is the truth);
tag-CI run green.
