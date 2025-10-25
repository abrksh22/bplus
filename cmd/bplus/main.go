package main

import (
	"flag"
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
	// Define command-line flags
	var (
		showVersion  = flag.Bool("version", false, "Show version information")
		showHelp     = flag.Bool("help", false, "Show help message")
		debugMode    = flag.Bool("debug", false, "Enable debug mode")
		fastMode     = flag.Bool("fast", false, "Run in Fast Mode (Layer 4 only)")
		thoroughMode = flag.Bool("thorough", false, "Run in Thorough Mode (all 7 layers)")
		configFile   = flag.String("config", "", "Path to config file")
	)

	// Short flags
	flag.BoolVar(showVersion, "v", false, "Show version information (shorthand)")
	flag.BoolVar(showHelp, "h", false, "Show help message (shorthand)")

	flag.Parse()

	// Handle --version flag
	if *showVersion {
		fmt.Printf("b+ (Be Positive) version %s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Built: %s\n", BuildTime)
		os.Exit(0)
	}

	// Handle --help flag
	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// TODO: Use debug, fast, thorough, and config flags in later phases
	// For now, store them for future use
	_ = debugMode
	_ = fastMode
	_ = thoroughMode
	_ = configFile

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

func printHelp() {
	fmt.Printf(`b+ (Be Positive) - Intelligent, model-agnostic, privacy-first agentic terminal coding assistant

Usage:
  bplus [flags]

Core Flags:
  -h, --help              Show this help message
  -v, --version           Show version information
      --debug             Enable debug mode with verbose logging

Execution Modes:
      --fast              Run in Fast Mode (Layer 4 only) - default
      --thorough          Run in Thorough Mode (all 7 layers active)

Configuration:
      --config <path>     Path to config file (default: ~/.config/bplus/config.yaml)

Examples:
  bplus                   # Start in Fast Mode with default settings
  bplus --thorough        # Start in Thorough Mode for complex tasks
  bplus --debug           # Start with debug logging enabled
  bplus --version         # Show version information

For more information, visit: https://github.com/abrksh22/bplus
`)
}
