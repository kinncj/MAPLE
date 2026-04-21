package main

import (
	"os"
	"path/filepath"
	"strings"
)

type superpowerDef struct {
	name        string
	description string
	tags        []string
	stageCount  int
}

// loadSuperpowers reads .claude/superpowers/*.yaml, skipping schema.yaml.
// Returns an empty slice (never nil) when the directory is absent or empty.
func loadSuperpowers() []superpowerDef {
	entries, err := filepath.Glob(".claude/superpowers/*.yaml")
	if err != nil || len(entries) == 0 {
		return []superpowerDef{}
	}

	var out []superpowerDef
	for _, path := range entries {
		if filepath.Base(path) == "schema.yaml" {
			continue
		}
		sp := parseSuperpowerFile(path)
		if sp.name != "" {
			out = append(out, sp)
		}
	}
	return out
}

func parseSuperpowerFile(path string) superpowerDef {
	data, err := os.ReadFile(path)
	if err != nil {
		return superpowerDef{}
	}
	text := string(data)

	sp := superpowerDef{}

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "name:") {
			sp.name = strings.Trim(strings.TrimPrefix(line, "name:"), " \"'")
			continue
		}
		if strings.HasPrefix(line, "description:") {
			sp.description = strings.Trim(strings.TrimPrefix(line, "description:"), " \"'")
			continue
		}
		if strings.HasPrefix(line, "tags:") {
			raw := strings.TrimPrefix(line, "tags:")
			raw = strings.Trim(raw, " []")
			for _, t := range strings.Split(raw, ",") {
				t = strings.Trim(t, " \"'")
				if t != "" {
					sp.tags = append(sp.tags, t)
				}
			}
			continue
		}
		if strings.HasPrefix(line, "- name:") {
			sp.stageCount++
		}
	}

	// Fall back to filename if name field is absent
	if sp.name == "" {
		base := filepath.Base(path)
		sp.name = strings.TrimSuffix(base, ".yaml")
	}

	return sp
}
