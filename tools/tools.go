// Package tools documents this repository's pinned development tools
// (actionlint, govulncheck). The pins themselves live in this module's
// go.mod as Go `tool` directives — the successor to blank-import tools.go
// files, which the Go 1.26 toolchain rejects ("is a program, not an
// importable package").
//
// This is its own nested module (not the library module) on purpose: current
// golang.org/x/* tool versions declare `go 1.26.0`, which would force the
// library's go.mod off its recorded relaxed `go 1.26` directive (guarded by
// TestModuleGoDirectiveStaysPinned). Consumers of
// github.com/larsartmann/go-retry download none of this — the library
// module's dependency surface is unchanged.
//
// Install all pinned tools into ~/go/bin (rerun after a bump):
//
//	go -C tools install github.com/rhysd/actionlint/cmd/actionlint golang.org/x/vuln/cmd/govulncheck
//
// Or run one ad hoc (note: executes with working directory tools/):
//
//	go -C tools tool actionlint -version
//
// Bump a tool like any dependency (Dependabot's gomod watcher covers /tools):
//
//	go -C tools get -tool golang.org/x/vuln/cmd/govulncheck@latest
//
// dprint (markdown/JSON/YAML/Dockerfile formatter) is a Rust binary, not a Go
// tool, so it cannot be pinned here. Its version is pinned where it runs in
// CI (.github/workflows/ci.yml, dprint step) and its plugin set is pinned by
// URL in dprint.json; locally run it via `nix run nixpkgs#dprint -- check`.
package tools
