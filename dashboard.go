package main

import (
	"fmt"
	"os"
	"os/exec"
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
	path     string // path to Story.md
}

type prRow struct {
	number int
	title  string
	state  string // OPEN / CLOSED / MERGED
}

type agentRow struct {
	agent     string
	op        string
	file      string
	ts        string
	source    string // "claude", "opencode", "maple"
	sessionID string // for session detail drill-down
}

// ─── Dashboard exit actions ───────────────────────────────────────────────────

type dashAction int

const (
	dashActionNone    dashAction = iota
	dashActionQuit              // plain quit — no follow-up workflow
	dashActionReq               // quit and run req (Gherkin converter)
	dashActionUpdate            // quit and run init --force (re-sync template)
	dashActionLabels            // quit and run labels bootstrap
	dashActionProject           // quit and run project creation
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

type prDetailLoadedMsg struct {
	lines []string
	title string
	err   string
}

type dashRefreshMsg struct{}
type statusClearMsg struct{}
type dashTickMsg struct{} // periodic local-data refresh (no network)

type qaTestResultMsg struct {
	output string
	failed bool
}

const dashTickInterval = 5 * time.Second

func dashTickCmd() tea.Cmd {
	return tea.Tick(dashTickInterval, func(time.Time) tea.Msg { return dashTickMsg{} })
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
	showHelp        bool
	showSkills      bool
	skillsTabSearch bool // false = Installed tab, true = Search tab
	skillsQuery     string
	skillsItems     []skillRow
	skillsCur       int
	skillsLoading   bool
	skillsErr       string
	skillsSearched  bool // true after first search attempt
	installedSkills []installedSkillRow
	installedCur    int
	installedLoading bool
	installedErr    string
	npxPath         string // cached npx binary path
	cmdMode       bool
	cmdBuf     string
	searchMode bool
	searchBuf  string
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
	qaFiles     []string // paths to .feature files
	qaFileCur   int

	designTree []string
	designCur  int

	logLines []string
	logsCur  int

	lastKey    string // for gg double-key detection
	debugMode  bool   // :debug — tee state to .claude/logs/tui.log

	// story detail overlay
	showStory      bool
	storyLines     []string
	storyScroll    int
	storyTitle     string
	storyDir       string // directory of the open story (for re-edit cleanup)

	// session detail overlay
	showSession   bool
	sessionLines  []string
	sessionScroll int
	sessionTitle  string
	sessionSource string

	// QA feature file overlay
	showQAFile   bool
	qaFileLines  []string
	qaFileScroll int
	qaFileTitle  string

	// PR detail overlay
	showPRDetail    bool
	prDetailLines   []string
	prDetailScroll  int
	prDetailTitle   string
	prDetailLoading bool

	exitAction dashAction
}

func newDashboard(t Theme, noAnimate bool) *dashboardModel {
	npx, _ := exec.LookPath("npx")
	m := &dashboardModel{
		theme:      t,
		noAnimate:  noAnimate,
		fullscreen: -1,
		npxPath:    npx,
	}
	m.reload()
	return m
}

// reload refreshes all local (fast) data sources.
func (m *dashboardModel) reload() {
	m.stories = loadStories()
	m.agents = loadAgents()
	m.qaFeatures, m.qaScenarios, m.qaFiles = loadQA()
	m.designTree = loadDesignTree()
	m.logLines = loadLogLines(200)
	m.projectName = loadProjectName()
	// clamp cursors
	m.clampCursor(&m.storiesCur, len(m.stories))
	m.clampCursor(&m.agentsCur, len(m.agents))
	m.clampCursor(&m.qaFileCur, len(m.qaFiles))
	m.clampCursor(&m.designCur, len(m.designTree))
	m.clampCursor(&m.logsCur, len(m.logLines))
}

// setStatus sets a status message and returns a Cmd to auto-clear it after 3s.
func (m *dashboardModel) setStatus(msg string, isErr bool) tea.Cmd {
	m.status = msg
	m.statusErr = isErr
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg { return statusClearMsg{} })
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
	return tea.Batch(loadPRsCmd(), dashTickCmd())
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

	case prDetailLoadedMsg:
		m.prDetailLoading = false
		if msg.err != "" {
			m.prDetailLines = []string{"error: " + msg.err}
		} else {
			m.prDetailLines = msg.lines
		}
		m.prDetailTitle = msg.title
		m.prDetailScroll = 0
		m.showPRDetail = true

	case dashRefreshMsg:
		m.reload()
		m.prsLoading = true
		return m, loadPRsCmd()

	case dashTickMsg:
		m.reload()
		return m, dashTickCmd() // re-schedule; tea.Tick does not repeat

	case qaTestResultMsg:
		if msg.failed {
			return m, m.setStatus("✗ "+msg.output, true)
		}
		return m, m.setStatus("✓ "+msg.output, false)

	case statusClearMsg:
		m.status = ""
		m.statusErr = false

	case skillsSearchedMsg:
		m.skillsLoading = false
		m.skillsSearched = msg.searched
		if msg.err != nil {
			m.skillsErr = msg.err.Error()
		} else {
			m.skillsItems = msg.items
			m.skillsErr = ""
			m.skillsCur = 0
		}

	case skillInstalledMsg:
		m.skillsLoading = false
		if msg.err != nil {
			return m, m.setStatus("✗ install "+msg.pkg+": "+msg.err.Error(), true)
		}
		m.showSkills = false
		return m, m.setStatus("✓ installed "+msg.pkg, false)

	case installedLoadedMsg:
		m.installedLoading = false
		if msg.err != nil {
			m.installedErr = msg.err.Error()
		} else {
			m.installedSkills = msg.items
			m.installedErr = ""
			m.clampCursor(&m.installedCur, len(m.installedSkills))
		}

	case skillRemovedMsg:
		m.installedLoading = false
		if msg.err != nil {
			return m, m.setStatus("✗ remove "+msg.name+": "+msg.err.Error(), true)
		}
		// Reload installed list
		m.installedLoading = true
		return m, tea.Batch(listInstalledSkillsCmd(), m.setStatus("✓ removed "+msg.name, false))

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *dashboardModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	// Search mode input
	if m.searchMode {
		switch k {
		case "enter", "esc", "ctrl+c":
			m.searchMode = false
			if k == "esc" || k == "ctrl+c" {
				m.searchBuf = ""
			}
		case "backspace":
			if len(m.searchBuf) > 0 {
				_, size := utf8.DecodeLastRuneInString(m.searchBuf)
				m.searchBuf = m.searchBuf[:len(m.searchBuf)-size]
			}
		default:
			if len(k) == 1 {
				m.searchBuf += k
			}
		}
		return m, nil
	}

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
				return m, m.setStatus(result, strings.HasPrefix(result, "✗"))
			}
		case "ctrl+c", "esc":
			m.cmdMode = false
			m.cmdBuf = ""
			m.status = ""
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

	// Skills marketplace browser
	if m.showSkills {
		switch k {
		case "esc", "ctrl+c", "F":
			m.showSkills = false
			m.skillsQuery = ""
		case "tab":
			m.skillsTabSearch = !m.skillsTabSearch
		case "up", "k":
			if m.skillsTabSearch {
				if m.skillsCur > 0 {
					m.skillsCur--
				}
			} else {
				if m.installedCur > 0 {
					m.installedCur--
				}
			}
		case "down", "j":
			if m.skillsTabSearch {
				if m.skillsCur < len(m.skillsItems)-1 {
					m.skillsCur++
				}
			} else {
				if m.installedCur < len(m.installedSkills)-1 {
					m.installedCur++
				}
			}
		case "d":
			if !m.skillsTabSearch && !m.installedLoading && len(m.installedSkills) > 0 {
				sk := m.installedSkills[m.installedCur]
				m.installedLoading = true
				return m, removeSkillCmd(sk.name)
			}
		case "enter":
			if m.skillsTabSearch {
				if m.skillsLoading {
					// ignore while loading
				} else if len(m.skillsItems) > 0 && m.skillsCur < len(m.skillsItems) {
					sk := m.skillsItems[m.skillsCur]
					m.skillsLoading = true
					return m, installSkillCmd(sk.pkg)
				} else {
					m.skillsLoading = true
					m.skillsItems = nil
					m.skillsCur = 0
					m.skillsErr = ""
					return m, searchSkillsCmd(m.skillsQuery)
				}
			}
		case "backspace":
			if m.skillsTabSearch && len(m.skillsQuery) > 0 {
				_, size := utf8.DecodeLastRuneInString(m.skillsQuery)
				m.skillsQuery = m.skillsQuery[:len(m.skillsQuery)-size]
				m.skillsItems = nil
			}
		default:
			if m.skillsTabSearch && len(k) == 1 {
				m.skillsQuery += k
				m.skillsItems = nil
			}
		}
		return m, nil
	}

	// Help overlay — any key closes
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	// Session detail overlay
	if m.showSession {
		maxScroll := len(m.sessionLines) - (m.height - 14)
		if maxScroll < 0 {
			maxScroll = 0
		}
		switch k {
		case "j", "down":
			if m.sessionScroll < maxScroll {
				m.sessionScroll++
			}
		case "k", "up":
			if m.sessionScroll > 0 {
				m.sessionScroll--
			}
		case "g":
			m.sessionScroll = 0
		case "G":
			m.sessionScroll = maxScroll
		case "q", "esc", "b", "ctrl+c":
			m.showSession = false
		}
		return m, nil
	}

	// Story detail overlay
	if m.showStory {
		maxScroll := len(m.storyLines) - (m.height - 14)
		if maxScroll < 0 {
			maxScroll = 0
		}
		switch k {
		case "j", "down":
			if m.storyScroll < maxScroll {
				m.storyScroll++
			}
		case "k", "up":
			if m.storyScroll > 0 {
				m.storyScroll--
			}
		case "g":
			m.storyScroll = 0
		case "G":
			m.storyScroll = maxScroll
		case "e":
			// Extract Gherkin from story content and write to handoff file
			// so runReq pre-loads it into the textarea.
			gherkin := extractGherkinFromLines(m.storyLines)
			_ = os.MkdirAll(".claude/state", 0o755)
			_ = os.WriteFile(".claude/state/maple-edit.txt", []byte(gherkin), 0o644)
			if m.storyDir != "" {
				_ = os.RemoveAll(m.storyDir)
				m.storyDir = ""
			}
			m.exitAction = dashActionReq
			m.showStory = false
			return m, tea.Quit
		case "q", "esc", "b", "ctrl+c":
			m.showStory = false
		}
		return m, nil
	}

	// QA feature file overlay
	if m.showQAFile {
		maxScroll := len(m.qaFileLines) - (m.height - 14)
		if maxScroll < 0 {
			maxScroll = 0
		}
		switch k {
		case "j", "down":
			if m.qaFileScroll < maxScroll {
				m.qaFileScroll++
			}
		case "k", "up":
			if m.qaFileScroll > 0 {
				m.qaFileScroll--
			}
		case "g":
			m.qaFileScroll = 0
		case "G":
			m.qaFileScroll = maxScroll
		case "r":
			m.showQAFile = false
			return m, m.runFeatureTestCmd(m.qaFileTitle)
		case "q", "esc", "b", "ctrl+c":
			m.showQAFile = false
		}
		return m, nil
	}

	// PR detail overlay
	if m.showPRDetail {
		maxScroll := len(m.prDetailLines) - (m.height - 14)
		if maxScroll < 0 {
			maxScroll = 0
		}
		switch k {
		case "j", "down":
			if m.prDetailScroll < maxScroll {
				m.prDetailScroll++
			}
		case "k", "up":
			if m.prDetailScroll > 0 {
				m.prDetailScroll--
			}
		case "g":
			m.prDetailScroll = 0
		case "G":
			m.prDetailScroll = maxScroll
		case "o":
			if len(m.prList) > 0 && m.prsCur < len(m.prList) {
				_ = exec.Command("gh", "pr", "view", fmt.Sprintf("%d", m.prList[m.prsCur].number), "--web").Start()
			}
		case "q", "esc", "b", "ctrl+c":
			m.showPRDetail = false
		}
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
	case "/":
		m.searchMode = true
		m.searchBuf = ""
	case "F":
		m.showSkills = true
		m.skillsTabSearch = false
		m.skillsQuery = ""
		m.skillsItems = nil
		m.skillsCur = 0
		m.skillsErr = ""
		m.skillsSearched = false
		m.installedSkills = nil
		m.installedCur = 0
		m.installedErr = ""
		m.installedLoading = true
		return m, listInstalledSkillsCmd()
	case "u":
		m.exitAction = dashActionUpdate
		return m, tea.Quit
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
	case "enter":
		if m.focus == paneStories && m.storiesCur < len(m.stories) {
			m.openStoryDetail(m.stories[m.storiesCur])
		} else if m.focus == paneAgents && m.agentsCur < len(m.agents) {
			m.openSessionDetail(m.agents[m.agentsCur])
		} else if m.focus == paneQA && m.qaFileCur < len(m.qaFiles) {
			m.openQAFile(m.qaFiles[m.qaFileCur])
		} else if m.focus == panePRs && m.prsCur < len(m.prList) {
			pr := m.prList[m.prsCur]
			m.prDetailLoading = true
			m.prDetailLines = nil
			m.showPRDetail = false
			return m, loadPRDetailCmd(pr.number, pr.title)
		}
	case "o":
		if m.focus == panePRs && m.prsCur < len(m.prList) {
			_ = exec.Command("gh", "pr", "view", fmt.Sprintf("%d", m.prList[m.prsCur].number), "--web").Start()
			return m, m.setStatus("opening PR in browser…", false)
		}
	case "r":
		if m.focus == paneQA {
			return m, m.runFeatureTestCmd("")
		}
		m.reload()
		m.prsLoading = true
		return m, tea.Batch(loadPRsCmd(), m.setStatus("✓ reloading…", false))
	case "j", "down":
		m.moveCursorDown()
	case "k", "up":
		m.moveCursorUp()
	case "g":
		if m.lastKey == "g" {
			m.moveCursorTop()
			m.lastKey = ""
		} else {
			m.lastKey = "g"
		}
		return m, nil
	case "G":
		m.moveCursorBottom()
	default:
		m.lastKey = ""
		return m, nil
	}
	m.lastKey = ""
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
		case paneQA:
			if m.qaFileCur < len(m.qaFiles)-1 {
				m.qaFileCur++
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
		case paneQA:
			if m.qaFileCur > 0 {
				m.qaFileCur--
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
		case paneQA:
			m.qaFileCur = 0
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
		case paneQA:
			if len(m.qaFiles) > 0 {
				m.qaFileCur = len(m.qaFiles) - 1
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
	// ── Navigation / quit (vim-style) ────────────────────────────────────────
	case "q", "q!", "quit", "wq", "x":
		m.exitAction = dashActionQuit
		return ""
	// ── Reload (vim :e / :e!) ────────────────────────────────────────────────
	case "e", "e!", "r", "reload", "sync":
		m.reload()
		m.prsLoading = true
		return "✓ reloading…"
	// ── Theme ────────────────────────────────────────────────────────────────
	case "theme", "colorscheme", "colo":
		if len(parts) < 2 {
			return "usage: theme <name>  (tokyo-night | catppuccin-mocha | gruvbox | nord | everforest)"
		}
		m.theme = themeByName(parts[1])
		return "✓ theme → " + parts[1]
	// ── Story / requirements ─────────────────────────────────────────────────
	case "req", "new", "story", "n":
		m.exitAction = dashActionReq
		return ""
	// ── Template update ──────────────────────────────────────────────────────
	case "update", "upgrade", "sync-template", "u":
		m.exitAction = dashActionUpdate
		return ""
	// ── GitHub labels ────────────────────────────────────────────────────────
	case "labels":
		m.exitAction = dashActionLabels
		return ""
	// ── GitHub Project v2 ────────────────────────────────────────────────────
	case "project":
		m.exitAction = dashActionProject
		return ""
	// ── Debug ────────────────────────────────────────────────────────────────
	case "debug":
		m.debugMode = !m.debugMode
		if m.debugMode {
			_ = os.MkdirAll(".claude/logs", 0o755)
			return "✓ debug logging → .claude/logs/tui.log"
		}
		return "✓ debug logging off"
	// ── Help ─────────────────────────────────────────────────────────────────
	case "help", "h", "?":
		m.showHelp = true
		return ""
	default:
		return "✗ unknown: " + parts[0] + "  (try :help)"
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
	if m.showSession {
		return m.header() + m.sessionDetailView() + m.footer()
	}
	if m.showStory {
		return m.header() + m.storyDetailView() + m.footer()
	}
	if m.showQAFile {
		return m.header() + m.qaFileDetailView() + m.footer()
	}
	if m.showPRDetail || m.prDetailLoading {
		return m.header() + m.prDetailView() + m.footer()
	}
	if m.showHelp {
		return m.header() + m.helpView() + m.footer()
	}
	if m.showSkills {
		return m.header() + m.skillsBrowserView() + m.footer()
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
	info := lipgloss.NewStyle().Foreground(t.Muted).Render("  project: " + name + " · theme: " + t.Name)
	return logoCompact(t.Primary) + info + "\n"
}

func (m *dashboardModel) footer() string {
	t := m.theme
	keys := "  [Tab] cycle · [s/a/p/Q] pane · [j/k] nav · [Enter] open · [o] browser (PR) · [r] run tests (QA) · [n] new · [u] update · [F] skills · [?] help · [q] quit"
	if m.showSkills {
		if !m.skillsTabSearch {
			if m.installedLoading {
				keys = "  loading installed skills…"
			} else {
				keys = "  [Tab] switch tab · [j/k] navigate · [d] remove · Esc close"
			}
		} else {
			if m.skillsLoading {
				keys = "  searching skills.sh…"
			} else if len(m.skillsItems) > 0 {
				keys = "  [Tab] switch tab · [j/k] navigate · [Enter] install · Esc close"
			} else {
				keys = "  [Tab] switch tab · type a query · [Enter] search · Esc close"
			}
		}
	} else if m.showSession {
		keys = "  [j/k] scroll · [Esc] close"
	} else if m.showStory {
		keys = "  [j/k] scroll · [e] re-edit · [Esc] close"
	} else if m.showQAFile {
		keys = "  [j/k] scroll · [r] run test · [Esc] close"
	} else if m.showPRDetail || m.prDetailLoading {
		keys = "  [j/k] scroll · [o] open in browser · [Esc] close"
	} else if m.searchMode {
		keys = "  /" + m.searchBuf + "█"
	} else if m.cmdMode {
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
	innerH := m.height - 10 // subtract compact header(6) + footer(2) + separators
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

// openQAFile loads a .feature file into the QA file overlay.
func (m *dashboardModel) openQAFile(path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		m.status = "✗ could not read " + path
		m.statusErr = true
		return
	}
	m.qaFileLines = strings.Split(string(raw), "\n")
	m.qaFileScroll = 0
	m.qaFileTitle = path
	m.showQAFile = true
}

// runFeatureTestCmd runs the test suite. If path is non-empty, runs only that file.
func (m *dashboardModel) runFeatureTestCmd(path string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		if path != "" {
			cmd = exec.Command("make", "test-features", "FEATURE="+path)
		} else {
			cmd = exec.Command("make", "test-features")
		}
		out, err := cmd.CombinedOutput()
		summary := strings.TrimSpace(string(out))
		// Keep last 72 chars of output as the status summary
		lines := strings.Split(summary, "\n")
		last := strings.TrimSpace(lines[len(lines)-1])
		if last == "" && len(lines) > 1 {
			last = strings.TrimSpace(lines[len(lines)-2])
		}
		if len(last) > 72 {
			last = "…" + last[len(last)-72:]
		}
		if last == "" {
			last = "tests ran"
		}
		return qaTestResultMsg{output: last, failed: err != nil}
	}
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
	writeRecoveryMarker("running")
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	writeRecoveryMarker("exited")
	if err != nil {
		return dashActionNone, err
	}
	return final.(*dashboardModel).exitAction, nil
}

func writeRecoveryMarker(state string) {
	_ = os.MkdirAll(".claude/state", 0o755)
	data := fmt.Sprintf(`{"state":%q,"ts":%q}`, state, time.Now().UTC().Format(time.RFC3339))
	_ = os.WriteFile(".claude/state/maple.json", []byte(data+"\n"), 0o644)
}
