package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunChangelogCreatesStagedEntry(t *testing.T) {
	root := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	if err := Run([]string{"changelog", "feature", "-m", "Add branch aliases"}); err != nil {
		t.Fatalf("run changelog: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(root, "changelogs"))
	if err != nil {
		t.Fatalf("read changelogs: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected a changelog entry to be created")
	}
	data, err := os.ReadFile(filepath.Join(root, "changelogs", entries[0].Name()))
	if err != nil {
		t.Fatalf("read entry: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "### Added") || !strings.Contains(got, "Add branch aliases") {
		t.Fatalf("entry content = %q", got)
	}
}

func TestRunChangelogAcceptsAliases(t *testing.T) {
	cases := []struct {
		name string
		kind string
		want string
	}{
		{name: "added", kind: "added", want: "### Added"},
		{name: "change", kind: "change", want: "### Changed"},
		{name: "removed", kind: "removed", want: "### Removed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			oldwd, err := os.Getwd()
			if err != nil {
				t.Fatalf("get wd: %v", err)
			}
			if err := os.Chdir(root); err != nil {
				t.Fatalf("chdir: %v", err)
			}
			t.Cleanup(func() { _ = os.Chdir(oldwd) })

			if err := Run([]string{"changelog", tc.kind, "-m", "Alias support"}); err != nil {
				t.Fatalf("run changelog: %v", err)
			}
			entries, err := os.ReadDir(filepath.Join(root, "changelogs"))
			if err != nil {
				t.Fatalf("read changelogs: %v", err)
			}
			if len(entries) == 0 {
				t.Fatal("expected a changelog entry to be created")
			}
			data, err := os.ReadFile(filepath.Join(root, "changelogs", entries[0].Name()))
			if err != nil {
				t.Fatalf("read entry: %v", err)
			}
			got := string(data)
			if !strings.Contains(got, tc.want) || !strings.Contains(got, "Alias support") {
				t.Fatalf("entry content = %q", got)
			}
		})
	}
}

func TestRunReleaseConfigCreatesConfig(t *testing.T) {
	root := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdin: %v", err)
	}
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	os.Stdin = stdinR
	os.Stdout = stdoutW
	t.Cleanup(func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	})

	go func() {
		_, _ = stdinW.Write([]byte("dev\nmain\n"))
		_ = stdinW.Close()
	}()

	if err := runRelease([]string{"config"}); err != nil {
		t.Fatalf("run release config: %v", err)
	}
	_ = stdoutW.Close()
	_ = stdoutR.Close()

	data, err := os.ReadFile(filepath.Join(root, "release.config.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "dev") || !strings.Contains(got, "main") {
		t.Fatalf("config content = %s", got)
	}
}

func TestRunReleaseStageCommandsRejectExtraArgs(t *testing.T) {
	root := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	for _, stage := range []string{"alpha", "beta", "rc", "final"} {
		t.Run(stage, func(t *testing.T) {
			err := runRelease([]string{stage, "0.2.0"})
			if err == nil {
				t.Fatal("expected extra args to fail")
			}
			want := "usage: thisisckm release " + stage
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error = %q, want %q", err, want)
			}
		})
	}
}

func TestRunAddSkillsInstallsCodexSkills(t *testing.T) {
	root := fixtureRepo(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	if err := Run([]string{"add-skills", "codex"}); err != nil {
		t.Fatalf("run add-skills: %v", err)
	}
	for _, skill := range []string{"prepare-pr", "prepare-release-pr"} {
		path := filepath.Join(home, ".codex", "skills", skill, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing installed skill %s: %v", skill, err)
		}
	}
}

func TestRunAddSkillsRequiresExplicitTarget(t *testing.T) {
	if err := Run([]string{"add-skills"}); err == nil {
		t.Fatal("expected add-skills without target to fail")
	}
}

func TestRunAddSkillsFromBuiltBinaryUsesEmbeddedSkills(t *testing.T) {
	repoRoot := mustRepoRoot(t)
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "thisisckm")
	build := exec.Command("go", "build", "-o", binPath, "./cmd/thisisckm")
	build.Dir = repoRoot
	build.Env = os.Environ()
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build binary: %v\n%s", err, string(out))
	}

	home := t.TempDir()
	runDir := t.TempDir()
	run := exec.Command(binPath, "add-skills", "codex")
	run.Dir = runDir
	run.Env = append(os.Environ(), "HOME="+home)
	if out, err := run.CombinedOutput(); err != nil {
		t.Fatalf("run add-skills from temp dir: %v\n%s", err, string(out))
	}

	for _, skill := range []string{"prepare-pr", "prepare-release-pr"} {
		path := filepath.Join(home, ".codex", "skills", skill, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing installed skill %s: %v", skill, err)
		}
	}
}

func mustRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repo root not found: %v", err)
	}
	return root
}

func fixtureRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	for _, skill := range []string{"prepare-pr", "prepare-release-pr"} {
		dir := filepath.Join(root, "skills", skill)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir skill dir: %v", err)
		}
		content := []byte("---\nname: " + skill + "\n---\n\nthisisckm changelog\nthisisckm release\n")
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), content, 0o644); err != nil {
			t.Fatalf("write skill file: %v", err)
		}
	}
	return root
}
