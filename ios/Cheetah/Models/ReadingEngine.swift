// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import Foundation
import Combine

// MARK: - ReadingEngine

/// Core RSVP reading engine.  Drives word advancement via a `Timer` and
/// publishes state changes for SwiftUI views to observe.
@MainActor
final class ReadingEngine: ObservableObject {

    // MARK: Published state

    @Published var currentWord: String = ""
    @Published var previousWord: String = ""
    @Published var nextWords: [String] = []
    @Published var wordIndex: Int = 0
    @Published var totalWords: Int = 0
    @Published var paragraphIndex: Int = 0
    @Published var totalParagraphs: Int = 0
    @Published var wpm: Int = 300
    @Published var isPaused: Bool = true
    @Published var progress: Double = 0
    @Published var elapsedMs: Int64 = 0
    @Published var documentLoaded: Bool = false
    @Published var documentTitle: String = ""

    // MARK: Private state

    private var document: Document?
    private var timer: Timer?
    private var startTime: Date?
    private var accumulatedElapsed: TimeInterval = 0
    private var settings: Settings

    // Auto-save callback
    var onAutoSave: ((Int, Int) -> Void)?   // (wordIndex, wpm)

    // MARK: Configuration

    let minWPM = 50
    let maxWPM = 2000

    // Punctuation delay multipliers (matching Go backend)
    private let sentencePunctuationMultiplier = 1.3
    private let clausePunctuationMultiplier = 1.15

    // MARK: Init

    init(settings: Settings = Settings()) {
        self.settings = settings
        self.wpm = settings.defaultWPM
    }

    // MARK: - Document loading

    /// Load a parsed document, optionally resuming at a saved word index.
    func load(_ doc: Document, resumeAt wordIndex: Int = 0, resumeWPM: Int? = nil) {
        timer?.invalidate()
        timer = nil

        document = doc
        totalWords = doc.totalWords
        totalParagraphs = doc.totalParagraphs
        documentTitle = doc.title
        documentLoaded = true
        isPaused = true
        accumulatedElapsed = 0
        startTime = nil
        elapsedMs = 0

        wpm = resumeWPM ?? settings.defaultWPM
        jump(to: wordIndex)
    }

    // MARK: - Playback controls

    /// Start or resume reading.
    func play() {
        guard documentLoaded, isPaused else { return }
        guard wordIndex < totalWords else { return }

        isPaused = false
        startTime = Date()
        scheduleNextWord()
    }

    /// Pause reading.
    func pause() {
        guard !isPaused else { return }
        timer?.invalidate()
        timer = nil
        if let start = startTime {
            accumulatedElapsed += Date().timeIntervalSince(start)
        }
        startTime = nil
        isPaused = true
        elapsedMs = Int64(accumulatedElapsed * 1000)
    }

    /// Toggle play/pause.
    func toggle() {
        if isPaused { play() } else { pause() }
    }

    // MARK: - Speed

    /// Set WPM, clamped to [minWPM, maxWPM].
    func setWPM(_ newWPM: Int) {
        wpm = min(maxWPM, max(minWPM, newWPM))
    }

    /// Increase WPM by 50.
    func increaseWPM() { setWPM(wpm + 50) }

    /// Decrease WPM by 50.
    func decreaseWPM() { setWPM(wpm - 50) }

    // MARK: - Navigation

    /// Jump to a specific flat word index.
    func jump(to index: Int) {
        guard let doc = document else { return }
        wordIndex = max(0, min(index, doc.totalWords - 1))
        updatePublishedState()
    }

    /// Jump to the first word of a paragraph.
    func jumpToParagraph(_ index: Int) {
        guard let doc = document else { return }
        let safeIndex = max(0, min(index, doc.totalParagraphs - 1))
        jump(to: doc.wordIndex(forParagraph: safeIndex))
    }

    /// Advance to the next paragraph.
    func nextParagraph() {
        guard let doc = document else { return }
        let current = doc.paragraphIndex(forWordIndex: wordIndex)
        let next = current + 1
        guard next < doc.totalParagraphs else { return }
        jumpToParagraph(next)
    }

    /// Return to the previous paragraph (or start of current).
    func prevParagraph() {
        guard let doc = document else { return }
        let current = doc.paragraphIndex(forWordIndex: wordIndex)
        let paraStart = doc.wordIndex(forParagraph: current)
        if wordIndex > paraStart {
            // Go to start of current paragraph
            jumpToParagraph(current)
        } else if current > 0 {
            jumpToParagraph(current - 1)
        }
    }

    /// Return to start of document.
    func returnToStart() {
        jump(to: 0)
    }

    /// Jump to a percentage (0...100) of the document.
    func jumpToPercent(_ percent: Double) {
        guard totalWords > 0 else { return }
        let target = Int(Double(totalWords) * min(1, max(0, percent / 100.0)))
        jump(to: target)
    }

    // MARK: - Private scheduling

    private func scheduleNextWord() {
        guard !isPaused, let doc = document else { return }
        guard wordIndex < doc.totalWords else {
            pause()
            return
        }

        let word = doc.words[wordIndex]
        let delay = wordDelay(for: word)

        timer?.invalidate()
        timer = Timer.scheduledTimer(withTimeInterval: delay, repeats: false) { [weak self] _ in
            Task { @MainActor in
                self?.advanceWord()
            }
        }
    }

    private func advanceWord() {
        guard !isPaused, let doc = document else { return }

        wordIndex += 1

        if wordIndex >= doc.totalWords {
            // End of document
            wordIndex = doc.totalWords
            isPaused = true
            if let start = startTime {
                accumulatedElapsed += Date().timeIntervalSince(start)
                startTime = nil
            }
            elapsedMs = Int64(accumulatedElapsed * 1000)
            updatePublishedState()
            return
        }

        updatePublishedState()

        // Auto-save every N words
        let interval = settings.autoSaveInterval
        if settings.autoSave, interval > 0, wordIndex % interval == 0 {
            onAutoSave?(wordIndex, wpm)
        }

        scheduleNextWord()
    }

    /// Seconds per word, adjusted for punctuation.
    private func wordDelay(for word: String) -> TimeInterval {
        let base = 60.0 / Double(wpm)
        if word.hasSuffix(".") || word.hasSuffix("!") || word.hasSuffix("?") {
            return base * sentencePunctuationMultiplier
        } else if word.hasSuffix(",") || word.hasSuffix(";") || word.hasSuffix(":") {
            return base * clausePunctuationMultiplier
        }
        return base
    }

    // MARK: - State update

    private func updatePublishedState() {
        guard let doc = document else { return }

        let idx = wordIndex
        currentWord  = idx < doc.words.count ? doc.words[idx] : ""
        previousWord = idx > 0 ? doc.words[idx - 1] : ""

        nextWords = (1...3).compactMap { offset -> String? in
            let i = idx + offset
            return i < doc.words.count ? doc.words[i] : nil
        }

        paragraphIndex = doc.paragraphIndex(forWordIndex: idx)
        progress = totalWords > 0 ? Double(idx) / Double(totalWords) : 0

        if !isPaused, let start = startTime {
            elapsedMs = Int64((accumulatedElapsed + Date().timeIntervalSince(start)) * 1000)
        } else {
            elapsedMs = Int64(accumulatedElapsed * 1000)
        }
    }

    // MARK: - Cleanup

    func stopAndCleanUp() {
        timer?.invalidate()
        timer = nil
        isPaused = true
    }
}
