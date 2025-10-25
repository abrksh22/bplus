package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid fast mode config",
			config: &Config{
				Mode: "fast",
				Models: ModelConfig{
					Default: "anthropic/claude-sonnet-4-5",
				},
				Layers: LayerConfig{
					MainAgent: MainAgentLayerConfig{
						Enabled: true,
					},
					ContextManagement: ContextLayerConfig{
						Enabled: true,
					},
					Validation: ValidationLayerConfig{
						MaxIterations: 3,
					},
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			wantErr: false,
		},
		{
			name: "valid thorough mode config",
			config: &Config{
				Mode: "thorough",
				Models: ModelConfig{
					Default: "anthropic/claude-sonnet-4-5",
				},
				Layers: LayerConfig{
					IntentClarification: IntentLayerConfig{
						Enabled: true,
					},
					ParallelPlanning: PlanningLayerConfig{
						Enabled:  true,
						NumPlans: 4,
						Models: []string{
							"anthropic/claude-opus-4-1",
							"openai/gpt-4-turbo",
							"gemini/gemini-2-5-pro",
							"ollama/deepseek-coder:33b",
						},
					},
					Synthesis: SynthesisLayerConfig{
						Enabled: true,
					},
					MainAgent: MainAgentLayerConfig{
						Enabled: true,
					},
					Validation: ValidationLayerConfig{
						Enabled:       true,
						MaxIterations: 3,
					},
					ContextManagement: ContextLayerConfig{
						Enabled: true,
					},
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid mode",
			config: &Config{
				Mode: "invalid",
				Models: ModelConfig{
					Default: "anthropic/claude-sonnet-4-5",
				},
				Layers: LayerConfig{
					MainAgent: MainAgentLayerConfig{
						Enabled: true,
					},
					ContextManagement: ContextLayerConfig{
						Enabled: true,
					},
					Validation: ValidationLayerConfig{
						MaxIterations: 3,
					},
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			wantErr: true,
			errMsg:  "invalid mode",
		},
		{
			name: "missing default model",
			config: &Config{
				Mode: "fast",
				Models: ModelConfig{
					Default: "",
				},
				Layers: LayerConfig{
					MainAgent: MainAgentLayerConfig{
						Enabled: true,
					},
					ContextManagement: ContextLayerConfig{
						Enabled: true,
					},
					Validation: ValidationLayerConfig{
						MaxIterations: 3,
					},
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			wantErr: true,
			errMsg:  "default model must be specified",
		},
		{
			name: "main agent disabled",
			config: &Config{
				Mode: "fast",
				Models: ModelConfig{
					Default: "anthropic/claude-sonnet-4-5",
				},
				Layers: LayerConfig{
					MainAgent: MainAgentLayerConfig{
						Enabled: false,
					},
					ContextManagement: ContextLayerConfig{
						Enabled: true,
					},
					Validation: ValidationLayerConfig{
						MaxIterations: 3,
					},
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			wantErr: true,
			errMsg:  "main agent layer (Layer 4) cannot be disabled",
		},
		{
			name: "context management disabled",
			config: &Config{
				Mode: "fast",
				Models: ModelConfig{
					Default: "anthropic/claude-sonnet-4-5",
				},
				Layers: LayerConfig{
					MainAgent: MainAgentLayerConfig{
						Enabled: true,
					},
					ContextManagement: ContextLayerConfig{
						Enabled: false,
					},
					Validation: ValidationLayerConfig{
						MaxIterations: 3,
					},
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			wantErr: true,
			errMsg:  "context management layer (Layer 6) cannot be disabled",
		},
		{
			name: "invalid num_plans",
			config: &Config{
				Mode: "thorough",
				Models: ModelConfig{
					Default: "anthropic/claude-sonnet-4-5",
				},
				Layers: LayerConfig{
					ParallelPlanning: PlanningLayerConfig{
						Enabled:  true,
						NumPlans: 10, // Invalid: must be 2-8
						Models:   []string{"model1", "model2"},
					},
					MainAgent: MainAgentLayerConfig{
						Enabled: true,
					},
					ContextManagement: ContextLayerConfig{
						Enabled: true,
					},
					Validation: ValidationLayerConfig{
						MaxIterations: 3,
					},
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			wantErr: true,
			errMsg:  "num_plans must be between 2 and 8",
		},
		{
			name: "mismatched planning models",
			config: &Config{
				Mode: "thorough",
				Models: ModelConfig{
					Default: "anthropic/claude-sonnet-4-5",
				},
				Layers: LayerConfig{
					ParallelPlanning: PlanningLayerConfig{
						Enabled:  true,
						NumPlans: 4,
						Models:   []string{"model1", "model2"}, // Only 2 models for 4 plans
					},
					MainAgent: MainAgentLayerConfig{
						Enabled: true,
					},
					ContextManagement: ContextLayerConfig{
						Enabled: true,
					},
					Validation: ValidationLayerConfig{
						MaxIterations: 3,
					},
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			wantErr: true,
			errMsg:  "number of planning models",
		},
		{
			name: "invalid validation iterations",
			config: &Config{
				Mode: "fast",
				Models: ModelConfig{
					Default: "anthropic/claude-sonnet-4-5",
				},
				Layers: LayerConfig{
					MainAgent: MainAgentLayerConfig{
						Enabled: true,
					},
					ContextManagement: ContextLayerConfig{
						Enabled: true,
					},
					Validation: ValidationLayerConfig{
						MaxIterations: 10, // Invalid: must be 1-5
					},
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			wantErr: true,
			errMsg:  "validation max_iterations must be between 1 and 5",
		},
		{
			name: "invalid logging level",
			config: &Config{
				Mode: "fast",
				Models: ModelConfig{
					Default: "anthropic/claude-sonnet-4-5",
				},
				Layers: LayerConfig{
					MainAgent: MainAgentLayerConfig{
						Enabled: true,
					},
					ContextManagement: ContextLayerConfig{
						Enabled: true,
					},
					Validation: ValidationLayerConfig{
						MaxIterations: 3,
					},
				},
				Logging: LoggingConfig{
					Level: "invalid",
				},
			},
			wantErr: true,
			errMsg:  "invalid logging level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetConfigDir(t *testing.T) {
	// Save original env vars
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	tests := []struct {
		name         string
		xdgConfig    string
		wantContains string
	}{
		{
			name:         "with XDG_CONFIG_HOME",
			xdgConfig:    "/custom/config",
			wantContains: "/custom/config/bplus",
		},
		{
			name:         "without XDG_CONFIG_HOME",
			xdgConfig:    "",
			wantContains: ".config/bplus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("XDG_CONFIG_HOME", tt.xdgConfig)

			dir, err := GetConfigDir()
			require.NoError(t, err)
			assert.Contains(t, dir, tt.wantContains)
		})
	}
}

func TestGetDataDir(t *testing.T) {
	// Save original env vars
	originalXDG := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", originalXDG)

	tests := []struct {
		name         string
		xdgData      string
		wantContains string
	}{
		{
			name:         "with XDG_DATA_HOME",
			xdgData:      "/custom/data",
			wantContains: "/custom/data/bplus",
		},
		{
			name:         "without XDG_DATA_HOME",
			xdgData:      "",
			wantContains: ".local/share/bplus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("XDG_DATA_HOME", tt.xdgData)

			dir, err := GetDataDir()
			require.NoError(t, err)
			assert.Contains(t, dir, tt.wantContains)
		})
	}
}

func TestGetCacheDir(t *testing.T) {
	// Save original env vars
	originalXDG := os.Getenv("XDG_CACHE_HOME")
	defer os.Setenv("XDG_CACHE_HOME", originalXDG)

	tests := []struct {
		name         string
		xdgCache     string
		wantContains string
	}{
		{
			name:         "with XDG_CACHE_HOME",
			xdgCache:     "/custom/cache",
			wantContains: "/custom/cache/bplus",
		},
		{
			name:         "without XDG_CACHE_HOME",
			xdgCache:     "",
			wantContains: ".cache/bplus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("XDG_CACHE_HOME", tt.xdgCache)

			dir, err := GetCacheDir()
			require.NoError(t, err)
			assert.Contains(t, dir, tt.wantContains)
		})
	}
}

func TestLoader_Load(t *testing.T) {
	// Create a temporary directory for test configs
	tmpDir := t.TempDir()

	// Create a test config file
	testConfig := `
mode: fast
models:
  default: "anthropic/claude-sonnet-4-5"
providers:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    base_url: "https://api.anthropic.com"
    timeout: 300s
    max_retries: 3
layers:
  main_agent:
    enabled: true
    model: "anthropic/claude-sonnet-4-5"
  context_management:
    enabled: true
    model: "openai/gpt-4-turbo"
    max_context_tokens: 200000
  validation:
    max_iterations: 3
logging:
  level: info
`

	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	require.NoError(t, err)

	// Set environment variable for substitution
	os.Setenv("ANTHROPIC_API_KEY", "test-api-key-123")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	// Load configuration
	loader := NewLoader()
	config, err := loader.Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify loaded config
	assert.Equal(t, "fast", config.Mode)
	assert.Equal(t, "anthropic/claude-sonnet-4-5", config.Models.Default)
	assert.Equal(t, "test-api-key-123", config.Providers["anthropic"].APIKey)
	assert.Equal(t, "https://api.anthropic.com", config.Providers["anthropic"].BaseURL)
	assert.Equal(t, 300*time.Second, config.Providers["anthropic"].Timeout)
	assert.Equal(t, 3, config.Providers["anthropic"].MaxRetries)
	assert.True(t, config.Layers.MainAgent.Enabled)
	assert.True(t, config.Layers.ContextManagement.Enabled)
	assert.Equal(t, 3, config.Layers.Validation.MaxIterations)
	assert.Equal(t, "info", config.Logging.Level)
}

func TestLoader_LoadWithDefaults(t *testing.T) {
	// Load with non-existent config file to test defaults
	loader := NewLoader()
	config, err := loader.Load("/nonexistent/config.yaml")

	// Should succeed and use defaults
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify defaults
	assert.Equal(t, "fast", config.Mode)
	assert.Equal(t, "anthropic/claude-sonnet-4-5", config.Models.Default)
	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, 4, config.Layers.ParallelPlanning.NumPlans)
	assert.True(t, config.Layers.MainAgent.Enabled)
	assert.True(t, config.Layers.ContextManagement.Enabled)
}

func TestProviderConfig_Timeout(t *testing.T) {
	provider := ProviderConfig{
		Timeout: 300 * time.Second,
	}

	assert.Equal(t, 300*time.Second, provider.Timeout)
	assert.Equal(t, 5*time.Minute, provider.Timeout)
}
