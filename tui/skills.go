package main

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type skillRow struct {
	pkg      string // "owner/repo@skill"
	installs string // "21.1K"
	url      string // "https://skills.sh/..."
}

type skillsSearchedMsg struct {
	items []skillRow
	err   error
}

type skillInstalledMsg struct {
	pkg string
	err error
}

func searchSkillsCmd(query string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"--yes", "skills", "find"}
		if query != "" {
			args = append(args, query)
		}
		out, err := exec.Command("npx", args...).Output()
		if err != nil {
			return skillsSearchedMsg{err: fmt.Errorf("npx skills find: %s", strings.TrimSpace(err.Error()))}
		}
		return skillsSearchedMsg{items: parseSkillsOutput(string(out))}
	}
}

func installSkillCmd(pkg string) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("npx", "--yes", "skills", "add", pkg, "--all", "-y").CombinedOutput()
		if err != nil {
			msg := strings.TrimSpace(string(out))
			if msg == "" {
				msg = err.Error()
			}
			return skillInstalledMsg{pkg: pkg, err: fmt.Errorf("%s", msg)}
		}
		return skillInstalledMsg{pkg: pkg}
	}
}

// parseSkillsOutput parses the plain-text output of `npx skills find <query>`.
// Each result looks like:
//
//	vercel-labs/agent-browser@dogfood 21.1K installs
//	└ https://skills.sh/vercel-labs/agent-browser/dogfood
func parseSkillsOutput(s string) []skillRow {
	var rows []skillRow
	lastWasPkg := false
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			lastWasPkg = false
			continue
		}
		if strings.HasPrefix(line, "└ ") {
			if lastWasPkg && len(rows) > 0 {
				rows[len(rows)-1].url = strings.TrimPrefix(line, "└ ")
			}
			lastWasPkg = false
			continue
		}
		// "owner/repo@skill N installs"
		parts := strings.Fields(line)
		if len(parts) >= 3 && parts[len(parts)-1] == "installs" && strings.Contains(parts[0], "/") {
			rows = append(rows, skillRow{
				pkg:      parts[0],
				installs: parts[len(parts)-2],
			})
			lastWasPkg = true
		} else {
			lastWasPkg = false
		}
	}
	return rows
}
