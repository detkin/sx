package handlers

import (
	"context"
	"fmt"

	"github.com/sleuth-io/skills/internal/metadata"
)

// ArtifactHandler handles installation and removal of artifacts
type ArtifactHandler interface {
	// Install extracts and installs the artifact
	// targetBase is the base directory (e.g., ~/.claude/ or {repo}/.claude/)
	Install(ctx context.Context, zipData []byte, targetBase string) error

	// Remove uninstalls the artifact
	Remove(ctx context.Context, targetBase string) error

	// GetInstallPath returns the installation path relative to targetBase
	GetInstallPath() string

	// Validate checks if the zip structure is valid for this artifact type
	Validate(zipData []byte) error
}

// NewHandler creates an appropriate handler for the given artifact type
func NewHandler(meta *metadata.Metadata) (ArtifactHandler, error) {
	switch meta.Artifact.Type {
	case "skill":
		return NewSkillHandler(meta), nil
	case "agent":
		return NewAgentHandler(meta), nil
	case "command":
		return NewCommandHandler(meta), nil
	case "hook":
		return NewHookHandler(meta), nil
	case "mcp":
		return NewMCPHandler(meta), nil
	case "mcp-remote":
		return NewMCPRemoteHandler(meta), nil
	default:
		return nil, fmt.Errorf("unsupported artifact type: %s", meta.Artifact.Type)
	}
}
