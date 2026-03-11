// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

/**
 * EPUB parser using JSZip
 * Extracts text content from EPUB files
 */

import JSZip from 'jszip';

/**
 * Parse an EPUB file and extract words
 * @param {File} file - The EPUB file to parse
 * @returns {Promise<{words: string[], title: string}>}
 */
export async function parseEpubFile(file) {
  const arrayBuffer = await file.arrayBuffer();
  const zip = await JSZip.loadAsync(arrayBuffer);

  let title = file.name.replace(/\.epub$/i, '');
  let fullText = '';

  // Find and parse the container.xml to get the content.opf location
  const containerFile = zip.file('META-INF/container.xml');
  if (!containerFile) {
    throw new Error('Invalid EPUB: missing container.xml');
  }

  const containerXml = await containerFile.async('string');
  const containerDoc = new DOMParser().parseFromString(containerXml, 'application/xml');
  const rootfileEl = containerDoc.querySelector('rootfile');
  if (!rootfileEl) {
    throw new Error('Invalid EPUB: missing rootfile');
  }

  const opfPath = rootfileEl.getAttribute('full-path');
  const opfDir = opfPath.substring(0, opfPath.lastIndexOf('/') + 1);

  // Parse the OPF file to get metadata and spine
  const opfFile = zip.file(opfPath);
  if (!opfFile) {
    throw new Error('Invalid EPUB: missing OPF file');
  }

  const opfXml = await opfFile.async('string');
  const opfDoc = new DOMParser().parseFromString(opfXml, 'application/xml');

  // Get title from metadata
  const titleEl = opfDoc.querySelector('metadata title, metadata dc\\:title');
  if (titleEl && titleEl.textContent) {
    title = titleEl.textContent;
  }

  // Get the spine order
  const spineItems = opfDoc.querySelectorAll('spine itemref');
  const manifest = {};

  // Build manifest lookup
  opfDoc.querySelectorAll('manifest item').forEach(item => {
    const id = item.getAttribute('id');
    const href = item.getAttribute('href');
    const mediaType = item.getAttribute('media-type');
    manifest[id] = { href, mediaType };
  });

  // Process spine items in order
  for (const itemRef of spineItems) {
    const idref = itemRef.getAttribute('idref');
    const manifestItem = manifest[idref];

    if (manifestItem && manifestItem.mediaType === 'application/xhtml+xml') {
      const contentPath = opfDir + manifestItem.href;
      const contentFile = zip.file(contentPath);

      if (contentFile) {
        const html = await contentFile.async('string');
        const text = extractTextFromHtml(html);
        fullText += text + '\n\n';
      }
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
 * Extract plain text from HTML content
 * @param {string} html - HTML content
 * @returns {string}
 */
function extractTextFromHtml(html) {
  const doc = new DOMParser().parseFromString(html, 'text/html');

  // Remove script and style elements
  doc.querySelectorAll('script, style').forEach(el => el.remove());

  // Get text content
  return doc.body ? doc.body.textContent || '' : '';
}

export default parseEpubFile;
