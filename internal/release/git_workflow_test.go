package release

import (
	"os/exec"
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
