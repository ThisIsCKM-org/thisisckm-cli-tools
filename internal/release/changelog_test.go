package release

import (
	"os"
	"strings"
	"testing"
)

func TestPromoteChangelogReplacesUnreleasedWithVersionedEntry(t *testing.T) {
	root := t.TempDir()
	content := `# Changelog

All notable changes to this project will be documented in this file.

## Unreleased
### Added
- Draft work that should be replaced.

## [0.0.1] - 2026-01-01
### Added
- Previous release.
`
	if err := os.WriteFile(ChangelogFile(root), []byte(content), 0o644); err != nil {
		t.Fatalf("write changelog: %v", err)
	}
	items := []string{
		"### Added\n- Release CLI alpha flow.",
		"### Fixed\n- PR title includes prerelease phase.",
	}
	if err := PromoteChangelog(root, "0.1.0", items); err != nil {
		t.Fatalf("promote changelog: %v", err)
	}
	data, err := os.ReadFile(ChangelogFile(root))
	if err != nil {
		t.Fatalf("read changelog: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "## Unreleased\n### Added\n### Fixed\n### Changed\n### Removed\n### Breaking") {
		t.Fatalf("unreleased scaffold was not reset: %q", got)
	}
	if !strings.Contains(got, "## [0.1.0] - ") {
		t.Fatalf("versioned release heading missing: %q", got)
	}
	if !strings.Contains(got, "Release CLI alpha flow") || !strings.Contains(got, "PR title includes prerelease phase") {
		t.Fatalf("staged changelog entries missing: %q", got)
	}
	if strings.Contains(got, "Draft work that should be replaced") {
		t.Fatalf("old unreleased content was not replaced: %q", got)
	}
	if !strings.Contains(got, "## [0.0.1] - 2026-01-01") {
		t.Fatalf("previous release history missing: %q", got)
	}
}
