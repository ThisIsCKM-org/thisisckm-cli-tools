package release

import (
	"fmt"
	"os"
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

func EnsureVersionFile(root string, version string) error {
	path := StateFile(root)
	if _, err := os.Stat(path); err == nil {
		current, loadErr := Load(path)
		if loadErr != nil {
			return loadErr
		}
		if current.BaseVersion != version {
			return fmt.Errorf("version.json already exists with base version %s", current.BaseVersion)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return Save(path, Seed(version))
}

func Initialize(root, version string) error {
	if err := EnsureVersionFile(root, version); err != nil {
		return err
	}
	if err := EnsureChangelog(root); err != nil {
		return err
	}
	return nil
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
	insertAt := strings.Index(content, "## Unreleased")
	if insertAt == -1 {
		return fmt.Errorf("missing Unreleased section in changelog")
	}
	heading := fmt.Sprintf("## [%s] - %s\n", version, time.Now().Format("2006-01-02"))
	var builder strings.Builder
	builder.WriteString(content[:insertAt])
	builder.WriteString(heading)
	if len(items) > 0 {
		for _, item := range items {
			builder.WriteString(item)
			if !strings.HasSuffix(item, "\n") {
				builder.WriteString("\n")
			}
		}
		builder.WriteString("\n")
	} else {
		builder.WriteString("\n")
	}
	builder.WriteString(content[insertAt:])
	return os.WriteFile(path, []byte(builder.String()), 0o644)
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
