// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package documents

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

// PDFParser handles PDF document parsing
type PDFParser struct{}

// Parse reads a PDF file and extracts its text content
func (p *PDFParser) Parse(path string) (*Document, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var textBuilder strings.Builder
	totalPages := r.NumPage()

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}

		textBuilder.WriteString(text)
		textBuilder.WriteString("\n\n") // Paragraph break between pages
	}

	text := textBuilder.String()
	if strings.TrimSpace(text) == "" {
		return nil, ErrEmptyDocument
	}

	// Process the extracted text
	processor := DefaultProcessor()
	doc := processor.Process(text, getTitleFromFilename(filepath.Base(path)), path)
	doc.Path = path

	if doc.TotalWords == 0 {
		return nil, ErrEmptyDocument
	}

	return doc, nil
}

// ParseBytes parses PDF content from bytes
func (p *PDFParser) ParseBytes(data []byte, filename string) (*Document, error) {
	// Create a temporary file since the PDF library requires a file
	tmpFile, err := os.CreateTemp("", "cheetah-*.pdf")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(data); err != nil {
		return nil, err
	}
	tmpFile.Close()

	doc, err := p.Parse(tmpFile.Name())
	if err != nil {
		return nil, err
	}

	// Update the title since we used a temp file
	doc.Title = getTitleFromFilename(filename)
	return doc, nil
}

// SupportedExtensions returns the extensions this parser handles
func (p *PDFParser) SupportedExtensions() []string {
	return []string{".pdf"}
}
