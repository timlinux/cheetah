// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

// Package documents provides document parsing functionality for various file formats.
// It extracts text content from PDF, DOCX, EPUB, ODT, TXT, and Markdown files.
package documents

import (
	"errors"
	"path/filepath"
	"strings"
)

// Document represents a parsed document with its content
type Document struct {
	// Title is the document title (from metadata or filename)
	Title string

	// Paragraphs contains all paragraphs of text
	Paragraphs []Paragraph

	// TotalWords is the total word count
	TotalWords int

	// Hash is a SHA-256 hash of the document content for identification
	Hash string

	// Path is the original file path
	Path string
}

// Paragraph represents a paragraph of text broken into words
type Paragraph struct {
	// Words contains the individual words in this paragraph
	Words []string

	// Index is the paragraph index (0-based)
	Index int
}

// Parser is the interface for document parsers
type Parser interface {
	// Parse reads a document and extracts its text content
	Parse(path string) (*Document, error)

	// ParseBytes parses document content from bytes
	ParseBytes(data []byte, filename string) (*Document, error)

	// SupportedExtensions returns the file extensions this parser supports
	SupportedExtensions() []string
}

// Errors
var (
	ErrUnsupportedFormat = errors.New("unsupported document format")
	ErrEmptyDocument     = errors.New("document contains no readable text")
	ErrParseError        = errors.New("failed to parse document")
)

// GetParser returns the appropriate parser for a file based on its extension
func GetParser(path string) (Parser, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".txt", ".md", ".markdown":
		return &TextParser{}, nil
	case ".pdf":
		return &PDFParser{}, nil
	case ".docx":
		return &DOCXParser{}, nil
	case ".epub":
		return &EPUBParser{}, nil
	case ".odt":
		return &ODTParser{}, nil
	default:
		return nil, ErrUnsupportedFormat
	}
}

// ParseFile parses a document file and returns its content
func ParseFile(path string) (*Document, error) {
	parser, err := GetParser(path)
	if err != nil {
		return nil, err
	}

	return parser.Parse(path)
}

// SupportedFormats returns a list of supported file formats
func SupportedFormats() []string {
	return []string{
		".txt",
		".md",
		".markdown",
		".pdf",
		".docx",
		".epub",
		".odt",
	}
}

// IsSupportedFormat checks if a file extension is supported
func IsSupportedFormat(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, supported := range SupportedFormats() {
		if ext == supported {
			return true
		}
	}
	return false
}
