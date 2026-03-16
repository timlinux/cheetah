// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

/**
 * Local storage wrapper for reading session persistence
 * Stores reading positions and preferences per document
 */

const STORAGE_KEY = 'cheetah_sessions';
const SETTINGS_KEY = 'cheetah_settings';

/**
 * Generate a hash for document content (for session identification)
 * Uses a simple hash for browser compatibility
 * @param {string} content - Document content to hash
 * @returns {string}
 */
async function hashContent(content) {
  // Use SubtleCrypto if available
  if (window.crypto && window.crypto.subtle) {
    const encoder = new TextEncoder();
    const data = encoder.encode(content);
    const hashBuffer = await window.crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
  }

  // Fallback: simple hash
  let hash = 0;
  for (let i = 0; i < content.length; i++) {
    const char = content.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // Convert to 32bit integer
  }
  return Math.abs(hash).toString(16);
}

/**
 * Get all saved sessions
 * @returns {Object}
 */
export function getSessions() {
  try {
    const data = localStorage.getItem(STORAGE_KEY);
    return data ? JSON.parse(data) : {};
  } catch (e) {
    console.error('Error reading sessions:', e);
    return {};
  }
}

/**
 * Save a reading session
 * @param {string} documentHash - Hash of document content
 * @param {Object} session - Session data
 */
export function saveSession(documentHash, session) {
  try {
    const sessions = getSessions();
    sessions[documentHash] = {
      ...session,
      lastAccessed: new Date().toISOString(),
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(sessions));
  } catch (e) {
    console.error('Error saving session:', e);
  }
}

/**
 * Get a specific session by document hash
 * @param {string} documentHash - Hash of document content
 * @returns {Object|null}
 */
export function getSession(documentHash) {
  const sessions = getSessions();
  return sessions[documentHash] || null;
}

/**
 * Delete a session
 * @param {string} documentHash - Hash of document content
 */
export function deleteSession(documentHash) {
  try {
    const sessions = getSessions();
    delete sessions[documentHash];
    localStorage.setItem(STORAGE_KEY, JSON.stringify(sessions));
  } catch (e) {
    console.error('Error deleting session:', e);
  }
}

/**
 * Get recent sessions sorted by last accessed, deduplicated by title
 * @param {number} limit - Maximum number of sessions to return
 * @returns {Array}
 */
export function getRecentSessions(limit = 10) {
  const sessions = getSessions();
  const sortedSessions = Object.entries(sessions)
    .map(([hash, data]) => ({ hash, ...data }))
    .sort((a, b) => new Date(b.lastAccessed) - new Date(a.lastAccessed));

  // Deduplicate by title, keeping the most recent entry for each title
  const seenTitles = new Set();
  const deduplicated = sortedSessions.filter(session => {
    const normalizedTitle = (session.title || '').toLowerCase().trim();
    if (seenTitles.has(normalizedTitle)) {
      return false;
    }
    seenTitles.add(normalizedTitle);
    return true;
  });

  return deduplicated.slice(0, limit);
}

/**
 * Create or update a session from document content
 * @param {string} content - First N characters of document for hashing
 * @param {string} title - Document title
 * @param {number} totalWords - Total word count
 * @param {number} position - Current word position
 * @param {number} wpm - Current WPM setting
 * @returns {Promise<string>} - Document hash
 */
export async function createOrUpdateSession(content, title, totalWords, position = 0, wpm = 300) {
  // Use first 10KB for hashing to keep it fast
  const hashContent$ = content.substring(0, 10240);
  const hash = await hashContent(hashContent$);

  saveSession(hash, {
    title,
    totalWords,
    position,
    wpm,
    progress: totalWords > 0 ? (position / totalWords) * 100 : 0,
  });

  return hash;
}

/**
 * Get user settings
 * @returns {Object}
 */
export function getSettings() {
  try {
    const data = localStorage.getItem(SETTINGS_KEY);
    const defaults = {
      defaultWpm: 300,
      showAds: true,
      displayAllCaps: true, // All caps enabled by default
    };
    return data ? { ...defaults, ...JSON.parse(data) } : defaults;
  } catch (e) {
    console.error('Error reading settings:', e);
    return { defaultWpm: 300, showAds: true, displayAllCaps: true };
  }
}

/**
 * Save user settings
 * @param {Object} settings - Settings to save
 */
export function saveSettings(settings) {
  try {
    const current = getSettings();
    localStorage.setItem(SETTINGS_KEY, JSON.stringify({ ...current, ...settings }));
  } catch (e) {
    console.error('Error saving settings:', e);
  }
}

/**
 * Clear all data
 */
export function clearAll() {
  localStorage.removeItem(STORAGE_KEY);
  localStorage.removeItem(SETTINGS_KEY);
}

export { hashContent };
