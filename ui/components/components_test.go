package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test InputComponent
func TestNewInput(t *testing.T) {
	input := NewInput("Test placeholder", 80, 3)
	assert.Equal(t, "Test placeholder", input.placeholder)
	assert.Equal(t, 80, input.width)
	assert.Equal(t, 3, input.height)
	assert.False(t, input.focused)
}

func TestInputComponent_SetValue(t *testing.T) {
	input := NewInput("", 80, 3)
	input.SetValue("test value")
	assert.Equal(t, "test value", input.Value())
}

func TestInputComponent_Clear(t *testing.T) {
	input := NewInput("", 80, 3)
	input.SetValue("test")
	input.Clear()
	assert.Empty(t, input.Value())
}

func TestInputComponent_History(t *testing.T) {
	input := NewInput("", 80, 3)
	input.addToHistory("command 1")
	input.addToHistory("command 2")
	history := input.GetHistory()
	assert.Len(t, history, 2)
	assert.Equal(t, "command 1", history[0])
	assert.Equal(t, "command 2", history[1])
}

// Test OutputComponent
func TestNewOutput(t *testing.T) {
	output := NewOutput(80, 24)
	assert.Equal(t, 80, output.width)
	assert.Equal(t, 24, output.height)
	assert.True(t, output.autoScroll)
	assert.Len(t, output.messages, 0)
}

func TestOutputComponent_AddMessage(t *testing.T) {
	output := NewOutput(80, 24)
	output.Init()
	output.AddMessage("user", "Hello!")
	messages := output.GetMessages()
	require.Len(t, messages, 1)
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "Hello!", messages[0].Content)
}

func TestOutputComponent_StreamToken(t *testing.T) {
	output := NewOutput(80, 24)
	output.Init()
	output.StreamToken("Hello")
	output.StreamToken(" ")
	output.StreamToken("World")
	messages := output.GetMessages()
	require.Len(t, messages, 1)
	assert.Equal(t, "Hello World", messages[0].Content)
	assert.True(t, messages[0].Streaming)
}

func TestOutputComponent_FinishStreaming(t *testing.T) {
	output := NewOutput(80, 24)
	output.Init()
	output.StreamToken("Test")
	output.FinishStreaming()
	messages := output.GetMessages()
	require.Len(t, messages, 1)
	assert.False(t, messages[0].Streaming)
}

func TestOutputComponent_Clear(t *testing.T) {
	output := NewOutput(80, 24)
	output.Init()
	output.AddMessage("user", "Test")
	output.Clear()
	assert.Len(t, output.GetMessages(), 0)
}

// Test StatusBar
func TestNewStatusBar(t *testing.T) {
	statusBar := NewStatusBar(80)
	assert.Equal(t, 80, statusBar.width)
	assert.Equal(t, "Fast", statusBar.mode)
	assert.False(t, statusBar.connected)
}

func TestStatusBar_SetModel(t *testing.T) {
	statusBar := NewStatusBar(80)
	statusBar.SetModel("anthropic/claude-sonnet-4-5")
	assert.Equal(t, "anthropic/claude-sonnet-4-5", statusBar.GetModel())
}

func TestStatusBar_SetTokens(t *testing.T) {
	statusBar := NewStatusBar(80)
	statusBar.SetTokens(100, 50)
	input, output := statusBar.GetTokens()
	assert.Equal(t, 100, input)
	assert.Equal(t, 50, output)
}

func TestStatusBar_AddTokens(t *testing.T) {
	statusBar := NewStatusBar(80)
	statusBar.SetTokens(100, 50)
	statusBar.AddTokens(25, 10)
	input, output := statusBar.GetTokens()
	assert.Equal(t, 125, input)
	assert.Equal(t, 60, output)
}

func TestStatusBar_SetCost(t *testing.T) {
	statusBar := NewStatusBar(80)
	statusBar.SetCost(1.25)
	assert.Equal(t, 1.25, statusBar.GetCost())
}

func TestStatusBar_AddCost(t *testing.T) {
	statusBar := NewStatusBar(80)
	statusBar.SetCost(1.0)
	statusBar.AddCost(0.5)
	assert.Equal(t, 1.5, statusBar.GetCost())
}

func TestStatusBar_Reset(t *testing.T) {
	statusBar := NewStatusBar(80)
	statusBar.SetTokens(100, 50)
	statusBar.SetCost(1.5)
	statusBar.Reset()
	input, output := statusBar.GetTokens()
	assert.Equal(t, 0, input)
	assert.Equal(t, 0, output)
	assert.Equal(t, 0.0, statusBar.GetCost())
}

// Test Spinner
func TestNewSpinner(t *testing.T) {
	spinner := NewSpinner("Loading...")
	assert.Equal(t, "Loading...", spinner.label)
	assert.False(t, spinner.active)
}

func TestSpinner_StartStop(t *testing.T) {
	spinner := NewSpinner("Loading...")
	assert.False(t, spinner.IsActive())
	spinner.Start()
	assert.True(t, spinner.IsActive())
	spinner.Stop()
	assert.False(t, spinner.IsActive())
}

func TestSpinner_SetLabel(t *testing.T) {
	spinner := NewSpinner("Loading...")
	spinner.SetLabel("Processing...")
	assert.Equal(t, "Processing...", spinner.label)
}

// Test ProgressBar
func TestNewProgressBar(t *testing.T) {
	progress := NewProgressBar(100)
	assert.Equal(t, 100, progress.total)
	assert.Equal(t, 0, progress.current)
	assert.Equal(t, "determinate", progress.mode)
}

func TestProgressBar_SetProgress(t *testing.T) {
	progress := NewProgressBar(100)
	progress.SetProgress(50)
	current, total := progress.GetProgress()
	assert.Equal(t, 50, current)
	assert.Equal(t, 100, total)
}

func TestProgressBar_Increment(t *testing.T) {
	progress := NewProgressBar(10)
	progress.Increment()
	progress.Increment()
	current, _ := progress.GetProgress()
	assert.Equal(t, 2, current)
}

func TestProgressBar_GetPercentage(t *testing.T) {
	progress := NewProgressBar(100)
	progress.SetProgress(50)
	assert.Equal(t, 50.0, progress.GetPercentage())
}

func TestProgressBar_IsComplete(t *testing.T) {
	progress := NewProgressBar(10)
	assert.False(t, progress.IsComplete())
	progress.SetProgress(10)
	assert.True(t, progress.IsComplete())
}

func TestProgressBar_Reset(t *testing.T) {
	progress := NewProgressBar(10)
	progress.SetProgress(5)
	progress.Reset()
	current, _ := progress.GetProgress()
	assert.Equal(t, 0, current)
}

// Test Modal
func TestNewModal(t *testing.T) {
	modal := NewModal("Test Title", "Test Content")
	assert.Equal(t, "Test Title", modal.title)
	assert.Equal(t, "Test Content", modal.content)
	assert.False(t, modal.visible)
}

func TestModal_ShowHide(t *testing.T) {
	modal := NewModal("Title", "Content")
	assert.False(t, modal.IsVisible())
	modal.Show()
	assert.True(t, modal.IsVisible())
	modal.Hide()
	assert.False(t, modal.IsVisible())
}

func TestModal_SetButtons(t *testing.T) {
	modal := NewModal("Title", "Content")
	modal.SetButtons([]string{"Yes", "No", "Cancel"})
	assert.Len(t, modal.buttons, 3)
}

// Test List
func TestNewList(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1", Description: "First item"},
		{Title: "Item 2", Description: "Second item"},
	}
	list := NewList(items)
	assert.Len(t, list.items, 2)
	assert.Equal(t, 0, list.selected)
}

func TestList_SelectedItem(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1"},
		{Title: "Item 2"},
	}
	list := NewList(items)
	selected := list.SelectedItem()
	require.NotNil(t, selected)
	assert.Equal(t, "Item 1", selected.Title)
}

func TestList_SetFilter(t *testing.T) {
	items := []ListItem{
		{Title: "Apple"},
		{Title: "Banana"},
		{Title: "Orange"},
	}
	list := NewList(items)
	list.SetFilter("an")
	filtered := list.getFilteredItems()
	assert.Len(t, filtered, 2) // Banana, Orange (both contain "an")
}

func TestList_Navigation(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1"},
		{Title: "Item 2"},
		{Title: "Item 3"},
	}
	list := NewList(items)

	// Test down navigation
	list.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, list.GetSelected())

	// Test up navigation
	list.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, list.GetSelected())
}

// Test SplitPane
func TestNewSplitPane(t *testing.T) {
	// Create simple models for testing
	left := &simpleModel{}
	right := &simpleModel{}
	splitPane := NewSplitPane(left, right, "horizontal")

	assert.Equal(t, "horizontal", splitPane.direction)
	assert.Equal(t, 0.5, splitPane.split)
	assert.Equal(t, "left", splitPane.focused)
}

func TestSplitPane_SetSplit(t *testing.T) {
	left := &simpleModel{}
	right := &simpleModel{}
	splitPane := NewSplitPane(left, right, "horizontal")

	splitPane.SetSplit(0.7)
	assert.Equal(t, 0.7, splitPane.GetSplit())

	// Test boundaries
	splitPane.SetSplit(0.05) // Too small
	assert.Equal(t, 0.1, splitPane.GetSplit())

	splitPane.SetSplit(0.95) // Too large
	assert.Equal(t, 0.9, splitPane.GetSplit())
}

func TestSplitPane_FocusToggle(t *testing.T) {
	left := &simpleModel{}
	right := &simpleModel{}
	splitPane := NewSplitPane(left, right, "horizontal")

	assert.Equal(t, "left", splitPane.GetFocused())
	splitPane.FocusRight()
	assert.Equal(t, "right", splitPane.GetFocused())
	splitPane.FocusLeft()
	assert.Equal(t, "left", splitPane.GetFocused())
}

// Simple model for testing
type simpleModel struct{}

func (m *simpleModel) Init() tea.Cmd                           { return nil }
func (m *simpleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *simpleModel) View() string                            { return "test" }
