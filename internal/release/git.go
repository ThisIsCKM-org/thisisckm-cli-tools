package release

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func NewWorkspaceFromCWD() (*Workspace, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return &Workspace{Root: root}, nil
}

type Workspace struct {
	Root string
}

func (w *Workspace) Init(version string) error {
	return Initialize(w.Root, version)
}

func (w *Workspace) New(version string) error {
	if err := ensureClean(w.Root); err != nil {
		return err
	}
	state, err := loadOrSeed(w.Root, version)
	if err != nil {
		return err
	}
	next, err := state.WithNew(version)
	if err != nil {
		return err
	}
	branch := fmt.Sprintf("release/v%s", version)
	if err := createBranch(w.Root, branch); err != nil {
		return err
	}
	if err := Save(StateFile(w.Root), next); err != nil {
		return err
	}
	if err := commitAll(w.Root, fmt.Sprintf("chore(release): start %s", version)); err != nil {
		return err
	}
	return nil
}

func (w *Workspace) Advance(channel Channel) error {
	if err := ensureClean(w.Root); err != nil {
		return err
	}
	state, err := loadCurrent(w.Root)
	if err != nil {
		return err
	}
	next, err := state.Advance(channel)
	if err != nil {
		return err
	}
	if err := Save(StateFile(w.Root), next); err != nil {
		return err
	}
	version, err := next.ReleaseVersion()
	if err != nil {
		return err
	}
	if err := commitAll(w.Root, fmt.Sprintf("chore(release): %s", version)); err != nil {
		return err
	}
	return nil
}

func (w *Workspace) Finalize() error {
	if err := ensureClean(w.Root); err != nil {
		return err
	}
	state, err := loadCurrent(w.Root)
	if err != nil {
		return err
	}
	next, err := state.Finalize()
	if err != nil {
		return err
	}
	if err := PromoteChangelog(w.Root, next.BaseVersion, nil); err != nil {
		return err
	}
	if err := Save(StateFile(w.Root), next); err != nil {
		return err
	}
	if err := commitAll(w.Root, fmt.Sprintf("chore(release): finalize %s", next.BaseVersion)); err != nil {
		return err
	}
	if err := openPullRequest(w.Root, branchName(w.Root), "main", next.BaseVersion); err != nil {
		return err
	}
	return nil
}

func loadCurrent(root string) (State, error) {
	statePath := StateFile(root)
	if _, err := os.Stat(statePath); err != nil {
		if os.IsNotExist(err) {
			return State{}, errors.New("release metadata not initialized; run `thisisckm release init <version>` first")
		}
		return State{}, err
	}
	return Load(statePath)
}

func loadOrSeed(root, version string) (State, error) {
	statePath := StateFile(root)
	if _, err := os.Stat(statePath); err != nil {
		if os.IsNotExist(err) {
			return Seed(version), nil
		}
		return State{}, err
	}
	return Load(statePath)
}

func ensureClean(root string) error {
	if !isGitRepo(root) {
		return nil
	}
	out, err := gitOutput(root, "status", "--porcelain")
	if err != nil {
		return err
	}
	if strings.TrimSpace(out) != "" {
		return errors.New("working tree is dirty; commit or stash changes before running release commands")
	}
	return nil
}

func isGitRepo(root string) bool {
	_, err := os.Stat(filepathJoin(root, ".git"))
	return err == nil
}

func branchName(root string) string {
	name, err := gitOutput(root, "branch", "--show-current")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(name)
}

func createBranch(root, branch string) error {
	if branch == "" {
		return errors.New("branch name is required")
	}
	if err := gitRun(root, "switch", "-c", branch); err != nil {
		return err
	}
	return nil
}

func commitAll(root, message string) error {
	if err := gitRun(root, "config", "user.name", "ThisIsCKM Release Bot"); err != nil {
		return err
	}
	if err := gitRun(root, "config", "user.email", "release@thisisckm.local"); err != nil {
		return err
	}
	if err := gitRun(root, "add", "version.json", "CHANGELOG.md"); err != nil {
		return err
	}
	if err := gitRun(root, "commit", "-m", message); err != nil {
		return err
	}
	return nil
}

func openPullRequest(root, head, base, version string) error {
	if head == "" || base == "" {
		return nil
	}
	if !strings.HasPrefix(head, "release/") {
		return nil
	}
	if _, err := exec.LookPath("gh"); err != nil {
		fmt.Printf("release branch prepared for %s; open a PR from %s into %s\n", version, head, base)
		return nil
	}
	cmd := exec.Command("gh", "pr", "create", "--base", base, "--head", head, "--title", fmt.Sprintf("Release %s", version), "--body", fmt.Sprintf("Prepare release %s.", version))
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitRun(root string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitOutput(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	return string(out), err
}

func filepathJoin(elem ...string) string {
	return strings.Join(elem, string(os.PathSeparator))
}
