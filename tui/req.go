package main

import (
"fmt"
"os"
"os/exec"
"path/filepath"
"regexp"
"strings"
"time"
"unicode"

"github.com/charmbracelet/bubbles/textarea"
tea "github.com/charmbracelet/bubbletea"
"github.com/charmbracelet/lipgloss"
)

// ─── Requirements command ─────────────────────────────────────────────────────

// availableAITools returns tools that support arbitrary prompt → text generation.
// gh copilot (explain/suggest) is excluded — it only handles shell commands.
func availableAITools(tools Tools) []aiOption {
var opts []aiOption
if tools.Claude != "" {
opts = append(opts, aiOption{"Claude Code", "claude", tools.Claude})
}
if tools.Copilot != "" {
opts = append(opts, aiOption{"GitHub Copilot", "copilot", tools.Copilot})
}
if tools.OpenCode != "" {
opts = append(opts, aiOption{"OpenCode", "opencode", tools.OpenCode})
}
return opts
}

func runReq(tools Tools) error {
opts := availableAITools(tools)
if len(opts) == 0 {
return fmt.Errorf("no AI tool available — install claude, copilot, or opencode")
}
m := newReqModel(tools)
p := tea.NewProgram(m, tea.WithAltScreen())
_, err := p.Run()
return err
}

// ─── Model ────────────────────────────────────────────────────────────────────

type reqStep int

const (
reqStepPickAI   reqStep = iota // shown only when >1 AI tool available
reqStepEdit                    // textarea for requirements
reqStepSending                 // waiting for AI
reqStepStories                 // list of converted stories
reqStepViewStory               // full content of one story
)

type gherkinStory struct {
title   string
gherkin string
savedTo string
}

type reqModel struct {
tools        Tools
aiOptions    []aiOption
selectedAI   aiOption
aiCursor     int
textarea     textarea.Model
step         reqStep
err          error
width        int
height       int
logoFrame    int
logoDone     bool
stories      []gherkinStory
storyCursor  int
viewingIdx   int
scrollOffset int
}

type reqDoneMsg struct {
stories []gherkinStory
err     error
}

func newReqModel(tools Tools) *reqModel {
ta := textarea.New()
ta.Placeholder = "Describe what you need…\n\nExamples:\n  As a user I want to export filtered results as CSV\n  so that I can share them with stakeholders offline.\n\n  The export must respect active filters.\n  Only the last 90 days of data should be included.\n  The file must be downloadable immediately."
ta.Focus()
ta.SetWidth(80)
ta.SetHeight(20)
ta.ShowLineNumbers = false

opts := availableAITools(tools)
firstStep := reqStepEdit
if len(opts) > 1 {
firstStep = reqStepPickAI
}

var selected aiOption
if len(opts) > 0 {
selected = opts[0]
}

return &reqModel{
tools:      tools,
aiOptions:  opts,
selectedAI: selected,
textarea:   ta,
step:       firstStep,
}
}

func (m *reqModel) Init() tea.Cmd {
return tea.Batch(textarea.Blink, logoTick(0))
}

func (m *reqModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
switch msg := msg.(type) {
case tea.WindowSizeMsg:
m.width = msg.Width
m.height = msg.Height
m.textarea.SetWidth(msg.Width - 4)
m.textarea.SetHeight(msg.Height - 12)
// Clamp scroll offset after resize
if m.step == reqStepViewStory && len(m.stories) > 0 {
lines := strings.Split(m.stories[m.viewingIdx].gherkin, "\n")
maxScroll := len(lines) - m.visibleLines()
if maxScroll < 0 {
maxScroll = 0
}
if m.scrollOffset > maxScroll {
m.scrollOffset = maxScroll
}
}

case logoTickMsg:
if !m.logoDone {
m.logoFrame++
if m.logoFrame >= logoFrameCount {
m.logoDone = true
} else {
return m, logoTick(m.logoFrame)
}
}

case tea.KeyMsg:
if !m.logoDone {
return m, nil
}

switch m.step {
case reqStepPickAI:
switch msg.String() {
case "up", "k":
if m.aiCursor > 0 {
m.aiCursor--
}
case "down", "j":
if m.aiCursor < len(m.aiOptions)-1 {
m.aiCursor++
}
case "enter", " ":
m.selectedAI = m.aiOptions[m.aiCursor]
m.step = reqStepEdit
return m, textarea.Blink
case "ctrl+c", "q", "esc":
return m, tea.Quit
}
return m, nil

case reqStepEdit:
switch msg.Type {
case tea.KeyCtrlD:
req := strings.TrimSpace(m.textarea.Value())
if req == "" {
return m, nil
}
m.step = reqStepSending
return m, m.convert(req)
case tea.KeyCtrlC, tea.KeyEsc:
return m, tea.Quit
}

case reqStepStories:
n := len(m.stories)
switch msg.String() {
case "up", "k":
if m.storyCursor > 0 {
m.storyCursor--
}
case "down", "j":
if m.storyCursor < n-1 {
m.storyCursor++
}
case "enter", " ":
m.viewingIdx = m.storyCursor
m.scrollOffset = 0
m.step = reqStepViewStory
case "e":
// Back to editor to refine
m.textarea.SetValue("")
m.step = reqStepEdit
return m, textarea.Blink
case "q", "esc":
return m, tea.Quit
}
return m, nil

case reqStepViewStory:
story := m.stories[m.viewingIdx]
lines := strings.Split(story.gherkin, "\n")
visible := m.visibleLines()
maxScroll := len(lines) - visible
if maxScroll < 0 {
maxScroll = 0
}
switch msg.String() {
case "up", "k":
if m.scrollOffset > 0 {
m.scrollOffset--
}
case "down", "j":
if m.scrollOffset < maxScroll {
m.scrollOffset++
}
case "b", "esc":
m.scrollOffset = 0
m.step = reqStepStories
case "q":
return m, tea.Quit
}
return m, nil
}

case reqDoneMsg:
if msg.err != nil {
m.err = msg.err
m.step = reqStepStories // show error in stories view
return m, nil
}
m.stories = msg.stories
m.storyCursor = 0
m.step = reqStepStories
}

if m.step == reqStepEdit {
var cmd tea.Cmd
m.textarea, cmd = m.textarea.Update(msg)
return m, cmd
}
return m, nil
}

func (m *reqModel) visibleLines() int {
// Reserve: logo (~9) + title (2) + footer (3)
v := m.height - 14
if v < 5 {
v = 5
}
return v
}

func (m *reqModel) View() string {
t := tokyoNight()
var hdr string
if m.logoDone {
hdr = logo()
} else {
hdr = logoAnimFrame(m.logoFrame)
}

cursor := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("❯")

switch m.step {
case reqStepPickAI:
var sb strings.Builder
title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("Select AI tool")
sb.WriteString("  " + title + "\n")
sb.WriteString(lipgloss.NewStyle().Foreground(t.Muted).Render("  "+strings.Repeat("─", 40)) + "\n\n")
for i, opt := range m.aiOptions {
kind := lipgloss.NewStyle().Foreground(t.Muted).Render("  " + opt.kind)
if i == m.aiCursor {
label := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render(fmt.Sprintf("%-20s", opt.label))
sb.WriteString("  " + cursor + " " + label + kind + "\n")
} else {
label := lipgloss.NewStyle().Foreground(t.Foreground).Render(fmt.Sprintf("%-20s", opt.label))
sb.WriteString("     " + label + kind + "\n")
}
}
sb.WriteString("\n" + lipgloss.NewStyle().Foreground(t.Muted).Render("  ↑/↓ Navigate   Enter Select   q Quit") + "\n")
return hdr + sb.String()

case reqStepEdit:
ai := lipgloss.NewStyle().Foreground(t.Muted).Render("  AI: " + m.selectedAI.label)
help := lipgloss.NewStyle().Foreground(t.Muted).Render(
"  Type your requirements below — describe a story or a whole epic.\n" +
"  Ctrl+D to convert → Gherkin   Esc to go back")
box := lipgloss.NewStyle().
Border(lipgloss.RoundedBorder()).
BorderForeground(t.Primary).
Padding(0, 1).
Render(m.textarea.View())
return hdr + ai + "\n\n" + help + "\n\n" + box + "\n"

case reqStepSending:
ai := lipgloss.NewStyle().Foreground(t.Muted).Render("  AI: " + m.selectedAI.label)
return hdr + ai + "\n\n" +
lipgloss.NewStyle().Foreground(t.Accent).Render("  Converting to Gherkin…") + "\n"

case reqStepStories:
if m.err != nil {
return hdr + "\n\n" +
lipgloss.NewStyle().Foreground(t.Error).Render("  ✗ "+m.err.Error()) + "\n\n" +
lipgloss.NewStyle().Foreground(t.Muted).Render("  e edit again   q back to menu") + "\n"
}
return hdr + m.storiesView(t, cursor)

case reqStepViewStory:
return hdr + m.storyDetailView(t)
}
return ""
}

func (m *reqModel) storiesView(t Theme, cursor string) string {
var sb strings.Builder
count := len(m.stories)
noun := "story"
if count != 1 {
noun = "stories"
}
title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).
Render(fmt.Sprintf("  %d %s generated", count, noun))
sb.WriteString(title + "\n")
sb.WriteString(lipgloss.NewStyle().Foreground(t.Muted).Render("  "+strings.Repeat("─", 54)) + "\n\n")

for i, s := range m.stories {
var savedNote string
if s.savedTo != "" {
savedNote = lipgloss.NewStyle().Foreground(t.Muted).Render("  " + s.savedTo)
}
if i == m.storyCursor {
label := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render(s.title)
sb.WriteString("  " + cursor + " " + label + savedNote + "\n")
} else {
label := lipgloss.NewStyle().Foreground(t.Foreground).Render(s.title)
sb.WriteString("     " + label + savedNote + "\n")
}
}

sb.WriteString("\n" + lipgloss.NewStyle().Foreground(t.Muted).Render(
"  ↑/↓ Navigate   Enter View   e Edit again   q Back to menu") + "\n")
return sb.String()
}

func (m *reqModel) storyDetailView(t Theme) string {
story := m.stories[m.viewingIdx]
var sb strings.Builder

title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("  " + story.title)
sb.WriteString(title + "\n")
if story.savedTo != "" {
saved := lipgloss.NewStyle().Foreground(t.Success).Render("  ✓ " + story.savedTo)
sb.WriteString(saved + "\n")
}
sb.WriteString(lipgloss.NewStyle().Foreground(t.Muted).Render("  "+strings.Repeat("─", 54)) + "\n\n")

lines := strings.Split(story.gherkin, "\n")
visible := m.visibleLines()
end := m.scrollOffset + visible
if end > len(lines) {
end = len(lines)
}
window := lines[m.scrollOffset:end]

for _, l := range window {
sb.WriteString("  " + lipgloss.NewStyle().Foreground(t.Foreground).Render(l) + "\n")
}

// Scroll indicator
total := len(lines)
scrollPct := ""
if total > visible {
pct := (m.scrollOffset * 100) / (total - visible)
scrollPct = fmt.Sprintf(" (%d%%)", pct)
}
sb.WriteString("\n" + lipgloss.NewStyle().Foreground(t.Muted).Render(
fmt.Sprintf("  ↑/↓ j/k Scroll%s   b/Esc Back to list   q Quit", scrollPct)) + "\n")
return sb.String()
}

func (m *reqModel) convert(requirements string) tea.Cmd {
ai := m.selectedAI
return func() tea.Msg {
gherkin, err := convertToGherkin(requirements, ai)
if err != nil {
return reqDoneMsg{err: err}
}
stories := parseStories(gherkin)
for i := range stories {
savedTo, saveErr := saveStory(stories[i].title, stories[i].gherkin, i+1)
if saveErr == nil {
stories[i].savedTo = savedTo
}
}
return reqDoneMsg{stories: stories}
}
}

// ─── AI conversion ────────────────────────────────────────────────────────────

const gherkinPrompt = `You are a Gherkin authoring expert. Analyze the requirements below.

Determine if this describes:
  A) A single user story → output ONE story block
  B) An epic (multiple related stories) → split into separate stories, one block each

For EACH story use this exact format:
=== STORY: <Concise Story Title> ===
Feature: <feature name>

  @story @priority:medium
  Scenario: <happy path name>
    Given ...
    When ...
    Then ...

  Scenario: <edge case name>
    Given ...
    When ...
    Then ...

Rules:
- Use Feature, Background (if needed), Scenario, Given/When/Then/And/But
- Plain language that non-technical stakeholders can understand
- Separate scenarios for happy path and key edge cases
- Output ONLY the === STORY: === blocks — no explanation, no prose

Requirements:
%s`

func convertToGherkin(requirements string, ai aiOption) (string, error) {
prompt := fmt.Sprintf(gherkinPrompt, requirements)
out, err := invokeAI(ai, prompt)
if err != nil {
return "", fmt.Errorf("%s: %w", ai.label, err)
}
return strings.TrimSpace(string(out)), nil
}

// invokeAI runs the selected AI tool and returns its stdout output.
// The prompt is passed as a positional argument (not stdin) to avoid inheriting
// the Bubble Tea raw-mode terminal state in the subprocess.
func invokeAI(ai aiOption, prompt string) ([]byte, error) {
var cmd *exec.Cmd
switch ai.kind {
case "claude":
// -p = --print (non-interactive), prompt is positional arg.
// --output-format text suppresses JSON wrappers.
// --no-session-persistence avoids writing session files.
cmd = exec.Command(ai.path, "-p", "--output-format", "text", "--no-session-persistence", prompt)
case "copilot":
// --prompt <text> is the non-interactive flag for the GitHub Copilot CLI.
cmd = exec.Command(ai.path, "--prompt", prompt)
case "opencode":
cmd = exec.Command(ai.path, "run")
cmd.Stdin = strings.NewReader(prompt)
default:
return nil, fmt.Errorf("unsupported AI tool: %s", ai.kind)
}

out, err := cmd.CombinedOutput()
if err != nil {
msg := strings.TrimSpace(string(out))
if msg == "" {
msg = err.Error()
}
return nil, fmt.Errorf("%s", msg)
}
return out, nil
}

// ─── Story parsing ────────────────────────────────────────────────────────────

var storyHeaderRe = regexp.MustCompile(`(?m)^=== STORY: (.+?) ===$`)

// parseStories splits AI output into individual gherkinStory values.
// If no === STORY: === delimiters are found the whole output is treated as one story.
func parseStories(output string) []gherkinStory {
matches := storyHeaderRe.FindAllStringIndex(output, -1)
if len(matches) == 0 {
return []gherkinStory{{
title:   extractFeatureTitle(output),
gherkin: stripFences(strings.TrimSpace(output)),
}}
}

var stories []gherkinStory
for i, match := range matches {
titleMatch := storyHeaderRe.FindStringSubmatch(output[match[0]:match[1]])
title := strings.TrimSpace(titleMatch[1])

start := match[1]
if start < len(output) && output[start] == '\n' {
start++
}
var content string
if i+1 < len(matches) {
content = output[start:matches[i+1][0]]
} else {
content = output[start:]
}
stories = append(stories, gherkinStory{
title:   title,
gherkin: stripFences(strings.TrimSpace(content)),
})
}
return stories
}

// extractFeatureTitle pulls the Feature: name from gherkin text, or returns "Untitled".
func extractFeatureTitle(s string) string {
for _, l := range strings.Split(s, "\n") {
l = strings.TrimSpace(l)
if strings.HasPrefix(l, "Feature:") {
t := strings.TrimSpace(strings.TrimPrefix(l, "Feature:"))
if t != "" {
return t
}
}
}
return "Untitled"
}

func stripFences(s string) string {
lines := strings.Split(s, "\n")
var out []string
for i, l := range lines {
trimmed := strings.TrimSpace(l)
if i == 0 && (trimmed == "```gherkin" || trimmed == "```") {
continue
}
if i == len(lines)-1 && trimmed == "```" {
continue
}
out = append(out, l)
}
return strings.Join(out, "\n")
}

// ─── Story file writer ────────────────────────────────────────────────────────

func saveStory(title, gherkin string, idx int) (string, error) {
if err := os.MkdirAll("docs/stories", 0755); err != nil {
return "", err
}

slug := slugify(title)
ts := time.Now().Format("20060102150405")
filename := fmt.Sprintf("docs/stories/%s-%s-%04d.md", slug, ts, idx)

displayTitle := title
if len(displayTitle) > 80 {
displayTitle = displayTitle[:80]
}

content := fmt.Sprintf(`---
id: "%s-%04d"
title: "%s"
epic: "%s"
priority: "medium"
ui: false
adr_required: false
milestone: null
labels:
  - "type:feature"
  - "priority:medium"
  - "phase:discover"
issue_number: null
issue_url: null
created_at: "%s"
---

# %s

## Scenarios

`+"```gherkin\n%s\n```"+`

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber scenarios green
- [ ] ADRs linked where required

## ADR Links

(populated by adr-author agent)
`,
slug, idx, displayTitle, slug, time.Now().Format(time.RFC3339),
displayTitle, gherkin,
)

return filepath.Clean(filename), os.WriteFile(filename, []byte(content), 0644)
}

func slugify(s string) string {
s = strings.ToLower(s)
re := regexp.MustCompile(`[^a-z0-9]+`)
s = re.ReplaceAllString(s, "-")
s = strings.Trim(s, "-")
runes := []rune(s)
if len(runes) > 40 {
runes = runes[:40]
for len(runes) > 0 && !unicode.IsLetter(runes[len(runes)-1]) && !unicode.IsDigit(runes[len(runes)-1]) {
runes = runes[:len(runes)-1]
}
}
return string(runes)
}
