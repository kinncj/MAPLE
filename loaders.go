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

	tea "github.com/charmbracelet/bubbletea"
)

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
		path:     path,
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
	cwd, _ := os.Getwd()

	var rows []agentRow
	seen := map[string]bool{}
	add := func(r agentRow) {
		key := r.agent + "|" + r.op + "|" + r.file
		if seen[key] {
			return
		}
		seen[key] = true
		rows = append(rows, r)
	}

	// Primary: MAPLE skills log
	if data, err := os.ReadFile(".claude/logs/skills.jsonl"); err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}
			var e map[string]interface{}
			if err := json.Unmarshal([]byte(line), &e); err != nil {
				continue
			}
			agent, _ := e["agent"].(string)
			if agent == "" {
				agent, _ = e["skill"].(string)
			}
			if agent == "" {
				continue
			}
			add(agentRow{
				agent:  agent,
				op:     str(e["op"]),
				file:   str(e["file"]),
				ts:     str(e["ts"]),
				source: "maple",
			})
		}
	}

	// Secondary: Claude Code native session logs for the current project
	rows = append(rows, loadClaudeSessionActivity(cwd)...)

	// Tertiary: OpenCode session activity for the current project
	rows = append(rows, loadOpenCodeActivity(cwd)...)

	// Re-deduplicate after merge
	final := rows[:0]
	finalSeen := map[string]bool{}
	for _, r := range rows {
		key := r.agent + "|" + r.op + "|" + r.file
		if finalSeen[key] {
			continue
		}
		finalSeen[key] = true
		final = append(final, r)
		if len(final) >= 15 {
			break
		}
	}
	return final
}

// loadClaudeSessionActivity reads Claude Code's native session JSONL for the
// current working directory and extracts recent tool-use activity.
func loadClaudeSessionActivity(cwd string) []agentRow {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	// Claude Code encodes the project path by replacing every "/" with "-"
	encoded := strings.ReplaceAll(cwd, "/", "-")
	projectDir := filepath.Join(home, ".claude", "projects", encoded)

	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil
	}

	// Collect JSONL files sorted newest-first by ModTime
	type mtime struct {
		path string
		t    time.Time
	}
	var files []mtime
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, mtime{filepath.Join(projectDir, e.Name()), info.ModTime()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].t.After(files[j].t) })

	var rows []agentRow
	seen := map[string]bool{}

	for _, f := range files {
		if len(rows) >= 8 {
			break
		}
		data, err := os.ReadFile(f.path)
		if err != nil {
			continue
		}
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		// Walk backward — most recent first
		for i := len(lines) - 1; i >= 0 && len(rows) < 8; i-- {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}
			var e map[string]interface{}
			if err := json.Unmarshal([]byte(line), &e); err != nil {
				continue
			}
			r, ok := claudeEntryToRow(e)
			if !ok {
				continue
			}
			r.source = "claude"
			r.sessionID = f.path
			key := r.agent + "|" + r.op + "|" + r.file
			if seen[key] {
				continue
			}
			seen[key] = true
			rows = append(rows, r)
		}
	}
	return rows
}

// loadOpenCodeActivity reads OpenCode's SQLite database and extracts recent
// tool-use activity for the given working directory.
func loadOpenCodeActivity(cwd string) []agentRow {
	sqlite3Path, err := exec.LookPath("sqlite3")
	if err != nil {
		return nil
	}
	home, _ := os.UserHomeDir()
	db := filepath.Join(home, ".local", "share", "opencode", "opencode.db")
	if _, err := os.Stat(db); err != nil {
		return nil
	}

	escapedCwd := strings.ReplaceAll(cwd, "'", "''")
	sessionsSQL := fmt.Sprintf(
		"SELECT s.id,s.title,datetime(s.time_updated,'unixepoch') FROM session s JOIN project p ON s.project_id=p.id WHERE p.worktree='%s' ORDER BY s.time_updated DESC LIMIT 5;",
		escapedCwd,
	)
	out, err := exec.Command(sqlite3Path, db, sessionsSQL).Output()
	if err != nil {
		return nil
	}

	var rows []agentRow
	seen := map[string]bool{}

	for _, sessLine := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		sessLine = strings.TrimSpace(sessLine)
		if sessLine == "" {
			continue
		}
		parts := strings.SplitN(sessLine, "|", 3)
		if len(parts) < 2 {
			continue
		}
		sessionID := parts[0]
		sessionTitle := parts[1]
		ts := ""
		if len(parts) >= 3 {
			ts = parts[2]
			if len(ts) > 16 {
				ts = ts[:16]
			}
		}

		key := "opencode|session|" + sessionID
		if !seen[key] {
			seen[key] = true
			rows = append(rows, agentRow{
				agent:     "opencode",
				op:        truncate(sessionTitle, 36),
				ts:        ts,
				source:    "opencode",
				sessionID: sessionID,
			})
		}

		partsSQL := fmt.Sprintf(
			"SELECT p.data FROM part p WHERE p.session_id='%s' AND json_extract(p.data,'$.type')='tool' ORDER BY p.time_created DESC LIMIT 8;",
			sessionID,
		)
		partsOut, err := exec.Command(sqlite3Path, db, partsSQL).Output()
		if err != nil {
			continue
		}
		for _, dataLine := range strings.Split(strings.TrimSpace(string(partsOut)), "\n") {
			dataLine = strings.TrimSpace(dataLine)
			if dataLine == "" {
				continue
			}
			var d map[string]interface{}
			if err := json.Unmarshal([]byte(dataLine), &d); err != nil {
				continue
			}
			toolName, _ := d["tool"].(string)
			if toolName == "" {
				continue
			}
			file := ""
			if state, ok := d["state"].(map[string]interface{}); ok {
				if inp, ok := state["input"].(map[string]interface{}); ok {
					for _, k := range []string{"file_path", "path", "command"} {
						if v, _ := inp[k].(string); v != "" {
							file = v
							if len(file) > 32 {
								file = "…" + file[len(file)-32:]
							}
							break
						}
					}
				}
			}
			rkey := "opencode|" + toolName + "|" + file
			if seen[rkey] {
				continue
			}
			seen[rkey] = true
			rows = append(rows, agentRow{
				agent:     "opencode",
				op:        toolName,
				file:      file,
				ts:        ts,
				source:    "opencode",
				sessionID: sessionID,
			})
			if len(rows) >= 10 {
				return rows
			}
		}
	}
	return rows
}

// claudeEntryToRow converts one Claude Code session JSONL entry into an agentRow.
func claudeEntryToRow(e map[string]interface{}) (agentRow, bool) {
	ts, _ := e["timestamp"].(string)
	if ts != "" && len(ts) > 19 {
		ts = ts[:19] // trim to "2006-01-02T15:04:05"
	}

	switch e["type"] {
	case "ai-title":
		title, _ := e["aiTitle"].(string)
		if title == "" {
			return agentRow{}, false
		}
		return agentRow{agent: "claude", op: title, ts: ts}, true

	case "last-prompt":
		prompt, _ := e["lastPrompt"].(string)
		if prompt == "" {
			return agentRow{}, false
		}
		if len(prompt) > 48 {
			prompt = prompt[:48] + "…"
		}
		return agentRow{agent: "user", op: prompt, ts: ts}, true

	case "user":
		// Sub-agent tool result entries carry agentType
		msg, _ := e["message"].(map[string]interface{})
		if msg == nil {
			return agentRow{}, false
		}
		content, _ := msg["content"].([]interface{})
		for _, c := range content {
			cm, _ := c.(map[string]interface{})
			if cm == nil {
				continue
			}
			agentType, _ := e["agentType"].(string)
			if agentType == "" {
				agentType, _ = cm["agentType"].(string)
			}
			if agentType != "" {
				toolCount := ""
				if v, ok := e["totalToolUseCount"].(float64); ok {
					toolCount = fmt.Sprintf("%d tools", int(v))
				}
				return agentRow{agent: agentType, op: toolCount, ts: ts}, true
			}
		}

	case "assistant":
		msg, _ := e["message"].(map[string]interface{})
		if msg == nil {
			return agentRow{}, false
		}
		content, _ := msg["content"].([]interface{})
		for _, c := range content {
			cm, _ := c.(map[string]interface{})
			if cm == nil || cm["type"] != "tool_use" {
				continue
			}
			toolName, _ := cm["name"].(string)
			if toolName == "" {
				continue
			}
			file := ""
			if inp, ok := cm["input"].(map[string]interface{}); ok {
				for _, k := range []string{"file_path", "path", "command"} {
					if v, _ := inp[k].(string); v != "" {
						file = v
						if len(file) > 32 {
							file = "…" + file[len(file)-32:]
						}
						break
					}
				}
			}
			return agentRow{agent: "claude", op: toolName, file: file, ts: ts}, true
		}
	}
	return agentRow{}, false
}

func str(v interface{}) string {
	s, _ := v.(string)
	return s
}

// loadClaudeSessionLines reads a Claude Code session JSONL file and returns
// human-readable lines for the session detail overlay.
func loadClaudeSessionLines(filePath string) []string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	var lines []string
	for _, raw := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		var e map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &e); err != nil {
			continue
		}
		ts, _ := e["timestamp"].(string)
		if len(ts) > 16 {
			ts = ts[:16]
		}
		switch e["type"] {
		case "ai-title":
			title, _ := e["aiTitle"].(string)
			if title != "" {
				lines = append(lines, fmt.Sprintf("[%s] session: %s", ts, title))
			}
		case "last-prompt":
			prompt, _ := e["lastPrompt"].(string)
			if prompt != "" {
				if len(prompt) > 72 {
					prompt = prompt[:72] + "…"
				}
				lines = append(lines, fmt.Sprintf("[%s] user: %s", ts, prompt))
			}
		case "assistant":
			msg, _ := e["message"].(map[string]interface{})
			if msg == nil {
				continue
			}
			content, _ := msg["content"].([]interface{})
			for _, c := range content {
				cm, _ := c.(map[string]interface{})
				if cm == nil || cm["type"] != "tool_use" {
					continue
				}
				toolName, _ := cm["name"].(string)
				if toolName == "" {
					continue
				}
				file := ""
				if inp, ok := cm["input"].(map[string]interface{}); ok {
					for _, k := range []string{"file_path", "path", "command"} {
						if v, _ := inp[k].(string); v != "" {
							file = v
							if len(file) > 48 {
								file = "…" + file[len(file)-48:]
							}
							break
						}
					}
				}
				if file != "" {
					lines = append(lines, fmt.Sprintf("[%s] tool: %-14s %s", ts, toolName, file))
				} else {
					lines = append(lines, fmt.Sprintf("[%s] tool: %s", ts, toolName))
				}
			}
		}
	}
	return lines
}

// loadOpenCodeSessionLines queries OpenCode's SQLite database for all parts
// of a given session and returns human-readable lines.
func loadOpenCodeSessionLines(sessionID string) []string {
	sqlite3Path, err := exec.LookPath("sqlite3")
	if err != nil {
		return nil
	}
	home, _ := os.UserHomeDir()
	db := filepath.Join(home, ".local", "share", "opencode", "opencode.db")
	if _, err := os.Stat(db); err != nil {
		return nil
	}

	escapedID := strings.ReplaceAll(sessionID, "'", "''")
	query := fmt.Sprintf(
		"SELECT datetime(time_created,'unixepoch'),data FROM part WHERE session_id='%s' ORDER BY time_created ASC;",
		escapedID,
	)
	out, err := exec.Command(sqlite3Path, db, query).Output()
	if err != nil {
		return nil
	}

	var lines []string
	for _, row := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		row = strings.TrimSpace(row)
		if row == "" {
			continue
		}
		idx := strings.Index(row, "|")
		if idx < 0 {
			continue
		}
		ts := row[:idx]
		if len(ts) > 16 {
			ts = ts[:16]
		}
		dataStr := row[idx+1:]
		var d map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &d); err != nil {
			continue
		}
		partType, _ := d["type"].(string)
		switch partType {
		case "text":
			text, _ := d["text"].(string)
			if text != "" {
				if len(text) > 72 {
					text = text[:72] + "…"
				}
				lines = append(lines, fmt.Sprintf("[%s] text: %s", ts, text))
			}
		case "tool":
			toolName, _ := d["tool"].(string)
			file := ""
			if state, ok := d["state"].(map[string]interface{}); ok {
				if inp, ok := state["input"].(map[string]interface{}); ok {
					for _, k := range []string{"file_path", "path", "command"} {
						if v, _ := inp[k].(string); v != "" {
							file = v
							if len(file) > 48 {
								file = "…" + file[len(file)-48:]
							}
							break
						}
					}
				}
			}
			if file != "" {
				lines = append(lines, fmt.Sprintf("[%s] tool: %-14s %s", ts, toolName, file))
			} else {
				lines = append(lines, fmt.Sprintf("[%s] tool: %s", ts, toolName))
			}
		case "step-start":
			title, _ := d["title"].(string)
			if title != "" {
				lines = append(lines, fmt.Sprintf("[%s] step: %s", ts, title))
			}
		}
	}
	return lines
}

func loadQA() (files int, scenarios int, paths []string) {
	entries, err := filepath.Glob("tests/features/*.feature")
	if err != nil {
		return
	}
	files = len(entries)
	paths = entries
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

func loadPRDetailCmd(number int, title string) tea.Cmd {
	return func() tea.Msg {
		ghPath, err := exec.LookPath("gh")
		if err != nil {
			return prDetailLoadedMsg{err: "gh not found", title: title}
		}
		out, err := exec.Command(ghPath, "pr", "view", fmt.Sprintf("%d", number)).Output()
		if err != nil {
			return prDetailLoadedMsg{err: strings.TrimSpace(string(out)), title: title}
		}
		lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
		return prDetailLoadedMsg{lines: lines, title: fmt.Sprintf("#%d %s", number, title)}
	}
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
