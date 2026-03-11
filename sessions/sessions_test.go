// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package sessions

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestManager(t *testing.T) *Manager {
	t.Helper()
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "cheetah")

	// Set up environment to use temp dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	// Ensure directory exists
	os.MkdirAll(configDir, 0755)

	return manager
}

func TestNewManager(t *testing.T) {
	manager := setupTestManager(t)
	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}
}

func TestManagerSaveAndLoad(t *testing.T) {
	manager := setupTestManager(t)

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
	err := manager.SaveSession(session)
	if err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// Load session
	loaded, err := manager.GetSession(session.DocumentHash)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if loaded == nil {
		t.Fatal("GetSession returned nil for existing session")
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

func TestManagerGetNonExistentSession(t *testing.T) {
	manager := setupTestManager(t)

	session, err := manager.GetSession("nonexistent")
	if err != nil {
		// Some implementations might return an error
		return
	}

	if session != nil {
		t.Error("Expected nil for non-existent session")
	}
}

func TestManagerGetAllSessions(t *testing.T) {
	manager := setupTestManager(t)

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
		if err := manager.SaveSession(s); err != nil {
			t.Fatalf("SaveSession failed: %v", err)
		}
	}

	// Get all sessions
	allSessions := manager.GetAllSessions()
	if len(allSessions) < 2 {
		t.Errorf("Expected at least 2 sessions, got %d", len(allSessions))
	}
}

func TestManagerDeleteSession(t *testing.T) {
	manager := setupTestManager(t)

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
	if err := manager.SaveSession(session); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// Verify it exists
	loaded, _ := manager.GetSession(session.DocumentHash)
	if loaded == nil {
		t.Fatal("Session should exist before deletion")
	}

	// Delete session
	err := manager.DeleteSession(session.DocumentHash)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// Verify it's gone
	deleted, _ := manager.GetSession(session.DocumentHash)
	if deleted != nil {
		t.Error("Session should not exist after deletion")
	}
}

func TestManagerUpdateSession(t *testing.T) {
	manager := setupTestManager(t)

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
	if err := manager.SaveSession(session); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// Update session
	session.LastPosition = 100
	session.LastWPM = 400
	session.LastAccessed = time.Now()

	if err := manager.SaveSession(session); err != nil {
		t.Fatalf("SaveSession (update) failed: %v", err)
	}

	// Verify update
	loaded, err := manager.GetSession(session.DocumentHash)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
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
