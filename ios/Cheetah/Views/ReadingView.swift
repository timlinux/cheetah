// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import SwiftUI

// MARK: - ReadingView

/// Full-screen RSVP reading experience.
/// Displays the current word in large typography with previous/next context,
/// a progress scrubber, speed controls, and paragraph navigation.
struct ReadingView: View {

    // MARK: Dependencies

    @StateObject private var engine: ReadingEngine
    @ObservedObject private var settings: Settings
    private let documentHash: String

    // MARK: Init

    init(document: Document,
         settings: Settings,
         initialPosition: Int = 0,
         initialWPM: Int? = nil) {
        let eng = ReadingEngine(settings: settings)

        // Configure auto-save callback before loading document
        eng.onAutoSave = { wordIndex, wpm in
            SessionStore.shared.save(
                documentHash: document.hash,
                title: document.title,
                position: wordIndex,
                totalWords: document.totalWords,
                wpm: wpm
            )
        }
        eng.load(document, resumeAt: initialPosition, resumeWPM: initialWPM)

        self._engine = StateObject(wrappedValue: eng)
        self._settings = ObservedObject(wrappedValue: settings)
        self.documentHash = document.hash
    }

    // MARK: UI state

    @Environment(\.dismiss) private var dismiss

    @State private var showSpeedPanel    = false
    @State private var showGoToSheet     = false
    @State private var goToPercentText   = ""
    @State private var showControlsOverlay = true
    @State private var lastTapTime: Date = .distantPast

    // Animation state
    @State private var wordScale: CGFloat = 1.0
    @State private var wordOpacity: Double = 1.0

    // MARK: Body

    var body: some View {
        ZStack {
            Color.black.ignoresSafeArea()

            VStack(spacing: 0) {
                headerBar
                Spacer(minLength: 0)
                wordDisplayArea
                Spacer(minLength: 0)
                progressSection
                controlBar
            }
        }
        .onAppear { showControlsOverlay = true }
        .onChange(of: engine.currentWord) { _ in animateWordChange() }
        .sheet(isPresented: $showSpeedPanel) { speedPanel }
        .sheet(isPresented: $showGoToSheet) { goToSheet }
        .gesture(
            TapGesture(count: 1).onEnded { _ in
                let now = Date()
                if now.timeIntervalSince(lastTapTime) < 0.3 {
                    // Double-tap: toggle play/pause
                    engine.toggle()
                } else {
                    engine.toggle()
                }
                lastTapTime = now
            }
        )
        .statusBarHidden(true)
        .preferredColorScheme(.dark)
    }

    // MARK: - Header bar

    private var headerBar: some View {
        HStack {
            // Exit button
            Button {
                saveAndExit()
            } label: {
                HStack(spacing: 4) {
                    Image(systemName: "chevron.left")
                    Text("Back")
                }
                .font(.subheadline)
                .foregroundColor(.gray)
            }

            Spacer()

            // Document title (truncated)
            VStack(spacing: 2) {
                Text(engine.documentTitle)
                    .font(.caption)
                    .fontWeight(.medium)
                    .foregroundColor(.white)
                    .lineLimit(1)
                Text("Word \(engine.wordIndex + 1) / \(engine.totalWords)")
                    .font(.caption2)
                    .foregroundColor(Color(white: 0.45))
            }

            Spacer()

            // WPM + speed button
            Button { showSpeedPanel = true } label: {
                Text("\(engine.wpm) WPM")
                    .font(.caption)
                    .fontWeight(.semibold)
                    .foregroundColor(.orange)
                    .padding(.horizontal, 10)
                    .padding(.vertical, 5)
                    .background(Color.orange.opacity(0.15))
                    .cornerRadius(8)
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 10)
    }

    // MARK: - Word display area

    private var wordDisplayArea: some View {
        VStack(spacing: 20) {
            // Previous word
            if settings.showPreviousWord {
                Text(engine.previousWord.isEmpty ? " " : engine.previousWord)
                    .font(.title3)
                    .foregroundColor(Color(white: 0.35))
                    .frame(height: 28)
            }

            // Horizontal rule
            Rectangle()
                .fill(Color(white: 0.15))
                .frame(height: 1)
                .padding(.horizontal, 40)

            // Current word – large display
            currentWordView
                .scaleEffect(wordScale)
                .opacity(wordOpacity)
                .animation(.easeInOut(duration: 0.06), value: wordScale)
                .frame(minHeight: 100)

            // Horizontal rule
            Rectangle()
                .fill(Color(white: 0.15))
                .frame(height: 1)
                .padding(.horizontal, 40)

            // Next words
            if settings.showNextWords {
                nextWordsView
            }
        }
        .padding(.horizontal, 24)
        .contentShape(Rectangle())
    }

    private var currentWordView: some View {
        Text(engine.currentWord.isEmpty ? "—" : engine.currentWord)
            .font(.system(size: dynamicFontSize(for: engine.currentWord), weight: .bold,
                          design: .rounded))
            .foregroundColor(.white)
            .minimumScaleFactor(0.4)
            .lineLimit(1)
            .multilineTextAlignment(.center)
    }

    private var nextWordsView: some View {
        VStack(spacing: 4) {
            ForEach(Array(engine.nextWords.prefix(settings.nextWordsCount).enumerated()), id: \.offset) { i, word in
                Text(word)
                    .font(.title3)
                    .foregroundColor(Color(white: 0.25 + Double(settings.nextWordsCount - i) * 0.04))
            }
        }
        .frame(height: CGFloat(settings.nextWordsCount) * 32)
    }

    // MARK: - Progress section

    private var progressSection: some View {
        VStack(spacing: 6) {
            // Progress bar (tappable to scrub)
            GeometryReader { geo in
                ZStack(alignment: .leading) {
                    RoundedRectangle(cornerRadius: 3)
                        .fill(Color(white: 0.15))
                        .frame(height: 6)

                    RoundedRectangle(cornerRadius: 3)
                        .fill(
                            LinearGradient(
                                colors: [.orange, .yellow],
                                startPoint: .leading,
                                endPoint: .trailing
                            )
                        )
                        .frame(width: geo.size.width * engine.progress, height: 6)
                }
                .contentShape(Rectangle().size(CGSize(width: geo.size.width, height: 44)))
                .gesture(
                    DragGesture(minimumDistance: 0)
                        .onChanged { value in
                            let fraction = max(0, min(1, value.location.x / geo.size.width))
                            engine.jump(to: Int(fraction * Double(engine.totalWords)))
                        }
                )
            }
            .frame(height: 6)
            .padding(.horizontal, 16)

            // Progress label + elapsed time
            HStack {
                Text(String(format: "%.1f%%", engine.progress * 100))
                    .font(.caption2)
                    .foregroundColor(.gray)
                Spacer()
                Text("¶ \(engine.paragraphIndex + 1)/\(engine.totalParagraphs)")
                    .font(.caption2)
                    .foregroundColor(.gray)
                Spacer()
                Text(formatElapsed(engine.elapsedMs))
                    .font(.caption2)
                    .foregroundColor(.gray)
            }
            .padding(.horizontal, 16)
        }
    }

    // MARK: - Control bar

    private var controlBar: some View {
        HStack(spacing: 0) {
            // Previous paragraph
            controlButton(icon: "backward.end.fill") { engine.prevParagraph() }

            // Decrease speed
            controlButton(icon: "minus.circle") { engine.decreaseWPM() }

            // Play / Pause
            Button {
                engine.toggle()
            } label: {
                Image(systemName: engine.isPaused ? "play.circle.fill" : "pause.circle.fill")
                    .font(.system(size: 48))
                    .foregroundColor(.orange)
                    .padding(.horizontal, 20)
            }
            .buttonStyle(.plain)

            // Increase speed
            controlButton(icon: "plus.circle") { engine.increaseWPM() }

            // Next paragraph
            controlButton(icon: "forward.end.fill") { engine.nextParagraph() }
        }
        .padding(.vertical, 12)
        .padding(.bottom, 8)
        .safeAreaPadding(.bottom)
    }

    private func controlButton(icon: String, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            Image(systemName: icon)
                .font(.title2)
                .foregroundColor(.gray)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 8)
        }
        .buttonStyle(.plain)
    }

    // MARK: - Speed panel sheet

    private var speedPanel: some View {
        NavigationView {
            VStack(spacing: 24) {
                // Current WPM display
                VStack(spacing: 4) {
                    Text("\(engine.wpm)")
                        .font(.system(size: 72, weight: .bold, design: .rounded))
                        .foregroundColor(wpmColor(engine.wpm))
                    Text("WPM")
                        .font(.subheadline)
                        .foregroundColor(.gray)
                    Text(Settings.speedLabel(for: engine.wpm))
                        .font(.caption)
                        .foregroundColor(wpmColor(engine.wpm).opacity(0.8))
                }

                // Slider
                Slider(
                    value: Binding(
                        get: { Double(engine.wpm) },
                        set: { engine.setWPM(Int($0)) }
                    ),
                    in: Double(engine.minWPM)...Double(engine.maxWPM),
                    step: 25
                )
                .tint(.orange)
                .padding(.horizontal, 20)

                HStack {
                    Text("\(engine.minWPM)").font(.caption).foregroundColor(.gray)
                    Spacer()
                    Text("\(engine.maxWPM)").font(.caption).foregroundColor(.gray)
                }
                .padding(.horizontal, 20)

                Divider()

                // Preset buttons 1–9
                VStack(alignment: .leading, spacing: 8) {
                    Text("Quick Presets")
                        .font(.subheadline)
                        .fontWeight(.semibold)
                        .padding(.horizontal, 20)

                    LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 3),
                              spacing: 10) {
                        ForEach(1...9, id: \.self) { key in
                            if let wpm = Settings.wpm(forPreset: key) {
                                Button {
                                    engine.setWPM(wpm)
                                } label: {
                                    VStack(spacing: 2) {
                                        Text("\(key)")
                                            .font(.caption2)
                                            .foregroundColor(.gray)
                                        Text("\(wpm)")
                                            .font(.subheadline)
                                            .fontWeight(.semibold)
                                            .foregroundColor(engine.wpm == wpm ? .orange : .white)
                                    }
                                    .frame(maxWidth: .infinity)
                                    .padding(.vertical, 10)
                                    .background(engine.wpm == wpm
                                                ? Color.orange.opacity(0.2)
                                                : Color.white.opacity(0.05))
                                    .cornerRadius(10)
                                    .overlay(
                                        RoundedRectangle(cornerRadius: 10)
                                            .stroke(engine.wpm == wpm
                                                    ? Color.orange.opacity(0.6)
                                                    : Color.clear, lineWidth: 1)
                                    )
                                }
                                .buttonStyle(.plain)
                            }
                        }
                    }
                    .padding(.horizontal, 20)
                }

                // Go-to button
                Button {
                    showSpeedPanel = false
                    showGoToSheet = true
                } label: {
                    Label("Jump to position…", systemImage: "arrow.forward.to.line")
                        .font(.subheadline)
                        .foregroundColor(.orange)
                }
                .padding(.top, 4)

                Spacer()
            }
            .padding(.top, 24)
            .background(Color(white: 0.07).ignoresSafeArea())
            .navigationTitle("Reading Speed")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") { showSpeedPanel = false }
                        .foregroundColor(.orange)
                }
            }
        }
        .presentationDetents([.medium, .large])
        .preferredColorScheme(.dark)
    }

    // MARK: - Go-to sheet

    private var goToSheet: some View {
        NavigationView {
            VStack(spacing: 20) {
                Text("Jump to a percentage of the document.")
                    .font(.subheadline)
                    .foregroundColor(.gray)

                HStack {
                    TextField("0–100", text: $goToPercentText)
                        .keyboardType(.decimalPad)
                        .textFieldStyle(.roundedBorder)
                        .frame(width: 100)
                        .colorScheme(.dark)
                    Text("%")
                        .foregroundColor(.white)
                }

                Text("Current: \(Int(engine.progress * 100))%")
                    .font(.caption)
                    .foregroundColor(.gray)

                Button("Jump") {
                    if let pct = Double(goToPercentText) {
                        engine.jumpToPercent(pct)
                    }
                    showGoToSheet = false
                    goToPercentText = ""
                }
                .buttonStyle(.borderedProminent)
                .tint(.orange)

                Spacer()
            }
            .padding(24)
            .background(Color(white: 0.07).ignoresSafeArea())
            .navigationTitle("Go to Position")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        showGoToSheet = false
                        goToPercentText = ""
                    }
                    .foregroundColor(.gray)
                }
            }
        }
        .presentationDetents([.height(260)])
        .preferredColorScheme(.dark)
    }

    // MARK: - Helpers

    /// Dynamic font size based on word length.
    private func dynamicFontSize(for word: String) -> CGFloat {
        switch word.count {
        case 0...4:  return 72
        case 5...7:  return 60
        case 8...10: return 50
        case 11...13: return 42
        default:     return 34
        }
    }

    /// Orange tint for the WPM colour indicator.
    private func wpmColor(_ wpm: Int) -> Color {
        switch wpm {
        case ..<200:  return .blue
        case ..<350:  return .green
        case ..<500:  return .yellow
        case ..<700:  return .orange
        default:      return .red
        }
    }

    private func formatElapsed(_ ms: Int64) -> String {
        let totalSeconds = Int(ms / 1000)
        let minutes = totalSeconds / 60
        let seconds = totalSeconds % 60
        return String(format: "%d:%02d", minutes, seconds)
    }

    private func animateWordChange() {
        wordScale   = 1.05
        wordOpacity = 0.8
        withAnimation(.easeOut(duration: 0.1)) {
            wordScale   = 1.0
            wordOpacity = 1.0
        }
    }

    private func saveAndExit() {
        engine.pause()
        if !documentHash.isEmpty {
            SessionStore.shared.save(
                documentHash: documentHash,
                title: engine.documentTitle,
                position: engine.wordIndex,
                totalWords: engine.totalWords,
                wpm: engine.wpm
            )
        }
        engine.stopAndCleanUp()
        dismiss()
    }
}
