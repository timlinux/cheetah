// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package documents

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

// Processor handles text processing - converting raw text into words and paragraphs
type Processor struct {
	// MinWordLength filters out words shorter than this
	MinWordLength int

	// MaxWordLength filters out words longer than this (0 = no limit)
	MaxWordLength int

	// PreserveCase keeps original case if true, otherwise converts to lowercase
	PreserveCase bool
}

// DefaultProcessor returns a processor with sensible defaults
func DefaultProcessor() *Processor {
	return &Processor{
		MinWordLength: 1,
		MaxWordLength: 0, // No limit
		PreserveCase:  true,
	}
}

// Process takes raw text and converts it into a structured document
func (p *Processor) Process(text string, title string, path string) *Document {
	paragraphs := p.ExtractParagraphs(text)

	totalWords := 0
	for _, para := range paragraphs {
		totalWords += len(para.Words)
	}

	// Calculate content hash
	hash := sha256.Sum256([]byte(text))
	hashStr := hex.EncodeToString(hash[:])

	return &Document{
		Title:      title,
		Paragraphs: paragraphs,
		TotalWords: totalWords,
		Hash:       hashStr,
		Path:       path,
	}
}

// ExtractParagraphs splits text into paragraphs and words
func (p *Processor) ExtractParagraphs(text string) []Paragraph {
	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// Split on double newlines (or more) to get paragraphs
	paraRegex := regexp.MustCompile(`\n\s*\n`)
	rawParagraphs := paraRegex.Split(text, -1)

	var paragraphs []Paragraph
	idx := 0

	for _, rawPara := range rawParagraphs {
		words := p.ExtractWords(rawPara)
		if len(words) == 0 {
			continue
		}

		paragraphs = append(paragraphs, Paragraph{
			Words: words,
			Index: idx,
		})
		idx++
	}

	return paragraphs
}

// ExtractWords splits a paragraph into individual words
func (p *Processor) ExtractWords(text string) []string {
	// Replace newlines with spaces
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	// Split on whitespace
	rawWords := strings.Fields(text)

	var words []string
	for _, word := range rawWords {
		// Clean the word
		cleaned := p.CleanWord(word)
		if cleaned == "" {
			continue
		}

		// Apply length filters
		runeCount := len([]rune(cleaned))
		if runeCount < p.MinWordLength {
			continue
		}
		if p.MaxWordLength > 0 && runeCount > p.MaxWordLength {
			continue
		}

		// Apply case transformation
		if !p.PreserveCase {
			cleaned = strings.ToLower(cleaned)
		}

		words = append(words, cleaned)
	}

	return words
}

// CleanWord removes unwanted characters from a word while preserving punctuation
// that's attached to words (like periods, commas, quotes)
func (p *Processor) CleanWord(word string) string {
	// Remove leading/trailing whitespace
	word = strings.TrimSpace(word)

	if word == "" {
		return ""
	}

	// Keep only letters, numbers, and common punctuation
	var cleaned strings.Builder
	for _, r := range word {
		if unicode.IsLetter(r) || unicode.IsNumber(r) ||
			r == '\'' || r == '-' || r == '.' || r == ',' ||
			r == '!' || r == '?' || r == ':' || r == ';' ||
			r == '"' || r == '\u201c' || r == '\u201d' || r == '\u2018' || r == '\u2019' {
			cleaned.WriteRune(r)
		}
	}

	result := cleaned.String()

	// Remove leading/trailing punctuation except quotes
	result = strings.Trim(result, ".,;:!?-")

	return result
}

// GetAllWords returns a flat list of all words in a document
func GetAllWords(doc *Document) []string {
	var words []string
	for _, para := range doc.Paragraphs {
		words = append(words, para.Words...)
	}
	return words
}

// GetWordAt returns the word at a specific position across all paragraphs
func GetWordAt(doc *Document, index int) (word string, paragraphIndex int, ok bool) {
	if index < 0 || index >= doc.TotalWords {
		return "", 0, false
	}

	current := 0
	for _, para := range doc.Paragraphs {
		if current+len(para.Words) > index {
			wordIndex := index - current
			return para.Words[wordIndex], para.Index, true
		}
		current += len(para.Words)
	}

	return "", 0, false
}

// GetParagraphStartIndex returns the starting word index for a paragraph
func GetParagraphStartIndex(doc *Document, paragraphIndex int) int {
	if paragraphIndex <= 0 {
		return 0
	}

	wordIndex := 0
	for i, para := range doc.Paragraphs {
		if i >= paragraphIndex {
			break
		}
		wordIndex += len(para.Words)
	}
	return wordIndex
}

// GetParagraphForWordIndex returns which paragraph contains a given word index
func GetParagraphForWordIndex(doc *Document, wordIndex int) int {
	if wordIndex < 0 {
		return 0
	}

	current := 0
	for _, para := range doc.Paragraphs {
		if current+len(para.Words) > wordIndex {
			return para.Index
		}
		current += len(para.Words)
	}

	// Return last paragraph if index is beyond end
	if len(doc.Paragraphs) > 0 {
		return doc.Paragraphs[len(doc.Paragraphs)-1].Index
	}
	return 0
}
