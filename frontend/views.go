// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package frontend

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/lipgloss"
	"github.com/timlinux/cheetah/backend"
	"github.com/timlinux/cheetah/font"
)

// Renderer handles all view rendering for the application
type Renderer struct {
	styles Styles
	width  int
	height int
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
	progressBar := r.renderProgressBar(state.Progress, state.WPM)

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
	carouselElements = append(carouselElements, progressBar)
	carouselElements = append(carouselElements, "")
	carouselElements = append(carouselElements, statusLine)

	// Center the main content
	mainContent := lipgloss.JoinVertical(lipgloss.Center, carouselElements...)

	return r.renderFullScreen(mainContent, state.DocumentTitle, width, height)
}

// renderProgressBar creates a progress bar for the reading progress
func (r *Renderer) renderProgressBar(progress float64, wpm int) string {
	const barWidth = 50

	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filledWidth := int(float64(barWidth) * progress)
	emptyWidth := barWidth - filledWidth

	var bar strings.Builder

	// Build the filled portion with gradient
	for i := 0; i < filledWidth; i++ {
		colourIdx := int(float64(i) / float64(barWidth) * float64(len(GradientColours)-1))
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

	return bar.String() + percentStyle.Render(percentLabel)
}

// renderFullScreen renders content with header and footer
func (r *Renderer) renderFullScreen(content, title string, width, height int) string {
	// Header
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	headerText := "🐆 CHEETAH"
	if title != "" {
		headerText = fmt.Sprintf("🐆 CHEETAH - %s", title)
	}
	header := lipgloss.PlaceHorizontal(width, lipgloss.Center, headerStyle.Render(headerText))

	// Help text
	helpText := "SPACE pause │ r restart │ j/k speed │ h/l paragraph │ 1-9 presets │ s save │ ? docs │ ESC back │ q quit"

	// Kartoza branding
	kartozaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	heartStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	linkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	kartozaLine := kartozaStyle.Render("Made with ") +
		heartStyle.Render("♥") +
		kartozaStyle.Render(" by ") +
		linkStyle.Render("Kartoza") +
		kartozaStyle.Render(" | ") +
		linkStyle.Render("Donate!") +
		kartozaStyle.Render(" | ") +
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

// RenderFilePicker renders the file picker screen
func (r *Renderer) RenderFilePicker(fp filepicker.Model, width, height int) string {
	title := r.styles.FilePickerTitle.Render("Select a document to read")
	helpText := "Navigate with ↑/↓ │ Enter to select │ Tab for resume list │ ? help/docs │ ESC to quit"

	// Supported formats info
	formatsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	formats := formatsStyle.Render("Supported: PDF, DOCX, EPUB, ODT, TXT, MD")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		formats,
		"",
		fp.View(),
	)

	// Kartoza branding
	kartozaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	heartStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	linkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	kartozaLine := kartozaStyle.Render("Made with ") +
		heartStyle.Render("♥") +
		kartozaStyle.Render(" by ") +
		linkStyle.Render("Kartoza") +
		kartozaStyle.Render(" | ") +
		linkStyle.Render("Donate!") +
		kartozaStyle.Render(" | ") +
		linkStyle.Render("GitHub")

	// Header
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	header := lipgloss.PlaceHorizontal(width, lipgloss.Center,
		headerStyle.Render("🐆 CHEETAH - RSVP Speed Reading"))

	footer := lipgloss.JoinVertical(lipgloss.Center,
		r.styles.Help.Render(helpText),
		kartozaLine,
	)
	footer = lipgloss.PlaceHorizontal(width, lipgloss.Center, footer)

	// Center content vertically
	headerHeight := 1
	footerHeight := 2
	contentHeight := strings.Count(content, "\n") + 1
	availableHeight := height - headerHeight - footerHeight - 2

	topPadding := (availableHeight - contentHeight) / 2
	if topPadding < 1 {
		topPadding = 1
	}

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
