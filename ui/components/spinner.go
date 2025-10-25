package components

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Spinner displays an animated spinner for loading states.
type Spinner struct {
	spinner spinner.Model
	label   string
	active  bool
	theme   SpinnerTheme
}

// SpinnerTheme defines the color scheme for the spinner.
type SpinnerTheme struct {
	Spinner lipgloss.Color
	Label   lipgloss.Color
}

// DefaultSpinnerTheme returns the default spinner theme.
func DefaultSpinnerTheme() SpinnerTheme {
	return SpinnerTheme{
		Spinner: lipgloss.Color("#7aa2f7"),
		Label:   lipgloss.Color("#c0caf5"),
	}
}

// NewSpinner creates a new spinner component.
func NewSpinner(label string) Spinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))

	return Spinner{
		spinner: s,
		label:   label,
		active:  false,
		theme:   DefaultSpinnerTheme(),
	}
}

// Update handles messages (Bubble Tea Update method).
func (s *Spinner) Update(msg tea.Msg) (*Spinner, tea.Cmd) {
	if !s.active {
		return s, nil
	}

	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return s, cmd
}

// View renders the spinner.
func (s *Spinner) View() string {
	if !s.active {
		return ""
	}

	spinnerStyle := lipgloss.NewStyle().Foreground(s.theme.Spinner)
	labelStyle := lipgloss.NewStyle().Foreground(s.theme.Label)

	spinnerView := spinnerStyle.Render(s.spinner.View())
	labelView := labelStyle.Render(s.label)

	return lipgloss.JoinHorizontal(lipgloss.Left, spinnerView, " ", labelView)
}

// Start starts the spinner animation.
func (s *Spinner) Start() tea.Cmd {
	s.active = true
	return s.spinner.Tick
}

// Stop stops the spinner animation.
func (s *Spinner) Stop() {
	s.active = false
}

// IsActive returns whether the spinner is active.
func (s *Spinner) IsActive() bool {
	return s.active
}

// SetLabel sets the label text.
func (s *Spinner) SetLabel(label string) {
	s.label = label
}

// SetStyle sets the spinner style (dot, line, moon, etc.).
func (s *Spinner) SetStyle(style spinner.Spinner) {
	s.spinner.Spinner = style
}

// SetTheme sets the color theme for the spinner.
func (s *Spinner) SetTheme(theme SpinnerTheme) {
	s.theme = theme
	s.spinner.Style = lipgloss.NewStyle().Foreground(theme.Spinner)
}
