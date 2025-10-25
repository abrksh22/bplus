// Package ui provides the terminal user interface for b+ using Bubble Tea.
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the main application state and implements tea.Model.
type Model struct {
	// Window dimensions
	width  int
	height int

	// Current view state
	ready            bool
	quitting         bool
	err              error
	view             ViewMode
	focusedComponent string

	// Application reference (Phase 6)
	app interface{} // Will be *app.Application, using interface{} to avoid circular import

	// UI Components (will be populated in Phase 3.2)
	// input      *InputComponent
	// output     *OutputComponent
	// statusBar  *StatusBarComponent
	// spinner    *SpinnerComponent
	// modal      *ModalComponent

	// Theme and styling
	theme *Theme

	// Key bindings
	keys KeyMap
}

// ViewMode represents the current view mode.
type ViewMode int

const (
	ViewStartup ViewMode = iota
	ViewChat
	ViewSettings
	ViewHelp
)

// New creates a new UI model with default settings.
func New() *Model {
	return &Model{
		ready:            false,
		quitting:         false,
		view:             ViewStartup,
		focusedComponent: "input",
		theme:            DefaultTheme(),
		keys:             DefaultKeyMap(),
	}
}

// NewWithApp creates a new UI model with application reference.
func NewWithApp(application interface{}) *Model {
	m := New()
	m.app = application
	return m
}

// Init initializes the model (Bubble Tea lifecycle method).
func (m *Model) Init() tea.Cmd {
	// Return commands to run on initialization
	// For now, we'll just return nil - more commands will be added as we build components
	return nil
}

// Width returns the current window width.
func (m *Model) Width() int {
	return m.width
}

// Height returns the current window height.
func (m *Model) Height() int {
	return m.height
}

// SetSize updates the window dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// IsReady returns whether the UI is ready to display.
func (m *Model) IsReady() bool {
	return m.ready
}

// SetReady marks the UI as ready to display.
func (m *Model) SetReady(ready bool) {
	m.ready = ready
}

// IsQuitting returns whether the application is quitting.
func (m *Model) IsQuitting() bool {
	return m.quitting
}

// Quit marks the application to quit.
func (m *Model) Quit() {
	m.quitting = true
}

// Error returns the current error, if any.
func (m *Model) Error() error {
	return m.err
}

// SetError sets an error to display.
func (m *Model) SetError(err error) {
	m.err = err
}

// View returns the current view mode.
func (m *Model) CurrentView() ViewMode {
	return m.view
}

// SetView changes the current view mode.
func (m *Model) SetView(view ViewMode) {
	m.view = view
}

// Theme returns the current theme.
func (m *Model) Theme() *Theme {
	return m.theme
}

// SetTheme changes the current theme.
func (m *Model) SetTheme(theme *Theme) {
	m.theme = theme
}

// FocusedComponent returns the name of the currently focused component.
func (m *Model) FocusedComponent() string {
	return m.focusedComponent
}

// SetFocus changes the focused component.
func (m *Model) SetFocus(component string) {
	m.focusedComponent = component
}

// String representation for debugging.
func (m *Model) String() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		"b+ UI Model State:",
		"  Ready: "+boolToString(m.ready),
		"  Quitting: "+boolToString(m.quitting),
		"  View: "+viewModeToString(m.view),
		"  Size: "+intToString(m.width)+"x"+intToString(m.height),
		"  Focus: "+m.focusedComponent,
	)
}

// Helper functions
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func viewModeToString(v ViewMode) string {
	switch v {
	case ViewStartup:
		return "Startup"
	case ViewChat:
		return "Chat"
	case ViewSettings:
		return "Settings"
	case ViewHelp:
		return "Help"
	default:
		return "Unknown"
	}
}

func intToString(i int) string {
	// Simple int to string - will use strconv in actual implementation
	return "..."
}
