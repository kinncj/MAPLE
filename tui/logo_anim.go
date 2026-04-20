package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// logoRows is the canonical glyph. DO NOT re-kern or redraw.
var logoRows = [8]string{
	"                                                                    ",
	"   ▄▄▄▄   ▄▄▄▄▄    ▄▄▄▄▄▄▄   ▄▄▄▄▄   ▄▄▄  ▄▄▄   ▄▄▄▄   ▄▄▄▄▄▄   ",
	"  ▄██▀▀██▄  ███    █████▀▀▀ ▄███████▄ ███  ███ ▄██▀▀██▄ ███▀▀██▄  ",
	"  ███  ███  ███     ▀████▄  ███   ███ ███  ███ ███  ███ ███  ███   ",
	"  ███▀▀███  ███       ▀████ ███▄█▄███ ███▄▄███ ███▀▀███ ███  ███   ",
	"  ███  ███ ▄███▄   ███████▀  ▀█████▀  ▀██████▀ ███  ███ ██████▀   ",
	"                          ▀▀                                        ",
	"                                                                    ",
}

// logoColor is the hacker-green used for the logo across all contexts.
const logoColor = lipgloss.Color("#73daca")

// logo returns the full static colored logo string (for Bubble Tea views).
func logo() string {
	style := lipgloss.NewStyle().Foreground(logoColor)
	var sb strings.Builder
	for _, row := range logoRows {
		sb.WriteString(style.Render(row))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// ─── Bubble Tea animation ─────────────────────────────────────────────────────

// logoTickMsg advances the logo animation by one frame.
type logoTickMsg struct{}

const (
	logoFrameCount = 5 // frames 0–4, then done
	logoFrameDelay = 70 * time.Millisecond
	logoFrame0Wait = 60 * time.Millisecond
)

// logoTick returns a Cmd that fires logoTickMsg after the right delay for frame n.
func logoTick(frame int) tea.Cmd {
	delay := logoFrameDelay
	if frame == 0 {
		delay = logoFrame0Wait
	}
	return tea.Tick(delay, func(time.Time) tea.Msg { return logoTickMsg{} })
}

// logoAnimFrame renders the logo at the given animation frame (0..5+).
//
//	frame 0: edge-on — single dim rule at row 3
//	frame 1: center 2 rows (3-4) revealed in green
//	frame 2: center 4 rows (2-5) revealed
//	frame 3: center 6 rows (1-6) revealed
//	frame 4+: full logo
func logoAnimFrame(frame int) string {
	N := len(logoRows)
	green := lipgloss.NewStyle().Foreground(logoColor)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#414868"))

	var sb strings.Builder

	if frame == 0 {
		// Edge-on: dim rule at row 3, blank elsewhere
		for i := 0; i < N; i++ {
			if i == 3 {
				sb.WriteString(dim.Render("  " + strings.Repeat("─", 66)))
			}
			sb.WriteByte('\n')
		}
		return sb.String()
	}

	// Frames 1-3: expand outward from center pair (rows 3-4)
	// f=2 → top=3, rows=2; f=1 → top=2, rows=4; f=0 → top=1, rows=6
	f := 3 - frame // frame 1→f=2, frame 2→f=1, frame 3→f=0
	if f < 0 {
		f = 0
	}
	top := f + 1
	rowsShown := 6 - 2*f

	for i := 0; i < N; i++ {
		if i >= top && i < top+rowsShown {
			sb.WriteString(green.Render(logoRows[i]))
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// ─── Sweep animation (PRD §5.10: left-to-right reveal + accent pulse) ────────

const (
	logoSweepFrameCount = 14
	logoSweepDelay      = 55 * time.Millisecond // 14 × 55ms ≈ 770ms
	logoPulseFrameCount = 4
	logoPulseDelay      = 80 * time.Millisecond
	logoShimmerWidth    = 68 // canonical logo width in runes
)

// logoSweepFrame reveals the logo left-to-right. frame ∈ [0, logoSweepFrameCount).
func logoSweepFrame(frame int, col lipgloss.Color) string {
	charsToShow := (frame + 1) * logoShimmerWidth / logoSweepFrameCount
	style := lipgloss.NewStyle().Foreground(col)
	var sb strings.Builder
	for _, row := range logoRows {
		runes := []rune(row)
		visible := make([]rune, len(runes))
		for i, r := range runes {
			if i < charsToShow {
				visible[i] = r
			} else {
				visible[i] = ' '
			}
		}
		sb.WriteString(style.Render(string(visible)))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// logoPulseFrame renders the logo in a single color for the post-sweep pulse.
// phase 0,2 = primary; phase 1 = accent (the "pulse"); phase 3+ = primary steady.
func logoPulseFrame(phase int, primary, accent lipgloss.Color) string {
	col := primary
	if phase == 1 {
		col = accent
	}
	style := lipgloss.NewStyle().Foreground(col)
	var sb strings.Builder
	for _, row := range logoRows {
		sb.WriteString(style.Render(row))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// logoShimmer renders the static logo with one bright cell traveling across it.
// shimmerPos ∈ [0, logoShimmerWidth). Pass -1 to render without shimmer.
func logoShimmer(shimmerPos int, primary, accent lipgloss.Color) string {
	normal := lipgloss.NewStyle().Foreground(primary)
	bright := lipgloss.NewStyle().Foreground(accent).Bold(true)
	var sb strings.Builder
	for _, row := range logoRows {
		runes := []rune(row)
		if shimmerPos < 0 || shimmerPos >= len(runes) || runes[shimmerPos] == ' ' {
			sb.WriteString(normal.Render(row))
		} else {
			before := string(runes[:shimmerPos])
			at := string(runes[shimmerPos : shimmerPos+1])
			after := string(runes[shimmerPos+1:])
			sb.WriteString(normal.Render(before))
			sb.WriteString(bright.Render(at))
			sb.WriteString(normal.Render(after))
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// ─── Raw stdout animation (for --help / non-TUI paths) ───────────────────────

// printLogoAnimated writes the animated logo directly to stdout.
// Falls back to a plain static render when stdout is not a TTY.
func printLogoAnimated() {
	if !isStdoutTTY() {
		printLogoStatic()
		return
	}

	N := len(logoRows)
	green := "\033[38;2;115;218;202m" // #73daca in 24-bit
	dimClr := "\033[2m"
	reset := "\033[0m"

	// Reserve N lines
	for i := 0; i < N; i++ {
		fmt.Println()
	}

	// Frame 0: edge-on
	fmt.Printf("\033[%dA", N)
	for i := 0; i < N; i++ {
		if i == 3 {
			fmt.Printf("\033[2K%s  %s%s\n", dimClr, strings.Repeat("─", 66), reset)
		} else {
			fmt.Print("\033[2K\n")
		}
	}
	time.Sleep(logoFrame0Wait)

	// Frames 1-3: expand from center outward
	for f := 2; f >= 0; f-- {
		top := f + 1
		rowsShown := 6 - 2*f
		fmt.Printf("\033[%dA", N)
		fmt.Print(green)
		for i := 0; i < N; i++ {
			if i >= top && i < top+rowsShown {
				fmt.Printf("\033[2K%s\n", logoRows[i])
			} else {
				fmt.Print("\033[2K\n")
			}
		}
		time.Sleep(logoFrameDelay)
	}
	fmt.Print(reset)
}

func printLogoStatic() {
	for _, row := range logoRows {
		fmt.Println(row)
	}
}

func isStdoutTTY() bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func isStdinTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// hasTTY reports whether a controlling terminal is available.
// It tries to open /dev/tty (POSIX) and returns false on Windows or in CI
// environments where no TTY exists.
func hasTTY() bool {
	f, err := os.Open("/dev/tty")
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}
