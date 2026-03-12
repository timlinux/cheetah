// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// Server handles HTTP requests for the reading API
type Server struct {
	config    Config
	addr      string
	sessions  map[string]*Engine
	mu        sync.RWMutex
	staticDir string
}

// NewServer creates a new HTTP server
func NewServer(config Config, addr string) (*Server, error) {
	return &Server{
		config:   config,
		addr:     addr,
		sessions: make(map[string]*Engine),
	}, nil
}

// SetStaticDir sets the directory for serving static files
func (s *Server) SetStaticDir(dir string) {
	s.staticDir = dir
}

// Start begins listening for HTTP requests (blocking)
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/api/health", s.handleHealth)

	// Session management
	mux.HandleFunc("POST /api/v1/sessions", s.handleCreateSession)
	mux.HandleFunc("DELETE /api/v1/sessions/{id}", s.handleDeleteSession)

	// Document operations
	mux.HandleFunc("POST /api/v1/sessions/{id}/document", s.handleUploadDocument)
	mux.HandleFunc("POST /api/v1/sessions/{id}/document/path", s.handleLoadDocumentPath)
	mux.HandleFunc("GET /api/v1/sessions/{id}/document/info", s.handleGetDocumentInfo)

	// Reading control
	mux.HandleFunc("GET /api/v1/sessions/{id}/state", s.handleGetState)
	mux.HandleFunc("POST /api/v1/sessions/{id}/play", s.handlePlay)
	mux.HandleFunc("POST /api/v1/sessions/{id}/pause", s.handlePause)
	mux.HandleFunc("POST /api/v1/sessions/{id}/toggle", s.handleToggle)
	mux.HandleFunc("POST /api/v1/sessions/{id}/speed", s.handleSetSpeed)
	mux.HandleFunc("POST /api/v1/sessions/{id}/paragraph/prev", s.handlePrevParagraph)
	mux.HandleFunc("POST /api/v1/sessions/{id}/paragraph/next", s.handleNextParagraph)
	mux.HandleFunc("POST /api/v1/sessions/{id}/word/{index}", s.handleJumpToWord)

	// Persistence
	mux.HandleFunc("POST /api/v1/sessions/{id}/save", s.handleSavePosition)
	mux.HandleFunc("GET /api/v1/saved", s.handleGetSavedSessions)
	mux.HandleFunc("POST /api/v1/saved/{hash}/resume", s.handleResumeSession)

	// Serve Hugo docs at /docs/
	docsDir := "hugo/public"
	if _, err := os.Stat(docsDir); err == nil {
		docsFS := http.FileServer(http.Dir(docsDir))
		mux.Handle("/docs/", http.StripPrefix("/docs/", docsFS))
	}

	// Static files for web frontend
	if s.staticDir != "" {
		fs := http.FileServer(http.Dir(s.staticDir))
		mux.Handle("/", fs)
	}

	return http.ListenAndServe(s.addr, mux)
}

// StartAsync starts the server in a goroutine
func (s *Server) StartAsync() {
	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()
}

// getEngine retrieves an engine for a session
func (s *Server) getEngine(sessionID string) *Engine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[sessionID]
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleCreateSession creates a new reading session
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate session ID
	sessionID := fmt.Sprintf("session-%d", len(s.sessions)+1)

	// Create engine
	engine := NewEngine(s.config)
	s.sessions[sessionID] = engine

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"session_id": sessionID})
}

// handleDeleteSession deletes a reading session
func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// handleUploadDocument handles multipart document upload
func (s *Server) handleUploadDocument(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	engine := s.getEngine(sessionID)
	if engine == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(100 << 20); err != nil { // 100MB max
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("document")
	if err != nil {
		http.Error(w, "No document provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read document", http.StatusInternalServerError)
		return
	}

	if err := engine.LoadDocumentBytes(data, header.Filename); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse document: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engine.GetDocumentInfo())
}

// handleLoadDocumentPath loads a document from filesystem path
func (s *Server) handleLoadDocumentPath(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	engine := s.getEngine(sessionID)
	if engine == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := engine.LoadDocument(req.Path); err != nil {
		http.Error(w, fmt.Sprintf("Failed to load document: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engine.GetDocumentInfo())
}

// handleGetDocumentInfo returns document metadata
func (s *Server) handleGetDocumentInfo(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	engine := s.getEngine(sessionID)
	if engine == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	info := engine.GetDocumentInfo()
	if info == nil {
		http.Error(w, "No document loaded", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// handleGetState returns the current reading state
func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	engine := s.getEngine(sessionID)
	if engine == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engine.GetState())
}

// handlePlay starts reading
func (s *Server) handlePlay(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	engine := s.getEngine(sessionID)
	if engine == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	engine.Play()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engine.GetState())
}

// handlePause pauses reading
func (s *Server) handlePause(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	engine := s.getEngine(sessionID)
	if engine == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	engine.Pause()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engine.GetState())
}

// handleToggle toggles play/pause
func (s *Server) handleToggle(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	engine := s.getEngine(sessionID)
	if engine == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	engine.Toggle()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engine.GetState())
}

// handleSetSpeed sets the reading speed
func (s *Server) handleSetSpeed(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	engine := s.getEngine(sessionID)
	if engine == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	var req struct {
		WPM int `json:"wpm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	engine.SetWPM(req.WPM)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engine.GetState())
}

// handlePrevParagraph moves to the previous paragraph
func (s *Server) handlePrevParagraph(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	engine := s.getEngine(sessionID)
	if engine == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	engine.PrevParagraph()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engine.GetState())
}

// handleNextParagraph moves to the next paragraph
func (s *Server) handleNextParagraph(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	engine := s.getEngine(sessionID)
	if engine == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	engine.NextParagraph()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engine.GetState())
}

// handleJumpToWord moves to a specific word
func (s *Server) handleJumpToWord(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	engine := s.getEngine(sessionID)
	if engine == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	var index int
	fmt.Sscanf(r.PathValue("index"), "%d", &index)

	engine.JumpToWord(index)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engine.GetState())
}

// handleSavePosition saves the current reading position
func (s *Server) handleSavePosition(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	engine := s.getEngine(sessionID)
	if engine == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if err := engine.SavePosition(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save position: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGetSavedSessions returns all saved sessions
func (s *Server) handleGetSavedSessions(w http.ResponseWriter, r *http.Request) {
	// Get saved sessions from any engine (they share the same storage)
	s.mu.RLock()
	var engine *Engine
	for _, e := range s.sessions {
		engine = e
		break
	}
	s.mu.RUnlock()

	var sessions []SavedSession
	if engine != nil {
		sessions = engine.GetSavedSessions()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// handleResumeSession resumes a saved session
func (s *Server) handleResumeSession(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")

	// Create a new session for resuming
	s.mu.Lock()
	sessionID := fmt.Sprintf("session-%d", len(s.sessions)+1)
	engine := NewEngine(s.config)
	s.sessions[sessionID] = engine
	s.mu.Unlock()

	if err := engine.ResumeSession(hash); err != nil {
		http.Error(w, fmt.Sprintf("Failed to resume session: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": sessionID,
		"state":      engine.GetState(),
	})
}
