package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputComponent handles user text input with history and auto-completion.
type InputComponent struct {
	textarea    textarea.Model
	history     []string
	historyIdx  int
	placeholder string
	focused     bool
	width       int
	height      int
	showCounter bool
	maxChars    int
	onSubmit    func(string)
	theme       InputTheme
}

// InputTheme defines the color scheme for the input component.
type InputTheme struct {
	Focused   lipgloss.Color
	Unfocused lipgloss.Color
	Cursor    lipgloss.Color
	Counter   lipgloss.Color
	Prompt    lipgloss.Color
}

// DefaultInputTheme returns the default input theme.
func DefaultInputTheme() InputTheme {
	return InputTheme{
		Focused:   lipgloss.Color("#7dcfff"),
		Unfocused: lipgloss.Color("#a9b1d6"),
		Cursor:    lipgloss.Color("#7aa2f7"),
		Counter:   lipgloss.Color("#565f89"),
		Prompt:    lipgloss.Color("#bb9af7"),
	}
}

// NewInput creates a new input component.
func NewInput(placeholder string, width, height int) InputComponent {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.ShowLineNumbers = false
	ta.CharLimit = 10000 // Reasonable limit
	ta.SetWidth(width - 4)
	ta.SetHeight(height - 2)

	return InputComponent{
		textarea:    ta,
		history:     make([]string, 0),
		historyIdx:  -1,
		placeholder: placeholder,
		focused:     false,
		width:       width,
		height:      height,
		showCounter: true,
		maxChars:    10000,
		theme:       DefaultInputTheme(),
	}
}

// Update handles messages (Bubble Tea Update method).
func (i *InputComponent) Update(msg tea.Msg) (*InputComponent, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Submit on Enter
			value := i.Value()
			if strings.TrimSpace(value) != "" {
				i.addToHistory(value)
				if i.onSubmit != nil {
					i.onSubmit(value)
				}
				i.Clear()
				return i, nil
			}
		case "up":
			// Navigate history up
			if i.historyIdx < len(i.history)-1 {
				i.historyIdx++
				if i.historyIdx < len(i.history) {
					i.SetValue(i.history[len(i.history)-1-i.historyIdx])
				}
				return i, nil
			}
		case "down":
			// Navigate history down
			if i.historyIdx > 0 {
				i.historyIdx--
				i.SetValue(i.history[len(i.history)-1-i.historyIdx])
				return i, nil
			} else if i.historyIdx == 0 {
				i.historyIdx = -1
				i.Clear()
				return i, nil
			}
		}
	}

	// Pass to underlying textarea
	i.textarea, cmd = i.textarea.Update(msg)
	return i, cmd
}

// View renders the input component.
func (i *InputComponent) View() string {
	// Determine border color based on focus
	borderColor := i.theme.Unfocused
	if i.focused {
		borderColor = i.theme.Focused
	}

	// Create the prompt
	promptStyle := lipgloss.NewStyle().Foreground(i.theme.Prompt).Bold(true)
	prompt := promptStyle.Render("> ")

	// Get textarea content
	textareaView := i.textarea.View()

	// Create counter if enabled
	counter := ""
	if i.showCounter {
		charCount := len(i.Value())
		counterText := ""
		if i.maxChars > 0 {
			counterText = lipgloss.NewStyle().
				Foreground(i.theme.Counter).
				Render(lipgloss.JoinHorizontal(
					lipgloss.Left,
					" ",
					lipgloss.NewStyle().Faint(true).Render("│"),
					" ",
					lipgloss.NewStyle().Render(lipgloss.JoinHorizontal(lipgloss.Left,
						lipgloss.NewStyle().Render(lipgloss.PlaceHorizontal(5, lipgloss.Right, lipgloss.NewStyle().Render(string(rune(charCount/1000+48))+string(rune((charCount/100)%10+48))+string(rune((charCount/10)%10+48))+string(rune(charCount%10+48))))),
						"/",
						lipgloss.NewStyle().Render(lipgloss.PlaceHorizontal(5, lipgloss.Right, lipgloss.NewStyle().Render(string(rune(i.maxChars/1000+48))+string(rune((i.maxChars/100)%10+48))+string(rune((i.maxChars/10)%10+48))+string(rune(i.maxChars%10+48))))),
					)),
				))
		}
		counter = counterText
	}

	// Build the view
	content := lipgloss.JoinHorizontal(lipgloss.Left, prompt, textareaView)

	// Add border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(i.width-2).
		Padding(0, 1)

	box := boxStyle.Render(content)

	// Add counter at the bottom right if enabled
	if counter != "" {
		box = lipgloss.JoinVertical(lipgloss.Left, box, counter)
	}

	return box
}

// Value returns the current input value.
func (i *InputComponent) Value() string {
	return i.textarea.Value()
}

// SetValue sets the input value.
func (i *InputComponent) SetValue(v string) {
	i.textarea.SetValue(v)
}

// Clear clears the input.
func (i *InputComponent) Clear() {
	i.textarea.Reset()
	i.historyIdx = -1
}

// Focus gives focus to the input.
func (i *InputComponent) Focus() tea.Cmd {
	i.focused = true
	return i.textarea.Focus()
}

// Blur removes focus from the input.
func (i *InputComponent) Blur() {
	i.focused = false
	i.textarea.Blur()
}

// IsFocused returns whether the input is focused.
func (i *InputComponent) IsFocused() bool {
	return i.focused
}

// SetWidth sets the width of the input.
func (i *InputComponent) SetWidth(width int) {
	i.width = width
	i.textarea.SetWidth(width - 6)
}

// SetHeight sets the height of the input.
func (i *InputComponent) SetHeight(height int) {
	i.height = height
	i.textarea.SetHeight(height - 2)
}

// SetPlaceholder sets the placeholder text.
func (i *InputComponent) SetPlaceholder(placeholder string) {
	i.placeholder = placeholder
	i.textarea.Placeholder = placeholder
}

// OnSubmit sets the callback function for when input is submitted.
func (i *InputComponent) OnSubmit(fn func(string)) {
	i.onSubmit = fn
}

// SetShowCounter sets whether to show the character counter.
func (i *InputComponent) SetShowCounter(show bool) {
	i.showCounter = show
}

// SetTheme sets the color theme for the input.
func (i *InputComponent) SetTheme(theme InputTheme) {
	i.theme = theme
}

// addToHistory adds a value to the history.
func (i *InputComponent) addToHistory(value string) {
	// Avoid duplicates of the last entry
	if len(i.history) == 0 || i.history[len(i.history)-1] != value {
		i.history = append(i.history, value)
		// Keep history to a reasonable size (last 100 entries)
		if len(i.history) > 100 {
			i.history = i.history[1:]
		}
	}
	i.historyIdx = -1
}

// GetHistory returns the command history.
func (i *InputComponent) GetHistory() []string {
	return i.history
}

// SetHistory sets the command history.
func (i *InputComponent) SetHistory(history []string) {
	i.history = history
	i.historyIdx = -1
}
