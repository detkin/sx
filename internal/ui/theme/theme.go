// Package theme provides theming support for the Skills CLI UI.
package theme

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// ColorPalette defines the colors used by a theme.
type ColorPalette struct {
	// Primary accent color (e.g., cyan for Claude Code)
	Primary lipgloss.AdaptiveColor
	// Secondary accent color (e.g., blue)
	Secondary lipgloss.AdaptiveColor

	// Status colors
	Success lipgloss.AdaptiveColor
	Error   lipgloss.AdaptiveColor
	Warning lipgloss.AdaptiveColor
	Info    lipgloss.AdaptiveColor

	// Text colors
	Text         lipgloss.AdaptiveColor
	TextMuted    lipgloss.AdaptiveColor
	TextFaint    lipgloss.AdaptiveColor
	TextEmphasis lipgloss.AdaptiveColor

	// UI element colors
	Border    lipgloss.AdaptiveColor
	Highlight lipgloss.AdaptiveColor
}

// Symbols defines the glyphs used for various states.
type Symbols struct {
	Success    string
	Error      string
	Warning    string
	Info       string
	Arrow      string
	Bullet     string
	Pending    string
	InProgress string
}

// Styles contains pre-composed lipgloss styles.
type Styles struct {
	// Message styles
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style

	// Layout styles
	Header    lipgloss.Style
	SubHeader lipgloss.Style

	// Text styles
	Bold     lipgloss.Style
	Muted    lipgloss.Style
	Faint    lipgloss.Style
	Emphasis lipgloss.Style

	// List styles
	ListItem   lipgloss.Style
	ListBullet lipgloss.Style
	Selected   lipgloss.Style
	Cursor     lipgloss.Style

	// Key-Value styles
	Key       lipgloss.Style
	Value     lipgloss.Style
	Separator lipgloss.Style

	// Progress/status styles
	Spinner  lipgloss.Style
	Progress lipgloss.Style
}

// Theme defines the visual styling for the CLI.
type Theme interface {
	// Name returns the theme identifier
	Name() string
	// Palette returns the color palette
	Palette() ColorPalette
	// Styles returns pre-composed lipgloss styles
	Styles() Styles
	// Symbols returns the glyphs used for various states
	Symbols() Symbols
}

var (
	currentTheme Theme
	themeMu      sync.RWMutex
)

func init() {
	// Set default theme
	currentTheme = NewClaudeCodeTheme()
}

// Current returns the active theme (thread-safe).
func Current() Theme {
	themeMu.RLock()
	defer themeMu.RUnlock()
	return currentTheme
}

// Set sets the active theme (thread-safe).
func Set(t Theme) {
	themeMu.Lock()
	defer themeMu.Unlock()
	currentTheme = t
}
