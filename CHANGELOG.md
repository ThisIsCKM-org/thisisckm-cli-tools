# Changelog

All notable changes to this project will be documented in this file.

Release notes are staged under `changelogs/` and consolidated here when a version is cut.

## Unreleased
### Added
### Fixed
### Changed
### Removed
### Breaking

## [0.1.0-alpha.1] - 2026-07-03
# Initial Change Log Entry

Date: 2026-06-13

## Added
- Initial `thisisckm-cli-tools` repository scaffold.
- Release CLI foundation with `init`, `new`, `alpha`, `beta`, `rc`, and `final` workflow commands.
- Branch-based release flow using `develop`, `release/*`, and `main`.
- Root `CHANGELOG.md` with an `Unreleased` section for ongoing work.
- Release process documentation with prerelease, selective phase release, and hotfix examples.
- Prerelease changelog promotion so alpha, beta, and rc releases update `CHANGELOG.md`.
- Prerelease branches use the exact computed version, such as `release/v0.1.0-alpha.2`.

# Sync Develop Release Gate

Date: 2026-07-01

## Added
- Added `thisisckm release sync-develop` to prepare a `sync/main-into-develop` PR instead of pushing directly to `develop`.
- Added a release gate that blocks `init`, `new`, `alpha`, `beta`, `rc`, and `final` when `develop` is behind `main`.

# Release Config and State Migration

Date: 2026-07-02

## Fixed
- Added `thisisckm release config` so repos can store their branch mapping in `release.config.json`.
- Added `release.json` as the canonical release state file and kept a compatibility fallback for legacy `version.json` state during migration.
- Updated the tag workflow and release commands to read the same release state the CLI writes.

### Fixed
- Fix prerelease release tagging and PR metadata

### Added
- Add installable AI agent skills and embedded installer

