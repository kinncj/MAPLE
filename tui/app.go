package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//Pane identifiers
type pane int

const (
	paneStories pane = iota
	paneAgents
	panePRs
	paneQA
	paneDesign
	paneLogs
	paneHelp
	paneSuperpower
	paneCount
)

var paneNames = map[pane]string{
	paneStories:    "Stories",
	paneAgents:     "Active Agents",
	panePRs:        "PRs",
	paneQA:         "QA / Gherkin",
	paneDesign:     "Design",
	paneLogs:       "Logs",
	paneHelp:       "Help",
	paneSuperpower: "Superpowers",
}

// App is the root Bubble Tea model.
type App struct {
	theme       Theme
	activePane  pane
	showHelp    bool
	showSplash  bool
	splashDone  bool
	noAnimate   bool
	splashTick  int
	shimmerPos  int
	width       int
	height      int
	cmdMode     bool
	cmdBuffer   string
	searchMode  bool
	searchQuery string

	stories    StoriesPane
	agents     AgentsPane
	prs        PRsPane
	qa         QAPane
	design     DesignPane
	logs       LogsPane
	superpower SuperpowerPane
}

type tickMsg time.Time
type shimmerTickMsg struct{}

func newApp() *App {
	return &App{
		theme:      loadTheme(),
		activePane: paneStories,
		showSplash: true,
		shimmerPos: -1,
	}
}

func (a *App) Init() tea.Cmd {
	if a.noAnimate {
		a.showSplash = false
	}
	return tea.Batch(
		tickCmd(),
		a.stories.Init(),
		a.logs.Init(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func shimmerCmd() tea.Cmd {
	return tea.Tick(10*time.Second, func(_ time.Time) tea.Msg {
		return shimmerTickMsg{}
	})
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.splashDone = true // skip splash on resize events after init
		return a, nil

	case tickMsg:
		if a.showSplash {
			a.splashTick++
			// Splash shows for ~800ms (1 tick), then transitions
			if a.splashTick >= 1 || a.noAnimate {
				a.showSplash = false
				return a, shimmerCmd()
			}
		}
		var cmds []tea.Cmd
		cmds = append(cmds, tickCmd())
		cmds = append(cmds, a.logs.Tick())
		return a, tea.Batch(cmds...)

	case shimmerTickMsg:
		// Advance shimmer highlight across logo
		a.shimmerPos = (a.shimmerPos + 1) % logoWidth
		return a, shimmerCmd()

	case tea.KeyMsg:
		return a.handleKey(msg)
	}

	return a, nil
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Command mode
	if a.cmdMode {
		return a.handleCmdMode(msg)
	}
	// Search mode
	if a.searchMode {
		return a.handleSearchMode(msg)
	}

	switch msg.String() {
	case "ctrl+c":
		return a, tea.Quit
	case "?":
		a.showHelp = !a.showHelp
	case "tab":
		a.activePane = (a.activePane + 1) % paneCount
	case "shift+tab":
		a.activePane = (a.activePane - 1 + paneCount) % paneCount
	case "s":
		a.activePane = paneStories
	case "a":
		a.activePane = paneAgents
	case "p":
		a.activePane = panePRs
	case "q":
		a.activePane = paneQA
	case "d":
		a.activePane = paneDesign
	case "l":
		a.activePane = paneLogs
	case "F":
		a.activePane = paneSuperpower
	case "/":
		a.searchMode = true
		a.searchQuery = ""
	case ":":
		a.cmdMode = true
		a.cmdBuffer = ""
	case "r":
		return a, a.refresh()
	case "j":
		a.moveFocus(1)
	case "k":
		a.moveFocus(-1)
	case "g":
		a.moveFocus(-9999)
	case "G":
		a.moveFocus(9999)
	}
	return a, nil
}

func (a *App) handleCmdMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		cmd := a.cmdBuffer
		a.cmdMode = false
		a.cmdBuffer = ""
		return a, a.executeCommand(cmd)
	case "esc":
		a.cmdMode = false
		a.cmdBuffer = ""
	case "backspace":
		if len(a.cmdBuffer) > 0 {
			a.cmdBuffer = a.cmdBuffer[:len(a.cmdBuffer)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			a.cmdBuffer += string(msg.Runes)
		}
	}
	return a, nil
}

func (a *App) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		a.searchMode = false
	case "backspace":
		if len(a.searchQuery) > 0 {
			a.searchQuery = a.searchQuery[:len(a.searchQuery)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			a.searchQuery += string(msg.Runes)
		}
	}
	return a, nil
}

func (a *App) executeCommand(cmd string) tea.Cmd {
	// Parse :theme <name> immediately; others TODO
	if len(cmd) > 6 && cmd[:6] == "theme " {
		a.theme = themeByName(cmd[6:])
	}
	// TODO: :kickoff, :sync, :a11y, :resume
	return nil
}

func (a *App) refresh() tea.Cmd {
	return nil // TODO: reload data for active pane
}

func (a *App) moveFocus(delta int) {
	// Delegate to active pane
}

const minWidth = 80

func (a *App) View() string {
	if a.showSplash {
		return renderSplash(a.theme, a.width, a.height, a.shimmerPos, a.noAnimate)
	}
	if a.width == 0 {
		return "Loading…"
	}
	// Narrow terminal: degrade to single-column scrolling log
	if a.width < minWidth {
		return renderNarrow(a.theme, a.width, a.height, a.getPaneContent(a.activePane, a.width-2, a.height-4))
	}
	if a.showHelp {
		return renderHelp(a.theme, a.width, a.height)
	}
	return a.renderDashboard()
}

func (a *App) renderDashboard() string {
	half := a.width / 2
	topH := (a.height - 4) / 2
	botH := a.height - 4 - topH

	topLeft  := a.renderPane(paneStories, half-1, topH)
	topRight := a.renderPane(paneAgents,  a.width-half-1, topH)
	botLeft  := a.renderPane(panePRs,     half-1, botH)
	botRight := a.renderPane(paneQA,      a.width-half-1, botH)

	top := lipgloss.JoinHorizontal(lipgloss.Top, topLeft, topRight)
	bot := lipgloss.JoinHorizontal(lipgloss.Top, botLeft, botRight)

	statusBar := a.renderStatusBar()
	footer    := a.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left,
		a.theme.Border.Render(fmt.Sprintf(" squad ─── %s ", a.projectName())),
		top,
		bot,
		statusBar,
		footer,
	)
}

func (a *App) renderPane(p pane, w, h int) string {
	style := a.theme.Pane
	if p == a.activePane {
		style = a.theme.ActivePane
	}
	title := paneNames[p]
	content := a.getPaneContent(p, w-2, h-2)
	return style.Width(w).Height(h).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			a.theme.PaneTitle.Render(title),
			content,
		),
	)
}

func (a *App) getPaneContent(p pane, w, h int) string {
	switch p {
	case paneStories:
		return a.stories.View(w, h)
	case paneAgents:
		return a.agents.View(w, h)
	case panePRs:
		return a.prs.View(w, h)
	case paneQA:
		return a.qa.View(w, h)
	case paneDesign:
		return a.design.View(w, h)
	case paneLogs:
		return a.logs.View(w, h)
	case paneSuperpower:
		return a.superpower.View(w, h)
	}
	return ""
}

func (a *App) renderStatusBar() string {
	if a.cmdMode {
		return a.theme.StatusBar.Render(fmt.Sprintf(":%s█", a.cmdBuffer))
	}
	if a.searchMode {
		return a.theme.StatusBar.Render(fmt.Sprintf("/%s█", a.searchQuery))
	}
	return a.theme.StatusBar.Render(
		fmt.Sprintf("  [Tab] cycle · [s]tories · [a]gents · [d]esign · [l]ogs · [F] superpower · [?] help"),
	)
}

func (a *App) renderFooter() string {
	return a.theme.Footer.Width(a.width).Render(
		fmt.Sprintf(" %s ", paneNames[a.activePane]),
	)
}

func (a *App) projectName() string {
	// TODO: read from project.config.yaml
	return "project"
}
