package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines all key bindings for the application.
type KeyMap struct {
	// Global keys (work in any view)
	Quit        key.Binding
	ForceQuit   key.Binding
	Help        key.Binding
	ClearScreen key.Binding
	Settings    key.Binding

	// Navigation keys
	FocusNext     key.Binding
	FocusPrevious key.Binding
	FocusInput    key.Binding
	FocusOutput   key.Binding
	FocusFiles    key.Binding
	FocusSession  key.Binding

	// Chat keys
	Send        key.Binding
	NewLine     key.Binding
	Cancel      key.Binding
	HistoryUp   key.Binding
	HistoryDown key.Binding

	// Editing keys
	Undo key.Binding
	Redo key.Binding

	// View keys
	TogglePlan     key.Binding
	ToggleMode     key.Binding
	CommandPalette key.Binding
	ToggleSidebar  key.Binding
	ToggleBrowser  key.Binding

	// Scroll keys
	ScrollUp   key.Binding
	ScrollDown key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	Home       key.Binding
	End        key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Global keys
		Quit: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		ClearScreen: key.NewBinding(
			key.WithKeys("ctrl+k"),
			key.WithHelp("ctrl+k", "clear screen"),
		),
		Settings: key.NewBinding(
			key.WithKeys("ctrl+/"),
			key.WithHelp("ctrl+/", "settings"),
		),

		// Navigation keys
		FocusNext: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "focus next"),
		),
		FocusPrevious: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "focus previous"),
		),
		FocusInput: key.NewBinding(
			key.WithKeys("ctrl+g"),
			key.WithHelp("ctrl+g", "focus input"),
		),
		FocusOutput: key.NewBinding(
			key.WithKeys("ctrl+o"),
			key.WithHelp("ctrl+o", "focus output"),
		),
		FocusFiles: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "focus files"),
		),
		FocusSession: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "focus sessions"),
		),

		// Chat keys
		Send: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "send message"),
		),
		NewLine: key.NewBinding(
			key.WithKeys("shift+enter"),
			key.WithHelp("shift+enter", "new line"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		HistoryUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "previous message"),
		),
		HistoryDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "next message"),
		),

		// Editing keys
		Undo: key.NewBinding(
			key.WithKeys("ctrl+z"),
			key.WithHelp("ctrl+z", "undo"),
		),
		Redo: key.NewBinding(
			key.WithKeys("ctrl+y"),
			key.WithHelp("ctrl+y", "redo"),
		),

		// View keys
		TogglePlan: key.NewBinding(
			key.WithKeys("shift+tab", "shift+tab"), // Double tap
			key.WithHelp("shift+tab×2", "toggle plan mode"),
		),
		ToggleMode: key.NewBinding(
			key.WithKeys("ctrl+m"),
			key.WithHelp("ctrl+m", "toggle fast/thorough"),
		),
		CommandPalette: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "command palette"),
		),
		ToggleSidebar: key.NewBinding(
			key.WithKeys("ctrl+\\"),
			key.WithHelp("ctrl+\\", "toggle sidebar"),
		),
		ToggleBrowser: key.NewBinding(
			key.WithKeys("ctrl+b"),
			key.WithHelp("ctrl+b", "toggle file browser"),
		),

		// Scroll keys
		ScrollUp: key.NewBinding(
			key.WithKeys("ctrl+up"),
			key.WithHelp("ctrl+↑", "scroll up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("ctrl+down"),
			key.WithHelp("ctrl+↓", "scroll down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdown", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "go to bottom"),
		),
	}
}

// ShortHelp returns a quick help text.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Help,
		k.Quit,
		k.Send,
		k.Cancel,
	}
}

// FullHelp returns the full help text.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Global
		{k.Quit, k.ForceQuit, k.Help, k.ClearScreen},
		// Navigation
		{k.FocusInput, k.FocusOutput, k.FocusFiles, k.FocusSession},
		// Chat
		{k.Send, k.NewLine, k.HistoryUp, k.HistoryDown},
		// Editing
		{k.Undo, k.Redo},
		// View
		{k.ToggleMode, k.CommandPalette, k.ToggleSidebar, k.ToggleBrowser},
	}
}

// Note: key.Binding matching is done directly in update.go using key.Matches()
