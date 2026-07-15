# Changelog

## v0.52.0 (2026-07-14)

### Changed

- replace the Cobra command parser with Kong while preserving documented lint, format, summary, and version invocations.
- apply command-line configuration overrides only when the corresponding flag was explicitly provided, preserving values loaded from config files and `CCLINT_*` environment variables.

## v0.47.1 (2026-06-04)

### Fixes

- load `.cclintrc.*` from the explicit `--root` project.
- serialize shared CUE validator access for race-free validation.
- update `golang.org/x/net` to the fixed `idna` vulnerability floor.

### Other

- delete unused DetermineSeverity helper.
