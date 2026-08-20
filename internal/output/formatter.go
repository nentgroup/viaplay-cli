// Package output provides formatting utilities for CLI output.
// It handles consistent styling, colours, and layout for terminal output.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Colour printers for consistent colours
var (
	// Base colours
	Success   = color.New(color.FgGreen).SprintFunc()
	Error     = color.New(color.FgRed).SprintFunc()
	Warning   = color.New(color.FgYellow).SprintFunc()
	Info      = color.New(color.FgCyan).SprintFunc()
	Primary   = color.New(color.FgBlue).SprintFunc()
	Secondary = color.New(color.FgMagenta).SprintFunc()
	Faint     = color.New(color.FgHiBlack).SprintFunc()
	Bold      = color.New(color.Bold).SprintFunc()
	Underline = color.New(color.Underline).SprintFunc()
	Inverse   = color.New(color.FgBlack, color.BgWhite).SprintFunc()

	// Compound styles
	SuccessBold   = color.New(color.FgGreen, color.Bold).SprintFunc()
	ErrorBold     = color.New(color.FgRed, color.Bold).SprintFunc()
	WarningBold   = color.New(color.FgYellow, color.Bold).SprintFunc()
	InfoBold      = color.New(color.FgCyan, color.Bold).SprintFunc()
	PrimaryBold   = color.New(color.FgBlue, color.Bold).SprintFunc()
	SecondaryBold = color.New(color.FgMagenta, color.Bold).SprintFunc()

	// Function variants
	Successf   = color.New(color.FgGreen).SprintfFunc()
	Errorf     = color.New(color.FgRed).SprintfFunc()
	Warningf   = color.New(color.FgYellow).SprintfFunc()
	Infof      = color.New(color.FgCyan).SprintfFunc()
	Primaryf   = color.New(color.FgBlue).SprintfFunc()
	Secondaryf = color.New(color.FgMagenta).SprintfFunc()
	Faintf     = color.New(color.FgHiBlack).SprintfFunc()

	DebugColor = color.New(color.FgHiMagenta).SprintFunc()
)

// EnableColors toggles colour output - useful for testing or when running in environments that don't support colours
func EnableColors(enabled bool) {
	color.NoColor = !enabled
}

// SuccessMessage prints a formatted success message
func SuccessMessage(message string) {
	fmt.Printf("\n%s %s\n", ActiveIcons.Success, SuccessBold(message))
}

// ErrorMessage prints a formatted error message
func ErrorMessage(message string) {
	fmt.Printf("\n%s %s\n", ActiveIcons.Error, ErrorBold(message))
}

// WarningMessage prints a formatted warning message
func WarningMessage(message string) {
	fmt.Printf("\n%s %s\n", ActiveIcons.Warning, WarningBold(message))
}

// InfoMessage prints a formatted info message
func InfoMessage(message string) {
	// Always add a space after the emoji for better terminal spacing
	fmt.Printf("\n%s %s\n", ActiveIcons.Info, InfoBold(message))
}

// ProcessingMessage prints a message indicating an operation is in progress
func ProcessingMessage(message string) {
	fmt.Printf("\n%s %s\n", ActiveIcons.Process, message)
}

// AuthMessage prints an authentication-related message
func AuthMessage(message string) {
	fmt.Printf("\n%s %s\n", ActiveIcons.Key, PrimaryBold(message))
}

// GitHubMessage prints a GitHub-related message
func GitHubMessage(message string) {
	fmt.Printf("\n%s %s\n", ActiveIcons.GitHub, PrimaryBold(message))
}

// ConfigMessage prints a configuration-related message
func ConfigMessage(message string) {
	fmt.Printf("\n%s %s\n", ActiveIcons.Config, InfoBold(message))
}

// CacheMessage prints a cache-related message
func CacheMessage(message string) {
	fmt.Printf("\n%s %s\n", ActiveIcons.Cache, InfoBold(message))
}

// Section prints a section header
func Section(title string) {
	fmt.Printf("\n%s %s\n%s\n", ActiveIcons.Brain, Bold(title), strings.Repeat("─", len(title)+2))
}

// FatalError prints an error message and exits the program
func FatalError(msg string) {
	ErrorMessage(msg)
	os.Exit(1)
}

// Table formats data as a table with aligned columns
func Table(headers []string, rows [][]string, indent int) string {
	if len(rows) == 0 {
		return ""
	}

	// Calculate column widths based on visible content
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	// Calculate maximum visible width for each column
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				// Get the visual width by stripping ANSI colour codes
				visualWidth := stripANSI(cell)
				if len(visualWidth) > colWidths[i] {
					colWidths[i] = len(visualWidth)
				}
			}
		}
	}

	// Create the table
	var sb strings.Builder
	indentStr := strings.Repeat(" ", indent)

	// Write headers
	sb.WriteString(indentStr)
	for i, header := range headers {
		format := fmt.Sprintf("%%-%ds", colWidths[i]+2)
		sb.WriteString(Bold(fmt.Sprintf(format, header)))
	}
	sb.WriteString("\n")

	// Write separator
	sb.WriteString(indentStr)
	for i := range headers {
		sb.WriteString(strings.Repeat("─", colWidths[i]+1) + " ")
	}
	sb.WriteString("\n")

	// Write rows
	for _, row := range rows {
		sb.WriteString(indentStr)
		for i, cell := range row {
			if i < len(colWidths) {
				// Calculate padding based on visible text length
				visibleLen := len(stripANSI(cell))
				padding := colWidths[i] - visibleLen + 2

				// Write the cell with its ANSI colours intact
				sb.WriteString(cell)

				// Add appropriate padding after the cell
				sb.WriteString(strings.Repeat(" ", padding))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// stripANSI removes ANSI colour sequences from a string to get its visual length
func stripANSI(str string) string {
	// ANSI escape sequence regex pattern
	ansiPattern := "\033\\[[0-9;]*m"
	re := regexp.MustCompile(ansiPattern)
	return re.ReplaceAllString(str, "")
}

// List creates a formatted list with bullets or numbers
func List(items []string, numbered bool, indent int) string {
	if len(items) == 0 {
		return ""
	}

	var sb strings.Builder
	indentStr := strings.Repeat(" ", indent)

	for i, item := range items {
		if numbered {
			fmt.Fprintf(&sb, "%s%d. %s\n", indentStr, i+1, item)
		} else {
			fmt.Fprintf(&sb, "%s%s %s\n", indentStr, ActiveIcons.Bullet, item)
		}
	}

	return sb.String()
}

// KeyValue formats a key-value pair with consistent spacing
func KeyValue(key, value string, indent int) string {
	indentStr := strings.Repeat(" ", indent)
	return fmt.Sprintf("%s%s: %s\n", indentStr, Bold(key), value)
}

// KeyValueTable creates a table of key-value pairs
func KeyValueTable(pairs map[string]string, indent int) string {
	if len(pairs) == 0 {
		return ""
	}

	// Find the longest key for alignment
	maxKeyLen := 0
	for key := range pairs {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	var sb strings.Builder
	indentStr := strings.Repeat(" ", indent)
	format := fmt.Sprintf("%%s%%-%ds  %%s\n", maxKeyLen)

	// Sort keys for consistent output
	keys := make([]string, 0, len(pairs))
	for key := range pairs {
		keys = append(keys, key)
	}
	// We could sort keys here if needed

	for _, key := range keys {
		fmt.Fprintf(&sb, format, indentStr, Bold(key), pairs[key])
	}

	return sb.String()
}

// Duration formats a time duration in a human-readable way
func Duration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1f seconds", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1f minutes", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	}
	return fmt.Sprintf("%.1f days", d.Hours()/24)
}

// FormatSize formats a byte size in a human-readable way
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// PrintIndentf prints a message with indentation
func PrintIndentf(indent int, format string, a ...interface{}) {
	indentStr := strings.Repeat(" ", indent)
	fmt.Printf("%s"+format+"\n", append([]interface{}{indentStr}, a...)...)
}

// Banner prints a banner with the application name
func Banner() {
	banner := `
██╗   ██╗██╗ █████╗ ██████╗ ██╗      █████╗ ██╗   ██╗      ██████╗██╗     ██╗
██║   ██║██║██╔══██╗██╔══██╗██║     ██╔══██╗╚██╗ ██╔╝     ██╔════╝██║     ██║
██║   ██║██║███████║██████╔╝██║     ███████║ ╚████╔╝█████╗██║     ██║     ██║
╚██╗ ██╔╝██║██╔══██║██╔═══╝ ██║     ██╔══██║  ╚██╔╝ ╚════╝██║     ██║     ██║
 ╚████╔╝ ██║██║  ██║██║     ███████╗██║  ██║   ██║        ╚██████╗███████╗██║
  ╚═══╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚══════╝╚═╝  ╚═╝   ╚═╝         ╚═════╝╚══════╝╚═╝
`
	fmt.Println(Primary(banner))
}

// Divider prints a horizontal divider line
func Divider() {
	fmt.Println(Faint(strings.Repeat("─", 80)))
}

// ProgressBar renders a simple progress bar
// current and total define the progress state
// width defines the visual width of the progress bar
func ProgressBar(current, total, width int) string {
	if total <= 0 || current <= 0 || current > total {
		return ""
	}

	percentage := float64(current) / float64(total)
	completed := int(percentage * float64(width))

	var sb strings.Builder
	sb.WriteString("[")
	sb.WriteString(strings.Repeat("█", completed))
	sb.WriteString(strings.Repeat(" ", width-completed))
	sb.WriteString("]")
	fmt.Fprintf(&sb, " %.1f%%", percentage*100)

	return sb.String()
}

var verboseEnabled bool

// SetVerbose enables or disables verbose output
func SetVerbose(enabled bool) {
	verboseEnabled = enabled
}

// VerboseMessage prints a message only if verbose mode is enabled
func VerboseMessage(message interface{}) {
	if verboseEnabled {
		var formatted string
		switch v := message.(type) {
		case string:
			formatted = v
		case fmt.Stringer:
			formatted = v.String()
		default:
			// Pretty-print structs/maps as JSON indented
			if b, err := json.MarshalIndent(v, "    ", "  "); err == nil {
				formatted = string(b)
			} else {
				formatted = fmt.Sprintf("%+v", v)
			}
		}
		fmt.Fprintf(os.Stderr, "\n%s %s %s\n", ActiveIcons.Debug, DebugColor("[DEBUG]"), formatted)
	}
}

// IsVerboseEnabled returns whether verbose mode is enabled
func IsVerboseEnabled() bool {
	return verboseEnabled
}
