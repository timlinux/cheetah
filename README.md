# 🐆 Cheetah - RSVP Speed Reading

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://go.dev/)

**Cheetah** is a Rapid Serial Visual Presentation (RSVP) speed reading application that displays words one at a time, allowing you to read at speeds up to 1000+ WPM.

![Cheetah Demo](demo/cheetah-demo.gif)

Note that my screen cast recorder (asciinema) does not accurately record the
font rendering or playback framerate, but hopeully the little demo above gives
you a sense of how it works.

## Features

- **RSVP Reading** - Words displayed one at a time in a fixed location
- **Multi-Format Support** - PDF, DOCX, EPUB, ODT, TXT, Markdown
- **Speed Control** - 50-2000 WPM with instant adjustment
- **Smart Pausing** - Extra time after punctuation for comprehension
- **Position Memory** - Resume reading where you left off
- **Beautiful TUI** - Large block letters with smooth animations
- **Web Interface** - Browser-based reading (coming soon)

## Why RSVP?

Traditional reading involves eye movements (saccades) that slow you down. RSVP eliminates this by presenting words in a fixed location, allowing your brain to process text faster. Research shows 2-3x faster reading with similar comprehension.

## Installation

### Using Nix (Recommended)

```bash
# Run directly
nix run github:timlinux/cheetah

# Or install to profile
nix profile install github:timlinux/cheetah
```

### From Source

```bash
git clone https://github.com/timlinux/cheetah.git
cd cheetah
nix develop  # Or ensure Go 1.25+ is installed
make build
./cheetah
```

### Pre-built Binaries

Download from [Releases](https://github.com/timlinux/cheetah/releases).

## Usage

### Open File Picker

```bash
cheetah
```

### Open Document Directly

```bash
cheetah /path/to/document.pdf
cheetah ~/Books/novel.epub
```

### Set Initial Speed

```bash
cheetah -wpm 400 book.txt
```

## Controls

| Key | Action |
|-----|--------|
| **Space** | Pause/resume reading |
| **j** / **←** | Decrease speed 50 WPM |
| **k** / **→** | Increase speed 50 WPM |
| **h** / **↑** | Previous paragraph |
| **l** / **↓** | Next paragraph |
| **1-9** | Speed presets (200-1000 WPM) |
| **s** | Save position |
| **Escape** | Return to file picker |
| **q** | Quit |

## Supported Formats

| Format | Extension | Parser |
|--------|-----------|--------|
| Plain Text | .txt | Standard library |
| Markdown | .md, .markdown | Standard library |
| PDF | .pdf | ledongthuc/pdf |
| Word | .docx | nguyenthenguyen/docx |
| EPUB | .epub | taylorskalyo/goreader |
| OpenDocument | .odt | archive/zip |

## Configuration

Settings are stored in `~/.config/cheetah/settings.json`:

```json
{
  "default_wpm": 300,
  "show_progress": true,
  "show_previous_word": true,
  "show_next_words": true,
  "next_words_count": 3,
  "auto_save": true,
  "auto_save_interval": 50
}
```

## Development

### Prerequisites

- Go 1.25+
- Node.js 20+ (for web frontend)
- Hugo (for documentation)

### Build

```bash
nix develop
make build
make test
```

### Run Development Server

```bash
# TUI
make run

# Web
make web-dev
```

### Documentation

```bash
make docs-dev
# Visit http://localhost:1313
```

## Architecture

```
cheetah/
├── main.go           # Entry point
├── backend/          # Reading engine + REST API
├── frontend/         # TUI (Bubble Tea)
├── documents/        # Document parsers
├── (uses: github.com/timlinux/blockfont)  # Block letter rendering
├── sessions/         # Position persistence
├── settings/         # User preferences
└── web/              # Web frontend (React)
```

See [PACKAGES.md](PACKAGES.md) for detailed architecture.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

---

Made with ♥ by [Kartoza](https://kartoza.com) | [Donate!](https://github.com/sponsors/timlinux) | [GitHub](https://github.com/timlinux/cheetah)
