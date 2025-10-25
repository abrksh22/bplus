package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/abrksh22/bplus/app"
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

	// Initialize application
	opts := &app.Options{
		Version:    Version,
		ConfigPath: *configFile,
		DebugMode:  *debugMode,
		FastMode:   *fastMode,
		Thorough:   *thoroughMode,
	}

	application, err := app.New(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize b+: %v\n", err)
		os.Exit(1)
	}
	defer application.Close()

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Create the UI model with application
	model := ui.NewWithApp(application)

	// Create the Bubble Tea program
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
		tea.WithContext(ctx),      // Use context for cancellation
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
