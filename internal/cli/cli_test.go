package cli

import (
	"os"
	"path/filepath"
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
