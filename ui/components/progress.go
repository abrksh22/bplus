package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressBar displays a progress bar for long-running operations.
type ProgressBar struct {
	progress progress.Model
	current  int
	total    int
	label    string
	mode     string // "determinate" or "indeterminate"
	theme    ProgressTheme
}

// ProgressTheme defines the color scheme for the progress bar.
type ProgressTheme struct {
	Full  lipgloss.Color
	Empty lipgloss.Color
	Label lipgloss.Color
}

// DefaultProgressTheme returns the default progress theme.
func DefaultProgressTheme() ProgressTheme {
	return ProgressTheme{
		Full:  lipgloss.Color("#7aa2f7"),
		Empty: lipgloss.Color("#3b4261"),
		Label: lipgloss.Color("#c0caf5"),
	}
}

// NewProgressBar creates a new progress bar component.
func NewProgressBar(total int) ProgressBar {
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 40

	mode := "determinate"
	if total <= 0 {
		mode = "indeterminate"
	}

	return ProgressBar{
		progress: p,
		current:  0,
		total:    total,
		label:    "",
		mode:     mode,
		theme:    DefaultProgressTheme(),
	}
}

// Update handles messages (Bubble Tea Update method).
func (p *ProgressBar) Update(msg tea.Msg) (*ProgressBar, tea.Cmd) {
	var cmd tea.Cmd
	var model tea.Model
	model, cmd = p.progress.Update(msg)
	p.progress = model.(progress.Model)
	return p, cmd
}

// View renders the progress bar.
func (p *ProgressBar) View() string {
	labelStyle := lipgloss.NewStyle().Foreground(p.theme.Label)

	var progressView string
	if p.mode == "indeterminate" {
		// For indeterminate mode, show a pulsing bar
		progressView = p.progress.ViewAs(0.5) // Show at 50%
	} else {
		// For determinate mode, show actual progress
		percent := float64(p.current) / float64(p.total)
		progressView = p.progress.ViewAs(percent)
	}

	// Build the display
	var display string
	if p.label != "" {
		labelView := labelStyle.Render(p.label)
		display = lipgloss.JoinVertical(lipgloss.Left, labelView, progressView)
	} else {
		display = progressView
	}

	// Add percentage/count if determinate
	if p.mode == "determinate" {
		percent := float64(p.current) / float64(p.total) * 100
		info := labelStyle.Render(fmt.Sprintf("%.0f%% (%d/%d)", percent, p.current, p.total))
		display = lipgloss.JoinVertical(lipgloss.Left, display, info)
	}

	return display
}

// SetProgress sets the current progress.
func (p *ProgressBar) SetProgress(current int) {
	p.current = current
	if p.current > p.total {
		p.current = p.total
	}
}

// Increment increments the progress by 1.
func (p *ProgressBar) Increment() {
	p.current++
	if p.current > p.total {
		p.current = p.total
	}
}

// SetTotal sets the total value.
func (p *ProgressBar) SetTotal(total int) {
	p.total = total
	if total <= 0 {
		p.mode = "indeterminate"
	} else {
		p.mode = "determinate"
	}
}

// SetLabel sets the label text.
func (p *ProgressBar) SetLabel(label string) {
	p.label = label
}

// SetWidth sets the width of the progress bar.
func (p *ProgressBar) SetWidth(width int) {
	p.progress.Width = width
}

// SetMode sets the mode (determinate or indeterminate).
func (p *ProgressBar) SetMode(mode string) {
	p.mode = mode
}

// SetTheme sets the color theme for the progress bar.
func (p *ProgressBar) SetTheme(theme ProgressTheme) {
	p.theme = theme
}

// GetProgress returns the current progress.
func (p *ProgressBar) GetProgress() (current, total int) {
	return p.current, p.total
}

// GetPercentage returns the progress as a percentage (0-100).
func (p *ProgressBar) GetPercentage() float64 {
	if p.total <= 0 {
		return 0
	}
	return float64(p.current) / float64(p.total) * 100
}

// IsComplete returns whether the progress is complete.
func (p *ProgressBar) IsComplete() bool {
	return p.current >= p.total && p.total > 0
}

// Reset resets the progress to 0.
func (p *ProgressBar) Reset() {
	p.current = 0
}
