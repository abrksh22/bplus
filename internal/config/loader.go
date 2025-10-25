package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Loader handles configuration loading and merging
type Loader struct {
	v *viper.Viper
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{
		v: viper.New(),
	}
}

// Load loads configuration from all sources and merges them
// Priority (highest to lowest):
// 1. CLI flags (handled externally by cobra)
// 2. Environment variables
// 3. Project config (.b+/config.yaml)
// 4. User config (~/.config/bplus/config.yaml)
// 5. Default values
func (l *Loader) Load(projectConfigPath string) (*Config, error) {
	// Set default values first
	l.setDefaults()

	// Configure Viper
	l.v.SetConfigType("yaml")
	l.v.SetEnvPrefix("BPLUS")
	l.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	l.v.AutomaticEnv()

	// Load user config (~/.config/bplus/config.yaml)
	if err := l.loadUserConfig(); err != nil {
		// User config is optional, so we just log and continue
		// In production, this would use proper logging
		fmt.Fprintf(os.Stderr, "Warning: could not load user config: %v\n", err)
	}

	// Load project config (.b+/config.yaml or specified path)
	if projectConfigPath != "" {
		// If path is explicitly specified, it should exist
		if _, err := os.Stat(projectConfigPath); os.IsNotExist(err) {
			// Project config is optional if not explicitly specified
			fmt.Fprintf(os.Stderr, "Info: project config not found: %s\n", projectConfigPath)
		} else if err := l.loadProjectConfig(projectConfigPath); err != nil {
			return nil, fmt.Errorf("failed to load project config: %w", err)
		}
	} else {
		// Try to load from .b+/config.yaml in current directory
		if err := l.loadProjectConfig(".b+/config.yaml"); err != nil {
			// Project config is optional
			fmt.Fprintf(os.Stderr, "Info: no project config found\n")
		}
	}

	// Parse environment variables
	l.bindEnvVars()

	// Unmarshal into Config struct
	var config Config
	if err := l.v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Perform environment variable substitution
	if err := l.substituteEnvVars(&config); err != nil {
		return nil, fmt.Errorf("failed to substitute environment variables: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// loadUserConfig loads configuration from user config directory
func (l *Loader) loadUserConfig() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// User config doesn't exist, which is fine
		return nil
	}

	l.v.AddConfigPath(configDir)
	l.v.SetConfigName("config")

	if err := l.v.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading user config: %w", err)
	}

	return nil
}

// loadProjectConfig loads configuration from project directory
func (l *Loader) loadProjectConfig(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("project config not found: %s", path)
	}

	l.v.SetConfigFile(path)
	if err := l.v.MergeInConfig(); err != nil {
		return fmt.Errorf("error reading project config: %w", err)
	}

	return nil
}

// bindEnvVars binds environment variables to config keys
func (l *Loader) bindEnvVars() {
	// Bind common environment variables
	envBindings := map[string]string{
		"ANTHROPIC_API_KEY":  "providers.anthropic.api_key",
		"OPENAI_API_KEY":     "providers.openai.api_key",
		"GOOGLE_API_KEY":     "providers.gemini.api_key",
		"OPENROUTER_API_KEY": "providers.openrouter.api_key",
		"BPLUS_MODE":         "mode",
		"BPLUS_MODEL":        "models.default",
		"BPLUS_LOG_LEVEL":    "logging.level",
	}

	for env, key := range envBindings {
		l.v.BindEnv(key, env)
	}
}

// substituteEnvVars performs ${VAR} substitution in string fields
func (l *Loader) substituteEnvVars(config *Config) error {
	// Substitute in provider API keys
	for name, provider := range config.Providers {
		provider.APIKey = os.ExpandEnv(provider.APIKey)
		config.Providers[name] = provider
	}

	// Substitute in MCP server environment variables
	for name, server := range config.Tools.MCPServers {
		for key, value := range server.Env {
			server.Env[key] = os.ExpandEnv(value)
		}
		config.Tools.MCPServers[name] = server
	}

	return nil
}

// setDefaults sets default configuration values
func (l *Loader) setDefaults() {
	// Core defaults
	l.v.SetDefault("mode", "fast")
	l.v.SetDefault("version", "1.0")

	// Model defaults
	l.v.SetDefault("models.default", "anthropic/claude-sonnet-4-5")

	// Provider defaults
	l.v.SetDefault("providers.anthropic.base_url", "https://api.anthropic.com")
	l.v.SetDefault("providers.anthropic.timeout", "300s")
	l.v.SetDefault("providers.anthropic.max_retries", 3)

	l.v.SetDefault("providers.openai.base_url", "https://api.openai.com/v1")
	l.v.SetDefault("providers.openai.timeout", "300s")
	l.v.SetDefault("providers.openai.max_retries", 3)

	l.v.SetDefault("providers.gemini.base_url", "https://generativelanguage.googleapis.com/v1beta")
	l.v.SetDefault("providers.gemini.timeout", "300s")
	l.v.SetDefault("providers.gemini.max_retries", 3)

	l.v.SetDefault("providers.ollama.base_url", "http://localhost:11434")
	l.v.SetDefault("providers.ollama.timeout", "300s")
	l.v.SetDefault("providers.ollama.max_retries", 3)

	l.v.SetDefault("providers.lmstudio.base_url", "http://localhost:1234")
	l.v.SetDefault("providers.lmstudio.timeout", "300s")
	l.v.SetDefault("providers.lmstudio.max_retries", 3)

	// Layer defaults
	l.v.SetDefault("layers.intent_clarification.enabled", true)
	l.v.SetDefault("layers.intent_clarification.model", "openai/gpt-4-turbo")
	l.v.SetDefault("layers.intent_clarification.max_turns", 5)

	l.v.SetDefault("layers.parallel_planning.enabled", true)
	l.v.SetDefault("layers.parallel_planning.num_plans", 4)
	l.v.SetDefault("layers.parallel_planning.models", []string{
		"anthropic/claude-opus-4-1",
		"anthropic/claude-sonnet-4-5",
		"gemini/gemini-2-5-pro",
		"ollama/deepseek-coder:33b",
	})

	l.v.SetDefault("layers.synthesis.enabled", true)
	l.v.SetDefault("layers.synthesis.model", "anthropic/claude-opus-4-1")

	l.v.SetDefault("layers.main_agent.enabled", true)
	l.v.SetDefault("layers.main_agent.model", "anthropic/claude-sonnet-4-5")

	l.v.SetDefault("layers.validation.enabled", true)
	l.v.SetDefault("layers.validation.model", "openai/gpt-4-turbo")
	l.v.SetDefault("layers.validation.max_iterations", 3)
	l.v.SetDefault("layers.validation.strict_mode", false)

	l.v.SetDefault("layers.context_management.enabled", true)
	l.v.SetDefault("layers.context_management.model", "openai/gpt-4-turbo")
	l.v.SetDefault("layers.context_management.max_context_tokens", 200000)

	// Tool defaults
	l.v.SetDefault("tools.enabled_tools", []string{}) // Empty means all enabled
	l.v.SetDefault("tools.disabled_tools", []string{})
	l.v.SetDefault("tools.auto_approve", []string{})

	// UI defaults
	l.v.SetDefault("ui.theme", "dark")
	l.v.SetDefault("ui.no_color", false)
	l.v.SetDefault("ui.quiet", false)
	l.v.SetDefault("ui.verbose", false)
	l.v.SetDefault("ui.show_cost", true)
	l.v.SetDefault("ui.show_tokens", true)
	l.v.SetDefault("ui.show_layers", true)

	// Session defaults
	l.v.SetDefault("session.auto_save", true)
	l.v.SetDefault("session.save_interval", "5m")
	l.v.SetDefault("session.checkpoint_enabled", false)
	l.v.SetDefault("session.checkpoint_interval", "5m")
	l.v.SetDefault("session.max_history_size", 1000)

	// Security defaults
	l.v.SetDefault("security.sandbox", false)
	l.v.SetDefault("security.auto_approve_read", false)
	l.v.SetDefault("security.auto_approve_write", false)
	l.v.SetDefault("security.auto_approve_exec", false)
	l.v.SetDefault("security.auto_approve_network", false)
	l.v.SetDefault("security.ignore_patterns", []string{
		"node_modules/**",
		".git/**",
		"**/*.log",
		"**/*.tmp",
	})

	// Cost defaults
	l.v.SetDefault("cost.budget_enabled", false)
	l.v.SetDefault("cost.session_budget", 0.0)
	l.v.SetDefault("cost.daily_budget", 0.0)
	l.v.SetDefault("cost.monthly_budget", 0.0)
	l.v.SetDefault("cost.alert_threshold", 80.0)
	l.v.SetDefault("cost.free_only", false)

	// Performance defaults
	l.v.SetDefault("performance.max_parallel", 4)
	l.v.SetDefault("performance.cache_enabled", true)
	l.v.SetDefault("performance.default_timeout", "5m")
	l.v.SetDefault("performance.max_context_size", 200000)

	// Logging defaults
	l.v.SetDefault("logging.level", "info")
	l.v.SetDefault("logging.format", "text")
	l.v.SetDefault("logging.file", "")
	l.v.SetDefault("logging.max_size", 100) // 100 MB
	l.v.SetDefault("logging.max_backups", 3)
	l.v.SetDefault("logging.max_age", 28) // 28 days
}

// SaveConfig saves the current configuration to a file
func SaveConfig(config *Config, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create viper instance for writing
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path)

	// Marshal config to viper
	// Note: We'd use a YAML library here for proper serialization
	// For now, this is a placeholder that would need yaml.Marshal

	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
