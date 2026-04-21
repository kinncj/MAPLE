package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// errNoNewTerminal is returned when no supported new-tab/window mechanism is available.
var errNoNewTerminal = errors.New("no supported new-terminal mechanism found")

// spawnInNewTerminal opens args in a new terminal tab or window, keeping the
// current terminal (and maple) alive. Detection order:
//
//  1. zellij     — action new-tab (ZELLIJ env var)
//  2. tmux       — new-window (TMUX env var)
//  3. GNU screen — new window via STY
//  4. WezTerm    — cli spawn
//  5. Kitty      — @launch --type=tab
//  6. macOS      — iTerm2 or Terminal.app via osascript
//  7. Linux      — first found: x-terminal-emulator, gnome-terminal, konsole, xfce4-terminal, xterm
//  8. Windows    — Windows Terminal (wt), then cmd start
//
// Returns errNoNewTerminal if nothing works; caller should fall back to
// suspend-and-resume in the same terminal.
//
// Tip: running maple inside tmux or zellij gives the best experience —
// harnesses open in a new tab automatically without any configuration.
func spawnInNewTerminal(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}

	// ── multiplexers (reliable on any OS, no display needed) ─────────────────

	if os.Getenv("ZELLIJ") != "" {
		// zellij action new-tab -- <cmd> [args...]
		return exec.Command("zellij", append([]string{"action", "new-tab", "--"}, args...)...).Start()
	}
	if os.Getenv("TMUX") != "" {
		return exec.Command("tmux", append([]string{"new-window", "--"}, args...)...).Start()
	}
	if sty := os.Getenv("STY"); sty != "" {
		// screen: send a new-window command to the current session
		title := args[0]
		c := exec.Command("screen", "-S", sty, "-X", "screen", "-t", title)
		c.Args = append(c.Args, args...)
		return c.Start()
	}
	if os.Getenv("WEZTERM_PANE") != "" {
		return exec.Command("wezterm", append([]string{"cli", "spawn", "--"}, args...)...).Start()
	}
	if os.Getenv("KITTY_PID") != "" || os.Getenv("KITTY_WINDOW_ID") != "" {
		return exec.Command("kitty", append([]string{"@", "launch", "--type=tab", "--"}, args...)...).Start()
	}

	// ── GUI terminal emulators — write a launcher script ─────────────────────

	script, err := writeLaunchScript(args)
	if err != nil {
		return errNoNewTerminal
	}

	switch runtime.GOOS {
	case "darwin":
		if os.Getenv("ITERM_SESSION_ID") != "" || strings.Contains(os.Getenv("TERM_PROGRAM"), "iTerm") {
			as := fmt.Sprintf(
				`tell application "iTerm2" to tell current window to create tab with default profile command %q`,
				script)
			return exec.Command("osascript", "-e", as).Start()
		}
		// Terminal.app (default macOS)
		as := fmt.Sprintf(`tell application "Terminal" to do script %q`, script)
		return exec.Command("osascript", "-e", as).Start()

	case "linux":
		for _, cand := range [][]string{
			{"x-terminal-emulator", "-e", script},
			{"gnome-terminal", "--", script},
			{"konsole", "-e", script},
			{"xfce4-terminal", "-e", script},
			{"xterm", "-e", script},
		} {
			if _, lerr := exec.LookPath(cand[0]); lerr == nil {
				return exec.Command(cand[0], cand[1:]...).Start()
			}
		}

	case "windows":
		if _, lerr := exec.LookPath("wt"); lerr == nil {
			return exec.Command("wt", "new-tab", "--", script).Start()
		}
		// cmd start opens a new window
		return exec.Command("cmd", "/c", "start", script).Start()
	}

	return errNoNewTerminal
}

// writeLaunchScript writes a temp shell script (or .bat on Windows) that execs args.
// The caller is responsible for not cleaning it up — it will be cleaned by the OS on reboot.
func writeLaunchScript(args []string) (string, error) {
	if runtime.GOOS == "windows" {
		f, err := os.CreateTemp("", "maple-launch-*.bat")
		if err != nil {
			return "", err
		}
		defer f.Close()
		var parts []string
		for _, a := range args {
			if strings.ContainsAny(a, ` "`) {
				parts = append(parts, `"`+strings.ReplaceAll(a, `"`, `""`)+`"`)
			} else {
				parts = append(parts, a)
			}
		}
		fmt.Fprintf(f, "@echo off\r\n%s\r\npause\r\n", strings.Join(parts, " "))
		return f.Name(), nil
	}

	f, err := os.CreateTemp("", "maple-launch-*.sh")
	if err != nil {
		return "", err
	}
	defer f.Close()
	fmt.Fprintf(f, "#!/bin/sh\nexec")
	for _, a := range args {
		fmt.Fprintf(f, " %s", shQuote(a))
	}
	fmt.Fprintln(f)
	_ = os.Chmod(f.Name(), 0o755)
	return f.Name(), nil
}

func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
