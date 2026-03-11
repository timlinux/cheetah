// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

/**
 * Plain text and Markdown parser
 * Extracts words from plain text files with heading and list metadata
 */

/**
 * Parse a text file and extract words with metadata
 * @param {File} file - The file to parse
 * @returns {Promise<{words: Array<{text: string, headingLevel?: number, listNumber?: number}>, title: string}>}
 */
export async function parseTextFile(file) {
  const text = await file.text();
  return parseTextContent(text, file.name);
}

/**
 * Parse text content and extract words with metadata
 * @param {string} text - The text content
 * @param {string} filename - The filename for title extraction
 * @returns {{words: Array<{text: string, headingLevel?: number, listNumber?: number}>, title: string}}
 */
export function parseTextContent(text, filename = 'Document') {
  const words = [];
  let currentListNumber = null;

  // Split into lines first to detect structure
  const lines = text.split(/\r?\n/);

  for (const line of lines) {
    const trimmedLine = line.trim();
    if (!trimmedLine) continue;

    // Detect markdown headings
    const headingMatch = trimmedLine.match(/^(#{1,6})\s+(.+)$/);
    if (headingMatch) {
      const level = headingMatch[1].length;
      const headingText = headingMatch[2]
        .replace(/\*\*([^*]+)\*\*/g, '$1')
        .replace(/\*([^*]+)\*/g, '$1')
        .replace(/__([^_]+)__/g, '$1')
        .replace(/_([^_]+)_/g, '$1');

      // Split heading into words, first word gets the heading level marker
      const headingWords = headingText.split(/\s+/).filter(w => w.length > 0);
      headingWords.forEach((word, idx) => {
        words.push({
          text: word,
          headingLevel: level,
          isHeadingStart: idx === 0,
          isHeadingEnd: idx === headingWords.length - 1,
        });
      });
      currentListNumber = null; // Reset list context
      continue;
    }

    // Detect numbered lists (1. 2. 3. etc.)
    const numberedListMatch = trimmedLine.match(/^(\d+)[\.\)]\s+(.+)$/);
    if (numberedListMatch) {
      currentListNumber = parseInt(numberedListMatch[1]);
      const listText = numberedListMatch[2]
        .replace(/\*\*([^*]+)\*\*/g, '$1')
        .replace(/\*([^*]+)\*/g, '$1');

      const listWords = listText.split(/\s+/).filter(w => w.length > 0);
      listWords.forEach((word, idx) => {
        words.push({
          text: word,
          listNumber: currentListNumber,
          isListStart: idx === 0,
          isListEnd: idx === listWords.length - 1,
        });
      });
      continue;
    }

    // Detect bullet lists (-, *, +)
    const bulletListMatch = trimmedLine.match(/^[-*+]\s+(.+)$/);
    if (bulletListMatch) {
      const listText = bulletListMatch[1]
        .replace(/\*\*([^*]+)\*\*/g, '$1')
        .replace(/\*([^*]+)\*/g, '$1');

      const listWords = listText.split(/\s+/).filter(w => w.length > 0);
      listWords.forEach((word, idx) => {
        words.push({
          text: word,
          isBullet: true,
          isListStart: idx === 0,
          isListEnd: idx === listWords.length - 1,
        });
      });
      continue;
    }

    // Regular paragraph - clear list context after a blank line or non-list content
    if (!numberedListMatch && !bulletListMatch) {
      currentListNumber = null;
    }

    // Clean up markdown formatting in regular text
    const cleanedLine = trimmedLine
      .replace(/\*\*([^*]+)\*\*/g, '$1')
      .replace(/\*([^*]+)\*/g, '$1')
      .replace(/__([^_]+)__/g, '$1')
      .replace(/_([^_]+)_/g, '$1')
      .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
      .replace(/`([^`]+)`/g, '$1');

    const lineWords = cleanedLine.split(/\s+/).filter(w => w.length > 0);
    lineWords.forEach(word => {
      words.push({ text: word });
    });
  }

  // Extract title from filename
  const title = filename
    .replace(/\.(txt|md|markdown)$/i, '')
    .replace(/[-_]/g, ' ');

  return { words, title };
}

export default parseTextFile;
