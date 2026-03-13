// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package settings

import (
	"path/filepath"
	"testing"
)

func setupTestSettings(t *testing.T) *Settings {
	t.Helper()
	tmpDir := t.TempDir()

	// Set up custom settings path for testing
	testSettingsPath := filepath.Join(tmpDir, "settings.json")
	SetSettingsPath(testSettingsPath)

	settings, err := Load()
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	return settings
}

func TestDefaultSettings(t *testing.T) {
	settings := DefaultSettings()

	if settings.DefaultWPM <= 0 {
		t.Error("DefaultWPM should be positive")
	}

	if settings.NextWordsCount <= 0 {
		t.Error("NextWordsCount should be positive")
	}

	if settings.AutoSaveInterval <= 0 {
		t.Error("AutoSaveInterval should be positive")
	}
}

func TestLoadSettings(t *testing.T) {
	settings := setupTestSettings(t)

	if settings == nil {
		t.Fatal("Load returned nil")
	}

	// Should have default values
	if settings.DefaultWPM <= 0 {
		t.Error("Loaded settings should have positive DefaultWPM")
	}
}

func TestSaveAndLoadSettings(t *testing.T) {
	tmpDir := t.TempDir()
	testSettingsPath := filepath.Join(tmpDir, "settings.json")
	SetSettingsPath(testSettingsPath)

	settings, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Modify settings
	settings.DefaultWPM = 450
	settings.ShowProgress = false
	settings.NextWordsCount = 5

	// Save
	err = settings.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Clear cache and load again
	SetSettingsPath(testSettingsPath)
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed after save: %v", err)
	}

	if loaded.DefaultWPM != 450 {
		t.Errorf("DefaultWPM not persisted: got %d, expected 450", loaded.DefaultWPM)
	}

	if loaded.ShowProgress != false {
		t.Error("ShowProgress not persisted")
	}

	if loaded.NextWordsCount != 5 {
		t.Errorf("NextWordsCount not persisted: got %d, expected 5", loaded.NextWordsCount)
	}
}

func TestSettingsValidation(t *testing.T) {
	settings := DefaultSettings()

	// Test that default values are reasonable
	if settings.DefaultWPM < 100 {
		t.Error("DefaultWPM should be at least 100")
	}

	if settings.DefaultWPM > 1000 {
		t.Error("DefaultWPM should not exceed 1000")
	}
}

func TestSettingsPath(t *testing.T) {
	path := GetSettingsPath()

	// Path should be absolute
	if !filepath.IsAbs(path) {
		t.Error("Settings path should be absolute")
	}

	// Path should end with .json
	if filepath.Ext(path) != ".json" {
		t.Errorf("Settings path should end with .json: %s", path)
	}
}

func TestSettingsDefaults(t *testing.T) {
	defaults := DefaultSettings()

	tests := []struct {
		name     string
		value    int
		minValue int
	}{
		{"DefaultWPM", defaults.DefaultWPM, 100},
		{"NextWordsCount", defaults.NextWordsCount, 1},
		{"AutoSaveInterval", defaults.AutoSaveInterval, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.minValue {
				t.Errorf("%s should be at least %d, got %d", tt.name, tt.minValue, tt.value)
			}
		})
	}
}
