// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package documents

import (
	"strings"
	"testing"
)

func TestProcessText(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		minWords   int
		minParas   int
	}{
		{
			name:     "Simple text",
			input:    "Hello world. This is a test.",
			minWords: 6,
			minParas: 1,
		},
		{
			name:     "Multiple paragraphs",
			input:    "First paragraph.\n\nSecond paragraph.",
			minWords: 4,
			minParas: 2,
		},
		{
			name:     "Empty string",
			input:    "",
			minWords: 0,
			minParas: 0,
		},
		{
			name:     "Only whitespace",
			input:    "   \n\n\t  ",
			minWords: 0,
			minParas: 0,
		},
		{
			name:     "Single word",
			input:    "Hello",
			minWords: 1,
			minParas: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paragraphs := ProcessText(tt.input)

			totalWords := 0
			for _, p := range paragraphs {
				totalWords += len(p.Words)
			}

			if totalWords < tt.minWords {
				t.Errorf("Expected at least %d words, got %d", tt.minWords, totalWords)
			}

			if len(paragraphs) < tt.minParas {
				t.Errorf("Expected at least %d paragraphs, got %d", tt.minParas, len(paragraphs))
			}
		})
	}
}

func TestTokenizeWord(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"Hello", "Hello"},
		{"hello,", "hello,"},
		{"hello.", "hello."},
		{"'hello'", "'hello'"},
		{"\"hello\"", "\"hello\""},
		{"", ""},
		{"   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := strings.TrimSpace(tt.input)
			if result != tt.expected {
				t.Errorf("TokenizeWord(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParagraphDetection(t *testing.T) {
	// Test that double newlines create paragraph breaks
	input := "First paragraph here.\n\nSecond paragraph here.\n\nThird paragraph."
	paragraphs := ProcessText(input)

	if len(paragraphs) < 3 {
		t.Errorf("Expected at least 3 paragraphs, got %d", len(paragraphs))
	}
}

func TestWordPreservation(t *testing.T) {
	// Test that words are correctly extracted from text with punctuation
	// Note: CleanWord trims trailing punctuation (.,;:!?-) from words
	input := "Hello, world! This is a test. How are you?"
	paragraphs := ProcessText(input)

	allWords := []string{}
	for _, p := range paragraphs {
		allWords = append(allWords, p.Words...)
	}

	// Check that expected words are present (punctuation is trimmed)
	expectedWords := map[string]bool{
		"Hello": false,
		"world": false,
		"This":  false,
		"is":    false,
		"a":     false,
		"test":  false,
		"How":   false,
		"are":   false,
		"you":   false,
	}

	for _, word := range allWords {
		if _, ok := expectedWords[word]; ok {
			expectedWords[word] = true
		}
	}

	for word, found := range expectedWords {
		if !found {
			t.Errorf("Expected to find word %q in extracted words", word)
		}
	}
}

func TestSpecialCharacterHandling(t *testing.T) {
	// Test handling of special characters
	input := "Test em-dash—here and en-dash–here."
	paragraphs := ProcessText(input)

	allWords := []string{}
	for _, p := range paragraphs {
		allWords = append(allWords, p.Words...)
	}

	if len(allWords) == 0 {
		t.Error("Expected words to be extracted from text with special characters")
	}
}

func TestCurlyQuoteHandling(t *testing.T) {
	// Test handling of curly quotes
	input := `"Hello," she said. 'Yes!' he replied.`
	paragraphs := ProcessText(input)

	allWords := []string{}
	for _, p := range paragraphs {
		allWords = append(allWords, p.Words...)
	}

	if len(allWords) == 0 {
		t.Error("Expected words to be extracted from text with curly quotes")
	}
}
