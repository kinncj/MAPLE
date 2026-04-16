package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Splash screen ────────────────────────────────────────────────────────────

const asciiLogo = `
  █████╗ ██╗      ███████╗ ██████╗ ██╗   ██╗ █████╗ ██████╗
 ██╔══██╗██║      ██╔════╝██╔═══██╗██║   ██║██╔══██╗██╔══██╗
 ███████║██║█████╗███████╗██║   ██║██║   ██║███████║██║  ██║
 ██╔══██║██║╚════╝╚════██║██║▄▄ ██║██║   ██║██╔══██║██║  ██║
 ██║  ██║██║      ███████║╚██████╔╝╚██████╔╝██║  ██║██████╔╝
 ╚═╝  ╚═╝╚═╝      ╚══════╝ ╚══▀▀═╝  ╚═════╝ ╚═╝  ╚═╝╚═════╝
`

const tagline = "Orchestrated multi-agent SDLC"

func renderSplash(t Theme, w, h int) string {
	logo := lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Render(asciiLogo)

	tag := lipgloss.NewStyle().
		Foreground(t.Accent).
		Italic(true).
		Render(tagline)

	version := lipgloss.NewStyle().
		Foreground(t.Muted).
		Render("v3.5.0")

	content := lipgloss.JoinVertical(lipgloss.Center, logo, tag, "", version)

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}

// ─── Help overlay ─────────────────────────────────────────────────────────────

func renderHelp(t Theme, w, h int) string {
	header := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("Keybindings")

	bindings := [][]string{
		{"Tab / Shift+Tab", "Cycle panes"},
		{"j / k", "Move down / up"},
		{"Enter", "Detail view"},
		{"s", "Stories pane"},
		{"a", "Agents pane"},
		{"p", "PRs pane"},
		{"q", "QA pane"},
		{"d", "Design pane"},
		{"l", "Logs pane"},
		{"F", "Fire superpower"},
		{"n", "New story / spike / ADR"},
		{"/", "Search"},
		{":", "Command mode"},
		{"r", "Refresh pane"},
		{"g g", "Go to top"},
		{"G", "Go to bottom"},
		{"?", "Toggle help"},
		{"Ctrl+c", "Quit"},
	}

	commands := [][]string{
		{":kickoff <id>", "Start pipeline for story"},
		{":sync", "Sync stories ↔ Issues"},
		{":a11y <id>", "Run a11y audit for story"},
		{":theme <name>", "Switch theme"},
		{":resume <sp>", "Resume paused superpower"},
		{":debug", "Toggle debug log tee"},
	}

	keyStyle := lipgloss.NewStyle().Foreground(t.Accent).Width(22)
	descStyle := lipgloss.NewStyle().Foreground(t.Foreground)

	var rows []string
	rows = append(rows, header, "")
	for _, b := range bindings {
		rows = append(rows, keyStyle.Render(b[0])+descStyle.Render(b[1]))
	}
	rows = append(rows, "", lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("Commands"))
	for _, c := range commands {
		rows = append(rows, keyStyle.Render(c[0])+descStyle.Render(c[1]))
	}
	rows = append(rows, "", lipgloss.NewStyle().Foreground(t.Muted).Render("Press ? to close"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(1, 3).
		Render(strings.Join(rows, "\n"))

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

// ─── Boot check ───────────────────────────────────────────────────────────────

type BootCheck struct {
	Label  string
	Check  func() bool
	Fix    string
}

func runBootChecks() []string {
	checks := []BootCheck{
		{"gh auth status", checkGHAuth, "Run: gh auth login"},
		{"project.config.yaml", checkProjectConfig, "Run: ai-squad init"},
		{"claude / opencode", checkAITool, "Install claude or opencode"},
	}

	var lines []string
	for _, c := range checks {
		if c.Check() {
			lines = append(lines, fmt.Sprintf("  ✓ %s", c.Label))
		} else {
			lines = append(lines, fmt.Sprintf("  ✗ %s  →  %s", c.Label, c.Fix))
		}
	}
	return lines
}

func checkGHAuth() bool {
	// TODO: exec gh auth status
	return true
}

func checkProjectConfig() bool {
	// TODO: os.Stat("project.config.yaml")
	return true
}

func checkAITool() bool {
	// TODO: exec.LookPath("claude") || exec.LookPath("opencode")
	return true
}

// ─── Pane stubs ───────────────────────────────────────────────────────────────
// Each pane is a lightweight struct with View(w, h int) string.
// Full implementations wire to gh CLI, JSONL log tails, etc.

type StoriesPane struct{ cursor int }
type AgentsPane struct{ cursor int }
type PRsPane struct{ cursor int }
type QAPane struct{ cursor int }
type DesignPane struct{ cursor int }
type LogsPane struct {
	cursor int
	lines  []string
	last   time.Time
}
type SuperpowerPane struct{ cursor int }

func (p *StoriesPane) Init() tea.Cmd { return nil }
func (p *AgentsPane) Init() tea.Cmd  { return nil }
func (l *LogsPane) Init() tea.Cmd    { return nil }
func (l *LogsPane) Tick() tea.Cmd    { return nil } // TODO: tail .claude/logs/skills.jsonl
	// TODO: read docs/stories/*.md and gh issue list
	placeholder := []string{
		"● 0042  export-csv       ▸ Implement",
		"○ 0041  auth-reset       ✔ Done",
		"◐ 0040  billing-page     ▸ QA",
		"◯ 0039  settings-ui      ▸ Ready",
		"◌ spike-perf-audit",
	}
	return strings.Join(placeholder[:min(h, len(placeholder))], "\n")
}

func (p *AgentsPane) View(w, h int) string {
	placeholder := []string{
		"orchestrator      idle            ···",
		"wireframe-arch    wireframe.md     ⏺",
		"ui-mockup-build   Button.tsx       ⏺",
		"qa-cucumber       waiting          ⏸",
	}
	return strings.Join(placeholder[:min(h, len(placeholder))], "\n")
}

func (p *PRsPane) View(w, h int) string {
	placeholder := []string{
		"#128  0042 export-csv   ◷ Review",
		"#127  0040 billing-ui   ✘ Failed",
		"#126  0039 settings     ✔ Green",
	}
	return strings.Join(placeholder[:min(h, len(placeholder))], "\n")
}

func (p *QAPane) View(w, h int) string {
	placeholder := []string{
		"✔ 23 scenarios passing",
		"✘  2 scenarios failing (0040)",
		"  → Scenario: invalid card number",
		"  → Scenario: expired card",
	}
	return strings.Join(placeholder[:min(h, len(placeholder))], "\n")
}

func (p *DesignPane) View(w, h int) string {
	// TODO: render wireframe ASCII, palette swatches, token tree
	return "Design artifacts\n(wireframes · mockups · tokens)\nPress Enter on a story to view wireframe"
}

func (l *LogsPane) View(w, h int) string {
	if len(l.lines) == 0 {
		return lipgloss.NewStyle().Faint(true).Render("Tailing .claude/logs/skills.jsonl…")
	}
	start := len(l.lines) - h
	if start < 0 {
		start = 0
	}
	return strings.Join(l.lines[start:], "\n")
}

func (p *SuperpowerPane) View(w, h int) string {
	// TODO: discover .claude/superpowers/*.yaml and render fuzzy list
	superpowers := []string{
		"  new-ui-feature    Full-stack UI with design + a11y",
		"  api-endpoint      API spec + Gherkin + pipeline",
		"  bugfix            Triage → red test → fix → green",
		"  design-refresh    Refresh palette + tokens + mockups",
	}
	header := "Fire Superpower  [Enter] to launch\n\n"
	return header + strings.Join(superpowers[:min(h-3, len(superpowers))], "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


