package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListItem represents a single item in the list.
type ListItem struct {
	Title       string
	Description string
	Icon        string
	Value       interface{}
}

// List displays a selectable list with keyboard navigation.
type List struct {
	items       []ListItem
	selected    int
	filter      string
	height      int
	width       int
	offset      int // For scrolling
	multiSelect bool
	selectedMap map[int]bool
	theme       ListTheme
}

// ListTheme defines the color scheme for the list.
type ListTheme struct {
	Selected    lipgloss.Style
	Unselected  lipgloss.Style
	Title       lipgloss.Color
	Description lipgloss.Color
	Icon        lipgloss.Color
	Border      lipgloss.Color
}

// DefaultListTheme returns the default list theme.
func DefaultListTheme() ListTheme {
	return ListTheme{
		Selected: lipgloss.NewStyle().
			Background(lipgloss.Color("#7aa2f7")).
			Foreground(lipgloss.Color("#1a1b26")).
			Bold(true).
			Padding(0, 1),
		Unselected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c0caf5")).
			Padding(0, 1),
		Title:       lipgloss.Color("#c0caf5"),
		Description: lipgloss.Color("#565f89"),
		Icon:        lipgloss.Color("#7aa2f7"),
		Border:      lipgloss.Color("#3b4261"),
	}
}

// NewList creates a new list component.
func NewList(items []ListItem) List {
	return List{
		items:       items,
		selected:    0,
		filter:      "",
		height:      10,
		width:       60,
		offset:      0,
		multiSelect: false,
		selectedMap: make(map[int]bool),
		theme:       DefaultListTheme(),
	}
}

// Update handles messages (Bubble Tea Update method).
func (l *List) Update(msg tea.Msg) (*List, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if l.selected > 0 {
				l.selected--
				l.ensureVisible()
			}
		case "down", "j":
			filteredItems := l.getFilteredItems()
			if l.selected < len(filteredItems)-1 {
				l.selected++
				l.ensureVisible()
			}
		case "home", "g":
			l.selected = 0
			l.offset = 0
		case "end", "G":
			filteredItems := l.getFilteredItems()
			l.selected = len(filteredItems) - 1
			l.ensureVisible()
		case "pgup":
			l.selected -= l.height
			if l.selected < 0 {
				l.selected = 0
			}
			l.ensureVisible()
		case "pgdown":
			filteredItems := l.getFilteredItems()
			l.selected += l.height
			if l.selected >= len(filteredItems) {
				l.selected = len(filteredItems) - 1
			}
			l.ensureVisible()
		case " ":
			// Toggle selection in multi-select mode
			if l.multiSelect {
				l.selectedMap[l.selected] = !l.selectedMap[l.selected]
			}
		}
	}

	return l, nil
}

// View renders the list.
func (l *List) View() string {
	filteredItems := l.getFilteredItems()
	if len(filteredItems) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565f89")).
			Faint(true).
			Align(lipgloss.Center).
			Width(l.width)
		return emptyStyle.Render("No items to display")
	}

	// Calculate visible range
	visibleStart := l.offset
	visibleEnd := l.offset + l.height
	if visibleEnd > len(filteredItems) {
		visibleEnd = len(filteredItems)
	}

	// Render visible items
	var renderedItems []string
	for i := visibleStart; i < visibleEnd; i++ {
		item := filteredItems[i]
		renderedItems = append(renderedItems, l.renderItem(item, i == l.selected, i))
	}

	// Join items
	content := strings.Join(renderedItems, "\n")

	// Add border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(l.theme.Border).
		Width(l.width).
		Height(l.height + 2)

	return boxStyle.Render(content)
}

// renderItem renders a single list item.
func (l *List) renderItem(item ListItem, isSelected bool, index int) string {
	// Build item content
	var parts []string

	// Selection indicator
	if l.multiSelect {
		checkbox := "☐"
		if l.selectedMap[index] {
			checkbox = "☑"
		}
		checkboxStyle := lipgloss.NewStyle().Foreground(l.theme.Icon)
		parts = append(parts, checkboxStyle.Render(checkbox))
	}

	// Icon
	if item.Icon != "" {
		iconStyle := lipgloss.NewStyle().Foreground(l.theme.Icon)
		parts = append(parts, iconStyle.Render(item.Icon))
	}

	// Title
	titleStyle := lipgloss.NewStyle().Foreground(l.theme.Title)
	parts = append(parts, titleStyle.Render(item.Title))

	// Description
	if item.Description != "" {
		descStyle := lipgloss.NewStyle().Foreground(l.theme.Description).Faint(true)
		parts = append(parts, descStyle.Render("- "+item.Description))
	}

	itemContent := strings.Join(parts, " ")

	// Apply selection style
	if isSelected {
		return l.theme.Selected.Width(l.width - 4).Render(itemContent)
	}
	return l.theme.Unselected.Width(l.width - 4).Render(itemContent)
}

// ensureVisible ensures the selected item is visible in the viewport.
func (l *List) ensureVisible() {
	if l.selected < l.offset {
		l.offset = l.selected
	} else if l.selected >= l.offset+l.height {
		l.offset = l.selected - l.height + 1
	}
}

// getFilteredItems returns items filtered by the current filter.
func (l *List) getFilteredItems() []ListItem {
	if l.filter == "" {
		return l.items
	}

	var filtered []ListItem
	filterLower := strings.ToLower(l.filter)
	for _, item := range l.items {
		titleLower := strings.ToLower(item.Title)
		descLower := strings.ToLower(item.Description)
		if strings.Contains(titleLower, filterLower) || strings.Contains(descLower, filterLower) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// SelectedItem returns the currently selected item.
func (l *List) SelectedItem() *ListItem {
	filteredItems := l.getFilteredItems()
	if l.selected >= 0 && l.selected < len(filteredItems) {
		return &filteredItems[l.selected]
	}
	return nil
}

// SelectedItems returns all selected items (for multi-select).
func (l *List) SelectedItems() []ListItem {
	if !l.multiSelect {
		item := l.SelectedItem()
		if item != nil {
			return []ListItem{*item}
		}
		return []ListItem{}
	}

	var selected []ListItem
	filteredItems := l.getFilteredItems()
	for i := range filteredItems {
		if l.selectedMap[i] {
			selected = append(selected, filteredItems[i])
		}
	}
	return selected
}

// SetFilter sets the filter string.
func (l *List) SetFilter(filter string) {
	l.filter = filter
	l.selected = 0
	l.offset = 0
}

// SetItems sets the list items.
func (l *List) SetItems(items []ListItem) {
	l.items = items
	l.selected = 0
	l.offset = 0
}

// SetHeight sets the visible height of the list.
func (l *List) SetHeight(height int) {
	l.height = height
}

// SetWidth sets the width of the list.
func (l *List) SetWidth(width int) {
	l.width = width
}

// SetMultiSelect enables or disables multi-select mode.
func (l *List) SetMultiSelect(enabled bool) {
	l.multiSelect = enabled
	if !enabled {
		l.selectedMap = make(map[int]bool)
	}
}

// SetTheme sets the color theme for the list.
func (l *List) SetTheme(theme ListTheme) {
	l.theme = theme
}

// GetItemCount returns the total number of items (filtered).
func (l *List) GetItemCount() int {
	return len(l.getFilteredItems())
}

// GetSelected returns the selected index.
func (l *List) GetSelected() int {
	return l.selected
}
