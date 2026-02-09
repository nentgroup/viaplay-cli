// Package output provides formatting utilities for CLI output.
package output

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// IconType represents the type of icons to use
type IconType int

const (
	// IconTypeUnicode uses plain Unicode characters
	IconTypeUnicode IconType = iota
	// IconTypeNerdFont uses Nerd Font icons when available
	IconTypeNerdFont
)

// currentIconType stores the detected or configured icon type
var currentIconType = IconTypeUnicode

// Icons holds all icon definitions with fallbacks
type Icons struct {
	// Status icons
	Success  string
	Error    string
	Warning  string
	Info     string
	Process  string
	Skip     string
	Debug    string
	Question string

	// Content type icons
	Key       string
	GitHub    string
	Template  string
	Brain     string
	Config    string
	Cache     string
	Clock     string
	Checkmark string
	Cross     string
	Bullet    string

	// Additional icons for summary display
	Clipboard string
	People    string
	Globe     string
	Lock      string
	Summary   string

	// Language icons
	LangGo      string
	LangNode    string
	LangRust    string
	LangPython  string
	LangJava    string
	LangGeneric string
}

// UnicodeIcons defines basic Unicode fallback icons
var UnicodeIcons = Icons{
	// Status icons
	Success:  "✓", // CHECK MARK (U+2713)
	Error:    "✗", // MULTIPLICATION X (U+2717)
	Warning:  "⚠", // WARNING SIGN (U+26A0)
	Info:     "ℹ", // INFORMATION SOURCE (U+2139)
	Process:  "↻", // ANTICLOCKWISE OPEN CIRCLE ARROW (U+21BB)
	Skip:     "↷", // RIGHTWARDS ARROW WITH HOOK (U+21B7)
	Debug:    "⬤", // BLACK CIRCLE (U+2B24)
	Question: "?", // QUESTION MARK

	// Content type icons
	Key:       "⚿", // KEY (U+26BF)
	GitHub:    "⟳", // CLOCKWISE GAPPED CIRCLE ARROW (U+27F3)
	Template:  "⊞", // SQUARED PLUS (U+229E)
	Brain:     "⌘", // PLACE OF INTEREST SIGN (U+2318)
	Config:    "⚙", // GEAR (U+2699)
	Cache:     "◉", // FISHEYE (U+25C9)
	Clock:     "⏱", // STOPWATCH (U+23F1)
	Checkmark: "✓", // CHECK MARK (U+2713)
	Cross:     "✗", // MULTIPLICATION X (U+2717)
	Bullet:    "•", // BULLET (U+2022)
	Clipboard: "📋", // CLIPBOARD (U+1F4CB)
	People:    "👥", // TWO MEN HOLDING HANDS (U+1F46B)
	Globe:     "🌐", // GLOBE WITH MERIDIANS (U+1F310)
	Lock:      "🔒", // LOCK (U+1F512)
	Summary:   "📝", // MEMO (U+1F4DD)

	// Language icons
	LangGo:      "G", // Simple "G" for Go
	LangNode:    "N", // Simple "N" for Node
	LangRust:    "R", // Simple "R" for Rust
	LangPython:  "P", // Simple "P" for Python
	LangJava:    "J", // Simple "J" for Java
	LangGeneric: "⊡", // SQUARED DOT OPERATOR (U+22A1) for generic
}

// NerdFontIcons defines icons using Nerd Font glyphs
var NerdFontIcons = Icons{
	// Status icons
	Success:  "\uf00c", // nf-fa-check (f00c)
	Error:    "\uf00d", // nf-fa-times (f00d)
	Warning:  "\uf071", // nf-fa-warning (f071)
	Info:     "\uf05a", // nf-fa-info_circle (f05a)
	Process:  "\uf021", // nf-fa-refresh (f021)
	Skip:     "\uf04e", // nf-fa-forward (f04e)
	Debug:    "\uf188", // nf-fa-bug (f188)
	Question: "\uf059", // nf-fa-question_circle (f059)

	// Content type icons
	Key:       "\uf084", // nf-fa-key (f084)
	GitHub:    "\uf09b", // nf-fa-github (f09b)
	Template:  "\uf15c", // nf-fa-file_text (f15c)
	Brain:     "\uf0eb", // nf-fa-lightbulb_o (f0eb)
	Config:    "\uf013", // nf-fa-cog (f013)
	Cache:     "\uf1c0", // nf-fa-database (f1c0)
	Clock:     "\uf017", // nf-fa-clock_o (f017)
	Checkmark: "\uf00c", // nf-fa-check (f00c)
	Cross:     "\uf00d", // nf-fa-times (f00d)
	Bullet:    "\uf111", // nf-fa-circle (f111)
	Clipboard: "\uf0ea", // nf-fa-clipboard (f0ea)
	People:    "\uf0c0", // nf-fa-users (f0c0)
	Globe:     "\uf0ac", // nf-fa-globe (f0ac)
	Lock:      "\uf023", // nf-fa-lock (f023)
	Summary:   "\uf15d", // nf-fa-file_alt (f15d)

	// Language icons
	LangGo:      "\ue65e", // nf-dev-go (e65e)
	LangNode:    "\ue718", // nf-dev-nodejs (e718)
	LangRust:    "\ue7a8", // nf-dev-rust (e7a8)
	LangPython:  "\ue73c", // nf-dev-python (e73c)
	LangJava:    "\ue738", // nf-dev-java (e738)
	LangGeneric: "\uf121", // nf-fa-code (f121)
}

// ColoredIcons defines colour emoji icons (mostly for compatibility)
var ColoredIcons = Icons{
	// Status icons
	Success:  "✅",
	Error:    "❌",
	Warning:  "⚠️",
	Info:     "💡",
	Process:  "🔄",
	Skip:     "⏭️",
	Debug:    "🐞",
	Question: "❓",

	// Content type icons
	Key:       "🔑",
	GitHub:    "🚀",
	Template:  "📄",
	Brain:     "🧠",
	Config:    "⚙️",
	Cache:     "📦",
	Clock:     "⏱️",
	Checkmark: "✓",
	Cross:     "✗",
	Bullet:    "•",
	Clipboard: "📋",
	People:    "👥",
	Globe:     "🌐",
	Lock:      "🔒",
	Summary:   "📝",

	// Language icons
	LangGo:      "🔷", // Blue diamond for Go
	LangNode:    "🟢", // Green circle for Node
	LangRust:    "🦀", // Crab for Rust
	LangPython:  "🐍", // Snake for Python
	LangJava:    "☕", // Coffee for Java
	LangGeneric: "📝", // Memo for generic
}

// ActiveIcons holds the current set of icons to use
var ActiveIcons = UnicodeIcons

// ShouldUseNerdFonts determines if Nerd Fonts should be used
func ShouldUseNerdFonts() bool {
	// Check environment variables that indicate terminal support
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	colorTerm := os.Getenv("COLORTERM")

	// Check for WSL
	isWSL := false
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/version"); err == nil {
			isWSL = strings.Contains(strings.ToLower(string(data)), "microsoft")
		}
	}

	// Common terminals known to support Nerd Fonts
	supportedTerms := []string{
		"kitty", "alacritty", "wezterm", "iterm", "iterm2",
		"vscode", "gnome-terminal", "konsole", "terminator",
	}

	// Look for known terminal indicators
	if termProgram != "" {
		termProgramLower := strings.ToLower(termProgram)
		for _, supported := range supportedTerms {
			if strings.Contains(termProgramLower, supported) {
				return true
			}
		}
	}

	// iTerm with special configurations
	if os.Getenv("ITERM_SESSION_ID") != "" {
		return true
	}

	// VS Code terminal usually supports Nerd Fonts
	if strings.Contains(strings.ToLower(term), "vscode") {
		return true
	}

	// Special case for some terminals that self-identify as supporting advanced features
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		return true
	}

	// Windows Terminal tends to have good font support
	if os.Getenv("WT_SESSION") != "" ||
		(runtime.GOOS == "windows" && !isWSL) {
		return true
	}

	// Check for explicit environment variable
	if os.Getenv("VIAPLAY_CLI_NERD_FONTS") == "1" {
		return true
	}

	return false
}

// InitIcons initialises the icons based on terminal capabilities
func InitIcons() {
	// Check if the VIAPLAY_CLI_ICON_TYPE environment variable is set
	iconTypeEnv := os.Getenv("VIAPLAY_CLI_ICON_TYPE")
	switch strings.ToLower(iconTypeEnv) {
	case "unicode":
		currentIconType = IconTypeUnicode
		ActiveIcons = UnicodeIcons
	case "nerdfont":
		currentIconType = IconTypeNerdFont
		ActiveIcons = NerdFontIcons
	case "emoji", "colored":
		ActiveIcons = ColoredIcons
	default:
		// Auto-detect based on terminal capabilities
		if ShouldUseNerdFonts() {
			currentIconType = IconTypeNerdFont
			ActiveIcons = NerdFontIcons
		} else {
			currentIconType = IconTypeUnicode
			ActiveIcons = UnicodeIcons
		}
	}

	// Force initialization of icons if they're somehow not set
	if ActiveIcons.Success == "" || ActiveIcons.Error == "" {
		// This is a fail-safe to ensure we always have visible icons
		ActiveIcons = UnicodeIcons
	}
}

// ForceIconType forces a specific icon type
func ForceIconType(iconType IconType) {
	currentIconType = iconType
	switch iconType {
	case IconTypeNerdFont:
		ActiveIcons = NerdFontIcons
	case IconTypeUnicode:
		ActiveIcons = UnicodeIcons
	}
}

// GetCurrentIconType returns the current icon type
func GetCurrentIconType() IconType {
	return currentIconType
}

// String gets an icon string with a trailing space for easy formatting
//
//nolint:cyclop
func (i Icons) String(name string) string {
	var icon string

	switch name {
	case "success":
		icon = i.Success
	case "error":
		icon = i.Error
	case "warning":
		icon = i.Warning
	case "info":
		icon = i.Info
	case "process":
		icon = i.Process
	case "skip":
		icon = i.Skip
	case "debug":
		icon = i.Debug
	case "question":
		icon = i.Question
	case "key":
		icon = i.Key
	case "github":
		icon = i.GitHub
	case "template":
		icon = i.Template
	case "brain":
		icon = i.Brain
	case "config":
		icon = i.Config
	case "cache":
		icon = i.Cache
	case "clock":
		icon = i.Clock
	case "checkmark":
		icon = i.Checkmark
	case "cross":
		icon = i.Cross
	case "bullet":
		icon = i.Bullet
	case "clipboard":
		icon = i.Clipboard
	case "people":
		icon = i.People
	case "globe":
		icon = i.Globe
	case "lock":
		icon = i.Lock
	case "summary":
		icon = i.Summary
	// Language icons
	case "lang-go":
		icon = i.LangGo
	case "lang-node":
		icon = i.LangNode
	case "lang-rust":
		icon = i.LangRust
	case "lang-python":
		icon = i.LangPython
	case "lang-java":
		icon = i.LangJava
	case "lang-generic":
		icon = i.LangGeneric
	default:
		return ""
	}

	return fmt.Sprintf("%s ", icon)
}

// Icon gets an icon string with a trailing space for easy formatting
func Icon(name string) string {
	return ActiveIcons.String(name)
}

// GetLanguageIcon returns the appropriate icon for a programming language
func GetLanguageIcon(language string) string {
	language = strings.ToLower(language)

	switch language {
	case "go":
		return Primary(Icon("lang-go"))
	case "node", "nodejs", "typescript", "javascript":
		return Success(Icon("lang-node"))
	case "rust":
		return Secondary(Icon("lang-rust"))
	case "python":
		return Info(Icon("lang-python"))
	case "java":
		return Warning(Icon("lang-java"))
	default:
		return Icon("lang-generic")
	}
}

func init() {
	// Initialise icons when the package is imported
	InitIcons()
}
