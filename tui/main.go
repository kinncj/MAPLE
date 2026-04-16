package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	noAnimate := false
	args := os.Args[1:]

	// Filter --no-animate before command dispatch
	filtered := args[:0]
	for _, a := range args {
		if a == "--no-animate" {
			noAnimate = true
		} else {
			filtered = append(filtered, a)
		}
	}
	args = filtered

	if len(args) > 0 {
		switch args[0] {
		case "--version", "-v":
			fmt.Println("squad v3.5.0")
			return
		case "--help", "-h":
			printHelp()
			return
		case ":resume":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "usage: squad :resume <superpower-name>")
				os.Exit(1)
			}
			fmt.Printf("Resuming superpower: %s\n", args[1])
			// TODO: load .claude/state/squad.json and resume
			return
		}
	}

	app := newApp()
	app.noAnimate = noAnimate
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "squad: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`squad — AI-Squad TUI

Usage:
  squad                   Launch interactive dashboard
  squad :resume <name>    Resume a paused superpower
  squad --version         Print version
  squad --help            Show this help
  squad --no-animate      Skip splash animation (useful on slow terminals / SSH)

Keybindings (in dashboard):
  Tab / Shift+Tab   Cycle panes
  j / k             Move down / up
  Enter             Open detail view
  s                 Stories pane
  a                 Agents pane
  p                 PRs pane
  q                 QA pane
  d                 Design pane
  l                 Logs pane
  n                 New story / spike / ADR
  /                 Search
  :                 Command mode
  F                 Fire superpower
  ?                 Help overlay
  Ctrl+c            Quit

`)
}
