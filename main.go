// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

// Cheetah - RSVP Speed Reading Application
//
// This is the entry point for the RSVP speed reading application.
//
// By default, the TUI runs as a standalone application with an embedded
// reading engine (no HTTP server required). For web mode, use the 'web'
// subcommand which starts an HTTP server.
//
// RSVP (Rapid Serial Visual Presentation) displays words one at a time
// in a fixed location, allowing users to read at speeds up to 1000+ WPM.
//
// Usage:
//
//	cheetah                     # Opens file picker to select a document
//	cheetah /path/to/book.pdf   # Opens document directly
//	cheetah web -port 8787      # Start web server with API
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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/timlinux/cheetah/backend"
	"github.com/timlinux/cheetah/frontend"
)

const Version = "0.2.0"

func main() {
	// Check for subcommands first
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "web":
			runWebCommand(os.Args[2:])
			return
		case "server":
			runServerCommand(os.Args[2:])
			return
		case "version":
			fmt.Printf("Cheetah v%s\n", Version)
			return
		case "help", "-h", "--help":
			printHelp()
			return
		}
	}

	// Parse command line flags for TUI mode
	wpm := flag.Int("wpm", 300, "Initial words per minute speed")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Cheetah v%s\n", Version)
		return
	}

	// Get document path from remaining args
	var documentPath string
	if flag.NArg() > 0 {
		documentPath = flag.Arg(0)
	}

	// Run TUI with embedded engine (standalone mode - no HTTP server)
	runStandaloneTUI(documentPath, *wpm)
}

// runStandaloneTUI runs the TUI with an embedded engine (no HTTP server needed).
func runStandaloneTUI(documentPath string, wpm int) {
	// Create and run TUI with embedded engine
	model := frontend.NewModel(documentPath, wpm)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// runServerCommand handles the 'server' subcommand for running API server only.
func runServerCommand(args []string) {
	serverFlags := flag.NewFlagSet("server", flag.ExitOnError)
	port := serverFlags.Int("port", 8787, "Port for the REST API server")

	serverFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: cheetah server [options]\n\n")
		fmt.Fprintf(os.Stderr, "Run the backend REST API server only.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		serverFlags.PrintDefaults()
	}

	if err := serverFlags.Parse(args); err != nil {
		os.Exit(1)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	runServerOnly(addr)
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

	fmt.Printf("🐆 Cheetah API server starting on %s\n", addr)
	fmt.Printf("   PID: %d (written to %s)\n", os.Getpid(), pidFile)
	fmt.Println("   Press Ctrl+C to stop")

	if err := server.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
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

// printHelp prints usage information.
func printHelp() {
	fmt.Println("🐆 Cheetah - RSVP Speed Reading Application")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  cheetah                     Open file picker (standalone TUI)")
	fmt.Println("  cheetah <file>              Open document directly")
	fmt.Println("  cheetah web [options]       Start web server with API")
	fmt.Println("  cheetah server [options]    Start API server only")
	fmt.Println("  cheetah version             Show version")
	fmt.Println("  cheetah help                Show this help")
	fmt.Println()
	fmt.Println("TUI Options:")
	fmt.Println("  -wpm int       Initial reading speed (default 300)")
	fmt.Println()
	fmt.Println("Web Options:")
	fmt.Println("  -port int      Port for web server (default 8787)")
	fmt.Println("  -dir string    Directory for web frontend (default web/dist)")
	fmt.Println()
	fmt.Println("Keyboard Controls (TUI):")
	fmt.Println("  Space          Pause/resume reading")
	fmt.Println("  j/k            Decrease/increase speed")
	fmt.Println("  h/l            Previous/next paragraph")
	fmt.Println("  1-9            Speed presets (200-1000 WPM)")
	fmt.Println("  r              Return to start")
	fmt.Println("  s              Save position")
	fmt.Println("  Esc            Return to file picker")
	fmt.Println("  q              Quit")
	fmt.Println()
	fmt.Println("Made with ❤️ by Kartoza | https://kartoza.com")
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
