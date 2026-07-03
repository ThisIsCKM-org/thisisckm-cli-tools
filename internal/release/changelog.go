package release

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const changelogScaffold = `# Changelog

All notable changes to this project will be documented in this file.

## Unreleased
### Added
### Fixed
### Changed
### Removed
### Breaking
`

func EnsureChangelog(root string) error {
	path := ChangelogFile(root)
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(changelogScaffold), 0o644)
}

func EnsureReleaseFile(root string, version string) error {
	path := StateFile(root)
	if current, existingPath, err := loadStateFile(root); err == nil {
		if current.BaseVersion != version {
			return fmt.Errorf("release.json already exists with base version %s", current.BaseVersion)
		}
		if existingPath != path {
			return Save(path, current)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return Save(path, Seed(version))
}

func Initialize(root, version string) error {
	if err := EnsureReleaseFile(root, version); err != nil {
		return err
	}
	if err := EnsureChangelog(root); err != nil {
		return err
	}
	if err := EnsureChangelogInbox(root); err != nil {
		return err
	}
	return nil
}

func EnsureChangelogInbox(root string) error {
	entriesDir := filepath.Join(root, "changelogs")
	if err := os.MkdirAll(entriesDir, 0o755); err != nil {
		return err
	}
	readmePath := filepath.Join(entriesDir, "README.md")
	if _, err := os.Stat(readmePath); err == nil {
		return nil
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	const inboxReadme = `# Changelog Entry Inbox

This folder stores one file per in-progress change log entry.

Release flow:
- add new entries here while work is in progress
- consolidate the staged entries into CHANGELOG.md when the version is released
- keep the root changelog as the published history

Entry files should be small, dated, and focused on a single change.
`
	return os.WriteFile(readmePath, []byte(inboxReadme), 0o644)
}

func CollectStagedEntries(root string) ([]string, error) {
	entriesDir := filepath.Join(root, "changelogs")
	matches, err := os.ReadDir(entriesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, entry := range matches {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "README.md" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	var items []string
	for _, name := range names {
		fullPath := filepath.Join(entriesDir, name)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, err
		}
		content := strings.TrimSpace(string(data))
		if content == "" {
			continue
		}
		items = append(items, content)
	}
	return items, nil
}

func ClearStagedEntries(root string) error {
	entriesDir := filepath.Join(root, "changelogs")
	matches, err := os.ReadDir(entriesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range matches {
		if entry.IsDir() || entry.Name() == "README.md" {
			continue
		}
		if err := os.Remove(filepath.Join(entriesDir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func CreateStagedChangelogEntry(root, kind, message string) (string, error) {
	if err := EnsureChangelog(root); err != nil {
		return "", err
	}
	if err := EnsureChangelogInbox(root); err != nil {
		return "", err
	}
	section, err := changelogSectionForKind(kind)
	if err != nil {
		return "", err
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return "", fmt.Errorf("changelog message is required")
	}
	date := time.Now().Format("2006-01-02")
	slug := changelogSlug(message)
	baseName := fmt.Sprintf("%s-%s-%s.md", date, kind, slug)
	path := filepath.Join(root, "changelogs", baseName)
	for suffix := 1; ; suffix++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		}
		baseName = fmt.Sprintf("%s-%s-%s-%d.md", date, kind, slug, suffix)
		path = filepath.Join(root, "changelogs", baseName)
	}
	content := fmt.Sprintf("### %s\n- %s\n", section, message)
	return path, os.WriteFile(path, []byte(content), 0o644)
}

func changelogSectionForKind(kind string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "bug", "fix", "fixed":
		return "Fixed", nil
	case "feature", "feat", "add", "added":
		return "Added", nil
	case "change", "changed", "update", "updated":
		return "Changed", nil
	case "removed", "remove", "delete", "deleted":
		return "Removed", nil
	default:
		return "", fmt.Errorf("unknown changelog kind %q", kind)
	}
}

func changelogSlug(message string) string {
	message = strings.ToLower(message)
	var builder strings.Builder
	lastDash := false
	for _, r := range message {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_' || r == '/' || r == '.':
			if !lastDash && builder.Len() > 0 {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}
	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		return "entry"
	}
	return slug
}

func PromoteChangelog(root, version string, items []string) error {
	path := ChangelogFile(root)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	if strings.Contains(content, fmt.Sprintf("## [%s] - ", version)) {
		return nil
	}
	start := strings.Index(content, "## Unreleased")
	if start == -1 {
		return fmt.Errorf("missing Unreleased section in changelog")
	}
	section := content[start:]
	nextRelease := strings.Index(section[len("## Unreleased"):], "\n## ")
	end := len(content)
	if nextRelease != -1 {
		end = start + len("## Unreleased") + nextRelease
	}
	previousReleases := strings.TrimLeft(content[end:], "\n")

	var builder strings.Builder
	builder.WriteString(strings.TrimRight(content[:start], "\n"))
	builder.WriteString("\n\n")
	builder.WriteString(emptyUnreleasedSection())
	builder.WriteString("\n\n")
	builder.WriteString(fmt.Sprintf("## [%s] - %s\n", version, time.Now().Format("2006-01-02")))
	if len(items) > 0 {
		for _, item := range items {
			builder.WriteString(strings.TrimSpace(item))
			builder.WriteString("\n\n")
		}
	} else {
		builder.WriteString("\n")
	}
	if strings.TrimSpace(previousReleases) != "" {
		builder.WriteString(strings.TrimRight(previousReleases, "\n"))
		builder.WriteString("\n")
	}
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}

func emptyUnreleasedSection() string {
	return strings.TrimSpace(`## Unreleased
### Added
### Fixed
### Changed
### Removed
### Breaking`)
}

func UnreleasedChangelog(root string) (string, error) {
	path := ChangelogFile(root)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	start := strings.Index(content, "## Unreleased")
	if start == -1 {
		return "", fmt.Errorf("missing Unreleased section in changelog")
	}
	section := content[start:]
	searchStart := len("## Unreleased")
	if next := strings.Index(section[searchStart:], "\n## "); next != -1 {
		section = section[:searchStart+next]
	}
	return strings.TrimSpace(section), nil
}

func VersionedChangelog(root, version string) (string, error) {
	path := ChangelogFile(root)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	headingPrefix := fmt.Sprintf("## [%s] - ", version)
	start := strings.Index(content, headingPrefix)
	if start == -1 {
		return "", fmt.Errorf("missing changelog section for %s", version)
	}
	section := content[start:]
	searchStart := len(headingPrefix)
	if next := strings.Index(section[searchStart:], "\n## "); next != -1 {
		section = section[:searchStart+next]
	}
	return strings.TrimSpace(section), nil
}

func AppendReleaseNote(root, line string) error {
	path := ChangelogFile(root)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	if !strings.Contains(content, "## Unreleased") {
		return fmt.Errorf("missing Unreleased section in changelog")
	}
	needle := "### Breaking\n"
	idx := strings.Index(content, needle)
	if idx == -1 {
		return fmt.Errorf("missing changelog scaffold")
	}
	insertAt := idx + len(needle)
	var builder strings.Builder
	builder.WriteString(content[:insertAt])
	builder.WriteString(line)
	if !strings.HasSuffix(line, "\n") {
		builder.WriteString("\n")
	}
	builder.WriteString(content[insertAt:])
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}
