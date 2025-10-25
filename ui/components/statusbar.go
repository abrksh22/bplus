package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusBar displays status information at the top or bottom of the screen.
type StatusBar struct {
	model       string
	tokensIn    int
	tokensOut   int
	cost        float64
	connected   bool
	processing  bool
	mode        string // "Fast" or "Thorough"
	theme       string
	width       int
	customLeft  string
	customRight string
	styleTheme  StatusBarTheme
}

// StatusBarTheme defines the color scheme for the status bar.
type StatusBarTheme struct {
	Background lipgloss.Color
	Foreground lipgloss.Color
	Highlight  lipgloss.Color
	Dim        lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
}

// DefaultStatusBarTheme returns the default status bar theme.
func DefaultStatusBarTheme() StatusBarTheme {
	return StatusBarTheme{
		Background: lipgloss.Color("#1a1b26"),
		Foreground: lipgloss.Color("#c0caf5"),
		Highlight:  lipgloss.Color("#7aa2f7"),
		Dim:        lipgloss.Color("#565f89"),
		Success:    lipgloss.Color("#9ece6a"),
		Warning:    lipgloss.Color("#e0af68"),
		Error:      lipgloss.Color("#f7768e"),
	}
}

// NewStatusBar creates a new status bar.
func NewStatusBar(width int) StatusBar {
	return StatusBar{
		model:      "No model selected",
		tokensIn:   0,
		tokensOut:  0,
		cost:       0.0,
		connected:  false,
		processing: false,
		mode:       "Fast",
		theme:      "dark",
		width:      width,
		styleTheme: DefaultStatusBarTheme(),
	}
}

// Update handles messages (Bubble Tea Update method).
func (s *StatusBar) Update(msg tea.Msg) (*StatusBar, tea.Cmd) {
	// Status bar is mostly passive, but can handle specific messages
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
	}
	return s, nil
}

// View renders the status bar.
func (s *StatusBar) View() string {
	// Build left section
	left := s.buildLeftSection()

	// Build right section
	right := s.buildRightSection()

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacing := s.width - leftWidth - rightWidth
	if spacing < 0 {
		spacing = 0
	}

	// Combine sections
	content := left + strings.Repeat(" ", spacing) + right

	// Apply style
	style := lipgloss.NewStyle().
		Background(s.styleTheme.Background).
		Foreground(s.styleTheme.Foreground).
		Width(s.width).
		Padding(0, 1)

	return style.Render(content)
}

// buildLeftSection builds the left section of the status bar.
func (s *StatusBar) buildLeftSection() string {
	if s.customLeft != "" {
		return s.customLeft
	}

	var parts []string

	// Mode indicator
	modeStyle := lipgloss.NewStyle().
		Foreground(s.styleTheme.Highlight).
		Bold(true)
	parts = append(parts, modeStyle.Render(s.mode+" Mode"))

	// Model name
	if s.model != "" {
		modelStyle := lipgloss.NewStyle().Foreground(s.styleTheme.Foreground)
		parts = append(parts, modelStyle.Render(s.model))
	}

	// Processing indicator
	if s.processing {
		processingStyle := lipgloss.NewStyle().
			Foreground(s.styleTheme.Warning).
			Bold(true)
		parts = append(parts, processingStyle.Render("⟳ Processing"))
	}

	// Connection status
	connStyle := lipgloss.NewStyle().Foreground(s.styleTheme.Dim)
	if s.connected {
		connStyle = lipgloss.NewStyle().Foreground(s.styleTheme.Success)
		parts = append(parts, connStyle.Render("● Connected"))
	} else {
		connStyle = lipgloss.NewStyle().Foreground(s.styleTheme.Error)
		parts = append(parts, connStyle.Render("○ Disconnected"))
	}

	return strings.Join(parts, " │ ")
}

// buildRightSection builds the right section of the status bar.
func (s *StatusBar) buildRightSection() string {
	if s.customRight != "" {
		return s.customRight
	}

	var parts []string

	// Cost
	costStyle := lipgloss.NewStyle().Foreground(s.styleTheme.Dim)
	if s.cost > 0 {
		parts = append(parts, costStyle.Render(fmt.Sprintf("$%.4f", s.cost)))
	} else {
		parts = append(parts, costStyle.Render("$0.00"))
	}

	// Tokens
	totalTokens := s.tokensIn + s.tokensOut
	tokenStyle := lipgloss.NewStyle().Foreground(s.styleTheme.Dim)
	if totalTokens > 0 {
		tokenStr := fmt.Sprintf("%d tokens (↑%d ↓%d)", totalTokens, s.tokensIn, s.tokensOut)
		parts = append(parts, tokenStyle.Render(tokenStr))
	} else {
		parts = append(parts, tokenStyle.Render("0 tokens"))
	}

	return strings.Join(parts, " │ ")
}

// SetModel sets the current model name.
func (s *StatusBar) SetModel(model string) {
	s.model = model
}

// SetTokens sets the token usage.
func (s *StatusBar) SetTokens(input, output int) {
	s.tokensIn = input
	s.tokensOut = output
}

// AddTokens adds to the token usage.
func (s *StatusBar) AddTokens(input, output int) {
	s.tokensIn += input
	s.tokensOut += output
}

// SetCost sets the total cost.
func (s *StatusBar) SetCost(cost float64) {
	s.cost = cost
}

// AddCost adds to the total cost.
func (s *StatusBar) AddCost(cost float64) {
	s.cost += cost
}

// SetConnected sets the connection status.
func (s *StatusBar) SetConnected(connected bool) {
	s.connected = connected
}

// SetProcessing sets the processing status.
func (s *StatusBar) SetProcessing(processing bool) {
	s.processing = processing
}

// SetMode sets the mode (Fast/Thorough).
func (s *StatusBar) SetMode(mode string) {
	s.mode = mode
}

// SetThemeName sets the theme name.
func (s *StatusBar) SetThemeName(theme string) {
	s.theme = theme
}

// SetWidth sets the width of the status bar.
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// SetCustomLeft sets custom content for the left section.
func (s *StatusBar) SetCustomLeft(content string) {
	s.customLeft = content
}

// SetCustomRight sets custom content for the right section.
func (s *StatusBar) SetCustomRight(content string) {
	s.customRight = content
}

// SetStyleTheme sets the color theme for the status bar.
func (s *StatusBar) SetStyleTheme(theme StatusBarTheme) {
	s.styleTheme = theme
}

// GetModel returns the current model name.
func (s *StatusBar) GetModel() string {
	return s.model
}

// GetTokens returns the current token usage.
func (s *StatusBar) GetTokens() (input, output int) {
	return s.tokensIn, s.tokensOut
}

// GetCost returns the current cost.
func (s *StatusBar) GetCost() float64 {
	return s.cost
}

// IsConnected returns the connection status.
func (s *StatusBar) IsConnected() bool {
	return s.connected
}

// IsProcessing returns the processing status.
func (s *StatusBar) IsProcessing() bool {
	return s.processing
}

// GetMode returns the current mode.
func (s *StatusBar) GetMode() string {
	return s.mode
}

// Reset resets all counters to zero.
func (s *StatusBar) Reset() {
	s.tokensIn = 0
	s.tokensOut = 0
	s.cost = 0.0
	s.processing = false
}
