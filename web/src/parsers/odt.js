// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

/**
 * ODT (OpenDocument Text) parser using JSZip
 * Extracts text content from OpenDocument files
 */

import JSZip from 'jszip';

/**
 * Parse an ODT file and extract words
 * @param {File} file - The ODT file to parse
 * @returns {Promise<{words: string[], title: string}>}
 */
export async function parseOdtFile(file) {
  const arrayBuffer = await file.arrayBuffer();
  const zip = await JSZip.loadAsync(arrayBuffer);

  let title = file.name.replace(/\.odt$/i, '');
  let fullText = '';

  // ODT files contain content in content.xml
  const contentFile = zip.file('content.xml');
  if (!contentFile) {
    throw new Error('Invalid ODT: missing content.xml');
  }

  const contentXml = await contentFile.async('string');
  const contentDoc = new DOMParser().parseFromString(contentXml, 'application/xml');

  // Extract text from all text:p (paragraph) and text:h (heading) elements
  const textElements = contentDoc.querySelectorAll('text\\:p, text\\:h, p, h');

  for (const el of textElements) {
    const text = extractOdtText(el);
    if (text.trim()) {
      fullText += text + '\n';
    }
  }

  // Try to get title from meta.xml
  const metaFile = zip.file('meta.xml');
  if (metaFile) {
    try {
      const metaXml = await metaFile.async('string');
      const metaDoc = new DOMParser().parseFromString(metaXml, 'application/xml');
      const titleEl = metaDoc.querySelector('title, dc\\:title');
      if (titleEl && titleEl.textContent) {
        title = titleEl.textContent;
      }
    } catch (e) {
      // Ignore meta parsing errors
    }
  }

  // Clean and split into words
  const cleaned = fullText
    .replace(/\s+/g, ' ')
    .replace(/[""]/g, '"')
    .replace(/['']/g, "'")
    .trim();

  const words = cleaned
    .split(/\s+/)
    .filter(word => word.length > 0);

  return { words, title };
}

/**
 * Recursively extract text from an ODT XML element
 * @param {Element} element - The XML element
 * @returns {string}
 */
function extractOdtText(element) {
  let text = '';

  for (const node of element.childNodes) {
    if (node.nodeType === Node.TEXT_NODE) {
      text += node.textContent;
    } else if (node.nodeType === Node.ELEMENT_NODE) {
      // Handle text:s (space) elements
      if (node.localName === 's') {
        const count = parseInt(node.getAttribute('text:c') || node.getAttribute('c') || '1');
        text += ' '.repeat(count);
      }
      // Handle text:tab elements
      else if (node.localName === 'tab') {
        text += ' ';
      }
      // Handle text:line-break elements
      else if (node.localName === 'line-break') {
        text += '\n';
      }
      // Recursively process other elements
      else {
        text += extractOdtText(node);
      }
    }
  }

  return text;
}

export default parseOdtFile;
