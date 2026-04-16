package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const version = "v3.5.0"

func main() {
	args := os.Args[1:]
	tools := Detect()
	tplDir := templateDir()

	// No args → interactive TUI menu
	if len(args) == 0 {
		result := runMenu(tools, tplDir)
		switch result.action {
		case menuInit:
			if err := runInit(tools, tplDir, false); err != nil {
				fatalf("init: %v", err)
			}
		case menuUpdate:
			if err := runInit(tools, tplDir, true); err != nil {
				fatalf("update: %v", err)
			}
		case menuReq:
			if err := runReq(tools); err != nil {
				fatalf("req: %v", err)
			}
		case menuLabels:
			if err := runLabels(tools.GH); err != nil {
				fatalf("labels: %v", err)
			}
		case menuProject:
			if err := runProject(tools.GH); err != nil {
				fatalf("project: %v", err)
			}
		case menuHelp:
			// Help is handled inline in the menu; this branch is unreachable
			// but kept for completeness.
			printHelpStatic()
		}
		return
	}

	// Subcommand mode
	switch args[0] {
	case "--version", "-v", "version":
		fmt.Println("squad " + version)

	case "--help", "-h", "help":
		printHelp()

	case "init":
		force := contains(args[1:], "--force")
		if err := runInit(tools, tplDir, force); err != nil {
			fatalf("init: %v", err)
		}

	case "req":
		if err := runReq(tools); err != nil {
			fatalf("req: %v", err)
		}

	case "labels":
		if err := runLabels(tools.GH); err != nil {
			fatalf("labels: %v", err)
		}

	case "project":
		if err := runProject(tools.GH); err != nil {
			fatalf("project: %v", err)
		}

	default:
		fmt.Fprintf(os.Stderr, "squad: unknown command %q\n\n", args[0])
		printHelpStatic()
		os.Exit(1)
	}
}

func printHelp() {
	printLogoAnimated()
	printHelpText()
}

func printHelpStatic() {
	printLogoStatic()
	printHelpText()
}

func printHelpText() {
	fmt.Printf(`squad %s — AI-Squad initialiser and project helper

Usage:
  squad                   Launch interactive menu
  squad init              Set up AI-Squad in the current directory
  squad init --force      Overwrite existing files
  squad req               Write requirements → generate Gherkin story
  squad labels            Bootstrap GitHub label set
  squad project           Create GitHub Project v2

  squad --version         Print version
  squad --help            Show this help
`, version)
}

// templateDir resolves the template source directory.
// Resolution order:
//  1. AI_SQUAD_TEMPLATE env var
//  2. <binary>/../template/  (installed alongside repo)
//  3. ./template/  (cwd fallback — running from repo root)
//  4. ~/.ai-squad/template/
func templateDir() string {
	if v := os.Getenv("AI_SQUAD_TEMPLATE"); v != "" {
		return v
	}
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "..", "template")
		if stat, err := os.Stat(candidate); err == nil && stat.IsDir() {
			abs, _ := filepath.Abs(candidate)
			return abs
		}
	}
	if stat, err := os.Stat("template"); err == nil && stat.IsDir() {
		abs, _ := filepath.Abs("template")
		return abs
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ai-squad", "template")
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "squad: "+format+"\n", args...)
	if runtime.GOOS == "windows" {
		os.Exit(1)
	}
	os.Exit(1)
}
