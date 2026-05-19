package main

import (
	"os/exec"
	"strings"
)

// Harness describes one launchable AI harness discovered on PATH.
type Harness struct {
	Key         string // cli key used by dashboard/session state
	Label       string // generic display label
	ReqLabel    string // label used in requirements picker
	Bin         string // resolved binary path
	SupportsReq bool   // supports requirements -> gherkin generation
}

// Tools holds the detection result for available AI and tooling.
type Tools struct {
	Claude    string // path or "" if not found
	OpenCode  string
	Copilot   string // GitHub Copilot CLI binary ("copilot"), supports -p for non-interactive
	Cursor    string // Cursor Agent CLI ("cursor-agent") or fallback "cursor"
	GHCopilot bool   // gh copilot extension (explain/suggest shell commands only)
	GH        string // gh CLI path
	Lefthook  string // lefthook binary path
	NPX       string // npx (Node.js) — needed for skills marketplace
	RTK       string // rtk CLI proxy — reduces LLM token consumption 60-90%
}

// Harnesses returns discovered launchable harnesses in canonical order.
func (t Tools) Harnesses() []Harness {
	var out []Harness
	if t.Claude != "" {
		out = append(out, Harness{Key: "claude", Label: "Claude Code", ReqLabel: "Claude Code", Bin: t.Claude, SupportsReq: true})
	}
	if t.Copilot != "" {
		out = append(out, Harness{Key: "copilot", Label: "GitHub Copilot", ReqLabel: "GitHub Copilot", Bin: t.Copilot, SupportsReq: true})
	}
	if t.OpenCode != "" {
		out = append(out, Harness{Key: "opencode", Label: "OpenCode", ReqLabel: "OpenCode", Bin: t.OpenCode, SupportsReq: true})
	}
	if t.Cursor != "" {
		out = append(out, Harness{Key: "cursor", Label: "Cursor", ReqLabel: "Cursor Agent", Bin: t.Cursor, SupportsReq: true})
	}
	return out
}

// Detect checks which tools are available on PATH.
func Detect() Tools {
	t := Tools{}
	t.Claude, _ = exec.LookPath("claude")
	t.OpenCode, _ = exec.LookPath("opencode")
	t.Copilot, _ = exec.LookPath("copilot")
	t.Cursor, _ = exec.LookPath("cursor-agent")
	if t.Cursor == "" {
		t.Cursor, _ = exec.LookPath("cursor")
	}
	t.GH, _ = exec.LookPath("gh")
	t.Lefthook, _ = exec.LookPath("lefthook")
	t.NPX, _ = exec.LookPath("npx")
	t.RTK, _ = exec.LookPath("rtk")

	// Check gh copilot extension (shell-command helper only)
	if t.GH != "" {
		out, err := exec.Command(t.GH, "extension", "list").Output()
		if err == nil && strings.Contains(string(out), "copilot") {
			t.GHCopilot = true
		}
	}
	return t
}

// aiOption represents one available AI tool that supports general-purpose prompt → text.
type aiOption struct {
	label string // display name
	kind  string // "claude" | "copilot" | "opencode" | "cursor"
	path  string // binary path
}

// HasAI returns true if at least one AI tool is available (for init/detection purposes).
func (t Tools) HasAI() bool {
	return len(t.Harnesses()) > 0 || t.GHCopilot
}

// HasReqAI returns true if at least one tool capable of requirements→Gherkin is available.
func (t Tools) HasReqAI() bool {
	for _, h := range t.Harnesses() {
		if h.SupportsReq {
			return true
		}
	}
	return false
}

// PreferredAI returns the name of the first available AI tool.
func (t Tools) PreferredAI() string {
	if hs := t.Harnesses(); len(hs) > 0 {
		return hs[0].Key
	}
	if t.GHCopilot {
		return "gh copilot"
	}
	return ""
}

// Summary returns compact tool status: found tools as pills, missing as a note.
func (t Tools) Summary() []string {
	var found, missing []string
	check := func(ok bool, label string) {
		if ok {
			found = append(found, label)
		} else {
			missing = append(missing, label)
		}
	}
	check(t.Claude != "", "claude")
	check(t.Copilot != "", "copilot")
	check(t.OpenCode != "", "opencode")
	check(t.Cursor != "", "cursor")
	check(t.GHCopilot, "gh-copilot")
	check(t.GH != "", "gh")
	check(t.Lefthook != "", "lefthook")
	check(t.NPX != "", "npx")
	check(t.RTK != "", "rtk")

	var lines []string
	if len(found) > 0 {
		lines = append(lines, "✓ "+strings.Join(found, " · "))
	}
	if len(missing) > 0 {
		lines = append(lines, "✗ "+strings.Join(missing, " · ")+" (not found)")
	}
	return lines
}
