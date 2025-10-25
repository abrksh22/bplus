package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Modal displays an overlay modal dialog with confirm/cancel buttons.
type Modal struct {
	title     string
	content   string
	buttons   []string
	selected  int
	visible   bool
	width     int
	height    int
	onConfirm func()
	onCancel  func()
	theme     ModalTheme
}

// ModalTheme defines the color scheme for the modal.
type ModalTheme struct {
	Overlay   lipgloss.Color
	Border    lipgloss.Color
	Title     lipgloss.Color
	Content   lipgloss.Color
	Button    lipgloss.Style
	ButtonSel lipgloss.Style
}

// DefaultModalTheme returns the default modal theme.
func DefaultModalTheme() ModalTheme {
	return ModalTheme{
		Overlay: lipgloss.Color("#000000"),
		Border:  lipgloss.Color("#7aa2f7"),
		Title:   lipgloss.Color("#7aa2f7"),
		Content: lipgloss.Color("#c0caf5"),
		Button: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c0caf5")).
			Background(lipgloss.Color("#3b4261")).
			Padding(0, 2),
		ButtonSel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1a1b26")).
			Background(lipgloss.Color("#7aa2f7")).
			Bold(true).
			Padding(0, 2),
	}
}

// NewModal creates a new modal dialog.
func NewModal(title, content string) Modal {
	return Modal{
		title:    title,
		content:  content,
		buttons:  []string{"OK", "Cancel"},
		selected: 0,
		visible:  false,
		width:    60,
		height:   20,
		theme:    DefaultModalTheme(),
	}
}

// Update handles messages (Bubble Tea Update method).
func (m *Modal) Update(msg tea.Msg) (*Modal, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.selected > 0 {
				m.selected--
			}
		case "right", "l":
			if m.selected < len(m.buttons)-1 {
				m.selected++
			}
		case "tab":
			m.selected = (m.selected + 1) % len(m.buttons)
		case "shift+tab":
			m.selected = (m.selected - 1 + len(m.buttons)) % len(m.buttons)
		case "enter":
			if m.selected == 0 && m.onConfirm != nil {
				m.onConfirm()
			} else if m.onCancel != nil {
				m.onCancel()
			}
			m.visible = false
		case "esc":
			if m.onCancel != nil {
				m.onCancel()
			}
			m.visible = false
		}
	}

	return m, nil
}

// View renders the modal.
func (m *Modal) View() string {
	if !m.visible {
		return ""
	}

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.Title).
		Bold(true).
		Align(lipgloss.Center)
	titleView := titleStyle.Render(m.title)

	// Content
	contentStyle := lipgloss.NewStyle().
		Foreground(m.theme.Content).
		Width(m.width - 4).
		Align(lipgloss.Left)
	contentView := contentStyle.Render(m.content)

	// Buttons
	var buttons []string
	for i, btn := range m.buttons {
		var btnView string
		if i == m.selected {
			btnView = m.theme.ButtonSel.Render(btn)
		} else {
			btnView = m.theme.Button.Render(btn)
		}
		buttons = append(buttons, btnView)
	}
	buttonsView := lipgloss.JoinHorizontal(lipgloss.Center, buttons...)
	buttonsView = lipgloss.NewStyle().Width(m.width - 4).Align(lipgloss.Center).Render(buttonsView)

	// Combine all parts
	dialogContent := lipgloss.JoinVertical(
		lipgloss.Center,
		titleView,
		"",
		contentView,
		"",
		buttonsView,
	)

	// Add border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Width(m.width).
		Padding(1).
		Align(lipgloss.Center)

	box := boxStyle.Render(dialogContent)

	return box
}

// Show shows the modal.
func (m *Modal) Show() {
	m.visible = true
}

// Hide hides the modal.
func (m *Modal) Hide() {
	m.visible = false
}

// IsVisible returns whether the modal is visible.
func (m *Modal) IsVisible() bool {
	return m.visible
}

// SetTitle sets the modal title.
func (m *Modal) SetTitle(title string) {
	m.title = title
}

// SetContent sets the modal content.
func (m *Modal) SetContent(content string) {
	m.content = content
}

// SetButtons sets the button labels.
func (m *Modal) SetButtons(buttons []string) {
	m.buttons = buttons
	if m.selected >= len(buttons) {
		m.selected = len(buttons) - 1
	}
}

// SetSize sets the width and height of the modal.
func (m *Modal) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// OnConfirm sets the callback function for the confirm button.
func (m *Modal) OnConfirm(fn func()) {
	m.onConfirm = fn
}

// OnCancel sets the callback function for the cancel button.
func (m *Modal) OnCancel(fn func()) {
	m.onCancel = fn
}

// SetTheme sets the color theme for the modal.
func (m *Modal) SetTheme(theme ModalTheme) {
	m.theme = theme
}

// Center centers the modal at the given screen dimensions.
func (m *Modal) Center(screenWidth, screenHeight int) string {
	modalView := m.View()
	if !m.visible {
		return ""
	}

	// Create a semi-transparent overlay (using spaces with background)
	overlayStyle := lipgloss.NewStyle().
		Background(m.theme.Overlay).
		Width(screenWidth).
		Height(screenHeight)

	// Calculate centered position (for future use)

	// Place modal in center
	centered := lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalView,
	)

	// Overlay with modal
	return overlayStyle.Render(centered)
}
