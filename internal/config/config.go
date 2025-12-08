package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sleuth-io/skills/internal/utils"
)

// RepositoryType represents the type of repository (sleuth or git)
type RepositoryType string

const (
	RepositoryTypeSleuth RepositoryType = "sleuth"
	RepositoryTypeGit    RepositoryType = "git"
)

// Config represents the configuration for the skills CLI
type Config struct {
	// Type of repository: "sleuth" or "git"
	Type RepositoryType `json:"type"`

	// ServerURL is the Sleuth server URL (only for type=sleuth)
	ServerURL string `json:"serverUrl,omitempty"`

	// AuthToken is the OAuth token for Sleuth server (only for type=sleuth)
	AuthToken string `json:"authToken,omitempty"`

	// RepositoryURL is the Git repository URL (only for type=git)
	RepositoryURL string `json:"repositoryUrl,omitempty"`
}

// getLegacyConfigFile returns the old config file path for backwards compatibility
func getLegacyConfigFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".claude", "plugins", "skills", "config.json"), nil
}

// Load loads the configuration from the config file
// Falls back to the old location (~/.claude/plugins/skills/config.json) for backwards compatibility
func Load() (*Config, error) {
	configFile, err := utils.GetConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get config file path: %w", err)
	}

	// Try new location first
	if utils.FileExists(configFile) {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}

		return &cfg, nil
	}

	// Fallback to legacy location
	legacyConfigFile, err := getLegacyConfigFile()
	if err == nil && utils.FileExists(legacyConfigFile) {
		data, err := os.ReadFile(legacyConfigFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read legacy config file: %w", err)
		}

		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse legacy config file: %w", err)
		}

		return &cfg, nil
	}

	return nil, fmt.Errorf("configuration not found. Run 'skills init' first")
}

// Save saves the configuration to the config file
func Save(cfg *Config) error {
	configFile, err := utils.GetConfigFile()
	if err != nil {
		return fmt.Errorf("failed to get config file path: %w", err)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configFile)
	if err := utils.EnsureDir(configDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON with indentation
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Exists checks if a configuration file exists
func Exists() bool {
	configFile, err := utils.GetConfigFile()
	if err != nil {
		return false
	}
	return utils.FileExists(configFile)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Type != RepositoryTypeSleuth && c.Type != RepositoryTypeGit {
		return fmt.Errorf("invalid repository type: %s (must be 'sleuth' or 'git')", c.Type)
	}

	switch c.Type {
	case RepositoryTypeSleuth:
		if c.ServerURL == "" {
			return fmt.Errorf("serverUrl is required for sleuth repository type")
		}
		if c.AuthToken == "" {
			return fmt.Errorf("authToken is required for sleuth repository type")
		}
	case RepositoryTypeGit:
		if c.RepositoryURL == "" {
			return fmt.Errorf("repositoryUrl is required for git repository type")
		}
	}

	return nil
}

// GetServerURL returns the Sleuth server URL, with environment override
func (c *Config) GetServerURL() string {
	if envURL := os.Getenv("SLEUTH_SERVER_URL"); envURL != "" {
		return envURL
	}
	return c.ServerURL
}

// IsSilent checks if silent mode is enabled via environment variable
func IsSilent() bool {
	return os.Getenv("SKILLS_SYNC_SILENT") == "true"
}
