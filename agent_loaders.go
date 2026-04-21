package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ─── Agent activity loaders ───────────────────────────────────────────────────

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

	rows = append(rows, loadClaudeSessionActivity(cwd)...)
	rows = append(rows, loadOpenCodeActivity(cwd)...)

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

func loadClaudeSessionActivity(cwd string) []agentRow {
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

func claudeEntryToRow(e map[string]interface{}) (agentRow, bool) {
	ts, _ := e["timestamp"].(string)
	if ts != "" && len(ts) > 19 {
		ts = ts[:19]
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
		if err := json.Unmarshal([]byte(raw), &e); err != nil {
			continue
		}
		ts, _ := e["timestamp"].(string)
		if len(ts) > 16 {
			ts = ts[:16]
		}
		switch e["type"] {
		case "ai-title":
			if title, _ := e["aiTitle"].(string); title != "" {
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
		switch d["type"] {
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
			if title, _ := d["title"].(string); title != "" {
				lines = append(lines, fmt.Sprintf("[%s] step: %s", ts, title))
			}
		}
	}
	return lines
}
