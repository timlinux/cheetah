// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package documents

import (
	"os"
	"path/filepath"
	"strings"
)

// TextParser handles plain text and Markdown files
type TextParser struct{}

// Parse reads a text file and extracts its content
func (p *TextParser) Parse(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return p.ParseBytes(data, filepath.Base(path))
}

// ParseBytes parses text content from bytes
func (p *TextParser) ParseBytes(data []byte, filename string) (*Document, error) {
	text := string(data)

	// For Markdown, strip common formatting
	if strings.HasSuffix(strings.ToLower(filename), ".md") ||
		strings.HasSuffix(strings.ToLower(filename), ".markdown") {
		text = stripMarkdown(text)
	}

	// Use default processor
	processor := DefaultProcessor()
	doc := processor.Process(text, getTitleFromFilename(filename), filename)

	if doc.TotalWords == 0 {
		return nil, ErrEmptyDocument
	}

	return doc, nil
}

// SupportedExtensions returns the extensions this parser handles
func (p *TextParser) SupportedExtensions() []string {
	return []string{".txt", ".md", ".markdown"}
}

// stripMarkdown removes common Markdown formatting
func stripMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		// Skip code fence markers
		if strings.HasPrefix(line, "```") {
			continue
		}

		// Remove heading markers
		line = strings.TrimLeft(line, "#")

		// Remove emphasis markers (* and _)
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "__", "")
		line = strings.ReplaceAll(line, "*", "")
		line = strings.ReplaceAll(line, "_", "")

		// Remove inline code backticks
		line = strings.ReplaceAll(line, "`", "")

		// Remove link syntax [text](url) -> text
		for strings.Contains(line, "](") {
			start := strings.Index(line, "[")
			end := strings.Index(line, "](")
			urlEnd := strings.Index(line[end:], ")")
			if start >= 0 && end > start && urlEnd > 0 {
				text := line[start+1 : end]
				line = line[:start] + text + line[end+urlEnd+1:]
			} else {
				break
			}
		}

		// Remove image syntax ![alt](url)
		for strings.Contains(line, "![") {
			start := strings.Index(line, "![")
			end := strings.Index(line[start:], ")")
			if start >= 0 && end > 0 {
				line = line[:start] + line[start+end+1:]
			} else {
				break
			}
		}

		// Remove blockquote markers
		line = strings.TrimPrefix(line, "> ")
		line = strings.TrimPrefix(line, ">")

		// Remove list markers
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")

		// Remove numbered list markers
		if len(line) > 2 {
			for i := 0; i < 10; i++ {
				prefix := strings.Repeat(" ", i) + "1. "
				line = strings.TrimPrefix(line, prefix)
			}
		}

		// Remove horizontal rules
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			continue
		}

		result = append(result, strings.TrimSpace(line))
	}

	return strings.Join(result, "\n")
}

// getTitleFromFilename extracts a title from a filename
func getTitleFromFilename(filename string) string {
	// Remove extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Replace common separators with spaces
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")

	// Title case
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}
