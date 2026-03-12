// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package frontend

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/timlinux/cheetah/backend"
	"github.com/timlinux/cheetah/font"
)

// ProgressBarInfo stores the position and dimensions of the progress bar
// This is used for mouse click detection on the scrubber
type ProgressBarInfo struct {
	X           int     // X position of the progress bar start (relative to left edge)
	Y           int     // Y position of the progress bar (row number from top)
	Width       int     // Width of the progress bar in characters
	Progress    float64 // Current progress value (0-1)
	TotalWords  int     // Total words in document (for calculating jump position)
	IsRendered  bool    // Whether the bar has been rendered this frame
}

// Renderer handles all view rendering for the application
type Renderer struct {
	styles         Styles
	width          int
	height         int
	progressBarInfo ProgressBarInfo
}

// NewRenderer creates a new renderer with the given dimensions
func NewRenderer(width, height int) *Renderer {
	return &Renderer{
		styles: NewStyles(),
		width:  width,
		height: height,
	}
}

// SetSize updates the renderer dimensions
func (r *Renderer) SetSize(width, height int) {
	r.width = width
	r.height = height
}

// GetProgressBarInfo returns the current progress bar position and dimensions
func (r *Renderer) GetProgressBarInfo() ProgressBarInfo {
	return r.progressBarInfo
}

// CalculateWordIndexFromClick calculates the word index from a click position on the progress bar
// Returns -1 if the click is outside the progress bar bounds
func (r *Renderer) CalculateWordIndexFromClick(clickX, clickY int) int {
	info := r.progressBarInfo
	if !info.IsRendered || info.TotalWords == 0 {
		return -1
	}

	// Check if click is on the progress bar row
	if clickY != info.Y {
		return -1
	}

	// Check if click is within the progress bar horizontal bounds
	if clickX < info.X || clickX >= info.X+info.Width {
		return -1
	}

	// Calculate the relative position within the bar (0.0 to 1.0)
	relativeX := float64(clickX-info.X) / float64(info.Width)

	// Calculate the word index
	wordIndex := int(relativeX * float64(info.TotalWords))

	// Clamp to valid range
	if wordIndex < 0 {
		wordIndex = 0
	}
	if wordIndex >= info.TotalWords {
		wordIndex = info.TotalWords - 1
	}

	return wordIndex
}

// RenderReadingScreen renders the main RSVP reading interface
func (r *Renderer) RenderReadingScreen(state *backend.ReadingState, animator *WordAnimator, width, height int) string {
	if state == nil || !state.DocumentLoaded {
		return r.RenderLoading(width, height)
	}

	currentWord := state.CurrentWord
	if currentWord == "" {
		currentWord = "..."
	}

	// Convert to lowercase for block font rendering
	displayWord := strings.ToLower(currentWord)

	// Render current word using custom block font
	letterLines := font.RenderWord(displayWord)

	// Build the block letter display
	var blockLetterLines []string
	for lineIdx := 0; lineIdx < font.LetterHeight; lineIdx++ {
		var lineBuilder strings.Builder
		for charIdx, letterLine := range letterLines[lineIdx] {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(ColourWord))
			lineBuilder.WriteString(style.Render(letterLine))
			if charIdx < len(letterLines[lineIdx])-1 {
				lineBuilder.WriteString(style.Render(" "))
			}
		}
		blockLetterLines = append(blockLetterLines, lineBuilder.String())
	}

	coloredWord := strings.Join(blockLetterLines, "\n")

	// Progress indicator
	progress := fmt.Sprintf("Word %d / %d", state.WordIndex+1, state.TotalWords)

	// Get animation values (default to fully visible if no animator)
	prevOpacity := 0.5
	currentOffset := 0
	nextOpacity := 0.6
	if animator != nil {
		prevOpacity = animator.GetPrevOpacity()
		currentOffset = animator.GetCurrentOffset()
		nextOpacity = animator.GetNextOpacity()
	}

	// Previous word display (animated opacity)
	prevWordDisplay := ""
	if state.PreviousWord != "" {
		greyLevel := 232 + int(prevOpacity*23)
		if greyLevel > 255 {
			greyLevel = 255
		}
		prevStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(fmt.Sprintf("%d", greyLevel))).
			Italic(true)
		prevWordDisplay = prevStyle.Render("· · · " + state.PreviousWord + " · · ·")
	}

	// Next words display (up to 3, with decreasing opacity)
	var nextWordsDisplay []string
	for i, word := range state.NextWords {
		wordOpacity := nextOpacity * (1.0 - float64(i)*0.2)
		if wordOpacity < 0.2 {
			wordOpacity = 0.2
		}
		greyLevel := 232 + int(wordOpacity*23)
		if greyLevel > 255 {
			greyLevel = 255
		}
		nextStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(fmt.Sprintf("%d", greyLevel))).
			Align(lipgloss.Center)
		if i == 0 {
			nextWordsDisplay = append(nextWordsDisplay, nextStyle.Render("▼  "+word+"  ▼"))
		} else {
			nextWordsDisplay = append(nextWordsDisplay, nextStyle.Render(word))
		}
	}

	// Progress bar
	progressBar := r.renderProgressBar(state.Progress, state.WPM, state.TotalWords)

	// Status indicator
	var statusIndicator string
	if state.IsPaused {
		statusIndicator = r.styles.Paused.Render("⏸ PAUSED")
	} else {
		statusIndicator = r.styles.Playing.Render("▶ PLAYING")
	}
	statusLine := fmt.Sprintf("%s │ %d WPM │ ¶ %d/%d",
		statusIndicator,
		state.WPM,
		state.ParagraphIndex+1,
		state.TotalParagraphs)

	// Build the carousel layout
	var carouselElements []string

	// Progress at top
	carouselElements = append(carouselElements, r.styles.Progress.Render(progress))
	carouselElements = append(carouselElements, "")

	// Previous word (above current, animated)
	if prevWordDisplay != "" {
		prevOffset := 0
		if animator != nil {
			prevOffset = animator.GetPrevOffset()
		}
		for i := 0; i < prevOffset; i++ {
			carouselElements = append(carouselElements, "")
		}
		carouselElements = append(carouselElements, prevWordDisplay)
		carouselElements = append(carouselElements, "")
	}

	// Decorative separator
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	carouselElements = append(carouselElements, separatorStyle.Render("─────────────────────────────────"))

	// Offset for current word animation
	for i := 0; i < currentOffset; i++ {
		carouselElements = append(carouselElements, "")
	}
	carouselElements = append(carouselElements, "")

	// Current word (large block letters) and original text below
	carouselElements = append(carouselElements, coloredWord)
	// Show original word below block letters for clarity
	originalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Align(lipgloss.Center)
	carouselElements = append(carouselElements, originalStyle.Render(currentWord))

	// Decorative separator
	carouselElements = append(carouselElements, "")
	carouselElements = append(carouselElements, separatorStyle.Render("─────────────────────────────────"))

	// Next words (below current, animated)
	if len(nextWordsDisplay) > 0 {
		nextOffset := 0
		if animator != nil {
			nextOffset = animator.GetNextOffset()
		}
		for i := 0; i < nextOffset; i++ {
			carouselElements = append(carouselElements, "")
		}
		carouselElements = append(carouselElements, "")
		for _, nextWord := range nextWordsDisplay {
			carouselElements = append(carouselElements, nextWord)
		}
	}

	carouselElements = append(carouselElements, "")
	progressBarRowIndex := len(carouselElements) // Track which row in content the progress bar is on
	carouselElements = append(carouselElements, progressBar)
	carouselElements = append(carouselElements, "")
	carouselElements = append(carouselElements, statusLine)

	// Center the main content
	mainContent := lipgloss.JoinVertical(lipgloss.Center, carouselElements...)

	// Calculate the actual Y position of the progress bar on screen
	// Content is vertically centered, so we need to calculate the top padding
	contentHeight := strings.Count(mainContent, "\n") + 1
	headerHeight := 1
	footerHeight := 2
	availableHeight := height - headerHeight - footerHeight - 2
	topPadding := (availableHeight - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	// Y position = header (1) + newline (1) + topPadding + row index within content
	r.progressBarInfo.Y = headerHeight + 1 + topPadding + progressBarRowIndex

	// X position = center of screen minus half the progress bar width
	r.progressBarInfo.X = (width - ProgressBarWidth) / 2

	return r.renderFullScreen(mainContent, state.DocumentTitle, width, height)
}

// ProgressBarWidth is the width of the progress bar in characters
const ProgressBarWidth = 50

// renderProgressBar creates a progress bar for the reading progress
// It also stores the bar position for mouse interaction
func (r *Renderer) renderProgressBar(progress float64, wpm int, totalWords int) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filledWidth := int(float64(ProgressBarWidth) * progress)
	emptyWidth := ProgressBarWidth - filledWidth

	var bar strings.Builder

	// Build the filled portion with gradient
	for i := 0; i < filledWidth; i++ {
		colourIdx := int(float64(i) / float64(ProgressBarWidth) * float64(len(GradientColours)-1))
		if colourIdx >= len(GradientColours) {
			colourIdx = len(GradientColours) - 1
		}
		colour := GradientColours[colourIdx]
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(colour))
		bar.WriteString(style.Render("█"))
	}

	// Build the empty portion
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColourEmptyBar))
	for i := 0; i < emptyWidth; i++ {
		bar.WriteString(emptyStyle.Render("░"))
	}

	// Progress percentage
	percentLabel := fmt.Sprintf(" %.1f%%", progress*100)
	percentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColourProgress))

	// Store progress bar info for mouse interaction
	// The X position will be calculated when we know where it's centered
	r.progressBarInfo = ProgressBarInfo{
		Width:      ProgressBarWidth,
		Progress:   progress,
		TotalWords: totalWords,
		IsRendered: true,
	}

	return bar.String() + percentStyle.Render(percentLabel)
}

// renderFullScreen renders content with header and footer
func (r *Renderer) renderFullScreen(content, title string, width, height int) string {
	// Header with Kartoza orange
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColourKartozaOrange)).Bold(true)
	headerText := "🐆 CHEETAH"
	if title != "" {
		headerText = fmt.Sprintf("🐆 CHEETAH - %s", title)
	}
	header := lipgloss.PlaceHorizontal(width, lipgloss.Center, headerStyle.Render(headerText))

	// Help text
	helpText := "SPACE pause │ r restart │ j/k speed │ h/l paragraph │ 1-9 presets │ g goto │ s save │ ESC back │ q quit"

	// Kartoza branding with proper colors
	kartozaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	heartStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColourKartozaOrange))
	linkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColourCheetahGold))
	kartozaLine := kartozaStyle.Render("Made with ") +
		heartStyle.Render("♥") +
		kartozaStyle.Render(" by ") +
		linkStyle.Render("Kartoza") +
		kartozaStyle.Render(" │ ") +
		linkStyle.Render("Donate!") +
		kartozaStyle.Render(" │ ") +
		linkStyle.Render("GitHub")

	footer := lipgloss.JoinVertical(lipgloss.Center,
		r.styles.Help.Render(helpText),
		kartozaLine,
	)
	footer = lipgloss.PlaceHorizontal(width, lipgloss.Center, footer)

	// Calculate heights
	headerHeight := 1
	footerHeight := 2
	contentHeight := strings.Count(content, "\n") + 1
	availableHeight := height - headerHeight - footerHeight - 2

	// Calculate top padding to center content
	topPadding := (availableHeight - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	// Build full screen
	var fullContent strings.Builder
	fullContent.WriteString(header)
	fullContent.WriteString("\n")

	for i := 0; i < topPadding; i++ {
		fullContent.WriteString("\n")
	}

	centeredContent := lipgloss.PlaceHorizontal(width, lipgloss.Center, content)
	fullContent.WriteString(centeredContent)

	currentHeight := headerHeight + 1 + topPadding + contentHeight
	for i := currentHeight; i < height-footerHeight; i++ {
		fullContent.WriteString("\n")
	}

	fullContent.WriteString(footer)

	return fullContent.String()
}

// RenderResumeList renders the saved sessions resume list
func (r *Renderer) RenderResumeList(sessions []backend.SavedSession, cursor int, width, height int) string {
	title := r.styles.ResumeTitle.Render("Resume Reading")

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	if len(sessions) == 0 {
		noSessionsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
		lines = append(lines, noSessionsStyle.Render("No saved reading sessions"))
	} else {
		for i, session := range sessions {
			progress := float64(session.LastPosition) / float64(session.TotalWords) * 100

			var line string
			if i == cursor {
				line = r.styles.ResumeSelected.Render(session.DocumentTitle)
			} else {
				line = r.styles.ResumeNormal.Render(session.DocumentTitle)
			}

			progressText := fmt.Sprintf(" %.0f%% @ %d WPM", progress, session.LastWPM)
			line += r.styles.ResumeProgress.Render(progressText)

			lines = append(lines, line)
		}
	}

	content := strings.Join(lines, "\n")

	return r.renderFullScreen(content, "", width, height)
}

// RenderLoading renders a loading screen
func (r *Renderer) RenderLoading(width, height int) string {
	loadingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	content := loadingStyle.Render("Loading document...")
	return r.renderFullScreen(content, "", width, height)
}

// RenderError renders an error screen
func (r *Renderer) RenderError(errMsg string, width, height int) string {
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	content := lipgloss.JoinVertical(lipgloss.Center,
		errorStyle.Render("Error loading document:"),
		"",
		errMsg,
		"",
		r.styles.Help.Render("Press ESC to go back"),
	)
	return r.renderFullScreen(content, "", width, height)
}

// RenderSettings renders the settings screen
func (r *Renderer) RenderSettings(cursor int, width, height int) string {
	title := r.styles.Title.Render("Settings")
	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		"Settings coming soon...",
		"",
		r.styles.Help.Render("Press ESC to go back"),
	)
	return r.renderFullScreen(content, "", width, height)
}

// RenderGotoOverlay renders the go-to percentage input overlay
func (r *Renderer) RenderGotoOverlay(gotoInput string, totalWords int) string {
	// Create a styled box for the input
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColourCheetahGold)).
		Padding(1, 2).
		Align(lipgloss.Center)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColourCheetahGold)).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	// Show current input or placeholder
	displayInput := gotoInput
	if displayInput == "" {
		displayInput = "0"
	}
	displayInput += "%"

	// Calculate what word index this would be
	wordHint := ""
	if totalWords > 0 && gotoInput != "" {
		if percentage, err := parseFloat(gotoInput); err == nil {
			wordIndex := int(percentage / 100.0 * float64(totalWords))
			if wordIndex >= totalWords {
				wordIndex = totalWords - 1
			}
			if wordIndex < 0 {
				wordIndex = 0
			}
			wordHint = fmt.Sprintf("→ Word %d / %d", wordIndex+1, totalWords)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("Go to Position"),
		"",
		inputStyle.Render(displayInput),
		"",
		hintStyle.Render(wordHint),
		"",
		hintStyle.Render("Enter percentage (0-100)"),
		hintStyle.Render("ENTER to jump │ ESC to cancel"),
	)

	return boxStyle.Render(content)
}

// parseFloat is a helper that wraps strconv.ParseFloat
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
