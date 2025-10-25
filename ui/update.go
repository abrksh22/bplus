package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all state changes in the application (Bubble Tea lifecycle method).
// This is the heart of the Elm-inspired architecture - all messages flow through here.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle standard Bubble Tea messages first
	switch msg := msg.(type) {

	// Window resizing
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	// Keyboard input
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	// Mouse events (if needed in future)
	case tea.MouseMsg:
		return m.handleMouse(msg)

	// Custom messages
	case WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case ErrorMsg:
		m.SetError(msg.Err)
		return m, nil

	case ClearErrorMsg:
		m.SetError(nil)
		return m, nil

	case ReadyMsg:
		m.SetReady(true)
		return m, nil

	case QuitMsg:
		m.Quit()
		return m, tea.Quit

	case ChangeViewMsg:
		m.SetView(msg.View)
		return m, nil

	case ChangeFocusMsg:
		m.SetFocus(msg.Component)
		return m, nil

	case ChangeThemeMsg:
		m.SetTheme(msg.Theme)
		return m, nil

	case UserInputMsg:
		return m.handleUserInput(msg)

	case StreamTokenMsg:
		return m.handleStreamToken(msg)

	case LoadingMsg:
		return m.handleLoading(msg)

	case ProgressMsg:
		return m.handleProgress(msg)

	case ShowModalMsg:
		return m.handleShowModal(msg)

	case ModalResultMsg:
		return m.handleModalResult(msg)

	case StatusUpdateMsg:
		return m.handleStatusUpdate(msg)

	case ShowHelpMsg:
		return m.handleShowHelp(msg)

	case ComponentMsg:
		return m.handleComponentMsg(msg)

	default:
		// Unknown message type - just return current model
		return m, nil
	}
}

// handleWindowSize handles window resize events.
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	// Mark as ready after first window size message
	if !m.ready {
		m.ready = true
	}

	// TODO: Update component sizes when components are implemented
	// m.input.SetWidth(m.width)
	// m.output.SetSize(m.width, m.height - statusBarHeight)

	return m, nil
}

// handleKeyPress handles keyboard input.
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global key bindings (work in any view)
	switch {
	case msg.String() == "ctrl+d":
		m.quitting = true
		return m, tea.Quit

	case msg.String() == "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case msg.String() == "?":
		// Toggle help overlay
		if m.view == ViewHelp {
			// Return to previous view
			m.view = ViewChat
		} else {
			m.view = ViewHelp
		}
		return m, nil

	case msg.String() == "ctrl+k":
		// Clear the screen
		// TODO: Implement when output component exists
		return m, tea.ClearScreen

	case msg.String() == "ctrl+/":
		// Toggle settings view
		if m.view == ViewSettings {
			m.view = ViewChat
		} else {
			m.view = ViewSettings
		}
		return m, nil
	}

	// View-specific key handling
	switch m.view {
	case ViewStartup:
		return m.handleStartupKeys(msg)
	case ViewChat:
		return m.handleChatKeys(msg)
	case ViewSettings:
		return m.handleSettingsKeys(msg)
	case ViewHelp:
		return m.handleHelpKeys(msg)
	}

	return m, nil
}

// handleStartupKeys handles keys in startup view.
func (m *Model) handleStartupKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Any key press in startup view transitions to chat view
	if msg.Type == tea.KeyEnter || msg.Type == tea.KeySpace {
		m.view = ViewChat
		return m, nil
	}
	return m, nil
}

// handleChatKeys handles keys in chat view.
func (m *Model) handleChatKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Component-specific handling based on focus
	switch m.focusedComponent {
	case "input":
		// TODO: Forward to input component when implemented
		// return m.input.Update(msg)
		return m, nil
	case "output":
		// TODO: Forward to output component when implemented
		// return m.output.Update(msg)
		return m, nil
	default:
		return m, nil
	}
}

// handleSettingsKeys handles keys in settings view.
func (m *Model) handleSettingsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement settings navigation
	if msg.Type == tea.KeyEsc {
		m.view = ViewChat
		return m, nil
	}
	return m, nil
}

// handleHelpKeys handles keys in help view.
func (m *Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Any key exits help
	if msg.Type == tea.KeyEsc || msg.String() == "?" {
		m.view = ViewChat
		return m, nil
	}
	return m, nil
}

// handleMouse handles mouse events.
func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement mouse handling for clicking components, etc.
	return m, nil
}

// handleUserInput handles user text input submission.
func (m *Model) handleUserInput(msg UserInputMsg) (tea.Model, tea.Cmd) {
	// TODO: Process user input
	// - Add to conversation history
	// - Send to agent for processing
	// - Update UI to show loading state
	return m, nil
}

// handleStreamToken handles streaming tokens from LLM.
func (m *Model) handleStreamToken(msg StreamTokenMsg) (tea.Model, tea.Cmd) {
	// TODO: Update output component with streaming token
	// if msg.Done {
	//     // Streaming complete
	// } else {
	//     // Append token to current message
	// }
	return m, nil
}

// handleLoading handles loading state changes.
func (m *Model) handleLoading(msg LoadingMsg) (tea.Model, tea.Cmd) {
	// TODO: Show/hide spinner component
	return m, nil
}

// handleProgress handles progress updates.
func (m *Model) handleProgress(msg ProgressMsg) (tea.Model, tea.Cmd) {
	// TODO: Update progress bar component
	return m, nil
}

// handleShowModal handles modal display requests.
func (m *Model) handleShowModal(msg ShowModalMsg) (tea.Model, tea.Cmd) {
	// TODO: Show modal component
	return m, nil
}

// handleModalResult handles modal button clicks.
func (m *Model) handleModalResult(msg ModalResultMsg) (tea.Model, tea.Cmd) {
	// TODO: Process modal result
	return m, nil
}

// handleStatusUpdate handles status bar updates.
func (m *Model) handleStatusUpdate(msg StatusUpdateMsg) (tea.Model, tea.Cmd) {
	// TODO: Update status bar component
	return m, nil
}

// handleShowHelp handles help overlay toggle.
func (m *Model) handleShowHelp(msg ShowHelpMsg) (tea.Model, tea.Cmd) {
	if msg.Show {
		m.view = ViewHelp
	} else {
		m.view = ViewChat
	}
	return m, nil
}

// handleComponentMsg handles component-specific messages.
func (m *Model) handleComponentMsg(msg ComponentMsg) (tea.Model, tea.Cmd) {
	// TODO: Route to appropriate component
	switch msg.Component {
	case "input":
		// Forward to input component
	case "output":
		// Forward to output component
	case "statusbar":
		// Forward to status bar
		// ... etc
	}
	return m, nil
}
