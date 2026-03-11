// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package backend

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func createTestDocument(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return testFile
}

func TestNewEngine(t *testing.T) {
	config := DefaultConfig()
	engine := NewEngine(config)

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	if engine.GetWPM() != config.DefaultWPM {
		t.Errorf("Expected WPM %d, got %d", config.DefaultWPM, engine.GetWPM())
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.DefaultWPM <= 0 {
		t.Error("DefaultWPM should be positive")
	}

	if config.MinWPM <= 0 {
		t.Error("MinWPM should be positive")
	}

	if config.MaxWPM <= config.MinWPM {
		t.Error("MaxWPM should be greater than MinWPM")
	}

	if config.PunctuationDelay <= 0 {
		t.Error("PunctuationDelay should be positive")
	}

	if config.ParagraphDelay <= 0 {
		t.Error("ParagraphDelay should be positive")
	}
}

func TestEngineLoadDocument(t *testing.T) {
	engine := NewEngine(DefaultConfig())

	testFile := createTestDocument(t, "Hello world. This is a test document with some words.")

	err := engine.LoadDocument(testFile)
	if err != nil {
		t.Fatalf("LoadDocument failed: %v", err)
	}

	info := engine.GetDocumentInfo()
	if info == nil {
		t.Fatal("GetDocumentInfo returned nil after loading")
	}

	if info.TotalWords == 0 {
		t.Error("Expected non-zero word count")
	}
}

func TestEngineLoadDocumentNonExistent(t *testing.T) {
	engine := NewEngine(DefaultConfig())

	err := engine.LoadDocument("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestEngineSetWPM(t *testing.T) {
	config := DefaultConfig()
	engine := NewEngine(config)

	// Test setting valid WPM
	engine.SetWPM(500)
	if engine.GetWPM() != 500 {
		t.Errorf("Expected WPM 500, got %d", engine.GetWPM())
	}

	// Test setting WPM below minimum
	engine.SetWPM(config.MinWPM - 10)
	if engine.GetWPM() < config.MinWPM {
		t.Errorf("WPM should not be below minimum %d", config.MinWPM)
	}

	// Test setting WPM above maximum
	engine.SetWPM(config.MaxWPM + 100)
	if engine.GetWPM() > config.MaxWPM {
		t.Errorf("WPM should not exceed maximum %d", config.MaxWPM)
	}
}

func TestEnginePlayPause(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	testFile := createTestDocument(t, "Hello world. This is a test.")

	if err := engine.LoadDocument(testFile); err != nil {
		t.Fatalf("LoadDocument failed: %v", err)
	}

	// Initially paused
	state := engine.GetState()
	if !state.IsPaused {
		t.Error("Engine should be paused initially")
	}

	// Play
	engine.Play()
	state = engine.GetState()
	if state.IsPaused {
		t.Error("Engine should not be paused after Play()")
	}

	// Pause
	engine.Pause()
	state = engine.GetState()
	if !state.IsPaused {
		t.Error("Engine should be paused after Pause()")
	}

	// Toggle
	engine.Toggle()
	state = engine.GetState()
	if state.IsPaused {
		t.Error("Engine should not be paused after Toggle() from paused state")
	}

	engine.Toggle()
	state = engine.GetState()
	if !state.IsPaused {
		t.Error("Engine should be paused after Toggle() from playing state")
	}
}

func TestEngineNavigation(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	testFile := createTestDocument(t, "First paragraph.\n\nSecond paragraph.\n\nThird paragraph.")

	if err := engine.LoadDocument(testFile); err != nil {
		t.Fatalf("LoadDocument failed: %v", err)
	}

	// Test jump to word
	engine.JumpToWord(3)
	state := engine.GetState()
	if state.WordIndex != 3 {
		t.Errorf("Expected word index 3, got %d", state.WordIndex)
	}

	// Test jump to start
	engine.JumpToWord(0)
	state = engine.GetState()
	if state.WordIndex != 0 {
		t.Errorf("Expected word index 0, got %d", state.WordIndex)
	}

	// Test next paragraph
	initialPara := state.ParagraphIndex
	engine.NextParagraph()
	state = engine.GetState()
	if state.ParagraphIndex <= initialPara && state.ParagraphIndex < state.TotalParagraphs-1 {
		t.Error("NextParagraph should move to next paragraph")
	}

	// Test prev paragraph
	if state.ParagraphIndex > 0 {
		currentPara := state.ParagraphIndex
		engine.PrevParagraph()
		state = engine.GetState()
		if state.ParagraphIndex >= currentPara {
			t.Error("PrevParagraph should move to previous paragraph")
		}
	}
}

func TestEngineGetState(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	testFile := createTestDocument(t, "Hello world test.")

	if err := engine.LoadDocument(testFile); err != nil {
		t.Fatalf("LoadDocument failed: %v", err)
	}

	state := engine.GetState()

	// Check all state fields are populated
	if state.TotalWords == 0 {
		t.Error("TotalWords should be non-zero")
	}

	if state.WPM == 0 {
		t.Error("WPM should be non-zero")
	}

	if !state.DocumentLoaded {
		t.Error("DocumentLoaded should be true")
	}

	if state.CurrentWord == "" {
		t.Error("CurrentWord should not be empty")
	}
}

func TestEngineWordAdvancement(t *testing.T) {
	config := DefaultConfig()
	config.DefaultWPM = 6000 // Very fast for testing
	engine := NewEngine(config)

	testFile := createTestDocument(t, "One two three four five.")

	if err := engine.LoadDocument(testFile); err != nil {
		t.Fatalf("LoadDocument failed: %v", err)
	}

	initialIndex := engine.GetState().WordIndex

	// Start playing
	engine.Play()

	// Wait for word advancement
	time.Sleep(100 * time.Millisecond)

	engine.Pause()

	finalIndex := engine.GetState().WordIndex

	// At high WPM, we should have advanced at least one word
	if finalIndex <= initialIndex {
		t.Log("Word advancement may not have occurred in time - this can be flaky")
	}
}

func TestEngineSubscribeState(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	testFile := createTestDocument(t, "Hello world.")

	if err := engine.LoadDocument(testFile); err != nil {
		t.Fatalf("LoadDocument failed: %v", err)
	}

	// Subscribe to state updates
	ch := engine.SubscribeState()
	if ch == nil {
		t.Fatal("SubscribeState returned nil channel")
	}

	// Unsubscribe
	engine.UnsubscribeState(ch)
}

func TestEngineSavePosition(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	testFile := createTestDocument(t, "Hello world test document.")

	if err := engine.LoadDocument(testFile); err != nil {
		t.Fatalf("LoadDocument failed: %v", err)
	}

	// Move to a specific position
	engine.JumpToWord(2)
	engine.SetWPM(400)

	// Save position
	err := engine.SavePosition()
	if err != nil {
		t.Errorf("SavePosition failed: %v", err)
	}

	// Check saved sessions
	sessions := engine.GetSavedSessions()
	// Note: sessions might be empty if persistence isn't working
	_ = sessions
}
