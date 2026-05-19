package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

type sessionRow struct {
	id        string // JSONL file path (claude) or SQLite session ID (opencode)
	title     string
	source    string // "claude", "opencode"
	ts        string // last activity timestamp
	toolCount int    // number of tool calls
}

type testEntry struct {
	path      string   // relative path to display
	framework string   // "gherkin", "go", "jest", "vitest", "mocha", "pytest", "rspec", "maven", "gradle", "phpunit", "cargo", "npm"
	runCmd    []string // command to run this test
	count     int      // test/scenario count if parseable (0 = unknown)
}

// ─── Dashboard exit actions ───────────────────────────────────────────────────

type dashAction int

const (
	dashActionNone      dashAction = iota
	dashActionQuit                 // plain quit — no follow-up workflow
	dashActionReq                  // quit and run req (Gherkin converter)
	dashActionUpdate               // quit and run init --force (re-sync template)
	dashActionLabels               // quit and run labels bootstrap
	dashActionProject              // quit and run project creation
	dashActionOpenAgent            // quit and exec a session in Claude/OpenCode
	dashActionLaunch               // quit and launch tool with optional command
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

type prApproveResultMsg struct {
	number int
	err    string
}

type dashRefreshMsg struct{}
type statusClearMsg struct{}
type dashTickMsg struct{}    // periodic local-data refresh (no network)
type dashNetTickMsg struct{} // periodic network refresh (gh pr list)

type testRunStartMsg struct {
	title string
}

type testRunDoneMsg struct {
	lines  []string
	failed bool
}

type shipSafeStartMsg struct{}

type shipSafeDoneMsg struct {
	lines  []string
	failed bool
}

type rtkInitDoneMsg struct {
	key string
	err string
}

type spawnSucceededMsg struct{ harness string }
type spawnFailedMsg struct{ args []string }
type designPortalResultMsg struct {
	err   string
	open  bool
	auto  bool
	stage string
}

const dashTickInterval = 5 * time.Second
const dashNetTickInterval = 60 * time.Second

func dashTickCmd() tea.Cmd {
	return tea.Tick(dashTickInterval, func(time.Time) tea.Msg { return dashTickMsg{} })
}

func dashNetTickCmd() tea.Cmd {
	return tea.Tick(dashNetTickInterval, func(time.Time) tea.Msg { return dashNetTickMsg{} })
}

// ─── Model ────────────────────────────────────────────────────────────────────

type dashboardModel struct {
	theme       Theme
	noAnimate   bool
	width       int
	height      int
	projectName string

	focus            dashPane
	fullscreen       dashPane // paneDesign or paneLogs, -1 = none
	showHelp         bool
	showSkills       bool
	skillsTabSearch  bool // false = Installed tab, true = Search tab
	skillsQuery      string
	skillsItems      []skillRow
	skillsCur        int
	skillsLoading    bool
	skillsErr        string
	skillsSearched   bool // true after first search attempt
	installedSkills  []installedSkillRow
	installedCur     int
	installedLoading bool
	installedErr     string
	npxPath          string // cached npx binary path
	cmdMode          bool
	cmdBuf           string
	searchMode       bool
	searchBuf        string
	status           string
	statusErr        bool

	// pane data
	stories    []storyRow
	storiesCur int

	prList     []prRow
	prsCur     int
	prsLoading bool
	prsErr     string

	sessions    []sessionRow
	sessionsCur int

	qaEntries  []testEntry
	qaEntryCur int

	designTree []string
	designCur  int

	logLines []string
	logsCur  int

	lastKey   string // for gg double-key detection
	debugMode bool   // :debug — tee state to .claude/logs/tui.log

	// story detail overlay
	showStory   bool
	storyLines  []string
	storyScroll int
	storyTitle  string
	storyDir    string // directory of the open story (for re-edit cleanup)

	// session detail overlay
	showSession   bool
	sessionLines  []string
	sessionScroll int
	sessionTitle  string
	sessionSource string

	// QA file viewer overlay
	showQAFile      bool
	qaFileLines     []string
	qaFileScroll    int
	qaFileTitle     string
	qaFileFramework string
	qaFileRunCmd    []string

	// QA test run output overlay
	showTestOut    bool
	testOutLines   []string
	testOutScroll  int
	testOutTitle   string
	testOutRunning bool
	testOutFailed  bool

	// PR detail overlay
	showPRDetail    bool
	prDetailLines   []string
	prDetailScroll  int
	prDetailTitle   string
	prDetailLoading bool
	prDetailNumber  int // number of the PR currently shown

	// ShipSafe audit overlay
	showShipSafe    bool
	shipSafeLines   []string
	shipSafeScroll  int
	shipSafeRunning bool
	shipSafeFailed  bool

	// Quick Prompt overlay ([x] key)
	showQuickPrompt bool
	quickMode       string      // "taffy" or "items" (skills+agents)
	taffyItems      []quickItem // loaded once on [x]
	quickItems      []quickItem // skills + agents
	quickItemCur    int
	quickItemScroll int    // top of visible window
	quickSearch     string // live filter string

	// Pipeline status overlay
	showPipeline    bool
	pipelineState   pipelineState
	approvalPending string // non-empty stage name when .claude/state/approval-pending.txt exists
	portalAutoStage string // last approval stage for which portal auto-start was attempted

	// Session launcher overlay
	showLauncher  bool
	launcherCur   int    // index into available tools list
	launcherCmd   string // command typed by user
	launcherInput bool   // true when user is typing the command

	// Pinned sessions (tool → session ID), loaded from .claude/state/sessions.json
	pinnedSessions map[string]string

	// Quick launch overlay (shown after picking an item)
	showQuickLaunch        bool
	quickLaunchName        string
	quickLaunchKind        string
	quickLaunchPrompt      string
	quickLaunchHarness     string
	quickLaunchPickHarness bool
	quickLaunchHarnessCur  int

	// Manual launch modal — shown when spawnInNewTerminal fails
	showManualLaunch   bool
	manualLaunchArgs   []string
	manualLaunchCopied bool // shows "copied!" briefly after [c]

	// RTK harness selector overlay
	showRTKHarness      bool
	rtkHarnessCur       int
	rtkHarnessToggled   map[string]bool // selected in current session (not yet installed)
	rtkHarnessInstalled map[string]bool // already installed (from .claude/state/rtk-harnesses.json)
	rtkHarnessRunning   bool

	openTarget []string // command to exec when exitAction == dashActionOpenAgent
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
	m.sessions = loadSessions()
	m.qaEntries = loadTestEntries()
	m.designTree = loadDesignTree()
	m.logLines = loadLogLines(200)
	m.projectName = loadProjectName()
	m.pinnedSessions = loadPinnedSessions()
	m.rtkHarnessInstalled = loadRTKHarnesses()
	// always keep pipeline state current — the skill updates maple.json every stage
	if ps, err := loadPipelineState(); err == nil {
		m.pipelineState = ps
	}
	m.approvalPending = approvalPending()
	// clamp cursors
	m.clampCursor(&m.storiesCur, len(m.stories))
	m.clampCursor(&m.sessionsCur, len(m.sessions))
	m.clampCursor(&m.qaEntryCur, len(m.qaEntries))
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
	return tea.Batch(loadPRsCmd(), dashTickCmd(), dashNetTickCmd())
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
			return m, m.setStatus("✗ "+msg.err, true)
		}
		m.prDetailLines = msg.lines
		m.prDetailTitle = msg.title
		m.prDetailScroll = 0
		m.showPRDetail = true

	case prApproveResultMsg:
		m.prDetailLoading = false
		if msg.err != "" {
			return m, m.setStatus("✗ "+msg.err, true)
		}
		return m, m.setStatus(fmt.Sprintf("✓ PR #%d approved", msg.number), false)

	case dashRefreshMsg:
		m.reload()
		m.prsLoading = true
		return m, loadPRsCmd()

	case dashTickMsg:
		m.reload()
		if m.approvalPending == "" {
			m.portalAutoStage = ""
			return m, dashTickCmd()
		}
		if m.approvalPending != m.portalAutoStage {
			m.portalAutoStage = m.approvalPending
			return m, tea.Batch(dashTickCmd(), designPortalCmd(false, true, m.approvalPending))
		}
		return m, dashTickCmd()

	case dashNetTickMsg:
		m.prsLoading = true
		return m, tea.Batch(loadPRsCmd(), dashNetTickCmd())

	case testRunStartMsg:
		m.testOutRunning = true
		m.testOutFailed = false
		m.testOutLines = []string{"running…"}
		m.testOutTitle = msg.title
		m.testOutScroll = 0
		m.showTestOut = true

	case testRunDoneMsg:
		m.testOutRunning = false
		m.testOutFailed = msg.failed
		m.testOutLines = msg.lines
		m.testOutScroll = len(msg.lines) - 1 // scroll to bottom
		if m.testOutScroll < 0 {
			m.testOutScroll = 0
		}

	case shipSafeStartMsg:
		m.shipSafeRunning = true
		m.shipSafeFailed = false
		m.shipSafeLines = []string{"running npx ship-safe audit …"}
		m.shipSafeScroll = 0
		m.showShipSafe = true

	case shipSafeDoneMsg:
		m.shipSafeRunning = false
		m.shipSafeFailed = msg.failed
		m.shipSafeLines = msg.lines
		m.shipSafeScroll = len(msg.lines) - 1
		if m.shipSafeScroll < 0 {
			m.shipSafeScroll = 0
		}

	case rtkInitDoneMsg:
		if msg.err != "" {
			return m, m.setStatus("✗ rtk init "+msg.key+": "+msg.err, true)
		}
		saveRTKHarness(msg.key)
		if m.rtkHarnessInstalled == nil {
			m.rtkHarnessInstalled = map[string]bool{}
		}
		m.rtkHarnessInstalled[msg.key] = true
		delete(m.rtkHarnessToggled, msg.key)
		// close overlay once all pending inits are done
		if len(m.rtkHarnessToggled) == 0 {
			m.rtkHarnessRunning = false
			m.showRTKHarness = false
			return m, m.setStatus("✓ rtk wired for selected harnesses", false)
		}

	case spawnSucceededMsg:
		return m, m.setStatus("✓ launched "+msg.harness+" in new terminal", false)

	case spawnFailedMsg:
		m.showManualLaunch = true
		m.manualLaunchArgs = msg.args
		m.manualLaunchCopied = false

	case designPortalResultMsg:
		if msg.err == "" {
			if !msg.auto && msg.open {
				return m, m.setStatus("✓ opened design review portal", false)
			}
			if msg.auto {
				return m, m.setStatus("✓ design review portal ready for stage "+msg.stage, false)
			}
			return m, nil
		}
		if msg.auto {
			return m, nil
		}
		return m, m.setStatus("✗ design portal: "+msg.err, true)

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

	// Test run output overlay
	if m.showTestOut {
		maxScroll := len(m.testOutLines) - (m.height - 14)
		if maxScroll < 0 {
			maxScroll = 0
		}
		switch k {
		case "j", "down":
			if m.testOutScroll < maxScroll {
				m.testOutScroll++
			}
		case "k", "up":
			if m.testOutScroll > 0 {
				m.testOutScroll--
			}
		case "g":
			m.testOutScroll = 0
		case "G":
			m.testOutScroll = maxScroll
		case "q", "esc", "b", "ctrl+c":
			if !m.testOutRunning {
				m.showTestOut = false
			}
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
			return m, m.runTestCmd(testEntry{path: m.qaFileTitle, framework: m.qaFileFramework, runCmd: m.qaFileRunCmd})
		case "q", "esc", "b", "ctrl+c":
			m.showQAFile = false
		}
		return m, nil
	}

	// Quick Prompt overlay
	if m.showQuickPrompt {
		activeItems := m.taffyItems
		if m.quickMode == "items" {
			activeItems = m.quickItems
		}
		filtered := quickFilter(activeItems, m.quickSearch)
		visibleItems := (m.height - 17) / 3
		if visibleItems < 3 {
			visibleItems = 3
		}
		switch k {
		case "t":
			if m.quickMode == "taffy" {
				m.quickMode = "items"
			} else {
				m.quickMode = "taffy"
			}
			m.quickItemCur = 0
			m.quickItemScroll = 0
		case "j", "down":
			if m.quickItemCur < len(filtered)-1 {
				m.quickItemCur++
				if m.quickItemCur >= m.quickItemScroll+visibleItems {
					m.quickItemScroll = m.quickItemCur - visibleItems + 1
				}
			}
		case "k", "up":
			if m.quickItemCur > 0 {
				m.quickItemCur--
				if m.quickItemCur < m.quickItemScroll {
					m.quickItemScroll = m.quickItemCur
				}
			}
		case "enter":
			if m.quickItemCur < len(filtered) {
				item := filtered[m.quickItemCur]
				m.showQuickPrompt = false
				m.quickLaunchName = item.name
				m.quickLaunchKind = item.kind
				if item.kind == "taffy" {
					m.quickLaunchName = "pipeline-runner " + item.name
				}
				m.quickLaunchPrompt = ""
				m.quickLaunchHarnessCur = 0
				m.quickLaunchHarness = ""
				m.quickLaunchPickHarness = true
				m.showQuickLaunch = true
			}
		case "backspace":
			if len(m.quickSearch) > 0 {
				_, size := utf8.DecodeLastRuneInString(m.quickSearch)
				m.quickSearch = m.quickSearch[:len(m.quickSearch)-size]
				m.quickItemCur = 0
				m.quickItemScroll = 0
			}
		case "q", "esc", "ctrl+c":
			if m.quickSearch != "" {
				m.quickSearch = ""
				m.quickItemCur = 0
				m.quickItemScroll = 0
			} else {
				m.showQuickPrompt = false
			}
		default:
			if len(k) == 1 && k >= " " {
				m.quickSearch += k
				m.quickItemCur = 0
				m.quickItemScroll = 0
			}
		}
		return m, nil
	}

	// Quick launch overlay
	if m.showQuickLaunch {
		tools := launcherTools()
		if m.quickLaunchPickHarness {
			// harness picker mode
			switch k {
			case "j", "down":
				if m.quickLaunchHarnessCur < len(tools)-1 {
					m.quickLaunchHarnessCur++
				}
			case "k", "up":
				if m.quickLaunchHarnessCur > 0 {
					m.quickLaunchHarnessCur--
				}
			case "enter":
				if m.quickLaunchHarnessCur < len(tools) {
					m.quickLaunchHarness = tools[m.quickLaunchHarnessCur]
					m.quickLaunchPickHarness = false
				}
			case "q", "esc", "ctrl+c":
				m.showQuickLaunch = false
			}
		} else {
			// prompt input mode
			switch k {
			case "enter":
				name := m.quickLaunchName
				prompt := strings.TrimSpace(m.quickLaunchPrompt)
				if m.quickLaunchKind == "taffy" && prompt == "" {
					return m, m.setStatus("add feature requirements before launching a taffy workflow", true)
				}
				// Write RUNNING state immediately so [P] reflects it before the agent responds
				writeQuickLaunchState(name, prompt)
				cmd := buildQuickPromptCmd(name, prompt, m.quickLaunchKind)
				m.showQuickLaunch = false
				target := buildLaunchCmd(m.quickLaunchHarness, cmd, m.pinnedSessions)
				return m, trySpawnCmdForHarness(m.quickLaunchHarness, target)
			case "esc":
				m.quickLaunchPickHarness = true
			case "ctrl+c":
				m.showQuickLaunch = false
			case "backspace":
				if len(m.quickLaunchPrompt) > 0 {
					_, size := utf8.DecodeLastRuneInString(m.quickLaunchPrompt)
					m.quickLaunchPrompt = m.quickLaunchPrompt[:len(m.quickLaunchPrompt)-size]
				}
			default:
				if len(k) == 1 {
					m.quickLaunchPrompt += k
				}
			}
		}
		return m, nil
	}

	// Pipeline status overlay
	if m.showPipeline {
		switch k {
		case "a":
			if m.approvalPending != "" {
				_ = os.Remove(".claude/state/approval-pending.txt")
				m.approvalPending = ""
				ps, _ := loadPipelineState()
				m.pipelineState = ps
				n := notifyAllPanesContinue()
				msg := "✓ approved — pipeline resuming"
				if n > 0 {
					msg = fmt.Sprintf("✓ approved — sent 'continue' to %d pane(s)", n)
				}
				return m, m.setStatus(msg, false)
			}
		case "v":
			return m, designPortalCmd(true, false, m.approvalPending)
		case "c":
			_ = os.Remove(".claude/state/maple.json")
			_ = os.Remove(".claude/state/approval-pending.txt")
			m.pipelineState = pipelineState{}
			m.approvalPending = ""
			m.portalAutoStage = ""
			m.showPipeline = false
			return m, m.setStatus("✓ pipeline state cleared", false)
		default:
			m.showPipeline = false
		}
		return m, nil
	}

	// Session launcher overlay
	if m.showLauncher {
		tools := launcherTools()
		switch {
		case m.launcherInput:
			switch k {
			case "enter":
				if len(tools) > 0 && m.launcherCur < len(tools) {
					tool := tools[m.launcherCur]
					m.showLauncher = false
					cmd := buildLaunchCmd(tool, m.launcherCmd, m.pinnedSessions)
					return m, trySpawnCmdForHarness(tool, cmd)
				}
			case "esc":
				m.launcherInput = false
			case "backspace":
				if len(m.launcherCmd) > 0 {
					_, size := utf8.DecodeLastRuneInString(m.launcherCmd)
					m.launcherCmd = m.launcherCmd[:len(m.launcherCmd)-size]
				}
			default:
				if len(k) == 1 {
					m.launcherCmd += k
				}
			}
		default:
			switch k {
			case "j", "down":
				if m.launcherCur < len(tools)-1 {
					m.launcherCur++
				}
			case "k", "up":
				if m.launcherCur > 0 {
					m.launcherCur--
				}
			case "enter":
				m.launcherInput = true
			case "q", "esc", "ctrl+c":
				m.showLauncher = false
			}
		}
		return m, nil
	}

	// RTK harness selector overlay
	if m.showRTKHarness {
		if m.rtkHarnessRunning {
			return m, nil
		}
		switch k {
		case "j", "down":
			if m.rtkHarnessCur < len(allRTKHarnesses)-1 {
				m.rtkHarnessCur++
			}
		case "k", "up":
			if m.rtkHarnessCur > 0 {
				m.rtkHarnessCur--
			}
		case " ":
			h := allRTKHarnesses[m.rtkHarnessCur]
			if !m.rtkHarnessInstalled[h.key] {
				m.rtkHarnessToggled[h.key] = !m.rtkHarnessToggled[h.key]
			}
		case "enter":
			// run rtk init for each toggled harness
			var cmds []tea.Cmd
			for _, h := range allRTKHarnesses {
				if m.rtkHarnessToggled[h.key] {
					h := h // capture
					cmds = append(cmds, rtkInitCmd(h))
				}
			}
			if len(cmds) == 0 {
				m.showRTKHarness = false
			} else {
				m.rtkHarnessRunning = true
				return m, tea.Batch(cmds...)
			}
		case "q", "esc", "ctrl+c":
			m.showRTKHarness = false
		}
		return m, nil
	}

	// Manual launch modal — shown when spawnInNewTerminal fails
	if m.showManualLaunch {
		switch k {
		case "c":
			if len(m.manualLaunchArgs) > 0 {
				// build a shell-pasteable command and copy to clipboard via pbcopy/xclip/clip
				var quoted []string
				for _, a := range m.manualLaunchArgs {
					quoted = append(quoted, shQuote(a))
				}
				line := strings.Join(quoted, " ")
				_ = copyToClipboard(line)
				m.manualLaunchCopied = true
			}
		case "q", "esc", "ctrl+c", "enter":
			m.showManualLaunch = false
		}
		return m, nil
	}

	// ShipSafe audit overlay
	if m.showShipSafe {
		maxScroll := len(m.shipSafeLines) - (m.height - 15)
		if maxScroll < 0 {
			maxScroll = 0
		}
		switch k {
		case "j", "down":
			if m.shipSafeScroll < maxScroll {
				m.shipSafeScroll++
			}
		case "k", "up":
			if m.shipSafeScroll > 0 {
				m.shipSafeScroll--
			}
		case "g":
			m.shipSafeScroll = 0
		case "G":
			m.shipSafeScroll = maxScroll
		case "q", "esc", "b", "ctrl+c":
			if !m.shipSafeRunning {
				m.showShipSafe = false
			}
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
			_ = exec.Command("gh", "pr", "view", fmt.Sprintf("%d", m.prDetailNumber), "--web").Start()
		case "a":
			m.prDetailLoading = true
			return m, approvePRCmd(m.prDetailNumber)
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
		// pin selected session when focus is on Agents pane; otherwise switch to PRs
		if m.focus == paneAgents && m.sessionsCur < len(m.sessions) {
			s := m.sessions[m.sessionsCur]
			id := sessionUUID(s)
			if id == "" {
				return m, m.setStatus("✗ cannot pin: no resumable ID for "+s.source, true)
			}
			savePinnedSession(s.source, id)
			m.pinnedSessions = loadPinnedSessions()
			return m, m.setStatus("✓ pinned "+s.source+" — saved to .claude/state/sessions.json", false)
		}
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
		} else if m.focus == paneAgents && m.sessionsCur < len(m.sessions) {
			m.openSessionDetail(m.sessions[m.sessionsCur])
		} else if m.focus == paneQA && m.qaEntryCur < len(m.qaEntries) {
			m.openQAFile(m.qaEntries[m.qaEntryCur])
		} else if m.focus == panePRs && m.prsCur < len(m.prList) {
			pr := m.prList[m.prsCur]
			m.prDetailLoading = true
			m.prDetailLines = nil
			m.prDetailNumber = pr.number
			m.showPRDetail = false
			return m, loadPRDetailCmd(pr.number, pr.title)
		}
	case "S":
		return m, m.runShipSafeCmd()
	case "x":
		m.taffyItems = loadTaffyItems()
		m.quickItems = loadQuickItems()
		m.quickMode = "taffy"
		m.quickItemCur = 0
		m.quickItemScroll = 0
		m.quickSearch = ""
		m.showQuickPrompt = true
	case "P":
		// pipelineState and approvalPending are kept fresh by reload() every tick
		m.showPipeline = true
	case "L":
		m.launcherCur = 0
		m.launcherCmd = ""
		m.launcherInput = false
		m.showLauncher = true
	case "R":
		m.rtkHarnessInstalled = loadRTKHarnesses()
		m.rtkHarnessToggled = map[string]bool{}
		m.rtkHarnessCur = 0
		m.rtkHarnessRunning = false
		m.showRTKHarness = true
	case "o":
		if m.focus == paneAgents {
			if m.sessionsCur >= len(m.sessions) {
				return m, m.setStatus("no sessions — navigate to [a] Agents pane and select one", true)
			}
			s := m.sessions[m.sessionsCur]
			cmd := agentOpenCmd(s)
			if len(cmd) == 0 {
				return m, m.setStatus("✗ cannot open: no launch command for "+s.source, true)
			}
			// auto-pin: use s.id directly — it's the UUID for claude and the DB id for opencode
			if s.id != "" {
				savePinnedSession(s.source, sessionUUID(s))
				m.pinnedSessions = loadPinnedSessions()
			}
			return m, trySpawnCmdForHarness(s.source, cmd)
		}
		if m.focus == panePRs && m.prsCur < len(m.prList) {
			_ = exec.Command("gh", "pr", "view", fmt.Sprintf("%d", m.prList[m.prsCur].number), "--web").Start()
			return m, m.setStatus("opening PR in browser…", false)
		}
		if m.focus != paneAgents {
			return m, m.setStatus("press [a] to switch to Agents pane, then [o] to open a session", false)
		}
	case "r":
		if m.focus == paneQA && m.qaEntryCur < len(m.qaEntries) {
			return m, m.runTestCmd(m.qaEntries[m.qaEntryCur])
		} else if m.focus == paneQA {
			return m, m.setStatus("no tests found", true)
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
			if m.sessionsCur < len(m.sessions)-1 {
				m.sessionsCur++
			}
		case panePRs:
			if m.prsCur < len(m.prList)-1 {
				m.prsCur++
			}
		case paneQA:
			if m.qaEntryCur < len(m.qaEntries)-1 {
				m.qaEntryCur++
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
			if m.sessionsCur > 0 {
				m.sessionsCur--
			}
		case panePRs:
			if m.prsCur > 0 {
				m.prsCur--
			}
		case paneQA:
			if m.qaEntryCur > 0 {
				m.qaEntryCur--
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
			m.sessionsCur = 0
		case panePRs:
			m.prsCur = 0
		case paneQA:
			m.qaEntryCur = 0
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
			if len(m.sessions) > 0 {
				m.sessionsCur = len(m.sessions) - 1
			}
		case panePRs:
			if len(m.prList) > 0 {
				m.prsCur = len(m.prList) - 1
			}
		case paneQA:
			if len(m.qaEntries) > 0 {
				m.qaEntryCur = len(m.qaEntries) - 1
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

// trySpawnCmd tries to open args in a new terminal tab/window as a tea.Cmd.
// On success it returns spawnSucceededMsg; on failure it returns spawnFailedMsg
// so the TUI can show the manual-launch modal — maple never exits.
func trySpawnCmd(args []string) tea.Cmd {
	return trySpawnCmdForHarness("", args)
}

// trySpawnCmdForHarness spawns args and records a paneRef keyed by harness so
// approvals can later send "continue" back to the running agent. Harness may be
// "" for launches that should not be tracked (e.g. story editors).
func trySpawnCmdForHarness(harness string, args []string) tea.Cmd {
	return func() tea.Msg {
		if len(args) == 0 {
			return spawnFailedMsg{}
		}
		label := harness
		if label == "" {
			label = args[0]
		}
		if harness != "" {
			p, err := spawnWithPane(harness, args)
			if err != nil {
				return spawnFailedMsg{args: args}
			}
			savePaneRef(harness, p)
			return spawnSucceededMsg{harness: label}
		}
		if err := spawnInNewTerminal(args); err != nil {
			return spawnFailedMsg{args: args}
		}
		return spawnSucceededMsg{harness: label}
	}
}

// copyToClipboard writes s to the system clipboard via pbcopy (macOS),
// xclip/xsel (Linux), or clip (Windows). Errors are silently ignored.
func copyToClipboard(s string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	default:
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard tool found")
		}
	}
	cmd.Stdin = strings.NewReader(s)
	return cmd.Run()
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
	if m.showShipSafe {
		return m.header() + m.shipSafeView() + m.footer()
	}
	if m.showSession {
		return m.header() + m.sessionDetailView() + m.footer()
	}
	if m.showStory {
		return m.header() + m.storyDetailView() + m.footer()
	}
	if m.showTestOut {
		return m.header() + m.testOutputView() + m.footer()
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
	if m.showQuickPrompt {
		return m.header() + m.quickPromptView() + m.footer()
	}
	if m.showQuickLaunch {
		return m.header() + m.quickLaunchView() + m.footer()
	}
	if m.showPipeline {
		return m.header() + m.pipelineStatusView() + m.footer()
	}
	if m.showLauncher {
		return m.header() + m.launcherView() + m.footer()
	}
	if m.showRTKHarness {
		return m.header() + m.rtkHarnessView() + m.footer()
	}
	if m.showManualLaunch {
		return m.header() + m.manualLaunchView() + m.footer()
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

	// Count Gherkin specs in docs/stories/ and Taffy workflows in .*/taffy/
	gherkinCount := 0
	if entries, err := os.ReadDir("docs/stories"); err == nil {
		for _, e := range entries {
			name := e.Name()
			// Skip template and special files
			if !e.IsDir() && strings.HasSuffix(name, ".md") && name != "_template.md" && !strings.HasPrefix(name, ".") {
				gherkinCount++
			}
		}
	}
	taffyCount := 0
	for _, harness := range []string{".claude", ".cursor", ".opencode"} {
		if entries, err := os.ReadDir(filepath.Join(harness, "taffy")); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".yml") {
					taffyCount++
				}
			}
		}
	}

	badges := lipgloss.NewStyle().Foreground(t.Accent).Render(fmt.Sprintf("  📋 Gherkin: %d | ▶️ Taffy: %d", gherkinCount, taffyCount))
	return logoCompact(t.Primary) + info + "\n" + badges + "\n"
}

func (m *dashboardModel) footer() string {
	t := m.theme
	// Status always wins — errors from approve/run must be visible over overlay hints
	if m.status != "" {
		col := t.Success
		if m.statusErr {
			col = t.Error
		}
		return "\n" + lipgloss.NewStyle().Foreground(col).Render("  "+m.status) + "\n"
	}

	keys := "  [Tab] cycle · [s/a/p/Q] pane · [Enter] open · [o] open+pin session · [p] pin · [n] story · [L] launch · [R] rtk harnesses · [S] ship-safe · [x] quick prompt · [P] pipeline · [F] skills · [?] help · [q] quit"
	switch {
	case m.showManualLaunch:
		if m.manualLaunchCopied {
			keys = "  copied! · [Esc] dismiss"
		} else {
			keys = "  [c] copy command · [Esc] dismiss"
		}
	case m.showRTKHarness:
		if m.rtkHarnessRunning {
			keys = "  installing…"
		} else {
			keys = "  [j/k] navigate · [Space] toggle · [Enter] install selected · [Esc] close"
		}
	case m.showLauncher:
		if m.launcherInput {
			keys = "  type command · [Enter] launch · [Esc] back"
		} else {
			keys = "  [j/k] navigate · [Enter] enter command · [Esc] close"
		}
	case m.showPipeline:
		switch {
		case m.approvalPending != "":
			keys = "  [a] approve stage · [v] open design portal · [c] clear state · any other key closes"
		case m.pipelineState.isStale():
			keys = "  [v] open design portal · [c] clear stale state · any other key closes"
		default:
			keys = "  [v] open design portal · [c] clear state · any other key closes"
		}
	case m.showQuickLaunch:
		if m.quickLaunchPickHarness {
			keys = "  [j/k] navigate · [Enter] select harness · [Esc] back"
		} else {
			keys = "  type context · [Enter] launch · [Esc] back"
		}
	case m.showQuickPrompt:
		filtered := quickFilter(m.quickItems, m.quickSearch)
		if m.quickSearch != "" {
			keys = fmt.Sprintf("  search: %s  · %d match(es) · [j/k] navigate · [Enter] select · [Esc] clear · [Backspace] edit", m.quickSearch, len(filtered))
		} else {
			keys = "  [j/k] navigate · [Enter] select · type to search · [Esc] close"
		}
	case m.showSkills:
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
	case m.showShipSafe:
		if m.shipSafeRunning {
			keys = "  auditing…"
		} else if m.shipSafeFailed {
			keys = "  ISSUES FOUND · [j/k] scroll · [Esc] close"
		} else {
			keys = "  CLEAN · [j/k] scroll · [Esc] close"
		}
	case m.showSession:
		keys = "  [j/k] scroll · [Esc] close"
	case m.showStory:
		keys = "  [j/k] scroll · [e] re-edit · [Esc] close"
	case m.showTestOut:
		if m.testOutRunning {
			keys = "  running…"
		} else if m.testOutFailed {
			keys = "  FAILED · [j/k] scroll · [Esc] close"
		} else {
			keys = "  PASSED · [j/k] scroll · [Esc] close"
		}
	case m.showQAFile:
		keys = "  [j/k] scroll · [r] run test · [Esc] close"
	case m.showPRDetail || m.prDetailLoading:
		keys = "  [j/k] scroll · [o] open in browser · [a] approve · [Esc] close"
	case m.searchMode:
		keys = "  /" + m.searchBuf + "█"
	case m.cmdMode:
		keys = "  :" + m.cmdBuf + "█"
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
		lipgloss.NewStyle().Foreground(t.Muted).Render(fmt.Sprintf("  %d stories · %d PRs · %d tests",
			len(m.stories), len(m.prList), len(m.qaEntries))),
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

// openQAFile loads a test file into the QA file overlay.
func (m *dashboardModel) openQAFile(e testEntry) {
	raw, err := os.ReadFile(e.path)
	if err != nil {
		m.status = "✗ could not read " + e.path
		m.statusErr = true
		return
	}
	m.qaFileLines = strings.Split(string(raw), "\n")
	m.qaFileScroll = 0
	m.qaFileTitle = e.path
	m.qaFileFramework = e.framework
	m.qaFileRunCmd = e.runCmd
	m.showQAFile = true
}

// runTestCmd runs a test entry and streams output to the test output overlay.
func (m *dashboardModel) runTestCmd(e testEntry) tea.Cmd {
	title := e.framework + ": " + e.path
	return tea.Batch(
		func() tea.Msg { return testRunStartMsg{title: title} },
		func() tea.Msg {
			if len(e.runCmd) == 0 {
				return testRunDoneMsg{lines: []string{"no run command configured"}, failed: true}
			}
			cmd := exec.Command(e.runCmd[0], e.runCmd[1:]...)
			out, err := cmd.CombinedOutput()
			lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
			if len(lines) == 0 {
				lines = []string{"(no output)"}
			}
			return testRunDoneMsg{lines: lines, failed: err != nil}
		},
	)
}

// runShipSafeCmd runs npx ship-safe audit . and streams output to the overlay.
func (m *dashboardModel) runShipSafeCmd() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return shipSafeStartMsg{} },
		func() tea.Msg {
			npx, err := exec.LookPath("npx")
			if err != nil {
				return shipSafeDoneMsg{lines: []string{"npx not found — install Node.js from nodejs.org"}, failed: true}
			}
			cmd := exec.Command(npx, "ship-safe", "audit", ".")
			out, err := cmd.CombinedOutput()
			lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
			if len(lines) == 0 {
				lines = []string{"(no output)"}
			}
			return shipSafeDoneMsg{lines: lines, failed: err != nil}
		},
	)
}

// agentOpenCmd returns the CLI command to resume a session in its native agent.
func agentOpenCmd(s sessionRow) []string {
	switch s.source {
	case "claude":
		uuid := sessionUUID(s)
		if uuid == "" {
			return nil
		}
		return []string{"claude", "--resume", uuid}
	case "opencode":
		// opencode accepts --session <id> to resume; fall back to plain opencode if no id
		if s.id != "" {
			return []string{"opencode", "--session", s.id}
		}
		return []string{"opencode"}
	case "copilot":
		if s.id != "" {
			return []string{"copilot", "--resume=" + s.id}
		}
		return []string{"copilot"}
	default:
		return nil
	}
}

// sessionUUID extracts the resumable session identifier for any source.
// For claude: strips the directory path and .jsonl extension from s.id (which is a file path).
// For opencode: s.id is already the DB session ID — return it directly.
func sessionUUID(s sessionRow) string {
	switch s.source {
	case "claude":
		base := s.id
		if idx := strings.LastIndex(base, "/"); idx >= 0 {
			base = base[idx+1:]
		}
		return strings.TrimSuffix(base, ".jsonl")
	case "opencode", "copilot":
		return s.id
	default:
		return s.id
	}
}

// launcherTools returns a list of tool names available for launching.
func launcherTools() []string {
	var tools []string
	if p, _ := exec.LookPath("claude"); p != "" {
		tools = append(tools, "claude")
	}
	if p, _ := exec.LookPath("opencode"); p != "" {
		tools = append(tools, "opencode")
	}
	if p, _ := exec.LookPath("copilot"); p != "" {
		tools = append(tools, "copilot")
	}
	if p, _ := exec.LookPath("cursor-agent"); p != "" {
		tools = append(tools, "cursor")
	} else if p, _ := exec.LookPath("cursor"); p != "" {
		tools = append(tools, "cursor")
	}
	if len(tools) == 0 {
		tools = append(tools, "claude") // show as option even if not detected
	}
	return tools
}

// buildLaunchCmd constructs the exec command for launching a tool, resuming a pinned
// session if one exists, otherwise starting fresh with the given command/prompt.
// When rtk is on PATH, RTK_HOOK_AUDIT=1 is prepended so hook activity is logged.
func buildLaunchCmd(tool, cmd string, pinned map[string]string) []string {
	pinnedID := pinned[tool]
	var args []string
	switch tool {
	case "claude":
		if pinnedID != "" {
			args = []string{"claude", "--resume", pinnedID}
			if cmd != "" {
				args = append(args, cmd)
			}
		} else if cmd != "" {
			args = []string{"claude", cmd}
		} else {
			args = []string{"claude"}
		}
	case "opencode":
		if cmd != "" {
			args = []string{"opencode", cmd}
		} else {
			args = []string{"opencode"}
		}
	case "copilot":
		if pinnedID != "" {
			args = []string{"copilot", "--resume=" + pinnedID}
			if cmd != "" {
				args = append(args, "-i", cmd)
			}
		} else if cmd != "" {
			args = []string{"copilot", "-i", cmd}
		} else {
			args = []string{"copilot"}
		}
	case "cursor":
		cursorBin := "cursor-agent"
		if p, _ := exec.LookPath("cursor-agent"); p == "" {
			cursorBin = "cursor"
		}
		if cmd != "" {
			args = []string{cursorBin, cmd}
		} else {
			args = []string{cursorBin}
		}
	default:
		args = []string{tool}
	}
	if rtkPath, err := exec.LookPath("rtk"); err == nil && rtkPath != "" {
		args = append([]string{"env", "RTK_HOOK_AUDIT=1"}, args...)
	}
	return args
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

// rtkInitCmd runs rtk init with the harness-specific flags and reports back.
func rtkInitCmd(h rtkHarness) tea.Cmd {
	return func() tea.Msg {
		rtkPath, err := exec.LookPath("rtk")
		if err != nil {
			return rtkInitDoneMsg{key: h.key, err: "rtk not found"}
		}
		out, err := exec.Command(rtkPath, h.flags...).CombinedOutput()
		if err != nil {
			return rtkInitDoneMsg{key: h.key, err: strings.TrimSpace(string(out))}
		}
		return rtkInitDoneMsg{key: h.key}
	}
}

func designPortalCmd(open, auto bool, stage string) tea.Cmd {
	return func() tea.Msg {
		script := "scripts/design-review-portal.sh"
		if _, err := os.Stat(script); err != nil {
			return designPortalResultMsg{err: "scripts/design-review-portal.sh not found", open: open, auto: auto, stage: stage}
		}
		action := "start"
		if open {
			action = "open"
		}
		cmd := exec.Command("bash", script, action)
		out, err := cmd.CombinedOutput()
		if err != nil {
			msg := strings.TrimSpace(string(out))
			if msg == "" {
				msg = err.Error()
			}
			return designPortalResultMsg{err: msg, open: open, auto: auto, stage: stage}
		}
		return designPortalResultMsg{open: open, auto: auto, stage: stage}
	}
}

// ─── Entry point ─────────────────────────────────────────────────────────────

// runDashboard runs the dashboard and returns the exit action and any open
// target command (used when dashActionOpenAgent is returned).
func runDashboard(t Theme, noAnimate bool) (dashAction, []string, error) {
	m := newDashboard(t, noAnimate)
	writeRecoveryMarker("running")
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	writeRecoveryMarker("exited")
	if err != nil {
		return dashActionNone, nil, err
	}
	dm := final.(*dashboardModel)
	return dm.exitAction, dm.openTarget, nil
}

// writeQuickLaunchState writes an initial RUNNING pipeline state to maple.json
// when the user launches from the quick-prompt overlay. Merges with existing content.
func writeQuickLaunchState(skill, stage string) {
	_ = os.MkdirAll(".claude/state", 0o755)
	merged := map[string]interface{}{}
	if raw, err := os.ReadFile(".claude/state/maple.json"); err == nil {
		_ = json.Unmarshal(raw, &merged)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	merged["taffy"] = skill
	merged["stage"] = stage
	merged["status"] = "RUNNING"
	merged["started_at"] = now
	merged["updated_at"] = now
	data, _ := json.Marshal(merged)
	_ = os.WriteFile(".claude/state/maple.json", append(data, '\n'), 0o644)
}

// buildQuickPromptCmd wraps the skill invocation with silent control instructions
// so the launched harness keeps maple.json updated and honors user-supplied context.
func buildQuickPromptCmd(skill, userPrompt, kind string) string {
	cmd := "/" + skill
	if userPrompt != "" {
		cmd += " " + userPrompt
	}
	taffyContext := ""
	if kind == "taffy" {
		taffyContext = `

<maple-taffy-context>
Treat the text after /` + skill + ` as the feature requirements for this workflow.
Do not replace it with canned or default prompts. For new-ui-feature, pass these
requirements into spec-kit so the generated story contains concrete Gherkin
scenarios for this request.
</maple-taffy-context>`
	}
	tracking := `

<maple-pipeline>
You were launched from the MAPLE quick-prompt. Keep .claude/state/maple.json updated as you work by writing (merge, never overwrite other keys):
  {"taffy":"` + skill + `","stage":"<current step>","status":"RUNNING","updated_at":"<ISO-8601 timestamp>"}
Set status to "DONE" when finished, "FAILED" if you cannot complete.
</maple-pipeline>`
	return cmd + taffyContext + tracking
}

func writeRecoveryMarker(state string) {
	_ = os.MkdirAll(".claude/state", 0o755)
	// Merge with existing content so taffy pipeline fields written by the
	// pipeline-runner skill are not overwritten when the TUI starts or exits.
	merged := map[string]interface{}{}
	if raw, err := os.ReadFile(".claude/state/maple.json"); err == nil {
		_ = json.Unmarshal(raw, &merged)
	}
	merged["state"] = state
	merged["ts"] = time.Now().UTC().Format(time.RFC3339)
	data, _ := json.Marshal(merged)
	_ = os.WriteFile(".claude/state/maple.json", append(data, '\n'), 0o644)
}
