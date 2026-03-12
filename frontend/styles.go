// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

// Package frontend provides the terminal user interface for the RSVP speed reading application.
// This package handles all rendering, user input, and visual presentation.
// It communicates with the backend exclusively through the REST API.
package frontend

import "github.com/charmbracelet/lipgloss"

// Kartoza branding colours
const (
	// Primary Kartoza orange
	ColourKartozaOrange = "#FF6B35"
	ColourKartozaDark   = "#1a1a2e"
	ColourKartozaAccent = "#16213e"

	// Cheetah theme (derived from Kartoza + cheetah spots)
	ColourCheetahGold   = "#FFB347" // Cheetah fur
	ColourCheetahSpots  = "#8B4513" // Dark brown spots
	ColourCheetahCream  = "#FFFDD0" // Light cream
	ColourCheetahAmber  = "#FFBF00" // Amber eyes
)

// Colour constants for consistent styling
const (
	ColourWord       = "#FFFFFF" // White for current word
	ColourPrevious   = "#888888" // Gray for previous word
	ColourNext       = "#AAAAAA" // Light gray for next words
	ColourTitle      = "#FF6B35" // Kartoza orange
	ColourProgress   = "#FFB347" // Cheetah gold
	ColourPaused     = "#FFBF00" // Amber (warning)
	ColourPlaying    = "#32CD32" // Lime green
	ColourHelp       = "#666666" // Dim gray
	ColourEmptyBar   = "#333333" // Dark gray
	ColourFilledBar  = "#FF6B35" // Kartoza orange
	ColourSeparator  = "#444444" // Separator lines
	ColourKartoza    = "#FF6B35" // Orange heart
	ColourLink       = "#FFB347" // Gold links
	ColourBorder     = "#FF6B35" // Kartoza orange borders
	ColourDirIcon    = "#FFB347" // Gold for directories
	ColourFileIcon   = "#FFFFFF" // White for files
	ColourSelected   = "#FF6B35" // Kartoza orange selection
	ColourSelectedFg = "#000000" // Black text on selection
	ColourHeader     = "#FF6B35" // Header text
	ColourPath       = "#FFB347" // Current path display
)

// Gradient colours for WPM bar (Kartoza-inspired warm tones)
var GradientColours = []string{
	"#8B0000", "#B22222", "#CD5C5C", "#F08080", // Reds (slow)
	"#FF6B35", "#FF8C00", "#FFA500", "#FFB347", // Oranges (medium)
	"#FFD700", "#ADFF2F", "#32CD32", "#00FF00", // Yellows to greens (fast)
}

// Box drawing characters for beautiful borders
const (
	BoxTopLeft     = "╭"
	BoxTopRight    = "╮"
	BoxBottomLeft  = "╰"
	BoxBottomRight = "╯"
	BoxHorizontal  = "─"
	BoxVertical    = "│"
	BoxTLeft       = "├"
	BoxTRight      = "┤"
	BoxTTop        = "┬"
	BoxTBottom     = "┴"
	BoxCross       = "┼"
)

// File type icons
const (
	IconFolder    = "📁"
	IconFolderUp  = "📂"
	IconPDF       = "📕"
	IconWord      = "📘"
	IconEpub      = "📗"
	IconText      = "📄"
	IconMarkdown  = "📝"
	IconODT       = "📙"
	IconFile      = "📄"
	IconCheetah   = "🐆"
	IconHeart     = "❤️"
)

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

	// File browser styles
	FileBrowserBorder       lipgloss.Style
	FileBrowserHeader       lipgloss.Style
	FileBrowserPath         lipgloss.Style
	FileBrowserSelected     lipgloss.Style
	FileBrowserNormal       lipgloss.Style
	FileBrowserDirectory    lipgloss.Style
	FileBrowserFile         lipgloss.Style
	FileBrowserStatusBar    lipgloss.Style
	FileBrowserHelp         lipgloss.Style
	FileBrowserColumnHeader lipgloss.Style

	// Legacy file picker styles (for compatibility)
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

		// File browser - beautiful bordered panel
		FileBrowserBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColourBorder)).
			Padding(0, 1),
		FileBrowserHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourHeader)).
			Bold(true).
			Align(lipgloss.Center),
		FileBrowserPath: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourPath)).
			Bold(true),
		FileBrowserSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourSelectedFg)).
			Background(lipgloss.Color(ColourSelected)).
			Bold(true).
			Padding(0, 1),
		FileBrowserNormal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")).
			Padding(0, 1),
		FileBrowserDirectory: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourDirIcon)).
			Bold(true),
		FileBrowserFile: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourFileIcon)),
		FileBrowserStatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourHelp)).
			Align(lipgloss.Center),
		FileBrowserHelp: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),
		FileBrowserColumnHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourHeader)).
			Bold(true).
			Underline(true),

		// Legacy file picker (for compatibility)
		FilePickerTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourTitle)).
			Bold(true).
			MarginBottom(1),
		FilePickerSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourSelectedFg)).
			Background(lipgloss.Color(ColourSelected)).
			Bold(true).
			Padding(0, 1),
		FilePickerNormal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")),
		FilePickerDir: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourDirIcon)),

		// Resume list
		ResumeTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourTitle)).
			Bold(true).
			MarginBottom(1),
		ResumeSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColourSelectedFg)).
			Background(lipgloss.Color(ColourSelected)).
			Bold(true).
			Padding(0, 1),
		ResumeNormal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")),
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
