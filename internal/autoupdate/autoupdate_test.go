package autoupdate

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sleuth-io/sx/internal/buildinfo"
	"github.com/sleuth-io/sx/internal/cache"
)

func TestShouldCheckDevBuild(t *testing.T) {
	// Save original version
	originalVersion := buildinfo.Version
	defer func() { buildinfo.Version = originalVersion }()

	// Set version to "dev"
	buildinfo.Version = "dev"

	// Should return early for dev builds
	err := checkAndUpdate()
	if err != nil {
		t.Errorf("Expected no error for dev build, got: %v", err)
	}
}

func TestShouldCheckWithNoCache(t *testing.T) {
	// Clean up any existing cache
	cacheDir, err := cache.GetCacheDir()
	if err != nil {
		t.Fatalf("Failed to get cache dir: %v", err)
	}
	lastCheckFile := filepath.Join(cacheDir, updateCacheFile)
	_ = os.Remove(lastCheckFile)

	// Should check when there's no cache file
	if !shouldCheck() {
		t.Error("Expected shouldCheck to return true when cache file doesn't exist")
	}
}

func TestShouldCheckWithRecentCache(t *testing.T) {
	// Create a recent cache file
	cacheDir, err := cache.GetCacheDir()
	if err != nil {
		t.Fatalf("Failed to get cache dir: %v", err)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	lastCheckFile := filepath.Join(cacheDir, updateCacheFile)

	// Create cache file with current timestamp
	if err := updateCheckTimestamp(); err != nil {
		t.Fatalf("Failed to update timestamp: %v", err)
	}

	// Should not check when cache is recent
	if shouldCheck() {
		t.Error("Expected shouldCheck to return false when cache is recent")
	}

	// Clean up
	_ = os.Remove(lastCheckFile)
}

func TestShouldCheckWithOldCache(t *testing.T) {
	// Create an old cache file
	cacheDir, err := cache.GetCacheDir()
	if err != nil {
		t.Fatalf("Failed to get cache dir: %v", err)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	lastCheckFile := filepath.Join(cacheDir, updateCacheFile)

	// Create file and set old modification time
	f, err := os.Create(lastCheckFile)
	if err != nil {
		t.Fatalf("Failed to create cache file: %v", err)
	}
	f.Close()

	// Set modification time to 25 hours ago (past the 24 hour threshold)
	oldTime := time.Now().Add(-25 * time.Hour)
	if err := os.Chtimes(lastCheckFile, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set file time: %v", err)
	}

	// Should check when cache is old
	if !shouldCheck() {
		t.Error("Expected shouldCheck to return true when cache is old")
	}

	// Clean up
	_ = os.Remove(lastCheckFile)
}

func TestUpdateCheckTimestamp(t *testing.T) {
	// Create timestamp
	if err := updateCheckTimestamp(); err != nil {
		t.Fatalf("Failed to update timestamp: %v", err)
	}

	// Verify file exists
	cacheDir, err := cache.GetCacheDir()
	if err != nil {
		t.Fatalf("Failed to get cache dir: %v", err)
	}

	lastCheckFile := filepath.Join(cacheDir, updateCacheFile)
	if _, err := os.Stat(lastCheckFile); os.IsNotExist(err) {
		t.Error("Expected cache file to exist after updateCheckTimestamp")
	}

	// Clean up
	_ = os.Remove(lastCheckFile)
}
