package release

import (
	"os"
	"strings"
	"testing"
)

func TestReleasePullRequestTitleUsesChannel(t *testing.T) {
	cases := map[string]string{
		"0.1.0-alpha.1": "Alpha Release 0.1.0-alpha.1",
		"0.1.0-beta.1":  "Beta Release 0.1.0-beta.1",
		"0.1.0-rc.1":    "Release Candidate 0.1.0-rc.1",
		"0.1.0":         "Stable Release 0.1.0",
	}
	for version, want := range cases {
		if got := releasePullRequestTitle(version); got != want {
			t.Fatalf("title for %s = %q, want %q", version, got, want)
		}
	}
}

func TestUnreleasedChangelogExtractsCurrentSection(t *testing.T) {
	root := t.TempDir()
	content := `# Changelog

## Unreleased
### Added
- Release CLI PR body.

### Fixed
- Alpha release title.

## [0.0.1] - 2026-01-01
### Added
- Previous release.
`
	if err := os.WriteFile(ChangelogFile(root), []byte(content), 0o644); err != nil {
		t.Fatalf("write changelog: %v", err)
	}
	got, err := UnreleasedChangelog(root)
	if err != nil {
		t.Fatalf("extract unreleased: %v", err)
	}
	if !strings.Contains(got, "## Unreleased") || !strings.Contains(got, "Release CLI PR body") {
		t.Fatalf("unreleased section missing expected content: %q", got)
	}
	if strings.Contains(got, "Previous release") {
		t.Fatalf("unreleased section included old release: %q", got)
	}
}

func TestReleasePullRequestBodyIncludesUnreleasedChangelog(t *testing.T) {
	root := t.TempDir()
	content := `# Changelog

## Unreleased
### Added
- Documented alpha release notes.
`
	if err := os.WriteFile(ChangelogFile(root), []byte(content), 0o644); err != nil {
		t.Fatalf("write changelog: %v", err)
	}
	body := releasePullRequestBody(root, "0.1.0-alpha.1")
	if !strings.Contains(body, "## Release") || !strings.Contains(body, "0.1.0-alpha.1") {
		t.Fatalf("body missing release metadata: %q", body)
	}
	if !strings.Contains(body, "## Changelog") || !strings.Contains(body, "Documented alpha release notes") {
		t.Fatalf("body missing unreleased changelog: %q", body)
	}
}
