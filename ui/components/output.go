package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Message represents a single message in the conversation.
type Message struct {
	Role      string // "user", "assistant", "system"
	Content   string // Message text (supports markdown)
	Timestamp time.Time
	Streaming bool // Currently streaming
}

// OutputComponent displays the conversation messages with markdown rendering.
type OutputComponent struct {
	messages    []Message
	viewport    viewport.Model
	autoScroll  bool
	theme       OutputTheme
	width       int
	height      int
	renderer    *glamour.TermRenderer
	initialized bool
}

// OutputTheme defines the color scheme for the output component.
type OutputTheme struct {
	UserBubble      lipgloss.Style
	AssistantBubble lipgloss.Style
	SystemBubble    lipgloss.Style
	Timestamp       lipgloss.Color
	Border          lipgloss.Color
}

// DefaultOutputTheme returns the default output theme.
func DefaultOutputTheme() OutputTheme {
	return OutputTheme{
		UserBubble: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7dcfff")).
			Padding(0, 1).
			MarginBottom(1),
		AssistantBubble: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#9ece6a")).
			Padding(0, 1).
			MarginBottom(1),
		SystemBubble: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#565f89")).
			Padding(0, 1).
			MarginBottom(1).
			Faint(true),
		Timestamp: lipgloss.Color("#565f89"),
		Border:    lipgloss.Color("#3b4261"),
	}
}

// NewOutput creates a new output component.
func NewOutput(width, height int) OutputComponent {
	vp := viewport.New(width-2, height-2)
	vp.YPosition = 0

	// Create glamour renderer for markdown
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-6),
	)

	return OutputComponent{
		messages:    make([]Message, 0),
		viewport:    vp,
		autoScroll:  true,
		theme:       DefaultOutputTheme(),
		width:       width,
		height:      height,
		renderer:    renderer,
		initialized: false,
	}
}

// Init initializes the output component.
func (o *OutputComponent) Init() tea.Cmd {
	o.initialized = true
	return nil
}

// Update handles messages (Bubble Tea Update method).
func (o *OutputComponent) Update(msg tea.Msg) (*OutputComponent, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "pgup", "pgdown", "up", "down":
			// User is scrolling, disable auto-scroll
			o.autoScroll = false
		case "g":
			// 'g' to go to top
			o.viewport.GotoTop()
			o.autoScroll = false
		case "G":
			// 'G' to go to bottom and re-enable auto-scroll
			o.viewport.GotoBottom()
			o.autoScroll = true
		}
	}

	// Update viewport
	o.viewport, cmd = o.viewport.Update(msg)

	// Check if at bottom to re-enable auto-scroll
	if o.viewport.AtBottom() {
		o.autoScroll = true
	}

	return o, cmd
}

// View renders the output component.
func (o *OutputComponent) View() string {
	if !o.initialized {
		return "Loading..."
	}

	// Render all messages
	content := o.renderMessages()
	o.viewport.SetContent(content)

	// Create border around viewport
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(o.theme.Border).
		Width(o.width - 2).
		Height(o.height)

	return boxStyle.Render(o.viewport.View())
}

// AddMessage adds a new message to the conversation.
func (o *OutputComponent) AddMessage(role, content string) {
	msg := Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
		Streaming: false,
	}
	o.messages = append(o.messages, msg)

	// Auto-scroll to bottom if enabled
	if o.autoScroll {
		o.viewport.GotoBottom()
	}
}

// StreamToken adds a token to the last message (for streaming).
func (o *OutputComponent) StreamToken(token string) {
	if len(o.messages) == 0 {
		// No messages yet, create a new one
		o.messages = append(o.messages, Message{
			Role:      "assistant",
			Content:   token,
			Timestamp: time.Now(),
			Streaming: true,
		})
	} else {
		// Append to last message
		lastIdx := len(o.messages) - 1
		o.messages[lastIdx].Content += token
		o.messages[lastIdx].Streaming = true
	}

	// Auto-scroll to bottom if enabled
	if o.autoScroll {
		o.viewport.GotoBottom()
	}
}

// FinishStreaming marks the last message as complete.
func (o *OutputComponent) FinishStreaming() {
	if len(o.messages) > 0 {
		lastIdx := len(o.messages) - 1
		o.messages[lastIdx].Streaming = false
	}
}

// SetSize updates the dimensions of the output component.
func (o *OutputComponent) SetSize(width, height int) {
	o.width = width
	o.height = height
	o.viewport.Width = width - 4
	o.viewport.Height = height - 2

	// Recreate renderer with new width
	if o.renderer != nil {
		o.renderer, _ = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width-8),
		)
	}
}

// Clear clears all messages.
func (o *OutputComponent) Clear() {
	o.messages = make([]Message, 0)
	o.viewport.SetContent("")
}

// GetMessages returns all messages.
func (o *OutputComponent) GetMessages() []Message {
	return o.messages
}

// SetAutoScroll sets whether to auto-scroll to bottom on new messages.
func (o *OutputComponent) SetAutoScroll(enabled bool) {
	o.autoScroll = enabled
}

// IsAutoScroll returns whether auto-scroll is enabled.
func (o *OutputComponent) IsAutoScroll() bool {
	return o.autoScroll
}

// SetTheme sets the color theme for the output.
func (o *OutputComponent) SetTheme(theme OutputTheme) {
	o.theme = theme
}

// renderMessages renders all messages as a string.
func (o *OutputComponent) renderMessages() string {
	if len(o.messages) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#565f89")).Faint(true)
		return emptyStyle.Render("No messages yet. Start a conversation!")
	}

	var rendered []string
	for _, msg := range o.messages {
		rendered = append(rendered, o.renderMessage(msg))
	}

	return strings.Join(rendered, "\n")
}

// renderMessage renders a single message.
func (o *OutputComponent) renderMessage(msg Message) string {
	// Get bubble style based on role
	var bubbleStyle lipgloss.Style
	var roleLabel string
	switch msg.Role {
	case "user":
		bubbleStyle = o.theme.UserBubble
		roleLabel = "You"
	case "assistant":
		bubbleStyle = o.theme.AssistantBubble
		roleLabel = "b+"
	case "system":
		bubbleStyle = o.theme.SystemBubble
		roleLabel = "System"
	default:
		bubbleStyle = o.theme.SystemBubble
		roleLabel = msg.Role
	}

	// Render timestamp
	timestampStyle := lipgloss.NewStyle().Foreground(o.theme.Timestamp).Faint(true)
	timestamp := timestampStyle.Render(msg.Timestamp.Format("15:04:05"))

	// Role and timestamp header
	headerStyle := lipgloss.NewStyle().Bold(true)
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		headerStyle.Render(roleLabel),
		"  ",
		timestamp,
	)

	// Render content (try markdown, fall back to plain text)
	content := msg.Content
	if o.renderer != nil {
		rendered, err := o.renderer.Render(content)
		if err == nil {
			content = rendered
		}
	}

	// Add streaming indicator if streaming
	if msg.Streaming {
		content += " ▊" // Streaming cursor
	}

	// Combine header and content
	messageContent := lipgloss.JoinVertical(lipgloss.Left, header, "", content)

	// Apply bubble style
	return bubbleStyle.Width(o.width - 6).Render(messageContent)
}

// ExportToString exports all messages to a plain text string.
func (o *OutputComponent) ExportToString() string {
	var lines []string
	lines = append(lines, "=== b+ Conversation Export ===")
	lines = append(lines, fmt.Sprintf("Exported: %s", time.Now().Format(time.RFC3339)))
	lines = append(lines, fmt.Sprintf("Messages: %d", len(o.messages)))
	lines = append(lines, "")

	for i, msg := range o.messages {
		lines = append(lines, fmt.Sprintf("--- Message %d ---", i+1))
		lines = append(lines, fmt.Sprintf("Role: %s", msg.Role))
		lines = append(lines, fmt.Sprintf("Time: %s", msg.Timestamp.Format(time.RFC3339)))
		lines = append(lines, "")
		lines = append(lines, msg.Content)
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// ScrollPercentage returns the current scroll percentage (0-100).
func (o *OutputComponent) ScrollPercentage() float64 {
	return o.viewport.ScrollPercent() * 100
}
