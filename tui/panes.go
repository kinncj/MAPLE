package main

import (
	"encoding/json"
	"os"
	"os/exec"
)

type paneRef struct {
	Kind   string `json:"kind"`   // "tmux", "zellij", ""
	Target string `json:"target"` // pane id (tmux) or tab/session name
}

const panesFile = ".claude/state/panes.json"

func loadPanes() map[string]paneRef {
	data, err := os.ReadFile(panesFile)
	if err != nil {
		return map[string]paneRef{}
	}
	var m map[string]paneRef
	if err := json.Unmarshal(data, &m); err != nil {
		return map[string]paneRef{}
	}
	return m
}

func savePaneRef(harness string, p paneRef) {
	_ = os.MkdirAll(".claude/state", 0o755)
	m := loadPanes()
	m[harness] = p
	data, _ := json.Marshal(m)
	_ = os.WriteFile(panesFile, append(data, '\n'), 0o644)
}

// sendContinueToPane types "continue\n" into the target pane, as if the
// user typed it in the agent's terminal. Fails silently when the pane is
// gone or the multiplexer is unreachable.
func sendContinueToPane(p paneRef) bool {
	if p.Kind == "" || p.Target == "" {
		return false
	}
	switch p.Kind {
	case "tmux":
		if err := exec.Command("tmux", "send-keys", "-t", p.Target, "continue", "Enter").Run(); err == nil {
			return true
		}
	case "zellij":
		if err := exec.Command("zellij", "action", "go-to-tab-name", p.Target).Run(); err != nil {
			return false
		}
		if err := exec.Command("zellij", "action", "write-chars", "continue").Run(); err != nil {
			return false
		}
		if err := exec.Command("zellij", "action", "write", "13").Run(); err == nil {
			return true
		}
	}
	return false
}

// notifyAllPanesContinue sends "continue" to every recorded pane. Returns
// the number of panes that accepted the keys.
func notifyAllPanesContinue() int {
	n := 0
	for _, p := range loadPanes() {
		if sendContinueToPane(p) {
			n++
		}
	}
	return n
}
