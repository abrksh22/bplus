package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SplitPane provides horizontal or vertical split view with resizable panes.
type SplitPane struct {
	left      tea.Model
	right     tea.Model
	split     float64 // 0.0 to 1.0 (percentage for left/top pane)
	direction string  // "horizontal" or "vertical"
	focused   string  // "left" or "right"
	width     int
	height    int
	minSize   int // Minimum pane size
	theme     SplitPaneTheme
}

// SplitPaneTheme defines the color scheme for the split pane.
type SplitPaneTheme struct {
	Border        lipgloss.Color
	FocusedBorder lipgloss.Color
	Divider       lipgloss.Color
}

// DefaultSplitPaneTheme returns the default split pane theme.
func DefaultSplitPaneTheme() SplitPaneTheme {
	return SplitPaneTheme{
		Border:        lipgloss.Color("#3b4261"),
		FocusedBorder: lipgloss.Color("#7aa2f7"),
		Divider:       lipgloss.Color("#565f89"),
	}
}

// NewSplitPane creates a new split pane component.
func NewSplitPane(left, right tea.Model, direction string) SplitPane {
	return SplitPane{
		left:      left,
		right:     right,
		split:     0.5, // 50/50 split by default
		direction: direction,
		focused:   "left",
		width:     80,
		height:    24,
		minSize:   10,
		theme:     DefaultSplitPaneTheme(),
	}
}

// Update handles messages (Bubble Tea Update method).
func (s *SplitPane) Update(msg tea.Msg) (*SplitPane, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+w":
			// Toggle focus
			if s.focused == "left" {
				s.focused = "right"
			} else {
				s.focused = "left"
			}
			return s, nil
		case "ctrl+h":
			// Decrease left/top pane size
			s.split -= 0.05
			if s.split < 0.1 {
				s.split = 0.1
			}
			return s, nil
		case "ctrl+l":
			// Increase left/top pane size
			s.split += 0.05
			if s.split > 0.9 {
				s.split = 0.9
			}
			return s, nil
		case "ctrl+=":
			// Reset to 50/50
			s.split = 0.5
			return s, nil
		}
	}

	// Forward to focused pane
	if s.focused == "left" && s.left != nil {
		var cmd tea.Cmd
		s.left, cmd = s.left.Update(msg)
		cmds = append(cmds, cmd)
	} else if s.focused == "right" && s.right != nil {
		var cmd tea.Cmd
		s.right, cmd = s.right.Update(msg)
		cmds = append(cmds, cmd)
	}

	return s, tea.Batch(cmds...)
}

// View renders the split pane.
func (s *SplitPane) View() string {
	if s.left == nil || s.right == nil {
		return "Split pane not properly initialized"
	}

	if s.direction == "horizontal" {
		return s.renderHorizontal()
	}
	return s.renderVertical()
}

// renderHorizontal renders horizontal split (left/right).
func (s *SplitPane) renderHorizontal() string {
	// Calculate pane widths
	dividerWidth := 1
	leftWidth := int(float64(s.width-dividerWidth) * s.split)
	rightWidth := s.width - leftWidth - dividerWidth

	// Ensure minimum sizes
	if leftWidth < s.minSize {
		leftWidth = s.minSize
		rightWidth = s.width - leftWidth - dividerWidth
	}
	if rightWidth < s.minSize {
		rightWidth = s.minSize
		leftWidth = s.width - rightWidth - dividerWidth
	}

	// Get border colors
	leftBorder := s.theme.Border
	rightBorder := s.theme.Border
	if s.focused == "left" {
		leftBorder = s.theme.FocusedBorder
	} else {
		rightBorder = s.theme.FocusedBorder
	}

	// Style panes
	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(s.height).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(leftBorder)

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(s.height).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(rightBorder)

	dividerStyle := lipgloss.NewStyle().
		Width(dividerWidth).
		Height(s.height).
		Foreground(s.theme.Divider)

	// Render panes
	leftView := leftStyle.Render(s.left.View())
	rightView := rightStyle.Render(s.right.View())
	divider := dividerStyle.Render("│")

	// Combine horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, leftView, divider, rightView)
}

// renderVertical renders vertical split (top/bottom).
func (s *SplitPane) renderVertical() string {
	// Calculate pane heights
	dividerHeight := 1
	topHeight := int(float64(s.height-dividerHeight) * s.split)
	bottomHeight := s.height - topHeight - dividerHeight

	// Ensure minimum sizes
	if topHeight < s.minSize {
		topHeight = s.minSize
		bottomHeight = s.height - topHeight - dividerHeight
	}
	if bottomHeight < s.minSize {
		bottomHeight = s.minSize
		topHeight = s.height - bottomHeight - dividerHeight
	}

	// Get border colors
	topBorder := s.theme.Border
	bottomBorder := s.theme.Border
	if s.focused == "left" { // "left" means top in vertical mode
		topBorder = s.theme.FocusedBorder
	} else {
		bottomBorder = s.theme.FocusedBorder
	}

	// Style panes
	topStyle := lipgloss.NewStyle().
		Width(s.width).
		Height(topHeight).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(topBorder)

	bottomStyle := lipgloss.NewStyle().
		Width(s.width).
		Height(bottomHeight).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(bottomBorder)

	dividerStyle := lipgloss.NewStyle().
		Width(s.width).
		Foreground(s.theme.Divider)

	// Render panes
	topView := topStyle.Render(s.left.View())
	bottomView := bottomStyle.Render(s.right.View())
	divider := dividerStyle.Render(lipgloss.PlaceHorizontal(s.width, lipgloss.Center, "─"))

	// Combine vertically
	return lipgloss.JoinVertical(lipgloss.Left, topView, divider, bottomView)
}

// SetSplit sets the split ratio (0.0 to 1.0).
func (s *SplitPane) SetSplit(ratio float64) {
	if ratio < 0.1 {
		ratio = 0.1
	}
	if ratio > 0.9 {
		ratio = 0.9
	}
	s.split = ratio
}

// FocusLeft focuses the left/top pane.
func (s *SplitPane) FocusLeft() {
	s.focused = "left"
}

// FocusRight focuses the right/bottom pane.
func (s *SplitPane) FocusRight() {
	s.focused = "right"
}

// SetSize sets the dimensions of the split pane.
func (s *SplitPane) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetDirection sets the split direction.
func (s *SplitPane) SetDirection(direction string) {
	if direction == "horizontal" || direction == "vertical" {
		s.direction = direction
	}
}

// SetLeft sets the left/top pane model.
func (s *SplitPane) SetLeft(model tea.Model) {
	s.left = model
}

// SetRight sets the right/bottom pane model.
func (s *SplitPane) SetRight(model tea.Model) {
	s.right = model
}

// SetMinSize sets the minimum pane size.
func (s *SplitPane) SetMinSize(size int) {
	s.minSize = size
}

// SetTheme sets the color theme for the split pane.
func (s *SplitPane) SetTheme(theme SplitPaneTheme) {
	s.theme = theme
}

// GetFocused returns which pane is focused.
func (s *SplitPane) GetFocused() string {
	return s.focused
}

// GetSplit returns the current split ratio.
func (s *SplitPane) GetSplit() float64 {
	return s.split
}
