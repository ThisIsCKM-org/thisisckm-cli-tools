# Release Config and State Migration

Date: 2026-07-02

## Fixed
- Added `thisisckm release config` so repos can store their branch mapping in `release.config.json`.
- Added `release.json` as the canonical release state file and kept a compatibility fallback for legacy `version.json` state during migration.
- Updated the tag workflow and release commands to read the same release state the CLI writes.
