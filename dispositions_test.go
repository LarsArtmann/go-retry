package retry_test

import (
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The disposition guards below exist because the 2026-10-08 gate-restoration
// session proved that deliberate decisions silently reverted twice in one day
// when no gate read them: BuildFlow's go-mod-update step and go-structure-
// linter's go-version repair rule each re-pinned the go directive until config
// dispositions stopped them. Like TestModuleGoDirectiveStaysPinned, these turn
// those dispositions into enforced invariants: removing any of them must fail
// a test that names the consequence, not a code review.

// TestBuildflowDispositionsStayCommitted guards the `.buildflow.yml`
// dispositions behind the go-directive pin: `go-mod-update` stays skipped (its
// unconditional minor bump silently re-lost the pin three times in 20 commits;
// Dependabot owns gomod bumps here) and `respect_patch_floor` stays on (the
// tools module legitimately declares the patch-form `go 1.26.0` six x/*
// dependencies demand). Deleting either flips go.mod on the next full run.
func TestBuildflowDispositionsStayCommitted(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(".buildflow.yml")
	if err != nil {
		t.Fatalf("read .buildflow.yml: %v", err)
	}

	skipSteps, patchFloor := parseBuildflowDispositions(string(data))

	if !contains(skipSteps, "go-mod-update") {
		t.Fatalf(
			".buildflow.yml no longer skips go-mod-update (skip_steps = %v): the step unconditionally bumps the go directive to the installed minor; remove this skip only after re-deciding the pin policy and updating every doc that states go 1.26",
			skipSteps,
		)
	}

	if patchFloor != "true" {
		t.Fatalf(
			".buildflow.yml sets respect_patch_floor = %q, want \"true\": tools/go.mod declares the patch-form go 1.26.0 floor six golang.org/x/* dependencies demand, and without this option the go-version-auto-configure step rewrites it",
			patchFloor,
		)
	}
}

// parseBuildflowDispositions extracts the `skip_steps` list entries and the
// `tool_options.go-version-auto-configure.respect_patch_floor` value from the
// committed config. It understands exactly the shape this file uses (block
// list, two-level nesting, comments); any other shape fails the assertions
// above closed, which is deliberate: an unreadable config must send a human
// to look, not pass silently.
func parseBuildflowDispositions(data string) (skipSteps []string, patchFloor string) {
	const (
		topLevel    = 0
		nestedLevel = 1
		optionLevel = 2
	)

	section, nested := "", ""

	for line := range strings.SplitSeq(data, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := optionLevel
		switch {
		case !strings.HasPrefix(line, " "):
			indent = topLevel
		case !strings.HasPrefix(line, "    "):
			indent = nestedLevel
		}

		key, value, found := strings.Cut(trimmed, ":")
		value = strings.TrimSpace(value)

		switch indent {
		case topLevel:
			section, nested = key, ""
		case nestedLevel:
			if item, isItem := strings.CutPrefix(trimmed, "- "); isItem && section == "skip_steps" {
				skipSteps = append(skipSteps, strings.TrimSpace(item))

				continue
			}
			nested = key
		case optionLevel:
			if section == "tool_options" && nested == "go-version-auto-configure" && found && key == "respect_patch_floor" {
				patchFloor = value
			}
		}
	}

	return skipSteps, patchFloor
}

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

	output, err := exec.Command("go", "list", "-m", "-f", "{{.GoVersion}}", "github.com/larsartmann/go-error-family").Output()
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

func contains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}

	return false
}
