package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sleuth-io/skills/internal/utils"
)

// GetCacheDir returns the platform-specific cache directory for skills
func GetCacheDir() (string, error) {
	// Check for environment override
	if cacheDir := os.Getenv("SKILLS_CACHE_DIR"); cacheDir != "" {
		return cacheDir, nil
	}

	// Use os.UserCacheDir() with platform-specific fallbacks
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		// Fallback to platform-specific defaults
		cacheDir, err = getFallbackCacheDir()
		if err != nil {
			return "", fmt.Errorf("failed to determine cache directory: %w", err)
		}
	}

	return filepath.Join(cacheDir, "sleuth-sync"), nil
}

// getFallbackCacheDir returns platform-specific fallback cache directories
func getFallbackCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Caches"), nil
	case "linux":
		xdgCache := os.Getenv("XDG_CACHE_HOME")
		if xdgCache != "" {
			return xdgCache, nil
		}
		return filepath.Join(homeDir, ".cache"), nil
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			return localAppData, nil
		}
		return filepath.Join(homeDir, "AppData", "Local"), nil
	default:
		return filepath.Join(homeDir, ".cache"), nil
	}
}

// GetArtifactCacheDir returns the directory for caching artifacts
func GetArtifactCacheDir() (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "artifacts"), nil
}

// GetGitReposCacheDir returns the directory for caching git repositories
func GetGitReposCacheDir() (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "git-repos"), nil
}

// GetLockFileCacheDir returns the directory for caching lock files
func GetLockFileCacheDir() (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "lockfiles"), nil
}

// EnsureCacheDirs creates all necessary cache directories
func EnsureCacheDirs() error {
	dirs := []func() (string, error){
		GetCacheDir,
		GetArtifactCacheDir,
		GetGitReposCacheDir,
		GetLockFileCacheDir,
	}

	for _, dirFunc := range dirs {
		dir, err := dirFunc()
		if err != nil {
			return err
		}
		if err := utils.EnsureDir(dir); err != nil {
			return fmt.Errorf("failed to create cache directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetArtifactCachePath returns the cache path for a specific artifact
func GetArtifactCachePath(name, version string) (string, error) {
	artifactCacheDir, err := GetArtifactCacheDir()
	if err != nil {
		return "", err
	}
	// Use sanitized name for directory safety
	safeName := filepath.Base(filepath.Clean(name))
	return filepath.Join(artifactCacheDir, safeName, version+".zip"), nil
}

// GetGitRepoCachePath returns the cache path for a git repository
func GetGitRepoCachePath(repoURL string) (string, error) {
	gitReposDir, err := GetGitReposCacheDir()
	if err != nil {
		return "", err
	}
	urlHash := utils.URLHash(repoURL)
	return filepath.Join(gitReposDir, urlHash), nil
}

// ClearArtifactCache removes cached artifacts for cleanup
func ClearArtifactCache() error {
	artifactCacheDir, err := GetArtifactCacheDir()
	if err != nil {
		return err
	}
	return os.RemoveAll(artifactCacheDir)
}
