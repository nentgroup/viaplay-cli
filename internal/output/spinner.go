// Package output provides formatting utilities for CLI output.
package output

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
)

// Spinner represents a CLI spinner for showing progress
type Spinner struct {
	spinner      *spinner.Spinner
	message      string
	mu           sync.RWMutex
	isActive     bool
	lastMessage  string
	output       io.Writer
	successStyle lipgloss.Style
	errorStyle   lipgloss.Style
	skipStyle    lipgloss.Style
	warningStyle lipgloss.Style
}

// NewSpinner creates a new spinner
func NewSpinner() *Spinner {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Writer = os.Stderr
	if err := s.Color("red", "bold"); err != nil {
		// If colour can't be set, continue with default colour; ignore logging errors
		_, _ = fmt.Fprintf(os.Stderr, "warning: failed to set spinner color: %v\n", err)
	}

	return &Spinner{
		spinner:      s,
		output:       os.Stderr,
		successStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("76")),  // Green
		errorStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("161")), // Red
		skipStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")), // Gray
		warningStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("214")), // Yellow
	}
}

// Start begins the spinner animation with the given message
func (s *Spinner) Start(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isActive {
		s.message = message
		s.spinner.Suffix = " " + message
		return
	}

	s.isActive = true
	s.message = message
	s.lastMessage = message

	// Set the suffix to be our message
	s.spinner.Suffix = " " + message

	// Start the spinner
	s.spinner.Start()
}

// Update changes the spinner message while it's running
func (s *Spinner) Update(message string) {
	s.mu.Lock()
	s.message = message
	s.lastMessage = message
	s.spinner.Suffix = " " + message
	s.mu.Unlock()
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.isActive {
		s.mu.Unlock()
		return
	}
	s.isActive = false
	s.mu.Unlock()

	s.spinner.Stop()
}

// Success stops the spinner and displays a success message
func (s *Spinner) Success(message string) {
	s.Stop()
	_, _ = fmt.Fprintf(s.output, "%s %s\n", ActiveIcons.Success, s.successStyle.Render(message))
}

// Fail stops the spinner and displays an error message
func (s *Spinner) Fail(message string) {
	s.Stop()
	_, _ = fmt.Fprintf(s.output, "%s %s\n", ActiveIcons.Error, s.errorStyle.Render(message))
}

// Skip stops the spinner and displays a skipped message
func (s *Spinner) Skip(message string) {
	s.Stop()
	_, _ = fmt.Fprintf(s.output, "%s %s\n", ActiveIcons.Skip, s.skipStyle.Render(message))
}

// Warn stops the spinner and displays a warning message
func (s *Spinner) Warn(message string) {
	s.Stop()
	_, _ = fmt.Fprintf(s.output, "%s %s\n", ActiveIcons.Warning, s.warningStyle.Render(message))
}
