package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type skillRow struct {
	pkg      string // "owner/repo@skill"
	installs string // "21.1K"
	url      string // "https://skills.sh/..."
}

type skillsSearchedMsg struct {
	items    []skillRow
	err      error
	searched bool // true = search ran (even if 0 results)
}

type skillInstalledMsg struct {
	pkg string
	err error
}

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[mGKHF]`)

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

func searchSkillsCmd(query string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"--yes", "skills", "find"}
		if query != "" {
			args = append(args, query)
		}
		cmd := exec.Command("npx", args...)
		// Disable colors so output is clean ASCII for parsing
		cmd.Env = append(os.Environ(), "NO_COLOR=1", "FORCE_COLOR=0")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		_ = cmd.Run() // ignore exit code — parse whatever stdout we got

		items := parseSkillsOutput(stripANSI(stdout.String()))
		if len(items) == 0 {
			errStr := strings.TrimSpace(stderr.String())
			if errStr != "" {
				return skillsSearchedMsg{searched: true, err: fmt.Errorf("%s", stripANSI(errStr))}
			}
		}
		return skillsSearchedMsg{items: items, searched: true}
	}
}

func installSkillCmd(pkg string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("npx", "--yes", "skills", "add", pkg, "--all", "-y")
		cmd.Env = append(os.Environ(), "NO_COLOR=1", "FORCE_COLOR=0")
		out, err := cmd.CombinedOutput()
		if err != nil {
			msg := strings.TrimSpace(stripANSI(string(out)))
			if msg == "" {
				msg = err.Error()
			}
			return skillInstalledMsg{pkg: pkg, err: fmt.Errorf("%s", msg)}
		}
		return skillInstalledMsg{pkg: pkg}
	}
}

// parseSkillsOutput parses the output of `npx skills find <query>`.
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
		// "owner/repo@skill N installs" or "owner/repo@skill N,NNN installs"
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
