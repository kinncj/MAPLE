# ADR-001: TUI implemented in Go (overrides PRD §5.9.6 Rust+Ratatui spec)

## Status
Accepted

## Context
PRD v1.2 §5.9.6 specified the `squad` TUI be built in **Rust using Ratatui**, with the explicit rationale that Rust aligns with the maintainer's kernel/systems environment and produces a smaller static binary. Go + Bubble Tea was considered and rejected in §8.

Before that decision could be acted on, a Go-based TUI was already partially implemented in `tui/` using the Go standard library and the Charm/Bubbletea ecosystem. The maintainer reviewed both the PRD decision and the existing Go implementation and chose to keep Go.

## Goals
- Resolve the conflict between the PRD spec and the existing codebase.
- Establish Go as the canonical TUI language so all future TUI work is unambiguous.
- Document the trade-offs so the decision can be revisited if circumstances change.

## Non-goals
- Re-litigating the original Rust vs Go choice in the abstract.
- Migrating any existing Go code to Rust.

## Proposal
Accept the existing Go TUI (`tui/`) as canonical. All future TUI development targets Go. The PRD §5.9.6 language requirement is superseded by this ADR.

The full PRD §5.9 feature set (splash, boot check, 4-pane dashboard, keybindings, design pane, logs pane, superpower picker, Omarchy themes, image rendering detection, command mode, help overlay) is still required — only the implementation language changes from Rust to Go.

## Alternatives Considered

### Option A: Rewrite in Rust + Ratatui (PRD spec)
**Description:** Discard the Go TUI, start fresh with Ratatui.
**Pros:** Matches PRD; static binary with zero runtime deps; Ratatui ecosystem is mature and well-suited to complex multi-pane UIs.
**Cons:** Throws away existing working code; adds Rust toolchain as a build dependency; significant implementation cost with no user-visible difference at feature parity.

### Option B: Keep Go (chosen)
**Description:** Continue with the existing Go implementation.
**Pros:** No throwaway work; Go toolchain already present in the repo (CI builds the TUI today); faster path to feature parity.
**Cons:** Go binaries are larger than equivalent Rust binaries; goroutine overhead vs Rust's zero-cost async — negligible at TUI scale; deviates from the PRD.

### Option C: Ship both
**Description:** Maintain Go TUI for now, Rust TUI as a future parallel track.
**Cons:** Double maintenance burden; no clear migration path; rejected outright.

## Trade-offs and Risks

- **Binary size:** Go TUI will be ~10–20 MB vs a ~2–5 MB Rust binary. Acceptable for a developer tool distributed via GitHub Releases.
- **Startup latency:** Go cold start is slightly higher than Rust, but the PRD's 150 ms target is easily achievable in Go for a TUI.
- **Ecosystem:** `bubbletea` / `lipgloss` / `bubbles` (Charm) is a mature, actively maintained Go TUI stack with broad community adoption. `ratatui` is also mature; neither is a risk.
- **PRD alignment:** This is a deliberate, documented override. Any future contributor reading the PRD must also read this ADR.

## Impact

### Cost (FinOps)
No cloud cost impact. TUI is a local binary.

### Operations (SRE)
No change. The Go TUI binary is already cross-compiled and released via GitHub Actions (`make build-tui-all`).

### Security
No change. TUI shells out to `gh`, `claude`, `opencode`, and local skill scripts — same trust model regardless of language.

### Team
Go toolchain (`go >= 1.22`) is already required to build the repo. No new toolchain dependency.

## Decision
Keep the Go TUI. Rust + Ratatui is rejected for this project. All PRD §5.9 functional requirements remain in force; only the implementation language is changed.

**Decided by:** Kinn (maintainer), 2026-04-20.

## Next Steps
- [x] File this ADR.
- [ ] Audit Go TUI against the full PRD §5.9 feature checklist.
- [ ] Implement any missing features in Go.
