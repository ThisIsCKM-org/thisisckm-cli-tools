# ThisIsCKM CLI Tools

`thisisckm-cli-tools` is an umbrella repository for release and developer automation tools.

The first tool is the release CLI.

## Branch Flow

- `develop` is the active integration branch.
- `release/*` branches are created for release preparation.
- `main` is the protected release-ready branch.
- release tags are cut from `main` only.

## Commands

- `thisisckm release init <version>` bootstraps `version.json` and `CHANGELOG.md`.
- `thisisckm release new <version>` starts a release line and creates `release/v<version>`.
- `thisisckm release alpha` advances prerelease state.
- `thisisckm release beta` advances prerelease state.
- `thisisckm release rc` advances prerelease state.
- `thisisckm release final` finalizes the release branch and opens the PR to `main`.

## Changelog

- Store in-progress entries under `changelogs/` as one file per change log item.
- Consolidate staged entries into `CHANGELOG.md` whenever `alpha`, `beta`, `rc`, or `final` cuts a release version.
- Keep `CHANGELOG.md` as the published release history.

## Release Flow

1. Work lands in `develop`.
2. A release branch is created with `thisisckm release new <version>`.
3. The release branch is updated through `alpha`, `beta`, `rc`, and `final`.
4. The CLI opens a PR from `release/*` into `main`.
5. Tags are cut from `main` after the release PR is merged.

## Version Examples

For base version `0.1.0`, prerelease versions progress like this:

```text
0.1.0-alpha.1
0.1.0-alpha.2
0.1.0-beta.1
0.1.0-rc.1
0.1.0
```

Git tags use the same version with a `v` prefix:

```text
v0.1.0-alpha.1
v0.1.0-beta.1
v0.1.0-rc.1
v0.1.0
```

## Example Release Cycle

Start by initializing release metadata after the project scaffold is ready:

```bash
thisisckm release init 0.1.0
```

Feature and bugfix work then happens on short-lived branches and lands in `develop`:

```text
feature/login-hardening -> develop
bugfix/session-timeout -> develop
feature/audit-logging -> develop
```

When `develop` is ready for the first prerelease, start the release line and advance to alpha:

```bash
thisisckm release new 0.1.0
thisisckm release alpha
```

Each release command uses a branch named for the exact release version, such as `release/v0.1.0-alpha.1`, promotes staged changelog entries into a versioned `CHANGELOG.md` section such as `## [0.1.0-alpha.1] - 2026-06-30`, and clears the staged entry files.

The release branch is then stabilized through the prerelease stages:

```bash
thisisckm release beta
thisisckm release rc
thisisckm release final
```

`final` promotes staged changelog entries into `CHANGELOG.md`, marks the version as stable, and opens or updates the release PR into `main`.

## Selective Release Example

Sometimes `develop` contains multiple completed phases, but one phase is not ready to ship.

Example: product hardening for version `2.1.0` has five independent phases:

```text
Phase 1 -> merged into develop
Phase 2 -> merged into develop, but later found to have a bug
Phase 3 -> merged into develop
Phase 4 -> still in progress
Phase 5 -> still in progress
```

If Phase 1 and Phase 3 are ready for alpha or beta, but Phase 2 needs more time, do not cut the release from `develop` while the broken Phase 2 code is still present. A release branch created from that state would include Phase 2.

Recommended approach: revert Phase 2 from `develop`, then release Phase 1 and Phase 3.

```bash
git switch develop
git revert <phase-2-merge-commit>
thisisckm release new 2.1.0
thisisckm release alpha
```

The alpha version would be:

```text
2.1.0-alpha.1
```

After Phase 2 is fixed, merge the fix back into `develop` and continue the release train:

```bash
thisisckm release beta
```

That produces:

```text
2.1.0-beta.1
```

Alternative approach: build the release branch selectively from the last stable branch and cherry-pick only Phase 1 and Phase 3. This can work for advanced cases, but it requires careful commit tracking and is easier to get wrong than reverting the bad phase from `develop`.

```text
main -> release/v2.1.0
Phase 1 commits -> release/v2.1.0
Phase 3 commits -> release/v2.1.0
```

Use the selective approach only when reverting from `develop` is not practical.

## Hotfix Example

For urgent production fixes, branch from the stable release branch, then merge the fix back everywhere it is needed:

```text
main -> hotfix/security-patch -> main
hotfix/security-patch -> develop
hotfix/security-patch -> release/v2.1.0, if a release is in progress
```

Tags should still be cut from `main` after the final release or hotfix merge.
