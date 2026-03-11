// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

// Package frontend provides the terminal user interface for the RSVP speed reading application.
// This package handles all rendering, user input, and visual presentation.
// It communicates with the backend exclusively through the REST API.
package frontend

import "github.com/charmbracelet/lipgloss"

// Colour constants for consistent styling
const (
	ColourWord       = "15"  // White for current word
	ColourPrevious   = "8"   // Gray for previous word
	ColourNext       = "7"   // Light gray for next words
	ColourTitle      = "14"  // Cyan
	ColourProgress   = "6"   // Cyan
	ColourPaused     = "226" // Yellow
	ColourPlaying    = "46"  // Green
	ColourHelp       = "240" // Dim gray
	ColourEmptyBar   = "236" // Dark gray
	ColourFilledBar  = "39"  // Blue
	ColourSeparator  = "238" // Separator lines
	ColourKartoza    = "204" // Pink heart
	ColourLink       = "39"  // Blue links
)

// Gradient colours for WPM bar (from slow to fast)
var GradientColours = []string{
	"196", "202", "208", "214", "220", "226",
	"190", "154", "118", "82", "46", "47",
}

// Styles holds all the lipgloss styles used in the application
type Styles struct {
	// Reading screen styles
	Word     lipgloss.Style
	Previous lipgloss.Style
	Next     lipgloss.Style
	Progress lipgloss.Style
	Paused   lipgloss.Style
	Playing  lipgloss.Style
	Help     lipgloss.Style

	// Common styles
	Title     lipgloss.Style
	Separator lipgloss.Style

	// File picker styles
	FilePickerTitle    lipgloss.Style
	FilePickerSelected lipgloss.Style
	FilePickerNormal   lipgloss.Style
	FilePickerDir      lipgloss.Style

	// Resume list styles
	ResumeTitle    lipgloss.Style
	ResumeSelected lipgloss.Style
	ResumeNormal   lipgloss.Style
	ResumeProgress lipgloss.Style
}

// NewStyles creates a new Styles instance with all styles initialised
func NewStyles() Styles {
	return Styles{
		// Reading screen
		Word:     lipgloss.NewStyle().Foreground(lipgloss.Color(ColourWord)),
		Previous: lipgloss.NewStyle().Foreground(lipgloss.Color(ColourPrevious)).Italic(true),
		Next:     lipgloss.NewStyle().Foreground(lipgloss.Color(ColourNext)),
		Progress: lipgloss.NewStyle().Foreground(lipgloss.Color(ColourProgress)).Bold(true),
		Paused:   lipgloss.NewStyle().Foreground(lipgloss.Color(ColourPaused)).Bold(true),
		Playing:  lipgloss.NewStyle().Foreground(lipgloss.Color(ColourPlaying)).Bold(true),
		Help:     lipgloss.NewStyle().Foreground(lipgloss.Color(ColourHelp)),

		// Common
		Title:     lipgloss.NewStyle().Foreground(lipgloss.Color(ColourTitle)).Bold(true),
		Separator: lipgloss.NewStyle().Foreground(lipgloss.Color(ColourSeparator)),

		// File picker
		FilePickerTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourTitle)).
			Bold(true).
			MarginBottom(1),
		FilePickerSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("39")).
			Bold(true).
			Padding(0, 1),
		FilePickerNormal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		FilePickerDir: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")),

		// Resume list
		ResumeTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourTitle)).
			Bold(true).
			MarginBottom(1),
		ResumeSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("39")).
			Bold(true).
			Padding(0, 1),
		ResumeNormal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		ResumeProgress: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourProgress)),
	}
}

// GetWPMColour returns a colour for the WPM value
func GetWPMColour(wpm int) string {
	// Map WPM to gradient (0 = slow/red, 11 = fast/green)
	// Typical range: 100-600 WPM
	index := (wpm - 100) / 50
	if index < 0 {
		index = 0
	}
	if index >= len(GradientColours) {
		index = len(GradientColours) - 1
	}
	return GradientColours[index]
}

// GetProgressColour returns a colour based on reading progress
func GetProgressColour(progress float64) string {
	// Progress 0.0 to 1.0
	index := int(progress * float64(len(GradientColours)-1))
	if index < 0 {
		index = 0
	}
	if index >= len(GradientColours) {
		index = len(GradientColours) - 1
	}
	return GradientColours[index]
}
