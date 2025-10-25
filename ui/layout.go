package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Layout manages the arrangement of UI components on screen.
type Layout struct {
	name    string
	regions map[string]*Region
	width   int
	height  int
}

// Region represents a rectangular area containing a component.
type Region struct {
	x         int
	y         int
	width     int
	height    int
	component tea.Model
	border    bool
	padding   int
}

// LayoutPreset represents predefined layout configurations.
type LayoutPreset string

const (
	// LayoutDefault is the standard chat layout (70% output, 25% input, 5% status)
	LayoutDefault LayoutPreset = "default"
	// LayoutCompact minimizes chrome and maximizes output area
	LayoutCompact LayoutPreset = "compact"
	// LayoutSplitScreen provides side-by-side conversation and code viewer
	LayoutSplitScreen LayoutPreset = "split-screen"
	// LayoutFocus is output-only fullscreen with hidden input
	LayoutFocus LayoutPreset = "focus"
)

// NewLayout creates a new layout with the given name and dimensions.
func NewLayout(name string, width, height int) *Layout {
	return &Layout{
		name:    name,
		regions: make(map[string]*Region),
		width:   width,
		height:  height,
	}
}

// NewRegion creates a new region.
func NewRegion(x, y, width, height int, component tea.Model) *Region {
	return &Region{
		x:         x,
		y:         y,
		width:     width,
		height:    height,
		component: component,
		border:    false,
		padding:   0,
	}
}

// Update handles messages for the layout.
func (l *Layout) Update(msg tea.Msg) (*Layout, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle window resize
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.SetSize(msg.Width, msg.Height)
		// Recalculate region sizes based on new dimensions
		// This would be done by the preset that created the layout
	}

	// Update all regions' components
	for _, region := range l.regions {
		if region.component != nil {
			var cmd tea.Cmd
			region.component, cmd = region.component.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return l, tea.Batch(cmds...)
}

// View renders the layout.
func (l *Layout) View() string {
	// Create a grid for the layout
	grid := make([][]string, l.height)
	for i := range grid {
		grid[i] = make([]string, l.width)
		for j := range grid[i] {
			grid[i][j] = " " // Fill with spaces
		}
	}

	// Render each region into the grid
	for _, region := range l.regions {
		if region.component == nil {
			continue
		}

		// Get component view
		view := region.component.View()

		// Apply border if needed
		if region.border {
			borderStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Width(region.width - 2).
				Height(region.height - 2)
			view = borderStyle.Render(view)
		}

		// Apply padding if needed
		if region.padding > 0 {
			paddingStyle := lipgloss.NewStyle().
				Padding(region.padding)
			view = paddingStyle.Render(view)
		}

		// Place view at region position
		// This is a simplified placement - a real implementation
		// would parse the view and place it character by character
		_ = view // Just to use the variable
	}

	// For now, return a simplified layout
	// A full implementation would construct the view from the grid
	return l.renderSimple()
}

// renderSimple provides a simplified rendering (placeholder).
func (l *Layout) renderSimple() string {
	var views []string

	// Render regions in a simple vertical stack
	// A real implementation would use the x,y coordinates
	for name, region := range l.regions {
		if region.component != nil {
			view := region.component.View()
			if region.border {
				borderStyle := lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#3b4261")).
					Width(region.width - 2).
					Height(region.height - 2)
				view = borderStyle.Render(view)
			}
			views = append(views, lipgloss.JoinHorizontal(lipgloss.Left, name+": ", view))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, views...)
}

// SetSize updates the layout dimensions.
func (l *Layout) SetSize(width, height int) {
	l.width = width
	l.height = height
}

// AddRegion adds a region to the layout.
func (l *Layout) AddRegion(name string, region *Region) {
	l.regions[name] = region
}

// GetRegion retrieves a region by name.
func (l *Layout) GetRegion(name string) *Region {
	return l.regions[name]
}

// RemoveRegion removes a region from the layout.
func (l *Layout) RemoveRegion(name string) {
	delete(l.regions, name)
}

// SetBorder enables/disables border for a region.
func (r *Region) SetBorder(border bool) {
	r.border = border
}

// SetPadding sets the padding for a region.
func (r *Region) SetPadding(padding int) {
	r.padding = padding
}

// SetComponent sets the component for a region.
func (r *Region) SetComponent(component tea.Model) {
	r.component = component
}

// SetPosition sets the x,y position of the region.
func (r *Region) SetPosition(x, y int) {
	r.x = x
	r.y = y
}

// SetSize sets the width and height of the region.
func (r *Region) SetSize(width, height int) {
	r.width = width
	r.height = height
}

// GetComponent returns the region's component.
func (r *Region) GetComponent() tea.Model {
	return r.component
}

// CreateDefaultLayout creates the default chat layout.
func CreateDefaultLayout(width, height int) *Layout {
	layout := NewLayout("default", width, height)

	statusHeight := 1
	inputHeight := max(3, int(float64(height)*0.15)) // 15% of height, minimum 3
	outputHeight := height - statusHeight - inputHeight

	// Status bar at top
	statusRegion := NewRegion(0, 0, width, statusHeight, nil)
	layout.AddRegion("status", statusRegion)

	// Output in middle
	outputRegion := NewRegion(0, statusHeight, width, outputHeight, nil)
	outputRegion.SetBorder(true)
	layout.AddRegion("output", outputRegion)

	// Input at bottom
	inputRegion := NewRegion(0, statusHeight+outputHeight, width, inputHeight, nil)
	inputRegion.SetBorder(true)
	layout.AddRegion("input", inputRegion)

	return layout
}

// CreateCompactLayout creates a compact layout with minimal chrome.
func CreateCompactLayout(width, height int) *Layout {
	layout := NewLayout("compact", width, height)

	statusHeight := 1
	inputHeight := 2
	outputHeight := height - statusHeight - inputHeight

	// Minimal status
	statusRegion := NewRegion(0, 0, width, statusHeight, nil)
	layout.AddRegion("status", statusRegion)

	// Maximized output
	outputRegion := NewRegion(0, statusHeight, width, outputHeight, nil)
	layout.AddRegion("output", outputRegion)

	// Compact input
	inputRegion := NewRegion(0, statusHeight+outputHeight, width, inputHeight, nil)
	layout.AddRegion("input", inputRegion)

	return layout
}

// CreateSplitScreenLayout creates a split-screen layout.
func CreateSplitScreenLayout(width, height int) *Layout {
	layout := NewLayout("split-screen", width, height)

	statusHeight := 1
	inputHeight := 3
	contentHeight := height - statusHeight - inputHeight
	leftWidth := width / 2
	rightWidth := width - leftWidth

	// Status at top
	statusRegion := NewRegion(0, 0, width, statusHeight, nil)
	layout.AddRegion("status", statusRegion)

	// Left: Conversation
	leftRegion := NewRegion(0, statusHeight, leftWidth, contentHeight, nil)
	leftRegion.SetBorder(true)
	layout.AddRegion("conversation", leftRegion)

	// Right: Code viewer
	rightRegion := NewRegion(leftWidth, statusHeight, rightWidth, contentHeight, nil)
	rightRegion.SetBorder(true)
	layout.AddRegion("viewer", rightRegion)

	// Input at bottom
	inputRegion := NewRegion(0, statusHeight+contentHeight, width, inputHeight, nil)
	inputRegion.SetBorder(true)
	layout.AddRegion("input", inputRegion)

	return layout
}

// CreateFocusLayout creates a focus layout (fullscreen output).
func CreateFocusLayout(width, height int) *Layout {
	layout := NewLayout("focus", width, height)

	// Output takes full screen
	outputRegion := NewRegion(0, 0, width, height, nil)
	layout.AddRegion("output", outputRegion)

	return layout
}

// Helper function for max (since Go doesn't have built-in max for ints)
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
