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
// current terminal (and maple) alive. Detection order:
//
//  1. zellij     — action new-tab (ZELLIJ env var)
//  2. tmux       — new-window (TMUX env var)
//  3. GNU screen — new window via STY
//  4. WezTerm    — cli spawn (WEZTERM_PANE env var)
//  5. Kitty      — @launch --type=tab (KITTY_PID / KITTY_WINDOW_ID env var)
//  6. Ghostty    — new window via -- (TERM_PROGRAM=ghostty / GHOSTTY_RESOURCES_DIR env var)
//  7. Alacritty  — new window via -e (TERM=alacritty / ALACRITTY_SOCKET env var)
//  8. macOS      — iTerm2 new tab → Terminal.app via osascript
//     Linux      — GNOME Terminal (--tab) → Konsole (--new-tab) → generic list
//     Windows    — Windows Terminal (wt new-tab) → cmd start
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
		return exec.Command("zellij", append([]string{"action", "new-tab", "--"}, args...)...).Start()
	}
	if os.Getenv("TMUX") != "" {
		return exec.Command("tmux", append([]string{"new-window", "--"}, args...)...).Start()
	}
	if sty := os.Getenv("STY"); sty != "" {
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

	// ── cross-platform GUI terminals — detected via env vars set by the terminal itself ──
	// Checking env vars first ensures we open the same terminal the user is running in,
	// regardless of what other terminals are installed on the system.

	// Ghostty sets TERM_PROGRAM=ghostty and GHOSTTY_RESOURCES_DIR in every session.
	if os.Getenv("TERM_PROGRAM") == "ghostty" || os.Getenv("GHOSTTY_RESOURCES_DIR") != "" {
		if p, lerr := exec.LookPath("ghostty"); lerr == nil {
			return exec.Command(p, "--", script).Start()
		}
		// App bundle installed but not linked onto PATH (common on macOS).
		if runtime.GOOS == "darwin" {
			if err := exec.Command("open", "-na", "Ghostty", "--args", "--", script).Start(); err == nil {
				return nil
			}
		}
	}

	// Alacritty sets TERM=alacritty; ALACRITTY_SOCKET is an additional signal.
	if os.Getenv("TERM") == "alacritty" || os.Getenv("ALACRITTY_SOCKET") != "" {
		if p, lerr := exec.LookPath("alacritty"); lerr == nil {
			return exec.Command(p, "-e", script).Start()
		}
		if runtime.GOOS == "darwin" {
			if err := exec.Command("open", "-na", "Alacritty", "--args", "-e", script).Start(); err == nil {
				return nil
			}
		}
	}

	switch runtime.GOOS {
	case "darwin":
		// iTerm2 — new tab in the current window
		if os.Getenv("ITERM_SESSION_ID") != "" || strings.Contains(os.Getenv("TERM_PROGRAM"), "iTerm") {
			as := fmt.Sprintf(
				`tell application "iTerm2" to tell current window to create tab with default profile command %q`,
				script)
			return exec.Command("osascript", "-e", as).Start()
		}
		// Terminal.app fallback
		as := fmt.Sprintf(`tell application "Terminal" to do script %q`, script)
		return exec.Command("osascript", "-e", as).Start()

	case "linux":
		// GNOME Terminal — new tab in the running instance
		if os.Getenv("GNOME_TERMINAL_SCREEN") != "" || os.Getenv("GNOME_TERMINAL_SERVICE") != "" {
			if p, lerr := exec.LookPath("gnome-terminal"); lerr == nil {
				return exec.Command(p, "--tab", "--", script).Start()
			}
		}
		// Konsole — new tab in the running instance
		if os.Getenv("KONSOLE_VERSION") != "" {
			if p, lerr := exec.LookPath("konsole"); lerr == nil {
				return exec.Command(p, "--new-tab", "-e", script).Start()
			}
		}
		// Generic fallback — first terminal found on PATH wins
		for _, cand := range [][]string{
			{"x-terminal-emulator", "-e", script},
			{"gnome-terminal", "--", script},
			{"konsole", "-e", script},
			{"xfce4-terminal", "-e", script},
			{"alacritty", "-e", script},
			{"ghostty", "--", script},
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
