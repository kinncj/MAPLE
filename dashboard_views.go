package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ─── Pane content renderers ───────────────────────────────────────────────────

func (m *dashboardModel) storiesContent(height int) string {
	t := m.theme
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("Stories")
	if len(m.stories) == 0 {
		return title + "\n" + lipgloss.NewStyle().Foreground(t.Muted).Render("no stories yet — run maple req")
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
		return title + "\n" + lipgloss.NewStyle().Foreground(t.Muted).Render("no agent activity for this project")
	}
	lines := []string{title}
	cursor := lipgloss.NewStyle().Foreground(t.Accent).Render("▸")
	for i, a := range m.agents {
		if i >= height-2 {
			break
		}
		badge := agentSourceBadge(a.source, t)
		name := lipgloss.NewStyle().Foreground(t.Foreground).Render(truncate(a.agent, 10))
		op := lipgloss.NewStyle().Foreground(t.Muted).Render(truncate(a.op, 18))
		var line string
		if i == m.agentsCur && m.focus == paneAgents {
			line = cursor + " " + badge + " " + name + " " + op
		} else {
			line = "  " + badge + " " + name + " " + op
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func agentSourceBadge(source string, t Theme) string {
	switch source {
	case "claude":
		return lipgloss.NewStyle().Foreground(t.Primary).Render("[cc]")
	case "opencode":
		return lipgloss.NewStyle().Foreground(t.Accent).Render("[oc]")
	case "maple":
		return lipgloss.NewStyle().Foreground(t.Success).Render("[ml]")
	default:
		return lipgloss.NewStyle().Foreground(t.Muted).Render("[??]")
	}
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

func (m *dashboardModel) qaContent(height int) string {
	t := m.theme
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("QA / Tests")
	if len(m.qaEntries) == 0 {
		return title + "\n" + lipgloss.NewStyle().Foreground(t.Muted).Render("no tests found")
	}
	summary := lipgloss.NewStyle().Foreground(t.Muted).Render(
		fmt.Sprintf("  %d test file(s)", len(m.qaEntries)))
	lines := []string{title, summary}
	cursor := lipgloss.NewStyle().Foreground(t.Accent).Render("▸")
	for i, e := range m.qaEntries {
		if i >= height-3 {
			break
		}
		badge := testFrameworkBadge(e.framework, t)
		label := lipgloss.NewStyle().Foreground(t.Foreground).Render(truncate(filepath.Base(e.path), 24))
		extra := ""
		if e.count > 0 {
			extra = lipgloss.NewStyle().Foreground(t.Muted).Render(fmt.Sprintf(" (%d)", e.count))
		}
		var line string
		if i == m.qaEntryCur && m.focus == paneQA {
			line = cursor + " " + badge + " " + label + extra
		} else {
			line = "  " + badge + " " + label + extra
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func testFrameworkBadge(fw string, t Theme) string {
	switch fw {
	case "gherkin":
		return lipgloss.NewStyle().Foreground(t.Success).Render("[gherkin]")
	case "go":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Render("[go]    ")
	case "jest":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#C21325")).Render("[jest]  ")
	case "vitest":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FCC72B")).Render("[vitest]")
	case "mocha":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#8D6748")).Render("[mocha] ")
	case "npm":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#CB3837")).Render("[npm]   ")
	case "pytest":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3572A5")).Render("[pytest]")
	case "unittest":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3572A5")).Render("[py]    ")
	case "rspec":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#CC342D")).Render("[rspec] ")
	case "maven":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#C71A36")).Render("[maven] ")
	case "gradle":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#02303A")).Render("[gradle]")
	case "phpunit":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#777BB4")).Render("[php]   ")
	case "cargo":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#DEA584")).Render("[cargo] ")
	default:
		return lipgloss.NewStyle().Foreground(t.Muted).Render("[test]  ")
	}
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

func (m *dashboardModel) skillsBrowserView() string {
	t := m.theme

	// Tab headers
	installedStyle := lipgloss.NewStyle().Foreground(t.Muted)
	searchStyle := lipgloss.NewStyle().Foreground(t.Muted)
	if !m.skillsTabSearch {
		installedStyle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Underline(true)
	} else {
		searchStyle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Underline(true)
	}
	tabs := "  " + installedStyle.Render("Installed") + "  " + searchStyle.Render("Search")
	sep := lipgloss.NewStyle().Foreground(t.Muted).Render("  " + strings.Repeat("─", 62))
	lines := []string{tabs, sep}

	if m.npxPath == "" {
		lines = append(lines,
			"",
			lipgloss.NewStyle().Foreground(t.Warning).Render("  npx not found — install Node.js from nodejs.org"),
		)
		return strings.Join(lines, "\n")
	}

	if !m.skillsTabSearch {
		// ── Installed tab ─────────────────────────────────────────
		if m.installedLoading {
			lines = append(lines, "", lipgloss.NewStyle().Foreground(t.Muted).Render("  loading installed skills…"))
		} else if m.installedErr != "" {
			lines = append(lines, "", lipgloss.NewStyle().Foreground(t.Error).Render("  "+m.installedErr))
		} else if len(m.installedSkills) == 0 {
			lines = append(lines, "", lipgloss.NewStyle().Foreground(t.Muted).Render("  no skills installed — switch to Search tab to add some"))
		} else {
			cursor := lipgloss.NewStyle().Foreground(t.Accent).Render("▸")
			for i, sk := range m.installedSkills {
				name := lipgloss.NewStyle().Foreground(t.Foreground).Bold(true).Render(fmt.Sprintf("%-24s", sk.name))
				pkg := lipgloss.NewStyle().Foreground(t.Muted).Render(fmt.Sprintf("%-28s", sk.pkg))
				scope := lipgloss.NewStyle().Foreground(t.Muted).Render(sk.scope)
				if i == m.installedCur {
					lines = append(lines, "  "+cursor+" "+name+" "+pkg+" "+scope)
				} else {
					lines = append(lines, "      "+name+" "+pkg+" "+scope)
				}
			}
		}
	} else {
		// ── Search tab ────────────────────────────────────────────
		search := lipgloss.NewStyle().Foreground(t.Muted).Render("  search: ") +
			lipgloss.NewStyle().Foreground(t.Foreground).Render(m.skillsQuery+"█")
		lines = append(lines, search)

		if m.skillsLoading {
			lines = append(lines, "", lipgloss.NewStyle().Foreground(t.Muted).Render("  searching…"))
		} else if m.skillsErr != "" {
			lines = append(lines, "", lipgloss.NewStyle().Foreground(t.Error).Render("  "+m.skillsErr))
		} else if len(m.skillsItems) == 0 && m.skillsSearched {
			lines = append(lines, "", lipgloss.NewStyle().Foreground(t.Muted).Render("  no results — try a different query"))
		} else if len(m.skillsItems) == 0 {
			lines = append(lines, "", lipgloss.NewStyle().Foreground(t.Muted).Render("  type a query and press Enter to search"))
		} else {
			cursor := lipgloss.NewStyle().Foreground(t.Accent).Render("▸")
			for i, sk := range m.skillsItems {
				pkg := lipgloss.NewStyle().Foreground(t.Foreground).Bold(true).Render(fmt.Sprintf("%-42s", sk.pkg))
				installs := lipgloss.NewStyle().Foreground(t.Muted).Render(sk.installs + " installs")
				if i == m.skillsCur {
					lines = append(lines, "  "+cursor+" "+pkg+" "+installs)
					if sk.url != "" {
						lines = append(lines, "       "+lipgloss.NewStyle().Foreground(t.Muted).Render(sk.url))
					}
				} else {
					lines = append(lines, "      "+pkg+" "+installs)
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// storyDetailView renders the selected Story.md as a full-screen overlay.
func (m *dashboardModel) storyDetailView() string {
	t := m.theme
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("  " + m.storyTitle)
	dir := lipgloss.NewStyle().Foreground(t.Muted).Render("  " + m.storyDir)
	sep := lipgloss.NewStyle().Foreground(t.Muted).Render("  " + strings.Repeat("─", 62))

	visible := m.height - 14
	if visible < 4 {
		visible = 4
	}
	end := m.storyScroll + visible
	if end > len(m.storyLines) {
		end = len(m.storyLines)
	}
	window := m.storyLines[m.storyScroll:end]

	var sb strings.Builder
	sb.WriteString(title + "\n" + dir + "\n" + sep + "\n\n")
	for _, l := range window {
		sb.WriteString("  " + colorizeStoryLine(l, t) + "\n")
	}

	total := len(m.storyLines)
	if total > visible {
		pct := (m.storyScroll * 100) / (total - visible)
		sb.WriteString(fmt.Sprintf("\n  %s\n",
			lipgloss.NewStyle().Foreground(t.Muted).Render(
				fmt.Sprintf("(%d%%)  j/k scroll · e re-edit · Esc close", pct))))
	} else {
		sb.WriteString("\n  " + lipgloss.NewStyle().Foreground(t.Muted).Render("e re-edit · Esc close") + "\n")
	}
	return sb.String()
}

// colorizeStoryLine applies minimal markdown-aware colours to a Story.md line.
func colorizeStoryLine(line string, t Theme) string {
	trimmed := strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(trimmed, "# "):
		return lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render(line)
	case strings.HasPrefix(trimmed, "## "):
		return lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render(line)
	case strings.HasPrefix(trimmed, "### "):
		return lipgloss.NewStyle().Foreground(t.Warning).Render(line)
	case strings.HasPrefix(trimmed, "- [ ]"):
		check := lipgloss.NewStyle().Foreground(t.Muted).Render("- [ ]")
		rest := lipgloss.NewStyle().Foreground(t.Foreground).Render(line[strings.Index(line, "- [ ]")+5:])
		return check + rest
	case strings.HasPrefix(trimmed, "- [x]"), strings.HasPrefix(trimmed, "- [X]"):
		check := lipgloss.NewStyle().Foreground(t.Success).Render("- [x]")
		rest := lipgloss.NewStyle().Foreground(t.Muted).Render(line[strings.Index(line, "- [")+5:])
		return check + rest
	case strings.HasPrefix(trimmed, "```"):
		return lipgloss.NewStyle().Foreground(t.Muted).Render(line)
	case strings.HasPrefix(trimmed, "Feature:"),
		strings.HasPrefix(trimmed, "Scenario:"),
		strings.HasPrefix(trimmed, "Given "),
		strings.HasPrefix(trimmed, "When "),
		strings.HasPrefix(trimmed, "Then "),
		strings.HasPrefix(trimmed, "And "),
		strings.HasPrefix(trimmed, "But "),
		strings.HasPrefix(trimmed, "@"):
		return colorizeGherkin(line, t)
	default:
		return lipgloss.NewStyle().Foreground(t.Foreground).Render(line)
	}
}

func (m *dashboardModel) helpView() string {
	t := m.theme
	titleStyle := lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
	sep := lipgloss.NewStyle().Foreground(t.Muted).Render(strings.Repeat("─", 62))

	keyBindings := [][2]string{
		{"Tab / Shift+Tab", "cycle panes"},
		{"j / k  (↓ / ↑)", "navigate rows"},
		{"gg / G", "jump to top / bottom"},
		{"s  a  p  Q", "focus Stories / Agents / PRs / QA"},
		{"d", "toggle Design pane (full-screen)"},
		{"l", "toggle Logs pane (full-screen)"},
		{"n", "new story → Gherkin requirements wizard"},
		{"u", "update — re-sync template files"},
		{"r", "reload all pane data"},
		{"F", "Skills marketplace (skills.sh)"},
		{"/", "search within active pane"},
		{"?", "this help overlay"},
		{"q  /  Ctrl+C", "quit"},
	}

	cmdRef := [][2]string{
		{":q  :wq  :q!  :x", "quit"},
		{":e  :e!  :r  :reload", "reload data"},
		{":n  :req  :story", "new story wizard"},
		{":u  :update", "re-sync template"},
		{":labels", "bootstrap GitHub labels"},
		{":project", "create GitHub Project v2"},
		{":theme <name>", "switch colour theme"},
		{":colo <name>", "alias for :theme"},
		{":debug", "toggle debug log"},
		{":help  :h  :?", "this overlay"},
		{"", ""},
		{"themes:", "tokyo-night  catppuccin-mocha"},
		{"", "gruvbox  nord  everforest"},
	}

	renderCol := func(rows [][2]string) []string {
		var out []string
		for _, p := range rows {
			if p[0] == "" && p[1] == "" {
				out = append(out, "")
				continue
			}
			key := lipgloss.NewStyle().Foreground(t.Accent).Render(fmt.Sprintf("  %-24s", p[0]))
			val := lipgloss.NewStyle().Foreground(t.Foreground).Render(p[1])
			out = append(out, key+val)
		}
		return out
	}

	leftTitle := titleStyle.Render("  Keybindings")
	rightTitle := titleStyle.Render("  : Commands")
	leftLines := renderCol(keyBindings)
	rightLines := renderCol(cmdRef)

	// pad both columns to same length
	for len(leftLines) < len(rightLines) {
		leftLines = append(leftLines, "")
	}
	for len(rightLines) < len(leftLines) {
		rightLines = append(rightLines, "")
	}

	halfW := (m.width - 4) / 2
	colStyle := lipgloss.NewStyle().Width(halfW)

	var rows []string
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		colStyle.Render(leftTitle), colStyle.Render(rightTitle))
	rows = append(rows, header, sep)
	for i := range leftLines {
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			colStyle.Render(leftLines[i]), colStyle.Render(rightLines[i]))
		rows = append(rows, row)
	}
	rows = append(rows, "", lipgloss.NewStyle().Foreground(t.Muted).Render("  Press any key to close"))
	return strings.Join(rows, "\n")
}

// extractGherkinFromLines pulls the content inside the first ```gherkin ... ``` block.
// Falls back to returning all lines that look like Gherkin keywords.
func extractGherkinFromLines(lines []string) string {
	var inFence bool
	var out []string
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if !inFence && (t == "```gherkin" || t == "```") {
			inFence = true
			continue
		}
		if inFence {
			if t == "```" {
				break
			}
			out = append(out, l)
		}
	}
	if len(out) > 0 {
		return strings.TrimSpace(strings.Join(out, "\n"))
	}
	// No fence found — return any Gherkin-looking lines
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t == "" {
			continue
		}
		for _, kw := range []string{"Feature:", "Background:", "Scenario", "Given ", "When ", "Then ", "And ", "But ", "@"} {
			if strings.HasPrefix(t, kw) {
				out = append(out, l)
				break
			}
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// openStoryDetail reads a Story.md and sets up the overlay.
func (m *dashboardModel) openStoryDetail(s storyRow) {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		m.status = "✗ could not read " + s.path
		m.statusErr = true
		return
	}
	// Strip YAML frontmatter (between --- delimiters)
	content := string(raw)
	if strings.HasPrefix(content, "---") {
		end := strings.Index(content[3:], "\n---")
		if end >= 0 {
			content = strings.TrimSpace(content[3+end+4:])
		}
	}
	m.storyLines = strings.Split(content, "\n")
	m.storyScroll = 0
	m.storyTitle = s.id
	m.storyDir = filepath.Dir(s.path)
	m.showStory = true
}

// openSessionDetail loads session detail lines for the selected agent row and
// scrolls to the line closest to the row's timestamp.
func (m *dashboardModel) openSessionDetail(row agentRow) {
	m.sessionTitle = row.agent + ": " + truncate(row.op, 48)
	m.sessionSource = row.source
	m.sessionScroll = 0
	m.sessionLines = nil

	switch row.source {
	case "claude":
		m.sessionLines = loadClaudeSessionLines(row.sessionID)
	case "opencode":
		m.sessionLines = loadOpenCodeSessionLines(row.sessionID)
	default:
		m.sessionLines = []string{"(no detail available for source: " + row.source + ")"}
	}
	if len(m.sessionLines) == 0 {
		m.sessionLines = []string{"(no entries found)"}
	}
	// Scroll to the line that mentions this row's op or ts so different
	// rows from the same session open at different positions.
	needle := row.op
	if row.ts != "" {
		needle = row.ts[:min(len(row.ts), 16)]
	}
	if needle != "" {
		for i, l := range m.sessionLines {
			if strings.Contains(l, needle) {
				m.sessionScroll = i
				break
			}
		}
	}
	m.showSession = true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// sessionDetailView renders the session detail overlay.
func (m *dashboardModel) sessionDetailView() string {
	t := m.theme

	sourceColor := t.Primary
	switch m.sessionSource {
	case "opencode":
		sourceColor = t.Accent
	case "maple":
		sourceColor = t.Success
	}

	titleStyle := lipgloss.NewStyle().Foreground(sourceColor).Bold(true)
	sep := lipgloss.NewStyle().Foreground(t.Muted).Render("  " + strings.Repeat("─", 62))
	title := titleStyle.Render("  " + m.sessionTitle)
	src := lipgloss.NewStyle().Foreground(t.Muted).Render("  source: " + m.sessionSource)

	visible := m.height - 14
	if visible < 4 {
		visible = 4
	}
	end := m.sessionScroll + visible
	if end > len(m.sessionLines) {
		end = len(m.sessionLines)
	}
	window := m.sessionLines[m.sessionScroll:end]

	var sb strings.Builder
	sb.WriteString(title + "\n" + src + "\n" + sep + "\n\n")
	for _, l := range window {
		sb.WriteString("  " + lipgloss.NewStyle().Foreground(t.Foreground).Render(l) + "\n")
	}

	total := len(m.sessionLines)
	if total > visible {
		pct := (m.sessionScroll * 100) / (total - visible)
		sb.WriteString(fmt.Sprintf("\n  %s\n",
			lipgloss.NewStyle().Foreground(t.Muted).Render(
				fmt.Sprintf("(%d%%)  j/k scroll · Esc close", pct))))
	} else {
		sb.WriteString("\n  " + lipgloss.NewStyle().Foreground(t.Muted).Render("Esc close") + "\n")
	}
	return sb.String()
}

// qaFileDetailView renders a test file as a full-screen overlay.
func (m *dashboardModel) qaFileDetailView() string {
	t := m.theme
	badge := testFrameworkBadge(m.qaFileFramework, t)
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("  " + filepath.Base(m.qaFileTitle))
	path := lipgloss.NewStyle().Foreground(t.Muted).Render("  " + m.qaFileTitle)
	sep := lipgloss.NewStyle().Foreground(t.Muted).Render("  " + strings.Repeat("─", 62))

	visible := m.height - 14
	if visible < 4 {
		visible = 4
	}
	end := m.qaFileScroll + visible
	if end > len(m.qaFileLines) {
		end = len(m.qaFileLines)
	}
	window := m.qaFileLines[m.qaFileScroll:end]

	var sb strings.Builder
	sb.WriteString(badge + " " + title + "\n" + path + "\n" + sep + "\n\n")
	for _, l := range window {
		var rendered string
		if m.qaFileFramework == "gherkin" {
			rendered = colorizeGherkin(l, t)
		} else {
			rendered = lipgloss.NewStyle().Foreground(t.Foreground).Render(l)
		}
		sb.WriteString("  " + rendered + "\n")
	}

	total := len(m.qaFileLines)
	if total > visible {
		pct := (m.qaFileScroll * 100) / (total - visible)
		sb.WriteString(fmt.Sprintf("\n  %s\n",
			lipgloss.NewStyle().Foreground(t.Muted).Render(
				fmt.Sprintf("(%d%%)  j/k scroll · r run test · Esc close", pct))))
	} else {
		sb.WriteString("\n  " + lipgloss.NewStyle().Foreground(t.Muted).Render("r run test · Esc close") + "\n")
	}
	return sb.String()
}

// testOutputView renders the live/completed test runner output.
func (m *dashboardModel) testOutputView() string {
	t := m.theme

	titleCol := t.Success
	statusStr := "PASSED"
	if m.testOutRunning {
		titleCol = t.Muted
		statusStr = "running…"
	} else if m.testOutFailed {
		titleCol = t.Error
		statusStr = "FAILED"
	}

	title := lipgloss.NewStyle().Foreground(titleCol).Bold(true).Render("  [" + statusStr + "] " + m.testOutTitle)
	sep := lipgloss.NewStyle().Foreground(t.Muted).Render("  " + strings.Repeat("─", 62))

	visible := m.height - 14
	if visible < 4 {
		visible = 4
	}
	end := m.testOutScroll + visible
	if end > len(m.testOutLines) {
		end = len(m.testOutLines)
	}
	window := m.testOutLines[m.testOutScroll:end]

	var sb strings.Builder
	sb.WriteString(title + "\n" + sep + "\n\n")
	for _, l := range window {
		col := t.Foreground
		tl := strings.TrimSpace(l)
		if strings.HasPrefix(tl, "FAIL") || strings.HasPrefix(tl, "--- FAIL") ||
			strings.Contains(tl, "FAILED") || strings.HasPrefix(tl, "Error") {
			col = t.Error
		} else if strings.HasPrefix(tl, "ok ") || strings.HasPrefix(tl, "--- PASS") ||
			strings.Contains(tl, "passed") || strings.HasPrefix(tl, "PASS") {
			col = t.Success
		}
		sb.WriteString("  " + lipgloss.NewStyle().Foreground(col).Render(l) + "\n")
	}

	total := len(m.testOutLines)
	if total > visible {
		pct := (m.testOutScroll * 100) / (total - visible)
		sb.WriteString(fmt.Sprintf("\n  %s\n",
			lipgloss.NewStyle().Foreground(t.Muted).Render(
				fmt.Sprintf("(%d%%)  j/k scroll · Esc close", pct))))
	}
	return sb.String()
}

// prDetailView renders `gh pr view` output as a full-screen overlay.
func (m *dashboardModel) prDetailView() string {
	t := m.theme
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("  " + m.prDetailTitle)
	sep := lipgloss.NewStyle().Foreground(t.Muted).Render("  " + strings.Repeat("─", 62))

	if m.prDetailLoading {
		return title + "\n" + sep + "\n\n  " +
			lipgloss.NewStyle().Foreground(t.Muted).Render("loading…") + "\n"
	}

	visible := m.height - 14
	if visible < 4 {
		visible = 4
	}
	end := m.prDetailScroll + visible
	if end > len(m.prDetailLines) {
		end = len(m.prDetailLines)
	}
	window := m.prDetailLines[m.prDetailScroll:end]

	var sb strings.Builder
	sb.WriteString(title + "\n" + sep + "\n\n")
	for _, l := range window {
		sb.WriteString("  " + lipgloss.NewStyle().Foreground(t.Foreground).Render(l) + "\n")
	}

	total := len(m.prDetailLines)
	if total > visible {
		pct := (m.prDetailScroll * 100) / (total - visible)
		sb.WriteString(fmt.Sprintf("\n  %s\n",
			lipgloss.NewStyle().Foreground(t.Muted).Render(
				fmt.Sprintf("(%d%%)  j/k scroll · o browser · a approve · Esc close", pct))))
	} else {
		sb.WriteString("\n  " + lipgloss.NewStyle().Foreground(t.Muted).Render("o browser · a approve · Esc close") + "\n")
	}
	return sb.String()
}
