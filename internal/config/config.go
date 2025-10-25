package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the complete b+ configuration
type Config struct {
	// Core settings
	Mode        string            `mapstructure:"mode" yaml:"mode" json:"mode"`                      // "fast" or "thorough"
	Version     string            `mapstructure:"version" yaml:"version" json:"version"`             // Config file version
	Models      ModelConfig       `mapstructure:"models" yaml:"models" json:"models"`                // Model configuration
	Providers   ProviderConfigs   `mapstructure:"providers" yaml:"providers" json:"providers"`       // Provider configurations
	Layers      LayerConfig       `mapstructure:"layers" yaml:"layers" json:"layers"`                // Layer-specific settings
	Tools       ToolConfig        `mapstructure:"tools" yaml:"tools" json:"tools"`                   // Tool settings
	UI          UIConfig          `mapstructure:"ui" yaml:"ui" json:"ui"`                            // UI settings
	Session     SessionConfig     `mapstructure:"session" yaml:"session" json:"session"`             // Session management
	Security    SecurityConfig    `mapstructure:"security" yaml:"security" json:"security"`          // Security settings
	Cost        CostConfig        `mapstructure:"cost" yaml:"cost" json:"cost"`                      // Cost management
	Performance PerformanceConfig `mapstructure:"performance" yaml:"performance" json:"performance"` // Performance settings
	Logging     LoggingConfig     `mapstructure:"logging" yaml:"logging" json:"logging"`             // Logging configuration
}

// ModelConfig defines model selection for all layers
type ModelConfig struct {
	Default string            `mapstructure:"default" yaml:"default" json:"default"` // Default model for all layers
	Layers  map[string]string `mapstructure:"layers" yaml:"layers" json:"layers"`    // Per-layer model overrides
}

// ProviderConfigs contains all provider configurations
type ProviderConfigs map[string]ProviderConfig

// ProviderConfig defines configuration for a single provider
type ProviderConfig struct {
	APIKey     string            `mapstructure:"api_key" yaml:"api_key" json:"api_key"`
	BaseURL    string            `mapstructure:"base_url" yaml:"base_url" json:"base_url"`
	Timeout    time.Duration     `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	MaxRetries int               `mapstructure:"max_retries" yaml:"max_retries" json:"max_retries"`
	Extra      map[string]string `mapstructure:"extra" yaml:"extra" json:"extra"` // Provider-specific settings
}

// LayerConfig defines layer-specific settings
type LayerConfig struct {
	IntentClarification IntentLayerConfig     `mapstructure:"intent_clarification" yaml:"intent_clarification" json:"intent_clarification"`
	ParallelPlanning    PlanningLayerConfig   `mapstructure:"parallel_planning" yaml:"parallel_planning" json:"parallel_planning"`
	Synthesis           SynthesisLayerConfig  `mapstructure:"synthesis" yaml:"synthesis" json:"synthesis"`
	MainAgent           MainAgentLayerConfig  `mapstructure:"main_agent" yaml:"main_agent" json:"main_agent"`
	Validation          ValidationLayerConfig `mapstructure:"validation" yaml:"validation" json:"validation"`
	ContextManagement   ContextLayerConfig    `mapstructure:"context_management" yaml:"context_management" json:"context_management"`
}

// IntentLayerConfig for Layer 1
type IntentLayerConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Model    string `mapstructure:"model" yaml:"model" json:"model"`
	MaxTurns int    `mapstructure:"max_turns" yaml:"max_turns" json:"max_turns"`
}

// PlanningLayerConfig for Layer 2
type PlanningLayerConfig struct {
	Enabled  bool     `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	NumPlans int      `mapstructure:"num_plans" yaml:"num_plans" json:"num_plans"`
	Models   []string `mapstructure:"models" yaml:"models" json:"models"`
}

// SynthesisLayerConfig for Layer 3
type SynthesisLayerConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Model   string `mapstructure:"model" yaml:"model" json:"model"`
}

// MainAgentLayerConfig for Layer 4
type MainAgentLayerConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"` // Always true, but kept for consistency
	Model   string `mapstructure:"model" yaml:"model" json:"model"`
}

// ValidationLayerConfig for Layer 5
type ValidationLayerConfig struct {
	Enabled       bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Model         string `mapstructure:"model" yaml:"model" json:"model"`
	MaxIterations int    `mapstructure:"max_iterations" yaml:"max_iterations" json:"max_iterations"`
	StrictMode    bool   `mapstructure:"strict_mode" yaml:"strict_mode" json:"strict_mode"`
}

// ContextLayerConfig for Layer 6
type ContextLayerConfig struct {
	Enabled          bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"` // Always true
	Model            string `mapstructure:"model" yaml:"model" json:"model"`
	MaxContextTokens int    `mapstructure:"max_context_tokens" yaml:"max_context_tokens" json:"max_context_tokens"`
}

// ToolConfig defines tool settings
type ToolConfig struct {
	EnabledTools  []string                   `mapstructure:"enabled_tools" yaml:"enabled_tools" json:"enabled_tools"`
	DisabledTools []string                   `mapstructure:"disabled_tools" yaml:"disabled_tools" json:"disabled_tools"`
	AutoApprove   []string                   `mapstructure:"auto_approve" yaml:"auto_approve" json:"auto_approve"`
	MCPServers    map[string]MCPServerConfig `mapstructure:"mcp_servers" yaml:"mcp_servers" json:"mcp_servers"`
}

// MCPServerConfig defines MCP server configuration
type MCPServerConfig struct {
	Command   string            `mapstructure:"command" yaml:"command" json:"command"`
	Args      []string          `mapstructure:"args" yaml:"args" json:"args"`
	Env       map[string]string `mapstructure:"env" yaml:"env" json:"env"`
	Transport string            `mapstructure:"transport" yaml:"transport" json:"transport"` // "stdio", "http", "sse"
	URL       string            `mapstructure:"url" yaml:"url" json:"url"`                   // For http/sse transports
}

// UIConfig defines UI settings
type UIConfig struct {
	Theme      string `mapstructure:"theme" yaml:"theme" json:"theme"`
	NoColor    bool   `mapstructure:"no_color" yaml:"no_color" json:"no_color"`
	Quiet      bool   `mapstructure:"quiet" yaml:"quiet" json:"quiet"`
	Verbose    bool   `mapstructure:"verbose" yaml:"verbose" json:"verbose"`
	ShowCost   bool   `mapstructure:"show_cost" yaml:"show_cost" json:"show_cost"`
	ShowTokens bool   `mapstructure:"show_tokens" yaml:"show_tokens" json:"show_tokens"`
	ShowLayers bool   `mapstructure:"show_layers" yaml:"show_layers" json:"show_layers"`
}

// SessionConfig defines session management settings
type SessionConfig struct {
	AutoSave           bool          `mapstructure:"auto_save" yaml:"auto_save" json:"auto_save"`
	SaveInterval       time.Duration `mapstructure:"save_interval" yaml:"save_interval" json:"save_interval"`
	CheckpointEnabled  bool          `mapstructure:"checkpoint_enabled" yaml:"checkpoint_enabled" json:"checkpoint_enabled"`
	CheckpointInterval time.Duration `mapstructure:"checkpoint_interval" yaml:"checkpoint_interval" json:"checkpoint_interval"`
	MaxHistorySize     int           `mapstructure:"max_history_size" yaml:"max_history_size" json:"max_history_size"`
}

// SecurityConfig defines security settings
type SecurityConfig struct {
	Sandbox            bool     `mapstructure:"sandbox" yaml:"sandbox" json:"sandbox"`
	AutoApproveRead    bool     `mapstructure:"auto_approve_read" yaml:"auto_approve_read" json:"auto_approve_read"`
	AutoApproveWrite   bool     `mapstructure:"auto_approve_write" yaml:"auto_approve_write" json:"auto_approve_write"`
	AutoApproveExec    bool     `mapstructure:"auto_approve_exec" yaml:"auto_approve_exec" json:"auto_approve_exec"`
	AutoApproveNetwork bool     `mapstructure:"auto_approve_network" yaml:"auto_approve_network" json:"auto_approve_network"`
	IgnorePatterns     []string `mapstructure:"ignore_patterns" yaml:"ignore_patterns" json:"ignore_patterns"`
}

// CostConfig defines cost management settings
type CostConfig struct {
	BudgetEnabled  bool    `mapstructure:"budget_enabled" yaml:"budget_enabled" json:"budget_enabled"`
	SessionBudget  float64 `mapstructure:"session_budget" yaml:"session_budget" json:"session_budget"`    // In USD
	DailyBudget    float64 `mapstructure:"daily_budget" yaml:"daily_budget" json:"daily_budget"`          // In USD
	MonthlyBudget  float64 `mapstructure:"monthly_budget" yaml:"monthly_budget" json:"monthly_budget"`    // In USD
	AlertThreshold float64 `mapstructure:"alert_threshold" yaml:"alert_threshold" json:"alert_threshold"` // Percentage (0-100)
	FreeOnly       bool    `mapstructure:"free_only" yaml:"free_only" json:"free_only"`                   // Use only free models
}

// PerformanceConfig defines performance settings
type PerformanceConfig struct {
	MaxParallel    int           `mapstructure:"max_parallel" yaml:"max_parallel" json:"max_parallel"`
	CacheEnabled   bool          `mapstructure:"cache_enabled" yaml:"cache_enabled" json:"cache_enabled"`
	DefaultTimeout time.Duration `mapstructure:"default_timeout" yaml:"default_timeout" json:"default_timeout"`
	MaxContextSize int           `mapstructure:"max_context_size" yaml:"max_context_size" json:"max_context_size"`
}

// LoggingConfig defines logging settings
type LoggingConfig struct {
	Level      string `mapstructure:"level" yaml:"level" json:"level"`          // "debug", "info", "warn", "error"
	Format     string `mapstructure:"format" yaml:"format" json:"format"`       // "text", "json"
	File       string `mapstructure:"file" yaml:"file" json:"file"`             // Log file path
	MaxSize    int    `mapstructure:"max_size" yaml:"max_size" json:"max_size"` // Max size in MB
	MaxBackups int    `mapstructure:"max_backups" yaml:"max_backups" json:"max_backups"`
	MaxAge     int    `mapstructure:"max_age" yaml:"max_age" json:"max_age"` // Max age in days
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate mode
	if c.Mode != "fast" && c.Mode != "thorough" {
		return fmt.Errorf("invalid mode: %s (must be 'fast' or 'thorough')", c.Mode)
	}

	// Validate default model
	if c.Models.Default == "" {
		return fmt.Errorf("default model must be specified")
	}

	// Validate layer configuration
	if !c.Layers.MainAgent.Enabled {
		return fmt.Errorf("main agent layer (Layer 4) cannot be disabled")
	}

	if !c.Layers.ContextManagement.Enabled {
		return fmt.Errorf("context management layer (Layer 6) cannot be disabled")
	}

	// Validate parallel planning configuration
	if c.Layers.ParallelPlanning.Enabled {
		if c.Layers.ParallelPlanning.NumPlans < 2 || c.Layers.ParallelPlanning.NumPlans > 8 {
			return fmt.Errorf("num_plans must be between 2 and 8")
		}
		if len(c.Layers.ParallelPlanning.Models) != c.Layers.ParallelPlanning.NumPlans {
			return fmt.Errorf("number of planning models (%d) must match num_plans (%d)",
				len(c.Layers.ParallelPlanning.Models), c.Layers.ParallelPlanning.NumPlans)
		}
	}

	// Validate validation configuration
	if c.Layers.Validation.MaxIterations < 1 || c.Layers.Validation.MaxIterations > 5 {
		return fmt.Errorf("validation max_iterations must be between 1 and 5")
	}

	// Validate logging level
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid logging level: %s (must be debug, info, warn, or error)", c.Logging.Level)
	}

	return nil
}

// GetConfigDir returns the configuration directory based on XDG spec
func GetConfigDir() (string, error) {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "bplus"), nil
	}

	// Fall back to ~/.config/bplus
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".config", "bplus"), nil
}

// GetDataDir returns the data directory based on XDG spec
func GetDataDir() (string, error) {
	// Check XDG_DATA_HOME first
	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		return filepath.Join(xdgData, "bplus"), nil
	}

	// Fall back to ~/.local/share/bplus
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".local", "share", "bplus"), nil
}

// GetCacheDir returns the cache directory based on XDG spec
func GetCacheDir() (string, error) {
	// Check XDG_CACHE_HOME first
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		return filepath.Join(xdgCache, "bplus"), nil
	}

	// Fall back to ~/.cache/bplus
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".cache", "bplus"), nil
}
