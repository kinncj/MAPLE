package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCopilotWorkspace_Flat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "workspace.yaml")
	body := `id: abc-123
cwd: /tmp/foo
summary: Implement Features
summary_count: 0
created_at: 2026-04-22T17:21:06.975Z
updated_at: 2026-04-22T17:21:31.926Z
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	m := parseCopilotWorkspace(path)
	if m["id"] != "abc-123" {
		t.Errorf("id = %q", m["id"])
	}
	if m["cwd"] != "/tmp/foo" {
		t.Errorf("cwd = %q", m["cwd"])
	}
	if m["summary"] != "Implement Features" {
		t.Errorf("summary = %q", m["summary"])
	}
	if m["updated_at"] != "2026-04-22T17:21:31.926Z" {
		t.Errorf("updated_at = %q", m["updated_at"])
	}
}

func TestParseCopilotWorkspace_BlockScalar(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "workspace.yaml")
	body := "id: abc\n" +
		"cwd: /tmp\n" +
		"summary: |-\n" +
		"  First line\n" +
		"  Second line\n" +
		"\n" +
		"  After blank\n" +
		"updated_at: 2026-04-22T00:00:00.000Z\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	m := parseCopilotWorkspace(path)
	if m["summary"] != "First line\nSecond line\n\nAfter blank" {
		t.Errorf("summary = %q", m["summary"])
	}
	if m["updated_at"] != "2026-04-22T00:00:00.000Z" {
		t.Errorf("updated_at = %q", m["updated_at"])
	}
}

func TestParseCopilotWorkspace_FoldedBlockScalar(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "workspace.yaml")
	body := "summary: >-\n" +
		"  First line\n" +
		"  Second line\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	m := parseCopilotWorkspace(path)
	if m["summary"] != "First line Second line" {
		t.Errorf("summary = %q", m["summary"])
	}
}

func TestLoadCopilotSessions_FiltersByCwd(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	mkSession := func(id, cwd, summary, updated string) {
		dir := filepath.Join(fakeHome, ".copilot", "session-state", id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		body := "id: " + id + "\n" +
			"cwd: " + cwd + "\n" +
			"summary: " + summary + "\n" +
			"updated_at: " + updated + "\n"
		if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		events := `{"type":"tool.execution_start","data":{"toolName":"glob"}}` + "\n" +
			`{"type":"tool.execution_start","data":{"toolName":"bash"}}` + "\n"
		if err := os.WriteFile(filepath.Join(dir, "events.jsonl"), []byte(events), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mkSession("aaa", "/match/me", "Match", "2026-04-22T10:00:00.000Z")
	mkSession("bbb", "/match/me", "Older", "2026-04-20T10:00:00.000Z")
	mkSession("ccc", "/other", "Skip", "2026-04-22T10:00:00.000Z")

	rows := loadCopilotSessions("/match/me")
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d: %+v", len(rows), rows)
	}
	if rows[0].id != "aaa" {
		t.Errorf("expected newest first, got id=%q", rows[0].id)
	}
	if rows[0].source != "copilot" {
		t.Errorf("source = %q", rows[0].source)
	}
	if rows[0].title != "Match" {
		t.Errorf("title = %q", rows[0].title)
	}
	if rows[0].toolCount != 2 {
		t.Errorf("toolCount = %d", rows[0].toolCount)
	}
}
