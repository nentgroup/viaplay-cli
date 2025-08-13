// Package output provides formatting utilities for CLI output.
package output

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// HookOutputWriter is a writer that formats and prints hook output
type HookOutputWriter struct {
	title       string
	mu          sync.Mutex
	content     strings.Builder
	borderColor lipgloss.Color
}

// NewHookOutputWriter creates a new writer for hook output
func NewHookOutputWriter(title string) *HookOutputWriter {
	return &HookOutputWriter{
		title:       title,
		content:     strings.Builder{},
		borderColor: "#E6007A", // Viaplay pink colour
	}
}

// Write implements io.Writer
func (w *HookOutputWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Write to content buffer
	w.content.Write(p)

	// Also write to stdout directly
	os.Stdout.Write(p)

	return len(p), nil
}

// DisplayHookOutput runs the hook and displays its output with simple formatting
func DisplayHookOutput(title string, runHookFn func(stdout, stderr io.Writer) error) error {
	// Create title style
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E6007A")).
		Bold(true).
		Padding(0, 1)

	// Print formatted title
	fmt.Println()
	fmt.Println(titleStyle.Render(title))
	fmt.Println()

	// Create writer for output
	writer := NewHookOutputWriter(title)

	// Run the hook with our writer
	err := runHookFn(writer, writer)

	// Print completion message
	fmt.Println()
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Italic(true)

	if err != nil {
		fmt.Println(statusStyle.Render(fmt.Sprintf("Hook failed: %v", err)))
	} else {
		fmt.Println(statusStyle.Render("Hook completed successfully"))
	}
	fmt.Println()

	return err
}
