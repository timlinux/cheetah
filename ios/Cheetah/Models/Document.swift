// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import Foundation
import CryptoKit

// MARK: - Document

/// A parsed document ready for RSVP reading.
struct Document: Identifiable {
    let id = UUID()

    /// Title derived from metadata or filename.
    let title: String

    /// Original file path (nil for in-memory documents).
    let path: String?

    /// All paragraphs in reading order.
    let paragraphs: [Paragraph]

    /// Flat array of all words across all paragraphs, in order.
    let words: [String]

    /// SHA-256 hash of the first 10 KB of word content for identity.
    let hash: String

    /// Total word count.
    var totalWords: Int { words.count }

    /// Total paragraph count.
    var totalParagraphs: Int { paragraphs.count }
}

// MARK: - Paragraph

/// A single paragraph from a document.
struct Paragraph: Identifiable {
    let id = UUID()

    /// 0-based index into the document's paragraph array.
    let index: Int

    /// Words that make up this paragraph.
    let words: [String]

    /// Index of the first word in the flat `words` array.
    let startWordIndex: Int
}

// MARK: - Document builder helpers

extension Document {
    /// Builds a Document from an array of paragraph word lists.
    static func build(title: String, path: String?, paragraphWords: [[String]]) -> Document {
        var paragraphs: [Paragraph] = []
        var allWords: [String] = []
        var offset = 0

        for (i, words) in paragraphWords.enumerated() where !words.isEmpty {
            paragraphs.append(Paragraph(index: i, words: words, startWordIndex: offset))
            allWords.append(contentsOf: words)
            offset += words.count
        }

        // Hash: first 10 KB of joined words
        let sampleText = allWords.prefix(2000).joined(separator: " ")
        let sampleData = Data(sampleText.utf8.prefix(10240))
        let digest = SHA256.hash(data: sampleData)
        let hashString = digest.map { String(format: "%02x", $0) }.joined()

        return Document(
            title: title,
            path: path,
            paragraphs: paragraphs,
            words: allWords,
            hash: hashString
        )
    }
}

// MARK: - Navigation helpers

extension Document {
    /// Returns the paragraph index for a given flat word index.
    func paragraphIndex(forWordIndex wordIndex: Int) -> Int {
        for para in paragraphs.reversed() {
            if wordIndex >= para.startWordIndex {
                return para.index
            }
        }
        return 0
    }

    /// Returns the flat word index for the first word of a paragraph.
    func wordIndex(forParagraph paragraphIndex: Int) -> Int {
        guard paragraphIndex >= 0, paragraphIndex < paragraphs.count else { return 0 }
        return paragraphs[paragraphIndex].startWordIndex
    }
}
