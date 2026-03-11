// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

// Package backend provides the reading engine and API for the RSVP speed reading application.
// This package handles document management, reading state, and timing control.
// The frontend communicates exclusively through the ReaderAPI interface.
package backend

import (
	"time"

	"github.com/timlinux/cheetah/documents"
)

// ReaderAPI defines the public interface for the reading backend.
// All frontend interactions with the reading engine go through this interface.
type ReaderAPI interface {
	// Document Management
	// -------------------

	// LoadDocument loads a document from the filesystem
	LoadDocument(path string) error

	// LoadDocumentBytes loads a document from bytes (for web uploads)
	LoadDocumentBytes(data []byte, filename string) error

	// GetDocumentInfo returns metadata about the loaded document
	GetDocumentInfo() *DocumentInfo

	// Reading Control
	// ---------------

	// Play starts or resumes reading
	Play()

	// Pause pauses reading
	Pause()

	// Toggle toggles between play and pause
	Toggle()

	// SetWPM sets the reading speed in words per minute
	SetWPM(wpm int)

	// GetWPM returns the current WPM setting
	GetWPM() int

	// Navigation
	// ----------

	// JumpToWord moves to a specific word index
	JumpToWord(index int)

	// JumpToParagraph moves to the start of a specific paragraph
	JumpToParagraph(index int)

	// NextParagraph moves to the start of the next paragraph
	NextParagraph()

	// PrevParagraph moves to the start of the previous paragraph
	PrevParagraph()

	// State
	// -----

	// GetState returns the current reading state
	GetState() ReadingState

	// SubscribeState returns a channel for state updates
	SubscribeState() <-chan ReadingState

	// UnsubscribeState removes a state subscription
	UnsubscribeState(ch <-chan ReadingState)

	// Persistence
	// -----------

	// SavePosition saves the current reading position
	SavePosition() error

	// GetSavedSessions returns all saved reading sessions
	GetSavedSessions() []SavedSession
}

// ReadingState contains the current state of reading
type ReadingState struct {
	// CurrentWord is the word currently being displayed
	CurrentWord string

	// PreviousWord is the word that was just displayed
	PreviousWord string

	// NextWords contains the next 3 upcoming words
	NextWords []string

	// WordIndex is the current word position (0-based)
	WordIndex int

	// TotalWords is the total number of words
	TotalWords int

	// ParagraphIndex is the current paragraph (0-based)
	ParagraphIndex int

	// TotalParagraphs is the total number of paragraphs
	TotalParagraphs int

	// WPM is the current words per minute setting
	WPM int

	// IsPaused indicates whether reading is paused
	IsPaused bool

	// Progress is the reading progress (0.0 to 1.0)
	Progress float64

	// ElapsedMs is the elapsed reading time in milliseconds
	ElapsedMs int64

	// DocumentLoaded indicates whether a document is loaded
	DocumentLoaded bool

	// DocumentTitle is the title of the loaded document
	DocumentTitle string
}

// DocumentInfo contains metadata about a loaded document
type DocumentInfo struct {
	// Title is the document title
	Title string

	// Path is the file path (if loaded from file)
	Path string

	// TotalWords is the total word count
	TotalWords int

	// TotalParagraphs is the total paragraph count
	TotalParagraphs int

	// Hash is the content hash for identification
	Hash string
}

// SavedSession represents a saved reading position
type SavedSession struct {
	// DocumentHash is the unique identifier for the document
	DocumentHash string `json:"document_hash"`

	// DocumentPath is the file path
	DocumentPath string `json:"document_path"`

	// DocumentTitle is the document title
	DocumentTitle string `json:"document_title"`

	// LastPosition is the last word index
	LastPosition int `json:"last_position"`

	// TotalWords is the total word count
	TotalWords int `json:"total_words"`

	// LastWPM is the last used WPM setting
	LastWPM int `json:"last_wpm"`

	// LastAccessed is when the session was last accessed
	LastAccessed time.Time `json:"last_accessed"`
}

// Config holds configuration options for the reading engine
type Config struct {
	// DefaultWPM is the default reading speed
	DefaultWPM int

	// MinWPM is the minimum allowed WPM
	MinWPM int

	// MaxWPM is the maximum allowed WPM
	MaxWPM int

	// PunctuationDelay adds extra time after punctuation (multiplier)
	PunctuationDelay float64

	// ParagraphDelay adds extra time between paragraphs (multiplier)
	ParagraphDelay float64
}

// DefaultConfig returns the default reading engine configuration
func DefaultConfig() Config {
	return Config{
		DefaultWPM:       300,
		MinWPM:           50,
		MaxWPM:           2000,
		PunctuationDelay: 1.3,  // 30% extra time after punctuation
		ParagraphDelay:   2.0,  // Double time between paragraphs
	}
}

// NewEngine creates a new reading engine with the given configuration
func NewEngine(config Config) *Engine {
	return &Engine{
		config:      config,
		wpm:         config.DefaultWPM,
		isPaused:    true,
		subscribers: make(map[chan ReadingState]struct{}),
	}
}

// GetDocument returns the currently loaded document
func (e *Engine) GetDocument() *documents.Document {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.document
}
