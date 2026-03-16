// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import SwiftUI

// MARK: - SettingsView

/// User preferences screen.
struct SettingsView: View {

    @ObservedObject var settings: Settings
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationView {
            Form {
                // MARK: Reading speed
                Section {
                    VStack(alignment: .leading, spacing: 8) {
                        HStack {
                            Text("Default Speed")
                            Spacer()
                            Text("\(settings.defaultWPM) WPM")
                                .foregroundColor(.orange)
                                .fontWeight(.semibold)
                        }
                        Slider(
                            value: Binding(
                                get: { Double(settings.defaultWPM) },
                                set: { settings.defaultWPM = Int($0) }
                            ),
                            in: 50...2000,
                            step: 50
                        )
                        .tint(.orange)

                        Text(Settings.speedLabel(for: settings.defaultWPM))
                            .font(.caption)
                            .foregroundColor(.gray)
                    }
                    .padding(.vertical, 4)
                } header: {
                    Text("Reading Speed")
                } footer: {
                    Text("This is the WPM used when opening a new document. Speed can always be adjusted during reading.")
                }

                // MARK: Display
                Section {
                    Toggle("Show Previous Word", isOn: $settings.showPreviousWord)
                    Toggle("Show Upcoming Words", isOn: $settings.showNextWords)

                    if settings.showNextWords {
                        Stepper(
                            "Upcoming Words: \(settings.nextWordsCount)",
                            value: $settings.nextWordsCount,
                            in: 1...5
                        )
                    }
                } header: {
                    Text("Display")
                } footer: {
                    Text("Show context words above and below the current word to help maintain comprehension.")
                }

                // MARK: Auto-save
                Section {
                    Toggle("Auto-Save Position", isOn: $settings.autoSave)

                    if settings.autoSave {
                        Stepper(
                            "Save Every \(settings.autoSaveInterval) Words",
                            value: $settings.autoSaveInterval,
                            in: 10...500,
                            step: 10
                        )
                    }
                } header: {
                    Text("Auto-Save")
                } footer: {
                    Text("Automatically saves your reading position so you can resume later.")
                }

                // MARK: Speed presets reference
                Section {
                    ForEach(1...9, id: \.self) { key in
                        if let wpm = Settings.wpm(forPreset: key) {
                            HStack {
                                Text("Key \(key)")
                                    .foregroundColor(.gray)
                                Spacer()
                                Text("\(wpm) WPM")
                                    .fontWeight(.medium)
                                Text("·")
                                    .foregroundColor(.gray)
                                Text(Settings.speedLabel(for: wpm))
                                    .font(.caption)
                                    .foregroundColor(.gray)
                            }
                        }
                    }
                } header: {
                    Text("Speed Presets")
                } footer: {
                    Text("Tap a preset button (1–9) in the speed panel during reading to instantly set your speed.")
                }

                // MARK: About
                Section {
                    HStack {
                        Text("Version")
                        Spacer()
                        Text("1.0.0")
                            .foregroundColor(.gray)
                    }
                    Link(destination: URL(string: "https://github.com/timlinux/cheetah")!) {
                        HStack {
                            Text("GitHub")
                            Spacer()
                            Image(systemName: "arrow.up.right")
                                .font(.caption)
                                .foregroundColor(.gray)
                        }
                    }
                    Link(destination: URL(string: "https://kartoza.com")!) {
                        HStack {
                            Text("Made with ❤️ by Kartoza")
                            Spacer()
                            Image(systemName: "arrow.up.right")
                                .font(.caption)
                                .foregroundColor(.gray)
                        }
                    }
                } header: {
                    Text("About")
                }
            }
            .navigationTitle("Settings")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") { dismiss() }
                        .foregroundColor(.orange)
                }
            }
        }
        .preferredColorScheme(.dark)
    }
}

// MARK: - Preview

#if DEBUG
struct SettingsView_Previews: PreviewProvider {
    static var previews: some View {
        SettingsView(settings: Settings())
    }
}
#endif
