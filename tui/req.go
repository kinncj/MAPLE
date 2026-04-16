package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Requirements command ─────────────────────────────────────────────────────

func runReq(tools Tools) error {
	if !tools.HasAI() {
		return fmt.Errorf("no AI tool found (need claude, opencode, or gh copilot)")
	}
	m := newReqModel(tools)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// ─── Model ────────────────────────────────────────────────────────────────────

type reqStep int

const (
	reqStepEdit reqStep = iota
	reqStepSending
	reqStepDone
)

type reqModel struct {
	tools    Tools
	textarea textarea.Model
	step     reqStep
	result   string
	savedTo  string
	err      error
	width    int
	height   int
}

type reqDoneMsg struct {
	gherkin string
	savedTo string
	err     error
}

func newReqModel(tools Tools) *reqModel {
	ta := textarea.New()
	ta.Placeholder = "Describe what you need...\n\nExample:\n  As a user I want to export filtered results as CSV\n  so that I can share them with stakeholders offline.\n\n  The export must respect active filters.\n  Only the last 90 days of data should be included.\n  The file must be downloadable immediately."
	ta.Focus()
	ta.SetWidth(80)
	ta.SetHeight(20)
	ta.ShowLineNumbers = false
	return &reqModel{
		tools:    tools,
		textarea: ta,
		step:     reqStepEdit,
	}
}

func (m *reqModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m *reqModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(msg.Height - 10)

	case tea.KeyMsg:
		switch m.step {
		case reqStepEdit:
			switch msg.Type {
			case tea.KeyCtrlD:
				req := strings.TrimSpace(m.textarea.Value())
				if req == "" {
					return m, nil
				}
				m.step = reqStepSending
				return m, m.convert(req)
			case tea.KeyCtrlC:
				return m, tea.Quit
			}
		case reqStepDone:
			return m, tea.Quit
		}

	case reqDoneMsg:
		m.step = reqStepDone
		m.result = msg.gherkin
		m.savedTo = msg.savedTo
		m.err = msg.err
	}

	if m.step == reqStepEdit {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *reqModel) View() string {
	t := tokyoNight()
	header := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render(logo())
	ai := lipgloss.NewStyle().Foreground(t.Muted).Render("  AI: " + m.tools.PreferredAI())

	switch m.step {
	case reqStepEdit:
		help := lipgloss.NewStyle().Foreground(t.Muted).Render(
			"Type your requirements below. Ctrl+D to convert, Ctrl+C to quit.")
		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary).
			Padding(0, 1).
			Render(m.textarea.View())
		return header + "  " + ai + "\n\n" + help + "\n\n" + box + "\n"

	case reqStepSending:
		return header + "  " + ai + "\n\n" +
			lipgloss.NewStyle().Foreground(t.Accent).Render("  Converting to Gherkin...") + "\n"

	case reqStepDone:
		if m.err != nil {
			return header + "\n\n" +
				lipgloss.NewStyle().Foreground(t.Error).Render("  ✗ "+m.err.Error()) + "\n" +
				lipgloss.NewStyle().Foreground(t.Muted).Render("  Press any key to exit.") + "\n"
		}
		saved := lipgloss.NewStyle().Foreground(t.Success).Bold(true).Render("  ✓ Saved: " + m.savedTo)
		preview := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Muted).
			Padding(0, 1).
			Render(m.result)
		return header + "\n\n" + saved + "\n\n" + preview + "\n" +
			lipgloss.NewStyle().Foreground(t.Muted).Render("  Press any key to exit.") + "\n"
	}
	return ""
}

func (m *reqModel) convert(requirements string) tea.Cmd {
	tools := m.tools
	return func() tea.Msg {
		gherkin, err := convertToGherkin(requirements, tools)
		if err != nil {
			return reqDoneMsg{err: err}
		}
		savedTo, err := saveStory(requirements, gherkin)
		return reqDoneMsg{gherkin: gherkin, savedTo: savedTo, err: err}
	}
}

// ─── AI conversion ────────────────────────────────────────────────────────────

const gherkinPrompt = `You are a Gherkin authoring expert. Convert the following requirements into a Gherkin Feature file.

Rules:
- Use Feature, Background (if needed), Scenario, Given/When/Then/And/But
- Write in plain language a non-technical stakeholder can understand
- One scenario per distinct case (happy path + key edge cases)
- Add @tags: @story, @priority:medium (adjust if clear from requirements)
- Do NOT include step definitions — only the .feature file content
- Output ONLY the Gherkin block, no explanation

Requirements:
%s`

func convertToGherkin(requirements string, tools Tools) (string, error) {
	prompt := fmt.Sprintf(gherkinPrompt, requirements)

	var cmd *exec.Cmd
	switch {
	case tools.Claude != "":
		cmd = exec.Command(tools.Claude, "--print", prompt)
	case tools.OpenCode != "":
		cmd = exec.Command(tools.OpenCode, "run", prompt)
	case tools.GHCopilot:
		cmd = exec.Command(tools.GH, "copilot", "explain", prompt)
	default:
		return "", fmt.Errorf("no AI tool available")
	}

	out, err := cmd.Output()
	if err != nil {
		// Fallback: try piping via stdin for claude
		if tools.Claude != "" {
			cmd2 := exec.Command(tools.Claude, "--print")
			cmd2.Stdin = strings.NewReader(prompt)
			out, err = cmd2.Output()
		}
		if err != nil {
			return "", fmt.Errorf("%s: %w", tools.PreferredAI(), err)
		}
	}

	result := strings.TrimSpace(string(out))
	// Strip markdown code fences if the AI wrapped the output
	result = stripFences(result)
	return result, nil
}

func stripFences(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if i == 0 && (trimmed == "```gherkin" || trimmed == "```") {
			continue
		}
		if i == len(lines)-1 && trimmed == "```" {
			continue
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}

// ─── Story file writer ────────────────────────────────────────────────────────

func saveStory(requirements, gherkin string) (string, error) {
	if err := os.MkdirAll("docs/stories", 0755); err != nil {
		return "", err
	}

	slug := slugify(firstLine(requirements))
	ts := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("docs/stories/%s-%s-0001.md", slug, ts)

	// Derive a title from first line of requirements
	title := firstLine(requirements)
	if len(title) > 80 {
		title = title[:80]
	}

	content := fmt.Sprintf(`---
id: "%s-0001"
title: "%s"
epic: "%s"
priority: "medium"
ui: false
adr_required: false
milestone: null
labels:
  - "type:feature"
  - "priority:medium"
  - "phase:discover"
issue_number: null
issue_url: null
created_at: "%s"
---

# %s

## Narrative

%s

## Scenarios

`+"```gherkin\n%s\n```"+`

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber scenarios green
- [ ] ADRs linked where required

## ADR Links

(populated by adr-author agent)
`,
		slug, title, slug, time.Now().Format(time.RFC3339),
		title, requirements, gherkin,
	)

	return filepath.Clean(filename), os.WriteFile(filename, []byte(content), 0644)
}

func firstLine(s string) string {
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			return l
		}
	}
	return "untitled"
}

func slugify(s string) string {
	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	// Trim to reasonable length, break at word boundary
	runes := []rune(s)
	if len(runes) > 40 {
		runes = runes[:40]
		for len(runes) > 0 && !unicode.IsLetter(runes[len(runes)-1]) && !unicode.IsDigit(runes[len(runes)-1]) {
			runes = runes[:len(runes)-1]
		}
	}
	return string(runes)
}
