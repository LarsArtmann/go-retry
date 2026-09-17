//go:build tools

// Package tools pins this repository's development tools so that local
// sessions and CI resolve identical versions from go.mod instead of floating
// mechanisms ("go run pkg@version", "go install ...@latest", nixpkgs refs).
//
// The blank imports below are compiled out of the library: this file carries
// the `tools` build tag, so consumers of github.com/larsartmann/go-retry never
// download these dependencies.
//
// Usage (version comes from go.mod, never suffix an @version):
//
//	go run github.com/rhysd/actionlint/cmd/actionlint -verbose
//	go run golang.org/x/vuln/cmd/govulncheck ./...
//
// Bump a tool like any dependency (Dependabot covers the gomod group):
//
//	go get github.com/rhysd/actionlint/cmd/actionlint@v1.7.13
//
// dprint (markdown/JSON/YAML/Dockerfile formatter) is a Rust binary, not a Go
// tool, so it cannot be pinned here. Its version is pinned where it runs in CI
// (.github/workflows/ci.yml, dprint step) and its plugin set is pinned by URL
// in dprint.json; locally run it via `nix run nixpkgs#dprint -- check`.
package tools

import (
	_ "github.com/rhysd/actionlint/cmd/actionlint" // GitHub workflow schema gate
	_ "golang.org/x/vuln/cmd/govulncheck"          // Go vulnerability scanner
)
