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
