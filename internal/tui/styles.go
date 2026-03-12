// Package tui implements the Bubble Tea terminal interface.
package tui

import "github.com/charmbracelet/lipgloss"

// Color palette — matches fleetctl's Charmbracelet styling.
var (
	ColorInfo     = lipgloss.Color("117") // sky blue
	ColorSuccess  = lipgloss.Color("120") // mint green
	ColorWarning  = lipgloss.Color("222") // peach
	ColorDanger   = lipgloss.Color("210") // coral
	ColorAccent   = lipgloss.Color("183") // lavender
	ColorMuted    = lipgloss.Color("246") // grey
	ColorSelected = lipgloss.Color("115") // teal
	ColorHeading  = lipgloss.Color("147") // periwinkle

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Padding(0, 2)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	StyleMuted = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	StyleDanger = lipgloss.NewStyle().
			Foreground(ColorDanger)

	StyleWarning = lipgloss.NewStyle().
			Foreground(ColorWarning)

	StyleInfo = lipgloss.NewStyle().
			Foreground(ColorInfo)

	StyleHighConfidence = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	StyleMedConfidence = lipgloss.NewStyle().
				Foreground(ColorWarning)

	StyleLowConfidence = lipgloss.NewStyle().
				Foreground(ColorDanger)

	StyleStatusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	StyleHelpKey = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	StyleHelpDesc = lipgloss.NewStyle().
			Foreground(ColorMuted)
)
