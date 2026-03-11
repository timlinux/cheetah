// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

// Package sessions handles persistence of reading positions and user sessions.
package sessions

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Session represents a saved reading session
type Session struct {
	// DocumentHash is the unique identifier for the document (SHA-256)
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

// SessionStore holds all saved sessions
type SessionStore struct {
	Version  int                `json:"version"`
	Sessions map[string]Session `json:"sessions"` // Keyed by document hash
}

var (
	mu         sync.RWMutex
	storePath  string
	storeCache *SessionStore
)

// ErrSessionNotFound indicates a session was not found
var ErrSessionNotFound = errors.New("session not found")

func init() {
	// Determine config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	storePath = filepath.Join(configDir, "cheetah", "sessions.json")
}

// GetStorePath returns the path to the sessions file
func GetStorePath() string {
	return storePath
}

// SetStorePath sets a custom path for the sessions file (useful for testing)
func SetStorePath(path string) {
	mu.Lock()
	defer mu.Unlock()
	storePath = path
	storeCache = nil
}

// loadStore loads the session store from disk
func loadStore() (*SessionStore, error) {
	mu.RLock()
	if storeCache != nil {
		store := *storeCache
		mu.RUnlock()
		return &store, nil
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	// Check again after acquiring write lock
	if storeCache != nil {
		store := *storeCache
		return &store, nil
	}

	data, err := os.ReadFile(storePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty store
			store := &SessionStore{
				Version:  1,
				Sessions: make(map[string]Session),
			}
			storeCache = store
			return store, nil
		}
		return nil, err
	}

	var store SessionStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}

	if store.Sessions == nil {
		store.Sessions = make(map[string]Session)
	}

	storeCache = &store
	return &store, nil
}

// saveStore saves the session store to disk
func saveStore(store *SessionStore) error {
	mu.Lock()
	defer mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(storePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(storePath, data, 0644); err != nil {
		return err
	}

	storeCache = store
	return nil
}

// Save saves or updates a session
func Save(session Session) error {
	store, err := loadStore()
	if err != nil {
		return err
	}

	session.LastAccessed = time.Now()
	store.Sessions[session.DocumentHash] = session

	return saveStore(store)
}

// Load retrieves a session by document hash
func Load(hash string) (*Session, error) {
	store, err := loadStore()
	if err != nil {
		return nil, err
	}

	session, ok := store.Sessions[hash]
	if !ok {
		return nil, ErrSessionNotFound
	}

	return &session, nil
}

// LoadAll returns all saved sessions, sorted by last accessed (most recent first)
func LoadAll() []Session {
	store, err := loadStore()
	if err != nil {
		return nil
	}

	sessions := make([]Session, 0, len(store.Sessions))
	for _, s := range store.Sessions {
		sessions = append(sessions, s)
	}

	// Sort by last accessed (most recent first)
	for i := 0; i < len(sessions); i++ {
		for j := i + 1; j < len(sessions); j++ {
			if sessions[j].LastAccessed.After(sessions[i].LastAccessed) {
				sessions[i], sessions[j] = sessions[j], sessions[i]
			}
		}
	}

	return sessions
}

// Delete removes a session by document hash
func Delete(hash string) error {
	store, err := loadStore()
	if err != nil {
		return err
	}

	delete(store.Sessions, hash)
	return saveStore(store)
}

// HasSession checks if a session exists for a document hash
func HasSession(hash string) bool {
	store, err := loadStore()
	if err != nil {
		return false
	}

	_, ok := store.Sessions[hash]
	return ok
}

// Clear removes all saved sessions
func Clear() error {
	store := &SessionStore{
		Version:  1,
		Sessions: make(map[string]Session),
	}
	return saveStore(store)
}
