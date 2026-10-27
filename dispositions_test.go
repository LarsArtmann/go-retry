package retry_test

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The disposition guards below turn deliberate repo policy into enforced
// invariants: the 2026-10-08 gate-restoration session proved that decisions
// silently revert when no gate reads them. Like
// TestModuleGoDirectiveStaysPinned, removing a guarded disposition must fail
// a test that names the consequence, not a code review.
//
// The former `.buildflow.yml` dispositions (skip go-mod-update,
// respect_patch_floor, art-dupl emit-suppressed-accepted) were removed with
// the file on 2026-10-27 (owner decision: no repo-level BuildFlow config).
// The pin guards remain; the BuildFlow-side protections are pursued
// upstream — see ROADMAP.md and docs/engineering-notes.md.

// TestTerminalSentinelsStayInterfaceTyped guards the erraudit disposition: all
// three terminal sentinels are declared with the `error` interface type, not
// the concrete `*errorfamily.InfrastructureError`. Retyping any of them to the
// concrete type reintroduces the erraudit finding the 2026-10-08 session
// closed and forces every `errors.Is` caller through the concrete type.
func TestTerminalSentinelsStayInterfaceTyped(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("retry.go")
	if err != nil {
		t.Fatalf("read retry.go: %v", err)
	}

	source := string(data)

	for _, sentinel := range []string{"ErrExhausted", "ErrCanceled", "ErrDeadlineExceeded"} {
		pattern := `(?m)^var ` + sentinel + ` error = errorfamily\.NewInfrastructure\(`
		if regexp.MustCompile(pattern).MatchString(source) {
			continue
		}

		t.Errorf(
			"retry.go no longer declares `var %s error = errorfamily.NewInfrastructure(...)`: sentinels must keep the error interface type or the erraudit finding returns; retype deliberately only with a new disposition",
			sentinel,
		)
	}
}

// TestErrorFamilyFloorStaysWithinGoPin guards the dependency side of the
// go 1.26 pin: go-error-family must not declare a go directive above 1.26,
// because a dependency's floor forces this module's directive to rise and
// breaks every doc that states go 1.26. If it fires, re-pin deliberately (or
// bump the pin with the full doc sweep) instead of silencing the test.
func TestErrorFamilyFloorStaysWithinGoPin(t *testing.T) {
	t.Parallel()

	const wantMinor = 26

	const dep = "github.com/larsartmann/go-error-family"

	output, err := exec.CommandContext(context.Background(), "go", "list", "-m", "-f", "{{.GoVersion}}", dep).Output()
	if err != nil {
		t.Fatalf("go list go-error-family GoVersion: %v", err)
	}

	version := strings.TrimSpace(string(output))

	match := regexp.MustCompile(`^(\d+)\.(\d+)`).FindStringSubmatch(version)
	if match == nil {
		t.Fatalf("go-error-family declares unparsable go version %q", version)
	}

	major, err := strconv.Atoi(match[1])
	if err != nil {
		t.Fatalf("parse go-error-family major version %q: %v", match[1], err)
	}

	minor, err := strconv.Atoi(match[2])
	if err != nil {
		t.Fatalf("parse go-error-family minor version %q: %v", match[2], err)
	}

	if major != 1 || minor > wantMinor {
		t.Fatalf(
			"go-error-family declares go %d.%d, want at most go 1.%d: its floor forces this module's directive to rise; re-pin deliberately and update every doc that states go 1.26",
			major,
			minor,
			wantMinor,
		)
	}
}

// TestLycheeExcludesPrivateNamespace guards the `lychee.toml` disposition:
// private LarsArtmann repositories 404 under unauthenticated link checks, and
// the fleet authenticate-vs-exclude policy is undecided, so the namespace
// stays excluded. Removing the pattern turns every link scan red on this
// repo's own docs.
func TestLycheeExcludesPrivateNamespace(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("lychee.toml")
	if err != nil {
		t.Fatalf("read lychee.toml: %v", err)
	}

	config := string(data)

	const pattern = `github\.com/larsartmann`
	if !strings.Contains(config, "(?i)") || !strings.Contains(config, pattern) {
		t.Fatalf(
			"lychee.toml no longer excludes %q case-insensitively: private LarsArtmann repos 404 under unauthenticated lychee checks; remove the exclude only after the fleet authenticate-vs-exclude policy is decided",
			pattern,
		)
	}
}
