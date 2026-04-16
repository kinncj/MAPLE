package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Println("squad v3.5.0")
			return
		case "--help", "-h":
			printHelp()
			return
		case ":resume":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "usage: squad :resume <superpower-name>")
				os.Exit(1)
			}
			fmt.Printf("Resuming superpower: %s\n", os.Args[2])
			// TODO: load .claude/state/squad.json and resume
			return
		}
	}

	app := newApp()
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
