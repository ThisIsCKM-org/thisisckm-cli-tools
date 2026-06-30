package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseBranchNameUsesBaseVersion(t *testing.T) {
	if got, want := releaseBranchName("0.1.0"), "release/v0.1.0"; got != want {
		t.Fatalf("releaseBranchName = %q, want %q", got, want)
	}
}

func TestNewCreatesDocumentedReleaseBranch(t *testing.T) {
	root := setupGitReleaseRepo(t)
	workspace := Workspace{Root: root}

	if err := workspace.New("0.1.0"); err != nil {
		t.Fatalf("new release: %v", err)
	}
	if got, want := currentGitBranch(t, root), "release/v0.1.0"; got != want {
		t.Fatalf("branch = %q, want %q", got, want)
	}
	state, err := Load(StateFile(root))
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if got, want := state.Channel, ChannelNone; got != want {
		t.Fatalf("channel = %s, want %s", got, want)
	}
	if got, want := state.State, StateInProgress; got != want {
		t.Fatalf("state = %s, want %s", got, want)
	}
}

func TestAdvancePromotesStagedChangelogEntries(t *testing.T) {
	root := setupGitReleaseRepo(t)
	entryPath := filepath.Join(root, "changelogs", "2026-06-30-add-sub-helper.md")
	entry := `### Added
- Added a sub(float, float) helper.
`
	if err := os.WriteFile(entryPath, []byte(entry), 0o644); err != nil {
		t.Fatalf("write staged changelog entry: %v", err)
	}
	workspace := Workspace{Root: root}

	if err := workspace.Advance(ChannelAlpha); err != nil {
		t.Fatalf("alpha release: %v", err)
	}
	if got, want := currentGitBranch(t, root), "release/v0.1.0-alpha.1"; got != want {
		t.Fatalf("branch = %q, want %q", got, want)
	}
	data, err := os.ReadFile(ChangelogFile(root))
	if err != nil {
		t.Fatalf("read changelog: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "## [0.1.0-alpha.1] - ") {
		t.Fatalf("alpha changelog heading missing: %q", got)
	}
	if !strings.Contains(got, "Added a sub(float, float) helper") {
		t.Fatalf("staged changelog entry missing from changelog: %q", got)
	}
	if _, err := os.Stat(entryPath); !os.IsNotExist(err) {
		t.Fatalf("staged changelog entry was not cleared, stat err: %v", err)
	}
}

func TestAdvanceUsesComputedPrereleaseBranch(t *testing.T) {
	root := setupGitReleaseRepo(t)
	workspace := Workspace{Root: root}

	if err := workspace.New("0.1.0"); err != nil {
		t.Fatalf("new release: %v", err)
	}
	if got, want := currentGitBranch(t, root), "release/v0.1.0"; got != want {
		t.Fatalf("branch after new = %q, want %q", got, want)
	}
	if err := workspace.Advance(ChannelAlpha); err != nil {
		t.Fatalf("alpha release: %v", err)
	}
	if got, want := currentGitBranch(t, root), "release/v0.1.0-alpha.1"; got != want {
		t.Fatalf("branch after alpha = %q, want %q", got, want)
	}
	if err := workspace.Advance(ChannelAlpha); err != nil {
		t.Fatalf("second alpha release: %v", err)
	}
	if got, want := currentGitBranch(t, root), "release/v0.1.0-alpha.2"; got != want {
		t.Fatalf("branch after alpha.2 = %q, want %q", got, want)
	}
}

func TestInvalidAdvanceDoesNotCreateReleaseBranchOrCommit(t *testing.T) {
	root := setupGitReleaseRepo(t)
	beforeBranch := currentGitBranch(t, root)
	beforeCommits := gitOutputForTest(t, root, "rev-list", "--count", "HEAD")
	workspace := Workspace{Root: root}

	if err := workspace.Advance(ChannelBeta); err == nil {
		t.Fatal("expected beta to fail before alpha")
	}
	if got := currentGitBranch(t, root); got != beforeBranch {
		t.Fatalf("branch changed to %q, want %q", got, beforeBranch)
	}
	if releaseBranchExists(t, root, "release/v0.1.0") {
		t.Fatal("release branch was created for invalid beta transition")
	}
	if got := gitOutputForTest(t, root, "rev-list", "--count", "HEAD"); got != beforeCommits {
		t.Fatalf("commit count = %q, want %q", got, beforeCommits)
	}
}

func TestInvalidFinalizeDoesNotCreateReleaseBranchOrCommit(t *testing.T) {
	root := setupGitReleaseRepo(t)
	beforeBranch := currentGitBranch(t, root)
	beforeCommits := gitOutputForTest(t, root, "rev-list", "--count", "HEAD")
	workspace := Workspace{Root: root}

	if err := workspace.Finalize(); err == nil {
		t.Fatal("expected final to fail before rc")
	}
	if got := currentGitBranch(t, root); got != beforeBranch {
		t.Fatalf("branch changed to %q, want %q", got, beforeBranch)
	}
	if releaseBranchExists(t, root, "release/v0.1.0") {
		t.Fatal("release branch was created for invalid final transition")
	}
	if got := gitOutputForTest(t, root, "rev-list", "--count", "HEAD"); got != beforeCommits {
		t.Fatalf("commit count = %q, want %q", got, beforeCommits)
	}
}

func setupGitReleaseRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGitForTest(t, root, "init")
	runGitForTest(t, root, "config", "user.name", "Test User")
	runGitForTest(t, root, "config", "user.email", "test@example.invalid")
	if err := Initialize(root, "0.1.0"); err != nil {
		t.Fatalf("initialize release files: %v", err)
	}
	runGitForTest(t, root, "add", "-A")
	runGitForTest(t, root, "commit", "-m", "initial")
	return root
}

func currentGitBranch(t *testing.T, root string) string {
	t.Helper()
	return gitOutputForTest(t, root, "branch", "--show-current")
}

func releaseBranchExists(t *testing.T, root, branch string) bool {
	t.Helper()
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	cmd.Dir = root
	return cmd.Run() == nil
}

func runGitForTest(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func gitOutputForTest(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}
