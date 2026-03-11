// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package documents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()

	expectedFormats := []string{".txt", ".md", ".markdown", ".pdf", ".docx", ".epub", ".odt"}

	for _, expected := range expectedFormats {
		found := false
		for _, format := range formats {
			if format == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected format %s not found in SupportedFormats()", expected)
		}
	}
}

func TestGetParser(t *testing.T) {
	tests := []struct {
		filename    string
		expectError bool
	}{
		{"test.txt", false},
		{"test.md", false},
		{"test.markdown", false},
		{"test.pdf", false},
		{"test.docx", false},
		{"test.epub", false},
		{"test.odt", false},
		{"test.doc", true}, // Unsupported
		{"test.rtf", true}, // Unsupported
		{"test", true},     // No extension
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			parser, err := GetParser(tt.filename)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tt.filename)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.filename, err)
				}
				if parser == nil {
					t.Errorf("Expected parser for %s, got nil", tt.filename)
				}
			}
		})
	}
}

func TestTextParser(t *testing.T) {
	// Create a temporary test file
	content := `# Test Document

This is a test paragraph with some words.

## Section One

Here is another paragraph.
And another line.

1. First item
2. Second item
3. Third item

The end.`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := &TextParser{}
	doc, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	// Check document properties
	if doc.Title == "" {
		t.Error("Expected non-empty title")
	}

	if doc.TotalWords == 0 {
		t.Error("Expected non-zero word count")
	}

	if len(doc.Paragraphs) == 0 {
		t.Error("Expected at least one paragraph")
	}

	// Verify words are extracted
	allWords := doc.GetAllWords()
	if len(allWords) == 0 {
		t.Error("Expected words to be extracted")
	}

	// Check that "test" appears in the document
	found := false
	for _, word := range allWords {
		if strings.ToLower(word) == "test" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find word 'test' in document")
	}
}

func TestTextParserEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := &TextParser{}
	doc, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Unexpected error parsing empty file: %v", err)
	}

	if doc.TotalWords != 0 {
		t.Errorf("Expected 0 words, got %d", doc.TotalWords)
	}
}

func TestTextParserNonExistentFile(t *testing.T) {
	parser := &TextParser{}
	_, err := parser.Parse("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestDocumentGetWord(t *testing.T) {
	doc := &Document{
		Paragraphs: []Paragraph{
			{Words: []string{"Hello", "world"}},
			{Words: []string{"Foo", "bar", "baz"}},
		},
		TotalWords: 5,
	}

	tests := []struct {
		index    int
		expected string
	}{
		{0, "Hello"},
		{1, "world"},
		{2, "Foo"},
		{3, "bar"},
		{4, "baz"},
		{5, ""},  // Out of bounds
		{-1, ""}, // Negative
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			word := doc.GetWord(tt.index)
			if word != tt.expected {
				t.Errorf("GetWord(%d) = %q, expected %q", tt.index, word, tt.expected)
			}
		})
	}
}

func TestDocumentGetParagraphForWord(t *testing.T) {
	doc := &Document{
		Paragraphs: []Paragraph{
			{Words: []string{"Hello", "world"}},        // words 0-1
			{Words: []string{"Foo", "bar", "baz"}},     // words 2-4
			{Words: []string{"One", "two", "three"}},   // words 5-7
		},
		TotalWords: 8,
	}

	tests := []struct {
		wordIndex int
		expected  int
	}{
		{0, 0},
		{1, 0},
		{2, 1},
		{3, 1},
		{4, 1},
		{5, 2},
		{7, 2},
		{8, -1},  // Out of bounds
		{-1, -1}, // Negative
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			paraIndex := doc.GetParagraphForWord(tt.wordIndex)
			if paraIndex != tt.expected {
				t.Errorf("GetParagraphForWord(%d) = %d, expected %d", tt.wordIndex, paraIndex, tt.expected)
			}
		})
	}
}

func TestDocumentGetParagraphStartWord(t *testing.T) {
	doc := &Document{
		Paragraphs: []Paragraph{
			{Words: []string{"Hello", "world"}},      // words 0-1
			{Words: []string{"Foo", "bar", "baz"}},   // words 2-4
			{Words: []string{"One", "two", "three"}}, // words 5-7
		},
		TotalWords: 8,
	}

	tests := []struct {
		paraIndex int
		expected  int
	}{
		{0, 0},
		{1, 2},
		{2, 5},
		{3, -1},  // Out of bounds
		{-1, -1}, // Negative
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			wordIndex := doc.GetParagraphStartWord(tt.paraIndex)
			if wordIndex != tt.expected {
				t.Errorf("GetParagraphStartWord(%d) = %d, expected %d", tt.paraIndex, wordIndex, tt.expected)
			}
		})
	}
}

func TestDocumentHash(t *testing.T) {
	doc1 := &Document{
		Paragraphs: []Paragraph{
			{Words: []string{"Hello", "world"}},
		},
		TotalWords: 2,
	}

	doc2 := &Document{
		Paragraphs: []Paragraph{
			{Words: []string{"Hello", "world"}},
		},
		TotalWords: 2,
	}

	doc3 := &Document{
		Paragraphs: []Paragraph{
			{Words: []string{"Different", "content"}},
		},
		TotalWords: 2,
	}

	hash1 := doc1.Hash()
	hash2 := doc2.Hash()
	hash3 := doc3.Hash()

	// Same content should have same hash
	if hash1 != hash2 {
		t.Errorf("Same content should have same hash: %s != %s", hash1, hash2)
	}

	// Different content should have different hash
	if hash1 == hash3 {
		t.Error("Different content should have different hash")
	}

	// Hash should be non-empty
	if hash1 == "" {
		t.Error("Hash should not be empty")
	}
}
