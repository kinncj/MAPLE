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

// ─── Session list loaders ─────────────────────────────────────────────────────

func loadSessions() []sessionRow {
	cwd, _ := os.Getwd()
	var out []sessionRow
	out = append(out, loadClaudeSessions(cwd)...)
	out = append(out, loadOpenCodeSessions(cwd)...)
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
