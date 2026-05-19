package main

import "strings"

func harnessInstructionMarkdowns(harness string) []string {
	base := []string{
		"AGENTS.md",
		".github/copilot-instructions.md",
		".github/instructions/stories.instructions.md",
	}
	switch harness {
	case "opencode":
		return append([]string{"OPENCODE.md"}, base...)
	case "cursor":
		return append([]string{"CURSOR.md"}, base...)
	default:
		return append([]string{"CLAUDE.md"}, base...)
	}
}

func governanceBootstrapBlock(harness string) string {
	files := harnessInstructionMarkdowns(harness)
	var sb strings.Builder
	sb.WriteString("\n<maple-governance-bootstrap>\n")
	sb.WriteString("Before running any TAFFY stage, read and enforce these markdown files in order:\n")
	for _, f := range files {
		sb.WriteString("- " + f + "\n")
	}
	sb.WriteString("Treat them as mandatory constraints for every delegated skill/agent call.\n")
	sb.WriteString("Runtime code and tests must never import from docs/, .github/, or .claude/ paths.\n")
	sb.WriteString("Copying or adapting artifact content from docs into app/test code is allowed; path-based imports/references are not.\n")
	sb.WriteString("</maple-governance-bootstrap>\n")
	return sb.String()
}
