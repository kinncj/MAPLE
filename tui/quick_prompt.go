package main

import (
	"os"
	"path/filepath"
	"strings"
)

type quickItem struct {
	name        string
	description string
	kind        string // "skill" or "agent"
}

// loadQuickItems scans .claude/skills/ and .claude/agents/ for installed items.
func loadQuickItems() []quickItem {
	var out []quickItem

	skillDirs, _ := filepath.Glob(".claude/skills/*")
	for _, dir := range skillDirs {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		item := parseSkillDir(dir)
		if item.name != "" {
			out = append(out, item)
		}
	}

	agentFiles, _ := filepath.Glob(".claude/agents/*.md")
	for _, f := range agentFiles {
		item := parseAgentFile(f)
		if item.name != "" {
			out = append(out, item)
		}
	}

	return out
}

func parseSkillDir(dir string) quickItem {
	name := filepath.Base(dir)
	description := ""
	data, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "name:") {
				v := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				v = strings.Trim(v, "\"'")
				if v != "" {
					name = v
				}
			}
			if strings.HasPrefix(line, "description:") {
				v := strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				v = strings.Trim(v, "\"'")
				if v != "" {
					description = v
				}
			}
		}
	}
	return quickItem{name: name, description: description, kind: "skill"}
}

func parseAgentFile(path string) quickItem {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, ".md")
	description := ""
	data, err := os.ReadFile(path)
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "description:") {
				v := strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				v = strings.Trim(v, "\"'")
				if v != "" {
					description = v
				}
			}
		}
	}
	return quickItem{name: name, description: description, kind: "agent"}
}

// quickFilter returns items whose name or description contains all words in query.
func quickFilter(items []quickItem, query string) []quickItem {
	query = strings.TrimSpace(query)
	if query == "" {
		return items
	}
	words := strings.Fields(strings.ToLower(query))
	var out []quickItem
	for _, item := range items {
		haystack := strings.ToLower(item.name + " " + item.description + " " + item.kind)
		match := true
		for _, w := range words {
			if !strings.Contains(haystack, w) {
				match = false
				break
			}
		}
		if match {
			out = append(out, item)
		}
	}
	return out
}
