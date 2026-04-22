# Session — 2026-04-22 (continued)

> Earlier session covered v4.9.0/v4.9.1 fixes and roadmap setup. This section covers the second session block.

## What was done

- **RTK install fixed for Linux/Windows** (`scripts/install.sh`, `scripts/install.ps1`): Isolated `install_rtk()` function so `set -euo pipefail` failures don't abort the outer script. Binary location now uses `find -maxdepth 4` instead of assuming archive root — handles both root-level and subdirectory layouts (rtk releases vary).
- **`maple init` now installs RTK on any OS**: Decoupled RTK install into `tui/rtk_install.go` — `installRTK()`, `latestRTKVersion()`, `rtkPlatformTriple()`, `extractFromTarGz/Zip()`, `compareSemver()`. Handles tar.gz and zip, root-level or subdirectory binary, triple-named binaries (e.g. `rtk-aarch64-apple-darwin`).
- **Unit tests added** (`tui/rtk_install_test.go`): 11 tests — `compareSemver` (8 cases including `v4.9.2 < v4.10.0`), `rtkBinaryName`, `rtkPlatformTriple`, `isRTKBinaryName` (8 cases), `extractFromTarGz` (4 cases), `extractFromZip` (2 cases), network version format test. All passing.
- **CLAUDE.md created** at repo root: build dance, release process, versioning table, commit style rules, architecture invariants, open issues table, style guidelines.
- **README.md rewritten**: Accurate keybindings table, shared state protocol table, multiplexer setup, Prerequisites table, CLI commands reference.
- **v4.10.0 milestone closed** (shipped). **v4.11.0 milestone created**, issues #14–#18 moved there, due 2026-05-31.
- **Memory updated** to reflect v4.11.0 as active milestone.

## Decisions made

- **SOLID decoupling for RTK**: Pure functions, no global state, independently unit-testable. Network test gated behind `SKIP_NETWORK_TESTS`.
- **Semver comparison numeric, not lexicographic**: `compareSemver()` correctly handles `v4.9.2 < v4.10.0`.
- **v4.10.0 shipped, v4.11.0 is next**: Same May 31 deadline, same 5 issues.

## Fixes applied

| Bug | Root cause | Fix |
|-----|-----------|-----|
| RTK not installing on Arch/Ubuntu/Fedora | `mv "$TMP/rtk"` assumed archive root; `set -e` killed script on RTK failure | `find`-based binary location; isolated install function |
| `maple init` silently skipped RTK on Linux | Same archive layout assumption in Go code | `extractFromTarGz/Zip` now walks entire tree |

## Unfinished / follow-up

Issues #14–#18 are open on v4.11.0 (https://github.com/kinncj/MAPLE/milestone/2, due 2026-05-31):

| # | Title | Priority |
|---|-------|----------|
| [#18](https://github.com/kinncj/MAPLE/issues/18) | Verify `o` key opens session end-to-end | **Highest** |
| [#15](https://github.com/kinncj/MAPLE/issues/15) | Remove dead `dashActionOpenAgent`/`dashActionLaunch` cases | High |
| [#16](https://github.com/kinncj/MAPLE/issues/16) | `manualLaunchCopied` auto-reset after 2s | Medium |
| [#14](https://github.com/kinncj/MAPLE/issues/14) | Automate `tui/template` build dance in Makefile | Medium |
| [#17](https://github.com/kinncj/MAPLE/issues/17) | `PostToolUse` hooks for near-instant TUI refresh | Medium |

Unit tests still missing for:
- `tui/test_discovery.go` — per-framework detector functions
- `tui/terminal_spawn.go` — `shQuote`, `writeLaunchScript`

## Commits (this block)

```
69eca79 add unit tests for rtk installer
fa6443b maple init installs rtk on any OS; add CLAUDE.md; refresh README
464e77d fix rtk install: find binary in archive tree, isolate from set -e
```

## Releases cut

| Version | Summary |
|---------|---------|
| v4.9.0  | maple never exits — harness launch is async tea.Cmd; manual-launch modal |
| v4.9.1  | Test discovery fixes across all frameworks + id[:8] panic |
| v4.10.0 | RTK auto-install on any OS; CLAUDE.md; README rewrite; unit tests for rtk_install |
