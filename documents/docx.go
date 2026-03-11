// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package documents

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DOCXParser handles Microsoft Word DOCX document parsing using standard library
type DOCXParser struct{}

// Parse reads a DOCX file and extracts its text content
func (p *DOCXParser) Parse(path string) (*Document, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var textBuilder strings.Builder

	// Find and parse word/document.xml
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			text := p.extractDocumentText(f)
			textBuilder.WriteString(text)
			break
		}
	}

	text := textBuilder.String()
	if strings.TrimSpace(text) == "" {
		return nil, ErrEmptyDocument
	}

	// Process the extracted text
	processor := DefaultProcessor()
	result := processor.Process(text, getTitleFromFilename(filepath.Base(path)), path)
	result.Path = path

	if result.TotalWords == 0 {
		return nil, ErrEmptyDocument
	}

	return result, nil
}

// ParseBytes parses DOCX content from bytes
func (p *DOCXParser) ParseBytes(data []byte, filename string) (*Document, error) {
	tmpFile, err := createTempFile("cheetah-*.docx", data)
	if err != nil {
		return nil, err
	}
	defer removeTempFile(tmpFile)

	doc, err := p.Parse(tmpFile)
	if err != nil {
		return nil, err
	}

	doc.Title = getTitleFromFilename(filename)
	return doc, nil
}

// SupportedExtensions returns the extensions this parser handles
func (p *DOCXParser) SupportedExtensions() []string {
	return []string{".docx"}
}

// extractDocumentText extracts text from document.xml
func (p *DOCXParser) extractDocumentText(f *zip.File) string {
	rc, err := f.Open()
	if err != nil {
		return ""
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return ""
	}

	return p.parseWordML(data)
}

// parseWordML extracts text from WordprocessingML (DOCX XML format)
func (p *DOCXParser) parseWordML(data []byte) string {
	var result strings.Builder

	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var inText bool
	var inParagraph bool

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Track paragraph elements for line breaks
			if t.Name.Local == "p" {
				inParagraph = true
			}
			// Track text elements
			if t.Name.Local == "t" {
				inText = true
			}
			// Handle line breaks and tabs
			if t.Name.Local == "br" {
				result.WriteString("\n")
			}
			if t.Name.Local == "tab" {
				result.WriteString(" ")
			}

		case xml.EndElement:
			if t.Name.Local == "t" {
				inText = false
			}
			if t.Name.Local == "p" {
				if inParagraph {
					result.WriteString("\n\n")
				}
				inParagraph = false
			}

		case xml.CharData:
			if inText {
				result.WriteString(string(t))
			}
		}
	}

	return result.String()
}

// Helper functions for temp file management

func createTempFile(pattern string, data []byte) (string, error) {
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", err
	}

	tmpFile.Close()
	return tmpFile.Name(), nil
}

func removeTempFile(path string) {
	os.Remove(path)
}
