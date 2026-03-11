// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

/**
 * DOCX parser using mammoth.js
 * Extracts text content from Word documents
 */

import mammoth from 'mammoth';

/**
 * Parse a DOCX file and extract words
 * @param {File} file - The DOCX file to parse
 * @returns {Promise<{words: string[], title: string}>}
 */
export async function parseDocxFile(file) {
  const arrayBuffer = await file.arrayBuffer();

  // Extract raw text using mammoth
  const result = await mammoth.extractRawText({ arrayBuffer });
  const text = result.value;

  // Clean and split into words
  const cleaned = text
    // Normalize whitespace
    .replace(/\s+/g, ' ')
    // Fix common issues
    .replace(/[""]/g, '"')
    .replace(/['']/g, "'")
    .trim();

  const words = cleaned
    .split(/\s+/)
    .filter(word => word.length > 0);

  // Extract title from filename
  const title = file.name.replace(/\.docx?$/i, '');

  return { words, title };
}

export default parseDocxFile;
