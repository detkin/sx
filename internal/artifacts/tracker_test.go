package artifacts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetTrackerPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")

	tests := []struct {
		name       string
		targetBase string
		wantInPath string // what should be in the path
	}{
		{
			name:       "global installation",
			targetBase: claudeDir,
			wantInPath: ".cache/skills/installed-state/global.json",
		},
		{
			name:       "repo-scoped installation",
			targetBase: "/home/user/myrepo/.claude",
			wantInPath: ".cache/skills/installed-state/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTrackerPath(tt.targetBase)

			// Verify it's in the cache directory
			if !strings.Contains(got, tt.wantInPath) {
				t.Errorf("GetTrackerPath(%q) = %q, want path containing %q", tt.targetBase, got, tt.wantInPath)
			}

			// Verify it's NOT in the targetBase directory anymore
			if strings.HasPrefix(got, tt.targetBase) {
				t.Errorf("GetTrackerPath(%q) = %q, should NOT be under targetBase (should be in cache)", tt.targetBase, got)
			}

			// Verify it ends with .json
			if !strings.HasSuffix(got, ".json") {
				t.Errorf("GetTrackerPath(%q) = %q, want path ending with .json", tt.targetBase, got)
			}
		})
	}
}

func TestTrackerPathGlobalVsRepoScoped(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	globalPath := GetTrackerPath(claudeDir)
	repoPath := GetTrackerPath("/some/repo/.claude")

	// Global and repo-scoped should have different paths
	if globalPath == repoPath {
		t.Errorf("Global and repo-scoped tracker paths should be different, both are: %s", globalPath)
	}

	// Global should contain "global"
	if !strings.Contains(globalPath, "global") {
		t.Errorf("Global tracker path should contain 'global', got: %s", globalPath)
	}

	// Repo-scoped should NOT contain "global"
	if strings.Contains(repoPath, "global.json") {
		t.Errorf("Repo-scoped tracker path should not contain 'global.json', got: %s", repoPath)
	}
}
