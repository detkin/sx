package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sleuth-io/skills/internal/metadata"
	"github.com/sleuth-io/skills/internal/utils"
)

// AgentHandler handles agent artifact installation
type AgentHandler struct {
	metadata *metadata.Metadata
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(meta *metadata.Metadata) *AgentHandler {
	return &AgentHandler{
		metadata: meta,
	}
}

// Install extracts and installs the agent artifact
func (h *AgentHandler) Install(ctx context.Context, zipData []byte, targetBase string) error {
	// Validate zip structure
	if err := h.Validate(zipData); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Determine installation path
	installPath := filepath.Join(targetBase, h.GetInstallPath())

	// Remove existing installation if present
	if utils.IsDirectory(installPath) {
		if err := os.RemoveAll(installPath); err != nil {
			return fmt.Errorf("failed to remove existing installation: %w", err)
		}
	}

	// Create installation directory
	if err := utils.EnsureDir(installPath); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Extract zip to installation directory
	if err := utils.ExtractZip(zipData, installPath); err != nil {
		return fmt.Errorf("failed to extract zip: %w", err)
	}

	return nil
}

// Remove uninstalls the agent artifact
func (h *AgentHandler) Remove(ctx context.Context, targetBase string) error {
	installPath := filepath.Join(targetBase, h.GetInstallPath())

	if !utils.IsDirectory(installPath) {
		// Already removed or never installed
		return nil
	}

	if err := os.RemoveAll(installPath); err != nil {
		return fmt.Errorf("failed to remove agent: %w", err)
	}

	return nil
}

// GetInstallPath returns the installation path relative to targetBase
func (h *AgentHandler) GetInstallPath() string {
	return filepath.Join("agents", h.metadata.Artifact.Name)
}

// Validate checks if the zip structure is valid for an agent artifact
func (h *AgentHandler) Validate(zipData []byte) error {
	// List files in zip
	files, err := utils.ListZipFiles(zipData)
	if err != nil {
		return fmt.Errorf("failed to list zip files: %w", err)
	}

	// Check that metadata.toml exists
	if !containsFile(files, "metadata.toml") {
		return fmt.Errorf("metadata.toml not found in zip")
	}

	// Extract and validate metadata
	metadataBytes, err := utils.ReadZipFile(zipData, "metadata.toml")
	if err != nil {
		return fmt.Errorf("failed to read metadata.toml: %w", err)
	}

	meta, err := metadata.Parse(metadataBytes)
	if err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Validate metadata with file list
	if err := meta.ValidateWithFiles(files); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	// Verify artifact type matches
	if meta.Artifact.Type != "agent" {
		return fmt.Errorf("artifact type mismatch: expected agent, got %s", meta.Artifact.Type)
	}

	// Check that prompt file exists
	if meta.Agent == nil {
		return fmt.Errorf("[agent] section missing in metadata")
	}

	if !containsFile(files, meta.Agent.PromptFile) {
		return fmt.Errorf("prompt file not found in zip: %s", meta.Agent.PromptFile)
	}

	return nil
}
