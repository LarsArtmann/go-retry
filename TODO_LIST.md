# TODO List

Short-term, **actionable** work for `go-retry`. Each item is bounded and cites
its evidence. This file lists open work only — completed items move to
`CHANGELOG.md`, long-term/unbounded ideas live in `ROADMAP.md`.

Priority uses a simple Pareto ranking: **P1** = high impact, do first;
**P2** = valuable but not blocking; **P3** = nice-to-have polish.

---

## P1 — Correctness of public-facing docs

### T1. Rewrite `README.md` into an accurate package description

The current README is a generic project template that actively misleads:

- Title is literally `# .` and description is "A Go project."
  (`README.md:1-5`).
- Install command is fake: `go get github.com/username/.`
  (`README.md:9-11`) — the real module path is `github.com/larsartmann/go-retry`
  (`go.mod:1`), and **no git remote exists yet** (`git remote -v` is empty).
- "Development" section references `just build` / `just test` / `just lint`
  (`README.md:29-38`) but there is **no `justfile`, `flake.nix`, or `Makefile`**
  in the repo. The real commands are `go test ./... -race` and
  `golangci-lint run ./...` (also documented in `AGENTS.md` → Commands).

**Fix:** describe what the package is (dependency-light retry loop with
exponential backoff + jitter; the no-CQRS/no-OTel core), the v0.1.0 API surface
(`Do`, `Config`, `Backoff`), a minimal usage example, and the real install/
dev commands. If no remote will exist soon, drop the `go get` line rather than
ship a broken one.

### T2. Resolve the `CONTRIBUTING.md` lint-config gap

`CONTRIBUTING.md:16-17` instructs contributors to run
`golangci-lint run ./...`, but **no `.golangci.yml` is committed**, so the
behavior contributors get is whatever their local defaults are — which may flag
or miss things inconsistently. Pick one: (a) commit a minimal `.golangci.yml`
that pins the linters the project cares about (and reconciles with the
intentional `//nolint:exhaustruct`, `//nolint:mnd,gosec` directives already in
`config.go:50` and `retry.go:121`), or (b) document explicitly that the project
runs golangci-lint with defaults and that the in-source `//nolint:` markers are
deliberate.

## P2 — Usability & developer experience

### T3. Add godoc-displayable `Example` functions

There is no `ExampleDo` (or similar) in `retry_test.go` — `grep -nE 'func
(Example|Benchmark)' retry_test.go` returns nothing. For a public library,
`Example*` functions render in `pkg.go.dev` and give users a copy-pasteable
entry point. Add at least one `ExampleDo` (success path) and one showing a
custom `IsRetryable`.

### T4. Add a benchmark for the backoff path

`Backoff` and `ComputeDelay` (`retry.go:104`, `retry.go:114`) are hot paths for
callers, yet there is no `Benchmark*` in `retry_test.go`. Add
`BenchmarkComputeDelay` so jitter allocation cost is visible. This also gives a
concrete number to cite if/when the jitter implementation is revisited (see
`ROADMAP.md` → "Deterministic RNG option").

### T5. Document the coverage workflow

`reports/coverage.out` exists (100% statement coverage today) but the
`reports/` directory is gitignored (`.gitignore:43`). Nothing tells a
contributor how to regenerate it. Add a one-liner to `CONTRIBUTING.md`:
`go test ./... -race -coverprofile=reports/coverage.out && go tool cover
-func=reports/coverage.out`.

## P3 — Release hygiene

### T6. Decide on CI

There is no `.github/workflows/` directory and no remote, so nothing runs tests
on push. Once a remote exists (see T1), add a minimal workflow running
`go test ./... -race` and `golangci-lint run ./...` on push/PR. Not actionable
until a remote is published, hence P3 rather than P2.

### T7. Add version-compare links to `CHANGELOG.md` once a remote exists

`CHANGELOG.md` currently has no `[Unreleased]`/version compare links (correct:
there is no remote URL pattern to build them from — see `CHANGELOG.md` note).
When a remote is published, add the standard `keepachangelog` footer links.
