package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

var version = "dev" // overridden at build time: -ldflags "-X main.version=vX.Y.Z"

func main() {
	args := os.Args[1:]
	tplFS, _ := resolveTemplateFS()

	noAnimate := contains(args, "--no-animate")
	// strip --no-animate so subcommand parsing below is unaffected
	var cleanArgs []string
	for _, a := range args {
		if a != "--no-animate" {
			cleanArgs = append(cleanArgs, a)
		}
	}
	args = cleanArgs

	// No args → interactive TUI (boot check + dashboard if initialized, menu otherwise)
	if len(args) == 0 {
		runInteractive(tplFS, noAnimate)
		return
	}

	tools := Detect()

	// Subcommand mode
	switch args[0] {
	case "--version", "-v", "version":
		fmt.Println("squad " + version)

	case "--help", "-h", "help":
		printHelp()

	case "init":
		force := contains(args[1:], "--force")
		if err := runInit(tools, tplFS, force); err != nil {
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

func runInteractive(fsys fs.FS, noAnimate bool) {
	if !isStdinTTY() {
		printHelpStatic()
		return
	}

	// If project is initialized, run boot check then launch dashboard.
	if _, err := os.Stat("project.config.yaml"); err == nil {
		t, ok := runBoot()
		if ok {
			if err := runDashboard(t, noAnimate); err != nil {
				fmt.Fprintf(os.Stderr, "dashboard: %v\n", err)
			}
		}
		return
	}

	// Not yet initialized — show the setup menu.
	for {
		tools := Detect()
		result := runMenu(tools, fsys)
		switch result.action {
		case menuQuit:
			return
		case menuInit:
			if err := runInitFromMenu(tools, fsys, false); err != nil {
				fmt.Fprintf(os.Stderr, "init: %v\n", err)
			}
		case menuUpdate:
			if err := runInitFromMenu(tools, fsys, true); err != nil {
				fmt.Fprintf(os.Stderr, "update: %v\n", err)
			}
		case menuReq:
			if err := runReq(tools); err != nil {
				fmt.Fprintf(os.Stderr, "req: %v\n", err)
			}
		case menuLabels:
			if err := runLabels(tools.GH); err != nil {
				fmt.Fprintf(os.Stderr, "labels: %v\n", err)
			}
		case menuProject:
			if err := runProject(tools.GH); err != nil {
				fmt.Fprintf(os.Stderr, "project: %v\n", err)
			}
		case menuHelp:
			// handled inline in menu
		}
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

  squad --no-animate      Skip logo animations (SSH / slow terminals)
  squad --version         Print version
  squad --help            Show this help
`, version)
}

// resolveTemplateFS resolves the template source as an fs.FS.
// Resolution order:
//  1. AI_SQUAD_TEMPLATE env var (resolved to absolute path, uses OS filesystem)
//  2. <binary_dir>/template/ if it exists on disk (dev checkout)
//  3. ./template/ in cwd if exists (running from repo root in dev)
//  4. Embedded FS (always works for released binaries)
func resolveTemplateFS() (fs.FS, string) {
	if v := os.Getenv("AI_SQUAD_TEMPLATE"); v != "" {
		if abs, err := filepath.Abs(v); err == nil {
			v = abs
		}
		return os.DirFS(v), v
	}
	exe, _ := os.Executable()
	for _, c := range []string{
		filepath.Join(filepath.Dir(exe), "template"),
		"template",
	} {
		if stat, err := os.Stat(c); err == nil && stat.IsDir() {
			abs, _ := filepath.Abs(c)
			return os.DirFS(abs), abs
		}
	}
	sub, _ := fs.Sub(embeddedTemplate, "template")
	return sub, "(embedded)"
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
