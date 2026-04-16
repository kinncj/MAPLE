package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Init command ─────────────────────────────────────────────────────────────

func runInit(tools Tools, templateDir string, force bool) error {
	m := newInitModel(tools, templateDir, force, false)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func runInitFromMenu(tools Tools, templateDir string, force bool) error {
	m := newInitModel(tools, templateDir, force, true)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// ─── Wizard model ─────────────────────────────────────────────────────────────

type initStep int

const (
	stepWelcome initStep = iota
	stepConfirm
	stepRun
	stepDone
)

type initModel struct {
	tools       Tools
	templateDir string
	force       bool
	fromMenu    bool // skip welcome step; jump straight to confirm
	step        initStep
	spinner     spinner.Model
	logs        []string
	err         error
	width       int
	logoFrame   int
	logoDone    bool
}

type initDoneMsg struct{ logs []string; err error }

func newInitModel(tools Tools, templateDir string, force bool, fromMenu bool) *initModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))
	initialStep := stepWelcome
	if fromMenu {
		initialStep = stepConfirm
	}
	return &initModel{
		tools:       tools,
		templateDir: templateDir,
		force:       force,
		fromMenu:    fromMenu,
		step:        initialStep,
		spinner:     s,
	}
}

func (m *initModel) Init() tea.Cmd {
	return logoTick(0)
}

func (m *initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	// Suppress Enter until logo is done
	case tea.KeyMsg:
		if !m.logoDone {
			return m, nil
		}
		switch m.step {
		case stepWelcome:
			if msg.String() == "enter" || msg.String() == "y" {
				m.step = stepConfirm
			} else if msg.String() == "ctrl+c" || msg.String() == "q" {
				return m, tea.Quit
			}
		case stepConfirm:
			switch msg.String() {
			case "enter", "y":
				m.step = stepRun
				return m, tea.Batch(m.spinner.Tick, m.runSetup())
			case "ctrl+c", "q", "n":
				return m, tea.Quit
			}
		case stepDone:
			return m, tea.Quit
		}

	case spinner.TickMsg:
		if m.step == stepRun {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case initDoneMsg:
		m.step = stepDone
		m.logs = msg.logs
		m.err = msg.err
	}
	return m, nil
}

func (m *initModel) View() string {
	t := tokyoNight()
	var header string
	if m.logoDone {
		header = logo()
	} else {
		header = logoAnimFrame(m.logoFrame)
	}
	switch m.step {
	case stepWelcome:
		detected := strings.Join(m.tools.Summary(), "\n")
		body := lipgloss.NewStyle().Foreground(t.Foreground).Render(
			"\nDetected tools:\n" + detected + "\n\n" +
				"This will copy agents, skills, and hooks into the current directory.\n")
		prompt := lipgloss.NewStyle().Foreground(t.Accent).Render("Press Enter to continue, q to quit.")
		return header + body + "\n" + prompt + "\n"

	case stepConfirm:
		cwd, _ := os.Getwd()
		body := lipgloss.NewStyle().Foreground(t.Foreground).Render(
			"\nTarget: " + cwd + "\n\n" +
				"Will set up:\n" +
				m.platformList() + "\n")
		prompt := lipgloss.NewStyle().Foreground(t.Warning).Render("Press Enter to confirm, n to cancel.")
		return header + body + "\n" + prompt + "\n"

	case stepRun:
		return header + "\n" + m.spinner.View() + "  Setting up AI-Squad...\n"

	case stepDone:
		if m.err != nil {
			msg := lipgloss.NewStyle().Foreground(t.Error).Render("\n✗ Setup failed: " + m.err.Error())
			return header + msg + "\n"
		}
		var sb strings.Builder
		sb.WriteString(lipgloss.NewStyle().Foreground(t.Success).Bold(true).Render("\n✓ AI-Squad initialized\n\n"))
		for _, l := range m.logs {
			sb.WriteString("  " + l + "\n")
		}
		sb.WriteString("\n")
		sb.WriteString(lipgloss.NewStyle().Foreground(t.Muted).Render(
			"Next steps:\n" +
				"  • Open your project in Claude Code, OpenCode, or Copilot CLI\n" +
				"  • Run: squad req  — to write requirements and generate a story\n" +
				"  • Run: squad labels  — to bootstrap GitHub labels\n"))
		return header + sb.String() + "\n"
	}
	return ""
}

func (m *initModel) platformList() string {
	var items []string
	if m.tools.Claude != "" {
		items = append(items, "  • .claude/ agents + skills + hooks")
	}
	if m.tools.OpenCode != "" {
		items = append(items, "  • .opencode/ agents + skills")
	}
	if m.tools.GHCopilot {
		items = append(items, "  • .github/copilot-instructions.md")
	}
	items = append(items, "  • docs/ story templates, specs, design structure")
	items = append(items, "  • Makefile, lefthook.yml, scripts/sdlc/")
	if m.tools.Lefthook != "" {
		items = append(items, "  • lefthook install (hooks wired)")
	}
	return strings.Join(items, "\n")
}

func (m *initModel) runSetup() tea.Cmd {
	force := m.force
	return func() tea.Msg {
		logs, err := doInit(m.tools, m.templateDir, force)
		return initDoneMsg{logs: logs, err: err}
	}
}

// ─── Actual setup logic ───────────────────────────────────────────────────────

func doInit(tools Tools, templateDir string, force bool) ([]string, error) {
	var logs []string
	log := func(s string) { logs = append(logs, s) }

	cwd, err := os.Getwd()
	if err != nil {
		return logs, err
	}

	// Always copy: docs structure, Makefile, scripts, .github
	pairs := []struct{ src, dst string }{
		{"docs",    "docs"},
		{"scripts", "scripts"},
		{"Makefile", "Makefile"},
		{"lefthook.yml", "lefthook.yml"},
		{".github",  ".github"},
	}

	// Platform-specific copies
	if tools.Claude != "" {
		pairs = append(pairs,
			struct{ src, dst string }{".claude", ".claude"},
			struct{ src, dst string }{"CLAUDE.md", "CLAUDE.md"},
		)
	}
	if tools.OpenCode != "" {
		pairs = append(pairs,
			struct{ src, dst string }{".opencode", ".opencode"},
			struct{ src, dst string }{"AGENTS.md", "AGENTS.md"},
			struct{ src, dst string }{"opencode.json", "opencode.json"},
		)
	}

	for _, p := range pairs {
		src := filepath.Join(templateDir, p.src)
		dst := filepath.Join(cwd, p.dst)
		if err := copyPath(src, dst, force); err != nil {
			log("✗ " + p.dst + ": " + err.Error())
		} else {
			if force {
				log("↺ " + p.dst + " (updated)")
			} else {
				log("✓ " + p.dst)
			}
		}
	}

	// Write project.config.yaml if not present
	cfgPath := filepath.Join(cwd, "project.config.yaml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		projectName := filepath.Base(cwd)
		cfg := projectConfigYAML(projectName)
		if err := os.WriteFile(cfgPath, []byte(cfg), 0644); err != nil {
			log("✗ project.config.yaml: " + err.Error())
		} else {
			log("✓ project.config.yaml")
		}
	} else {
		log("~ project.config.yaml (already exists, skipped)")
	}

	// Wire lefthook
	if tools.Lefthook != "" {
		if out, err := exec.Command(tools.Lefthook, "install").CombinedOutput(); err != nil {
			log("✗ lefthook install: " + strings.TrimSpace(string(out)))
		} else {
			log("✓ lefthook hooks wired")
		}
	}

	return logs, nil
}

func copyPath(src, dst string, force bool) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dst, force)
	}
	return copyFile(src, dst, force)
}

func copyDir(src, dst string, force bool) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target, force)
	})
}

func copyFile(src, dst string, force bool) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	if !force {
		if _, err := os.Stat(dst); err == nil {
			return nil // skip existing
		}
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func projectConfigYAML(name string) string {
	return fmt.Sprintf(`project:
  name: "%s"
  created_at: "%s"

sdlc:
  mode: standard          # standard | spike | quick
  require_adr_for:
    - new_dependency
    - cross_boundary_change
    - data_model_change
    - auth_change
    - mcp_adoption
    - visual_identity_change

qa:
  bdd: cucumber           # cucumber | behave | pytest-bdd
  coverage_threshold: 80

design:
  ui_library: null        # mantine | tailwind | shadcn | null
  token_format: dtcg      # W3C DTCG

github:
  project_number: null
  project_node_id: null
`, name, time.Now().Format(time.RFC3339))
}
