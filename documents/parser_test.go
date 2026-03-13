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
	allWords := GetAllWords(doc)
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
	_, err := parser.Parse(testFile)
	if err != ErrEmptyDocument {
		t.Errorf("Expected ErrEmptyDocument for empty file, got %v", err)
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
			word, _, ok := GetWordAt(doc, tt.index)
			if !ok && tt.expected != "" {
				t.Errorf("GetWordAt(%d) returned not ok, expected %q", tt.index, tt.expected)
			} else if ok && word != tt.expected {
				t.Errorf("GetWordAt(%d) = %q, expected %q", tt.index, word, tt.expected)
			}
		})
	}
}

func TestDocumentGetParagraphForWord(t *testing.T) {
	doc := &Document{
		Paragraphs: []Paragraph{
			{Words: []string{"Hello", "world"}, Index: 0},      // words 0-1
			{Words: []string{"Foo", "bar", "baz"}, Index: 1},   // words 2-4
			{Words: []string{"One", "two", "three"}, Index: 2}, // words 5-7
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
		{8, 2},  // Out of bounds returns last paragraph
		{-1, 0}, // Negative returns first paragraph
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			paraIndex := GetParagraphForWordIndex(doc, tt.wordIndex)
			if paraIndex != tt.expected {
				t.Errorf("GetParagraphForWordIndex(%d) = %d, expected %d", tt.wordIndex, paraIndex, tt.expected)
			}
		})
	}
}

func TestDocumentGetParagraphStartWord(t *testing.T) {
	doc := &Document{
		Paragraphs: []Paragraph{
			{Words: []string{"Hello", "world"}, Index: 0},      // words 0-1
			{Words: []string{"Foo", "bar", "baz"}, Index: 1},   // words 2-4
			{Words: []string{"One", "two", "three"}, Index: 2}, // words 5-7
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
		{3, 8},  // Out of bounds returns total words
		{-1, 0}, // Negative returns 0
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			wordIndex := GetParagraphStartIndex(doc, tt.paraIndex)
			if wordIndex != tt.expected {
				t.Errorf("GetParagraphStartIndex(%d) = %d, expected %d", tt.paraIndex, wordIndex, tt.expected)
			}
		})
	}
}

func TestDocumentHash(t *testing.T) {
	processor := DefaultProcessor()

	doc1 := processor.Process("Hello world", "test1", "/test1.txt")
	doc2 := processor.Process("Hello world", "test2", "/test2.txt")
	doc3 := processor.Process("Different content", "test3", "/test3.txt")

	// Same content should have same hash
	if doc1.Hash != doc2.Hash {
		t.Errorf("Same content should have same hash: %s != %s", doc1.Hash, doc2.Hash)
	}

	// Different content should have different hash
	if doc1.Hash == doc3.Hash {
		t.Error("Different content should have different hash")
	}

	// Hash should be non-empty
	if doc1.Hash == "" {
		t.Error("Hash should not be empty")
	}
}
