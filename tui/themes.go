package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme holds all styled renderers for the TUI.
type Theme struct {
	Name       string
	Primary    lipgloss.Color
	Accent     lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
	Muted      lipgloss.Color
	Background lipgloss.Color
	Foreground lipgloss.Color

	Border     lipgloss.Style
	Pane       lipgloss.Style
	ActivePane lipgloss.Style
	PaneTitle  lipgloss.Style
	StatusBar  lipgloss.Style
	Footer     lipgloss.Style

	// Text helpers
	PrimaryText lipgloss.Style
	MutedText   lipgloss.Style
	SuccessText lipgloss.Style
	ErrorText   lipgloss.Style
	WarningText lipgloss.Style
}

// loadTheme returns the default theme (tokyo-night).
// Use themeByName to switch at runtime via :theme <name>.
func loadTheme() Theme {
	return tokyoNight()
}

func buildTheme(name string, primary, accent, success, warning, errColor, muted, bg, fg string) Theme {
	p := lipgloss.Color(primary)
	a := lipgloss.Color(accent)
	s := lipgloss.Color(success)
	w := lipgloss.Color(warning)
	e := lipgloss.Color(errColor)
	m := lipgloss.Color(muted)
	b := lipgloss.Color(bg)
	f := lipgloss.Color(fg)

	pane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m).
		Background(b)

	return Theme{
		Name:       name,
		Primary:    p,
		Accent:     a,
		Success:    s,
		Warning:    w,
		Error:      e,
		Muted:      m,
		Background: b,
		Foreground: f,

		Border: lipgloss.NewStyle().
			Foreground(p).Bold(true),

		Pane:       pane,
		ActivePane: pane.Copy().BorderForeground(p),

		PaneTitle: lipgloss.NewStyle().
			Foreground(p).Bold(true).
			PaddingLeft(1),

		StatusBar: lipgloss.NewStyle().
			Foreground(f).Background(m).
			PaddingLeft(1),

		Footer: lipgloss.NewStyle().
			Foreground(m).
			PaddingLeft(1),

		PrimaryText: lipgloss.NewStyle().Foreground(p),
		MutedText:   lipgloss.NewStyle().Foreground(m),
		SuccessText: lipgloss.NewStyle().Foreground(s),
		ErrorText:   lipgloss.NewStyle().Foreground(e),
		WarningText: lipgloss.NewStyle().Foreground(w),
	}
}

// ─── Built-in themes ──────────────────────────────────────────────────────────

// logo returns the canonical AI-Squad ASCII block art header.
// DO NOT re-kern or redraw this glyph — only rendering (colors, animation) may change.
func logo() string {
	return "" +
		"                                                                    \n" +
		"   ▄▄▄▄   ▄▄▄▄▄    ▄▄▄▄▄▄▄   ▄▄▄▄▄   ▄▄▄  ▄▄▄   ▄▄▄▄   ▄▄▄▄▄▄   \n" +
		"  ▄██▀▀██▄  ███    █████▀▀▀ ▄███████▄ ███  ███ ▄██▀▀██▄ ███▀▀██▄  \n" +
		"  ███  ███  ███     ▀████▄  ███   ███ ███  ███ ███  ███ ███  ███   \n" +
		"  ███▀▀███  ███       ▀████ ███▄█▄███ ███▄▄███ ███▀▀███ ███  ███   \n" +
		"  ███  ███ ▄███▄   ███████▀  ▀█████▀  ▀██████▀ ███  ███ ██████▀   \n" +
		"                          ▀▀                                        \n"
}

func tokyoNight() Theme {
	return buildTheme("tokyo-night",
		"#7aa2f7", // primary (blue)
		"#bb9af7", // accent (purple)
		"#9ece6a", // success (green)
		"#e0af68", // warning (orange)
		"#f7768e", // error (red)
		"#565f89", // muted
		"#1a1b26", // background
		"#c0caf5", // foreground
	)
}

func catppuccinMocha() Theme {
	return buildTheme("catppuccin-mocha",
		"#89b4fa", "#cba6f7", "#a6e3a1", "#fab387", "#f38ba8",
		"#585b70", "#1e1e2e", "#cdd6f4",
	)
}

func gruvbox() Theme {
	return buildTheme("gruvbox",
		"#83a598", "#d3869b", "#b8bb26", "#fabd2f", "#fb4934",
		"#928374", "#282828", "#ebdbb2",
	)
}

func nord() Theme {
	return buildTheme("nord",
		"#88c0d0", "#b48ead", "#a3be8c", "#ebcb8b", "#bf616a",
		"#4c566a", "#2e3440", "#eceff4",
	)
}

func everforest() Theme {
	return buildTheme("everforest",
		"#7fbbb3", "#d699b6", "#a7c080", "#dbbc7f", "#e67e80",
		"#859289", "#2d353b", "#d3c6aa",
	)
}

// themeByName returns a named theme or falls back to tokyo-night.
func themeByName(name string) Theme {
	switch name {
	case "catppuccin-mocha":
		return catppuccinMocha()
	case "gruvbox":
		return gruvbox()
	case "nord":
		return nord()
	case "everforest":
		return everforest()
	default:
		return tokyoNight()
	}
}
