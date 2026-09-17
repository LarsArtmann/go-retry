package retry_test

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// actionInputAllowlist names every input key each pinned remote action
// accepts, so a typo'd `with:` key (the 2026-09-13 `namee:` class) fails here
// even though actionlint only validates workflow schema, not remote-action
// inputs. Verified from each action's action.yml at the exact pinned commit
// SHA on 2026-09-17 via:
//
//	gh api "repos/<owner>/<repo>/contents/action.yml?ref=<sha>" --jq .content | base64 -d
//
// When pinning a new action or re-pinning an existing one, re-verify the
// inputs at the new SHA and update this table in the same change.
var actionInputAllowlist = map[string][]string{
	"actions/checkout": {
		"repository", "ref", "token", "ssh-key", "ssh-known-hosts", "ssh-strict",
		"ssh-user", "persist-credentials", "path", "clean", "filter",
		"sparse-checkout", "sparse-checkout-cone-mode", "fetch-depth", "fetch-tags",
		"show-progress", "lfs", "submodules", "set-safe-directory",
		"github-server-url", "allow-unsafe-pr-checkout",
	},
	"actions/setup-go": {
		"go-version", "go-version-file", "check-latest", "token", "cache",
		"cache-dependency-path", "architecture", "go-download-base-url",
	},
	"actions/upload-artifact": {
		"name", "path", "if-no-files-found", "retention-days", "compression-level",
		"overwrite", "include-hidden-files", "archive",
	},
	"dprint/check": {
		"dprint-version", "config-path", "args", "working-directory", "cache",
		"annotations", "verify-attestation",
	},
	"golang/govulncheck-action": {
		"go-version-input", "check-latest", "cache", "cache-dependency-path",
		"go-package", "work-dir", "repo-checkout", "go-version-file",
		"output-format", "output-file",
	},
	"golangci/golangci-lint-action": {
		"version", "version-file", "install-mode", "install-only", "working-directory",
		"github-token", "verify", "only-new-issues", "args", "skip-cache",
		"skip-save-cache", "cache-invalidation-interval", "problem-matchers",
		"debug", "experimental",
	},
}

// workflowActionUse is one `uses:` step found in a workflow file, together
// with the `with:` input keys it passes.
type workflowActionUse struct {
	workflow string
	line     int
	action   string
	ref      string
	inputs   []string
}

var (
	workflowUsesRe = regexp.MustCompile(`^(\s*)(?:- )?uses:[ \t]+([^\s#]+)(?:[ \t]+#.*)?$`)
	workflowWithRe = regexp.MustCompile(`^(\s*)(?:- )?with:\s*$`)
	workflowKeyRe  = regexp.MustCompile(`^(\s+)([A-Za-z][A-Za-z0-9_-]*):(?:\s|$)`)
	workflowShaRe  = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

// TestRemoteActionInputsAreAllowlisted walks every workflow file, extracts
// each pinned `uses:` step and its `with:` keys, and fails when a key is not
// a real input of that action. actionlint cannot see remote-action inputs
// (the runner silently ignores unknown keys), so this test is the only gate
// for that failure class.
func TestRemoteActionInputsAreAllowlisted(t *testing.T) {
	t.Parallel()

	uses := collectWorkflowActionUses(t)
	if len(uses) == 0 {
		t.Fatal("no `uses:` steps found under .github/workflows — the parser is broken, not the workflows")
	}

	seen := map[string]bool{}

	for _, step := range uses {
		allowed, known := actionInputAllowlist[step.action]
		if !known {
			t.Fatalf(
				"%s:%d: action %q is pinned but missing from actionInputAllowlist — verify its inputs from action.yml at the pinned SHA and add them",
				step.workflow, step.line, step.action,
			)
		}
		seen[step.action] = true

		if !workflowShaRe.MatchString(step.ref) {
			t.Fatalf(
				"%s:%d: %s is pinned as %q — repo convention is a 40-hex commit SHA",
				step.workflow, step.line, step.action, step.ref,
			)
		}

		for _, key := range step.inputs {
			if !slices.Contains(allowed, key) {
				t.Fatalf(
					"%s:%d: input %q is not a valid input of %s@%s — typo'd key (the 2026-09-13 `namee:` class), or a stale allowlist after a re-pin",
					step.workflow, step.line, key, step.action, step.ref,
				)
			}
		}
	}

	for action := range actionInputAllowlist {
		if !seen[action] {
			t.Fatalf(
				"actionInputAllowlist entry %q matches no `uses:` in any workflow — prune it",
				action,
			)
		}
	}
}

// collectWorkflowActionUses scans .github/workflows/*.yml. The parser below
// is deliberately fail-closed: any YAML shape it cannot attribute to a
// `uses:`/`with:` structure aborts the test instead of silently skipping.
func collectWorkflowActionUses(t *testing.T) []workflowActionUse {
	t.Helper()

	entries, err := os.ReadDir(".github/workflows")
	if err != nil {
		t.Fatalf("read .github/workflows: %v", err)
	}

	var uses []workflowActionUse

	for _, entry := range entries {
		if entry.IsDir() || (strings.HasSuffix(entry.Name(), ".yml") && !strings.HasSuffix(entry.Name(), ".yaml")) {
			path := filepath.Join(".github/workflows", entry.Name())

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}

			uses = append(uses, parseWorkflowActionUses(t, path, string(data))...)
		}
	}

	return uses
}

// parseWorkflowActionUses extracts (action, ref, with-keys) triples from one
// workflow file using indentation tracking. It supports exactly the YAML
// subset these workflows use — block mappings with plain scalar or block
// scalar values — and fatals on flow mappings, anchors, aliases, and merge
// keys, all of which could hide an input key from extraction.
func parseWorkflowActionUses(t *testing.T, path, content string) []workflowActionUse {
	t.Helper()

	var (
		uses     []workflowActionUse
		current  *workflowActionUse
		withNest = -1
		lineNo   int
	)

	for line := range strings.SplitSeq(content, "\n") {
		lineNo++

		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		if withNest >= 0 {
			if indent > withNest {
				parseWithBlockLine(t, path, lineNo, line, current)
				continue
			}
			withNest = -1 // dedent ends the with: block; fall through
		}

		if match := workflowUsesRe.FindStringSubmatch(line); match != nil {
			current = &workflowActionUse{workflow: path, line: lineNo, ref: match[2]}
			current.action, current.ref, _ = strings.Cut(match[2], "@")

			if current.action == "" || current.ref == "" {
				t.Fatalf(
					"%s:%d: `uses: %s` is not owner/repo@ref — local or docker actions need parser support first",
					path, lineNo, match[2],
				)
			}
			if strings.HasPrefix(current.action, "./") || strings.HasPrefix(current.action, "docker://") {
				t.Fatalf("%s:%d: unsupported `uses:` form %q — extend the parser", path, lineNo, match[2])
			}

			uses = append(uses, *current)

			continue
		}

		if strings.HasPrefix(trimmed, "uses:") {
			t.Fatalf(
				"%s:%d: unparseable `uses:` line %q — flow style or trailing content; extend the parser",
				path, lineNo, trimmed,
			)
		}

		if workflowWithRe.MatchString(line) {
			if current == nil {
				t.Fatalf("%s:%d: `with:` without a preceding `uses:` in the same step", path, lineNo)
			}
			withNest = indent
			current = &uses[len(uses)-1]

			continue
		}

		if strings.HasPrefix(trimmed, "with:") {
			t.Fatalf(
				"%s:%d: `with:` carries inline content %q (flow mapping?) — rewrite as a block mapping or extend the parser",
				path, lineNo, trimmed,
			)
		}
	}

	return uses
}

// parseWithBlockLine records one `key:` line inside a with: block on the
// in-progress step use. Deeper lines that are not keys are value
// continuations (block scalars, nested lists) and are skipped.
func parseWithBlockLine(t *testing.T, path string, lineNo int, line string, current *workflowActionUse) {
	t.Helper()

	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "&") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "<<:") {
		t.Fatalf(
			"%s:%d: anchor/alias/merge key %q inside a with: block could hide an input — rewrite without it",
			path, lineNo, trimmed,
		)
	}

	if match := workflowKeyRe.FindStringSubmatch(line); match != nil {
		if !slices.Contains(current.inputs, match[2]) {
			current.inputs = append(current.inputs, match[2])
		}
	}
}
