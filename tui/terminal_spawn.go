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

// spawnWithPane launches args and returns a paneRef that identifies where the
// agent is running so the TUI can later send "continue" keystrokes on approve.
//
// Resolution order:
//   1. Outer tmux      → tmux new-window, capture pane id
//   2. Outer zellij    → zellij new-tab --name maple-<harness>
//   3. tmux installed  → wrap in detached tmux session "maple-<harness>",
//                        then spawn a terminal that attaches to it
//   4. no multiplexer  → plain spawn, pane ref kind=""
//
// The harness label is used to name inner tmux sessions and zellij tabs so
// approvals can locate the correct target.
func spawnWithPane(harness string, args []string) (paneRef, error) {
	if len(args) == 0 {
		return paneRef{}, errors.New("empty command")
	}

	if os.Getenv("TMUX") != "" {
		c := exec.Command("tmux", append([]string{"new-window", "-PF", "#{pane_id}", "--"}, args...)...)
		out, err := c.Output()
		if err != nil {
			return paneRef{}, err
		}
		return paneRef{Kind: "tmux", Target: strings.TrimSpace(string(out))}, nil
	}

	if os.Getenv("ZELLIJ") != "" {
		tab := "maple-" + harness
		if err := exec.Command("zellij", append([]string{"action", "new-tab", "--name", tab, "--"}, args...)...).Start(); err != nil {
			return paneRef{}, err
		}
		return paneRef{Kind: "zellij", Target: tab}, nil
	}

	if _, err := exec.LookPath("tmux"); err == nil {
		session := "maple-" + harness
		_ = exec.Command("tmux", "kill-session", "-t", session).Run()
		startArgs := append([]string{"new-session", "-d", "-s", session, "--"}, args...)
		if err := exec.Command("tmux", startArgs...).Run(); err != nil {
			return paneRef{}, err
		}
		if err := spawnInNewTerminal([]string{"tmux", "attach", "-t", session}); err != nil {
			_ = exec.Command("tmux", "kill-session", "-t", session).Run()
			return paneRef{}, err
		}
		return paneRef{Kind: "tmux", Target: session}, nil
	}

	if err := spawnInNewTerminal(args); err != nil {
		return paneRef{}, err
	}
	return paneRef{}, nil
}

// spawnInNewTerminal opens args in a new terminal tab or window, keeping the
// current terminal (and maple) alive.
//
// Detection order — strictly "same terminal the user is running in":
//
//  1. Multiplexers: ZELLIJ → TMUX → STY (screen) — tabs, cross-platform
//  2. GPU terminals with IPC: WEZTERM_PANE → KITTY_PID/KITTY_WINDOW_ID
//  3. TERM_PROGRAM (canonical per-terminal env var set by most modern terminals):
//     ghostty · iTerm.app · Apple_Terminal · WarpTerminal · Hyper
//  4. Secondary per-session signals (terminals that don't set TERM_PROGRAM):
//     ITERM_SESSION_ID → iTerm2
//     TERM=alacritty / ALACRITTY_SOCKET → Alacritty
//     GNOME_TERMINAL_SCREEN / _SERVICE → GNOME Terminal (tab)
//     KONSOLE_VERSION / KONSOLE_DBUS_* → Konsole (new tab)
//     TILIX_ID → Tilix
//     TERMINATOR_UUID → Terminator
//     WT_SESSION → Windows Terminal (new tab)
//
// Returns errNoNewTerminal when the running terminal is unidentified or its
// launcher is unreachable. The caller shows a manual-launch modal.
// No generic OS fallback — opening the wrong terminal is worse than the modal.
func spawnInNewTerminal(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}

	// ── 1. multiplexers — tab support, no display needed ─────────────────────

	if os.Getenv("ZELLIJ") != "" {
		return exec.Command("zellij", append([]string{"action", "new-tab", "--"}, args...)...).Start()
	}
	if os.Getenv("TMUX") != "" {
		return exec.Command("tmux", append([]string{"new-window", "--"}, args...)...).Start()
	}
	if sty := os.Getenv("STY"); sty != "" {
		c := exec.Command("screen", "-S", sty, "-X", "screen", "-t", args[0])
		c.Args = append(c.Args, args...)
		return c.Start()
	}

	// ── 2. GPU terminals with native IPC ─────────────────────────────────────

	if os.Getenv("WEZTERM_PANE") != "" {
		return exec.Command("wezterm", append([]string{"cli", "spawn", "--"}, args...)...).Start()
	}
	if os.Getenv("KITTY_PID") != "" || os.Getenv("KITTY_WINDOW_ID") != "" {
		return exec.Command("kitty", append([]string{"@", "launch", "--type=tab", "--"}, args...)...).Start()
	}

	// Remaining mechanisms need a launcher script.
	script, err := writeLaunchScript(args)
	if err != nil {
		return errNoNewTerminal
	}

	// ── 3. TERM_PROGRAM — canonical terminal identity ─────────────────────────
	// Most modern terminals set this to a unique value. Checked first so we
	// open exactly the terminal the user is in, not one that happens to be
	// installed on the machine. If a match is found but the launcher fails,
	// return errNoNewTerminal rather than silently opening a different terminal.

	switch os.Getenv("TERM_PROGRAM") {
	case "ghostty":
		return spawnGhostty(script)
	case "iTerm.app":
		return spawnITerm2(script)
	case "Apple_Terminal":
		return spawnTerminalApp(script)
	case "WarpTerminal":
		// Warp has no stable CLI for opening new windows from an external process.
		return errNoNewTerminal
	case "Hyper":
		if p, lerr := exec.LookPath("hyper"); lerr == nil {
			return exec.Command(p, script).Start()
		}
		return errNoNewTerminal
	}

	// ── 4. Secondary per-session env vars ────────────────────────────────────
	// For terminals that don't set TERM_PROGRAM; each is unique to that app.

	if os.Getenv("ITERM_SESSION_ID") != "" {
		return spawnITerm2(script)
	}
	if os.Getenv("TERM") == "alacritty" || os.Getenv("ALACRITTY_SOCKET") != "" {
		return spawnAlacritty(script)
	}
	if os.Getenv("GNOME_TERMINAL_SCREEN") != "" || os.Getenv("GNOME_TERMINAL_SERVICE") != "" {
		if p, lerr := exec.LookPath("gnome-terminal"); lerr == nil {
			return exec.Command(p, "--tab", "--", script).Start()
		}
		return errNoNewTerminal
	}
	if os.Getenv("KONSOLE_VERSION") != "" ||
		os.Getenv("KONSOLE_DBUS_SESSION") != "" ||
		os.Getenv("KONSOLE_DBUS_WINDOW") != "" {
		if p, lerr := exec.LookPath("konsole"); lerr == nil {
			return exec.Command(p, "--new-tab", "-e", script).Start()
		}
		return errNoNewTerminal
	}
	if os.Getenv("TILIX_ID") != "" {
		if p, lerr := exec.LookPath("tilix"); lerr == nil {
			return exec.Command(p, "--action=app-new-session", "-e", script).Start()
		}
		return errNoNewTerminal
	}
	if os.Getenv("TERMINATOR_UUID") != "" {
		if p, lerr := exec.LookPath("terminator"); lerr == nil {
			return exec.Command(p, "-e", script).Start()
		}
		return errNoNewTerminal
	}
	if os.Getenv("WT_SESSION") != "" {
		if p, lerr := exec.LookPath("wt"); lerr == nil {
			return exec.Command(p, "new-tab", "--", script).Start()
		}
		return errNoNewTerminal
	}

	return errNoNewTerminal
}

func spawnGhostty(script string) error {
	if p, err := exec.LookPath("ghostty"); err == nil {
		return exec.Command(p, "-e", script).Start()
	}
	if runtime.GOOS == "darwin" {
		return exec.Command("open", "-na", "Ghostty", "--args", "-e", script).Start()
	}
	return errNoNewTerminal
}

func spawnAlacritty(script string) error {
	if p, err := exec.LookPath("alacritty"); err == nil {
		return exec.Command(p, "-e", script).Start()
	}
	if runtime.GOOS == "darwin" {
		return exec.Command("open", "-na", "Alacritty", "--args", "-e", script).Start()
	}
	return errNoNewTerminal
}

func spawnITerm2(script string) error {
	as := fmt.Sprintf(
		`tell application "iTerm2" to tell current window to create tab with default profile command %q`,
		script)
	return exec.Command("osascript", "-e", as).Start()
}

func spawnTerminalApp(script string) error {
	as := fmt.Sprintf(`tell application "Terminal" to do script %q`, script)
	return exec.Command("osascript", "-e", as).Start()
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

	// Resolve to a full path so the script works in minimal-PATH shells
	// (e.g. /bin/sh on macOS via Terminal.app doesn't have Homebrew on PATH).
	bin := args[0]
	if full, err := exec.LookPath(bin); err == nil {
		bin = full
	}

	f, err := os.CreateTemp("", "maple-launch-*.sh")
	if err != nil {
		return "", err
	}
	defer f.Close()
	fmt.Fprintf(f, "#!/bin/sh\nexec %s", shQuote(bin))
	for _, a := range args[1:] {
		fmt.Fprintf(f, " %s", shQuote(a))
	}
	fmt.Fprintln(f)
	_ = os.Chmod(f.Name(), 0o755)
	return f.Name(), nil
}

func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
