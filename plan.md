# Skills CLI Tool - Go Implementation Plan

## Overview

Create a new standalone Go CLI tool called `skills` that provisions AI artifacts (skills, agents, MCPs, etc.) from remote Sleuth servers or Git repositories. This is a greenfield project using the JavaScript pulse-client implementation as reference for specifications and behavior.

## Requirements Summary

### Commands to Implement
1. **init** - Initialize configuration (authenticate with Sleuth server OR configure Git repo as source)
   - Interactive mode by default
   - Support non-interactive mode with flags (`--type`, `--server-url`, `--repo-url`)
2. **install** (default command) - Read lock file, fetch artifacts, install locally
   - Runs when no subcommand specified (if lock file exists)
3. **add** - Take local zip file, install to repository and update lock file
   - Auto-detect metadata from zip contents
   - Always prompt user to confirm/edit detected values

### Key Constraints
- Ignore Claude Code plugin integration (no plugin wrapper mode)
- Direct installation to `~/.claude/` for global artifacts
- Clear boundaries between code areas
- RepositoryType interface abstraction for different source types
- Git operations via shell (not go-git library) to reuse user's git config
- Full dependency resolution with topological sort
- Platform-specific cache directories (macOS, Linux, Windows)

### User Clarifications
- **Add command**: Extract/detect metadata from zip, always ask user to confirm values (even if detected)
- **Plugin mode**: Direct installation only (no wrapper)
- **Git implementation**: Shell out to git CLI for auth compatibility
- **Dependencies**: Full resolution from the start
- **Binary name**: `skills`
- **Logging**: Standard Go CLI logging (simple by default, structured available)
- **Progress**: Use progress bars for downloads
- **Version embedding**: Use ldflags at build time
- **Release**: Use goreleaser with GitHub Actions
- **Init command**: Support both interactive and non-interactive (flags) modes
- **Default command**: `skills` alone runs install if lock file exists
- **Git commits**: Simple format: "Add {artifact-name} {version}"
- **Project location**: /home/mrdon/dev/skills/ (current directory)
- **Source layout**: Standard Go layout with cmd/ and internal/

---

## Project Structure

```
/home/mrdon/dev/skills/
  cmd/
    skills/
      main.go                    # Entry point, CLI setup
  internal/
    commands/
      init.go                    # Init command implementation
      install.go                 # Install command implementation
      add.go                     # Add command implementation
    config/
      config.go                  # Configuration management
      auth.go                    # OAuth device code flow
    repository/
      repository.go              # Repository interface (unified concept)
      sleuth.go                  # Sleuth HTTP server implementation
      git.go                     # Git repository implementation
      http.go                    # HTTP source (for individual artifact URLs)
      path.go                    # Local path source
    lockfile/
      lockfile.go               # Lock file parsing and management
      parser.go                 # TOML parser
      validation.go             # Lock file validation
    artifacts/
      artifact.go               # Artifact structures
      installer.go              # Installation orchestration
      dependency.go             # Dependency resolution
    handlers/
      handler.go                # Handler interface
      skill.go                  # Skill handler
      agent.go                  # Agent handler
      command.go                # Command handler
      hook.go                   # Hook handler
      mcp.go                    # MCP handler
      mcp_remote.go             # MCP remote handler
    cache/
      cache.go                  # Cache management
    scope/
      scope.go                  # Scope resolution (global/repo/path)
    metadata/
      metadata.go               # Metadata TOML parsing
      validation.go             # Metadata validation
    gitutil/
      gitutil.go                # Git context detection utilities
    utils/
      paths.go                  # Path utilities
      zip.go                    # Zip extraction utilities
      hash.go                   # Hash computation
  Makefile                      # Build, test, release targets
  .goreleaser.yml               # GoReleaser configuration
  .github/
    workflows/
      release.yml               # GitHub Actions release workflow
  go.mod
  go.sum
  README.md
  LICENSE
```

---

## Core Interfaces

### Repository Interface (Unified Concept)

```go
package repository

// Repository represents a source of artifacts with read and write capabilities
// This interface unifies the concepts of "repository" and "source fetcher"
type Repository interface {
    // Authenticate performs authentication with the repository
    // Returns an auth token or empty string if no auth needed
    Authenticate(ctx context.Context) (string, error)

    // GetLockFile retrieves the lock file from the repository
    // Returns lock file content and ETag for caching
    GetLockFile(ctx context.Context, cachedETag string) (content []byte, etag string, notModified bool, err error)

    // GetArtifact downloads an artifact using its source configuration from the lock file
    // The sourceConfig is the artifact's source table (e.g., source-http, source-git, source-path)
    GetArtifact(ctx context.Context, artifact *Artifact, sourceConfig map[string]interface{}) ([]byte, error)

    // AddArtifact uploads an artifact to the repository
    // Updates the lock file with the new artifact entry
    AddArtifact(ctx context.Context, artifact *Artifact, zipData []byte) error

    // GetVersionList retrieves available versions for an artifact (for resolution)
    // Only applicable to repositories with version management (Sleuth, not Git)
    GetVersionList(ctx context.Context, name string) ([]string, error)

    // GetMetadata retrieves metadata for a specific artifact version
    // Only applicable to repositories with version management (Sleuth, not Git)
    GetMetadata(ctx context.Context, name, version string) (*Metadata, error)

    // VerifyIntegrity checks hashes and sizes for downloaded artifacts
    VerifyIntegrity(data []byte, hashes map[string]string, size int64) error
}

// Note: Individual source types (HTTP URL, Git ref, Path) are handled internally
// by the Repository implementation when processing the sourceConfig from lock file artifacts
```

### ArtifactHandler Interface

```go
package handlers

type ArtifactHandler interface {
    // Install extracts and installs the artifact
    Install(ctx context.Context, zipData []byte, targetBase string) error

    // Remove uninstalls the artifact
    Remove(ctx context.Context, targetBase string) error

    // GetInstallPath returns the installation path relative to targetBase
    GetInstallPath() string

    // Validate checks if the zip structure is valid for this artifact type
    Validate(zipEntries []string) error
}
```

---

## Implementation Phases

### Phase 1: Core Infrastructure

**Files to create:**
- `cmd/pulse/main.go` - CLI framework using cobra/cli library
- `internal/config/config.go` - Configuration management (JSON storage)
- `internal/cache/cache.go` - Platform-specific cache directories
- `internal/utils/paths.go` - Path resolution utilities
- `internal/utils/zip.go` - Zip extraction using archive/zip
- `internal/utils/hash.go` - SHA256/SHA512 computation

**Key details:**
- Use `os.UserHomeDir()` for home directory
- Use `os.UserCacheDir()` with fallbacks for cache directories
- Config stored at `~/.claude/plugins/sleuth-sync/config.json`
- Cache stored at platform-specific locations (same as JS implementation)

### Phase 2: Lock File and Metadata

**Files to create:**
- `internal/lockfile/lockfile.go` - Lock file structures
- `internal/lockfile/parser.go` - TOML parsing using BurntSushi/toml
- `internal/lockfile/validation.go` - Lock file validation
- `internal/metadata/metadata.go` - Metadata structures and parsing
- `internal/metadata/validation.go` - Type-specific metadata validation

**Key details:**
- Parse TOML using github.com/BurntSushi/toml
- Support both inline tables and separate sections for sources
- Validate semantic versions using github.com/Masterminds/semver
- Extract metadata.toml from zips without full extraction

### Phase 3: Repository Implementation (Unified Concept)

**Files to create:**
- `internal/repository/repository.go` - Interface definition
- `internal/repository/sleuth.go` - Sleuth HTTP server implementation
- `internal/repository/git.go` - Git repository implementation
- `internal/repository/http.go` - HTTP source handler (for source-http in lock file)
- `internal/repository/path.go` - Path source handler (for source-path in lock file)
- `internal/config/auth.go` - OAuth device code flow

**Key details:**

**Sleuth Repository:**
- Implements Repository interface for Sleuth server
- HTTP client with Bearer token authentication
- ETag-based caching for lock file
- Endpoints: `/api/oauth/device-authorization/`, `/api/oauth/token/`, `/api/skills/lock`
- Poll interval: 5 seconds for OAuth flow
- Use `open` package or `xdg-open`/`open`/`start` commands to open browser
- GetArtifact: Dispatches to http.go, git.go, or path.go based on source table in artifact
- AddArtifact: Uploads to Sleuth server, server manages lock file updates

**Git Repository:**
- Implements Repository interface for Git repos
- Shell out to git CLI using `os/exec.Command`
- Clone to `{CACHE_DIR}/git-repos/{urlHash}/`
- Lock file at `sleuth.lock` in repo root
- GetLockFile: git pull, then read sleuth.lock
- GetArtifact: Dispatches to http.go, git.go, or path.go based on source table in artifact
- AddArtifact: Add artifact to repo structure, update sleuth.lock, commit, and push
- Use user's existing git credentials/SSH keys
- Hash URL using SHA256, take first 16 chars for directory name

**HTTP Source Handler (http.go):**
- Handles artifacts with `source-http` table in lock file
- Use net/http with context for downloads
- Integrity verification (SHA256/SHA512)
- Used by both Sleuth and Git repositories

**Git Source Handler (git.go):**
- Handles artifacts with `source-git` table in lock file
- Clone/fetch/checkout using exec.Command
- Resolves branches/tags to commit SHAs
- Used by both Sleuth and Git repositories

**Path Source Handler (path.go):**
- Handles artifacts with `source-path` table in lock file
- Use os.ReadFile, support tilde expansion
- Used by both Sleuth and Git repositories

### Phase 4: Artifact Handlers

**Files to create:**
- `internal/handlers/handler.go` - Interface and factory
- `internal/handlers/skill.go` - Extract to `skills/{name}/`
- `internal/handlers/agent.go` - Extract to `agents/{name}/`
- `internal/handlers/command.go` - Extract to `commands/{name}.md`
- `internal/handlers/hook.go` - Extract to `hooks/{name}/`, update settings.json
- `internal/handlers/mcp.go` - Extract to `mcp-servers/{name}/`, update .mcp.json
- `internal/handlers/mcp_remote.go` - No extraction, update .mcp.json only

**Key details:**
- Use archive/zip for extraction
- Read config files from zip without full extraction
- Merge/update JSON config files (encoding/json)
- Convert relative paths to absolute in configs
- Tag resources with `_artifact` field for cleanup tracking

### Phase 5: Core Installation Logic

**Files to create:**
- `internal/artifacts/artifact.go` - Artifact structures
- `internal/artifacts/fetcher.go` - Download orchestration
- `internal/artifacts/installer.go` - Installation orchestration
- `internal/artifacts/dependency.go` - Topological sort
- `internal/scope/scope.go` - Scope resolution
- `internal/git/git.go` - Git context detection

**Key details:**
- Parallel downloads with goroutines and sync.WaitGroup
- Concurrent limit: 10 artifacts at a time
- Dependency resolution: DFS with cycle detection
- Scope matching: normalize URLs, support path wildcards
- Git detection: shell out to git commands

### Phase 6: Commands Implementation

**Files to create:**
- `internal/commands/init.go` - Init command
- `internal/commands/install.go` - Install command
- `internal/commands/add.go` - Add command

**Init command flow:**
1. Interactive mode (default):
   - Ask: Sleuth server or Git repository?
   - If Sleuth: Run OAuth device code flow, save token
   - If Git: Ask for repository URL, clone to cache, save config
   - Save config to `~/.claude/plugins/sleuth-sync/config.json`
2. Non-interactive mode (with flags):
   - `skills init --type=sleuth [--server-url=URL]`
   - `skills init --type=git --repo-url=URL`
   - Fail if required values missing

**Default command behavior:**
- Running `skills` with no subcommand:
  - If lock file exists: Run install command
  - If no lock file: Show help/error message

**Install command flow:**
1. Load config
2. Get repository instance (Sleuth or Git)
3. Fetch lock file (with ETag caching)
4. Validate lock file
5. Detect git context (repo URL, relative path)
6. Filter by client compatibility (claude-code)
7. Resolve applicable artifacts by scope
8. Resolve dependencies (topological sort)
9. Download artifacts in parallel (max 10 concurrent)
10. Install artifacts in dependency order
11. Cleanup removed artifacts (compare with cached previous lock file)
12. Cleanup old cached versions

**Add command flow:**
1. Prompt for zip file path (or accept as argument)
2. Extract and detect artifact type (look for SKILL.md, AGENT.md, etc.)
3. Try to extract metadata values (name, version, type)
4. Always prompt user to confirm/edit detected values (even if complete)
   - For skills/agents: Default version to "1.0.0" if not found
   - For MCPs: Prompt for command, args, env configuration
5. Create/update metadata.toml in artifact
6. Repackage zip with metadata.toml
7. Get repository instance
8. Call repository.AddArtifact() to upload
9. Update lock file locally
10. Commit with message: "Add {artifact-name} {version}"

---

## Dependencies (go.mod)

```go
module github.com/sleuth-io/skills

go 1.21

require (
    github.com/BurntSushi/toml v1.3.2                         // TOML parsing
    github.com/Masterminds/semver/v3 v3.2.1                   // Semantic versioning
    github.com/spf13/cobra v1.8.0                             // CLI framework
    github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // Open browser for OAuth
    github.com/schollz/progressbar/v3 v3.14.1                 // Progress bars for downloads
    github.com/rs/zerolog v1.31.0                             // Structured logging (optional)
)
```

---

## Key Design Decisions

### 1. Unified Repository Concept
- **Single interface** combines repository and source concepts
- Repository implementations (Sleuth, Git) handle lock file operations
- Source handlers (HTTP, Git, Path) handle individual artifact fetching
- Repository.GetArtifact() dispatches to appropriate source handler based on artifact's source table
- **Clear boundaries**: Repository = lock file management, Source handlers = artifact fetching
- Each repository handles its own authentication method

### 2. Git Integration
- **Shell out to git CLI** to leverage user's existing configuration
- Avoids credential management complexity
- Works with SSH keys, credential helpers, etc.
- Use `exec.CommandContext` for cancellation support

### 3. Error Handling
- **Collect errors** during installation, don't fail-fast
- Report all errors at the end
- Use `multierror` pattern for error collection
- Log progress to stderr, errors to stderr

### 4. Caching Strategy
- **Lock file**: ETag-based for HTTP, git pull for Git repos
- **Artifacts**: Cache by name/version, validate with zip magic bytes
- **Git repos**: Clone once, fetch updates incrementally
- **Platform-specific** cache directories

### 5. Concurrency
- **Parallel downloads**: Goroutines with WaitGroup, limit to 10 concurrent
- **Sequential installation**: Respect dependency order
- **Context support**: All network operations support cancellation

### 6. Configuration
- **JSON format** for config file (simple, standard library support)
- **Environment overrides**: SLEUTH_SERVER_URL, SLEUTH_CONFIG_DIR
- **Silent mode**: SKILLS_SYNC_SILENT env var

---

## Implementation Strategy

### Stage 1: Core CLI (Current Scope)
Implement init, install, add commands as standalone Go binary with:
- Cross-platform builds via goreleaser
- Progress bars for user feedback
- Simple logging by default
- Makefile for common tasks
- GitHub Actions for releases

### Stage 2: Future Enhancements (Out of Scope)
- Lock file generation from requirements file (sleuth.txt)
- Version resolution from list.txt
- Publish command for uploading artifacts
- Search command for finding artifacts
- Update command for upgrading artifacts
- Plugin integration with Claude Code

---

## Testing Approach

### Unit Tests
- Lock file parsing and validation
- Metadata extraction and validation
- Dependency resolution (topological sort)
- Scope resolution logic
- Path utilities (tilde expansion, normalization)

### Integration Tests
- OAuth device code flow (mock server)
- Git operations (test repository)
- Artifact installation (temp directories)
- Configuration management

### End-to-End Tests
- Full init → install workflow
- Add artifact workflow
- Cleanup and caching

---

## Reference Materials

### JavaScript Implementation (Reference Only - Read-Only)
The pulse-client directory contains the JavaScript implementation for reference:
- `/home/mrdon/dev/skills/pulse-client/skills-sync/src/sync-flow.js` - Main orchestration
- `/home/mrdon/dev/skills/pulse-client/skills-sync/src/config.js` - Config management
- `/home/mrdon/dev/skills/pulse-client/skills-sync/src/auth.js` - OAuth flow
- `/home/mrdon/dev/skills/pulse-client/skills-sync/src/lockfile.js` - Lock file handling
- `/home/mrdon/dev/skills/pulse-client/skills-sync/src/source-fetchers/` - Source implementations
- `/home/mrdon/dev/skills/pulse-client/skills-sync/src/artifact-handlers/` - Handler implementations
- `/home/mrdon/dev/skills/pulse-client/Makefile` - Build patterns to follow

### Specifications (Reference Only - Read-Only)
- `/home/mrdon/dev/skills/pulse-client/docs/lock-spec.md` - Lock file format
- `/home/mrdon/dev/skills/pulse-client/docs/metadata-spec.md` - Metadata format
- `/home/mrdon/dev/skills/pulse-client/docs/repository-spec.md` - Repository structure
- `/home/mrdon/dev/skills/pulse-client/docs/requirements-spec.md` - Requirements format

### New Project Location
All new Go code will be created at: `/home/mrdon/dev/skills/`

---

## Build & Release

### Makefile Targets
Based on pulse-client/Makefile pattern:
- `make help` - Show available targets (default)
- `make build` - Build the binary
- `make install` - Install binary to $GOPATH/bin
- `make test` - Run tests
- `make lint` - Run linters (golangci-lint)
- `make format` - Format code (gofmt)
- `make clean` - Clean build artifacts
- `make release` - Create release with goreleaser (for local testing)

### GoReleaser Configuration
- Build for: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64
- Archive formats: tar.gz for Unix, zip for Windows
- Checksums: SHA256
- Binary name: `skills`
- Version embedding via ldflags

### GitHub Actions Workflow
- Trigger on: Tag push (v*)
- Run tests
- Run goreleaser
- Create GitHub release with binaries and checksums

---

## Success Criteria

✅ Go binary successfully initializes with Sleuth server authentication
✅ Go binary successfully initializes with Git repository
✅ Install command reads lock file and installs artifacts correctly
✅ All artifact types install to correct locations (skills, agents, commands, hooks, MCPs)
✅ Scope resolution works (global, repo-specific, path-specific)
✅ Dependency resolution and topological sort works correctly
✅ Add command packages and uploads artifacts
✅ Git integration works with user's existing credentials
✅ Caching works (lock file ETag, artifact versions)
✅ Cross-compiles for major platforms (macOS, Linux, Windows)
✅ Cleanup removes old artifacts correctly
