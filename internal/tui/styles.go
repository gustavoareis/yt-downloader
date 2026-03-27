// Package tui contains all terminal user-interface code.
package tui

import "github.com/charmbracelet/lipgloss"

// ─── Color tokens ────────────────────────────────────────────────────────────

var (
	colorPurple = lipgloss.Color("#A855F7")
	colorPink   = lipgloss.Color("#EC4899")
	colorCyan   = lipgloss.Color("#06B6D4")
	colorGreen  = lipgloss.Color("#22C55E")
	colorRed    = lipgloss.Color("#EF4444")
	colorYellow = lipgloss.Color("#F59E0B")
	colorBgDark = lipgloss.Color("#0F0F1A")
	colorHover  = lipgloss.Color("#252545")
	colorMuted  = lipgloss.Color("#6B7280")
)

// ─── Component styles ────────────────────────────────────────────────────────

var (
	styleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(1, 3)

	styleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	styleSubtitle = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleMuted = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleSuccess = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	styleError = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	styleWarning = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true)

	styleItemSelected = lipgloss.NewStyle().
				Background(colorHover).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPurple).
				Padding(0, 2)

	styleItemNormal = lipgloss.NewStyle().
			Foreground(colorMuted).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#2D2D4E")).
			Padding(0, 2)

	styleKeyHint = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	styleKeyLabel = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F0F1A")).
			Background(colorPurple).
			Bold(true).
			Padding(0, 1)

	styleLogBox = lipgloss.NewStyle().
			Background(lipgloss.Color("#0A0A14")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#2D2D4E")).
			Padding(0, 1)

	styleSpinner = lipgloss.NewStyle().
			Foreground(colorPurple)
)
