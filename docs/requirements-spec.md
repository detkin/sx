# Sleuth Requirements File Specification

## Overview

This specification defines `sleuth.txt`, a simple requirements format for declaring AI client artifacts before resolution into a lock file. Inspired by pip's `requirements.txt`, it prioritizes simplicity and human editability.

## File Naming

Requirements files must be named:

- `sleuth.txt` (default)
- `sleuth-<name>.txt` (named variants)

## Format

Plain text file with one requirement per line:

```txt
# Comments start with #
# Blank lines are ignored

# Registry artifacts with version specifiers
github-mcp==1.2.3
code-reviewer>=3.0.0
database-mcp~=2.0.0
helper-agent>=1.0,<2.0
awesome-skill

# Git sources
git+https://github.com/user/repo.git@main#name=artifact-name
git+https://github.com/user/repo.git@v1.2.3#name=artifact-name&path=subdir

# Local paths
./relative/path/artifact.zip
~/home/relative/artifact.zip
/absolute/path/artifact.zip

# HTTP sources
https://example.com/artifacts/skill.zip
```

## Requirement Types

### Registry Artifacts

Format: `<name>[<version-spec>]`

```txt
# Exact version
github-mcp==1.2.3

# Minimum version
code-reviewer>=3.0.0

# Compatible version (>= 2.0.0, < 2.1.0)
database-mcp~=2.0.0

# Range
helper-agent>=1.0,<2.0

# Latest version (no specifier)
awesome-skill
```

**Version Specifiers**:

- `==X.Y.Z` - Exact version
- `>=X.Y.Z` - Minimum version
- `>X.Y.Z` - Greater than
- `<=X.Y.Z` - Maximum version
- `<X.Y.Z` - Less than
- `~=X.Y.Z` - Compatible release (>= X.Y.Z, < X.(Y+1).0)
- `X.Y.Z` - Exact version (same as ==)
- Multiple specifiers separated by comma: `>=1.0,<2.0`

**Resolution**:

- Uses default repository configured in `config.toml` (see `repository-spec.md`)
- Queries repository for available versions
- Filters versions matching specifier
- Resolves dependencies recursively
- Selects highest compatible version
- Generates lock file entry with concrete source (HTTP or path)

### Git Sources

Format: `git+<url>@<ref>#name=<artifact-name>[&path=<subdirectory>]`

```txt
# Branch reference
git+https://github.com/user/repo.git@main#name=custom-agent

# Tag reference
git+https://github.com/user/repo.git@v1.2.3#name=my-mcp

# Commit SHA
git+https://github.com/user/repo.git@abc123def456#name=pinned-skill

# With subdirectory
git+https://github.com/user/monorepo.git@main#name=api-agent&path=packages/agents
```

**Components**:

- `git+` prefix (required)
- URL: Repository URL (HTTPS or SSH)
- `@<ref>`: Git reference - branch, tag, or commit SHA
- `#name=<name>`: Artifact name (required)
- `&path=<subdir>`: Subdirectory within repo (optional)

**Resolution**:

- Branch/tag: Resolved to commit SHA via `git ls-remote`
- Commit SHA: Used as-is
- Client clones repo and extracts artifact from `path` (or root)

### Local Paths

Format: `<path>`

```txt
# Relative to current directory
./skills/my-skill.zip

# Relative to home directory
~/dev/artifacts/my-agent.zip

# Absolute path
/var/artifacts/production-mcp.zip
```

**Resolution**:

- Used as-is (no version resolution)
- Must point to valid `.zip` file
- Version extracted from zip metadata or generated from file mtime

### HTTP Sources

Format: `<url>`

```txt
https://example.com/artifacts/skill-1.2.3.zip
https://cdn.company.com/mcps/custom.zip
```

**Resolution**:

- Downloaded directly from URL
- Version extracted from zip metadata or HTTP `Last-Modified` header
- No dependency resolution (unless metadata found in zip)

## Version Detection

When version is not specified in requirement (git, path, HTTP sources), it's determined by:

### From Artifact Metadata

Client extracts zip and checks for version in:

1. `package.json` - read `version` field
2. `metadata.yml` - read `version` field
3. `metadata.toml` - read `version` field

Example `package.json`:

```json
{
  "name": "my-skill",
  "version": "1.2.3"
}
```

Example `metadata.yml`:

```yaml
name: my-skill
version: 1.2.3
type: skill
```

### From Source Timestamps

If no metadata found, generate version: `0.0.0+YYYYMMDD`

**Local paths**: Use file system `mtime`

```txt
./my-skill.zip  →  0.0.0+20250125  (if file modified on 2025-01-25)
```

**HTTP sources**: Use `Last-Modified` header, fallback to `Date` header, fallback to current date

```txt
https://example.com/skill.zip  →  0.0.0+20250120  (if Last-Modified: 2025-01-20)
```

**Git sources**: Use commit timestamp

```txt
git+https://github.com/user/repo.git@main#name=agent  →  0.0.0+20250118  (if commit date is 2025-01-18)
```

## Comments and Whitespace

```txt
# Full-line comments start with #

github-mcp==1.2.3  # Inline comments not supported

# Blank lines are ignored


# Indentation is ignored
  code-reviewer>=3.0.0
```

## Dependencies

Requirements file specifies **top-level** artifacts only. Dependencies are:

- Declared in artifact metadata (for registry artifacts)
- Declared in `package.json` or `metadata.yml` (for git/path/HTTP artifacts)
- Resolved recursively during lock file generation

## Lock File Generation

Command: `sleuth lock`

Process:

1. Parse `sleuth.txt`
2. For each requirement:
   - Registry artifacts: Query repository (see `repository-spec.md`), select best match
   - Git: Resolve refs to commit SHAs, extract artifact
   - Path/HTTP: Download, extract version from metadata or timestamp
3. Resolve dependencies recursively
4. Detect conflicts (multiple artifacts require incompatible versions)
5. Generate `sleuth.lock` with:
   - Exact versions for all artifacts
   - Commit SHAs for git sources
   - Hashes for HTTP sources
   - Full dependency graph

See `lock-spec.md` for lock file format and `repository-spec.md` for repository structure.

## Examples

### Simple Project

`sleuth.txt`:

```txt
# Core MCPs
github-mcp>=1.2.0
database-mcp~=2.0.0

# Skills
code-reviewer>=3.0.0
```

Generates `sleuth.lock` with resolved versions, dependencies, and hashes.

### Mixed Sources

`sleuth.txt`:

```txt
# From registry
github-mcp==1.2.3

# From git (internal tool)
git+https://github.com/company/agents.git@main#name=api-helper&path=dist

# Local development
./local-skills/debug-skill.zip

# External URL
https://cdn.example.com/skills/formatter.zip
```

### Development Workflow

1. Create `sleuth.txt` with high-level requirements
2. Run `sleuth lock` to generate `sleuth.lock`
3. Commit both files to version control
4. Team members run `sleuth sync` to install from lock file
5. Update `sleuth.txt` when adding/changing artifacts
6. Run `sleuth lock` to regenerate lock file

## Comparison with Lock File

| Aspect           | Requirements (sleuth.txt) | Lock File (sleuth.lock)   |
| ---------------- | ------------------------- | ------------------------- |
| **Purpose**      | Declare desired artifacts | Pin exact versions        |
| **Format**       | Plain text, line-based    | TOML, structured          |
| **Versions**     | Ranges (>=, ~=, etc.)     | Exact versions only       |
| **Git refs**     | Branches, tags, commits   | Commit SHAs only          |
| **Dependencies** | Top-level only            | Full dependency graph     |
| **Hashes**       | Not included              | Required for HTTP sources |
| **Editability**  | Hand-edited by users      | Machine-generated         |
| **Scope**        | Implicit (file location)  | Explicit per artifact     |

## Edge Cases

### Conflicting Versions

If multiple artifacts require incompatible versions:

```txt
artifact-a  # depends on helper>=2.0
artifact-b  # depends on helper<2.0
```

Resolution fails with clear error message listing conflict.

### Missing Git Ref

```txt
git+https://github.com/user/repo.git@nonexistent#name=foo
```

Resolution fails: "Git ref 'nonexistent' not found in repository"

### Invalid Path

```txt
./nonexistent/artifact.zip
```

Resolution fails: "File not found: ./nonexistent/artifact.zip"

### Circular Dependencies

If artifact A depends on B, and B depends on A, resolution fails with circular dependency error.

## Future Enhancements

Potential additions:

- Environment markers: `github-mcp>=1.2.0 ; client=="claude-code"`
- Constraints files: `-c constraints.txt`
- Include other requirements: `-r base-requirements.txt`
- Editable installs: `-e ./local-dev-artifact`
