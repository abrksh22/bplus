package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines all colors and styles for the UI.
type Theme struct {
	// Base colors
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Background lipgloss.Color
	Foreground lipgloss.Color
	Border     lipgloss.Color

	// Semantic colors
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
	Info    lipgloss.Color
	Subtle  lipgloss.Color
	Dim     lipgloss.Color

	// Component-specific colors
	InputBorder  lipgloss.Color
	OutputBorder lipgloss.Color
	StatusBarBg  lipgloss.Color
	StatusBarFg  lipgloss.Color
	ErrorBg      lipgloss.Color
	ErrorFg      lipgloss.Color

	// Syntax highlighting colors
	Keyword  lipgloss.Color
	String   lipgloss.Color
	Number   lipgloss.Color
	Comment  lipgloss.Color
	Function lipgloss.Color

	// Styles (pre-configured lipgloss styles)
	Bold      lipgloss.Style
	Italic    lipgloss.Style
	Underline lipgloss.Style
	Render    lipgloss.Style
	StatusBar lipgloss.Style
}

// DefaultTheme returns the default dark theme.
func DefaultTheme() *Theme {
	return DarkTheme()
}

// DarkTheme returns a dark color scheme.
func DarkTheme() *Theme {
	primary := lipgloss.Color("#7C3AED")    // Purple
	secondary := lipgloss.Color("#3B82F6")  // Blue
	background := lipgloss.Color("#1E1E2E") // Dark background
	foreground := lipgloss.Color("#CDD6F4") // Light text
	border := lipgloss.Color("#45475A")     // Gray border
	success := lipgloss.Color("#A6E3A1")    // Green
	warning := lipgloss.Color("#F9E2AF")    // Yellow
	errorColor := lipgloss.Color("#F38BA8") // Red
	info := lipgloss.Color("#89DCEB")       // Cyan
	subtle := lipgloss.Color("#BAC2DE")     // Light gray
	dim := lipgloss.Color("#6C7086")        // Dim gray

	return &Theme{
		Primary:      primary,
		Secondary:    secondary,
		Background:   background,
		Foreground:   foreground,
		Border:       border,
		Success:      success,
		Warning:      warning,
		Error:        errorColor,
		Info:         info,
		Subtle:       subtle,
		Dim:          dim,
		InputBorder:  primary,
		OutputBorder: border,
		StatusBarBg:  lipgloss.Color("#313244"),
		StatusBarFg:  foreground,
		ErrorBg:      errorColor,
		ErrorFg:      background,
		Keyword:      lipgloss.Color("#CBA6F7"),
		String:       lipgloss.Color("#A6E3A1"),
		Number:       lipgloss.Color("#FAB387"),
		Comment:      dim,
		Function:     lipgloss.Color("#89B4FA"),

		// Styles
		Bold:      lipgloss.NewStyle().Bold(true).Foreground(foreground),
		Italic:    lipgloss.NewStyle().Italic(true).Foreground(subtle),
		Underline: lipgloss.NewStyle().Underline(true),
		Render:    lipgloss.NewStyle().Foreground(foreground),
		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("#313244")).
			Foreground(foreground).
			Bold(true),
	}
}

// LightTheme returns a light color scheme.
func LightTheme() *Theme {
	primary := lipgloss.Color("#7C3AED")    // Purple
	secondary := lipgloss.Color("#3B82F6")  // Blue
	background := lipgloss.Color("#EFF1F5") // Light background
	foreground := lipgloss.Color("#4C4F69") // Dark text
	border := lipgloss.Color("#ACB0BE")     // Gray border
	success := lipgloss.Color("#40A02B")    // Green
	warning := lipgloss.Color("#DF8E1D")    // Yellow
	errorColor := lipgloss.Color("#D20F39") // Red
	info := lipgloss.Color("#209FB5")       // Cyan
	subtle := lipgloss.Color("#6C6F85")     // Gray
	dim := lipgloss.Color("#9CA0B0")        // Dim gray

	return &Theme{
		Primary:      primary,
		Secondary:    secondary,
		Background:   background,
		Foreground:   foreground,
		Border:       border,
		Success:      success,
		Warning:      warning,
		Error:        errorColor,
		Info:         info,
		Subtle:       subtle,
		Dim:          dim,
		InputBorder:  primary,
		OutputBorder: border,
		StatusBarBg:  lipgloss.Color("#DCE0E8"),
		StatusBarFg:  foreground,
		ErrorBg:      errorColor,
		ErrorFg:      background,
		Keyword:      lipgloss.Color("#8839EF"),
		String:       lipgloss.Color("#40A02B"),
		Number:       lipgloss.Color("#FE640B"),
		Comment:      dim,
		Function:     lipgloss.Color("#1E66F5"),

		// Styles
		Bold:      lipgloss.NewStyle().Bold(true).Foreground(foreground),
		Italic:    lipgloss.NewStyle().Italic(true).Foreground(subtle),
		Underline: lipgloss.NewStyle().Underline(true),
		Render:    lipgloss.NewStyle().Foreground(foreground),
		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("#DCE0E8")).
			Foreground(foreground).
			Bold(true),
	}
}

// SolarizedDarkTheme returns the Solarized Dark color scheme.
func SolarizedDarkTheme() *Theme {
	base03 := lipgloss.Color("#002b36")
	base02 := lipgloss.Color("#073642")
	base01 := lipgloss.Color("#586e75")
	base0 := lipgloss.Color("#839496")
	base1 := lipgloss.Color("#93a1a1")
	yellow := lipgloss.Color("#b58900")
	orange := lipgloss.Color("#cb4b16")
	red := lipgloss.Color("#dc322f")
	magenta := lipgloss.Color("#d33682")
	violet := lipgloss.Color("#6c71c4")
	blue := lipgloss.Color("#268bd2")
	cyan := lipgloss.Color("#2aa198")
	green := lipgloss.Color("#859900")

	return &Theme{
		Primary:      violet,
		Secondary:    blue,
		Background:   base03,
		Foreground:   base0,
		Border:       base01,
		Success:      green,
		Warning:      yellow,
		Error:        red,
		Info:         cyan,
		Subtle:       base1,
		Dim:          base01,
		InputBorder:  violet,
		OutputBorder: base01,
		StatusBarBg:  base02,
		StatusBarFg:  base0,
		ErrorBg:      red,
		ErrorFg:      base03,
		Keyword:      magenta,
		String:       green,
		Number:       orange,
		Comment:      base01,
		Function:     blue,

		// Styles
		Bold:      lipgloss.NewStyle().Bold(true).Foreground(base0),
		Italic:    lipgloss.NewStyle().Italic(true).Foreground(base1),
		Underline: lipgloss.NewStyle().Underline(true),
		Render:    lipgloss.NewStyle().Foreground(base0),
		StatusBar: lipgloss.NewStyle().
			Background(base02).
			Foreground(base0).
			Bold(true),
	}
}

// SolarizedLightTheme returns the Solarized Light color scheme.
func SolarizedLightTheme() *Theme {
	base3 := lipgloss.Color("#fdf6e3")
	base2 := lipgloss.Color("#eee8d5")
	base1 := lipgloss.Color("#93a1a1")
	base00 := lipgloss.Color("#657b83")
	base01 := lipgloss.Color("#586e75")
	yellow := lipgloss.Color("#b58900")
	orange := lipgloss.Color("#cb4b16")
	red := lipgloss.Color("#dc322f")
	magenta := lipgloss.Color("#d33682")
	violet := lipgloss.Color("#6c71c4")
	blue := lipgloss.Color("#268bd2")
	cyan := lipgloss.Color("#2aa198")
	green := lipgloss.Color("#859900")

	return &Theme{
		Primary:      violet,
		Secondary:    blue,
		Background:   base3,
		Foreground:   base00,
		Border:       base1,
		Success:      green,
		Warning:      yellow,
		Error:        red,
		Info:         cyan,
		Subtle:       base01,
		Dim:          base1,
		InputBorder:  violet,
		OutputBorder: base1,
		StatusBarBg:  base2,
		StatusBarFg:  base00,
		ErrorBg:      red,
		ErrorFg:      base3,
		Keyword:      magenta,
		String:       green,
		Number:       orange,
		Comment:      base1,
		Function:     blue,

		// Styles
		Bold:      lipgloss.NewStyle().Bold(true).Foreground(base00),
		Italic:    lipgloss.NewStyle().Italic(true).Foreground(base01),
		Underline: lipgloss.NewStyle().Underline(true),
		Render:    lipgloss.NewStyle().Foreground(base00),
		StatusBar: lipgloss.NewStyle().
			Background(base2).
			Foreground(base00).
			Bold(true),
	}
}

// NordTheme returns the Nord color scheme.
func NordTheme() *Theme {
	nord0 := lipgloss.Color("#2E3440")
	nord1 := lipgloss.Color("#3B4252")
	nord3 := lipgloss.Color("#4C566A")
	nord4 := lipgloss.Color("#D8DEE9")
	nord6 := lipgloss.Color("#ECEFF4")
	nord8 := lipgloss.Color("#88C0D0")
	nord9 := lipgloss.Color("#81A1C1")
	nord10 := lipgloss.Color("#5E81AC")
	nord11 := lipgloss.Color("#BF616A")
	nord12 := lipgloss.Color("#D08770")
	nord13 := lipgloss.Color("#EBCB8B")
	nord14 := lipgloss.Color("#A3BE8C")
	nord15 := lipgloss.Color("#B48EAD")

	return &Theme{
		Primary:      nord10,
		Secondary:    nord9,
		Background:   nord0,
		Foreground:   nord4,
		Border:       nord3,
		Success:      nord14,
		Warning:      nord13,
		Error:        nord11,
		Info:         nord8,
		Subtle:       nord6,
		Dim:          nord3,
		InputBorder:  nord10,
		OutputBorder: nord3,
		StatusBarBg:  nord1,
		StatusBarFg:  nord4,
		ErrorBg:      nord11,
		ErrorFg:      nord0,
		Keyword:      nord15,
		String:       nord14,
		Number:       nord12,
		Comment:      nord3,
		Function:     nord8,

		// Styles
		Bold:      lipgloss.NewStyle().Bold(true).Foreground(nord4),
		Italic:    lipgloss.NewStyle().Italic(true).Foreground(nord6),
		Underline: lipgloss.NewStyle().Underline(true),
		Render:    lipgloss.NewStyle().Foreground(nord4),
		StatusBar: lipgloss.NewStyle().
			Background(nord1).
			Foreground(nord4).
			Bold(true),
	}
}

// DraculaTheme returns the Dracula color scheme.
func DraculaTheme() *Theme {
	background := lipgloss.Color("#282a36")
	currentLine := lipgloss.Color("#44475a")
	foreground := lipgloss.Color("#f8f8f2")
	comment := lipgloss.Color("#6272a4")
	cyan := lipgloss.Color("#8be9fd")
	green := lipgloss.Color("#50fa7b")
	orange := lipgloss.Color("#ffb86c")
	pink := lipgloss.Color("#ff79c6")
	purple := lipgloss.Color("#bd93f9")
	red := lipgloss.Color("#ff5555")
	yellow := lipgloss.Color("#f1fa8c")

	return &Theme{
		Primary:      purple,
		Secondary:    pink,
		Background:   background,
		Foreground:   foreground,
		Border:       currentLine,
		Success:      green,
		Warning:      yellow,
		Error:        red,
		Info:         cyan,
		Subtle:       comment,
		Dim:          comment,
		InputBorder:  purple,
		OutputBorder: currentLine,
		StatusBarBg:  currentLine,
		StatusBarFg:  foreground,
		ErrorBg:      red,
		ErrorFg:      background,
		Keyword:      pink,
		String:       yellow,
		Number:       orange,
		Comment:      comment,
		Function:     green,

		// Styles
		Bold:      lipgloss.NewStyle().Bold(true).Foreground(foreground),
		Italic:    lipgloss.NewStyle().Italic(true).Foreground(comment),
		Underline: lipgloss.NewStyle().Underline(true),
		Render:    lipgloss.NewStyle().Foreground(foreground),
		StatusBar: lipgloss.NewStyle().
			Background(currentLine).
			Foreground(foreground).
			Bold(true),
	}
}

// GetThemeByName returns a theme by name.
func GetThemeByName(name string) *Theme {
	switch name {
	case "dark":
		return DarkTheme()
	case "light":
		return LightTheme()
	case "solarized-dark":
		return SolarizedDarkTheme()
	case "solarized-light":
		return SolarizedLightTheme()
	case "nord":
		return NordTheme()
	case "dracula":
		return DraculaTheme()
	default:
		return DefaultTheme()
	}
}

// ThemeNames returns a list of available theme names.
func ThemeNames() []string {
	return []string{
		"dark",
		"light",
		"solarized-dark",
		"solarized-light",
		"nord",
		"dracula",
	}
}
