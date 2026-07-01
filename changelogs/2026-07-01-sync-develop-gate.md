# Sync Develop Release Gate

Date: 2026-07-01

## Added
- Added `thisisckm release sync-develop` to prepare a `sync/main-into-develop` PR instead of pushing directly to `develop`.
- Added a release gate that blocks `init`, `new`, `alpha`, `beta`, `rc`, and `final` when `develop` is behind `main`.
