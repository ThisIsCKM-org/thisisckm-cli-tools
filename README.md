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

Keep release notes under `CHANGELOG.md` with an `Unreleased` section for active work.

## Release Flow

1. Work lands in `develop`.
2. A release branch is created with `thisisckm release new <version>`.
3. The release branch is updated through `alpha`, `beta`, `rc`, and `final`.
4. The CLI opens a PR from `release/*` into `main`.
5. GitHub Actions tags the merge commit and publishes the release from the tag.
