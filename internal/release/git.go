package release

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
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
	if err := w.ensureDevelopSynced(); err != nil {
		return err
	}
	return Initialize(w.Root, version)
}

func (w *Workspace) New(version string) error {
	if err := w.ensureDevelopSynced(); err != nil {
		return err
	}
	return w.startNew(version, true)
}

func (w *Workspace) startNew(version string, openPR bool) error {
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
	branch := releaseBranchName(version)
	if err := createBranch(w.Root, branch); err != nil {
		return err
	}
	if err := Save(StateFile(w.Root), next); err != nil {
		return err
	}
	if err := commitAll(w.Root, fmt.Sprintf("chore(release): start %s", version)); err != nil {
		return err
	}
	if err := pushBranch(w.Root, branch); err != nil {
		return err
	}
	if !openPR {
		return nil
	}
	return openPullRequest(w.Root, branch, resolveBranchName(w.Root, "main"), version, true)
}

func (w *Workspace) Advance(channel Channel) error {
	if err := w.ensureDevelopSynced(); err != nil {
		return err
	}
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
	version, err := next.ReleaseVersion()
	if err != nil {
		return err
	}
	branch := releaseBranchName(version)
	if branchName(w.Root) != branch {
		if err := createBranch(w.Root, branch); err != nil {
			return err
		}
	}
	if err := promoteStagedChangelog(w.Root, version); err != nil {
		return err
	}
	if err := Save(StateFile(w.Root), next); err != nil {
		return err
	}
	if err := commitAll(w.Root, fmt.Sprintf("chore(release): %s", version)); err != nil {
		return err
	}
	if err := pushBranch(w.Root, branch); err != nil {
		return err
	}
	return openPullRequest(w.Root, branch, resolveBranchName(w.Root, "main"), version, true)
}

func (w *Workspace) Finalize() error {
	if err := w.ensureDevelopSynced(); err != nil {
		return err
	}
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
	branch := releaseBranchName(next.BaseVersion)
	if branchName(w.Root) != branch {
		if err := createBranch(w.Root, branch); err != nil {
			return err
		}
	}
	if err := promoteStagedChangelog(w.Root, next.BaseVersion); err != nil {
		return err
	}
	if err := Save(StateFile(w.Root), next); err != nil {
		return err
	}
	if err := commitAll(w.Root, fmt.Sprintf("chore(release): finalize %s", next.BaseVersion)); err != nil {
		return err
	}
	if err := pushBranch(w.Root, branch); err != nil {
		return err
	}
	return openPullRequest(w.Root, branch, resolveBranchName(w.Root, "main"), next.BaseVersion, false)
}

func (w *Workspace) SyncDevelop() error {
	return w.SyncBranches("main", "develop")
}

func (w *Workspace) SyncBranches(source, target string) error {
	if source == "" || target == "" {
		return errors.New("source and target branches are required")
	}
	if source == target {
		return errors.New("source and target branches must be different")
	}
	if !isGitRepo(w.Root) {
		return errors.New("release sync requires a git repository")
	}
	if err := ensureFullyClean(w.Root); err != nil {
		return err
	}
	sourceName, targetName := branchNamesForRoles(w.Root, source, target)
	if hasOriginRemote(w.Root) {
		if err := gitRun(w.Root, "fetch", "origin", sourceName, targetName); err != nil {
			return err
		}
	}
	sourceRef := bestBranchRef(w.Root, source)
	targetRef := bestBranchRef(w.Root, target)
	if sourceRef == "" || targetRef == "" {
		return fmt.Errorf("could not find %s or %s branches for sync", sourceName, targetName)
	}
	syncBranch := syncBranchName(sourceName, targetName)
	if localBranchExists(w.Root, syncBranch) {
		if err := gitRun(w.Root, "switch", syncBranch); err != nil {
			return err
		}
		if err := gitRun(w.Root, "merge", "--no-edit", targetRef); err != nil {
			return err
		}
	} else if err := gitRun(w.Root, "switch", "-c", syncBranch, targetRef); err != nil {
		return err
	}
	if err := gitRun(w.Root, "merge", "--no-edit", sourceRef); err != nil {
		return err
	}
	if err := pushBranch(w.Root, syncBranch); err != nil {
		return err
	}
	return openSyncPullRequest(w.Root, syncBranch, targetName, sourceName)
}

func (w *Workspace) ensureDevelopSynced() error {
	return ensureBranchContains(w.Root, "main", "develop")
}

func ensureBranchContains(root, source, target string) error {
	if !isGitRepo(root) {
		return nil
	}
	sourceName, targetName := branchNamesForRoles(root, source, target)
	if hasOriginRemote(root) {
		if err := gitRun(root, "fetch", "origin", sourceName, targetName); err != nil {
			return err
		}
	}
	sourceRef := bestBranchRef(root, source)
	targetRef := bestBranchRef(root, target)
	if sourceRef == "" || targetRef == "" {
		return nil
	}
	cmd := exec.Command("git", "merge-base", "--is-ancestor", sourceRef, targetRef)
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s is out of sync with %s; run `thisisckm release sync-develop` before release commands", targetName, sourceName)
	}
	return nil
}

func bestBranchRef(root, branch string) string {
	for _, candidate := range branchCandidates(root, branch) {
		if candidate == "" {
			continue
		}
		remoteRef := "refs/remotes/origin/" + candidate
		if gitRefExists(root, remoteRef) {
			return "origin/" + candidate
		}
		localRef := "refs/heads/" + candidate
		if gitRefExists(root, localRef) {
			return candidate
		}
	}
	return ""
}

func gitRefExists(root, ref string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", ref)
	cmd.Dir = root
	return cmd.Run() == nil
}

func loadCurrent(root string) (State, error) {
	state, _, err := loadStateFile(root)
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, errors.New("release metadata not initialized; run `thisisckm release init <version>` first")
		}
		return State{}, err
	}
	return state, nil
}

func loadOrSeed(root, version string) (State, error) {
	state, _, err := loadStateFile(root)
	if err != nil {
		if os.IsNotExist(err) {
			return Seed(version), nil
		}
		return State{}, err
	}
	return state, nil
}

func promoteStagedChangelog(root, version string) error {
	items, err := CollectStagedEntries(root)
	if err != nil {
		return err
	}
	if err := PromoteChangelog(root, version, items); err != nil {
		return err
	}
	return ClearStagedEntries(root)
}

func releaseBranchName(version string) string {
	return fmt.Sprintf("release/v%s", version)
}

func syncBranchName(source, target string) string {
	return fmt.Sprintf("sync/%s-into-%s", source, target)
}

func ensureClean(root string) error {
	if !isGitRepo(root) {
		return nil
	}
	out, err := gitStatus(root)
	if err != nil {
		return err
	}
	for _, rawLine := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.TrimSpace(rawLine) == "" {
			continue
		}
		path := releasePathFromStatus(rawLine)
		if path == "" {
			return errors.New("unable to parse git status output")
		}
		if !isControlledReleasePath(path) {
			return errors.New("working tree has unrelated changes; commit or stash them before running release commands")
		}
	}
	return nil
}

func ensureFullyClean(root string) error {
	out, err := gitStatus(root)
	if err != nil {
		return err
	}
	if strings.TrimSpace(out) != "" {
		return errors.New("working tree is dirty; commit or stash changes before syncing branches")
	}
	return nil
}

func gitStatus(root string) (string, error) {
	return gitOutput(root, "status", "--porcelain=v1", "--untracked-files=all")
}

func releasePathFromStatus(line string) string {
	if len(line) < 4 {
		return ""
	}
	return strings.TrimSpace(line[3:])
}

func isControlledReleasePath(path string) bool {
	switch path {
	case "release.config.json", "release.json", "CHANGELOG.md":
		return true
	}
	return strings.HasPrefix(path, "changelogs/")
}

func isGitRepo(root string) bool {
	_, err := os.Stat(filepathJoin(root, ".git"))
	return err == nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
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
	if localBranchExists(root, branch) {
		return gitRun(root, "switch", branch)
	}
	if remoteBranchExists(root, "origin", branch) {
		if err := gitRun(root, "fetch", "origin", branch); err != nil {
			return err
		}
		return gitRun(root, "switch", "--track", "-c", branch, "origin/"+branch)
	}
	if err := gitRun(root, "switch", "-c", branch); err != nil {
		return err
	}
	return nil
}

func localBranchExists(root, branch string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	cmd.Dir = root
	return cmd.Run() == nil
}

func commitAll(root, message string) error {
	if err := gitRun(root, "config", "user.name", "ThisIsCKM Release Bot"); err != nil {
		return err
	}
	if err := gitRun(root, "config", "user.email", "release@thisisckm.local"); err != nil {
		return err
	}
	args := []string{"add", "-A", "--"}
	if fileExists(filepathJoin(root, "release.config.json")) {
		args = append(args, "release.config.json")
	}
	args = append(args, "release.json", "CHANGELOG.md", "changelogs")
	if err := gitRun(root, args...); err != nil {
		return err
	}
	if err := gitRun(root, "commit", "-m", message); err != nil {
		return err
	}
	return nil
}

func pushBranch(root, branch string) error {
	if branch == "" {
		return nil
	}
	remoteURL, err := originRemoteURL(root)
	if err != nil {
		return nil
	}
	if !looksLikeGitHubRemote(strings.TrimSpace(remoteURL)) {
		return nil
	}
	if err := gitRun(root, "push", "-u", "origin", branch); err != nil {
		return err
	}
	return nil
}

func openPullRequest(root, head, base, version string, draft bool) error {
	if head == "" || base == "" {
		return nil
	}
	if !strings.HasPrefix(head, "release/") {
		return nil
	}
	remoteURL, err := gitOutput(root, "remote", "get-url", "origin")
	if err != nil {
		return nil
	}
	if !looksLikeGitHubRemote(strings.TrimSpace(remoteURL)) {
		return nil
	}
	if !remoteBranchExists(root, "origin", base) {
		fmt.Printf("release branch prepared for %s; create or sync %s before opening a PR from %s\n", version, base, head)
		return nil
	}
	if _, err := exec.LookPath("gh"); err != nil {
		fmt.Printf("release branch prepared for %s; open a PR from %s into %s\n", version, head, base)
		return nil
	}
	title := releasePullRequestTitle(version)
	body := releasePullRequestBody(root, version)
	if number, err := existingPullRequestNumber(root, head, base); err != nil {
		return err
	} else if number > 0 {
		return editPullRequest(root, number, title, body)
	}
	args := []string{"pr", "create", "--base", base, "--head", head, "--title", title, "--body", body}
	if draft {
		args = append(args, "--draft")
	}
	cmd := exec.Command("gh", args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func openSyncPullRequest(root, head, base, source string) error {
	if head == "" || base == "" {
		return nil
	}
	remoteURL, err := originRemoteURL(root)
	if err != nil {
		return nil
	}
	if !looksLikeGitHubRemote(strings.TrimSpace(remoteURL)) {
		return nil
	}
	if !remoteBranchExists(root, "origin", base) {
		fmt.Printf("sync branch prepared from %s; create or sync %s before opening a PR from %s\n", source, base, head)
		return nil
	}
	if _, err := exec.LookPath("gh"); err != nil {
		fmt.Printf("sync branch prepared; open a PR from %s into %s\n", head, base)
		return nil
	}
	title := fmt.Sprintf("Sync %s into %s", source, base)
	body := fmt.Sprintf("## Sync\nMerge `%s` back into `%s` after release metadata updates.\n", source, base)
	if number, err := existingPullRequestNumber(root, head, base); err != nil {
		return err
	} else if number > 0 {
		return editPullRequest(root, number, title, body)
	}
	cmd := exec.Command("gh", "pr", "create", "--base", base, "--head", head, "--title", title, "--body", body)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func existingPullRequestNumber(root, head, base string) (int, error) {
	cmd := exec.Command("gh", "pr", "list", "--head", head, "--base", base, "--state", "open", "--json", "number", "--limit", "1")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	var prs []struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(out, &prs); err != nil {
		return 0, err
	}
	if len(prs) == 0 {
		return 0, nil
	}
	return prs[0].Number, nil
}

func editPullRequest(root string, number int, title, body string) error {
	cmd := exec.Command("gh", "pr", "edit", strconv.Itoa(number), "--title", title, "--body", body)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func releasePullRequestTitle(version string) string {
	switch {
	case strings.Contains(version, "-alpha."):
		return fmt.Sprintf("Alpha Release %s", version)
	case strings.Contains(version, "-beta."):
		return fmt.Sprintf("Beta Release %s", version)
	case strings.Contains(version, "-rc."):
		return fmt.Sprintf("Release Candidate %s", version)
	default:
		return fmt.Sprintf("Stable Release %s", version)
	}
}

func releasePullRequestBody(root, version string) string {
	unreleased, err := UnreleasedChangelog(root)
	if err != nil || strings.TrimSpace(unreleased) == "" {
		unreleased = "No unreleased changelog entries found."
	}
	return fmt.Sprintf("## Release\n%s\n\n## Changelog\n%s\n", version, strings.TrimSpace(unreleased))
}

func remoteBranchExists(root, remote, branch string) bool {
	cmd := exec.Command("git", "ls-remote", "--heads", remote, branch)
	cmd.Dir = root
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}

func hasOriginRemote(root string) bool {
	_, err := originRemoteURL(root)
	return err == nil
}

func originRemoteURL(root string) (string, error) {
	return gitOutput(root, "remote", "get-url", "origin")
}

func looksLikeGitHubRemote(remote string) bool {
	return strings.Contains(remote, "github.com")
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
