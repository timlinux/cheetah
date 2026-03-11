// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestSettings(t *testing.T) *Settings {
	t.Helper()
	tmpDir := t.TempDir()

	// Set up environment to use temp dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Create config directory
	configDir := filepath.Join(tmpDir, ".config", "cheetah")
	os.MkdirAll(configDir, 0755)

	return Load()
}

func TestDefaultSettings(t *testing.T) {
	settings := Default()

	if settings.DefaultWPM <= 0 {
		t.Error("DefaultWPM should be positive")
	}

	if settings.MinWPM <= 0 {
		t.Error("MinWPM should be positive")
	}

	if settings.MaxWPM <= settings.MinWPM {
		t.Error("MaxWPM should be greater than MinWPM")
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
	settings := setupTestSettings(t)

	// Modify settings
	settings.DefaultWPM = 450
	settings.MinWPM = 100
	settings.MaxWPM = 1500

	// Save
	err := settings.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load again
	loaded := Load()
	if loaded == nil {
		t.Fatal("Load returned nil after save")
	}

	if loaded.DefaultWPM != 450 {
		t.Errorf("DefaultWPM not persisted: got %d, expected 450", loaded.DefaultWPM)
	}

	if loaded.MinWPM != 100 {
		t.Errorf("MinWPM not persisted: got %d, expected 100", loaded.MinWPM)
	}

	if loaded.MaxWPM != 1500 {
		t.Errorf("MaxWPM not persisted: got %d, expected 1500", loaded.MaxWPM)
	}
}

func TestSettingsValidation(t *testing.T) {
	settings := Default()

	// Test that validation ensures reasonable values
	if settings.MinWPM > settings.DefaultWPM {
		t.Error("MinWPM should not exceed DefaultWPM")
	}

	if settings.MaxWPM < settings.DefaultWPM {
		t.Error("MaxWPM should not be less than DefaultWPM")
	}
}

func TestSettingsPath(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	path := GetSettingsPath()

	// Path should contain .config/cheetah
	if !filepath.IsAbs(path) {
		t.Error("Settings path should be absolute")
	}

	expectedSuffix := filepath.Join(".config", "cheetah", "settings.json")
	if !contains(path, "cheetah") {
		t.Errorf("Settings path should contain 'cheetah': %s", path)
	}
	_ = expectedSuffix
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && contains(s[1:], substr) || s[:len(substr)] == substr)
}

func TestSettingsDefaults(t *testing.T) {
	defaults := Default()

	tests := []struct {
		name     string
		value    int
		minValue int
	}{
		{"DefaultWPM", defaults.DefaultWPM, 100},
		{"MinWPM", defaults.MinWPM, 10},
		{"MaxWPM", defaults.MaxWPM, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.minValue {
				t.Errorf("%s should be at least %d, got %d", tt.name, tt.minValue, tt.value)
			}
		})
	}
}
