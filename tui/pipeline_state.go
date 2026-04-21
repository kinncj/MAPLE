package main

import (
	"encoding/json"
	"os"
	"strings"
)

// pipelineState mirrors the state written by the superpower-runner skill
// to .claude/state/maple.json
type pipelineState struct {
	Superpower      string `json:"superpower"`
	Stage           string `json:"stage"`
	Status          string `json:"status"` // RUNNING | PAUSED | DONE | FAILED
	AwaitingApproval string `json:"awaiting_approval"`
	StartedAt       string `json:"started_at"`
	UpdatedAt       string `json:"updated_at"`
	// recovery marker fields written by the TUI itself
	State string `json:"state"`
	TS    string `json:"ts"`
}

func (p pipelineState) isSuperpower() bool {
	return p.Superpower != ""
}

func (p pipelineState) statusIcon() string {
	switch strings.ToUpper(p.Status) {
	case "RUNNING":
		return "▶"
	case "PAUSED":
		return "⏸"
	case "DONE":
		return "✓"
	case "FAILED":
		return "✗"
	default:
		return "·"
	}
}

// approvalPending returns the stage name from .claude/state/approval-pending.txt,
// or "" if no approval is waiting. The superpower-runner skill writes this file at
// human-approval gates; the TUI deletes it when the user presses [a].
func approvalPending() string {
	data, err := os.ReadFile(".claude/state/approval-pending.txt")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func loadPipelineState() (pipelineState, error) {
	data, err := os.ReadFile(".claude/state/maple.json")
	if err != nil {
		return pipelineState{}, err
	}
	var ps pipelineState
	if err := json.Unmarshal(data, &ps); err != nil {
		return pipelineState{}, err
	}
	return ps, nil
}
