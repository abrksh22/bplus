package ui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew tests the creation of a new UI model.
func TestNew(t *testing.T) {
	m := New()

	assert.NotNil(t, m, "Model should not be nil")
	assert.False(t, m.ready, "Model should not be ready initially")
	assert.False(t, m.quitting, "Model should not be quitting initially")
	assert.Equal(t, ViewStartup, m.view, "Initial view should be Startup")
	assert.Equal(t, "input", m.focusedComponent, "Initial focus should be input")
	assert.NotNil(t, m.theme, "Theme should be initialized")
	assert.NotNil(t, m.keys, "Keys should be initialized")
}

// TestModelGettersSetters tests model getter and setter methods.
func TestModelGettersSetters(t *testing.T) {
	m := New()

	// Test Width/Height
	m.SetSize(100, 50)
	assert.Equal(t, 100, m.Width())
	assert.Equal(t, 50, m.Height())

	// Test Ready
	assert.False(t, m.IsReady())
	m.SetReady(true)
	assert.True(t, m.IsReady())

	// Test Quit
	assert.False(t, m.IsQuitting())
	m.Quit()
	assert.True(t, m.IsQuitting())

	// Test Error
	assert.Nil(t, m.Error())
	err := errors.New("test error")
	m.SetError(err)
	assert.Equal(t, err, m.Error())

	// Test View
	assert.Equal(t, ViewStartup, m.CurrentView())
	m.SetView(ViewChat)
	assert.Equal(t, ViewChat, m.CurrentView())

	// Test Theme
	theme := LightTheme()
	m.SetTheme(theme)
	assert.Equal(t, theme, m.Theme())

	// Test Focus
	assert.Equal(t, "input", m.FocusedComponent())
	m.SetFocus("output")
	assert.Equal(t, "output", m.FocusedComponent())
}

// TestInit tests the Init method.
func TestInit(t *testing.T) {
	m := New()
	cmd := m.Init()
	assert.Nil(t, cmd, "Init should return nil command")
}

// TestUpdateWindowSize tests window size updates.
func TestUpdateWindowSize(t *testing.T) {
	m := New()

	// Send window size message
	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 60})
	assert.NotNil(t, updated)
	assert.Nil(t, cmd)

	// Check size was updated
	assert.Equal(t, 120, m.Width())
	assert.Equal(t, 60, m.Height())
	assert.True(t, m.IsReady(), "Should be ready after first window size")
}

// TestUpdateCustomMessages tests custom message handling.
func TestUpdateCustomMessages(t *testing.T) {
	m := New()

	tests := []struct {
		name    string
		msg     tea.Msg
		checkFn func(*testing.T, *Model)
	}{
		{
			name: "ErrorMsg",
			msg:  NewErrorMsg(errors.New("test error")),
			checkFn: func(t *testing.T, m *Model) {
				assert.NotNil(t, m.Error())
				assert.Equal(t, "test error", m.Error().Error())
			},
		},
		{
			name: "ClearErrorMsg",
			msg:  ClearErrorMsg{},
			checkFn: func(t *testing.T, m *Model) {
				assert.Nil(t, m.Error())
			},
		},
		{
			name: "ReadyMsg",
			msg:  ReadyMsg{},
			checkFn: func(t *testing.T, m *Model) {
				assert.True(t, m.IsReady())
			},
		},
		{
			name: "ChangeViewMsg",
			msg:  NewChangeViewMsg(ViewSettings),
			checkFn: func(t *testing.T, m *Model) {
				assert.Equal(t, ViewSettings, m.CurrentView())
			},
		},
		{
			name: "ChangeFocusMsg",
			msg:  NewChangeFocusMsg("output"),
			checkFn: func(t *testing.T, m *Model) {
				assert.Equal(t, "output", m.FocusedComponent())
			},
		},
		{
			name: "ChangeThemeMsg",
			msg:  NewChangeThemeMsg(LightTheme()),
			checkFn: func(t *testing.T, m *Model) {
				// Just verify it doesn't panic
				assert.NotNil(t, m.Theme())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, _ := m.Update(tt.msg)
			updatedModel := updated.(*Model)
			tt.checkFn(t, updatedModel)
		})
	}
}

// TestUpdateQuit tests quit message handling.
func TestUpdateQuit(t *testing.T) {
	m := New()

	updated, cmd := m.Update(QuitMsg{})
	assert.NotNil(t, updated)
	assert.NotNil(t, cmd, "Quit should return tea.Quit command")

	updatedModel := updated.(*Model)
	assert.True(t, updatedModel.IsQuitting())
}

// TestView tests view rendering.
func TestView(t *testing.T) {
	m := New()
	m.SetSize(80, 24)

	tests := []struct {
		name    string
		setupFn func(*Model)
		checkFn func(*testing.T, string)
	}{
		{
			name: "Loading view",
			setupFn: func(m *Model) {
				m.SetReady(false)
			},
			checkFn: func(t *testing.T, view string) {
				assert.Contains(t, view, "Initializing")
			},
		},
		{
			name: "Quitting view",
			setupFn: func(m *Model) {
				m.Quit()
			},
			checkFn: func(t *testing.T, view string) {
				assert.Contains(t, view, "Thanks for using b+")
			},
		},
		{
			name: "Startup view",
			setupFn: func(m *Model) {
				m.SetReady(true)
				m.SetView(ViewStartup)
			},
			checkFn: func(t *testing.T, view string) {
				assert.Contains(t, view, "Be Positive")
				assert.Contains(t, view, "Press Enter")
			},
		},
		{
			name: "Chat view",
			setupFn: func(m *Model) {
				m.SetReady(true)
				m.SetView(ViewChat)
			},
			checkFn: func(t *testing.T, view string) {
				assert.Contains(t, view, "Conversation")
			},
		},
		{
			name: "Settings view",
			setupFn: func(m *Model) {
				m.SetReady(true)
				m.SetView(ViewSettings)
			},
			checkFn: func(t *testing.T, view string) {
				assert.Contains(t, view, "Settings")
			},
		},
		{
			name: "Help view",
			setupFn: func(m *Model) {
				m.SetReady(true)
				m.SetView(ViewHelp)
			},
			checkFn: func(t *testing.T, view string) {
				assert.Contains(t, view, "Help")
				assert.Contains(t, view, "Keyboard Shortcuts")
			},
		},
		{
			name: "Error display",
			setupFn: func(m *Model) {
				m.SetReady(true)
				m.SetView(ViewChat)
				m.SetError(errors.New("test error"))
			},
			checkFn: func(t *testing.T, view string) {
				assert.Contains(t, view, "Error:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset model for each test
			m = New()
			m.SetSize(80, 24)
			tt.setupFn(m)

			view := m.View()
			assert.NotEmpty(t, view, "View should not be empty")
			tt.checkFn(t, view)
		})
	}
}

// TestThemes tests all built-in themes.
func TestThemes(t *testing.T) {
	themes := []struct {
		name  string
		theme *Theme
	}{
		{"Dark", DarkTheme()},
		{"Light", LightTheme()},
		{"SolarizedDark", SolarizedDarkTheme()},
		{"SolarizedLight", SolarizedLightTheme()},
		{"Nord", NordTheme()},
		{"Dracula", DraculaTheme()},
	}

	for _, tt := range themes {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.theme, "Theme should not be nil")
			assert.NotEmpty(t, tt.theme.Primary, "Primary color should be set")
			assert.NotEmpty(t, tt.theme.Background, "Background color should be set")
			assert.NotEmpty(t, tt.theme.Foreground, "Foreground color should be set")
			assert.NotNil(t, tt.theme.Bold, "Bold style should be set")
			assert.NotNil(t, tt.theme.StatusBar, "StatusBar style should be set")
		})
	}
}

// TestGetThemeByName tests theme retrieval by name.
func TestGetThemeByName(t *testing.T) {
	tests := []struct {
		name         string
		expected     *Theme
		shouldBeDark bool
	}{
		{"dark", DarkTheme(), true},
		{"light", LightTheme(), false},
		{"solarized-dark", SolarizedDarkTheme(), true},
		{"solarized-light", SolarizedLightTheme(), false},
		{"nord", NordTheme(), true},
		{"dracula", DraculaTheme(), true},
		{"unknown", DefaultTheme(), true}, // Should return default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := GetThemeByName(tt.name)
			assert.NotNil(t, theme)
			// Just verify we got a valid theme
			assert.NotEmpty(t, theme.Primary)
		})
	}
}

// TestThemeNames tests the list of theme names.
func TestThemeNames(t *testing.T) {
	names := ThemeNames()
	assert.NotEmpty(t, names)
	assert.Contains(t, names, "dark")
	assert.Contains(t, names, "light")
	assert.Contains(t, names, "nord")
	assert.Contains(t, names, "dracula")
	assert.Len(t, names, 6)
}

// TestKeyBindings tests key bindings.
func TestKeyBindings(t *testing.T) {
	keys := DefaultKeyMap()

	assert.NotNil(t, keys.Quit)
	assert.NotNil(t, keys.ForceQuit)
	assert.NotNil(t, keys.Help)
	assert.NotNil(t, keys.Send)
	assert.NotNil(t, keys.Cancel)
	assert.NotNil(t, keys.FocusInput)
	assert.NotNil(t, keys.FocusOutput)

	// Test help methods
	shortHelp := keys.ShortHelp()
	assert.NotEmpty(t, shortHelp)

	fullHelp := keys.FullHelp()
	assert.NotEmpty(t, fullHelp)
}

// TestMessageHelpers tests message helper functions.
func TestMessageHelpers(t *testing.T) {
	t.Run("NewErrorMsg", func(t *testing.T) {
		err := errors.New("test")
		msg := NewErrorMsg(err)
		assert.Equal(t, err, msg.Err)
	})

	t.Run("NewWindowSizeMsg", func(t *testing.T) {
		msg := NewWindowSizeMsg(100, 50)
		assert.Equal(t, 100, msg.Width)
		assert.Equal(t, 50, msg.Height)
	})

	t.Run("NewChangeViewMsg", func(t *testing.T) {
		msg := NewChangeViewMsg(ViewChat)
		assert.Equal(t, ViewChat, msg.View)
	})

	t.Run("NewUserInputMsg", func(t *testing.T) {
		msg := NewUserInputMsg("test input")
		assert.Equal(t, "test input", msg.Input)
	})

	t.Run("NewStreamTokenMsg", func(t *testing.T) {
		msg := NewStreamTokenMsg("token", true)
		assert.Equal(t, "token", msg.Token)
		assert.True(t, msg.Done)
	})

	t.Run("NewLoadingMsg", func(t *testing.T) {
		msg := NewLoadingMsg(true, "loading...")
		assert.True(t, msg.Loading)
		assert.Equal(t, "loading...", msg.Message)
	})
}

// BenchmarkUpdate benchmarks the Update method.
func BenchmarkUpdate(b *testing.B) {
	m := New()
	m.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Update(msg)
	}
}

// BenchmarkView benchmarks the View rendering.
func BenchmarkView(b *testing.B) {
	m := New()
	m.SetSize(80, 24)
	m.SetReady(true)
	m.SetView(ViewChat)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}
