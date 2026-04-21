package main

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Result ───────────────────────────────────────────────────────────────────

type menuAction int

const (
	menuNone menuAction = iota
	menuInit
	menuUpdate
	menuReq
	menuLabels
	menuProject
	menuHelp
	menuQuit
)

type menuResult struct {
	action menuAction
}

// ─── Model ────────────────────────────────────────────────────────────────────

type menuItem struct {
	action   menuAction
	label    string
	desc     string
	disabled bool
	why      string
}

type menuModel struct {
	items       []menuItem
	cursor      int
	result      menuResult
	done        bool
	showHelp    bool
	logoFrame   int
	logoDone    bool
	initialized bool
	cwd         string
	tools       Tools
	width       int
}

func runMenu(tools Tools, fsys fs.FS) menuResult {
	if !isStdinTTY() {
		printHelpStatic()
		return menuResult{action: menuQuit}
	}

	_, err := os.Stat("project.config.yaml")
	initialized := err == nil

	cwd, _ := os.Getwd()
	m := &menuModel{
		items:       buildMenuItems(tools, initialized),
		initialized: initialized,
		tools:       tools,
		cwd:         cwd,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	final, runErr := p.Run()
	if runErr != nil {
		return menuResult{action: menuQuit}
	}
	return final.(*menuModel).result
}

func buildMenuItems(tools Tools, initialized bool) []menuItem {
	items := []menuItem{
		{action: menuInit, label: "Init", desc: "Set up MAPLE in this directory"},
	}

	if initialized {
		items = append(items, menuItem{
			action: menuUpdate,
			label:  "Update",
			desc:   "Re-sync agents, skills, and hooks with latest templates",
		})
	}

	req := menuItem{action: menuReq, label: "Requirements", desc: "Write requirements → Gherkin story"}
	if !tools.HasReqAI() {
		req.disabled = true
		req.why = "needs claude or opencode"
	}
	items = append(items, req)

	labels := menuItem{action: menuLabels, label: "Labels", desc: "Bootstrap GitHub label set in current repo"}
	if tools.GH == "" {
		labels.disabled = true
		labels.why = "needs gh CLI"
	}
	items = append(items, labels)

	proj := menuItem{action: menuProject, label: "Project", desc: "Create GitHub Project v2"}
	if tools.GH == "" {
		proj.disabled = true
		proj.why = "needs gh CLI"
	}
	items = append(items, proj)

	items = append(items, menuItem{action: menuHelp, label: "Help", desc: "Show documentation"})
	return items
}

// ─── Bubble Tea lifecycle ─────────────────────────────────────────────────────

func (m *menuModel) Init() tea.Cmd {
	return logoTick(0)
}

func (m *menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case logoTickMsg:
		if !m.logoDone {
			m.logoFrame++
			if m.logoFrame >= logoFrameCount {
				m.logoDone = true
			} else {
				return m, logoTick(m.logoFrame)
			}
		}

	case tea.KeyMsg:
		// Suppress all input until logo is done
		if !m.logoDone {
			return m, nil
		}

		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.result = menuResult{action: menuQuit}
			m.done = true
			return m, tea.Quit

		case "up", "k":
			m.moveCursor(-1)

		case "down", "j":
			m.moveCursor(1)

		case "enter", " ":
			item := m.items[m.cursor]
			if item.disabled {
				return m, nil
			}
			if item.action == menuHelp {
				m.showHelp = true
				return m, nil
			}
			m.result = menuResult{action: item.action}
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *menuModel) moveCursor(dir int) {
	n := len(m.items)
	for i := 0; i < n; i++ {
		m.cursor = (m.cursor + dir + n) % n
		if !m.items[m.cursor].disabled {
			return
		}
	}
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m *menuModel) View() string {
	t := tokyoNight()

	var hdr string
	if m.logoDone {
		hdr = logo()
	} else {
		hdr = logoAnimFrame(m.logoFrame)
	}

	if m.showHelp {
		return hdr + m.helpView(t)
	}

	return hdr + m.menuView(t)
}

func (m *menuModel) menuView(t Theme) string {
	var sb strings.Builder

	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("MAPLE")
	ver := lipgloss.NewStyle().Foreground(t.Muted).Render(" · " + version)
	sb.WriteString("  " + title + ver + "\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(t.Muted).Render("  " + strings.Repeat("─", 54)) + "\n\n")

	cursor := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("❯")
	space := "  "

	for i, item := range m.items {
		var line string
		if item.disabled {
			label := lipgloss.NewStyle().Foreground(t.Muted).Render(fmt.Sprintf("%-14s", item.label))
			why := lipgloss.NewStyle().Foreground(t.Muted).Italic(true).Render(item.why)
			line = "    " + label + "  " + why
		} else if i == m.cursor {
			label := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render(fmt.Sprintf("%-14s", item.label))
			desc := lipgloss.NewStyle().Foreground(t.Foreground).Render(item.desc)
			line = "  " + cursor + " " + label + "  " + desc
		} else {
			label := lipgloss.NewStyle().Foreground(t.Foreground).Render(fmt.Sprintf("%-14s", item.label))
			desc := lipgloss.NewStyle().Foreground(t.Muted).Render(item.desc)
			line = space + "  " + label + "  " + desc
		}
		sb.WriteString(line + "\n")
	}

	sb.WriteString("\n" + lipgloss.NewStyle().Foreground(t.Muted).Render("  "+strings.Repeat("─", 54)) + "\n")

	keys := lipgloss.NewStyle().Foreground(t.Muted).Render(
		"  ↑/↓ j/k Navigate   Enter Select   q Quit")
	sb.WriteString(keys + "\n\n")

	// Status line
	cwdStr := m.cwd
	if len(cwdStr) > 50 {
		cwdStr = "…" + cwdStr[len(cwdStr)-49:]
	}
	var initStatus string
	if m.initialized {
		initStatus = lipgloss.NewStyle().Foreground(t.Success).Render("● Initialized")
	} else {
		initStatus = lipgloss.NewStyle().Foreground(t.Muted).Render("○ Not initialized")
	}

	summaryLines := m.tools.Summary()
	toolLine := strings.Join(summaryLines, "  ")
	toolStr := lipgloss.NewStyle().Foreground(t.Muted).Render(toolLine)

	sb.WriteString("  " + lipgloss.NewStyle().Foreground(t.Muted).Render(cwdStr) + "  " + initStatus + "\n")
	sb.WriteString("  " + toolStr + "\n")

	return sb.String()
}

func (m *menuModel) helpView(t Theme) string {
	var sb strings.Builder
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("Documentation")
	sb.WriteString("  " + title + "\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(t.Muted).Render("  "+strings.Repeat("─", 54)) + "\n\n")

	sections := [][2]string{
		{"Init", "Copies agents, skills, hooks, and config into the current\n                  directory for each detected AI tool."},
		{"Update", "Re-syncs managed files (agents, skills, hooks) with the\n                  latest templates. Never overwrites project.config.yaml."},
		{"Requirements", "Interactive editor. Type plain-text requirements, press\n                  Ctrl+D to convert via detected AI tool → Gherkin story\n                  saved to docs/stories/."},
		{"Labels", "Creates the canonical MAPLE GitHub label set in the\n                  current repo using gh CLI."},
		{"Project", "Creates a GitHub Project v2 and writes the project number\n                  and node ID into project.config.yaml."},
	}

	for _, s := range sections {
		label := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render(fmt.Sprintf("  %-16s", s[0]))
		desc := lipgloss.NewStyle().Foreground(t.Foreground).Render(s[1])
		sb.WriteString(label + desc + "\n\n")
	}

	sb.WriteString(lipgloss.NewStyle().Foreground(t.Muted).Render("  Press any key to return.") + "\n")
	return sb.String()
}
