// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import Foundation
import Combine

// MARK: - Settings

/// User preferences stored in `UserDefaults` via `@AppStorage`.
/// Mirrors the settings schema used by the Go backend and web frontend.
class Settings: ObservableObject {

    // MARK: Keys

    enum Key: String {
        case defaultWPM       = "cheetah.defaultWPM"
        case showPreviousWord = "cheetah.showPreviousWord"
        case showNextWords    = "cheetah.showNextWords"
        case nextWordsCount   = "cheetah.nextWordsCount"
        case autoSave         = "cheetah.autoSave"
        case autoSaveInterval = "cheetah.autoSaveInterval"
        case lastDirectory    = "cheetah.lastDirectory"
    }

    // MARK: Defaults

    static let defaultDefaultWPM       = 300
    static let defaultShowPreviousWord = true
    static let defaultShowNextWords    = true
    static let defaultNextWordsCount   = 3
    static let defaultAutoSave         = true
    static let defaultAutoSaveInterval = 50

    // MARK: Published properties

    @Published var defaultWPM: Int {
        didSet { UserDefaults.standard.set(defaultWPM, forKey: Key.defaultWPM.rawValue) }
    }

    @Published var showPreviousWord: Bool {
        didSet { UserDefaults.standard.set(showPreviousWord, forKey: Key.showPreviousWord.rawValue) }
    }

    @Published var showNextWords: Bool {
        didSet { UserDefaults.standard.set(showNextWords, forKey: Key.showNextWords.rawValue) }
    }

    @Published var nextWordsCount: Int {
        didSet { UserDefaults.standard.set(nextWordsCount, forKey: Key.nextWordsCount.rawValue) }
    }

    @Published var autoSave: Bool {
        didSet { UserDefaults.standard.set(autoSave, forKey: Key.autoSave.rawValue) }
    }

    @Published var autoSaveInterval: Int {
        didSet { UserDefaults.standard.set(autoSaveInterval, forKey: Key.autoSaveInterval.rawValue) }
    }

    @Published var lastDirectory: String {
        didSet { UserDefaults.standard.set(lastDirectory, forKey: Key.lastDirectory.rawValue) }
    }

    // MARK: Init

    init() {
        let ud = UserDefaults.standard

        // Register defaults once
        ud.register(defaults: [
            Key.defaultWPM.rawValue:       Settings.defaultDefaultWPM,
            Key.showPreviousWord.rawValue:  Settings.defaultShowPreviousWord,
            Key.showNextWords.rawValue:     Settings.defaultShowNextWords,
            Key.nextWordsCount.rawValue:    Settings.defaultNextWordsCount,
            Key.autoSave.rawValue:          Settings.defaultAutoSave,
            Key.autoSaveInterval.rawValue:  Settings.defaultAutoSaveInterval,
            Key.lastDirectory.rawValue:     "",
        ])

        defaultWPM       = ud.integer(forKey: Key.defaultWPM.rawValue)
        showPreviousWord = ud.bool(forKey: Key.showPreviousWord.rawValue)
        showNextWords    = ud.bool(forKey: Key.showNextWords.rawValue)
        nextWordsCount   = ud.integer(forKey: Key.nextWordsCount.rawValue)
        autoSave         = ud.bool(forKey: Key.autoSave.rawValue)
        autoSaveInterval = ud.integer(forKey: Key.autoSaveInterval.rawValue)
        lastDirectory    = ud.string(forKey: Key.lastDirectory.rawValue) ?? ""
    }

    // MARK: Speed presets

    /// Maps key 1–9 to WPM values (matching TUI/web behaviour).
    static func wpm(forPreset key: Int) -> Int? {
        let presets = [200, 300, 400, 500, 600, 700, 800, 900, 1000]
        guard key >= 1, key <= presets.count else { return nil }
        return presets[key - 1]
    }

    /// Returns a human-readable label for a WPM value.
    static func speedLabel(for wpm: Int) -> String {
        switch wpm {
        case ..<200:  return "Relaxed"
        case ..<350:  return "Normal"
        case ..<500:  return "Fast"
        case ..<700:  return "Very Fast"
        default:      return "Speed Demon"
        }
    }
}
