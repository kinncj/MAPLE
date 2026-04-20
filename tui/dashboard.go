package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Data types ───────────────────────────────────────────────────────────────

type storyRow struct {
	id       string
	slug     string
	priority string
	phase    string
	ui       bool
	issue    int
}

type prRow struct {
	number int
	title  string
	state  string // OPEN / CLOSED / MERGED
}

type agentRow struct {
	agent string
	op    string
	file  string
	ts    string
}

type spRow struct {
	name string
	desc string
	path string
}

// ─── Dashboard exit actions ───────────────────────────────────────────────────

type dashAction int

const (
	dashActionNone dashAction = iota
	dashActionReq             // quit dashboard and run req (Gherkin converter)
)

// ─── Pane IDs ─────────────────────────────────────────────────────────────────

type dashPane int

const (
	paneStories dashPane = iota
	paneAgents
	panePRs
	paneQA
	paneCount = 4 // panes in the 2×2 grid
	// full-screen overlays
	paneDesign
	paneLogs
)

// ─── Async messages ───────────────────────────────────────────────────────────

type prsLoadedMsg struct {
	items []prRow
	err   string
}

type shimmerTickMsg struct{}
type dashRefreshMsg struct{}

func shimmerTick() tea.Cmd {
	return tea.Tick(10*time.Second, func(time.Time) tea.Msg { return shimmerTickMsg{} })
}

// ─── Model ────────────────────────────────────────────────────────────────────

type dashboardModel struct {
	theme       Theme
	noAnimate   bool
	width       int
	height      int
	projectName string

	focus      dashPane
	fullscreen dashPane // paneDesign or paneLogs, -1 = none
	showHelp   bool
	showSP     bool
	spFilter   string
	cmdMode    bool
	cmdBuf     string
	status     string
	statusErr  bool

	// pane data
	stories    []storyRow
	storiesCur int

	prList     []prRow
	prsCur     int
	prsLoading bool
	prsErr     string

	agents    []agentRow
	agentsCur int

	qaFeatures  int
	qaScenarios int

	designTree []string
	designCur  int

	logLines []string
	logsCur  int

	spItems []spRow
	spCur   int

	shimmerPos int
	shimmerDir int

	exitAction dashAction
}

func newDashboard(t Theme, noAnimate bool) *dashboardModel {
	m := &dashboardModel{
		theme:      t,
		noAnimate:  noAnimate,
		fullscreen: -1,
		shimmerDir: 1,
	}
	m.reload()
	return m
}

// reload refreshes all local (fast) data sources.
func (m *dashboardModel) reload() {
	m.stories = loadStories()
	m.agents = loadAgents()
	m.qaFeatures, m.qaScenarios = loadQA()
	m.designTree = loadDesignTree()
	m.logLines = loadLogLines(200)
	m.spItems = loadSuperpowers()
	m.projectName = loadProjectName()
	// clamp cursors
	m.clampCursor(&m.storiesCur, len(m.stories))
	m.clampCursor(&m.agentsCur, len(m.agents))
	m.clampCursor(&m.designCur, len(m.designTree))
	m.clampCursor(&m.logsCur, len(m.logLines))
	m.clampCursor(&m.spCur, len(m.spItems))
}

func (m *dashboardModel) clampCursor(c *int, n int) {
	if n == 0 {
		*c = 0
		return
	}
	if *c >= n {
		*c = n - 1
	}
	if *c < 0 {
		*c = 0
	}
}

// ─── Init ─────────────────────────────────────────────────────────────────────

func (m *dashboardModel) Init() tea.Cmd {
	cmds := []tea.Cmd{loadPRsCmd()}
	if !m.noAnimate {
		cmds = append(cmds, shimmerTick())
	}
	return tea.Batch(cmds...)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (m *dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case prsLoadedMsg:
		m.prsLoading = false
		m.prsErr = msg.err
		m.prList = msg.items
		m.clampCursor(&m.prsCur, len(m.prList))

	case shimmerTickMsg:
		m.shimmerPos += m.shimmerDir
		if m.shimmerPos >= logoShimmerWidth {
			m.shimmerDir = -1
			m.shimmerPos = logoShimmerWidth - 2
		} else if m.shimmerPos < 0 {
			m.shimmerDir = 1
			m.shimmerPos = 1
		}
		return m, shimmerTick()

	case dashRefreshMsg:
		m.reload()
		m.prsLoading = true
		return m, loadPRsCmd()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *dashboardModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	// Command mode input
	if m.cmdMode {
		switch k {
		case "enter":
			result := m.execCmd(m.cmdBuf)
			m.cmdMode = false
			m.cmdBuf = ""
			if m.exitAction != dashActionNone {
				return m, tea.Quit
			}
			if result != "" {
				m.status = result
				m.statusErr = strings.HasPrefix(result, "✗")
			}
		case "ctrl+c", "esc":
			m.cmdMode = false
			m.cmdBuf = ""
		case "backspace":
			if len(m.cmdBuf) > 0 {
				_, size := utf8.DecodeLastRuneInString(m.cmdBuf)
				m.cmdBuf = m.cmdBuf[:len(m.cmdBuf)-size]
			}
		default:
			if len(k) == 1 {
				m.cmdBuf += k
			}
		}
		return m, nil
	}

	// Superpower picker input
	if m.showSP {
		switch k {
		case "esc", "ctrl+c", "F":
			m.showSP = false
			m.spFilter = ""
		case "enter":
			m.showSP = false
			m.spFilter = ""
			if m.spCur < len(m.filteredSP()) {
				sp := m.filteredSP()[m.spCur]
				m.status = "▸ Superpower: " + sp.name + " — use /feature in Claude Code to run"
			}
		case "backspace":
			if len(m.spFilter) > 0 {
				_, size := utf8.DecodeLastRuneInString(m.spFilter)
				m.spFilter = m.spFilter[:len(m.spFilter)-size]
				m.spCur = 0
			}
		case "up", "k":
			if m.spCur > 0 {
				m.spCur--
			}
		case "down", "j":
			fsp := m.filteredSP()
			if m.spCur < len(fsp)-1 {
				m.spCur++
			}
		default:
			if len(k) == 1 {
				m.spFilter += k
				m.spCur = 0
			}
		}
		return m, nil
	}

	// Help overlay
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	// Global keys
	switch k {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if m.fullscreen < 0 {
			return m, tea.Quit
		}
		m.fullscreen = -1
	case "n":
		m.exitAction = dashActionReq
		return m, tea.Quit
	case "?":
		m.showHelp = true
	case ":":
		m.cmdMode = true
		m.cmdBuf = ""
	case "F":
		m.showSP = true
		m.spFilter = ""
		m.spCur = 0
	case "r":
		m.reload()
		m.prsLoading = true
		return m, loadPRsCmd()
	case "d":
		if m.fullscreen == paneDesign {
			m.fullscreen = -1
		} else {
			m.fullscreen = paneDesign
		}
	case "l":
		if m.fullscreen == paneLogs {
			m.fullscreen = -1
		} else {
			m.fullscreen = paneLogs
			m.logLines = loadLogLines(200)
			m.logsCur = len(m.logLines) - 1
			if m.logsCur < 0 {
				m.logsCur = 0
			}
		}
	case "s":
		m.focus = paneStories
		m.fullscreen = -1
	case "a":
		m.focus = paneAgents
		m.fullscreen = -1
	case "p":
		m.focus = panePRs
		m.fullscreen = -1
	case "Q":
		m.focus = paneQA
		m.fullscreen = -1
	case "tab":
		m.focus = (m.focus + 1) % paneCount
		m.fullscreen = -1
	case "shift+tab":
		m.focus = (m.focus - 1 + paneCount) % paneCount
		m.fullscreen = -1
	case "j", "down":
		m.moveCursorDown()
	case "k", "up":
		m.moveCursorUp()
	case "g":
		m.moveCursorTop()
	case "G":
		m.moveCursorBottom()
	}
	return m, nil
}

func (m *dashboardModel) moveCursorDown() {
	switch {
	case m.fullscreen == paneDesign:
		m.clampCursor(&m.designCur, len(m.designTree))
		if m.designCur < len(m.designTree)-1 {
			m.designCur++
		}
	case m.fullscreen == paneLogs:
		if m.logsCur < len(m.logLines)-1 {
			m.logsCur++
		}
	default:
		switch m.focus {
		case paneStories:
			if m.storiesCur < len(m.stories)-1 {
				m.storiesCur++
			}
		case paneAgents:
			if m.agentsCur < len(m.agents)-1 {
				m.agentsCur++
			}
		case panePRs:
			if m.prsCur < len(m.prList)-1 {
				m.prsCur++
			}
		}
	}
}

func (m *dashboardModel) moveCursorUp() {
	switch {
	case m.fullscreen == paneDesign:
		if m.designCur > 0 {
			m.designCur--
		}
	case m.fullscreen == paneLogs:
		if m.logsCur > 0 {
			m.logsCur--
		}
	default:
		switch m.focus {
		case paneStories:
			if m.storiesCur > 0 {
				m.storiesCur--
			}
		case paneAgents:
			if m.agentsCur > 0 {
				m.agentsCur--
			}
		case panePRs:
			if m.prsCur > 0 {
				m.prsCur--
			}
		}
	}
}

func (m *dashboardModel) moveCursorTop() {
	switch {
	case m.fullscreen == paneDesign:
		m.designCur = 0
	case m.fullscreen == paneLogs:
		m.logsCur = 0
	default:
		switch m.focus {
		case paneStories:
			m.storiesCur = 0
		case paneAgents:
			m.agentsCur = 0
		case panePRs:
			m.prsCur = 0
		}
	}
}

func (m *dashboardModel) moveCursorBottom() {
	switch {
	case m.fullscreen == paneDesign:
		if len(m.designTree) > 0 {
			m.designCur = len(m.designTree) - 1
		}
	case m.fullscreen == paneLogs:
		if len(m.logLines) > 0 {
			m.logsCur = len(m.logLines) - 1
		}
	default:
		switch m.focus {
		case paneStories:
			if len(m.stories) > 0 {
				m.storiesCur = len(m.stories) - 1
			}
		case paneAgents:
			if len(m.agents) > 0 {
				m.agentsCur = len(m.agents) - 1
			}
		case panePRs:
			if len(m.prList) > 0 {
				m.prsCur = len(m.prList) - 1
			}
		}
	}
}

func (m *dashboardModel) execCmd(input string) string {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return ""
	}
	switch parts[0] {
	case "theme":
		if len(parts) < 2 {
			return "usage: theme <name>  (tokyo-night | catppuccin-mocha | gruvbox | nord | everforest)"
		}
		m.theme = themeByName(parts[1])
		return "✓ theme switched to " + parts[1]
	case "reload", "sync":
		m.reload()
		m.prsLoading = true
		return "✓ reloading…"
	case "req", "new", "story":
		m.exitAction = dashActionReq
		// return empty so Quit fires on next render cycle
		return ""
	case "help":
		m.showHelp = true
		return ""
	default:
		return "✗ unknown command: " + parts[0]
	}
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m *dashboardModel) View() string {
	if m.width == 0 {
		return ""
	}

	// Narrow terminal: degrade to single-column log mode
	if m.width < 80 {
		return m.narrowView()
	}

	// Full-screen overlays
	if m.fullscreen == paneDesign {
		return m.header() + m.designView() + m.footer()
	}
	if m.fullscreen == paneLogs {
		return m.header() + m.logsView() + m.footer()
	}

	// Overlays (rendered over the grid)
	if m.showHelp {
		return m.header() + m.helpView() + m.footer()
	}
	if m.showSP {
		return m.header() + m.spPickerView() + m.footer()
	}

	// Normal 2×2 dashboard
	return m.header() + m.gridView() + m.footer()
}

func (m *dashboardModel) header() string {
	t := m.theme
	name := m.projectName
	if name == "" {
		name = "—"
	}
	left := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("  squad")
	mid := lipgloss.NewStyle().Foreground(t.Muted).Render(" · project: " + name + " · theme: " + t.Name)
	bar := left + mid
	if m.noAnimate {
		return logo() + "\n" + bar + "\n"
	}
	return logoShimmer(m.shimmerPos, t.Primary, t.Accent) + bar + "\n"
}

func (m *dashboardModel) footer() string {
	t := m.theme
	keys := "  [Tab] cycle · [s/a/p/Q] pane · [d]esign · [l]ogs · [n] new story · [F] superpowers · [:] cmd · [?] help · [q] quit"
	if m.cmdMode {
		keys = "  :" + m.cmdBuf + "█"
	} else if m.status != "" {
		col := t.Success
		if m.statusErr {
			col = t.Error
		}
		keys = lipgloss.NewStyle().Foreground(col).Render("  " + m.status)
	}
	return "\n" + lipgloss.NewStyle().Foreground(t.Muted).Render(keys) + "\n"
}

// gridView renders the 2×2 pane layout.
func (m *dashboardModel) gridView() string {
	t := m.theme
	innerH := m.height - 14 // subtract header(9) + footer(2) + separators
	if innerH < 6 {
		innerH = 6
	}
	paneH := innerH / 2
	halfW := (m.width - 4) / 2

	style := func(active bool) lipgloss.Style {
		base := lipgloss.NewStyle().
			Width(halfW).Height(paneH).
			Border(lipgloss.RoundedBorder()).
			PaddingLeft(1)
		if active {
			return base.BorderForeground(t.Primary)
		}
		return base.BorderForeground(t.Muted)
	}

	topLeft := style(m.focus == paneStories).Render(m.storiesContent(paneH - 2))
	topRight := style(m.focus == paneAgents).Render(m.agentsContent(paneH - 2))
	botLeft := style(m.focus == panePRs).Render(m.prsContent(paneH - 2))
	botRight := style(m.focus == paneQA).Render(m.qaContent(paneH - 2))

	top := lipgloss.JoinHorizontal(lipgloss.Top, topLeft, "  ", topRight)
	bot := lipgloss.JoinHorizontal(lipgloss.Top, botLeft, "  ", botRight)
	return lipgloss.JoinVertical(lipgloss.Left, top, bot)
}

// ─── Pane content renderers ───────────────────────────────────────────────────

func (m *dashboardModel) storiesContent(height int) string {
	t := m.theme
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("Stories")
	if len(m.stories) == 0 {
		return title + "\n" + lipgloss.NewStyle().Foreground(t.Muted).Render("no stories yet — run squad req")
	}
	lines := []string{title}
	cursor := lipgloss.NewStyle().Foreground(t.Accent).Render("▸")
	for i, s := range m.stories {
		if i >= height-2 {
			break
		}
		phaseTag := lipgloss.NewStyle().Foreground(t.Muted).Render(truncate(s.phase, 8))
		idStr := lipgloss.NewStyle().Foreground(t.Foreground).Render(truncate(s.id, 20))
		var line string
		if i == m.storiesCur && m.focus == paneStories {
			line = cursor + " " + idStr + " " + phaseTag
		} else {
			line = "  " + idStr + " " + phaseTag
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m *dashboardModel) agentsContent(height int) string {
	t := m.theme
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("Recent Agents")
	if len(m.agents) == 0 {
		return title + "\n" + lipgloss.NewStyle().Foreground(t.Muted).Render("no activity in .claude/logs/")
	}
	lines := []string{title}
	cursor := lipgloss.NewStyle().Foreground(t.Accent).Render("▸")
	for i, a := range m.agents {
		if i >= height-2 {
			break
		}
		name := lipgloss.NewStyle().Foreground(t.Foreground).Render(truncate(a.agent, 14))
		op := lipgloss.NewStyle().Foreground(t.Muted).Render(truncate(a.op, 16))
		var line string
		if i == m.agentsCur && m.focus == paneAgents {
			line = cursor + " " + name + " " + op
		} else {
			line = "  " + name + " " + op
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m *dashboardModel) prsContent(height int) string {
	t := m.theme
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("PRs")
	if m.prsLoading {
		return title + "\n" + lipgloss.NewStyle().Foreground(t.Muted).Render("loading…")
	}
	if m.prsErr != "" {
		return title + "\n" + lipgloss.NewStyle().Foreground(t.Warning).Render(truncate(m.prsErr, 40))
	}
	if len(m.prList) == 0 {
		return title + "\n" + lipgloss.NewStyle().Foreground(t.Muted).Render("no open PRs")
	}
	lines := []string{title}
	cursor := lipgloss.NewStyle().Foreground(t.Accent).Render("▸")
	for i, pr := range m.prList {
		if i >= height-2 {
			break
		}
		stateCol := t.Success
		stateIcon := "✓"
		if pr.state == "OPEN" {
			stateCol = t.Accent
			stateIcon = "●"
		} else if pr.state == "CLOSED" {
			stateCol = t.Muted
			stateIcon = "✗"
		}
		num := lipgloss.NewStyle().Foreground(stateCol).Render(fmt.Sprintf("#%d %s", pr.number, stateIcon))
		ttl := lipgloss.NewStyle().Foreground(t.Foreground).Render(truncate(pr.title, 28))
		var line string
		if i == m.prsCur && m.focus == panePRs {
			line = cursor + " " + num + " " + ttl
		} else {
			line = "  " + num + " " + ttl
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m *dashboardModel) qaContent(_ int) string {
	t := m.theme
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("QA / Gherkin")
	if m.qaFeatures == 0 {
		return title + "\n" + lipgloss.NewStyle().Foreground(t.Muted).Render("no .feature files in tests/features/")
	}
	fLine := lipgloss.NewStyle().Foreground(t.Success).Render(
		fmt.Sprintf("  %d feature file(s)", m.qaFeatures))
	sLine := lipgloss.NewStyle().Foreground(t.Foreground).Render(
		fmt.Sprintf("  %d scenario(s) total", m.qaScenarios))
	hint := lipgloss.NewStyle().Foreground(t.Muted).Render("  make test-features-sync to regenerate")
	return strings.Join([]string{title, fLine, sLine, hint}, "\n")
}

func (m *dashboardModel) designView() string {
	t := m.theme
	innerH := m.height - 14
	if innerH < 4 {
		innerH = 4
	}
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("Design Artifacts")
	sep := lipgloss.NewStyle().Foreground(t.Muted).Render(strings.Repeat("─", m.width-4))
	if len(m.designTree) == 0 {
		return title + "\n" + sep + "\n" + lipgloss.NewStyle().Foreground(t.Muted).Render("docs/design/ is empty")
	}
	lines := []string{title, sep}
	cursor := lipgloss.NewStyle().Foreground(t.Accent).Render("▸")
	start := m.designCur - innerH/2
	if start < 0 {
		start = 0
	}
	for i, entry := range m.designTree[start:] {
		if i >= innerH {
			break
		}
		abs := start + i
		var line string
		if abs == m.designCur {
			line = cursor + " " + lipgloss.NewStyle().Foreground(t.Foreground).Render(entry)
		} else {
			line = "  " + lipgloss.NewStyle().Foreground(t.Muted).Render(entry)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m *dashboardModel) logsView() string {
	t := m.theme
	innerH := m.height - 14
	if innerH < 4 {
		innerH = 4
	}
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("Skill Logs (.claude/logs/skills.jsonl)")
	sep := lipgloss.NewStyle().Foreground(t.Muted).Render(strings.Repeat("─", m.width-4))
	if len(m.logLines) == 0 {
		return title + "\n" + sep + "\n" + lipgloss.NewStyle().Foreground(t.Muted).Render("no log entries yet")
	}
	lines := []string{title, sep}
	start := m.logsCur - innerH + 2
	if start < 0 {
		start = 0
	}
	for i, entry := range m.logLines[start:] {
		if i >= innerH {
			break
		}
		abs := start + i
		col := t.Muted
		if abs == m.logsCur {
			col = t.Foreground
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(col).Render("  "+truncate(entry, m.width-6)))
	}
	hint := lipgloss.NewStyle().Foreground(t.Muted).Render(
		fmt.Sprintf("  j/k scroll · %d/%d", m.logsCur+1, len(m.logLines)))
	lines = append(lines, hint)
	return strings.Join(lines, "\n")
}

func (m *dashboardModel) filteredSP() []spRow {
	if m.spFilter == "" {
		return m.spItems
	}
	f := strings.ToLower(m.spFilter)
	var out []spRow
	for _, sp := range m.spItems {
		if strings.Contains(strings.ToLower(sp.name), f) || strings.Contains(strings.ToLower(sp.desc), f) {
			out = append(out, sp)
		}
	}
	return out
}

func (m *dashboardModel) spPickerView() string {
	t := m.theme
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("Superpowers")
	filter := lipgloss.NewStyle().Foreground(t.Muted).Render("  filter: ") +
		lipgloss.NewStyle().Foreground(t.Foreground).Render(m.spFilter+"█")
	sep := lipgloss.NewStyle().Foreground(t.Muted).Render(strings.Repeat("─", 52))
	fsp := m.filteredSP()
	lines := []string{title, filter, sep}
	cursor := lipgloss.NewStyle().Foreground(t.Accent).Render("▸")
	if len(fsp) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(t.Muted).Render("  no matches"))
	}
	for i, sp := range fsp {
		name := lipgloss.NewStyle().Foreground(t.Foreground).Bold(true).Render(fmt.Sprintf("%-22s", sp.name))
		desc := lipgloss.NewStyle().Foreground(t.Muted).Render(truncate(sp.desc, 36))
		var line string
		if i == m.spCur {
			line = cursor + " " + name + " " + desc
		} else {
			line = "    " + name + " " + desc
		}
		lines = append(lines, line)
	}
	lines = append(lines, "", lipgloss.NewStyle().Foreground(t.Muted).Render("  Enter select · Esc cancel · type to filter"))
	return strings.Join(lines, "\n")
}

func (m *dashboardModel) helpView() string {
	t := m.theme
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("Keybindings")
	sep := lipgloss.NewStyle().Foreground(t.Muted).Render(strings.Repeat("─", 60))
	pairs := [][2]string{
		{"Tab / Shift+Tab", "cycle panes"},
		{"j / k", "move down / up"},
		{"g / G", "top / bottom"},
		{"s a p Q", "focus Stories / Agents / PRs / QA"},
		{"d", "toggle Design pane (full-screen)"},
		{"l", "toggle Logs pane (full-screen)"},
		{"n", "new story — open Gherkin requirements editor"},
		{"F", "Superpower picker"},
		{":", "command mode  (:theme, :reload, :req / :new / :story)"},
		{"r", "refresh data"},
		{"?", "this help"},
		{"q / Ctrl+C", "quit"},
	}
	lines := []string{title, sep}
	for _, p := range pairs {
		key := lipgloss.NewStyle().Foreground(t.Accent).Render(fmt.Sprintf("  %-22s", p[0]))
		val := lipgloss.NewStyle().Foreground(t.Foreground).Render(p[1])
		lines = append(lines, key+val)
	}
	lines = append(lines, "", lipgloss.NewStyle().Foreground(t.Muted).Render("  Press any key to close"))
	return strings.Join(lines, "\n")
}

func (m *dashboardModel) narrowView() string {
	t := m.theme
	lines := []string{
		lipgloss.NewStyle().Foreground(t.Warning).Render("  Terminal < 80 cols — narrow mode"),
		lipgloss.NewStyle().Foreground(t.Muted).Render(fmt.Sprintf("  %d stories · %d PRs · %d scenarios",
			len(m.stories), len(m.prList), m.qaScenarios)),
		"",
	}
	for i, s := range m.stories {
		if i >= 10 {
			lines = append(lines, lipgloss.NewStyle().Foreground(t.Muted).Render("  …"))
			break
		}
		lines = append(lines, "  "+s.id+" "+s.phase)
	}
	lines = append(lines, "", lipgloss.NewStyle().Foreground(t.Muted).Render("  q quit"))
	return strings.Join(lines, "\n")
}

// ─── Data loaders ─────────────────────────────────────────────────────────────

func loadStories() []storyRow {
	var rows []storyRow
	// New format: docs/stories/*/Story.md
	dirs, _ := filepath.Glob("docs/stories/*/Story.md")
	for _, p := range dirs {
		if r, ok := parseStoryFile(p); ok {
			rows = append(rows, r)
		}
	}
	// Legacy format: docs/stories/*.md (excluding _template.md)
	files, _ := filepath.Glob("docs/stories/*.md")
	for _, p := range files {
		if strings.HasPrefix(filepath.Base(p), "_") {
			continue
		}
		if r, ok := parseStoryFile(p); ok {
			rows = append(rows, r)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].id < rows[j].id })
	return rows
}

func parseStoryFile(path string) (storyRow, bool) {
	f, err := os.Open(path)
	if err != nil {
		return storyRow{}, false
	}
	defer f.Close()

	fm := extractFrontmatter(f)
	if len(fm) == 0 {
		return storyRow{}, false
	}

	r := storyRow{
		id:       fm["id"],
		slug:     fm["story_slug"],
		priority: fm["priority"],
	}
	if r.id == "" {
		r.id = fm["story_id"]
	}
	if r.id == "" {
		r.id = filepath.Base(filepath.Dir(path))
	}
	r.phase = extractPhaseFromLabels(fm["labels"])
	if fm["ui"] == "true" {
		r.ui = true
	}
	if n, err := strconv.Atoi(fm["issue_number"]); err == nil {
		r.issue = n
	}
	return r, true
}

func extractFrontmatter(f *os.File) map[string]string {
	m := map[string]string{}
	scanner := bufio.NewScanner(f)
	inFM := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			if !inFM {
				inFM = true
				continue
			}
			break
		}
		if !inFM {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		k := strings.TrimSpace(line[:idx])
		v := strings.TrimSpace(line[idx+1:])
		v = strings.Trim(v, `"'`)
		m[k] = v
	}
	return m
}

func extractPhaseFromLabels(labels string) string {
	// labels field looks like: ["type:feature", "phase:implement"]
	for _, part := range strings.Split(labels, ",") {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `[]"' `)
		if strings.HasPrefix(part, "phase:") {
			return strings.TrimPrefix(part, "phase:")
		}
	}
	return "discover"
}

func loadPRsCmd() tea.Cmd {
	return func() tea.Msg {
		ghPath, err := exec.LookPath("gh")
		if err != nil {
			return prsLoadedMsg{err: "gh not found"}
		}
		out, err := exec.Command(ghPath, "pr", "list",
			"--json", "number,title,state",
			"--limit", "20",
		).Output()
		if err != nil {
			return prsLoadedMsg{err: "gh pr list: " + strings.TrimSpace(string(out))}
		}
		var raw []struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
			State  string `json:"state"`
		}
		if err := json.Unmarshal(out, &raw); err != nil {
			return prsLoadedMsg{err: "parse error"}
		}
		rows := make([]prRow, len(raw))
		for i, r := range raw {
			rows[i] = prRow{r.Number, r.Title, r.State}
		}
		return prsLoadedMsg{items: rows}
	}
}

func loadAgents() []agentRow {
	data, err := os.ReadFile(".claude/logs/skills.jsonl")
	if err != nil {
		return nil
	}
	var rows []agentRow
	seen := map[string]bool{}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	// Walk backward — most recent first
	for i := len(lines) - 1; i >= 0 && len(rows) < 8; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		agent, _ := entry["agent"].(string)
		if agent == "" {
			agent, _ = entry["skill"].(string)
		}
		if agent == "" {
			continue
		}
		op, _ := entry["op"].(string)
		file, _ := entry["file"].(string)
		ts, _ := entry["ts"].(string)
		key := agent + "|" + op
		if seen[key] {
			continue
		}
		seen[key] = true
		rows = append(rows, agentRow{agent: agent, op: op, file: file, ts: ts})
	}
	return rows
}

func loadQA() (files int, scenarios int) {
	entries, err := filepath.Glob("tests/features/*.feature")
	if err != nil {
		return
	}
	files = len(entries)
	for _, p := range entries {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, l := range strings.Split(string(data), "\n") {
			t := strings.TrimSpace(l)
			if strings.HasPrefix(t, "Scenario:") || strings.HasPrefix(t, "Scenario Outline:") {
				scenarios++
			}
		}
	}
	return
}

func loadDesignTree() []string {
	var entries []string
	_ = filepath.WalkDir("docs/design", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if filepath.Base(path) == ".gitkeep" {
			return nil
		}
		rel := strings.TrimPrefix(path, "docs/design/")
		if rel == "" || rel == "docs/design" {
			return nil
		}
		prefix := strings.Repeat("  ", strings.Count(rel, string(os.PathSeparator)))
		icon := "📄 "
		if d.IsDir() {
			icon = "📁 "
		}
		entries = append(entries, prefix+icon+filepath.Base(path))
		return nil
	})
	return entries
}

func loadLogLines(n int) []string {
	data, err := os.ReadFile(".claude/logs/skills.jsonl")
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	// Pretty-format each JSON line
	var out []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			out = append(out, l)
			continue
		}
		parts := []string{}
		for _, k := range []string{"ts", "agent", "skill", "op", "file", "duration", "error"} {
			if v, ok := m[k]; ok && v != nil && v != "" {
				parts = append(parts, fmt.Sprintf("%s=%v", k, v))
			}
		}
		out = append(out, strings.Join(parts, "  "))
	}
	return out
}

func loadSuperpowers() []spRow {
	var rows []spRow
	for _, dir := range []string{".claude/superpowers", "template/.claude/superpowers"} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".yaml") || e.Name() == "schema.yaml" {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			name, desc := parseSuperpowerYAML(string(data))
			if name != "" {
				rows = append(rows, spRow{name: name, desc: desc, path: filepath.Join(dir, e.Name())})
			}
		}
		if len(rows) > 0 {
			break
		}
	}
	return rows
}

func parseSuperpowerYAML(content string) (name, desc string) {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			name = strings.Trim(name, `"'`)
		}
		if strings.HasPrefix(line, "description:") {
			desc = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			desc = strings.Trim(desc, `"'`)
		}
	}
	return
}

func loadProjectName() string {
	data, err := os.ReadFile("project.config.yaml")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			n := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			return strings.Trim(n, `"'`)
		}
	}
	return ""
}

// ─── Utility ─────────────────────────────────────────────────────────────────

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return string(runes[:n-1]) + "…"
}

// ─── Entry point ─────────────────────────────────────────────────────────────

// runDashboard runs the dashboard and returns the exit action so the caller
// can decide what to do next (e.g. launch the req workflow).
func runDashboard(t Theme, noAnimate bool) (dashAction, error) {
	m := newDashboard(t, noAnimate)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return dashActionNone, err
	}
	return final.(*dashboardModel).exitAction, nil
}
