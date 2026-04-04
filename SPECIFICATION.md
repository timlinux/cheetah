# Cheetah - RSVP Speed Reading Application

## Technical Specification

Version: 0.2.0

---

## 1. Overview

Cheetah is a Rapid Serial Visual Presentation (RSVP) speed reading application that displays words one at a time from any document format, allowing users to read at speeds up to 1000+ WPM.

### 1.1 RSVP Technique

RSVP is a speed reading technique where:
- Words are displayed one at a time in a fixed location
- Eliminates eye movement (saccades) that slow down traditional reading
- Breaks subvocalization habit at higher speeds
- Research shows 2-3x faster reading with similar comprehension

### 1.2 Target Platforms

- **TUI**: Cross-platform terminal interface (Linux, macOS, Windows)
- **Web**: Browser-based interface (Chrome, Firefox, Safari, Edge)

---

## 2. Architecture

```
┌────────────────────────────────────────────────────────────────────┐
│                         FRONTEND LAYER                              │
├──────────────────────────┬─────────────────────────────────────────┤
│ Terminal UI (TUI)        │ Web UI                                  │
│ (Bubble Tea + Lipgloss)  │ (React + Chakra UI + Framer Motion)    │
│ - File picker            │ - Drag-drop document upload             │
│ - Large block word       │ - Large animated word display           │
│ - Speed/pause controls   │ - Client-side document parsing          │
│                          │ - localStorage persistence              │
├──────────────────────────┴─────────────────────────────────────────┤
│                    REST API (JSON over HTTP)                        │
├────────────────────────────────────────────────────────────────────┤
│                    GO BACKEND SERVER                                │
├────────────────────────────────────────────────────────────────────┤
│  Reading Engine │ Document Parser │ Session Manager │ Persistence  │
└────────────────────────────────────────────────────────────────────┘
```

### 2.1 Component Responsibilities

| Component | Responsibility |
|-----------|---------------|
| Reading Engine | Timer-based word advancement, speed control |
| Document Parser | Extract text from PDF, DOCX, EPUB, ODT, TXT, MD |
| Session Manager | Track reading positions per document |
| REST API | Expose engine functionality via HTTP |
| TUI Frontend | Terminal-based user interface |
| Web Frontend | Browser-based user interface |

---

## 3. User Stories

### 3.1 Document Loading

**US-001: Load document from file picker**
> As a user, I want to browse and select a document to read using a file picker, so I can choose any document on my system.

**US-002: Load document from command line**
> As a user, I want to open a document directly from the command line, so I can start reading immediately.

**US-003: Drag-and-drop document (Web)**
> As a web user, I want to drag and drop a document onto the interface, so I can quickly start reading.

**US-004: Resume reading from resume list**
> As a user, I want to resume reading a previously opened document from where I left off using the resume list.

**US-005: Auto-resume from file picker**
> As a TUI user, when I open a document from the file picker that I have previously read, I want to automatically resume from my last position and speed setting.

### 3.2 Reading Controls

**US-010: Play/Pause reading**
> As a user, I want to pause and resume reading with the space bar, so I can take breaks or re-read content.

**US-011: Adjust reading speed**
> As a user, I want to increase or decrease the reading speed, so I can find my optimal reading pace.

**US-012: Speed presets**
> As a user, I want to quickly jump to preset speeds (200-1000 WPM), so I can easily switch between speeds.

**US-013: Navigate paragraphs**
> As a user, I want to jump to the next or previous paragraph, so I can skip or re-read sections.

### 3.3 Display

**US-020: Large word display**
> As a user, I want the current word displayed in large block letters, so it's easy to read at a glance.

**US-021: Context words**
> As a user, I want to see the previous and upcoming words, so I maintain reading context.

**US-022: Progress indicator**
> As a user, I want to see my progress through the document, so I know how much is left.

**US-023: Status display**
> As a user, I want to see the current WPM, play/pause status, and paragraph position.

**US-024: All caps display toggle**
> As a user, I want a toggle to display words in all capital letters or normal case, with all caps enabled by default.

**US-025: Interactive scrubber (TUI)**
> As a TUI user, I want to click or drag on the progress bar to jump to any position in the document.

**US-026: Interactive scrubber (Web)**
> As a web user, I want to use a slider to scrub through the document to any position.

**US-027: Go-to percentage**
> As a user, I want to press 'g' and enter a percentage to jump directly to that position in the document.

**US-028: Distraction-free mode**
> As a user, when I am actively reading (play mode), I want all UI elements except the current word to gradually fade away after a few seconds of keyboard inactivity, so I can focus entirely on the reading content without visual distractions. Pressing any key should instantly restore all UI elements, which will then start fading again if no further keyboard activity occurs.

### 3.4 Persistence

**US-030: Auto-save position**
> As a user, I want my reading position automatically saved, so I never lose my place.

**US-031: Manual save**
> As a user, I want to manually save my position, so I can be sure it's saved before exiting.

---

## 4. Functional Requirements

### 4.1 Document Support

| ID | Requirement |
|----|-------------|
| FR-001 | System SHALL parse plain text files (.txt) |
| FR-002 | System SHALL parse Markdown files (.md, .markdown) |
| FR-003 | System SHALL parse PDF files (.pdf) |
| FR-004 | System SHALL parse Microsoft Word files (.docx) |
| FR-005 | System SHALL parse EPUB ebooks (.epub) |
| FR-006 | System SHALL parse OpenDocument text files (.odt) |
| FR-007 | System SHALL extract document title from metadata when available |
| FR-008 | System SHALL generate document title from filename as fallback |
| FR-009 | System SHALL calculate SHA-256 hash of document content for identification |

### 4.2 Reading Engine

| ID | Requirement |
|----|-------------|
| FR-010 | System SHALL support WPM range of 50-2000 |
| FR-011 | System SHALL default to 300 WPM |
| FR-012 | System SHALL adjust speed in increments of 50 WPM |
| FR-013 | System SHALL add punctuation delay after sentence-ending punctuation (. ! ?) |
| FR-014 | System SHALL add smaller delay after clause punctuation (, ; :) |
| FR-015 | System SHALL support instant play/pause toggle |
| FR-016 | System SHALL track elapsed reading time |
| FR-017 | System SHALL calculate reading progress as percentage |

### 4.3 Navigation

| ID | Requirement |
|----|-------------|
| FR-020 | System SHALL support jumping to next paragraph |
| FR-021 | System SHALL support jumping to previous paragraph |
| FR-022 | System SHALL support jumping to start of current paragraph |
| FR-023 | System SHALL support jumping to specific word index |
| FR-024 | System SHALL support clicking/dragging on progress bar to jump to position (TUI) |
| FR-025 | System SHALL support slider scrubbing to any position (Web) |
| FR-026 | System SHALL support go-to mode ('g' key) to jump to a percentage |
| FR-027 | System SHALL auto-pause during scrubbing/go-to and optionally resume after |

### 4.4 Persistence

| ID | Requirement |
|----|-------------|
| FR-030 | System SHALL save reading position to ~/.config/cheetah/sessions.json |
| FR-031 | System SHALL save last used WPM per document |
| FR-032 | System SHALL save last access timestamp |
| FR-033 | System SHALL identify documents by content hash |
| FR-034 | System SHALL list all saved sessions sorted by last access |
| FR-035 | System SHALL auto-resume position and WPM when opening previously read document from file picker (TUI) |
| FR-036 | System SHALL provide toggle for all caps display (default: enabled) |
| FR-037 | System SHALL persist all caps preference in settings |

### 4.5 Distraction-Free Mode

| ID | Requirement |
|----|-------------|
| FR-040 | System SHALL start fading non-essential UI elements after 2 seconds of keyboard inactivity during play mode |
| FR-041 | System SHALL complete the fade-out animation over 3 seconds |
| FR-042 | System SHALL instantly restore all UI elements to full visibility on any key press |
| FR-043 | System SHALL keep the current word (block letters) fully visible at all times |
| FR-044 | System SHALL keep the Kartoza branding visible even when other elements have faded |
| FR-045 | System SHALL only activate distraction-free fading during active playback (not when paused) |

### 4.6 TUI Controls

| Key | Action |
|-----|--------|
| **Space** | Pause/resume reading |
| **j / Left Arrow** | Decrease speed 50 WPM |
| **k / Right Arrow** | Increase speed 50 WPM |
| **h / Up Arrow** | Previous paragraph |
| **l / Down Arrow** | Next paragraph |
| **1-9** | Speed presets (1=200, 2=300, ..., 9=1000 WPM) |
| **g** | Go-to mode (enter percentage to jump) |
| **c** | Toggle all caps display |
| **r** | Return to start of document |
| **s** | Save position |
| **Escape** | Return to document picker |
| **q** | Quit application |
| **Mouse click** | Click on progress bar to jump to position |
| **Mouse drag** | Drag on progress bar to scrub through document |
| **Mouse wheel** | Scroll to move forward/backward 10 words |

---

## 5. API Specification

### 5.1 Endpoints

#### Session Management

```
POST /api/v1/sessions
→ Creates new reading session
← {"session_id": "session-1"}

DELETE /api/v1/sessions/{id}
→ Deletes session
← 204 No Content
```

#### Document Operations

```
POST /api/v1/sessions/{id}/document
Content-Type: multipart/form-data
→ Uploads document file
← {"title": "...", "total_words": 45000, ...}

POST /api/v1/sessions/{id}/document/path
→ {"path": "/path/to/document.pdf"}
← {"title": "...", "total_words": 45000, ...}

GET /api/v1/sessions/{id}/document/info
← {"title": "...", "total_words": 45000, "total_paragraphs": 200, "hash": "..."}
```

#### Reading Control

```
GET /api/v1/sessions/{id}/state
← ReadingState JSON

POST /api/v1/sessions/{id}/play
← ReadingState JSON

POST /api/v1/sessions/{id}/pause
← ReadingState JSON

POST /api/v1/sessions/{id}/toggle
← ReadingState JSON

POST /api/v1/sessions/{id}/speed
→ {"wpm": 350}
← ReadingState JSON

POST /api/v1/sessions/{id}/paragraph/prev
← ReadingState JSON

POST /api/v1/sessions/{id}/paragraph/next
← ReadingState JSON
```

#### Persistence

```
POST /api/v1/sessions/{id}/save
← 204 No Content

GET /api/v1/saved
← [SavedSession, ...]

POST /api/v1/saved/{hash}/resume
← {"session_id": "...", "state": ReadingState}
```

### 5.2 Data Schemas

#### ReadingState

```json
{
  "current_word": "beautiful",
  "previous_word": "was",
  "next_words": ["day", "in", "the"],
  "word_index": 1523,
  "total_words": 45000,
  "paragraph_index": 45,
  "total_paragraphs": 200,
  "wpm": 350,
  "is_paused": true,
  "progress": 0.034,
  "elapsed_ms": 260800,
  "document_loaded": true,
  "document_title": "Example Book"
}
```

#### SavedSession

```json
{
  "document_hash": "sha256-abc123...",
  "document_path": "/home/user/book.epub",
  "document_title": "Example Book",
  "last_position": 1523,
  "total_words": 45000,
  "last_wpm": 350,
  "last_accessed": "2026-03-11T14:30:00Z"
}
```

---

## 6. UI Specification

### 6.1 TUI Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  🐆 CHEETAH - Document Title                   Word 1523 / 45000    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│                         · · · was · · ·                              │
│                                                                      │
│  ─────────────────────────────────────────────────────────────────  │
│                                                                      │
│         ██████  ████   ████  ██  ██  ██████  ████  ████  ██  ██     │
│         ██  ██  ██     ██  ██  ██  ██    ██    ██    ██  ██  ██     │
│         ██████  ████   ████████  ██    ██    ██    ██████  ██       │
│         ██  ██  ██     ██  ██  ██    ██    ██    ██  ██  ██  ██     │
│         ██████  ████   ██  ██  ████  ██    ████  ██  ██  ████       │
│                            beautiful                                 │
│  ─────────────────────────────────────────────────────────────────  │
│                                                                      │
│                            ▼ day ▼                                   │
│                              in                                      │
│                              the                                     │
│                                                                      │
│  ████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  3.4%        │
│                                                                      │
│                    ⏸ PAUSED │ 350 WPM │ ¶ 45/200                     │
│                                                                      │
├─────────────────────────────────────────────────────────────────────┤
│ SPACE pause │ j/k speed │ h/l para │ 1-9 presets │ c caps │ ESC back  │
│ Made with ♥ by Kartoza │ Donate! │ GitHub                           │
└─────────────────────────────────────────────────────────────────────┘
```

### 6.2 Layout Zones

| Zone | Content | Distraction-Free Behavior |
|------|---------|---------------------------|
| Header | App icon, document title, word count | Fades to invisible |
| Previous | Faded previous word | Fades to invisible |
| Current | Large block letters with original text | Always visible |
| Next | Upcoming 3 words with decreasing opacity | Fades to invisible |
| Progress | Gradient progress bar with percentage | Fades to invisible |
| Status | Play/pause, WPM, paragraph position | Fades to invisible |
| Footer | Key bindings, Kartoza branding | Help fades, branding stays visible |

---

## 7. Testing Requirements

### 7.1 Unit Tests

- Document parsers for each format
- Text processor (paragraph/word extraction)
- Reading engine state management
- Session persistence

### 7.2 Integration Tests

- REST API endpoints
- Full document load and reading flow
- Position save/resume

### 7.3 Manual Testing

- TUI on Linux, macOS, Windows terminals
- Web UI on Chrome, Firefox, Safari, Edge
- Various document sizes (1KB to 10MB)

---

## 8. Non-Functional Requirements

### 8.1 Performance

| Requirement | Target |
|-------------|--------|
| Document parse time | < 2s for 1MB document |
| Word advance latency | < 10ms |
| API response time | < 50ms |
| Memory usage | < 100MB typical |

### 8.2 Usability

| Requirement | Target |
|-------------|--------|
| Time to first word | < 3s from document selection |
| Learning curve | < 5 minutes to understand controls |
| Keyboard-only operation | Full functionality without mouse |

### 8.3 Privacy

| Requirement |
|-------------|
| Web documents processed entirely client-side |
| No document content sent to server |
| Sessions stored locally only |

---

## 9. Dependencies

### 9.1 Go Dependencies

| Package | Purpose |
|---------|---------|
| charmbracelet/bubbletea | TUI framework |
| charmbracelet/lipgloss | TUI styling |
| charmbracelet/harmonica | Spring animations |
| charmbracelet/bubbles | TUI components (file picker) |
| ledongthuc/pdf | PDF text extraction |
| nguyenthenguyen/docx | DOCX parsing |
| taylorskalyo/goreader | EPUB parsing |

### 9.2 Web Dependencies

| Package | Purpose |
|---------|---------|
| react, react-dom | UI framework |
| @chakra-ui/react | UI components |
| framer-motion | Animations |
| pdfjs-dist | PDF parsing |
| mammoth | DOCX parsing |
| epubjs | EPUB parsing |

---

## 10. Changelog

### Version 0.2.0 (Current)

- **Distraction-free mode**: UI elements (header, progress indicator, previous/next words, status line, help text) gradually fade away during active reading after 2 seconds of keyboard inactivity
- Current word block letters remain fully visible for uninterrupted reading focus
- Kartoza branding remains visible (for web UI ad display)
- Any key press instantly restores all UI elements

### Version 0.1.0 (Initial Release)

- Basic document loading (TXT, MD, PDF, DOCX, EPUB, ODT)
- RSVP reading engine with timer-based advancement
- Pause/play, speed control, paragraph navigation
- Position persistence per document
- TUI with block letter display and carousel animation
- REST API for frontend communication

---

Made with ♥ by [Kartoza](https://kartoza.com) | [Donate!](https://github.com/sponsors/timlinux) | [GitHub](https://github.com/timlinux/cheetah)
