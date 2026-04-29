# Vern Agents Guide

This document provides context for AI agents working on the Vern project.

## Project Overview

Vern (Version Number Manager) is a programming language version installation manager written in Go. It follows the pattern of tools like `asdf`, `pyenv`, `rbenv` but with a focus on simplicity and sensible defaults.

## Architecture

### Core Packages

- **cmd/vern/** - CLI commands using Cobra
  - `root.go` - Main entry point, version variable, ASCII logo display
  - `install.go` - `vern install` command (with `--verbose` flag)
  - `search.go` - `vern search` command (partial version filtering)
  - `list.go` - `vern list` command (semver sorted)
  - `which.go` - `vern which` command (resolved version for current dir)
  - `remove.go` - `vern remove` command (multi-select)
  - `default.go` - `vern default` command (promptui select)
  - `init.go` - `vern init` command (create .vern files)
  - `stats.go` - `vern stats` command (disk usage)
  - `update.go` - `vern update` command (self + languages)
  - `setup.go` - `vern setup` command (shims + version manager detection)
  - `implode.go` - `vern implode` command (selective uninstall)

- **internal/config/** - Configuration management
  - `config.go` - Language definitions, config loading
  - Uses explicit paths: `~/.config/vern` and `~/.local/share/vern`
  - Supports versioned language lists (separate from binary version)

- **internal/install/** - Installation logic
  - `install.go` - Download, extract, compile from source
  - Supports `tar.gz`, `tar.xz` (via xz command)
  - Source compilation (`./configure && make && make install`)
  - Rust install via `install.sh --prefix`
  - Download progress bar (MB/percentage)
  - Verbose mode for build output

- **internal/version/** - Version resolution
  - `version.go` - Version parsing, comparison, fetching available versions from URLs

- **internal/ui/** - Terminal output
  - `ui.go` - Dracula-themed colored output (Success, Warn, Error, Info, Dim, Accent)
  - ASCII logo constant

### Key Concepts

- **Languages YAML** (`~/.config/vern/languages.yaml`) - Defines supported languages, download URLs, version regex, build commands
- **Installs Directory** (`~/.local/share/vern/installs/`) - Where language versions are installed
- **Shims Directory** (`~/.local/share/vern/shims/`) - Executable scripts that resolve versions via `.vern` files or defaults
- **Defaults** (`~/.local/share/vern/defaults.yaml`) - Global default versions per language
- **`.vern` files** - Project-level version files (format: `language version`)

## Supported Languages

- Go - pre-compiled binary
- Python - compiled from source
- Node.js - pre-compiled binary
- Ruby - compiled from source
- Rust - installed via `install.sh --prefix`
- Zig - pre-compiled binary

## Template Variables

Download URL templates support these variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `{{.Version}}` | Full version | `1.21.0` |
| `{{.MajorMinor}}` | Major.Minor | `1.21` |
| `{{.OS}}` | Go's GOOS | `linux`, `darwin` |
| `{{.Arch}}` | Go's GOARCH | `amd64`, `arm64` |
| `{{.ArchAlt}}` | Node-style arch | `x64`, `arm64` |
| `{{.ArchGNU}}` | GNU-style arch | `x86_64`, `aarch64` |
| `{{.OSAlt}}` | Zig-style OS | `macos`, `linux` |
| `{{.RustTarget}}` | Rust target triple | `aarch64-apple-darwin` |
| `{{.InstallDir}}` | Install directory | (used in build commands) |

## Testing

Tests are in `tests/` directory:
- `version_test.go` - Version parsing and comparison
- `platform_test.go` - Arch/OS mapping functions
- `config_test.go` - Config and .vern file parsing
- `install_test.go` - Version sorting, installed languages

Run tests: `go test ./tests/ -v`

## CI/CD

- **CI** (`.github/workflows/ci.yml`) - Runs `go vet` and tests on every push to main/staging and PRs
- **Release** (`.github/workflows/release.yml`) - Tests must pass before building release binaries
- Release builds for: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64
- SHA256 checksums published alongside binaries

## Build & Release

### Building
```bash
go build -o vern cmd/vern/main.go
```

### Release Process
1. Bump version in `cmd/vern/cmd/root.go`
2. Bump language list version in `languages/manifest.json` (if languages changed)
3. Commit and push to main
4. Tag: `git tag vX.Y.Z && git push origin vX.Y.Z`
5. GitHub Action runs tests, then builds binaries

### Installer
```bash
curl -fsSL https://raw.githubusercontent.com/chris-roerig/vern/main/scripts/install.sh | bash
```

## Project Structure
```
vern/
├── cmd/vern/           # CLI entry point and commands
│   ├── main.go
│   └── cmd/            # Command implementations
├── internal/           # Internal packages
│   ├── config/         # Configuration
│   ├── install/        # Installation logic
│   ├── ui/             # Colored terminal output
│   └── version/        # Version resolution
├── tests/              # Test suite
├── languages/          # Versioned language lists
│   ├── manifest.json
│   └── v1.3.0.yaml
├── scripts/            # Installer script
├── docs/               # GitHub Pages site
├── .github/workflows/  # CI/CD
│   ├── ci.yml
│   └── release.yml
├── README.md
└── AGENTS.md
```
