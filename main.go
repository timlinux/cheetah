// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

// Cheetah - RSVP Speed Reading Application
//
// This is the entry point that starts the REST API server (backend)
// and connects the terminal UI (frontend) to it via HTTP.
//
// RSVP (Rapid Serial Visual Presentation) displays words one at a time
// in a fixed location, allowing users to read at speeds up to 1000+ WPM.
//
// Architecture:
//   - Backend: REST API server handling document parsing and reading state
//   - Frontend: Bubble Tea TUI communicating via REST client
//
// Usage:
//
//	cheetah                     # Opens file picker to select a document
//	cheetah /path/to/book.pdf   # Opens document directly
//	cheetah -port 8787          # Use custom port for REST API
//	cheetah -server             # Run backend server only (blocking)
//	cheetah -client             # Run frontend only (connect to existing backend)
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/timlinux/cheetah/backend"
	"github.com/timlinux/cheetah/frontend"
)

const Version = "0.1.0"

func main() {
	// Check for subcommands first
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "web":
			runWebCommand(os.Args[2:])
			return
		case "version":
			fmt.Printf("Cheetah v%s\n", Version)
			return
		}
	}

	// Parse command line flags
	port := flag.Int("port", 8787, "Port for the REST API server")
	serverOnly := flag.Bool("server", false, "Run backend server only (no TUI)")
	clientOnly := flag.Bool("client", false, "Run frontend only (connect to existing backend)")
	wpm := flag.Int("wpm", 300, "Initial words per minute speed")
	flag.Parse()

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	baseURL := fmt.Sprintf("http://%s", addr)

	// Get document path from remaining args
	var documentPath string
	if flag.NArg() > 0 {
		documentPath = flag.Arg(0)
	}

	// Validate flags
	if *serverOnly && *clientOnly {
		fmt.Println("Error: cannot use both -server and -client flags")
		os.Exit(1)
	}

	// Server-only mode: run backend and block
	if *serverOnly {
		runServerOnly(addr)
		return
	}

	// Client-only mode: connect to existing backend
	if *clientOnly {
		runClientOnly(baseURL, documentPath, *wpm)
		return
	}

	// Default mode: start backend and frontend together
	runCombined(addr, baseURL, documentPath, *wpm)
}

// runServerOnly starts the backend server and blocks until interrupted.
func runServerOnly(addr string) {
	config := backend.DefaultConfig()

	server, err := backend.NewServer(config, addr)
	if err != nil {
		fmt.Printf("Error creating server: %v\n", err)
		os.Exit(1)
	}

	// Write PID file for management scripts
	pidFile := getPIDFilePath()
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		fmt.Printf("Warning: could not write PID file: %v\n", err)
	}
	defer os.Remove(pidFile)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down server...")
		os.Remove(pidFile)
		os.Exit(0)
	}()

	fmt.Printf("Cheetah backend server starting on %s\n", addr)
	fmt.Printf("PID: %d (written to %s)\n", os.Getpid(), pidFile)
	fmt.Println("Press Ctrl+C to stop")

	if err := server.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

// runClientOnly connects to an existing backend server.
func runClientOnly(baseURL, documentPath string, wpm int) {
	client := frontend.NewClient(baseURL)

	// Wait for server to be ready
	fmt.Printf("Connecting to backend at %s...\n", baseURL)
	if err := client.WaitForServer(5 * time.Second); err != nil {
		fmt.Printf("Error: Could not connect to backend: %v\n", err)
		fmt.Println("Make sure the backend is running with: cheetah -server")
		os.Exit(1)
	}

	// Create a session on the server
	if err := client.CreateSession(); err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		os.Exit(1)
	}
	defer client.DeleteSession()

	// Create and run TUI
	model := frontend.NewModel(client, documentPath, wpm)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// runCombined starts both backend and frontend together (default mode).
func runCombined(addr, baseURL, documentPath string, wpm int) {
	config := backend.DefaultConfig()

	server, err := backend.NewServer(config, addr)
	if err != nil {
		fmt.Printf("Error creating server: %v\n", err)
		os.Exit(1)
	}

	// Start server in background
	server.StartAsync()

	client := frontend.NewClient(baseURL)

	// Wait for server to be ready
	if err := client.WaitForServer(2 * time.Second); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Create a session on the server
	if err := client.CreateSession(); err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		os.Exit(1)
	}
	defer client.DeleteSession()

	// Create and run TUI
	model := frontend.NewModel(client, documentPath, wpm)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// runWebCommand handles the 'web' subcommand for serving the web frontend.
func runWebCommand(args []string) {
	webFlags := flag.NewFlagSet("web", flag.ExitOnError)
	port := webFlags.Int("port", 8787, "Port for the web server")
	webDir := webFlags.String("dir", "web/dist", "Directory containing built web frontend")

	webFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: cheetah web [options]\n\n")
		fmt.Fprintf(os.Stderr, "Serve the web frontend with the backend API.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		webFlags.PrintDefaults()
	}

	if err := webFlags.Parse(args); err != nil {
		os.Exit(1)
	}

	addr := fmt.Sprintf("0.0.0.0:%d", *port)

	config := backend.DefaultConfig()

	server, err := backend.NewServer(config, addr)
	if err != nil {
		fmt.Printf("Error creating server: %v\n", err)
		os.Exit(1)
	}

	// Check if web directory exists
	if _, err := os.Stat(*webDir); os.IsNotExist(err) {
		fmt.Printf("Warning: Web directory '%s' not found.\n", *webDir)
		fmt.Println("Run 'make web-build' first to build the web frontend.")
		fmt.Println("Starting API server only...")
	} else {
		// Set the static file directory
		server.SetStaticDir(*webDir)
	}

	// Write PID file for management scripts
	pidFile := getPIDFilePath()
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		fmt.Printf("Warning: could not write PID file: %v\n", err)
	}
	defer os.Remove(pidFile)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down server...")
		os.Remove(pidFile)
		os.Exit(0)
	}()

	fmt.Printf("🐆 Cheetah web server starting on http://%s\n", addr)
	fmt.Println("   Press Ctrl+C to stop")

	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

// getPIDFilePath returns the path to the PID file.
func getPIDFilePath() string {
	// Use XDG runtime dir if available, otherwise /tmp
	runDir := os.Getenv("XDG_RUNTIME_DIR")
	if runDir == "" {
		runDir = "/tmp"
	}
	return filepath.Join(runDir, "cheetah.pid")
}
