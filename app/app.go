// Package app provides the main application structure and initialization.
package app

import (
	"context"
	"os"
	"path/filepath"

	"github.com/abrksh22/bplus/internal/config"
	"github.com/abrksh22/bplus/internal/errors"
	"github.com/abrksh22/bplus/internal/logging"
	"github.com/abrksh22/bplus/internal/storage"
	"github.com/abrksh22/bplus/layers/execution"
	"github.com/abrksh22/bplus/models"
	"github.com/abrksh22/bplus/models/providers/anthropic"
	"github.com/abrksh22/bplus/models/providers/gemini"
	"github.com/abrksh22/bplus/models/providers/lmstudio"
	"github.com/abrksh22/bplus/models/providers/ollama"
	"github.com/abrksh22/bplus/models/providers/openai"
	"github.com/abrksh22/bplus/models/providers/openrouter"
	"github.com/abrksh22/bplus/prompts"
	"github.com/abrksh22/bplus/security"
	"github.com/abrksh22/bplus/tools"
	"github.com/abrksh22/bplus/tools/exec"
	"github.com/abrksh22/bplus/tools/file"
)

// Application holds all the components needed to run b+.
type Application struct {
	Config         *config.Config
	Logger         *logging.Logger
	DB             *storage.SQLiteDB
	Provider       models.Provider
	ToolRegistry   *tools.Registry
	PermManager    *security.PermissionManager
	Agent          *execution.Agent
	SessionManager *execution.SessionManager
}

// New creates a new Application with all components initialized.
func New(opts *Options) (*Application, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Initialize logging
	logger := logging.NewDefaultLogger()

	logger.Info("Initializing b+ application", "version", opts.Version)

	// Load configuration
	cfg, err := loadConfig(opts)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeConfigInvalid, "failed to load configuration")
	}

	// Initialize database
	dbPath := getDBPath(cfg)
	db, err := storage.NewSQLiteDB(dbPath)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to initialize database")
	}

	logger.Info("Database initialized", "path", dbPath)

	// Initialize provider
	provider, err := createProvider(cfg)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeProvider, "failed to create provider")
	}

	logger.Info("Provider initialized", "provider", provider.Name())

	// Initialize tool registry
	toolReg := tools.NewRegistry()
	if err := registerTools(toolReg); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to register tools")
	}

	logger.Info("Tools registered", "count", len(toolReg.List()))

	// Initialize permission manager
	// For Phase 6 MVP, use a simple prompt handler
	promptHandler := func(ctx context.Context, req *security.PermissionRequest) (bool, error) {
		// TODO: Implement proper prompting in Phase 7
		// For now, auto-approve in interactive mode
		return true, nil
	}
	permManager := security.NewPermissionManager(security.ModeInteractive, promptHandler)

	// Create agent configuration
	agentConfig := &execution.AgentConfig{
		ModelName:     cfg.Models.Default,
		SystemPrompt:  prompts.GetLayer4Prompt(),
		MaxIterations: 10,
		Temperature:   0.7,
		MaxTokens:     4096,
		Streaming:     true,
	}

	// Create agent
	agent, err := execution.NewAgent(provider, agentConfig, toolReg, permManager)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to create agent")
	}

	logger.Info("Agent initialized")

	// Create session manager
	sessionManager := execution.NewSessionManager(db)

	return &Application{
		Config:         cfg,
		Logger:         logger,
		DB:             db,
		Provider:       provider,
		ToolRegistry:   toolReg,
		PermManager:    permManager,
		Agent:          agent,
		SessionManager: sessionManager,
	}, nil
}

// Close closes all resources.
func (app *Application) Close() error {
	if app.DB != nil {
		if err := app.DB.Close(); err != nil {
			return errors.Wrap(err, errors.ErrCodeInternal, "failed to close database")
		}
	}
	return nil
}

// Options holds application initialization options.
type Options struct {
	Version    string
	ConfigPath string
	DebugMode  bool
	FastMode   bool
	Thorough   bool
}

// DefaultOptions returns default options.
func DefaultOptions() *Options {
	return &Options{
		Version:   "dev",
		FastMode:  true,
		Thorough:  false,
		DebugMode: false,
	}
}

// loadConfig loads configuration from file or defaults.
func loadConfig(opts *Options) (*config.Config, error) {
	// For Phase 6 MVP, use sensible defaults
	cfg := &config.Config{
		Mode: "fast",
		Models: config.ModelConfig{
			Default: "anthropic/claude-sonnet-4-5",
		},
		Providers: config.ProviderConfigs{
			"anthropic": config.ProviderConfig{
				APIKey:  os.Getenv("ANTHROPIC_API_KEY"),
				BaseURL: "https://api.anthropic.com",
			},
			"openai": config.ProviderConfig{
				APIKey:  os.Getenv("OPENAI_API_KEY"),
				BaseURL: "https://api.openai.com/v1",
			},
			"gemini": config.ProviderConfig{
				APIKey:  os.Getenv("GEMINI_API_KEY"),
				BaseURL: "https://generativelanguage.googleapis.com/v1beta",
			},
			"openrouter": config.ProviderConfig{
				APIKey:  os.Getenv("OPENROUTER_API_KEY"),
				BaseURL: "https://openrouter.ai/api/v1",
			},
			"ollama": config.ProviderConfig{
				BaseURL: "http://localhost:11434",
			},
			"lmstudio": config.ProviderConfig{
				BaseURL: "http://localhost:1234/v1",
			},
		},
	}

	// Load from file if specified (TODO: implement config.LoadConfig in Phase 7)
	// if opts.ConfigPath != "" {
	// 	loaded, err := config.LoadConfig(opts.ConfigPath)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	cfg = loaded
	// }

	// Override with mode flags
	if opts.Thorough {
		cfg.Mode = "thorough"
	}

	return cfg, nil
}

// getDBPath returns the database path from config or default.
func getDBPath(cfg *config.Config) string {
	// Default to ~/.local/share/bplus/bplus.db
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./bplus.db"
	}

	dataDir := filepath.Join(homeDir, ".local", "share", "bplus")
	return filepath.Join(dataDir, "bplus.db")
}

// createProvider creates the appropriate provider based on configuration.
func createProvider(cfg *config.Config) (models.Provider, error) {
	// Parse model name to extract provider
	providerName, _, err := models.ParseModelName(cfg.Models.Default)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeConfigInvalid, "invalid model name")
	}

	// Get provider config
	providerCfg, ok := cfg.Providers[providerName]
	if !ok {
		return nil, errors.Newf(errors.ErrCodeToolNotFound, "provider %s not configured", providerName)
	}

	// Create provider based on name
	switch providerName {
	case "anthropic":
		if providerCfg.APIKey == "" {
			return nil, errors.New(errors.ErrCodeConfigInvalid, "ANTHROPIC_API_KEY not set")
		}
		var opts []anthropic.Option
		if providerCfg.BaseURL != "" {
			opts = append(opts, anthropic.WithBaseURL(providerCfg.BaseURL))
		}
		return anthropic.New(providerCfg.APIKey, opts...), nil

	case "openai":
		if providerCfg.APIKey == "" {
			return nil, errors.New(errors.ErrCodeConfigInvalid, "OPENAI_API_KEY not set")
		}
		var opts []openai.Option
		if providerCfg.BaseURL != "" {
			opts = append(opts, openai.WithBaseURL(providerCfg.BaseURL))
		}
		return openai.New(providerCfg.APIKey, opts...), nil

	case "gemini":
		if providerCfg.APIKey == "" {
			return nil, errors.New(errors.ErrCodeConfigInvalid, "GEMINI_API_KEY not set")
		}
		var opts []gemini.Option
		if providerCfg.BaseURL != "" {
			opts = append(opts, gemini.WithBaseURL(providerCfg.BaseURL))
		}
		return gemini.New(providerCfg.APIKey, opts...), nil

	case "openrouter":
		if providerCfg.APIKey == "" {
			return nil, errors.New(errors.ErrCodeConfigInvalid, "OPENROUTER_API_KEY not set")
		}
		var opts []openrouter.Option
		if providerCfg.BaseURL != "" {
			opts = append(opts, openrouter.WithBaseURL(providerCfg.BaseURL))
		}
		return openrouter.New(providerCfg.APIKey, opts...), nil

	case "ollama":
		var opts []ollama.Option
		baseURL := providerCfg.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		opts = append(opts, ollama.WithBaseURL(baseURL))
		return ollama.New(opts...), nil

	case "lmstudio":
		var opts []lmstudio.Option
		baseURL := providerCfg.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:1234/v1"
		}
		opts = append(opts, lmstudio.WithBaseURL(baseURL))
		return lmstudio.New(opts...), nil

	default:
		return nil, errors.Newf(errors.ErrCodeToolNotFound, "unsupported provider: %s", providerName)
	}
}

// registerTools registers all available tools.
func registerTools(registry *tools.Registry) error {
	// File tools
	if err := registry.Register(file.NewReadTool()); err != nil {
		return err
	}
	if err := registry.Register(file.NewWriteTool()); err != nil {
		return err
	}
	if err := registry.Register(file.NewEditTool()); err != nil {
		return err
	}
	if err := registry.Register(file.NewGlobTool()); err != nil {
		return err
	}
	if err := registry.Register(file.NewGrepTool()); err != nil {
		return err
	}

	// Exec tools
	if err := registry.Register(exec.NewBashTool()); err != nil {
		return err
	}

	return nil
}

// Execute runs the agent with the given request.
func (app *Application) Execute(ctx context.Context, req *execution.AgentRequest) (*execution.AgentResponse, error) {
	return app.Agent.Execute(ctx, req)
}
