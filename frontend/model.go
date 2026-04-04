// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package frontend

import (
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/timlinux/cheetah/backend"
	"github.com/timlinux/cheetah/sessions"
	"github.com/timlinux/cheetah/settings"
)

// AppState represents the current state of the application UI
type AppState int

const (
	StateFilePicker AppState = iota
	StateResumeList
	StateReading
	StateSettings
)

// tickMsg is sent periodically to update the display
type tickMsg time.Time

// animTickMsg is sent to update animations
type animTickMsg time.Time

// distractionFadeTickMsg is sent to update distraction-free fade animation
type distractionFadeTickMsg time.Time

// stateUpdateMsg is sent when the backend state changes
type stateUpdateMsg backend.ReadingState

// documentLoadedMsg is sent when a document finishes loading
type documentLoadedMsg struct {
	err error
}

// Model is the Bubble Tea model for the reading application
type Model struct {
	// Engine adapter (embedded or HTTP client)
	engine EngineAdapter

	// UI state
	state    AppState
	width    int
	height   int
	renderer *Renderer
	animator *WordAnimator

	// File browser (custom beautiful file picker)
	fileBrowser  *FileBrowser
	selectedFile string
	initialPath  string

	// Resume list
	savedSessions []backend.SavedSession
	resumeCursor  int

	// Reading state
	readingState    *backend.ReadingState
	lastWordIndex   int
	initialWPM      int
	documentLoading bool
	loadError       string

	// Scrubber state for mouse interaction
	isDragging     bool // Whether user is dragging the scrubber
	wasPausedBeforeDrag bool // Track pause state before drag started

	// Go-to mode state
	gotoMode   bool   // Whether user is in go-to mode (typing percentage)
	gotoInput  string // Current input for go-to mode

	// Distraction-free mode state
	lastKeyboardActivity    time.Time // When the last keyboard activity occurred
	distractionFreeOpacity  float64   // Opacity of non-essential UI elements (1.0 = visible, 0.0 = hidden)
	distractionFadeActive   bool      // Whether the fade animation is currently running

	// Settings
	settingsCursor int
}

// NewModel creates a new Model with the embedded engine (standalone mode)
func NewModel(documentPath string, initialWPM int) Model {
	// Create embedded engine (no HTTP needed)
	engine := NewEmbeddedEngine()

	// Initialize custom file browser
	fb := NewFileBrowser()

	// Determine initial state
	initialState := StateFilePicker
	if documentPath != "" {
		initialState = StateReading
	}

	return Model{
		engine:                 engine,
		state:                  initialState,
		renderer:               NewRenderer(80, 24),
		animator:               NewWordAnimator(),
		fileBrowser:            fb,
		initialPath:            documentPath,
		initialWPM:             initialWPM,
		readingState:           &backend.ReadingState{IsPaused: true, WPM: initialWPM},
		lastKeyboardActivity:   time.Now(),
		distractionFreeOpacity: 1.0,
	}
}

// NewModelWithEngine creates a model with a custom engine adapter
func NewModelWithEngine(engine EngineAdapter, documentPath string, initialWPM int) Model {
	fb := NewFileBrowser()

	initialState := StateFilePicker
	if documentPath != "" {
		initialState = StateReading
	}

	return Model{
		engine:                 engine,
		state:                  initialState,
		renderer:               NewRenderer(80, 24),
		animator:               NewWordAnimator(),
		fileBrowser:            fb,
		initialPath:            documentPath,
		initialWPM:             initialWPM,
		readingState:           &backend.ReadingState{IsPaused: true, WPM: initialWPM},
		lastKeyboardActivity:   time.Now(),
		distractionFreeOpacity: 1.0,
	}
}

// Init initializes the model and returns the initial command
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	// If we have a document path, load it
	if m.initialPath != "" {
		cmds = append(cmds, m.loadDocumentCmd(m.initialPath))
	}

	// Start tick for state polling
	cmds = append(cmds, tickCmd())

	return tea.Batch(cmds...)
}

// Update handles messages and updates the model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case documentLoadedMsg:
		m.documentLoading = false
		if msg.err != nil {
			m.loadError = msg.err.Error()
		} else {
			m.loadError = ""
			m.state = StateReading

			// Check if there's a saved session for this document
			docInfo, err := m.engine.GetDocumentInfo()
			if err == nil && docInfo != nil && docInfo.Hash != "" {
				session, err := sessions.Load(docInfo.Hash)
				if err == nil && session != nil {
					// Restore saved position and WPM
					m.engine.JumpToWord(session.LastPosition)
					if session.LastWPM > 0 {
						m.engine.SetWPM(session.LastWPM)
					}
				} else if m.initialWPM > 0 {
					// No saved session, use initial WPM if provided
					m.engine.SetWPM(m.initialWPM)
				}
			} else if m.initialWPM > 0 {
				// Couldn't get document info, use initial WPM if provided
				m.engine.SetWPM(m.initialWPM)
			}
		}
		return m, nil

	case tickMsg:
		// Poll engine state directly
		if m.state == StateReading && m.engine != nil {
			state, err := m.engine.GetState()
			if err == nil && state != nil {
				// Check if word changed for animation
				if state.WordIndex != m.lastWordIndex && !state.IsPaused {
					m.animator.TriggerTransition()
					m.lastWordIndex = state.WordIndex
				}
				m.readingState = state
			}
		}
		return m, tickCmd()

	case animTickMsg:
		// Handle animations
		if m.animator != nil && m.animator.IsAnimating {
			m.animator.Update()
			if m.animator.IsAnimating {
				return m, animTickCmd()
			}
		}
		return m, nil

	case distractionFadeTickMsg:
		// Handle distraction-free fade animation during play mode
		if m.state == StateReading && m.readingState != nil && !m.readingState.IsPaused {
			elapsed := time.Since(m.lastKeyboardActivity)

			// Distraction-free timing constants
			const fadeStartDelay = 2 * time.Second   // Start fading after 2 seconds of inactivity
			const fadeDuration = 3 * time.Second     // Complete fade over 3 seconds

			if elapsed < fadeStartDelay {
				// Before fade starts, keep fully visible
				m.distractionFreeOpacity = 1.0
			} else {
				// Calculate fade progress (0.0 to 1.0)
				fadeElapsed := elapsed - fadeStartDelay
				fadeProgress := float64(fadeElapsed) / float64(fadeDuration)
				if fadeProgress > 1.0 {
					fadeProgress = 1.0
				}
				// Opacity goes from 1.0 to 0.0
				m.distractionFreeOpacity = 1.0 - fadeProgress
			}

			// Continue fade animation while playing
			m.distractionFadeActive = true
			return m, distractionFadeTickCmd()
		} else {
			// Not in play mode or paused - restore full visibility
			m.distractionFreeOpacity = 1.0
			m.distractionFadeActive = false
		}
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case StateFilePicker:
			return m.handleFilePickerInput(msg)
		case StateResumeList:
			return m.handleResumeListInput(msg)
		case StateReading:
			return m.handleReadingInput(msg)
		case StateSettings:
			return m.handleSettingsInput(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.renderer.SetSize(msg.Width, msg.Height)

		// Update file browser size
		if m.fileBrowser != nil {
			m.fileBrowser.SetSize(msg.Width, msg.Height)
		}

	case tea.MouseMsg:
		if m.state == StateReading {
			return m.handleMouseInput(msg)
		}
	}

	return m, nil
}

// handleMouseInput processes mouse events during reading
func (m Model) handleMouseInput(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonLeft:
		switch msg.Action {
		case tea.MouseActionPress:
			// Check if click is on progress bar
			wordIndex := m.renderer.CalculateWordIndexFromClick(msg.X, msg.Y)
			if wordIndex >= 0 {
				// Start dragging
				m.isDragging = true
				// Remember pause state and pause during drag
				if m.readingState != nil {
					m.wasPausedBeforeDrag = m.readingState.IsPaused
				}
				m.engine.Pause()
				// Jump to position
				m.engine.JumpToWord(wordIndex)
			}

		case tea.MouseActionRelease:
			if m.isDragging {
				m.isDragging = false
				// Resume if it was playing before drag started
				if !m.wasPausedBeforeDrag {
					m.engine.Play()
				}
			}

		case tea.MouseActionMotion:
			// Handle drag motion
			if m.isDragging {
				wordIndex := m.renderer.CalculateWordIndexFromClick(msg.X, msg.Y)
				if wordIndex >= 0 {
					m.engine.JumpToWord(wordIndex)
				}
			}
		}

	case tea.MouseButtonWheelUp:
		// Scroll up = go back in document
		if m.readingState != nil && m.readingState.WordIndex > 0 {
			newIndex := m.readingState.WordIndex - 10
			if newIndex < 0 {
				newIndex = 0
			}
			m.engine.JumpToWord(newIndex)
		}

	case tea.MouseButtonWheelDown:
		// Scroll down = go forward in document
		if m.readingState != nil && m.readingState.WordIndex < m.readingState.TotalWords-1 {
			newIndex := m.readingState.WordIndex + 10
			if newIndex >= m.readingState.TotalWords {
				newIndex = m.readingState.TotalWords - 1
			}
			m.engine.JumpToWord(newIndex)
		}
	}

	return m, nil
}

// View renders the current state
func (m Model) View() string {
	switch m.state {
	case StateFilePicker:
		return m.renderFileBrowser()
	case StateResumeList:
		return m.renderer.RenderResumeList(m.savedSessions, m.resumeCursor, m.width, m.height)
	case StateReading:
		if m.documentLoading {
			return m.renderer.RenderLoading(m.width, m.height)
		}
		if m.loadError != "" {
			return m.renderer.RenderError(m.loadError, m.width, m.height)
		}
		// Render reading screen with distraction-free opacity
		baseView := m.renderer.RenderReadingScreen(m.readingState, m.animator, m.width, m.height, m.distractionFreeOpacity)

		// Overlay go-to input if in go-to mode
		if m.gotoMode {
			totalWords := 0
			if m.readingState != nil {
				totalWords = m.readingState.TotalWords
			}
			gotoOverlay := m.renderer.RenderGotoOverlay(m.gotoInput, totalWords)
			return overlayContent(baseView, gotoOverlay, m.width, m.height)
		}
		return baseView
	case StateSettings:
		return m.renderer.RenderSettings(m.settingsCursor, m.width, m.height)
	}
	return ""
}

// renderFileBrowser renders the file browser centered on screen
func (m Model) renderFileBrowser() string {
	if m.fileBrowser == nil {
		return ""
	}

	browserView := m.fileBrowser.View()

	// Center horizontally and vertically
	return centerContent(browserView, m.width, m.height)
}

// overlayContent places an overlay centered on top of base content
func overlayContent(base, overlay string, width, height int) string {
	// Split base and overlay into lines
	baseLines := splitLines(base)
	overlayLines := splitLines(overlay)

	// Calculate overlay dimensions
	overlayHeight := len(overlayLines)
	overlayWidth := maxLineWidth(overlayLines)

	// Calculate starting position for overlay (centered)
	startY := (height - overlayHeight) / 2
	if startY < 0 {
		startY = 0
	}
	startX := (width - overlayWidth) / 2
	if startX < 0 {
		startX = 0
	}

	// Overlay the content
	for i, overlayLine := range overlayLines {
		y := startY + i
		if y < len(baseLines) {
			baseLine := baseLines[y]
			// Insert overlay line into base line
			baseLines[y] = insertAt(baseLine, overlayLine, startX)
		}
	}

	return joinLines(baseLines)
}

// splitLines splits content into individual lines
func splitLines(content string) []string {
	return strings.Split(content, "\n")
}

// joinLines joins lines back into content
func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

// maxLineWidth returns the maximum width of any line (visible characters only)
func maxLineWidth(lines []string) int {
	maxWidth := 0
	for _, line := range lines {
		width := visibleWidth(line)
		if width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}

// visibleWidth returns the visible width of a string (excluding ANSI escape codes)
func visibleWidth(s string) int {
	// Simple approximation - count runes, excluding ANSI escape sequences
	width := 0
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
		width++
	}
	return width
}

// insertAt inserts overlay text at position x in the base line
func insertAt(base, overlay string, x int) string {
	// Ensure base line is long enough
	baseWidth := visibleWidth(base)
	if baseWidth < x {
		// Pad base with spaces
		base += strings.Repeat(" ", x-baseWidth)
	}

	overlayWidth := visibleWidth(overlay)

	// Simple replacement: just truncate base after x and append overlay
	// This is a simplified version - full implementation would preserve ANSI codes
	result := padToWidth(base, x) + overlay

	// Append rest of base line after overlay if it extends beyond
	restStart := x + overlayWidth
	if restStart < baseWidth {
		// Get the part of base after the overlay
		baseRunes := []rune(base)
		if restStart < len(baseRunes) {
			result += string(baseRunes[restStart:])
		}
	}

	return result
}

// padToWidth returns the first n visible characters of a string
func padToWidth(s string, n int) string {
	var result strings.Builder
	width := 0
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			result.WriteRune(r)
			continue
		}
		if inEscape {
			result.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		if width >= n {
			break
		}
		result.WriteRune(r)
		width++
	}
	// Pad with spaces if needed
	for width < n {
		result.WriteRune(' ')
		width++
	}
	return result.String()
}

// centerContent centers content on screen
func centerContent(content string, width, height int) string {
	// Count lines in content
	lines := countLines(content)

	// Calculate padding
	topPadding := (height - lines) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	var result string
	for i := 0; i < topPadding; i++ {
		result += "\n"
	}
	result += content

	return result
}

func countLines(s string) int {
	count := 1
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}

// handleFilePickerInput processes keyboard input in file picker
func (m Model) handleFilePickerInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEsc:
		return m, tea.Quit

	case tea.KeyUp, tea.KeyShiftTab:
		m.fileBrowser.MoveUp()
		return m, nil

	case tea.KeyDown:
		m.fileBrowser.MoveDown()
		return m, nil

	case tea.KeyTab:
		// Tab switches to resume list if we have saved sessions
		sessions, _ := m.engine.GetSavedSessions()
		if len(sessions) > 0 {
			m.savedSessions = sessions
			m.resumeCursor = 0
			m.state = StateResumeList
			return m, nil
		}
		return m, nil

	case tea.KeyPgUp:
		m.fileBrowser.PageUp()
		return m, nil

	case tea.KeyPgDown:
		m.fileBrowser.PageDown()
		return m, nil

	case tea.KeyHome:
		m.fileBrowser.GoHome()
		return m, nil

	case tea.KeyEnd:
		m.fileBrowser.GoEnd()
		return m, nil

	case tea.KeyEnter:
		if m.fileBrowser.Enter() {
			// File was selected
			m.selectedFile = m.fileBrowser.SelectedFile
			m.documentLoading = true
			m.state = StateReading
			return m, m.loadDocumentCmd(m.selectedFile)
		}
		return m, nil

	case tea.KeyBackspace:
		m.fileBrowser.GoBack()
		return m, nil

	case tea.KeyRunes:
		char := string(msg.Runes)
		switch char {
		case "q", "Q":
			return m, tea.Quit
		case "h", "H":
			m.fileBrowser.GoBack()
		case "j":
			m.fileBrowser.MoveDown()
		case "k":
			m.fileBrowser.MoveUp()
		case "l":
			if m.fileBrowser.Enter() {
				m.selectedFile = m.fileBrowser.SelectedFile
				m.documentLoading = true
				m.state = StateReading
				return m, m.loadDocumentCmd(m.selectedFile)
			}
		case "g":
			m.fileBrowser.GoHome()
		case "G":
			m.fileBrowser.GoEnd()
		case ".":
			m.fileBrowser.ToggleHidden()
		case "~":
			// Go to home directory
			if home, err := os.UserHomeDir(); err == nil {
				m.fileBrowser.CurrentDir = home
				m.fileBrowser.Refresh()
			}
		}
		return m, nil
	}

	return m, nil
}

// handleResumeListInput processes keyboard input in resume list
func (m Model) handleResumeListInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		m.state = StateFilePicker
		return m, nil

	case tea.KeyUp, tea.KeyShiftTab:
		if m.resumeCursor > 0 {
			m.resumeCursor--
		} else {
			m.resumeCursor = len(m.savedSessions) - 1
		}

	case tea.KeyDown, tea.KeyTab:
		if m.resumeCursor < len(m.savedSessions)-1 {
			m.resumeCursor++
		} else {
			m.resumeCursor = 0
		}

	case tea.KeyEnter:
		if m.resumeCursor < len(m.savedSessions) {
			session := m.savedSessions[m.resumeCursor]
			m.documentLoading = true
			m.state = StateReading
			return m, m.resumeSessionCmd(session.DocumentHash)
		}

	case tea.KeyRunes:
		char := string(msg.Runes)
		switch char {
		case "j":
			if m.resumeCursor < len(m.savedSessions)-1 {
				m.resumeCursor++
			} else {
				m.resumeCursor = 0
			}
		case "k":
			if m.resumeCursor > 0 {
				m.resumeCursor--
			} else {
				m.resumeCursor = len(m.savedSessions) - 1
			}
		}
	}

	return m, nil
}

// handleReadingInput processes keyboard input during reading
func (m Model) handleReadingInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Reset keyboard activity time on any key press (for distraction-free mode)
	m.lastKeyboardActivity = time.Now()
	m.distractionFreeOpacity = 1.0

	// Start fade timer if not already running and playing
	var fadeCmd tea.Cmd
	if !m.distractionFadeActive && m.readingState != nil && !m.readingState.IsPaused {
		fadeCmd = distractionFadeTickCmd()
		m.distractionFadeActive = true
	}

	// Handle go-to mode input
	if m.gotoMode {
		newModel, cmd := m.handleGotoModeInput(msg)
		if fadeCmd != nil {
			return newModel, tea.Batch(cmd, fadeCmd)
		}
		return newModel, cmd
	}

	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEsc:
		// Save position and return to file picker
		m.engine.SavePosition()
		m.engine.Pause()
		m.state = StateFilePicker
		m.loadError = ""
		return m, nil

	case tea.KeySpace:
		// Toggle play/pause
		m.engine.Toggle()
		// After toggling, check if we're now playing and start fade timer
		// We need to wait for state to update, so start timer regardless
		// (it will check play state and do nothing if paused)
		if !m.distractionFadeActive {
			m.distractionFadeActive = true
			return m, distractionFadeTickCmd()
		}
		return m, fadeCmd

	case tea.KeyRunes:
		char := string(msg.Runes)
		switch char {
		case "q":
			m.engine.SavePosition()
			return m, tea.Quit

		case "j", "J":
			// Decrease speed
			if m.readingState != nil {
				newWPM := m.readingState.WPM - 50
				if newWPM < 50 {
					newWPM = 50
				}
				m.engine.SetWPM(newWPM)
			}

		case "k", "K":
			// Increase speed
			if m.readingState != nil {
				newWPM := m.readingState.WPM + 50
				if newWPM > 2000 {
					newWPM = 2000
				}
				m.engine.SetWPM(newWPM)
			}

		case "h", "H":
			// Previous paragraph
			m.engine.PrevParagraph()

		case "l", "L":
			// Next paragraph
			m.engine.NextParagraph()

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Speed presets
			preset := int(msg.Runes[0] - '0')
			wpm := 100 + (preset * 100) // 1=200, 2=300, ..., 9=1000
			m.engine.SetWPM(wpm)

		case "s", "S":
			// Save position
			m.engine.SavePosition()

		case "r", "R":
			// Return to start
			m.engine.ReturnToStart()
			m.engine.Pause()

		case "g":
			// Enter go-to mode
			m.gotoMode = true
			m.gotoInput = ""
			// Pause while entering position
			if m.readingState != nil && !m.readingState.IsPaused {
				m.wasPausedBeforeDrag = false
				m.engine.Pause()
			} else {
				m.wasPausedBeforeDrag = true
			}

		case "c", "C":
			// Toggle all caps display
			_, _ = settings.ToggleAllCaps()
		}

	case tea.KeyLeft:
		// Decrease speed
		if m.readingState != nil {
			newWPM := m.readingState.WPM - 50
			if newWPM < 50 {
				newWPM = 50
			}
			m.engine.SetWPM(newWPM)
		}

	case tea.KeyRight:
		// Increase speed
		if m.readingState != nil {
			newWPM := m.readingState.WPM + 50
			if newWPM > 2000 {
				newWPM = 2000
			}
			m.engine.SetWPM(newWPM)
		}

	case tea.KeyUp:
		// Previous paragraph
		m.engine.PrevParagraph()

	case tea.KeyDown:
		// Next paragraph
		m.engine.NextParagraph()
	}

	return m, fadeCmd
}

// handleGotoModeInput handles keyboard input when in go-to percentage mode
func (m Model) handleGotoModeInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Cancel go-to mode
		m.gotoMode = false
		m.gotoInput = ""
		// Resume if it was playing before
		if !m.wasPausedBeforeDrag {
			m.engine.Play()
		}
		return m, nil

	case tea.KeyEnter:
		// Execute the jump
		if m.gotoInput != "" && m.readingState != nil && m.readingState.TotalWords > 0 {
			percentage, err := strconv.ParseFloat(m.gotoInput, 64)
			if err == nil {
				// Clamp percentage to 0-100
				if percentage < 0 {
					percentage = 0
				}
				if percentage > 100 {
					percentage = 100
				}
				// Calculate word index
				wordIndex := int(percentage / 100.0 * float64(m.readingState.TotalWords))
				if wordIndex >= m.readingState.TotalWords {
					wordIndex = m.readingState.TotalWords - 1
				}
				m.engine.JumpToWord(wordIndex)
			}
		}
		m.gotoMode = false
		m.gotoInput = ""
		// Resume if it was playing before
		if !m.wasPausedBeforeDrag {
			m.engine.Play()
		}
		return m, nil

	case tea.KeyBackspace:
		// Delete last character
		if len(m.gotoInput) > 0 {
			m.gotoInput = m.gotoInput[:len(m.gotoInput)-1]
		}
		return m, nil

	case tea.KeyRunes:
		// Only accept digits and decimal point
		for _, r := range msg.Runes {
			if unicode.IsDigit(r) || (r == '.' && !containsRune(m.gotoInput, '.')) {
				// Limit input length
				if len(m.gotoInput) < 6 {
					m.gotoInput += string(r)
				}
			}
		}
		return m, nil
	}

	return m, nil
}

// containsRune checks if a string contains a specific rune
func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

// handleSettingsInput processes keyboard input in settings
func (m Model) handleSettingsInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEsc:
		m.state = StateReading
		return m, nil
	}

	return m, nil
}

// loadDocumentCmd returns a command to load a document
func (m *Model) loadDocumentCmd(path string) tea.Cmd {
	engine := m.engine
	return func() tea.Msg {
		err := engine.LoadDocument(path)
		return documentLoadedMsg{err: err}
	}
}

// resumeSessionCmd returns a command to resume a saved session
func (m *Model) resumeSessionCmd(hash string) tea.Cmd {
	engine := m.engine
	return func() tea.Msg {
		err := engine.ResumeSession(hash)
		return documentLoadedMsg{err: err}
	}
}

// tickCmd returns a command that sends tick messages for state polling
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// animTickCmd returns a command that sends animation tick messages
func animTickCmd() tea.Cmd {
	return tea.Tick(GetAnimationInterval(), func(t time.Time) tea.Msg {
		return animTickMsg(t)
	})
}

// distractionFadeTickCmd returns a command for distraction-free fade animation
func distractionFadeTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return distractionFadeTickMsg(t)
	})
}
