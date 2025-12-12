package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSessionCache_HasSession(t *testing.T) {
	// Create temp cache dir
	tmpDir := t.TempDir()
	t.Setenv("SKILLS_CACHE_DIR", tmpDir)

	cache, err := NewSessionCache("test-client")
	if err != nil {
		t.Fatalf("Failed to create session cache: %v", err)
	}

	// Test empty cache
	if cache.HasSession("session-1") {
		t.Error("Expected HasSession to return false for empty cache")
	}

	// Record a session
	if err := cache.RecordSession("session-1"); err != nil {
		t.Fatalf("Failed to record session: %v", err)
	}

	// Test that session is now found
	if !cache.HasSession("session-1") {
		t.Error("Expected HasSession to return true after recording")
	}

	// Test that other sessions are not found
	if cache.HasSession("session-2") {
		t.Error("Expected HasSession to return false for unrecorded session")
	}
}

func TestSessionCache_RecordSession(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("SKILLS_CACHE_DIR", tmpDir)

	cache, err := NewSessionCache("test-client")
	if err != nil {
		t.Fatalf("Failed to create session cache: %v", err)
	}

	// Record multiple sessions
	sessions := []string{"session-1", "session-2", "session-3"}
	for _, s := range sessions {
		if err := cache.RecordSession(s); err != nil {
			t.Fatalf("Failed to record session %s: %v", s, err)
		}
	}

	// Verify all are recorded
	for _, s := range sessions {
		if !cache.HasSession(s) {
			t.Errorf("Expected session %s to be recorded", s)
		}
	}

	// Verify file format
	data, err := os.ReadFile(cache.FilePath())
	if err != nil {
		t.Fatalf("Failed to read session file: %v", err)
	}

	// File should have 3 lines
	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	if lines != 3 {
		t.Errorf("Expected 3 lines in session file, got %d", lines)
	}
}

func TestSessionCache_EmptySessionID(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("SKILLS_CACHE_DIR", tmpDir)

	cache, err := NewSessionCache("test-client")
	if err != nil {
		t.Fatalf("Failed to create session cache: %v", err)
	}

	// Empty session ID should not be recorded
	if err := cache.RecordSession(""); err != nil {
		t.Fatalf("RecordSession with empty ID should not error: %v", err)
	}

	// File should not exist
	if _, err := os.Stat(cache.FilePath()); !os.IsNotExist(err) {
		t.Error("Expected no file to be created for empty session ID")
	}

	// HasSession should return false for empty
	if cache.HasSession("") {
		t.Error("Expected HasSession to return false for empty session ID")
	}
}

func TestSessionCache_CullOldEntries(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("SKILLS_CACHE_DIR", tmpDir)

	cache, err := NewSessionCache("test-client")
	if err != nil {
		t.Fatalf("Failed to create session cache: %v", err)
	}

	// Write some entries with old timestamps manually
	oldTime := time.Now().Add(-10 * 24 * time.Hour).UTC().Format(time.RFC3339)
	newTime := time.Now().UTC().Format(time.RFC3339)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(cache.FilePath()), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	content := "old-session " + oldTime + "\n" +
		"new-session " + newTime + "\n"

	if err := os.WriteFile(cache.FilePath(), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	// Cull entries older than 5 days
	if err := cache.CullOldEntries(5 * 24 * time.Hour); err != nil {
		t.Fatalf("Failed to cull old entries: %v", err)
	}

	// Old session should be gone
	if cache.HasSession("old-session") {
		t.Error("Expected old session to be culled")
	}

	// New session should still exist
	if !cache.HasSession("new-session") {
		t.Error("Expected new session to still exist after culling")
	}
}

func TestSessionCache_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("SKILLS_CACHE_DIR", tmpDir)

	cache, err := NewSessionCache("test-client")
	if err != nil {
		t.Fatalf("Failed to create session cache: %v", err)
	}

	// Record a session
	if err := cache.RecordSession("session-1"); err != nil {
		t.Fatalf("Failed to record session: %v", err)
	}

	// Clear cache
	if err := cache.Clear(); err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	// Session should no longer exist
	if cache.HasSession("session-1") {
		t.Error("Expected session to be cleared")
	}
}

func TestSessionCache_MultipleClients(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("SKILLS_CACHE_DIR", tmpDir)

	cache1, err := NewSessionCache("client-1")
	if err != nil {
		t.Fatalf("Failed to create cache 1: %v", err)
	}

	cache2, err := NewSessionCache("client-2")
	if err != nil {
		t.Fatalf("Failed to create cache 2: %v", err)
	}

	// Record to client 1
	if err := cache1.RecordSession("session-a"); err != nil {
		t.Fatalf("Failed to record to cache 1: %v", err)
	}

	// Client 1 should have it
	if !cache1.HasSession("session-a") {
		t.Error("Client 1 should have session-a")
	}

	// Client 2 should NOT have it
	if cache2.HasSession("session-a") {
		t.Error("Client 2 should not have session-a")
	}

	// Verify different files
	if cache1.FilePath() == cache2.FilePath() {
		t.Error("Different clients should have different cache files")
	}
}
