// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import XCTest
@testable import Cheetah

// MARK: - DocumentParserTests

final class DocumentParserTests: XCTestCase {

    // MARK: splitIntoParagraphs

    func testSplitIntoParagraphsOnDoubleNewline() {
        let text = "Hello world\n\nSecond paragraph here.\n\nThird."
        let paragraphs = DocumentParser.splitIntoParagraphs(text)
        XCTAssertEqual(paragraphs.count, 3)
        XCTAssertEqual(paragraphs[0], "Hello world")
        XCTAssertEqual(paragraphs[1], "Second paragraph here.")
        XCTAssertEqual(paragraphs[2], "Third.")
    }

    func testSplitIntoParagraphsIgnoresEmptyBlocks() {
        let text = "\n\nFirst paragraph.\n\n\n\nSecond paragraph.\n\n"
        let paragraphs = DocumentParser.splitIntoParagraphs(text)
        XCTAssertEqual(paragraphs.count, 2)
    }

    func testSplitIntoParagraphsSingleParagraph() {
        let text = "Just one paragraph with no blank lines."
        let paragraphs = DocumentParser.splitIntoParagraphs(text)
        XCTAssertEqual(paragraphs.count, 1)
        XCTAssertEqual(paragraphs[0], "Just one paragraph with no blank lines.")
    }

    func testSplitNormalisesWindowsLineEndings() {
        let text = "First\r\n\r\nSecond"
        let paragraphs = DocumentParser.splitIntoParagraphs(text)
        XCTAssertEqual(paragraphs.count, 2)
        XCTAssertEqual(paragraphs[0], "First")
        XCTAssertEqual(paragraphs[1], "Second")
    }

    // MARK: tokenize

    func testTokenizeBasic() {
        let words = DocumentParser.tokenize("Hello world foo bar")
        XCTAssertEqual(words, ["Hello", "world", "foo", "bar"])
    }

    func testTokenizeFiltersEmptyTokens() {
        let words = DocumentParser.tokenize("  hello   world  ")
        XCTAssertEqual(words, ["hello", "world"])
    }

    func testTokenizePreservesPunctuation() {
        let words = DocumentParser.tokenize("Hello, world.")
        XCTAssertEqual(words, ["Hello,", "world."])
    }

    func testTokenizeEmptyString() {
        let words = DocumentParser.tokenize("")
        XCTAssertTrue(words.isEmpty)
    }
}

// MARK: - DocumentBuildTests

final class DocumentBuildTests: XCTestCase {

    func testBuildDocumentWordCount() {
        let doc = Document.build(
            title: "Test",
            path: nil,
            paragraphWords: [["hello", "world"], ["foo", "bar", "baz"]]
        )
        XCTAssertEqual(doc.totalWords, 5)
        XCTAssertEqual(doc.totalParagraphs, 2)
    }

    func testBuildDocumentParagraphIndices() {
        let doc = Document.build(
            title: "Test",
            path: nil,
            paragraphWords: [["a", "b", "c"], ["d", "e"]]
        )
        // Paragraph 0 starts at word 0, paragraph 1 starts at word 3
        XCTAssertEqual(doc.paragraphs[0].startWordIndex, 0)
        XCTAssertEqual(doc.paragraphs[1].startWordIndex, 3)
    }

    func testBuildDocumentHash() {
        let doc1 = Document.build(title: "T", path: nil,
                                  paragraphWords: [["hello", "world"]])
        let doc2 = Document.build(title: "T", path: nil,
                                  paragraphWords: [["hello", "world"]])
        let doc3 = Document.build(title: "T", path: nil,
                                  paragraphWords: [["different", "content"]])
        XCTAssertEqual(doc1.hash, doc2.hash)
        XCTAssertNotEqual(doc1.hash, doc3.hash)
    }

    func testBuildDocumentHashNotEmpty() {
        let doc = Document.build(title: "T", path: nil,
                                 paragraphWords: [["hello"]])
        XCTAssertFalse(doc.hash.isEmpty)
    }

    func testParagraphIndexForWordIndex() {
        let doc = Document.build(
            title: "T",
            path: nil,
            paragraphWords: [["a", "b"], ["c", "d", "e"], ["f"]]
        )
        XCTAssertEqual(doc.paragraphIndex(forWordIndex: 0), 0)
        XCTAssertEqual(doc.paragraphIndex(forWordIndex: 1), 0)
        XCTAssertEqual(doc.paragraphIndex(forWordIndex: 2), 1)
        XCTAssertEqual(doc.paragraphIndex(forWordIndex: 4), 1)
        XCTAssertEqual(doc.paragraphIndex(forWordIndex: 5), 2)
    }

    func testWordIndexForParagraph() {
        let doc = Document.build(
            title: "T",
            path: nil,
            paragraphWords: [["a", "b"], ["c", "d", "e"], ["f"]]
        )
        XCTAssertEqual(doc.wordIndex(forParagraph: 0), 0)
        XCTAssertEqual(doc.wordIndex(forParagraph: 1), 2)
        XCTAssertEqual(doc.wordIndex(forParagraph: 2), 5)
    }

    func testBuildDocumentSkipsEmptyParagraphs() {
        let doc = Document.build(
            title: "T",
            path: nil,
            paragraphWords: [["a"], [], ["b"]]
        )
        XCTAssertEqual(doc.totalParagraphs, 2)
        XCTAssertEqual(doc.totalWords, 2)
    }
}

// MARK: - ReadingEngineTests

@MainActor
final class ReadingEngineTests: XCTestCase {

    private var engine: ReadingEngine!
    private var settings: Settings!
    private var testDoc: Document!

    override func setUp() async throws {
        settings = Settings()
        settings.defaultWPM = 300
        settings.autoSave = false

        engine = ReadingEngine(settings: settings)

        testDoc = Document.build(
            title: "Test Document",
            path: nil,
            paragraphWords: [
                ["Hello", "world,", "this", "is", "a", "test."],
                ["Second", "paragraph", "here."],
                ["Third."]
            ]
        )
        engine.load(testDoc)
    }

    func testInitialState() {
        XCTAssertTrue(engine.isPaused)
        XCTAssertEqual(engine.wordIndex, 0)
        XCTAssertEqual(engine.totalWords, 10)
        XCTAssertEqual(engine.totalParagraphs, 3)
        XCTAssertTrue(engine.documentLoaded)
        XCTAssertEqual(engine.documentTitle, "Test Document")
    }

    func testInitialCurrentWord() {
        XCTAssertEqual(engine.currentWord, "Hello")
    }

    func testJumpToWord() {
        engine.jump(to: 3)
        XCTAssertEqual(engine.wordIndex, 3)
        XCTAssertEqual(engine.currentWord, "is")
    }

    func testJumpToWordClampsLow() {
        engine.jump(to: -5)
        XCTAssertEqual(engine.wordIndex, 0)
    }

    func testJumpToWordClampsHigh() {
        engine.jump(to: 9999)
        XCTAssertEqual(engine.wordIndex, testDoc.totalWords - 1)
    }

    func testJumpToParagraph() {
        engine.jumpToParagraph(1)
        XCTAssertEqual(engine.wordIndex, 6)  // Second paragraph starts at word 6
        XCTAssertEqual(engine.currentWord, "Second")
    }

    func testNextParagraph() {
        engine.jump(to: 0)
        engine.nextParagraph()
        XCTAssertEqual(engine.paragraphIndex, 1)
    }

    func testPrevParagraphFromMidParagraph() {
        engine.jump(to: 3)  // Middle of paragraph 0
        engine.prevParagraph()
        XCTAssertEqual(engine.wordIndex, 0)  // Goes to start of paragraph 0
    }

    func testPrevParagraphFromParagraphStart() {
        engine.jumpToParagraph(2)
        engine.prevParagraph()
        XCTAssertEqual(engine.paragraphIndex, 1)
    }

    func testSetWPMClampsLow() {
        engine.setWPM(0)
        XCTAssertEqual(engine.wpm, engine.minWPM)
    }

    func testSetWPMClampsHigh() {
        engine.setWPM(99999)
        XCTAssertEqual(engine.wpm, engine.maxWPM)
    }

    func testIncreaseWPM() {
        engine.setWPM(300)
        engine.increaseWPM()
        XCTAssertEqual(engine.wpm, 350)
    }

    func testDecreaseWPM() {
        engine.setWPM(300)
        engine.decreaseWPM()
        XCTAssertEqual(engine.wpm, 250)
    }

    func testProgressAt0() {
        engine.jump(to: 0)
        XCTAssertEqual(engine.progress, 0.0, accuracy: 0.001)
    }

    func testJumpToPercent() {
        engine.jumpToPercent(50)
        let expected = Int(Double(testDoc.totalWords) * 0.5)
        XCTAssertEqual(engine.wordIndex, expected)
    }

    func testJumpToPercentClamps() {
        engine.jumpToPercent(-10)
        XCTAssertEqual(engine.wordIndex, 0)
        engine.jumpToPercent(110)
        XCTAssertEqual(engine.wordIndex, testDoc.totalWords - 1)
    }

    func testReturnToStart() {
        engine.jump(to: 5)
        engine.returnToStart()
        XCTAssertEqual(engine.wordIndex, 0)
    }

    func testNextWords() {
        engine.jump(to: 0)
        XCTAssertEqual(engine.nextWords.count, 3)
        XCTAssertEqual(engine.nextWords[0], "world,")
        XCTAssertEqual(engine.nextWords[1], "this")
        XCTAssertEqual(engine.nextWords[2], "is")
    }

    func testPreviousWord() {
        engine.jump(to: 0)
        XCTAssertEqual(engine.previousWord, "")
        engine.jump(to: 1)
        XCTAssertEqual(engine.previousWord, "Hello")
    }

    func testPauseDoesNothingWhenAlreadyPaused() {
        XCTAssertTrue(engine.isPaused)
        engine.pause()  // Should not crash
        XCTAssertTrue(engine.isPaused)
    }

    func testLoadDocumentResetsState() {
        engine.jump(to: 5)
        let newDoc = Document.build(title: "New", path: nil,
                                    paragraphWords: [["one", "two"]])
        engine.load(newDoc)
        XCTAssertEqual(engine.wordIndex, 0)
        XCTAssertEqual(engine.totalWords, 2)
        XCTAssertEqual(engine.documentTitle, "New")
    }

    func testResumeAtPosition() {
        engine.load(testDoc, resumeAt: 5, resumeWPM: 400)
        XCTAssertEqual(engine.wordIndex, 5)
        XCTAssertEqual(engine.wpm, 400)
    }
}

// MARK: - SettingsTests

final class SettingsTests: XCTestCase {

    func testSpeedPresets() {
        XCTAssertEqual(Settings.wpm(forPreset: 1), 200)
        XCTAssertEqual(Settings.wpm(forPreset: 5), 600)
        XCTAssertEqual(Settings.wpm(forPreset: 9), 1000)
    }

    func testSpeedPresetOutOfRange() {
        XCTAssertNil(Settings.wpm(forPreset: 0))
        XCTAssertNil(Settings.wpm(forPreset: 10))
    }

    func testSpeedLabel() {
        XCTAssertEqual(Settings.speedLabel(for: 150), "Relaxed")
        XCTAssertEqual(Settings.speedLabel(for: 300), "Normal")
        XCTAssertEqual(Settings.speedLabel(for: 450), "Fast")
        XCTAssertEqual(Settings.speedLabel(for: 600), "Very Fast")
        XCTAssertEqual(Settings.speedLabel(for: 800), "Speed Demon")
    }
}

// MARK: - SessionStoreTests

final class SessionStoreTests: XCTestCase {

    func testSaveAndRetrieve() {
        let store = SessionStore.shared
        let hash = "test-hash-\(UUID())"
        store.save(documentHash: hash, title: "Test Book",
                   position: 100, totalWords: 1000, wpm: 350)

        let session = store.session(for: hash)
        XCTAssertNotNil(session)
        XCTAssertEqual(session?.documentTitle, "Test Book")
        XCTAssertEqual(session?.lastPosition, 100)
        XCTAssertEqual(session?.lastWPM, 350)
        XCTAssertEqual(session?.totalWords, 1000)

        // Clean up
        store.delete(hash: hash)
    }

    func testDelete() {
        let store = SessionStore.shared
        let hash = "delete-test-\(UUID())"
        store.save(documentHash: hash, title: "Delete Test",
                   position: 0, totalWords: 100, wpm: 300)
        store.delete(hash: hash)
        XCTAssertNil(store.session(for: hash))
    }

    func testProgress() {
        let store = SessionStore.shared
        let hash = "progress-test-\(UUID())"
        store.save(documentHash: hash, title: "Progress Test",
                   position: 250, totalWords: 1000, wpm: 300)

        let session = store.session(for: hash)
        XCTAssertEqual(session?.progress ?? 0, 0.25, accuracy: 0.001)

        store.delete(hash: hash)
    }

    func testRecentSessionsLimit() {
        let store = SessionStore.shared
        var hashes: [String] = []
        for i in 0..<8 {
            let hash = "limit-test-\(i)-\(UUID())"
            hashes.append(hash)
            store.save(documentHash: hash, title: "Book \(i)",
                       position: 0, totalWords: 100, wpm: 300)
        }

        let recent = store.recentSessions(limit: 5)
        XCTAssertLessThanOrEqual(recent.count, 5)

        // Clean up
        hashes.forEach { store.delete(hash: $0) }
    }
}
