package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallCodexCopiesAllSkills(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	actions, err := Install(AgentCodex, Options{})
	if err != nil {
		t.Fatalf("install codex: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("actions len = %d, want 2", len(actions))
	}
	for _, name := range []string{"prepare-pr", "prepare-release-pr"} {
		path := filepath.Join(home, ".codex", "skills", name, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing codex skill %s: %v", name, err)
		}
	}
}

func TestInstallCursorRendersRules(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	actions, err := Install(AgentCursor, Options{})
	if err != nil {
		t.Fatalf("install cursor: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("actions len = %d, want 2", len(actions))
	}
	data, err := os.ReadFile(filepath.Join(home, ".cursor", "rules", "thisisckm-prepare-pr.mdc"))
	if err != nil {
		t.Fatalf("read cursor rule: %v", err)
	}
	if !strings.Contains(string(data), "prepare-pr") || !strings.Contains(string(data), "thisisckm changelog") {
		t.Fatalf("cursor rule content = %s", string(data))
	}
}

func TestInstallSkipsExistingFilesWithoutForce(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	existing := filepath.Join(home, ".codex", "skills", "prepare-pr")
	if err := os.MkdirAll(existing, 0o755); err != nil {
		t.Fatalf("seed existing: %v", err)
	}
	seed := filepath.Join(existing, "SKILL.md")
	if err := os.WriteFile(seed, []byte("local changes"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	actions, err := Install(AgentCodex, Options{})
	if err != nil {
		t.Fatalf("install codex: %v", err)
	}
	if !actions[0].Skipped {
		t.Fatal("expected existing skill to be skipped")
	}
	data, err := os.ReadFile(seed)
	if err != nil {
		t.Fatalf("read seed: %v", err)
	}
	if string(data) != "local changes" {
		t.Fatalf("seed file was overwritten: %s", string(data))
	}
}

func TestInstallForceOverwritesExistingFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	target := filepath.Join(home, ".codex", "skills", "prepare-pr")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte("stale"), 0o644); err != nil {
		t.Fatalf("seed stale: %v", err)
	}
	if _, err := Install(AgentCodex, Options{Force: true}); err != nil {
		t.Fatalf("force install: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(target, "SKILL.md"))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	if strings.Contains(string(data), "stale") {
		t.Fatal("expected stale content to be replaced")
	}
}

func TestParseAgent(t *testing.T) {
	if _, err := ParseAgent("codex"); err != nil {
		t.Fatalf("parse codex: %v", err)
	}
	if _, err := ParseAgent("unknown"); err == nil {
		t.Fatal("expected parse failure for unknown agent")
	}
}
