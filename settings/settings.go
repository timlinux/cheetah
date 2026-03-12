// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

// Package settings handles user preferences and configuration.
package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Settings holds user preferences
type Settings struct {
	// DefaultWPM is the default words per minute
	DefaultWPM int `json:"default_wpm"`

	// ShowProgress shows progress bar
	ShowProgress bool `json:"show_progress"`

	// ShowPreviousWord shows the previous word above the current word
	ShowPreviousWord bool `json:"show_previous_word"`

	// ShowNextWords shows upcoming words below the current word
	ShowNextWords bool `json:"show_next_words"`

	// NextWordsCount is how many upcoming words to show
	NextWordsCount int `json:"next_words_count"`

	// AutoSave automatically saves reading position
	AutoSave bool `json:"auto_save"`

	// AutoSaveInterval is how often to auto-save (in words read)
	AutoSaveInterval int `json:"auto_save_interval"`

	// LastDirectory is the last directory visited in the file picker
	LastDirectory string `json:"last_directory"`
}

var (
	mu           sync.RWMutex
	settingsPath string
	cache        *Settings
)

func init() {
	// Determine config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	settingsPath = filepath.Join(configDir, "cheetah", "settings.json")
}

// DefaultSettings returns the default settings
func DefaultSettings() *Settings {
	return &Settings{
		DefaultWPM:       300,
		ShowProgress:     true,
		ShowPreviousWord: true,
		ShowNextWords:    true,
		NextWordsCount:   3,
		AutoSave:         true,
		AutoSaveInterval: 50, // Every 50 words
	}
}

// Load loads settings from disk, returning defaults if file doesn't exist
func Load() (*Settings, error) {
	mu.RLock()
	if cache != nil {
		s := *cache
		mu.RUnlock()
		return &s, nil
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	// Check again after acquiring write lock
	if cache != nil {
		s := *cache
		return &s, nil
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return defaults
			settings := DefaultSettings()
			cache = settings
			return settings, nil
		}
		return nil, err
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	// Apply defaults for any missing fields
	defaults := DefaultSettings()
	if settings.DefaultWPM == 0 {
		settings.DefaultWPM = defaults.DefaultWPM
	}
	if settings.NextWordsCount == 0 {
		settings.NextWordsCount = defaults.NextWordsCount
	}
	if settings.AutoSaveInterval == 0 {
		settings.AutoSaveInterval = defaults.AutoSaveInterval
	}

	cache = &settings
	return &settings, nil
}

// Save saves settings to disk
func (s *Settings) Save() error {
	mu.Lock()
	defer mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return err
	}

	cache = s
	return nil
}

// GetSettingsPath returns the path to the settings file
func GetSettingsPath() string {
	return settingsPath
}

// SetSettingsPath sets a custom path for the settings file (useful for testing)
func SetSettingsPath(path string) {
	mu.Lock()
	defer mu.Unlock()
	settingsPath = path
	cache = nil
}

// Reset resets settings to defaults
func Reset() error {
	settings := DefaultSettings()
	return settings.Save()
}
