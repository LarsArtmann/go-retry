# Draft: BuildFlow issue — go-mod-update pin-blind go-line bump

Filed 2026-10-27 against LarsArtmann/BuildFlow after source-level verification
(`tools/gomod/update_1.go` at `49ced3e` = installed binary) and a live repro
(scratch module, replace-only dep, `go version` shimmed to 1.27.0).

---

## Problem

`go-mod-update` bumps the `go` directive to the installed toolchain's minor whenever the module declares an older one — unconditionally, with no notion of a deliberate pin. `tools/gomod/update_1.go` (`determineGoLineTargets`), verified at `49ced3e` (the installed binary):

```go
case goModVersion.Major < installedVersion.Major ||
    (goModVersion.Major == installedVersion.Major && goModVersion.Minor < installedVersion.Minor):
    r.Logger.Infof("Updating go line %s -> %s", goModVersion, installedVersion.MinorString())
    return goVersionTargets{GoLine: installedVersion.MinorString()}, nil
```

Live repro (binary `49ced3e`): scratch module pinned `go 1.26`, sole dependency behind a local `replace` (nothing updatable), `go version` shimmed to 1.27.0:

```
INFO  Updating go line 1.26 -> 1.27
DEBUG Running: go mod edit -go=1.27
```

go.mod ends at `go 1.27` with zero dependency changes — the bump is the only writer.

go-retry pins `go 1.26` (guarded by `TestModuleGoDirectiveStaysPinned`, stated in every doc); this policy re-lost the pin three times in 20 commits, so the repo carried `skip_steps: [go-mod-update]` in a committed `.buildflow.yml`. Removed today (owner decision: the repo should need no BuildFlow config) — which re-exposes the pin to every `buildflow update`/`--fix` run. Not a duplicate of #58: that one was plain runs auto-enabling `--fix`; here the repair itself is pin-blind even when repairs are wanted.

## Why the bump is never required

The go directive is a floor, not a target: any newer toolchain builds a `go 1.26` module unchanged. The only case where the directive must rise — a dependency's own floor — is already handled natively (`go get`/`go mod tidy` raise it, hard error otherwise), and go-mod-normalize already gates dep-floor cases (`HeldBackByDepFloor`).

## Ask

- Default: raise the go line only when a dependency floor forces it (or the declared floor fails to build). Directive bumps become opt-in, like Dependabot's explicit `go` bumps.
- If the auto-bump stays: a respected pin marker (e.g. a go.mod comment) so protection does not require a repo config file.

💘 Generated with Crush
