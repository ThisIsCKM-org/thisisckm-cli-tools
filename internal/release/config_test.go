package release

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigUsesDefaultsWhenFileMissing(t *testing.T) {
	root := t.TempDir()
	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got, want := cfg.Branches.Develop, BranchName("develop"); got != want {
		t.Fatalf("develop branch = %q, want %q", got, want)
	}
	if got, want := cfg.Branches.Main, BranchName("main"); got != want {
		t.Fatalf("main branch = %q, want %q", got, want)
	}
}

func TestEnsureConfigWritesDefaultConfig(t *testing.T) {
	root := t.TempDir()
	if err := EnsureConfig(root); err != nil {
		t.Fatalf("ensure config: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "release.config.json")); err != nil {
		t.Fatalf("release.config.json missing: %v", err)
	}
}

func TestConfigurePromptsAndWritesConfig(t *testing.T) {
	root := t.TempDir()
	in := strings.NewReader("dev\nmaster\n")
	out := &bytes.Buffer{}
	if err := Configure(root, in, out); err != nil {
		t.Fatalf("configure: %v", err)
	}
	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got, want := cfg.Branches.Develop, BranchName("dev"); got != want {
		t.Fatalf("develop branch = %q, want %q", got, want)
	}
	if got, want := cfg.Branches.Main, BranchName("master"); got != want {
		t.Fatalf("main branch = %q, want %q", got, want)
	}
	if !strings.Contains(out.String(), "What is the development branch?") || !strings.Contains(out.String(), "What is the main branch?") {
		t.Fatalf("prompt output missing questions: %q", out.String())
	}
}

func TestConfigureUsesExistingConfigAsPromptDefaults(t *testing.T) {
	root := t.TempDir()
	seed := []byte("{\"branches\":{\"develop\":\"feature-dev\",\"main\":\"release-main\"}}\n")
	if err := os.WriteFile(filepath.Join(root, "release.config.json"), seed, 0o644); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	in := strings.NewReader("\n\n")
	out := &bytes.Buffer{}
	if err := Configure(root, in, out); err != nil {
		t.Fatalf("configure: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "release.config.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "feature-dev") || !strings.Contains(got, "release-main") {
		t.Fatalf("config content = %s", got)
	}
	if !strings.Contains(out.String(), "What is the development branch? [feature-dev]") || !strings.Contains(out.String(), "What is the main branch? [release-main]") {
		t.Fatalf("prompt output missing existing defaults: %q", out.String())
	}
}

func TestInitializeCreatesReleaseFileWithoutConfigFile(t *testing.T) {
	root := t.TempDir()
	if err := Initialize(root, "0.1.0"); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "release.json")); err != nil {
		t.Fatalf("release.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "release.config.json")); !os.IsNotExist(err) {
		t.Fatalf("release.config.json should not be created by init, stat err: %v", err)
	}
}
