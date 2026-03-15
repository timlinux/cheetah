# Cheetah - Package Architecture

This document provides an annotated overview of all packages in the Cheetah software architecture.

---

## Package Dependency Graph

```
main.go
├── backend/
│   ├── api.go       → documents/, sessions/
│   ├── engine.go    → documents/, sessions/
│   └── server.go    → backend/api
├── frontend/
│   ├── model.go     → backend/, documents/, bubbles/filepicker
│   ├── views.go     → backend/, blockfont (external)
│   ├── client.go    → backend/
│   ├── styles.go    → lipgloss
│   └── animations.go → harmonica
├── documents/
│   ├── parser.go    → (interface definitions)
│   ├── processor.go → (text processing)
│   ├── text.go      → (standard library)
│   ├── pdf.go       → ledongthuc/pdf
│   ├── docx.go      → nguyenthenguyen/docx
│   ├── epub.go      → taylorskalyo/goreader
│   └── odt.go       → archive/zip, encoding/xml
├── (external: github.com/timlinux/blockfont)
├── sessions/
│   └── sessions.go  → (JSON persistence)
└── settings/
    └── settings.go  → (JSON persistence)
```

---

## Core Packages

### `main.go`

**Purpose:** Application entry point and CLI handling

**Responsibilities:**
- Parse command-line flags (-server, -client, -port, -wpm)
- Handle subcommands (web, version)
- Start backend server (embedded or standalone)
- Initialize frontend and run TUI

**Key Functions:**
- `main()` - Entry point
- `runServerOnly()` - Run backend as standalone server
- `runClientOnly()` - Connect TUI to existing server
- `runCombined()` - Run backend + TUI together
- `runWebCommand()` - Serve web frontend

---

### `backend/`

**Purpose:** Reading engine and REST API server

#### `backend/api.go`

**Purpose:** Interface definitions and data types

**Key Types:**
- `ReaderAPI` - Interface for reading engine operations
- `ReadingState` - Current reading state snapshot
- `DocumentInfo` - Document metadata
- `SavedSession` - Saved reading position
- `Config` - Engine configuration

#### `backend/engine.go`

**Purpose:** RSVP reading engine implementation

**Responsibilities:**
- Load and parse documents
- Timer-based word advancement
- Speed control (WPM)
- Paragraph navigation
- Position tracking
- State change notifications

**Key Methods:**
- `LoadDocument(path)` - Load document from filesystem
- `Play()` / `Pause()` / `Toggle()` - Playback control
- `SetWPM(wpm)` - Set reading speed
- `NextParagraph()` / `PrevParagraph()` - Navigation
- `GetState()` - Get current reading state
- `SavePosition()` - Persist reading position

**Concurrency:**
- Uses `sync.RWMutex` for thread-safe state access
- Ticker goroutine for word advancement
- Channel-based state subscriptions

#### `backend/server.go`

**Purpose:** HTTP REST API server

**Responsibilities:**
- Session management (create/delete)
- Document upload and loading
- Playback control endpoints
- State retrieval
- Position persistence
- Static file serving (web frontend)

**Endpoints:** See SPECIFICATION.md Section 5

---

### `documents/`

**Purpose:** Document parsing and text extraction

#### `documents/parser.go`

**Purpose:** Parser interface and factory

**Key Functions:**
- `GetParser(path)` - Return appropriate parser for file type
- `ParseFile(path)` - Parse a document file
- `SupportedFormats()` - List supported file extensions

**Key Types:**
- `Parser` - Interface for document parsers
- `Document` - Parsed document with paragraphs and words
- `Paragraph` - Collection of words with index

#### `documents/processor.go`

**Purpose:** Text processing utilities

**Responsibilities:**
- Split text into paragraphs
- Split paragraphs into words
- Clean and normalize words
- Calculate content hash

**Key Functions:**
- `ExtractParagraphs(text)` - Split text into paragraphs
- `ExtractWords(text)` - Split paragraph into words
- `CleanWord(word)` - Normalize word characters
- `GetWordAt(doc, index)` - Get word at position
- `GetParagraphStartIndex(doc, paraIndex)` - Get word index for paragraph

#### `documents/text.go`

**Purpose:** Plain text and Markdown parser

**Responsibilities:**
- Read .txt, .md, .markdown files
- Strip Markdown formatting
- Extract title from filename

#### `documents/pdf.go`

**Purpose:** PDF document parser

**Dependencies:** `github.com/ledongthuc/pdf`

**Responsibilities:**
- Extract text from PDF pages
- Handle multi-page documents

#### `documents/docx.go`

**Purpose:** Microsoft Word DOCX parser

**Dependencies:** `github.com/nguyenthenguyen/docx`

**Responsibilities:**
- Extract text content from DOCX files

#### `documents/epub.go`

**Purpose:** EPUB ebook parser

**Dependencies:** `github.com/taylorskalyo/goreader/epub`

**Responsibilities:**
- Extract text from EPUB chapters
- Handle spine ordering
- Strip HTML tags
- Extract title from metadata

#### `documents/odt.go`

**Purpose:** OpenDocument Text parser

**Dependencies:** `archive/zip`, `encoding/xml`

**Responsibilities:**
- Extract content.xml from ODT archive
- Parse XML to extract text
- Extract title from meta.xml

---

### External: `github.com/timlinux/blockfont`

**Purpose:** Block letter rendering for TUI (external dependency)

**Contents:**
- `BlockLetters` map - 6-line tall block letter definitions
- Uses Unicode block elements: █ ◢ ◣ ◤ ◥
- Supports a-z, A-Z, 0-9, punctuation

**Key Functions:**
- `RenderWord(word)` - Render word as block letters
- `RenderText(text)` - Render text with line breaks
- `GetLetterWidth(char)` - Get width of letter
- `GetTotalWidth(word)` - Get total width of word

**Key Constants:**
- `LetterHeight` - Height of block letters (6 lines)
- `LetterSpacing` - Spacing between letters

---

### `frontend/`

**Purpose:** Terminal user interface (TUI)

#### `frontend/model.go`

**Purpose:** Bubble Tea model and state machine

**States:**
- `StateFilePicker` - File selection
- `StateResumeList` - Saved sessions list
- `StateReading` - Active reading
- `StateSettings` - Settings screen

**Key Methods:**
- `Init()` - Initialize model
- `Update(msg)` - Handle messages
- `View()` - Render current state
- `handleReadingInput(msg)` - Process reading controls

#### `frontend/views.go`

**Purpose:** View rendering functions

**Key Functions:**
- `RenderReadingScreen()` - Main RSVP display
- `RenderFilePicker()` - File browser
- `RenderResumeList()` - Saved sessions
- `RenderLoading()` - Loading indicator
- `RenderError()` - Error display

**Layout Components:**
- Header (app icon, title)
- Previous word (faded)
- Current word (block letters)
- Next words (cascade)
- Progress bar (gradient)
- Status line (WPM, paragraph)
- Footer (help, branding)

#### `frontend/client.go`

**Purpose:** REST API client

**Key Methods:**
- `CreateSession()` / `DeleteSession()`
- `LoadDocument(path)` / `UploadDocument(path)`
- `GetState()` - Poll reading state
- `Play()` / `Pause()` / `Toggle()`
- `SetWPM(wpm)`
- `PrevParagraph()` / `NextParagraph()`
- `SavePosition()`
- `GetSavedSessions()` / `ResumeSession(hash)`

#### `frontend/styles.go`

**Purpose:** Lipgloss style definitions

**Color Constants:**
- `ColourWord` - Current word (white)
- `ColourPrevious` - Previous word (gray)
- `ColourPaused` - Pause indicator (yellow)
- `ColourPlaying` - Play indicator (green)

**Gradient Colors:** Red → Yellow → Green for progress

#### `frontend/animations.go`

**Purpose:** Spring-based animations

**Uses:** `charmbracelet/harmonica`

**Key Type:** `WordAnimator`
- Smooth word transitions
- Previous word fade-out
- Current word slide-in
- Next words cascade-in

---

### `sessions/`

**Purpose:** Reading position persistence

#### `sessions/sessions.go`

**Storage Location:** `~/.config/cheetah/sessions.json`

**Key Functions:**
- `Save(session)` - Save or update session
- `Load(hash)` - Load session by document hash
- `LoadAll()` - List all sessions (sorted by last access)
- `Delete(hash)` - Remove session
- `HasSession(hash)` - Check if session exists

**Data Format:**
```json
{
  "version": 1,
  "sessions": {
    "sha256-abc123...": {
      "document_path": "...",
      "document_title": "...",
      "last_position": 1523,
      "total_words": 45000,
      "last_wpm": 350,
      "last_accessed": "2026-03-11T14:30:00Z"
    }
  }
}
```

---

### `settings/`

**Purpose:** User preferences

#### `settings/settings.go`

**Storage Location:** `~/.config/cheetah/settings.json`

**Settings:**
- `DefaultWPM` - Default reading speed
- `ShowProgress` - Show progress bar
- `ShowPreviousWord` - Show previous word
- `ShowNextWords` - Show upcoming words
- `NextWordsCount` - Number of next words
- `AutoSave` - Automatic position saving
- `AutoSaveInterval` - Save frequency (words)

---

## External Dependencies

### TUI Stack (Charm)

| Package | Purpose |
|---------|---------|
| `charmbracelet/bubbletea` | TUI framework (Elm architecture) |
| `charmbracelet/lipgloss` | Styling and layout |
| `charmbracelet/harmonica` | Spring physics animations |

### Block Font Rendering

| Package | Purpose |
|---------|---------|
| `timlinux/blockfont` | Unicode block letter rendering for TUI |

### Document Parsers

| Package | Purpose |
|---------|---------|
| `ledongthuc/pdf` | PDF text extraction |
| `nguyenthenguyen/docx` | DOCX parsing |
| `taylorskalyo/goreader` | EPUB parsing |

### Standard Library

| Package | Purpose |
|---------|---------|
| `archive/zip` | ODT file extraction |
| `encoding/json` | Settings/session persistence |
| `encoding/xml` | ODT content parsing |
| `crypto/sha256` | Document content hashing |
| `net/http` | REST API server |

---

Made with ♥ by [Kartoza](https://kartoza.com) | [Donate!](https://github.com/sponsors/timlinux) | [GitHub](https://github.com/timlinux/cheetah)
