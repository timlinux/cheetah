// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import Foundation

// MARK: - ReadingSession

/// A saved reading position for a document.
struct ReadingSession: Codable, Identifiable {
    /// Unique identifier = document hash.
    var id: String { documentHash }

    /// SHA-256 hash of the document content.
    let documentHash: String

    /// User-visible document title.
    let documentTitle: String

    /// Word index of the last reading position.
    let lastPosition: Int

    /// Total word count for the document.
    let totalWords: Int

    /// WPM at last use.
    let lastWPM: Int

    /// When this session was last accessed.
    let lastAccessed: Date

    /// Reading progress (0.0 – 1.0).
    var progress: Double {
        totalWords > 0 ? Double(lastPosition) / Double(totalWords) : 0
    }
}

// MARK: - SessionStore

/// Persists reading sessions to `UserDefaults`.
/// Matches the session schema used by the Go backend (`sessions.json`).
final class SessionStore: ObservableObject {

    // MARK: Singleton

    static let shared = SessionStore()

    // MARK: Published state

    @Published private(set) var sessions: [ReadingSession] = []

    // MARK: Private

    private let defaultsKey = "cheetah.sessions.v1"

    private init() {
        load()
    }

    // MARK: - Public interface

    /// Save or update a reading session.
    func save(documentHash: String,
              title: String,
              position: Int,
              totalWords: Int,
              wpm: Int) {
        let session = ReadingSession(
            documentHash: documentHash,
            documentTitle: title,
            lastPosition: position,
            totalWords: totalWords,
            lastWPM: wpm,
            lastAccessed: Date()
        )

        if let idx = sessions.firstIndex(where: { $0.documentHash == documentHash }) {
            sessions[idx] = session
        } else {
            sessions.append(session)
        }

        // Keep most-recently-accessed first
        sessions.sort { $0.lastAccessed > $1.lastAccessed }

        persist()
    }

    /// Retrieve a session by document hash.
    func session(for hash: String) -> ReadingSession? {
        sessions.first { $0.documentHash == hash }
    }

    /// Remove a session by document hash.
    func delete(hash: String) {
        sessions.removeAll { $0.documentHash == hash }
        persist()
    }

    /// Remove all sessions.
    func clearAll() {
        sessions.removeAll()
        persist()
    }

    /// Returns up to `limit` sessions sorted by most-recently-accessed.
    func recentSessions(limit: Int = 5) -> [ReadingSession] {
        Array(sessions.prefix(limit))
    }

    // MARK: - Persistence

    private func load() {
        guard let data = UserDefaults.standard.data(forKey: defaultsKey) else { return }
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        if let loaded = try? decoder.decode([ReadingSession].self, from: data) {
            sessions = loaded.sorted { $0.lastAccessed > $1.lastAccessed }
        }
    }

    private func persist() {
        let encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
        guard let data = try? encoder.encode(sessions) else { return }
        UserDefaults.standard.set(data, forKey: defaultsKey)
    }
}
