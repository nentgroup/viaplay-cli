package scaffolding

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HookOutputModel is a Bubble Tea model for displaying hook output
type HookOutputModel struct {
	viewport    viewport.Model
	content     strings.Builder
	ready       bool
	done        bool
	err         error
	title       string
	width       int
	height      int
	borderColor lipgloss.Color
}

// NewHookOutputModel creates a new model for displaying hook output
func NewHookOutputModel(title string) *HookOutputModel {
	return &HookOutputModel{
		content:     strings.Builder{},
		title:       title,
		width:       80,                        // Default width
		height:      20,                        // Default height
		borderColor: lipgloss.Color("#E6007A"), // Viaplay pink color
	}
}

// Init initializes the model
func (m *HookOutputModel) Init() tea.Cmd {
	// No initialization commands needed
	return nil
}

// Update updates the model based on messages
func (m *HookOutputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle key presses
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		// Handle window resize
		headerHeight := 4 // Space for header and footer
		m.width = msg.Width
		m.height = msg.Height - headerHeight
		if !m.ready {
			m.viewport = viewport.New(m.width, m.height)
			m.viewport.Style = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(m.borderColor)

			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = m.height
		}
		m.viewport.SetContent(m.content.String())

	case hookOutputMsg:
		// Append new output
		m.content.WriteString(string(msg))
		if m.ready {
			m.viewport.SetContent(m.content.String())
			// Auto-scroll to bottom
			m.viewport.GotoBottom()
		}

	case hookDoneMsg:
		// Mark as done when hooks are complete
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	}

	// Only pass through viewport commands if we're ready
	if m.ready {
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

// View renders the model
func (m *HookOutputModel) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(m.borderColor).
		Bold(true).
		Padding(0, 1)

	// Status line at the bottom
	statusMsg := "Press q to quit"
	if m.done {
		if m.err != nil {
			statusMsg = fmt.Sprintf("Hooks failed: %v (press q to quit)", m.err)
		} else {
			statusMsg = "Hooks completed successfully (press q to quit)"
		}
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Italic(true).
		Padding(0, 1)

	// Combine all parts
	return fmt.Sprintf("%s\n%s\n%s",
		titleStyle.Render(m.title),
		m.viewport.View(),
		statusStyle.Render(statusMsg))
}

// Custom messages for the Bubble Tea model
type hookOutputMsg string
type hookDoneMsg struct {
	err error
}

// OutputWriter is a writer that sends output to the Bubble Tea model
type OutputWriter struct {
	program *tea.Program
}

// Write implements io.Writer
func (w *OutputWriter) Write(p []byte) (n int, err error) {
	// Send output to the Bubble Tea model
	w.program.Send(hookOutputMsg(p))
	return len(p), nil
}

// DisplayHookOutputWithTUI runs the hook and displays its output in a TUI
func DisplayHookOutputWithTUI(title string, runHookFn func(stdout, stderr io.Writer) error) error {
	// Create a model
	model := NewHookOutputModel(title)

	// Initialize the program
	p := tea.NewProgram(model)

	// Create writer for output
	writer := &OutputWriter{program: p}

	// Run the hook in a goroutine
	go func() {
		err := runHookFn(writer, writer)

		// Signal completion
		time.Sleep(100 * time.Millisecond) // Give time for final output to process
		p.Send(hookDoneMsg{err: err})
	}()

	// Run the UI
	_, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
	}

	// Return the error from the hook, if any
	return model.err
}
