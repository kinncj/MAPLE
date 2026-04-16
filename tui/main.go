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

	if len(args) == 0 {
		printHelp()
		return
	}

	tools := Detect()

	switch args[0] {
	case "--version", "-v", "version":
		fmt.Println("squad " + version)

	case "--help", "-h", "help":
		printHelp()

	case "init":
		force := contains(args[1:], "--force")
		_ = force // copyFile already skips existing; --force would overwrite — future flag
		tplDir := templateDir()
		if err := runInit(tools, tplDir); err != nil {
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
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(logo())
	fmt.Printf(`squad %s — AI-Squad initialiser and project helper

Usage:
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
//  2. <binary>/../template/  (local dev / installed alongside repo)
//  3. ~/.ai-squad/template/
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
	// Try relative to cwd (e.g. running from repo root)
	if stat, err := os.Stat("template"); err == nil && stat.IsDir() {
		abs, _ := filepath.Abs("template")
		return abs
	}
	// Fallback: ~/.ai-squad/template
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
