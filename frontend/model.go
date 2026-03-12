// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package frontend

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/timlinux/cheetah/backend"
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
		engine:       engine,
		state:        initialState,
		renderer:     NewRenderer(80, 24),
		animator:     NewWordAnimator(),
		fileBrowser:  fb,
		initialPath:  documentPath,
		initialWPM:   initialWPM,
		readingState: &backend.ReadingState{IsPaused: true, WPM: initialWPM},
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
		engine:       engine,
		state:        initialState,
		renderer:     NewRenderer(80, 24),
		animator:     NewWordAnimator(),
		fileBrowser:  fb,
		initialPath:  documentPath,
		initialWPM:   initialWPM,
		readingState: &backend.ReadingState{IsPaused: true, WPM: initialWPM},
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
			// Set initial WPM
			if m.initialWPM > 0 {
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
		return m.renderer.RenderReadingScreen(m.readingState, m.animator, m.width, m.height)
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
		return m, nil

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

	return m, nil
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
