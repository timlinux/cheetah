// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package frontend

import "github.com/timlinux/cheetah/backend"

// EngineAdapter defines the interface for reading engine access.
// This allows the TUI to work with either an embedded engine (standalone mode)
// or an HTTP client (server mode).
type EngineAdapter interface {
	// LoadDocument loads a document from the filesystem
	LoadDocument(path string) error

	// GetState returns the current reading state
	GetState() (*backend.ReadingState, error)

	// Play starts reading
	Play() error

	// Pause pauses reading
	Pause() error

	// Toggle toggles play/pause
	Toggle() error

	// SetWPM sets the reading speed
	SetWPM(wpm int) error

	// PrevParagraph moves to the previous paragraph
	PrevParagraph() error

	// NextParagraph moves to the next paragraph
	NextParagraph() error

	// JumpToWord moves to a specific word index
	JumpToWord(index int) error

	// ReturnToStart returns to the beginning
	ReturnToStart() error

	// SavePosition saves the current reading position
	SavePosition() error

	// GetSavedSessions returns all saved sessions
	GetSavedSessions() ([]backend.SavedSession, error)

	// ResumeSession resumes a saved session
	ResumeSession(hash string) error

	// Close cleans up resources
	Close() error
}

// EmbeddedEngine wraps the backend Engine for direct use without HTTP
type EmbeddedEngine struct {
	engine *backend.Engine
}

// NewEmbeddedEngine creates a new embedded engine adapter
func NewEmbeddedEngine() *EmbeddedEngine {
	config := backend.DefaultConfig()
	return &EmbeddedEngine{
		engine: backend.NewEngine(config),
	}
}

// LoadDocument loads a document from the filesystem
func (e *EmbeddedEngine) LoadDocument(path string) error {
	return e.engine.LoadDocument(path)
}

// GetState returns the current reading state
func (e *EmbeddedEngine) GetState() (*backend.ReadingState, error) {
	state := e.engine.GetState()
	return &state, nil
}

// Play starts reading
func (e *EmbeddedEngine) Play() error {
	e.engine.Play()
	return nil
}

// Pause pauses reading
func (e *EmbeddedEngine) Pause() error {
	e.engine.Pause()
	return nil
}

// Toggle toggles play/pause
func (e *EmbeddedEngine) Toggle() error {
	e.engine.Toggle()
	return nil
}

// SetWPM sets the reading speed
func (e *EmbeddedEngine) SetWPM(wpm int) error {
	e.engine.SetWPM(wpm)
	return nil
}

// PrevParagraph moves to the previous paragraph
func (e *EmbeddedEngine) PrevParagraph() error {
	e.engine.PrevParagraph()
	return nil
}

// NextParagraph moves to the next paragraph
func (e *EmbeddedEngine) NextParagraph() error {
	e.engine.NextParagraph()
	return nil
}

// JumpToWord moves to a specific word index
func (e *EmbeddedEngine) JumpToWord(index int) error {
	e.engine.JumpToWord(index)
	return nil
}

// ReturnToStart returns to the beginning
func (e *EmbeddedEngine) ReturnToStart() error {
	e.engine.JumpToWord(0)
	return nil
}

// SavePosition saves the current reading position
func (e *EmbeddedEngine) SavePosition() error {
	return e.engine.SavePosition()
}

// GetSavedSessions returns all saved sessions
func (e *EmbeddedEngine) GetSavedSessions() ([]backend.SavedSession, error) {
	return e.engine.GetSavedSessions(), nil
}

// ResumeSession resumes a saved session
func (e *EmbeddedEngine) ResumeSession(hash string) error {
	return e.engine.ResumeSession(hash)
}

// Close cleans up resources
func (e *EmbeddedEngine) Close() error {
	e.engine.Pause()
	return nil
}
