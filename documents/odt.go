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

// ODTParser handles OpenDocument Text (ODT) file parsing
type ODTParser struct{}

// Parse reads an ODT file and extracts its text content
func (p *ODTParser) Parse(path string) (*Document, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var text string
	var title string

	for _, f := range r.File {
		switch f.Name {
		case "content.xml":
			rc, err := f.Open()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}
			text = extractODTText(data)

		case "meta.xml":
			rc, err := f.Open()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}
			title = extractODTTitle(data)
		}
	}

	if strings.TrimSpace(text) == "" {
		return nil, ErrEmptyDocument
	}

	if title == "" {
		title = getTitleFromFilename(filepath.Base(path))
	}

	// Process the extracted text
	processor := DefaultProcessor()
	doc := processor.Process(text, title, path)
	doc.Path = path

	if doc.TotalWords == 0 {
		return nil, ErrEmptyDocument
	}

	return doc, nil
}

// ParseBytes parses ODT content from bytes
func (p *ODTParser) ParseBytes(data []byte, filename string) (*Document, error) {
	// Create a temporary file since zip.OpenReader requires a file
	tmpFile, err := os.CreateTemp("", "cheetah-*.odt")
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
func (p *ODTParser) SupportedExtensions() []string {
	return []string{".odt"}
}

// extractODTText extracts text content from ODT content.xml
func extractODTText(data []byte) string {
	var result strings.Builder

	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var inText bool
	var textDepth int

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Track text:p and text:h elements
			if t.Name.Local == "p" || t.Name.Local == "h" {
				if t.Name.Space == "urn:oasis:names:tc:opendocument:xmlns:text:1.0" ||
					strings.HasSuffix(t.Name.Space, ":text") {
					inText = true
					textDepth++
					if textDepth == 1 {
						result.WriteString("\n")
					}
				}
			}
			// Handle line breaks
			if t.Name.Local == "line-break" {
				result.WriteString("\n")
			}
			// Handle tabs
			if t.Name.Local == "tab" {
				result.WriteString(" ")
			}
			// Handle spaces
			if t.Name.Local == "s" {
				result.WriteString(" ")
			}

		case xml.EndElement:
			if t.Name.Local == "p" || t.Name.Local == "h" {
				if inText {
					textDepth--
					if textDepth <= 0 {
						inText = false
						textDepth = 0
						result.WriteString("\n")
					}
				}
			}

		case xml.CharData:
			if inText {
				text := strings.TrimSpace(string(t))
				if text != "" {
					result.WriteString(text)
					result.WriteString(" ")
				}
			}
		}
	}

	return result.String()
}

// extractODTTitle extracts the title from ODT meta.xml
func extractODTTitle(data []byte) string {
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var inTitle bool

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "title" {
				inTitle = true
			}

		case xml.EndElement:
			if t.Name.Local == "title" {
				inTitle = false
			}

		case xml.CharData:
			if inTitle {
				title := strings.TrimSpace(string(t))
				if title != "" {
					return title
				}
			}
		}
	}

	return ""
}
