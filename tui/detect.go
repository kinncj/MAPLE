package main

import (
	"os/exec"
	"strings"
)

// Tools holds the detection result for available AI and tooling.
type Tools struct {
	Claude    string // path or "" if not found
	OpenCode  string
	GHCopilot bool   // gh copilot extension present
	GH        string // gh CLI path
	Lefthook  string // lefthook binary path
}

// Detect checks which tools are available on PATH.
func Detect() Tools {
	t := Tools{}
	t.Claude, _ = exec.LookPath("claude")
	t.OpenCode, _ = exec.LookPath("opencode")
	t.GH, _ = exec.LookPath("gh")
	t.Lefthook, _ = exec.LookPath("lefthook")

	// Check gh copilot extension
	if t.GH != "" {
		out, err := exec.Command(t.GH, "extension", "list").Output()
		if err == nil && strings.Contains(string(out), "copilot") {
			t.GHCopilot = true
		}
	}
	return t
}

// HasAI returns true if at least one AI tool is available.
func (t Tools) HasAI() bool {
	return t.Claude != "" || t.OpenCode != "" || t.GHCopilot
}

// PreferredAI returns the name of the first available AI tool.
func (t Tools) PreferredAI() string {
	if t.Claude != "" {
		return "claude"
	}
	if t.OpenCode != "" {
		return "opencode"
	}
	if t.GHCopilot {
		return "gh copilot"
	}
	return ""
}

// Summary returns a human-readable list of detected tools.
func (t Tools) Summary() []string {
	var lines []string
	mark := func(ok bool, label string) string {
		if ok {
			return "  ✓  " + label
		}
		return "  ✗  " + label + " (not found)"
	}
	lines = append(lines, mark(t.Claude != "", "claude"))
	lines = append(lines, mark(t.OpenCode != "", "opencode"))
	lines = append(lines, mark(t.GHCopilot, "gh copilot"))
	lines = append(lines, mark(t.GH != "", "gh CLI"))
	lines = append(lines, mark(t.Lefthook != "", "lefthook"))
	return lines
}
