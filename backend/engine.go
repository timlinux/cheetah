// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package backend

import (
	"strings"
	"sync"
	"time"

	"github.com/timlinux/cheetah/documents"
	"github.com/timlinux/cheetah/sessions"
)

// Engine implements the ReaderAPI interface and handles RSVP reading
type Engine struct {
	config Config

	// Document state
	document  *documents.Document
	words     []string
	wordIndex int

	// Timing state
	wpm       int
	isPaused  bool
	startTime time.Time
	elapsed   time.Duration
	ticker    *time.Ticker
	stopChan  chan struct{}

	// Subscribers for state updates
	subscribers map[chan ReadingState]struct{}

	mu sync.RWMutex
}

// LoadDocument loads a document from the filesystem
func (e *Engine) LoadDocument(path string) error {
	doc, err := documents.ParseFile(path)
	if err != nil {
		return err
	}

	e.mu.Lock()
	e.document = doc
	e.words = documents.GetAllWords(doc)
	e.wordIndex = 0
	e.isPaused = true
	e.elapsed = 0
	e.mu.Unlock()

	e.notifySubscribers()
	return nil
}

// LoadDocumentBytes loads a document from bytes
func (e *Engine) LoadDocumentBytes(data []byte, filename string) error {
	parser, err := documents.GetParser(filename)
	if err != nil {
		return err
	}

	doc, err := parser.ParseBytes(data, filename)
	if err != nil {
		return err
	}

	e.mu.Lock()
	e.document = doc
	e.words = documents.GetAllWords(doc)
	e.wordIndex = 0
	e.isPaused = true
	e.elapsed = 0
	e.mu.Unlock()

	e.notifySubscribers()
	return nil
}

// GetDocumentInfo returns metadata about the loaded document
func (e *Engine) GetDocumentInfo() *DocumentInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.document == nil {
		return nil
	}

	return &DocumentInfo{
		Title:           e.document.Title,
		Path:            e.document.Path,
		TotalWords:      e.document.TotalWords,
		TotalParagraphs: len(e.document.Paragraphs),
		Hash:            e.document.Hash,
	}
}

// Play starts or resumes reading
func (e *Engine) Play() {
	e.mu.Lock()
	if e.document == nil || !e.isPaused {
		e.mu.Unlock()
		return
	}

	e.isPaused = false
	e.startTime = time.Now()

	// Start the ticker goroutine
	e.stopChan = make(chan struct{})
	e.mu.Unlock()

	go e.runTicker()
	e.notifySubscribers()
}

// Pause pauses reading
func (e *Engine) Pause() {
	e.mu.Lock()
	if e.isPaused {
		e.mu.Unlock()
		return
	}

	e.isPaused = true
	e.elapsed += time.Since(e.startTime)

	// Stop the ticker
	if e.stopChan != nil {
		close(e.stopChan)
		e.stopChan = nil
	}
	e.mu.Unlock()

	e.notifySubscribers()
}

// Toggle toggles between play and pause
func (e *Engine) Toggle() {
	e.mu.RLock()
	isPaused := e.isPaused
	e.mu.RUnlock()

	if isPaused {
		e.Play()
	} else {
		e.Pause()
	}
}

// SetWPM sets the reading speed
func (e *Engine) SetWPM(wpm int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if wpm < e.config.MinWPM {
		wpm = e.config.MinWPM
	}
	if wpm > e.config.MaxWPM {
		wpm = e.config.MaxWPM
	}

	e.wpm = wpm
}

// GetWPM returns the current WPM setting
func (e *Engine) GetWPM() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.wpm
}

// JumpToWord moves to a specific word index
func (e *Engine) JumpToWord(index int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if index < 0 {
		index = 0
	}
	if index >= len(e.words) {
		index = len(e.words) - 1
	}

	e.wordIndex = index
	e.notifySubscribersLocked()
}

// JumpToParagraph moves to the start of a specific paragraph
func (e *Engine) JumpToParagraph(index int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.document == nil {
		return
	}

	if index < 0 {
		index = 0
	}
	if index >= len(e.document.Paragraphs) {
		index = len(e.document.Paragraphs) - 1
	}

	e.wordIndex = documents.GetParagraphStartIndex(e.document, index)
	e.notifySubscribersLocked()
}

// NextParagraph moves to the start of the next paragraph
func (e *Engine) NextParagraph() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.document == nil {
		return
	}

	currentPara := documents.GetParagraphForWordIndex(e.document, e.wordIndex)
	nextPara := currentPara + 1

	if nextPara >= len(e.document.Paragraphs) {
		return
	}

	e.wordIndex = documents.GetParagraphStartIndex(e.document, nextPara)
	e.notifySubscribersLocked()
}

// PrevParagraph moves to the start of the previous paragraph
func (e *Engine) PrevParagraph() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.document == nil {
		return
	}

	currentPara := documents.GetParagraphForWordIndex(e.document, e.wordIndex)

	// If we're at the start of a paragraph, go to the previous one
	// Otherwise, go to the start of the current paragraph
	paraStart := documents.GetParagraphStartIndex(e.document, currentPara)
	if e.wordIndex == paraStart && currentPara > 0 {
		e.wordIndex = documents.GetParagraphStartIndex(e.document, currentPara-1)
	} else {
		e.wordIndex = paraStart
	}

	e.notifySubscribersLocked()
}

// GetState returns the current reading state
func (e *Engine) GetState() ReadingState {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.getStateLocked()
}

func (e *Engine) getStateLocked() ReadingState {
	state := ReadingState{
		WPM:      e.wpm,
		IsPaused: e.isPaused,
	}

	if e.document == nil {
		return state
	}

	state.DocumentLoaded = true
	state.DocumentTitle = e.document.Title
	state.TotalWords = len(e.words)
	state.TotalParagraphs = len(e.document.Paragraphs)
	state.WordIndex = e.wordIndex

	if e.wordIndex < len(e.words) {
		state.CurrentWord = e.words[e.wordIndex]
	}

	if e.wordIndex > 0 {
		state.PreviousWord = e.words[e.wordIndex-1]
	}

	// Get next 3 words
	for i := 1; i <= 3; i++ {
		idx := e.wordIndex + i
		if idx < len(e.words) {
			state.NextWords = append(state.NextWords, e.words[idx])
		}
	}

	state.ParagraphIndex = documents.GetParagraphForWordIndex(e.document, e.wordIndex)

	if state.TotalWords > 0 {
		state.Progress = float64(e.wordIndex) / float64(state.TotalWords)
	}

	elapsed := e.elapsed
	if !e.isPaused && !e.startTime.IsZero() {
		elapsed += time.Since(e.startTime)
	}
	state.ElapsedMs = elapsed.Milliseconds()

	return state
}

// SubscribeState returns a channel for state updates
func (e *Engine) SubscribeState() <-chan ReadingState {
	e.mu.Lock()
	defer e.mu.Unlock()

	ch := make(chan ReadingState, 10)
	e.subscribers[ch] = struct{}{}
	return ch
}

// UnsubscribeState removes a state subscription
func (e *Engine) UnsubscribeState(ch <-chan ReadingState) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Find and remove the channel
	for c := range e.subscribers {
		if c == ch {
			delete(e.subscribers, c)
			close(c)
			break
		}
	}
}

// SavePosition saves the current reading position
func (e *Engine) SavePosition() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.document == nil {
		return nil
	}

	session := sessions.Session{
		DocumentHash:  e.document.Hash,
		DocumentPath:  e.document.Path,
		DocumentTitle: e.document.Title,
		LastPosition:  e.wordIndex,
		TotalWords:    len(e.words),
		LastWPM:       e.wpm,
		LastAccessed:  time.Now(),
	}

	return sessions.Save(session)
}

// GetSavedSessions returns all saved reading sessions
func (e *Engine) GetSavedSessions() []SavedSession {
	stored := sessions.LoadAll()
	result := make([]SavedSession, len(stored))

	for i, s := range stored {
		result[i] = SavedSession{
			DocumentHash:  s.DocumentHash,
			DocumentPath:  s.DocumentPath,
			DocumentTitle: s.DocumentTitle,
			LastPosition:  s.LastPosition,
			TotalWords:    s.TotalWords,
			LastWPM:       s.LastWPM,
			LastAccessed:  s.LastAccessed,
		}
	}

	return result
}

// runTicker runs the word advancement ticker
func (e *Engine) runTicker() {
	for {
		e.mu.RLock()
		if e.isPaused || e.document == nil {
			e.mu.RUnlock()
			return
		}

		wpm := e.wpm
		wordIndex := e.wordIndex
		words := e.words
		stopChan := e.stopChan
		e.mu.RUnlock()

		if wordIndex >= len(words) {
			// End of document
			e.Pause()
			return
		}

		// Calculate delay for this word
		delay := e.calculateWordDelay(words[wordIndex], wpm)

		select {
		case <-stopChan:
			return
		case <-time.After(delay):
			e.mu.Lock()
			if e.isPaused {
				e.mu.Unlock()
				return
			}
			e.wordIndex++
			if e.wordIndex >= len(e.words) {
				e.isPaused = true
				e.elapsed += time.Since(e.startTime)
			}
			e.notifySubscribersLocked()
			e.mu.Unlock()
		}
	}
}

// calculateWordDelay calculates the delay for a word based on WPM and punctuation
func (e *Engine) calculateWordDelay(word string, wpm int) time.Duration {
	// Base delay: 60000ms / WPM = ms per word
	baseDelay := float64(60000) / float64(wpm)

	// Apply punctuation delay
	if strings.HasSuffix(word, ".") || strings.HasSuffix(word, "!") || strings.HasSuffix(word, "?") {
		baseDelay *= e.config.PunctuationDelay
	} else if strings.HasSuffix(word, ",") || strings.HasSuffix(word, ";") || strings.HasSuffix(word, ":") {
		baseDelay *= (1 + (e.config.PunctuationDelay-1)/2) // Half the extra delay
	}

	return time.Duration(baseDelay) * time.Millisecond
}

// notifySubscribers sends state updates to all subscribers
func (e *Engine) notifySubscribers() {
	e.mu.RLock()
	defer e.mu.RUnlock()
	e.notifySubscribersLocked()
}

func (e *Engine) notifySubscribersLocked() {
	state := e.getStateLocked()
	for ch := range e.subscribers {
		select {
		case ch <- state:
		default:
			// Channel full, skip
		}
	}
}

// ResumeSession resumes a saved session
func (e *Engine) ResumeSession(hash string) error {
	session, err := sessions.Load(hash)
	if err != nil {
		return err
	}

	// Load the document
	if err := e.LoadDocument(session.DocumentPath); err != nil {
		return err
	}

	// Restore position and WPM
	e.mu.Lock()
	e.wordIndex = session.LastPosition
	e.wpm = session.LastWPM
	e.mu.Unlock()

	e.notifySubscribers()
	return nil
}
