---
title: "API Reference"
weight: 3
---

Cheetah provides a REST API for controlling the reading engine.

## Base URL

```
http://localhost:8787/api/v1
```

## Endpoints

### Session Management

#### Create Session
```
POST /sessions
```

Returns:
```json
{
  "session_id": "abc123"
}
```

#### Delete Session
```
DELETE /sessions/{id}
```

### Document Operations

#### Load Document from Path
```
POST /sessions/{id}/document/path
Content-Type: application/json

{"path": "/path/to/document.pdf"}
```

#### Get Document Info
```
GET /sessions/{id}/document/info
```

Returns:
```json
{
  "title": "Example Document",
  "total_words": 45000,
  "total_paragraphs": 200
}
```

### Reading Control

#### Get Current State
```
GET /sessions/{id}/state
```

Returns:
```json
{
  "current_word": "beautiful",
  "word_index": 1523,
  "total_words": 45000,
  "wpm": 350,
  "is_paused": true,
  "progress": 0.034
}
```

#### Play/Pause/Toggle
```
POST /sessions/{id}/play
POST /sessions/{id}/pause
POST /sessions/{id}/toggle
```

#### Set Speed
```
POST /sessions/{id}/speed
Content-Type: application/json

{"wpm": 400}
```

#### Navigation
```
POST /sessions/{id}/paragraph/prev
POST /sessions/{id}/paragraph/next
POST /sessions/{id}/word/{index}
```

### Persistence

#### Save Position
```
POST /sessions/{id}/save
```

#### Get Saved Sessions
```
GET /saved
```

#### Resume Session
```
POST /saved/{hash}/resume
```
