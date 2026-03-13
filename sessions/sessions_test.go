// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package sessions

import (
	"path/filepath"
	"testing"
	"time"
)

func setupTestSessions(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()

	// Set up custom store path for testing
	testStorePath := filepath.Join(tmpDir, "sessions.json")
	SetStorePath(testStorePath)
}

func TestSaveAndLoad(t *testing.T) {
	setupTestSessions(t)

	session := Session{
		DocumentHash:  "testhash123",
		DocumentPath:  "/path/to/test.txt",
		DocumentTitle: "Test Document",
		LastPosition:  100,
		TotalWords:    500,
		LastWPM:       350,
		LastAccessed:  time.Now(),
	}

	// Save session
	err := Save(session)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load session
	loaded, err := Load(session.DocumentHash)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded == nil {
		t.Fatal("Load returned nil for existing session")
	}

	if loaded.DocumentHash != session.DocumentHash {
		t.Errorf("DocumentHash mismatch: got %s, expected %s", loaded.DocumentHash, session.DocumentHash)
	}

	if loaded.LastPosition != session.LastPosition {
		t.Errorf("LastPosition mismatch: got %d, expected %d", loaded.LastPosition, session.LastPosition)
	}

	if loaded.LastWPM != session.LastWPM {
		t.Errorf("LastWPM mismatch: got %d, expected %d", loaded.LastWPM, session.LastWPM)
	}
}

func TestLoadNonExistentSession(t *testing.T) {
	setupTestSessions(t)

	_, err := Load("nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestLoadAllSessions(t *testing.T) {
	setupTestSessions(t)

	// Save multiple sessions
	sessions := []Session{
		{
			DocumentHash:  "hash1",
			DocumentPath:  "/path/to/doc1.txt",
			DocumentTitle: "Document 1",
			LastPosition:  50,
			TotalWords:    200,
			LastWPM:       300,
			LastAccessed:  time.Now(),
		},
		{
			DocumentHash:  "hash2",
			DocumentPath:  "/path/to/doc2.txt",
			DocumentTitle: "Document 2",
			LastPosition:  100,
			TotalWords:    400,
			LastWPM:       400,
			LastAccessed:  time.Now(),
		},
	}

	for _, s := range sessions {
		if err := Save(s); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	// Get all sessions
	allSessions := LoadAll()
	if len(allSessions) < 2 {
		t.Errorf("Expected at least 2 sessions, got %d", len(allSessions))
	}
}

func TestDeleteSession(t *testing.T) {
	setupTestSessions(t)

	session := Session{
		DocumentHash:  "deleteme",
		DocumentPath:  "/path/to/delete.txt",
		DocumentTitle: "Delete Me",
		LastPosition:  0,
		TotalWords:    100,
		LastWPM:       300,
		LastAccessed:  time.Now(),
	}

	// Save session
	if err := Save(session); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify it exists
	if !HasSession(session.DocumentHash) {
		t.Fatal("Session should exist before deletion")
	}

	// Delete session
	err := Delete(session.DocumentHash)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	if HasSession(session.DocumentHash) {
		t.Error("Session should not exist after deletion")
	}
}

func TestUpdateSession(t *testing.T) {
	setupTestSessions(t)

	session := Session{
		DocumentHash:  "updateme",
		DocumentPath:  "/path/to/update.txt",
		DocumentTitle: "Update Me",
		LastPosition:  50,
		TotalWords:    200,
		LastWPM:       300,
		LastAccessed:  time.Now(),
	}

	// Save initial session
	if err := Save(session); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Update session
	session.LastPosition = 100
	session.LastWPM = 400
	session.LastAccessed = time.Now()

	if err := Save(session); err != nil {
		t.Fatalf("Save (update) failed: %v", err)
	}

	// Verify update
	loaded, err := Load(session.DocumentHash)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.LastPosition != 100 {
		t.Errorf("LastPosition not updated: got %d, expected 100", loaded.LastPosition)
	}

	if loaded.LastWPM != 400 {
		t.Errorf("LastWPM not updated: got %d, expected 400", loaded.LastWPM)
	}
}

func TestSessionProgress(t *testing.T) {
	session := Session{
		LastPosition: 50,
		TotalWords:   200,
	}

	progress := float64(session.LastPosition) / float64(session.TotalWords) * 100
	expected := 25.0

	if progress != expected {
		t.Errorf("Progress calculation wrong: got %.2f, expected %.2f", progress, expected)
	}
}

func TestClearSessions(t *testing.T) {
	setupTestSessions(t)

	// Save a session
	session := Session{
		DocumentHash:  "clearthis",
		DocumentPath:  "/path/to/clear.txt",
		DocumentTitle: "Clear This",
		LastPosition:  0,
		TotalWords:    100,
		LastWPM:       300,
		LastAccessed:  time.Now(),
	}

	if err := Save(session); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Clear all sessions
	if err := Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify it's gone
	if HasSession(session.DocumentHash) {
		t.Error("Session should not exist after clear")
	}
}

func TestHasSession(t *testing.T) {
	setupTestSessions(t)

	// Should not have session initially
	if HasSession("notexist") {
		t.Error("HasSession should return false for non-existent session")
	}

	// Save a session
	session := Session{
		DocumentHash:  "exists",
		DocumentPath:  "/path/to/exists.txt",
		DocumentTitle: "Exists",
		LastPosition:  0,
		TotalWords:    100,
		LastWPM:       300,
		LastAccessed:  time.Now(),
	}

	if err := Save(session); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Should have session now
	if !HasSession("exists") {
		t.Error("HasSession should return true for existing session")
	}
}
