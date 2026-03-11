// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package documents

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"path/filepath"
	"strings"
)

// EPUBParser handles EPUB ebook parsing using standard library
type EPUBParser struct{}

// Parse reads an EPUB file and extracts its text content
func (p *EPUBParser) Parse(path string) (*Document, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var textBuilder strings.Builder
	var title string

	// Find and parse container.xml to locate content.opf
	var rootfilePath string
	for _, f := range r.File {
		if f.Name == "META-INF/container.xml" {
			rootfilePath = p.parseContainer(f)
			break
		}
	}

	// If no container, try common paths
	if rootfilePath == "" {
		rootfilePath = "OEBPS/content.opf"
	}

	// Parse OPF file to get spine order and title
	var spineItems []string
	opfDir := filepath.Dir(rootfilePath)
	for _, f := range r.File {
		if f.Name == rootfilePath {
			title, spineItems = p.parseOPF(f, opfDir)
			break
		}
	}

	// If no spine found, just read all HTML/XHTML files
	if len(spineItems) == 0 {
		for _, f := range r.File {
			if strings.HasSuffix(f.Name, ".html") || strings.HasSuffix(f.Name, ".xhtml") ||
				strings.HasSuffix(f.Name, ".htm") {
				content := p.readFileContent(f)
				text := stripHTML(content)
				textBuilder.WriteString(text)
				textBuilder.WriteString("\n\n")
			}
		}
	} else {
		// Read files in spine order
		for _, item := range spineItems {
			for _, f := range r.File {
				if f.Name == item {
					content := p.readFileContent(f)
					text := stripHTML(content)
					textBuilder.WriteString(text)
					textBuilder.WriteString("\n\n")
					break
				}
			}
		}
	}

	text := textBuilder.String()
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

// ParseBytes parses EPUB content from bytes
func (p *EPUBParser) ParseBytes(data []byte, filename string) (*Document, error) {
	// Create a temporary file since zip.OpenReader requires a file
	tmpFile, err := createTempFile("cheetah-*.epub", data)
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
func (p *EPUBParser) SupportedExtensions() []string {
	return []string{".epub"}
}

// parseContainer extracts the rootfile path from container.xml
func (p *EPUBParser) parseContainer(f *zip.File) string {
	rc, err := f.Open()
	if err != nil {
		return ""
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return ""
	}

	// Simple XML parsing for container
	type rootfile struct {
		FullPath string `xml:"full-path,attr"`
	}
	type container struct {
		Rootfile rootfile `xml:"rootfiles>rootfile"`
	}

	var c container
	if err := xml.Unmarshal(data, &c); err != nil {
		return ""
	}

	return c.Rootfile.FullPath
}

// parseOPF extracts title and spine items from content.opf
func (p *EPUBParser) parseOPF(f *zip.File, opfDir string) (string, []string) {
	rc, err := f.Open()
	if err != nil {
		return "", nil
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return "", nil
	}

	// OPF structure
	type metadata struct {
		Title string `xml:"metadata>title"`
	}
	type manifestItem struct {
		ID   string `xml:"id,attr"`
		Href string `xml:"href,attr"`
	}
	type spineItemref struct {
		Idref string `xml:"idref,attr"`
	}
	type opfPackage struct {
		Metadata metadata       `xml:"metadata"`
		Manifest []manifestItem `xml:"manifest>item"`
		Spine    []spineItemref `xml:"spine>itemref"`
	}

	var pkg opfPackage
	if err := xml.Unmarshal(data, &pkg); err != nil {
		return "", nil
	}

	// Build ID -> href map
	idToHref := make(map[string]string)
	for _, item := range pkg.Manifest {
		idToHref[item.ID] = item.Href
	}

	// Build ordered list of spine items
	var spineItems []string
	for _, ref := range pkg.Spine {
		if href, ok := idToHref[ref.Idref]; ok {
			// Resolve relative path
			fullPath := href
			if opfDir != "" && opfDir != "." {
				fullPath = opfDir + "/" + href
			}
			spineItems = append(spineItems, fullPath)
		}
	}

	return pkg.Metadata.Title, spineItems
}

// readFileContent reads the content of a zip file
func (p *EPUBParser) readFileContent(f *zip.File) string {
	rc, err := f.Open()
	if err != nil {
		return ""
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return ""
	}

	return string(data)
}

// stripHTML removes HTML tags from text (moved from original epub.go)
func stripHTML(html string) string {
	var result strings.Builder
	inTag := false
	inScript := false
	inStyle := false

	lowerHTML := strings.ToLower(html)

	for i := 0; i < len(html); i++ {
		c := html[i]

		if c == '<' {
			// Check for script/style tags
			remaining := lowerHTML[i:]
			if strings.HasPrefix(remaining, "<script") {
				inScript = true
			} else if strings.HasPrefix(remaining, "</script") {
				inScript = false
			} else if strings.HasPrefix(remaining, "<style") {
				inStyle = true
			} else if strings.HasPrefix(remaining, "</style") {
				inStyle = false
			}

			// Check for block elements that should add newlines
			if strings.HasPrefix(remaining, "<p") ||
				strings.HasPrefix(remaining, "<div") ||
				strings.HasPrefix(remaining, "<br") ||
				strings.HasPrefix(remaining, "<h1") ||
				strings.HasPrefix(remaining, "<h2") ||
				strings.HasPrefix(remaining, "<h3") ||
				strings.HasPrefix(remaining, "<h4") ||
				strings.HasPrefix(remaining, "<h5") ||
				strings.HasPrefix(remaining, "<h6") ||
				strings.HasPrefix(remaining, "<li") {
				result.WriteString("\n")
			}

			inTag = true
			continue
		}

		if c == '>' {
			inTag = false
			continue
		}

		if !inTag && !inScript && !inStyle {
			// Decode common HTML entities
			if c == '&' && i+3 < len(html) {
				entity := html[i:]
				if strings.HasPrefix(entity, "&nbsp;") {
					result.WriteRune(' ')
					i += 5
					continue
				} else if strings.HasPrefix(entity, "&amp;") {
					result.WriteRune('&')
					i += 4
					continue
				} else if strings.HasPrefix(entity, "&lt;") {
					result.WriteRune('<')
					i += 3
					continue
				} else if strings.HasPrefix(entity, "&gt;") {
					result.WriteRune('>')
					i += 3
					continue
				} else if strings.HasPrefix(entity, "&quot;") {
					result.WriteRune('"')
					i += 5
					continue
				} else if strings.HasPrefix(entity, "&apos;") {
					result.WriteRune('\'')
					i += 5
					continue
				} else if strings.HasPrefix(entity, "&#") {
					// Skip numeric entities for now
					end := strings.Index(entity, ";")
					if end > 0 && end < 10 {
						result.WriteRune(' ')
						i += end
						continue
					}
				}
			}
			result.WriteByte(c)
		}
	}

	return result.String()
}
