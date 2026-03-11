// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

/**
 * Document parser factory
 * Selects the appropriate parser based on file type
 */

import parseTextFile from './text.js';
import parsePdfFile from './pdf.js';
import parseDocxFile from './docx.js';
import parseEpubFile from './epub.js';
import parseOdtFile from './odt.js';

/**
 * Parse a document file and extract words
 * @param {File} file - The file to parse
 * @returns {Promise<{words: string[], title: string}>}
 */
export async function parseDocument(file) {
  const extension = file.name.split('.').pop().toLowerCase();
  const mimeType = file.type;

  // Determine parser based on extension or MIME type
  switch (extension) {
    case 'txt':
    case 'md':
    case 'markdown':
      return parseTextFile(file);

    case 'pdf':
      return parsePdfFile(file);

    case 'docx':
    case 'doc':
      return parseDocxFile(file);

    case 'epub':
      return parseEpubFile(file);

    case 'odt':
      return parseOdtFile(file);

    default:
      // Try to determine by MIME type
      if (mimeType.startsWith('text/')) {
        return parseTextFile(file);
      }
      if (mimeType === 'application/pdf') {
        return parsePdfFile(file);
      }
      if (mimeType.includes('wordprocessingml')) {
        return parseDocxFile(file);
      }
      if (mimeType === 'application/epub+zip') {
        return parseEpubFile(file);
      }
      if (mimeType.includes('opendocument')) {
        return parseOdtFile(file);
      }

      throw new Error(`Unsupported file format: ${extension || mimeType}`);
  }
}

/**
 * Get supported file extensions
 * @returns {string[]}
 */
export function getSupportedExtensions() {
  return ['txt', 'md', 'markdown', 'pdf', 'docx', 'doc', 'epub', 'odt'];
}

/**
 * Check if a file is supported
 * @param {File} file - The file to check
 * @returns {boolean}
 */
export function isSupported(file) {
  const extension = file.name.split('.').pop().toLowerCase();
  return getSupportedExtensions().includes(extension);
}

export default parseDocument;
