---
name: prepare-release-pr
description: Prepare a release pull request for thisisckm-cli-tools. Use when the task is to review or stage release metadata and open the release PR with the release CLI.
---

# Prepare Release PR

Use this skill only for release PR preparation.

## Workflow

0. Verify the `thisisckm` CLI is installed and available before doing any release PR work. If it is missing, stop and ask the human to install it.
1. Read `release.json`, the current release branch, and `release.config.json` first so the release line is inferred from repo state when possible.
2. Do not assume whether the release is alpha, beta, rc, or final.
3. Ask the human for release type or version details only when the repo state is missing, conflicting, or otherwise cannot determine the release line safely.
4. Use `thisisckm release` as the source of truth once the release type is known.
5. Follow the existing release lifecycle:
   - `init`
   - `config`
   - `new`
   - `alpha`
   - `beta`
   - `rc`
   - `final`
   - `sync-develop`
6. Respect staged changelog flow under `changelogs/` and the published history in `CHANGELOG.md`.
7. Create or update the release pull request from the release branch into the resolved main branch when the repo and hosting tools support it.
8. If PR creation is not available, stop after preparing the release branch, version metadata, changelog, and PR body so the human can open it manually.
9. Call out when `sync-develop` is required before release work can continue.
10. Keep the skill focused on release PR prep, not normal feature PR flow.
