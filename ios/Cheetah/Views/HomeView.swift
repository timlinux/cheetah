// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import SwiftUI
import UniformTypeIdentifiers

// MARK: - HomeView

/// Landing screen: drop-zone / file picker, recent sessions, feature overview.
struct HomeView: View {

    // MARK: Dependencies

    @StateObject private var settings = Settings()
    @StateObject private var sessionStore = SessionStore.shared

    // MARK: Navigation state

    @State private var showDocumentPicker = false
    @State private var showSettings       = false
    @State private var isLoading          = false
    @State private var loadError: String? = nil
    @State private var loadedDocument: Document? = nil
    @State private var resumeSession: ReadingSession? = nil

    // MARK: Body

    var body: some View {
        NavigationStack {
            ZStack {
                // Background
                Color.black.ignoresSafeArea()

                ScrollView {
                    VStack(spacing: 32) {
                        heroSection
                        if let err = loadError { errorBanner(err) }
                        dropZoneButton
                        recentSessionsSection
                        featuresSection
                        footerSection
                    }
                    .padding(.horizontal, 20)
                    .padding(.vertical, 24)
                }
            }
            .navigationTitle("")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .principal) {
                    HStack(spacing: 8) {
                        Text("🐆").font(.title2)
                        Text("Cheetah")
                            .font(.headline)
                            .foregroundColor(.white)
                    }
                }
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button { showSettings = true } label: {
                        Image(systemName: "gear")
                            .foregroundColor(.gray)
                    }
                }
            }
        }
        // Document picker
        .fileImporter(
            isPresented: $showDocumentPicker,
            allowedContentTypes: allowedTypes,
            allowsMultipleSelection: false
        ) { result in
            handleImport(result)
        }
        // Settings sheet
        .sheet(isPresented: $showSettings) {
            SettingsView(settings: settings)
        }
        // Navigate to reading when document loaded
        .fullScreenCover(item: $loadedDocument) { doc in
            ReadingView(
                document: doc,
                settings: settings,
                initialPosition: resumeSession?.lastPosition ?? 0,
                initialWPM: resumeSession?.lastWPM ?? settings.defaultWPM
            )
        }
        // Loading overlay
        .overlay {
            if isLoading {
                ZStack {
                    Color.black.opacity(0.7).ignoresSafeArea()
                    VStack(spacing: 16) {
                        ProgressView()
                            .progressViewStyle(.circular)
                            .scaleEffect(1.5)
                            .tint(.orange)
                        Text("Parsing document…")
                            .foregroundColor(.gray)
                    }
                }
            }
        }
    }

    // MARK: - Hero section

    private var heroSection: some View {
        VStack(spacing: 12) {
            Text("Read at 1000+ WPM")
                .font(.system(size: 34, weight: .bold))
                .foregroundStyle(
                    LinearGradient(
                        colors: [.orange, .yellow],
                        startPoint: .leading,
                        endPoint: .trailing
                    )
                )
                .multilineTextAlignment(.center)

            Text("RSVP displays words one at a time, eliminating eye movement\nand unlocking your true reading speed.")
                .font(.subheadline)
                .foregroundColor(.gray)
                .multilineTextAlignment(.center)
        }
    }

    // MARK: - Error banner

    private func errorBanner(_ message: String) -> some View {
        HStack {
            Image(systemName: "exclamationmark.triangle.fill")
                .foregroundColor(.red)
            Text(message)
                .foregroundColor(.red)
                .font(.subheadline)
            Spacer()
            Button { loadError = nil } label: {
                Image(systemName: "xmark")
                    .foregroundColor(.gray)
            }
        }
        .padding()
        .background(Color.red.opacity(0.15))
        .cornerRadius(12)
    }

    // MARK: - Drop zone / file picker button

    private var dropZoneButton: some View {
        Button { showDocumentPicker = true } label: {
            VStack(spacing: 12) {
                Image(systemName: "doc.badge.plus")
                    .font(.system(size: 48))
                    .foregroundColor(.orange)

                Text("Tap to Open a Document")
                    .font(.title3)
                    .fontWeight(.semibold)
                    .foregroundColor(.white)

                Text("or use the Files app to share any document here")
                    .font(.caption)
                    .foregroundColor(.gray)

                // Supported format badges
                HStack(spacing: 6) {
                    ForEach(["PDF", "DOCX", "EPUB", "ODT", "TXT", "MD"], id: \.self) { fmt in
                        Text(fmt)
                            .font(.caption2)
                            .fontWeight(.semibold)
                            .padding(.horizontal, 8)
                            .padding(.vertical, 4)
                            .background(Color.white.opacity(0.08))
                            .foregroundColor(.gray)
                            .cornerRadius(6)
                    }
                }
            }
            .frame(maxWidth: .infinity)
            .padding(28)
            .overlay(
                RoundedRectangle(cornerRadius: 20)
                    .stroke(Color.orange.opacity(0.4), style: StrokeStyle(lineWidth: 2, dash: [8]))
            )
        }
        .buttonStyle(.plain)
    }

    // MARK: - Recent sessions

    @ViewBuilder
    private var recentSessionsSection: some View {
        let recent = sessionStore.recentSessions(limit: 5)
        if !recent.isEmpty {
            VStack(alignment: .leading, spacing: 12) {
                Text("Continue Reading")
                    .font(.headline)
                    .foregroundColor(.white)

                ForEach(recent) { session in
                    SessionCard(session: session) {
                        // Re-open using the files app – show picker with context
                        resumeSession = session
                        showDocumentPicker = true
                    } onDelete: {
                        sessionStore.delete(hash: session.documentHash)
                    }
                }

                Text("Tip: Reopen the same document to resume where you left off.")
                    .font(.caption2)
                    .foregroundColor(Color(white: 0.4))
                    .multilineTextAlignment(.center)
                    .frame(maxWidth: .infinity)
            }
        }
    }

    // MARK: - Features

    private var featuresSection: some View {
        VStack(spacing: 12) {
            HStack(spacing: 12) {
                FeatureCard(icon: "bolt.fill", title: "Lightning Fast",
                            description: "2–5× faster than traditional reading.")
                FeatureCard(icon: "lock.fill", title: "Privacy First",
                            description: "Documents stay on your device.")
            }
            HStack(spacing: 12) {
                FeatureCard(icon: "arrow.clockwise", title: "Auto-Save",
                            description: "Position saved after every 50 words.")
                FeatureCard(icon: "slider.horizontal.3", title: "Speed Control",
                            description: "50–2000 WPM with one-touch presets.")
            }
        }
    }

    // MARK: - Footer

    private var footerSection: some View {
        VStack(spacing: 6) {
            HStack(spacing: 8) {
                Text("Made with ❤️ by")
                    .foregroundColor(.gray)
                Link("Kartoza", destination: URL(string: "https://kartoza.com")!)
                    .foregroundColor(.orange)
                Text("·")
                    .foregroundColor(Color(white: 0.3))
                Link("GitHub", destination: URL(string: "https://github.com/timlinux/cheetah")!)
                    .foregroundColor(.gray)
            }
            .font(.footnote)

            HStack(spacing: 12) {
                Text("SPACE: pause").font(.caption2).foregroundColor(Color(white: 0.35))
                Text("j/k: speed").font(.caption2).foregroundColor(Color(white: 0.35))
                Text("1–9: presets").font(.caption2).foregroundColor(Color(white: 0.35))
            }
        }
        .padding(.top, 8)
    }

    // MARK: - File import handling

    private var allowedTypes: [UTType] {
        [
            .plainText,
            .pdf,
            .epub,
            UTType(filenameExtension: "docx") ?? .data,
            UTType(filenameExtension: "odt") ?? .data,
            UTType(filenameExtension: "md") ?? .plainText,
            UTType(filenameExtension: "markdown") ?? .plainText,
        ]
    }

    private func handleImport(_ result: Result<[URL], Error>) {
        switch result {
        case .failure(let error):
            loadError = error.localizedDescription
        case .success(let urls):
            guard let url = urls.first else { return }
            loadDocument(url: url)
        }
    }

    private func loadDocument(url: URL) {
        isLoading = true
        loadError = nil

        // Access security-scoped resource
        let secured = url.startAccessingSecurityScopedResource()

        Task.detached(priority: .userInitiated) {
            do {
                let doc = try DocumentParser.parse(url: url)

                // Check for saved session
                let savedSession = SessionStore.shared.session(for: doc.hash)

                await MainActor.run {
                    if secured { url.stopAccessingSecurityScopedResource() }
                    isLoading = false
                    resumeSession = savedSession
                    loadedDocument = doc
                }
            } catch {
                await MainActor.run {
                    if secured { url.stopAccessingSecurityScopedResource() }
                    isLoading = false
                    loadError = error.localizedDescription
                }
            }
        }
    }
}

// MARK: - SessionCard

private struct SessionCard: View {
    let session: ReadingSession
    let onOpen: () -> Void
    let onDelete: () -> Void

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                Text(session.documentTitle)
                    .font(.subheadline)
                    .fontWeight(.semibold)
                    .foregroundColor(.white)
                    .lineLimit(1)

                HStack(spacing: 8) {
                    Text("\(session.totalWords.formatted()) words")
                    Text("·")
                    Text("\(Int(session.progress * 100))% read")
                }
                .font(.caption)
                .foregroundColor(.gray)

                HStack(spacing: 6) {
                    Text("\(session.lastWPM) WPM")
                        .font(.caption2)
                        .padding(.horizontal, 6).padding(.vertical, 2)
                        .background(Color.orange.opacity(0.2))
                        .foregroundColor(.orange)
                        .cornerRadius(4)

                    Text(session.lastAccessed.formatted(date: .abbreviated, time: .omitted))
                        .font(.caption2)
                        .foregroundColor(Color(white: 0.4))
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)

            Button(action: onOpen) {
                Image(systemName: "play.circle.fill")
                    .font(.title2)
                    .foregroundColor(.orange)
            }
            .padding(.leading, 4)

            Button(action: onDelete) {
                Image(systemName: "xmark.circle")
                    .font(.body)
                    .foregroundColor(Color(white: 0.4))
            }
        }
        .padding(14)
        .background(Color.white.opacity(0.05))
        .cornerRadius(14)
    }
}

// MARK: - FeatureCard

private struct FeatureCard: View {
    let icon: String
    let title: String
    let description: String

    var body: some View {
        VStack(spacing: 8) {
            Image(systemName: icon)
                .font(.title2)
                .foregroundColor(.orange)
            Text(title)
                .font(.subheadline)
                .fontWeight(.semibold)
                .foregroundColor(.white)
            Text(description)
                .font(.caption)
                .foregroundColor(.gray)
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity)
        .padding(16)
        .background(Color.white.opacity(0.05))
        .cornerRadius(14)
    }
}
