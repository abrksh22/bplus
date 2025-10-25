package main

import (
	"fmt"
	"os"

	"github.com/abrksh22/bplus/ui"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	// Version information (set during build)
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	// Parse command-line flags (TODO: Implement in later phase)
	// For now, just start the UI

	// Create the UI model
	model := ui.New()

	// Create the Bubble Tea program
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Start the program
	finalModel, err := program.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running b+: %v\n", err)
		os.Exit(1)
	}

	// Check if there was an error in the final model
	if m, ok := finalModel.(*ui.Model); ok {
		if m.Error() != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", m.Error())
			os.Exit(1)
		}
	}

	os.Exit(0)
}
