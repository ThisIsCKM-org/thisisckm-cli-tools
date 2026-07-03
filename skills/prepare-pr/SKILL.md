---
name: prepare-pr
description: Prepare a normal feature or bug pull request for thisisckm-cli-tools. Use when the task is to stage changelog entries, validate a small code change, and draft a non-release PR.
---

# Prepare PR

Use this skill for normal feature and bug pull requests only.

## Workflow

0. Verify the `thisisckm` CLI is installed and available before doing any PR prep work. If it is missing, stop and ask the human to install it.
1. Resolve the active development branch in this order:
   - `release.config.json` if it exists
   - git branch metadata
   - the default `develop`
2. If the repo is not on the expected develop branch, stop and ask the human for approval before changing branches or proceeding.
3. Verify the worktree is safe to operate on before making PR-ready changes.
4. Add the required gates before drafting the PR:
   - create or update a staged changelog entry with `thisisckm changelog`
   - run a targeted test or quick validation for the changed path
   - keep the change scoped to the intended feature or bugfix
   - exclude unrelated release automation
5. Prefer staged changelog entries over editing `CHANGELOG.md` directly.
6. If multiple changes are unrelated, split them into separate changelog entries.
7. Draft the PR from the code diff plus the staged changelog entry.
8. Create the pull request from the feature branch into the resolved develop branch when the repo and hosting tools support it.
9. If PR creation is not available, stop after preparing the branch, diff, and PR body so the human can open it manually.
