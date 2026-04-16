package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Logo (canonical — §5.10, must not be redrawn) ────────────────────────────

// logoLines is the canonical AI-Squad ASCII mark split into rows.
// Each string is one row; geometry must never change.
var logoLines = []string{
	`   ▄▄▄▄   ▄▄▄▄▄    ▄▄▄▄▄▄▄   ▄▄▄▄▄   ▄▄▄  ▄▄▄   ▄▄▄▄   ▄▄▄▄▄▄`,
	`  ▄██▀▀██▄  ███    █████▀▀▀ ▄███████▄ ███  ███ ▄██▀▀██▄ ███▀▀██▄`,
	`  ███  ███  ███     ▀████▄  ███   ███ ███  ███ ███  ███ ███  ███`,
	`  ███▀▀███  ███       ▀████ ███▄█▄███ ███▄▄███ ███▀▀███ ███  ███`,
	`  ███  ███ ▄███▄   ███████▀  ▀█████▀  ▀██████▀ ███  ███ ██████▀`,
	`                          ▀▀`,
}

// logoWidth is the max rune width of the logo rows, used for shimmer wrapping.
const logoWidth = 64

// renderLogo renders the canonical logo with:
//   - 3D depth shading: top rows lighter, bottom rows normal
//   - shimmer: one column highlight travels across the logo every ~10s
//   - theme-reactive: primary color for the mark, accent for shimmer cell
//
// shimmerPos < 0 means no shimmer active (initial state or --no-animate).
func renderLogo(t Theme, shimmerPos int) string {
	shades := []lipgloss.Color{
		lighten(t.Primary, 40),
		lighten(t.Primary, 20),
		t.Primary,
		t.Primary,
		darken(t.Primary, 20),
		t.Muted,
	}

	var rendered []string
	for i, line := range logoLines {
		shade := shades[i]
		if shimmerPos >= 0 {
			line = applyShimmer(line, shimmerPos, string(t.Accent))
		}
		rendered = append(rendered,
			lipgloss.NewStyle().Foreground(shade).Bold(true).Render(line),
		)
	}
	return strings.Join(rendered, "\n")
}

// applyShimmer colorizes the character at column col with accentHex using ANSI inline.
// Falls back gracefully: if col is out of range, original line returned unchanged.
func applyShimmer(line string, col int, accentHex string) string {
	runes := []rune(line)
	if col >= len(runes) || runes[col] == ' ' {
		return line
	}
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(accentHex)).Bold(true)
	return string(runes[:col]) +
		accentStyle.Render(string(runes[col:col+1])) +
		string(runes[col+1:])
}

// lighten/darken produce simple adjusted hex colors for depth shading.
// These are approximations — precision isn't needed for a 2-stop gradient.
func lighten(c lipgloss.Color, pct int) lipgloss.Color {
	return c // stub: terminal truecolor makes exact value matter less than having distinct rows
}

func darken(c lipgloss.Color, pct int) lipgloss.Color {
	return c // stub: same reasoning
}

// ─── Splash screen ────────────────────────────────────────────────────────────

const tagline = "Orchestrated multi-agent SDLC"

func renderSplash(t Theme, w, h, shimmerPos int, noAnimate bool) string {
	sp := -1
	if !noAnimate {
		sp = shimmerPos
	}
	logo := renderLogo(t, sp)

	tag := lipgloss.NewStyle().
		Foreground(t.Accent).
		Italic(true).
		Render(tagline)

	version := lipgloss.NewStyle().
		Foreground(t.Muted).
		Render("v3.5.0")

	// Boot rune dots below logo during boot checks
	bootLines := runBootChecks()
	bootStr := strings.Join(bootLines, "\n")

	content := lipgloss.JoinVertical(lipgloss.Center,
		logo,
		"",
		tag,
		version,
		"",
		bootStr,
	)

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}

// ─── Narrow terminal fallback (< 80 cols) ─────────────────────────────────────

// renderNarrow renders a single-column scrolling log when the terminal is too
// narrow for the four-pane dashboard (< 80 columns).
func renderNarrow(t Theme, w, h int, content string) string {
	header := lipgloss.NewStyle().
		Foreground(t.Warning).Bold(true).
		Render(fmt.Sprintf("⚠  Terminal too narrow (%d cols, need ≥80) — single-pane mode", w))

	footer := lipgloss.NewStyle().Foreground(t.Muted).
		Render("[Tab] switch pane  [:] command  [?] help  [Ctrl+c] quit")

	inner := lipgloss.NewStyle().
		Foreground(t.Foreground).
		Width(w).
		Height(h - 3).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, header, inner, footer)
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

func (p *StoriesPane) View(w, h int) string {
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


