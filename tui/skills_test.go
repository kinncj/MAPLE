package main

import "testing"

func TestParseSkillsOutput_Single(t *testing.T) {
	input := `vercel-labs/agent-browser@dogfood 21.1K installs
└ https://skills.sh/vercel-labs/agent-browser/dogfood
`
	rows := parseSkillsOutput(input)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].pkg != "vercel-labs/agent-browser@dogfood" {
		t.Errorf("pkg = %q", rows[0].pkg)
	}
	if rows[0].installs != "21.1K" {
		t.Errorf("installs = %q", rows[0].installs)
	}
	if rows[0].url != "https://skills.sh/vercel-labs/agent-browser/dogfood" {
		t.Errorf("url = %q", rows[0].url)
	}
}

func TestParseSkillsOutput_Multiple(t *testing.T) {
	input := `vercel-labs/agent-browser@dogfood 21.1K installs
└ https://skills.sh/vercel-labs/agent-browser/dogfood
anthropic/claude-skills@tdd 4.2K installs
└ https://skills.sh/anthropic/claude-skills/tdd
microsoft/copilot-skills@review 1.8K installs
`
	rows := parseSkillsOutput(input)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d: %+v", len(rows), rows)
	}
	if rows[1].pkg != "anthropic/claude-skills@tdd" {
		t.Errorf("row[1].pkg = %q", rows[1].pkg)
	}
	if rows[1].installs != "4.2K" {
		t.Errorf("row[1].installs = %q", rows[1].installs)
	}
	if rows[2].url != "" {
		t.Errorf("row[2].url should be empty, got %q", rows[2].url)
	}
}

func TestParseSkillsOutput_CommaInstalls(t *testing.T) {
	input := `kinncj/maple@pipeline-runner 1,234 installs
└ https://skills.sh/kinncj/maple/pipeline-runner
`
	rows := parseSkillsOutput(input)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].installs != "1,234" {
		t.Errorf("installs = %q", rows[0].installs)
	}
}

func TestParseSkillsOutput_Empty(t *testing.T) {
	rows := parseSkillsOutput("")
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

func TestParseSkillsOutput_NoMatches(t *testing.T) {
	input := "No skills found matching your query.\n"
	rows := parseSkillsOutput(input)
	if len(rows) != 0 {
		t.Errorf("expected 0 rows for non-matching output, got %d", len(rows))
	}
}

func TestParseSkillsOutput_SkipsNonSkillLines(t *testing.T) {
	input := `SKILLS DIRECTORY — skills.sh
════════════════════════

vercel-labs/agent-browser@dogfood 21.1K installs
└ https://skills.sh/vercel-labs/agent-browser/dogfood

notaskill 100 installs
`
	rows := parseSkillsOutput(input)
	// only the one with "/" in pkg field should match
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d: %+v", len(rows), rows)
	}
	if rows[0].pkg != "vercel-labs/agent-browser@dogfood" {
		t.Errorf("pkg = %q", rows[0].pkg)
	}
}

func TestParseInstalledJSON_Array(t *testing.T) {
	s := `[{"name":"tdd-workflow","package":"kinncj/maple@tdd-workflow","agent":"claude-code"},{"name":"pipeline-runner","pkg":"kinncj/maple@pipeline-runner","agent":"claude-code"}]`
	rows := parseInstalledJSON(s, "project")
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].name != "tdd-workflow" {
		t.Errorf("name = %q", rows[0].name)
	}
	if rows[0].pkg != "kinncj/maple@tdd-workflow" {
		t.Errorf("pkg = %q", rows[0].pkg)
	}
	if rows[0].scope != "project" {
		t.Errorf("scope = %q", rows[0].scope)
	}
	// second row uses "pkg" field instead of "package"
	if rows[1].pkg != "kinncj/maple@pipeline-runner" {
		t.Errorf("row[1].pkg = %q", rows[1].pkg)
	}
}

func TestParseInstalledJSON_Map(t *testing.T) {
	s := `{"claude-code":[{"name":"tdd-workflow","package":"kinncj/maple@tdd-workflow"}],"cursor":[{"name":"playwright-cli","package":"kinncj/maple@playwright-cli"}]}`
	rows := parseInstalledJSON(s, "global")
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	names := map[string]bool{}
	for _, r := range rows {
		names[r.name] = true
		if r.scope != "global" {
			t.Errorf("scope = %q", r.scope)
		}
	}
	if !names["tdd-workflow"] || !names["playwright-cli"] {
		t.Errorf("missing expected skills: %+v", names)
	}
}

func TestParseInstalledJSON_Invalid(t *testing.T) {
	rows := parseInstalledJSON("not json at all", "project")
	if len(rows) != 0 {
		t.Errorf("expected 0 rows for invalid JSON, got %d", len(rows))
	}
}

func TestParseInstalledJSON_Empty(t *testing.T) {
	rows := parseInstalledJSON("[]", "project")
	if len(rows) != 0 {
		t.Errorf("expected 0 rows for empty array, got %d", len(rows))
	}
}

func TestParseInstalledText_AgentHeaders(t *testing.T) {
	s := `Project skills (claude-code):
  tdd-workflow   kinncj/maple@tdd-workflow
  pipeline-runner kinncj/maple@pipeline-runner
Project skills (cursor):
  playwright-cli kinncj/maple@playwright-cli
`
	rows := parseInstalledText(s, "project")
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d: %+v", len(rows), rows)
	}
	if rows[0].name != "tdd-workflow" {
		t.Errorf("rows[0].name = %q", rows[0].name)
	}
	if rows[0].agent != "claude-code" {
		t.Errorf("rows[0].agent = %q", rows[0].agent)
	}
	if rows[2].agent != "cursor" {
		t.Errorf("rows[2].agent = %q", rows[2].agent)
	}
}

func TestParseInstalledText_NoHeaders(t *testing.T) {
	s := `tdd-workflow   kinncj/maple@tdd-workflow
pipeline-runner kinncj/maple@pipeline-runner
`
	rows := parseInstalledText(s, "global")
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].name != "tdd-workflow" {
		t.Errorf("rows[0].name = %q", rows[0].name)
	}
	if rows[0].scope != "global" {
		t.Errorf("rows[0].scope = %q", rows[0].scope)
	}
	// agent is empty when no header
	if rows[0].agent != "" {
		t.Errorf("rows[0].agent should be empty, got %q", rows[0].agent)
	}
}

func TestParseInstalledText_Empty(t *testing.T) {
	rows := parseInstalledText("", "project")
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

func TestStripANSI(t *testing.T) {
	cases := []struct{ in, want string }{
		{"\x1b[32mfoo\x1b[0m", "foo"},
		{"\x1b[1;31mERROR\x1b[0m: bad", "ERROR: bad"},
		{"no ansi here", "no ansi here"},
		{"", ""},
	}
	for _, c := range cases {
		got := stripANSI(c.in)
		if got != c.want {
			t.Errorf("stripANSI(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
