package main

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

const stalePipelineThreshold = 10 * time.Minute

// pipelineState mirrors the state written by the pipeline-runner skill
// to .claude/state/maple.json
type pipelineState struct {
	Taffy           string `json:"taffy"`
	Stage           string `json:"stage"`
	Status          string `json:"status"` // RUNNING | PAUSED | DONE | FAILED
	AwaitingApproval string `json:"awaiting_approval"`
	StartedAt       string `json:"started_at"`
	UpdatedAt       string `json:"updated_at"`
	// recovery marker fields written by the TUI itself
	State string `json:"state"`
	TS    string `json:"ts"`
}

func (p pipelineState) isTaffy() bool {
	return p.Taffy != ""
}

// isStale returns true if the pipeline claims RUNNING but hasn't been updated
// in stalePipelineThreshold — meaning the agent likely died without writing DONE/FAILED.
func (p pipelineState) isStale() bool {
	if strings.ToUpper(p.Status) != "RUNNING" {
		return false
	}
	if p.UpdatedAt == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339, p.UpdatedAt)
	if err != nil {
		return false
	}
	return time.Since(t) > stalePipelineThreshold
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
// or "" if no approval is waiting. The pipeline-runner skill writes this file at
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
