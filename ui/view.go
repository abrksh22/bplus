package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current state to a string (Bubble Tea lifecycle method).
// This function is called on every update to re-render the UI.
func (m *Model) View() string {
	if m.quitting {
		return m.renderQuit()
	}

	if !m.ready {
		return m.renderLoading()
	}

	// Render based on current view mode
	switch m.view {
	case ViewStartup:
		return m.renderStartup()
	case ViewChat:
		return m.renderChat()
	case ViewSettings:
		return m.renderSettings()
	case ViewHelp:
		return m.renderHelp()
	default:
		return m.renderError(fmt.Errorf("unknown view mode: %d", m.view))
	}
}

// renderLoading renders the initial loading screen.
func (m *Model) renderLoading() string {
	style := lipgloss.NewStyle().Foreground(m.theme.Dim)
	return style.Render("Initializing b+...")
}

// renderQuit renders the quit message.
func (m *Model) renderQuit() string {
	style := lipgloss.NewStyle().Foreground(m.theme.Success)
	return style.Render("Thanks for using b+! 👋\n")
}

// renderStartup renders the startup/welcome screen.
func (m *Model) renderStartup() string {
	// ASCII art logo
	logo := `
    ██████╗    ██╗
    ██╔══██╗ ██████╗
    ██████╔╝ ╚═██╔═╝
    ██╔══██╗   ██║
    ██████╔╝   ██║
    ╚═════╝    ╚═╝

    Be Positive
`

	welcome := "Welcome to b+ — Your AI-powered coding assistant\n"
	subtitle := "Model-agnostic • Privacy-first • Developer-friendly"
	prompt := "\nPress Enter to start..."

	// Style the components
	logoStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
	subtleStyle := lipgloss.NewStyle().Foreground(m.theme.Subtle)
	dimStyle := lipgloss.NewStyle().Foreground(m.theme.Dim)

	styledLogo := logoStyle.Render(logo)
	styledWelcome := m.theme.Bold.Render(welcome)
	styledSubtitle := subtleStyle.Render(subtitle)
	styledPrompt := dimStyle.Render(prompt)

	// Combine and center
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styledLogo,
		"",
		styledWelcome,
		styledSubtitle,
		styledPrompt,
	)

	// Center on screen
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// renderChat renders the main chat interface.
func (m *Model) renderChat() string {
	// Calculate heights
	statusBarHeight := 1
	inputHeight := 3
	outputHeight := m.height - statusBarHeight - inputHeight - 2 // -2 for borders

	// Render components
	statusBar := m.renderStatusBar()
	output := m.renderOutput(outputHeight)
	input := m.renderInput(inputHeight)

	// Error display (if any)
	errorDisplay := ""
	if m.err != nil {
		errorDisplay = m.renderErrorBanner()
	}

	// Combine vertically
	view := lipgloss.JoinVertical(
		lipgloss.Left,
		statusBar,
		errorDisplay,
		output,
		input,
	)

	return view
}

// renderStatusBar renders the status bar.
func (m *Model) renderStatusBar() string {
	// TODO: Get actual values from status bar component
	mode := "Fast Mode"
	model := "anthropic/claude-sonnet-4-5"
	cost := "$0.00"
	tokens := "0"

	left := fmt.Sprintf(" %s | %s", mode, model)
	right := fmt.Sprintf("Cost: %s | Tokens: %s ", cost, tokens)

	// Calculate spacing
	spacing := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if spacing < 0 {
		spacing = 0
	}

	statusBar := left + strings.Repeat(" ", spacing) + right

	return m.theme.StatusBar.Width(m.width).Render(statusBar)
}

// renderOutput renders the output/conversation area.
func (m *Model) renderOutput(height int) string {
	// TODO: Replace with actual output component
	placeholder := "Conversation will appear here...\n\n"
	placeholder += "You can ask me to:\n"
	placeholder += "  • Write code\n"
	placeholder += "  • Fix bugs\n"
	placeholder += "  • Refactor\n"
	placeholder += "  • Generate tests\n"
	placeholder += "  • And much more!\n"

	box := lipgloss.NewStyle().
		Width(m.width - 2).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Padding(1).
		Render(placeholder)

	return box
}

// renderInput renders the input area.
func (m *Model) renderInput(height int) string {
	// TODO: Replace with actual input component
	placeholder := "Type your message... (Ctrl+D to quit, ? for help)"

	primaryStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
	subtleStyle := lipgloss.NewStyle().Foreground(m.theme.Subtle)

	prompt := primaryStyle.Render("> ")
	inputText := subtleStyle.Render(placeholder)

	box := lipgloss.NewStyle().
		Width(m.width-2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary).
		Padding(0, 1).
		Render(prompt + inputText)

	return box
}

// renderErrorBanner renders an error banner.
func (m *Model) renderErrorBanner() string {
	if m.err == nil {
		return ""
	}

	errorText := fmt.Sprintf("⚠️  Error: %s", m.err.Error())
	banner := lipgloss.NewStyle().
		Width(m.width).
		Background(m.theme.ErrorBg).
		Foreground(m.theme.ErrorFg).
		Padding(0, 1).
		Render(errorText)

	return banner
}

// renderSettings renders the settings view.
func (m *Model) renderSettings() string {
	subtleStyle := lipgloss.NewStyle().Foreground(m.theme.Subtle)
	dimStyle := lipgloss.NewStyle().Foreground(m.theme.Dim)

	title := m.theme.Bold.Render("⚙️  Settings\n\n")
	content := subtleStyle.Render("Settings interface coming soon...\n\n")
	hint := dimStyle.Render("Press ESC to return to chat")

	settingsContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		hint,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		settingsContent,
	)
}

// renderHelp renders the help overlay.
func (m *Model) renderHelp() string {
	dimStyle := lipgloss.NewStyle().Foreground(m.theme.Dim)

	title := m.theme.Bold.Render("📖 Help\n")

	helpText := `
Keyboard Shortcuts:
  Ctrl+D, Ctrl+C    Quit
  ?                 Toggle this help
  Ctrl+K            Clear screen
  Ctrl+/            Settings

Navigation:
  Tab               Focus next component
  Shift+Tab         Focus previous component

Chat:
  Enter             Send message
  Up/Down           Navigate history

Coming soon:
  • Custom commands
  • Model switching
  • Session management
  • And much more!
`

	styledHelp := m.theme.Render.Render(helpText)
	hint := dimStyle.Render("\nPress ? or ESC to close")

	helpContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		styledHelp,
		hint,
	)

	// Add border
	box := lipgloss.NewStyle().
		Width(m.width - 10).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary).
		Padding(2).
		Render(helpContent)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// renderError renders an error screen.
func (m *Model) renderError(err error) string {
	errorText := fmt.Sprintf("❌ Error: %s", err.Error())
	errorStyle := lipgloss.NewStyle().Foreground(m.theme.Error)
	styled := errorStyle.Render(errorText)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		styled,
	)
}

// Helper method to render a centered box with content.
func (m *Model) renderCenteredBox(title, content string) string {
	titleStyled := m.theme.Bold.Render(title + "\n\n")
	contentStyled := m.theme.Render.Render(content)

	combined := titleStyled + contentStyled

	box := lipgloss.NewStyle().
		Width(m.width - 10).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Padding(2).
		Render(combined)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}
