package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Message types for Bubble Tea's message passing system.
// These messages are sent through the update loop to trigger state changes.

// WindowSizeMsg is sent when the window is resized.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// ErrorMsg represents an error message to display.
type ErrorMsg struct {
	Err error
}

// ReadyMsg is sent when the UI is fully initialized and ready to display.
type ReadyMsg struct{}

// QuitMsg is sent when the application should quit.
type QuitMsg struct{}

// ChangeViewMsg is sent to switch between different views.
type ChangeViewMsg struct {
	View ViewMode
}

// ChangeFocusMsg is sent to change the focused component.
type ChangeFocusMsg struct {
	Component string
}

// ChangeThemeMsg is sent to switch themes.
type ChangeThemeMsg struct {
	Theme *Theme
}

// ClearErrorMsg is sent to clear the current error.
type ClearErrorMsg struct{}

// UserInputMsg represents user text input.
type UserInputMsg struct {
	Input string
}

// StreamTokenMsg represents a token received from streaming LLM response.
type StreamTokenMsg struct {
	Token string
	Done  bool
}

// LoadingMsg is sent to show/hide loading indicator.
type LoadingMsg struct {
	Loading bool
	Message string
}

// ProgressMsg is sent to update progress bar.
type ProgressMsg struct {
	Current int
	Total   int
	Message string
}

// ShowModalMsg is sent to display a modal dialog.
type ShowModalMsg struct {
	Title   string
	Message string
	Type    ModalType
	Buttons []string
}

// ModalType represents the type of modal dialog.
type ModalType int

const (
	ModalTypeInfo ModalType = iota
	ModalTypeWarning
	ModalTypeError
	ModalTypeConfirm
)

// ModalResultMsg is sent when a modal is dismissed with a selection.
type ModalResultMsg struct {
	Button string
}

// StatusUpdateMsg is sent to update the status bar.
type StatusUpdateMsg struct {
	Mode   string
	Model  string
	Cost   float64
	Tokens int
}

// ShowHelpMsg is sent to show/hide the help overlay.
type ShowHelpMsg struct {
	Show bool
}

// KeyPressMsg wraps tea.KeyMsg for internal handling.
type KeyPressMsg tea.KeyMsg

// ComponentMsg is a generic message for component-specific updates.
type ComponentMsg struct {
	Component string
	Data      interface{}
}

// Helper functions to create messages

// NewErrorMsg creates a new error message.
func NewErrorMsg(err error) ErrorMsg {
	return ErrorMsg{Err: err}
}

// NewWindowSizeMsg creates a new window size message.
func NewWindowSizeMsg(width, height int) WindowSizeMsg {
	return WindowSizeMsg{Width: width, Height: height}
}

// NewChangeViewMsg creates a new change view message.
func NewChangeViewMsg(view ViewMode) ChangeViewMsg {
	return ChangeViewMsg{View: view}
}

// NewChangeFocusMsg creates a new change focus message.
func NewChangeFocusMsg(component string) ChangeFocusMsg {
	return ChangeFocusMsg{Component: component}
}

// NewChangeThemeMsg creates a new change theme message.
func NewChangeThemeMsg(theme *Theme) ChangeThemeMsg {
	return ChangeThemeMsg{Theme: theme}
}

// NewUserInputMsg creates a new user input message.
func NewUserInputMsg(input string) UserInputMsg {
	return UserInputMsg{Input: input}
}

// NewStreamTokenMsg creates a new stream token message.
func NewStreamTokenMsg(token string, done bool) StreamTokenMsg {
	return StreamTokenMsg{Token: token, Done: done}
}

// NewLoadingMsg creates a new loading message.
func NewLoadingMsg(loading bool, message string) LoadingMsg {
	return LoadingMsg{Loading: loading, Message: message}
}

// NewProgressMsg creates a new progress message.
func NewProgressMsg(current, total int, message string) ProgressMsg {
	return ProgressMsg{Current: current, Total: total, Message: message}
}

// NewShowModalMsg creates a new show modal message.
func NewShowModalMsg(title, message string, modalType ModalType, buttons []string) ShowModalMsg {
	return ShowModalMsg{Title: title, Message: message, Type: modalType, Buttons: buttons}
}

// NewModalResultMsg creates a new modal result message.
func NewModalResultMsg(button string) ModalResultMsg {
	return ModalResultMsg{Button: button}
}

// NewStatusUpdateMsg creates a new status update message.
func NewStatusUpdateMsg(mode, model string, cost float64, tokens int) StatusUpdateMsg {
	return StatusUpdateMsg{Mode: mode, Model: model, Cost: cost, Tokens: tokens}
}

// NewShowHelpMsg creates a new show help message.
func NewShowHelpMsg(show bool) ShowHelpMsg {
	return ShowHelpMsg{Show: show}
}

// NewComponentMsg creates a new component-specific message.
func NewComponentMsg(component string, data interface{}) ComponentMsg {
	return ComponentMsg{Component: component, Data: data}
}
