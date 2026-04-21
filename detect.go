package main

import (
	"os/exec"
	"strings"
)

// Tools holds the detection result for available AI and tooling.
type Tools struct {
	Claude    string // path or "" if not found
	OpenCode  string
	Copilot   string // GitHub Copilot CLI binary ("copilot"), supports -p for non-interactive
	GHCopilot bool   // gh copilot extension (explain/suggest shell commands only)
	GH        string // gh CLI path
	Lefthook  string // lefthook binary path
	NPX       string // npx (Node.js) — needed for skills marketplace
}

// Detect checks which tools are available on PATH.
func Detect() Tools {
	t := Tools{}
	t.Claude, _ = exec.LookPath("claude")
	t.OpenCode, _ = exec.LookPath("opencode")
	t.Copilot, _ = exec.LookPath("copilot")
	t.GH, _ = exec.LookPath("gh")
	t.Lefthook, _ = exec.LookPath("lefthook")
	t.NPX, _ = exec.LookPath("npx")

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
	kind  string // "claude" | "copilot" | "opencode"
	path  string // binary path
}

// HasAI returns true if at least one AI tool is available (for init/detection purposes).
func (t Tools) HasAI() bool {
	return t.Claude != "" || t.Copilot != "" || t.OpenCode != "" || t.GHCopilot
}

// HasReqAI returns true if at least one tool capable of requirements→Gherkin is available.
func (t Tools) HasReqAI() bool {
	return t.Claude != "" || t.Copilot != "" || t.OpenCode != ""
}

// PreferredAI returns the name of the first available AI tool.
func (t Tools) PreferredAI() string {
	if t.Claude != "" {
		return "claude"
	}
	if t.Copilot != "" {
		return "copilot"
	}
	if t.OpenCode != "" {
		return "opencode"
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
	check(t.GHCopilot, "gh-copilot")
	check(t.GH != "", "gh")
	check(t.Lefthook != "", "lefthook")
	check(t.NPX != "", "npx")

	var lines []string
	if len(found) > 0 {
		lines = append(lines, "✓ "+strings.Join(found, " · "))
	}
	if len(missing) > 0 {
		lines = append(lines, "✗ "+strings.Join(missing, " · ")+" (not found)")
	}
	return lines
}
