//go:build tools

// Package tools pins this repository's development tools (actionlint,
// govulncheck) so local sessions and CI resolve identical versions from
// go.mod instead of floating mechanisms ("go run pkg@version",
// "go install ...@latest", unversioned nixpkgs refs).
//
// This file lives in its own nested module (not the library module) on
// purpose: current golang.org/x/* tool versions declare `go 1.26.0`, which
// would force the library's go.mod off its recorded relaxed `go 1.26`
// directive (guarded by TestModuleGoDirectiveStaysPinned). Consumers of
// github.com/larsartmann/go-retry download none of this — the library
// module's dependency surface is unchanged.
//
// Install all pinned tools into ~/go/bin (once per bump):
//
//	go -C tools install github.com/rhysd/actionlint/cmd/actionlint golang.org/x/vuln/cmd/govulncheck
//
// Bump a tool like any dependency (Dependabot's gomod watcher covers /tools):
//
//	go -C tools get github.com/rhysd/actionlint/cmd/actionlint@v1.7.13
//
// dprint (markdown/JSON/YAML/Dockerfile formatter) is a Rust binary, not a Go
// tool, so it cannot be pinned here. Its version is pinned where it runs in
// CI (.github/workflows/ci.yml, dprint step) and its plugin set is pinned by
// URL in dprint.json; locally run it via `nix run nixpkgs#dprint -- check`.
package tools

import (
	_ "github.com/rhysd/actionlint/cmd/actionlint" // GitHub workflow schema gate
	_ "golang.org/x/vuln/cmd/govulncheck"          // Go vulnerability scanner
)
