// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

/**
 * PDF parser using PDF.js
 * Extracts text content from PDF files
 */

import * as pdfjsLib from 'pdfjs-dist';

// Set up the worker - use CDN for simplicity
pdfjsLib.GlobalWorkerOptions.workerSrc = `https://cdnjs.cloudflare.com/ajax/libs/pdf.js/${pdfjsLib.version}/pdf.worker.min.js`;

/**
 * Parse a PDF file and extract words
 * @param {File} file - The PDF file to parse
 * @returns {Promise<{words: string[], title: string}>}
 */
export async function parsePdfFile(file) {
  const arrayBuffer = await file.arrayBuffer();
  const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise;

  let fullText = '';
  const numPages = pdf.numPages;

  // Extract text from each page
  for (let pageNum = 1; pageNum <= numPages; pageNum++) {
    const page = await pdf.getPage(pageNum);
    const textContent = await page.getTextContent();

    // Concatenate text items
    const pageText = textContent.items
      .map(item => item.str)
      .join(' ');

    fullText += pageText + '\n\n';
  }

  // Clean and split into words
  const cleaned = fullText
    // Normalize whitespace
    .replace(/\s+/g, ' ')
    // Fix common OCR issues
    .replace(/[""]/g, '"')
    .replace(/['']/g, "'")
    // Remove multiple spaces
    .replace(/  +/g, ' ')
    .trim();

  const words = cleaned
    .split(/\s+/)
    .filter(word => word.length > 0);

  // Try to get title from PDF metadata or filename
  let title = file.name.replace(/\.pdf$/i, '');
  try {
    const metadata = await pdf.getMetadata();
    if (metadata.info && metadata.info.Title) {
      title = metadata.info.Title;
    }
  } catch (e) {
    // Ignore metadata errors
  }

  return { words, title };
}

export default parsePdfFile;
