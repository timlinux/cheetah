// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package frontend

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/timlinux/blockfont"
	"github.com/timlinux/cheetah/backend"
	"github.com/timlinux/cheetah/settings"
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
// distractionFreeOpacity controls the visibility of non-essential UI elements (1.0 = fully visible, 0.0 = hidden)
func (r *Renderer) RenderReadingScreen(state *backend.ReadingState, animator *WordAnimator, width, height int, distractionFreeOpacity float64) string {
	if state == nil || !state.DocumentLoaded {
		return r.RenderLoading(width, height)
	}

	currentWord := state.CurrentWord
	if currentWord == "" {
		currentWord = "..."
	}

	// Check if all caps display is enabled
	allCapsEnabled := settings.IsAllCapsEnabled()

	// Convert to uppercase or lowercase for block font rendering based on caps setting
	var displayWord string
	if allCapsEnabled {
		displayWord = strings.ToUpper(currentWord)
	} else {
		displayWord = strings.ToLower(currentWord)
	}

	// Render current word using custom block font
	letterLines := blockfont.RenderWord(displayWord)

	// Build the block letter display
	var blockLetterLines []string
	for lineIdx := 0; lineIdx < blockfont.LetterHeight; lineIdx++ {
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

	// Progress indicator (fades with distraction-free mode)
	progressText := fmt.Sprintf("Word %d / %d", state.WordIndex+1, state.TotalWords)
	progress := applyOpacity(r.styles.Progress.Render(progressText), distractionFreeOpacity)

	// Get animation values (default to fully visible if no animator)
	prevOpacity := 0.5
	currentOffset := 0
	nextOpacity := 0.6
	if animator != nil {
		prevOpacity = animator.GetPrevOpacity()
		currentOffset = animator.GetCurrentOffset()
		nextOpacity = animator.GetNextOpacity()
	}

	// Previous word display (animated opacity + distraction-free fade)
	prevWordDisplay := ""
	if state.PreviousWord != "" {
		// Combine animation opacity with distraction-free opacity
		combinedOpacity := prevOpacity * distractionFreeOpacity
		greyLevel := 232 + int(combinedOpacity*23)
		if greyLevel > 255 {
			greyLevel = 255
		}
		prevStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(fmt.Sprintf("%d", greyLevel))).
			Italic(true)
		prevWord := state.PreviousWord
		if allCapsEnabled {
			prevWord = strings.ToUpper(prevWord)
		}
		prevWordDisplay = prevStyle.Render("· · · " + prevWord + " · · ·")
	}

	// Next words display (up to 3, with decreasing opacity + distraction-free fade)
	var nextWordsDisplay []string
	for i, word := range state.NextWords {
		// Combine animation opacity with distraction-free opacity
		wordOpacity := nextOpacity * (1.0 - float64(i)*0.2) * distractionFreeOpacity
		if wordOpacity < 0.2*distractionFreeOpacity {
			wordOpacity = 0.2 * distractionFreeOpacity
		}
		greyLevel := 232 + int(wordOpacity*23)
		if greyLevel > 255 {
			greyLevel = 255
		}
		nextStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(fmt.Sprintf("%d", greyLevel))).
			Align(lipgloss.Center)
		displayNextWord := word
		if allCapsEnabled {
			displayNextWord = strings.ToUpper(word)
		}
		if i == 0 {
			nextWordsDisplay = append(nextWordsDisplay, nextStyle.Render("▼  "+displayNextWord+"  ▼"))
		} else {
			nextWordsDisplay = append(nextWordsDisplay, nextStyle.Render(displayNextWord))
		}
	}

	// Progress bar (fades with distraction-free mode)
	progressBar := r.renderProgressBar(state.Progress, state.WPM, state.TotalWords, distractionFreeOpacity)

	// Status indicator (fades with distraction-free mode)
	var statusIndicator string
	if state.IsPaused {
		statusIndicator = r.styles.Paused.Render("⏸ PAUSED")
	} else {
		statusIndicator = r.styles.Playing.Render("▶ PLAYING")
	}
	statusLineText := fmt.Sprintf("%s │ %d WPM │ ¶ %d/%d",
		statusIndicator,
		state.WPM,
		state.ParagraphIndex+1,
		state.TotalParagraphs)
	statusLine := applyOpacity(statusLineText, distractionFreeOpacity)

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

	// Decorative separator (fades with distraction-free mode)
	// Use greyscale 232-238 for separator (darker range)
	var separatorStyle lipgloss.Style
	if distractionFreeOpacity <= 0.05 {
		separatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("232"))
	} else {
		separatorGreyLevel := 232 + int(distractionFreeOpacity*6) // 232 to 238
		separatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", separatorGreyLevel)))
	}
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
	displayCurrentWord := currentWord
	if allCapsEnabled {
		displayCurrentWord = strings.ToUpper(currentWord)
	}
	carouselElements = append(carouselElements, originalStyle.Render(displayCurrentWord))

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

	return r.renderFullScreen(mainContent, state.DocumentTitle, width, height, distractionFreeOpacity)
}

// ProgressBarWidth is the width of the progress bar in characters
const ProgressBarWidth = 50

// renderProgressBar creates a progress bar for the reading progress
// It also stores the bar position for mouse interaction
// opacity controls the visibility (1.0 = fully visible, 0.0 = hidden)
func (r *Renderer) renderProgressBar(progress float64, wpm int, totalWords int, opacity float64) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filledWidth := int(float64(ProgressBarWidth) * progress)
	emptyWidth := ProgressBarWidth - filledWidth

	var bar strings.Builder

	// Build the filled portion with gradient (faded by opacity)
	// Use greyscale 232-255 for fading (232=dark, 255=light)
	for i := 0; i < filledWidth; i++ {
		colourIdx := int(float64(i) / float64(ProgressBarWidth) * float64(len(GradientColours)-1))
		if colourIdx >= len(GradientColours) {
			colourIdx = len(GradientColours) - 1
		}
		colour := GradientColours[colourIdx]
		var style lipgloss.Style
		if opacity >= 0.9 {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color(colour))
		} else if opacity <= 0.05 {
			// Nearly invisible
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("232"))
		} else {
			// Fade to greyscale: 232 (dark) to ~245 (medium grey)
			greyLevel := 232 + int(opacity*13)
			style = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", greyLevel)))
		}
		bar.WriteString(style.Render("█"))
	}

	// Build the empty portion (faded by opacity)
	// Empty bar uses darker greyscale: 232-236
	var emptyStyle lipgloss.Style
	if opacity <= 0.05 {
		emptyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("232"))
	} else {
		emptyGreyLevel := 232 + int(opacity*4) // 232 to 236
		emptyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", emptyGreyLevel)))
	}
	for i := 0; i < emptyWidth; i++ {
		bar.WriteString(emptyStyle.Render("░"))
	}

	// Progress percentage (faded by opacity)
	percentLabel := fmt.Sprintf(" %.1f%%", progress*100)
	var percentStyle lipgloss.Style
	if opacity >= 0.9 {
		percentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColourProgress))
	} else if opacity <= 0.05 {
		percentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("232"))
	} else {
		// Greyscale: 232 (dark) to 250 (light grey)
		percentGreyLevel := 232 + int(opacity*18)
		percentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", percentGreyLevel)))
	}

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
// opacity controls the visibility of header and help text (1.0 = fully visible, 0.0 = hidden)
// Kartoza branding always remains visible (ads should stay visible per requirement)
func (r *Renderer) renderFullScreen(content, title string, width, height int, opacity float64) string {
	// Header with Kartoza orange (fades with distraction-free mode)
	// Use greyscale 232-255 for proper fading (232=near-black, 255=white)
	var header string
	headerText := "🐆 CHEETAH"
	if title != "" {
		headerText = fmt.Sprintf("🐆 CHEETAH - %s", title)
	}
	if opacity >= 0.9 {
		headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColourKartozaOrange)).Bold(true)
		header = lipgloss.PlaceHorizontal(width, lipgloss.Center, headerStyle.Render(headerText))
	} else if opacity <= 0.05 {
		// Nearly invisible - use darkest grey
		headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("232"))
		header = lipgloss.PlaceHorizontal(width, lipgloss.Center, headerStyle.Render(headerText))
	} else {
		// Faded header using greyscale: 232 (dark) to 250 (light)
		greyLevel := 232 + int(opacity*18)
		headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", greyLevel)))
		header = lipgloss.PlaceHorizontal(width, lipgloss.Center, headerStyle.Render(headerText))
	}

	// Help text (fades with distraction-free mode)
	helpText := "SPACE pause │ r restart │ j/k speed │ h/l paragraph │ 1-9 presets │ g goto │ c caps │ s save │ ESC back │ q quit"
	var renderedHelpText string
	if opacity >= 0.9 {
		renderedHelpText = r.styles.Help.Render(helpText)
	} else if opacity <= 0.05 {
		// Nearly invisible
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("232"))
		renderedHelpText = helpStyle.Render(helpText)
	} else {
		// Greyscale: 232 (dark) to 240 (medium-dark grey for help)
		greyLevel := 232 + int(opacity*8)
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", greyLevel)))
		renderedHelpText = helpStyle.Render(helpText)
	}

	// Kartoza branding with proper colors - ALWAYS VISIBLE (ads stay visible per requirement)
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
		renderedHelpText,
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

	return r.renderFullScreen(content, "", width, height, 1.0)
}

// RenderLoading renders a loading screen
func (r *Renderer) RenderLoading(width, height int) string {
	loadingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	content := loadingStyle.Render("Loading document...")
	return r.renderFullScreen(content, "", width, height, 1.0)
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
	return r.renderFullScreen(content, "", width, height, 1.0)
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
	return r.renderFullScreen(content, "", width, height, 1.0)
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

// applyOpacity applies opacity to rendered text by making it fade to black
// opacity: 1.0 = fully visible, 0.0 = hidden (rendered as nearly black)
func applyOpacity(text string, opacity float64) string {
	if opacity >= 1.0 {
		return text
	}
	if opacity <= 0.05 {
		// Nearly invisible - return empty string
		return ""
	}

	// Use 256-color greyscale range: 232 (near-black) to 255 (white)
	// opacity 0 → 232 (darkest), opacity 1 → 255 (lightest)
	greyLevel := 232 + int(opacity*23)
	if greyLevel < 232 {
		greyLevel = 232
	}
	if greyLevel > 255 {
		greyLevel = 255
	}
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", greyLevel)))
	return dimStyle.Render(stripAnsi(text))
}

// stripAnsi removes ANSI escape codes from a string
func stripAnsi(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}
