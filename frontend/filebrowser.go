// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package frontend

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/timlinux/cheetah/documents"
	"github.com/timlinux/cheetah/settings"
)

// FileEntry represents a file or directory in the browser
type FileEntry struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime time.Time
	Icon    string
}

// FileBrowser is a beautiful file browser like yazi/midnight commander
type FileBrowser struct {
	// Current directory
	CurrentDir string

	// List of entries
	Entries []FileEntry

	// Cursor position
	Cursor int

	// Scroll offset for large directories
	ScrollOffset int

	// Visible height (number of entries that fit)
	VisibleHeight int

	// Width of the browser
	Width int

	// Styles
	styles Styles

	// Error message if any
	Error string

	// Show hidden files
	ShowHidden bool

	// Allowed file extensions (empty = all)
	AllowedTypes []string

	// Selected file path (when user presses Enter on a file)
	SelectedFile string
}

// NewFileBrowser creates a new file browser
func NewFileBrowser() *FileBrowser {
	fb := &FileBrowser{
		styles:        NewStyles(),
		ShowHidden:    false,
		AllowedTypes:  documents.SupportedFormats(),
		VisibleHeight: 20,
		Width:         80,
	}

	// Load last directory from settings
	s, err := settings.Load()
	if err == nil && s.LastDirectory != "" {
		if info, err := os.Stat(s.LastDirectory); err == nil && info.IsDir() {
			fb.CurrentDir = s.LastDirectory
		}
	}

	// Default to home directory if no saved directory
	if fb.CurrentDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fb.CurrentDir = "/"
		} else {
			fb.CurrentDir = home
		}
	}

	fb.Refresh()
	return fb
}

// Refresh reloads the directory listing
func (fb *FileBrowser) Refresh() {
	fb.Entries = nil
	fb.Error = ""

	// Read directory
	entries, err := os.ReadDir(fb.CurrentDir)
	if err != nil {
		fb.Error = err.Error()
		return
	}

	// Always add parent directory entry (except at root)
	if fb.CurrentDir != "/" {
		fb.Entries = append(fb.Entries, FileEntry{
			Name:  "..",
			Path:  filepath.Dir(fb.CurrentDir),
			IsDir: true,
			Icon:  IconFolderUp,
		})
	}

	// Process entries
	var dirs []FileEntry
	var files []FileEntry

	for _, entry := range entries {
		// Skip hidden files unless enabled
		if !fb.ShowHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fe := FileEntry{
			Name:    entry.Name(),
			Path:    filepath.Join(fb.CurrentDir, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}

		if entry.IsDir() {
			fe.Icon = IconFolder
			dirs = append(dirs, fe)
		} else {
			// Check if file type is allowed
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if len(fb.AllowedTypes) > 0 && !contains(fb.AllowedTypes, ext) {
				continue
			}
			fe.Icon = getFileIcon(ext)
			files = append(files, fe)
		}
	}

	// Sort: directories first, then files, both alphabetically
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	fb.Entries = append(fb.Entries, dirs...)
	fb.Entries = append(fb.Entries, files...)

	// Reset cursor if out of bounds
	if fb.Cursor >= len(fb.Entries) {
		fb.Cursor = len(fb.Entries) - 1
	}
	if fb.Cursor < 0 {
		fb.Cursor = 0
	}

	// Save current directory to settings
	s, err := settings.Load()
	if err == nil {
		s.LastDirectory = fb.CurrentDir
		s.Save()
	}
}

// SetSize updates the browser dimensions
func (fb *FileBrowser) SetSize(width, height int) {
	fb.Width = width
	// Reserve space for header, footer, borders
	fb.VisibleHeight = height - 12
	if fb.VisibleHeight < 5 {
		fb.VisibleHeight = 5
	}
}

// MoveUp moves cursor up
func (fb *FileBrowser) MoveUp() {
	if fb.Cursor > 0 {
		fb.Cursor--
		fb.adjustScroll()
	}
}

// MoveDown moves cursor down
func (fb *FileBrowser) MoveDown() {
	if fb.Cursor < len(fb.Entries)-1 {
		fb.Cursor++
		fb.adjustScroll()
	}
}

// PageUp moves cursor up by page
func (fb *FileBrowser) PageUp() {
	fb.Cursor -= fb.VisibleHeight
	if fb.Cursor < 0 {
		fb.Cursor = 0
	}
	fb.adjustScroll()
}

// PageDown moves cursor down by page
func (fb *FileBrowser) PageDown() {
	fb.Cursor += fb.VisibleHeight
	if fb.Cursor >= len(fb.Entries) {
		fb.Cursor = len(fb.Entries) - 1
	}
	fb.adjustScroll()
}

// GoHome moves to first entry
func (fb *FileBrowser) GoHome() {
	fb.Cursor = 0
	fb.ScrollOffset = 0
}

// GoEnd moves to last entry
func (fb *FileBrowser) GoEnd() {
	fb.Cursor = len(fb.Entries) - 1
	fb.adjustScroll()
}

// Enter processes Enter key - navigates into directory or selects file
func (fb *FileBrowser) Enter() bool {
	if len(fb.Entries) == 0 || fb.Cursor >= len(fb.Entries) {
		return false
	}

	entry := fb.Entries[fb.Cursor]

	if entry.IsDir {
		fb.CurrentDir = entry.Path
		fb.Cursor = 0
		fb.ScrollOffset = 0
		fb.Refresh()
		return false
	}

	// File selected
	fb.SelectedFile = entry.Path
	return true
}

// GoBack navigates to parent directory
func (fb *FileBrowser) GoBack() {
	if fb.CurrentDir != "/" {
		fb.CurrentDir = filepath.Dir(fb.CurrentDir)
		fb.Cursor = 0
		fb.ScrollOffset = 0
		fb.Refresh()
	}
}

// ToggleHidden toggles showing hidden files
func (fb *FileBrowser) ToggleHidden() {
	fb.ShowHidden = !fb.ShowHidden
	fb.Refresh()
}

// adjustScroll ensures cursor is visible
func (fb *FileBrowser) adjustScroll() {
	// Scroll up if cursor is above visible area
	if fb.Cursor < fb.ScrollOffset {
		fb.ScrollOffset = fb.Cursor
	}

	// Scroll down if cursor is below visible area
	if fb.Cursor >= fb.ScrollOffset+fb.VisibleHeight {
		fb.ScrollOffset = fb.Cursor - fb.VisibleHeight + 1
	}
}

// View renders the file browser
func (fb *FileBrowser) View() string {
	var b strings.Builder

	// Calculate inner width (accounting for border padding)
	innerWidth := fb.Width - 4
	if innerWidth < 40 {
		innerWidth = 40
	}

	// Header with title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColourKartozaOrange)).
		Bold(true)
	headerIcon := IconCheetah + " "
	headerText := titleStyle.Render(headerIcon + "CHEETAH - Select Document")
	header := lipgloss.PlaceHorizontal(innerWidth, lipgloss.Center, headerText)
	b.WriteString(header)
	b.WriteString("\n")

	// Path line
	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColourPath))
	pathText := "📂 " + fb.CurrentDir
	if len(pathText) > innerWidth {
		pathText = "📂 ..." + fb.CurrentDir[len(fb.CurrentDir)-(innerWidth-6):]
	}
	b.WriteString(pathStyle.Render(pathText))
	b.WriteString("\n")

	// Separator
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColourBorder))
	separator := sepStyle.Render(strings.Repeat("─", innerWidth))
	b.WriteString(separator)
	b.WriteString("\n")

	// Column headers
	colHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColourHeader)).
		Bold(true)
	nameHeader := colHeaderStyle.Render(fmt.Sprintf("%-*s", innerWidth-22, "Name"))
	sizeHeader := colHeaderStyle.Render(fmt.Sprintf("%10s", "Size"))
	dateHeader := colHeaderStyle.Render(fmt.Sprintf("%10s", "Modified"))
	b.WriteString(nameHeader + sizeHeader + " " + dateHeader)
	b.WriteString("\n")
	b.WriteString(separator)
	b.WriteString("\n")

	// File listing
	if fb.Error != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		b.WriteString(errorStyle.Render("Error: " + fb.Error))
		b.WriteString("\n")
	} else if len(fb.Entries) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true)
		b.WriteString(emptyStyle.Render("  (empty directory)"))
		b.WriteString("\n")
	} else {
		// Calculate visible entries
		start := fb.ScrollOffset
		end := start + fb.VisibleHeight
		if end > len(fb.Entries) {
			end = len(fb.Entries)
		}

		for i := start; i < end; i++ {
			entry := fb.Entries[i]
			isSelected := i == fb.Cursor

			// Format name with icon
			icon := entry.Icon
			name := entry.Name
			maxNameLen := innerWidth - 24
			if len(name) > maxNameLen {
				name = name[:maxNameLen-3] + "..."
			}
			nameField := fmt.Sprintf("%s %-*s", icon, maxNameLen, name)

			// Format size
			var sizeField string
			if entry.IsDir {
				sizeField = fmt.Sprintf("%10s", "<DIR>")
			} else {
				sizeField = fmt.Sprintf("%10s", formatSize(entry.Size))
			}

			// Format date
			dateField := fmt.Sprintf("%10s", entry.ModTime.Format("Jan 02 15:04"))

			line := nameField + sizeField + " " + dateField

			if isSelected {
				selectedStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(ColourSelectedFg)).
					Background(lipgloss.Color(ColourSelected)).
					Bold(true)
				b.WriteString(selectedStyle.Render(line))
			} else if entry.IsDir {
				dirStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(ColourDirIcon))
				b.WriteString(dirStyle.Render(line))
			} else {
				fileStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#CCCCCC"))
				b.WriteString(fileStyle.Render(line))
			}
			b.WriteString("\n")
		}

		// Pad remaining visible height
		for i := end - start; i < fb.VisibleHeight; i++ {
			b.WriteString("\n")
		}
	}

	// Bottom separator
	b.WriteString(separator)
	b.WriteString("\n")

	// Status bar
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColourHelp))
	statusText := fmt.Sprintf("%d/%d items", fb.Cursor+1, len(fb.Entries))
	if fb.ShowHidden {
		statusText += " [hidden: on]"
	}
	b.WriteString(statusStyle.Render(statusText))
	b.WriteString("\n")

	// Supported formats
	formatsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	formats := formatsStyle.Render("Supported: PDF, DOCX, EPUB, ODT, TXT, MD")
	b.WriteString(lipgloss.PlaceHorizontal(innerWidth, lipgloss.Center, formats))
	b.WriteString("\n\n")

	// Help line
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	help := "↑/↓ navigate │ Enter select │ Backspace/h up │ Tab resume │ . hidden │ ? docs │ q quit"
	b.WriteString(helpStyle.Render(help))
	b.WriteString("\n")

	// Kartoza branding
	kartozaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	heartStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColourKartozaOrange))
	linkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColourCheetahGold))
	kartozaLine := kartozaStyle.Render("Made with ") +
		heartStyle.Render("♥") +
		kartozaStyle.Render(" by ") +
		linkStyle.Render("Kartoza") +
		kartozaStyle.Render(" │ ") +
		linkStyle.Render("Donate!") +
		kartozaStyle.Render(" │ ") +
		linkStyle.Render("GitHub")
	b.WriteString(lipgloss.PlaceHorizontal(innerWidth, lipgloss.Center, kartozaLine))

	// Wrap in border
	content := b.String()
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColourBorder)).
		Padding(0, 1)

	return borderStyle.Render(content)
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getFileIcon(ext string) string {
	switch ext {
	case ".pdf":
		return IconPDF
	case ".docx", ".doc":
		return IconWord
	case ".epub":
		return IconEpub
	case ".txt":
		return IconText
	case ".md", ".markdown":
		return IconMarkdown
	case ".odt":
		return IconODT
	default:
		return IconFile
	}
}

func formatSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1fG", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1fM", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1fK", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%dB", size)
	}
}
