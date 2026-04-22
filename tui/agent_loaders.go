package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ─── Session list loaders ─────────────────────────────────────────────────────

func loadSessions() []sessionRow {
	cwd, _ := os.Getwd()
	var out []sessionRow
	out = append(out, loadClaudeSessions(cwd)...)
	out = append(out, loadOpenCodeSessions(cwd)...)
	out = append(out, loadCopilotSessions(cwd)...)
	// Sort newest first across sources
	sort.Slice(out, func(i, j int) bool {
		return out[i].ts > out[j].ts
	})
	return out
}

// loadClaudeSessions returns one sessionRow per Claude Code JSONL file for cwd.
func loadClaudeSessions(cwd string) []sessionRow {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	encoded := strings.ReplaceAll(cwd, "/", "-")
	projectDir := filepath.Join(home, ".claude", "projects", encoded)

	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil
	}

	type finfo struct {
		path string
		t    time.Time
	}
	var files []finfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, finfo{filepath.Join(projectDir, e.Name()), info.ModTime()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].t.After(files[j].t) })

	var rows []sessionRow
	for _, f := range files {
		if len(rows) >= 10 {
			break
		}
		data, err := os.ReadFile(f.path)
		if err != nil {
			continue
		}

		title := ""
		toolCount := 0

		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var e map[string]interface{}
			if json.Unmarshal([]byte(line), &e) != nil {
				continue
			}
			if e["type"] == "ai-title" {
				if t, _ := e["aiTitle"].(string); t != "" {
					title = t
				}
			}
			if e["type"] == "assistant" {
				if msg, ok := e["message"].(map[string]interface{}); ok {
					if content, ok := msg["content"].([]interface{}); ok {
						for _, c := range content {
							if cm, ok := c.(map[string]interface{}); ok && cm["type"] == "tool_use" {
								toolCount++
							}
						}
					}
				}
			}
		}

		if title == "" {
			// Fall back to file name without extension as a unique identifier
			base := filepath.Base(f.path)
			title = strings.TrimSuffix(base, ".jsonl")
			if len(title) > 20 {
				title = title[:8] + "…" + title[len(title)-8:]
			}
		}

		rows = append(rows, sessionRow{
			id:        f.path,
			title:     title,
			source:    "claude",
			ts:        f.t.Format("2006-01-02 15:04"),
			toolCount: toolCount,
		})
	}
	return rows
}

// loadOpenCodeSessions returns one sessionRow per OpenCode session for cwd.
func loadOpenCodeSessions(cwd string) []sessionRow {
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
	// Include tool count via subquery
	query := fmt.Sprintf(
		"SELECT s.id, s.title, datetime(s.time_updated,'unixepoch'), "+
			"(SELECT COUNT(*) FROM part p WHERE p.session_id=s.id AND json_extract(p.data,'$.type')='tool') "+
			"FROM session s JOIN project pr ON s.project_id=pr.id "+
			"WHERE pr.worktree='%s' ORDER BY s.time_updated DESC LIMIT 10;",
		escapedCwd,
	)
	out, err := exec.Command(sqlite3Path, db, query).Output()
	if err != nil {
		return nil
	}

	var rows []sessionRow
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 2 {
			continue
		}
		title := parts[1]
		if title == "" {
			title = parts[0][:min(len(parts[0]), 16)]
		}
		ts := ""
		if len(parts) >= 3 {
			ts = parts[2]
			if len(ts) > 16 {
				ts = ts[:16]
			}
		}
		toolCount := 0
		if len(parts) >= 4 {
			fmt.Sscanf(parts[3], "%d", &toolCount)
		}
		rows = append(rows, sessionRow{
			id:        parts[0],
			title:     title,
			source:    "opencode",
			ts:        ts,
			toolCount: toolCount,
		})
	}
	return rows
}

// loadCopilotSessions returns one sessionRow per GitHub Copilot CLI session matching cwd.
// Copilot CLI stores session state under ~/.copilot/session-state/<uuid>/ with a
// workspace.yaml describing the session and an events.jsonl stream.
func loadCopilotSessions(cwd string) []sessionRow {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	root := filepath.Join(home, ".copilot", "session-state")
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	type cand struct {
		row       sessionRow
		updatedAt time.Time
	}
	var cands []cand
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		meta := parseCopilotWorkspace(filepath.Join(dir, "workspace.yaml"))
		if meta["cwd"] != cwd {
			continue
		}
		id := meta["id"]
		if id == "" {
			id = e.Name()
		}
		title := meta["summary"]
		if title == "" {
			if len(id) > 16 {
				title = id[:8] + "…" + id[len(id)-8:]
			} else {
				title = id
			}
		}
		tsRaw := meta["updated_at"]
		if tsRaw == "" {
			tsRaw = meta["created_at"]
		}
		ts := tsRaw
		updated := time.Time{}
		if t, err := time.Parse(time.RFC3339Nano, tsRaw); err == nil {
			ts = t.Format("2006-01-02 15:04")
			updated = t
		} else if t, err := time.Parse(time.RFC3339, tsRaw); err == nil {
			ts = t.Format("2006-01-02 15:04")
			updated = t
		} else if info, err := os.Stat(filepath.Join(dir, "events.jsonl")); err == nil {
			updated = info.ModTime()
			ts = updated.Format("2006-01-02 15:04")
		}

		toolCount := countCopilotTools(filepath.Join(dir, "events.jsonl"))

		cands = append(cands, cand{
			row: sessionRow{
				id:        id,
				title:     title,
				source:    "copilot",
				ts:        ts,
				toolCount: toolCount,
			},
			updatedAt: updated,
		})
	}

	sort.Slice(cands, func(i, j int) bool { return cands[i].updatedAt.After(cands[j].updatedAt) })
	if len(cands) > 10 {
		cands = cands[:10]
	}
	rows := make([]sessionRow, 0, len(cands))
	for _, c := range cands {
		rows = append(rows, c.row)
	}
	return rows
}

// parseCopilotWorkspace reads a Copilot workspace.yaml — mostly flat "key: value" pairs
// with occasional block scalars (|, |-, >, >-) for long summaries. Indented lines
// following a block-scalar header are joined into the value.
func parseCopilotWorkspace(path string) map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}
	}
	out := map[string]string{}
	lines := strings.Split(string(data), "\n")
	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if val == "|" || val == "|-" || val == ">" || val == ">-" {
			var parts []string
			for j := i + 1; j < len(lines); j++ {
				peek := lines[j]
				if strings.TrimSpace(peek) == "" {
					parts = append(parts, "")
					continue
				}
				if !strings.HasPrefix(peek, " ") && !strings.HasPrefix(peek, "\t") {
					break
				}
				parts = append(parts, strings.TrimLeft(peek, " \t"))
				i = j
			}
			sep := "\n"
			if val == ">" || val == ">-" {
				sep = " "
			}
			out[key] = strings.TrimSpace(strings.Join(parts, sep))
			continue
		}
		out[key] = strings.Trim(val, `"'`)
	}
	return out
}

// countCopilotTools returns the number of tool.execution_start events in events.jsonl.
func countCopilotTools(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	count := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if bytes.Contains(line, []byte(`"type":"tool.execution_start"`)) {
			count++
		}
	}
	return count
}

// ─── Session detail loaders ───────────────────────────────────────────────────

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
		if json.Unmarshal([]byte(raw), &e) != nil {
			continue
		}
		ts, _ := e["timestamp"].(string)
		if len(ts) > 16 {
			ts = ts[:16]
		}
		switch e["type"] {
		case "ai-title":
			if title, _ := e["aiTitle"].(string); title != "" {
				lines = append(lines, fmt.Sprintf("[%s] ── %s", ts, title))
			}
		case "last-prompt":
			prompt, _ := e["lastPrompt"].(string)
			if prompt != "" {
				if len(prompt) > 72 {
					prompt = prompt[:72] + "…"
				}
				lines = append(lines, fmt.Sprintf("[%s] ▶ %s", ts, prompt))
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
					lines = append(lines, fmt.Sprintf("[%s]   %-16s %s", ts, toolName, file))
				} else {
					lines = append(lines, fmt.Sprintf("[%s]   %s", ts, toolName))
				}
			}
		}
	}
	return lines
}

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
		var d map[string]interface{}
		if json.Unmarshal([]byte(row[idx+1:]), &d) != nil {
			continue
		}
		switch d["type"] {
		case "text":
			text, _ := d["text"].(string)
			if text != "" {
				if len(text) > 72 {
					text = text[:72] + "…"
				}
				lines = append(lines, fmt.Sprintf("[%s] ▶ %s", ts, text))
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
				lines = append(lines, fmt.Sprintf("[%s]   %-16s %s", ts, toolName, file))
			} else {
				lines = append(lines, fmt.Sprintf("[%s]   %s", ts, toolName))
			}
		case "step-start":
			if title, _ := d["title"].(string); title != "" {
				lines = append(lines, fmt.Sprintf("[%s] ── %s", ts, title))
			}
		}
	}
	return lines
}

// loadCopilotSessionLines parses a Copilot CLI events.jsonl into a compact activity timeline.
func loadCopilotSessionLines(sessionID string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	path := filepath.Join(home, ".copilot", "session-state", sessionID, "events.jsonl")
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var e map[string]interface{}
		if json.Unmarshal(sc.Bytes(), &e) != nil {
			continue
		}
		tsRaw, _ := e["timestamp"].(string)
		ts := tsRaw
		if len(ts) > 16 {
			ts = ts[:16]
		}
		data, _ := e["data"].(map[string]interface{})
		switch e["type"] {
		case "user.message":
			if data != nil {
				msg, _ := data["content"].(string)
				if msg == "" {
					msg, _ = data["message"].(string)
				}
				if msg != "" {
					if len(msg) > 72 {
						msg = msg[:72] + "…"
					}
					lines = append(lines, fmt.Sprintf("[%s] ▶ %s", ts, msg))
				}
			}
		case "assistant.message":
			if data != nil {
				msg, _ := data["content"].(string)
				if msg != "" {
					if len(msg) > 72 {
						msg = msg[:72] + "…"
					}
					lines = append(lines, fmt.Sprintf("[%s] ◀ %s", ts, msg))
				}
			}
		case "tool.execution_start":
			if data == nil {
				continue
			}
			toolName, _ := data["toolName"].(string)
			if toolName == "" {
				continue
			}
			file := ""
			if args, ok := data["arguments"].(map[string]interface{}); ok {
				for _, k := range []string{"file_path", "path", "command", "pattern", "intent"} {
					if v, _ := args[k].(string); v != "" {
						file = v
						if len(file) > 48 {
							file = "…" + file[len(file)-48:]
						}
						break
					}
				}
			}
			if file != "" {
				lines = append(lines, fmt.Sprintf("[%s]   %-16s %s", ts, toolName, file))
			} else {
				lines = append(lines, fmt.Sprintf("[%s]   %s", ts, toolName))
			}
		case "session.model_change":
			if data != nil {
				if model, _ := data["newModel"].(string); model != "" {
					lines = append(lines, fmt.Sprintf("[%s] ── model: %s", ts, model))
				}
			}
		}
	}
	return lines
}
