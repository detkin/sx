package components

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/sleuth-io/skills/internal/ui"
	"github.com/sleuth-io/skills/internal/ui/theme"
)

// confirmModel is the bubbletea model for the confirm component.
type confirmModel struct {
	message    string
	confirmed  bool
	defaultYes bool
	done       bool
	theme      theme.Theme
}

// confirmKeyMap defines the keybindings for the confirm component.
type confirmKeyMap struct {
	Yes    key.Binding
	No     key.Binding
	Toggle key.Binding
	Submit key.Binding
	Quit   key.Binding
}

var confirmKeys = confirmKeyMap{
	Yes: key.NewBinding(
		key.WithKeys("y", "Y"),
		key.WithHelp("y", "yes"),
	),
	No: key.NewBinding(
		key.WithKeys("n", "N"),
		key.WithHelp("n", "no"),
	),
	Toggle: key.NewBinding(
		key.WithKeys("left", "right", "h", "l", "tab"),
		key.WithHelp("←/→", "toggle"),
	),
	Submit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "esc"),
		key.WithHelp("q", "quit"),
	),
}

func newConfirmModel(message string, defaultYes bool) confirmModel {
	return confirmModel{
		message:    message,
		confirmed:  defaultYes,
		defaultYes: defaultYes,
		theme:      theme.Current(),
	}
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, confirmKeys.Quit):
			m.confirmed = false
			m.done = true
			return m, tea.Quit

		case key.Matches(msg, confirmKeys.Yes):
			m.confirmed = true
			m.done = true
			return m, tea.Quit

		case key.Matches(msg, confirmKeys.No):
			m.confirmed = false
			m.done = true
			return m, tea.Quit

		case key.Matches(msg, confirmKeys.Toggle):
			m.confirmed = !m.confirmed

		case key.Matches(msg, confirmKeys.Submit):
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m confirmModel) View() string {
	if m.done {
		return ""
	}

	styles := m.theme.Styles()

	var yes, no string
	if m.confirmed {
		yes = styles.Selected.Render("[Yes]")
		no = styles.Muted.Render(" No ")
	} else {
		yes = styles.Muted.Render(" Yes ")
		no = styles.Selected.Render("[No]")
	}

	return fmt.Sprintf("%s %s %s", m.message, yes, no)
}

// Confirm displays an interactive confirmation prompt.
// Returns true for yes, false for no.
// Falls back to Y/n prompt for non-TTY environments.
func Confirm(message string, defaultYes bool) (bool, error) {
	return ConfirmWithIO(message, defaultYes, os.Stdin, os.Stdout)
}

// ConfirmWithIO displays an interactive confirmation prompt using custom IO.
func ConfirmWithIO(message string, defaultYes bool, in io.Reader, out io.Writer) (bool, error) {
	// Fall back to simple prompt for non-TTY
	if !ui.IsStdoutTTY() || !ui.IsStdinTTY() {
		return confirmSimple(message, defaultYes, in, out)
	}

	m := newConfirmModel(message, defaultYes)
	p := tea.NewProgram(m, tea.WithOutput(out))

	result, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("confirm failed: %w", err)
	}

	final := result.(confirmModel)
	return final.confirmed, nil
}

// confirmSimple provides a simple Y/n fallback for non-TTY environments.
func confirmSimple(message string, defaultYes bool, in io.Reader, out io.Writer) (bool, error) {
	hint := "(y/N)"
	if defaultYes {
		hint = "(Y/n)"
	}

	fmt.Fprintf(out, "%s %s: ", message, hint)

	reader := bufio.NewReader(in)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultYes, nil
	}

	switch input {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return defaultYes, nil
	}
}
