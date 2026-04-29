# Vern Agents Guide

This document provides context for AI agents working on the Vern project.

## Project Overview

Vern (Version Number Manager) is a programming language version installation manager written in Go. It follows the pattern of tools like `asdf`, `pyenv`, `rbenv` but with a focus on simplicity and sensible defaults.

## Architecture

### Core Packages

- **cmd/vern/** - CLI commands using Cobra
  - `root.go` - Main entry point, version variable
  - `install.go` - `vern install` command
  - `list.go` - `vern list` command  
  - `remove.go` - `vern remove` command (multi-select)
  - `default.go` - `vern default` command
  - `update.go` - `vern update` command
  - `setup.go` - `vern setup` command (shims)

- **internal/config/** - Configuration management
  - `config.go` - Language definitions, config loading, XDG paths
  - Supports versioned language lists (separate from binary version)

- **internal/install/** - Installation logic
  - `install.go` - Download, extract, compile from source
  - Supports `tar.gz`, `tar.xz` (via xz command)
  - Source compilation like pyenv/rbenv (`./configure && make && make install`)

- **internal/version/** - Version resolution
  - `version.go` - Version parsing, comparison, fetching available versions from URLs

### Key Concepts

- **Languages YAML** (`~/.config/vern/languages.yaml`) - Defines supported languages, download URLs, version regex, build commands
- **Installs Directory** (`~/.local/share/vern/installs/`) - Where language versions are installed
- **Shims Directory** (`~/.local/share/vern/shims/`) - Executable scripts that resolve versions via `.vern` files or defaults
- **Defaults** (`~/.local/share/vern/defaults.yaml`) - Global default versions per language
- **`.vern` files** - Project-level version files (format: `language version`)

## Common Tasks

### Adding a New Language

1. Add entry to `internal/config/config.go` `createDefaultConfig()` 
2. Add entry to `languages/vX.X.X.yaml` for the versioned language list
3. Update `languages/manifest.json` with new version
4. Test installation: `vern install <lang> <version>`

Language config structure:
```yaml
- name: <lang>              # Language identifier
  binary_name: <bin>         # Executable name (e.g., python3, node, ruby)
  version_source:
    url: <url>               # URL to scrape for versions
    version_regex: <regex>       # Regex to extract versions (capture group 1)
  install:
    download_template: <url>     # URL template with {{.Version}}, {{.MajorMinor}}
    extract_type: tar.gz|tar.xz
    bin_rel_path: <path>         # Path to binary relative to install dir
    build_config: <cmd>          # Optional: ./configure --prefix={{.InstallDir}}
    build_command: <cmd>         # Optional: make -j$(nproc) && make install
```

### Version Resolution

The `version.ResolveVersion()` function handles:
- Empty string → latest version
- Single number (e.g., "3") → latest in that major version
- Two numbers (e.g., "3.11") → latest in that minor version  
- Full version (e.g., "3.11.2") → exact version

### Source Compilation

Languages like Python and Ruby compile from source (like pyenv/rbenv):
1. Download source tarball
2. Extract to temp directory
3. Run `build_config` (e.g., `./configure --prefix=...`)
4. Run `build_command` (e.g., `make && make install`)
5. Move built files to install directory

Pre-compiled languages (Go, Node.js) skip the build steps.

## Build & Release

### Building
```bash
go build -o vern cmd/vern/main.go
```

### Release Process
1. Bump version in `cmd/vern/cmd/root.go` (`var Version = "X.Y.Z"`)
2. Bump language list version in `languages/manifest.json`
3. Commit: `git commit -m "Release vX.Y.Z"`
4. Tag: `git tag vX.Y.Z`
5. Push: `git push origin main --tags`
6. GitHub Action automatically builds binaries for:
   - linux/amd64, linux/arm64
   - darwin/amd64, darwin/arm64

### Installer
The `scripts/install.sh` script:
- Detects OS/arch
- Downloads appropriate binary from GitHub Releases
- Installs to `~/.local/bin/vern`
- Downloads language list
- Optionally adds `~/.local/bin` to PATH

Run with: `curl -fsSL https://raw.githubusercontent.com/chris-roerig/vern/main/scripts/install.sh | bash`

## Known Issues

### Ruby Installation
- **Problem**: Ruby compilation requires a host Ruby (`baseruby`) to build
- **Workaround**: Install system Ruby first (`apt install ruby`), then compile
- **Alternative**: Use pre-compiled Ruby binaries from a different source
- **Status**: Python source compilation works; Ruby needs work

### Python Shared Libraries
- **Problem**: Pre-compiled Python from `python-build-standalone` may have library path issues
- **Solution**: Compile from source (now implemented)
- **Status**: Working with source compilation

## Development Tips

### Testing Locally
```bash
# Build and install locally
go build -o ~/.local/bin/vern cmd/vern/main.go

# Test installation
~/.local/bin/vern install go
~/.local/bin/vern install python 3.11
~/.local/bin/vern list

# Test shims
~/.local/bin/vern setup
export PATH="/home/chris/.local/share/vern/shims:$PATH"
echo "python 3.11.10" > /tmp/.vern
cd /tmp && python3 --version
```

### Debugging Extraction
The `extractTarGz()` and `extractTarXz()` functions strip top-level directories from tarballs. If files aren't where expected, check:
1. What the tarball structure is: `tar -tzf <file> | head`
2. What `BinRelPath` is set to in config
3. That the extraction is stripping prefixes correctly

### PATH Issues
If shims aren't being used:
1. Check `echo $PATH` includes shims directory
2. Run `hash -r` to clear shell cache
3. Source shell config: `source ~/.zshrc` or `source ~/.bashrc`
4. Verify with `which <lang>` - should show shims path

## Project Structure
```
vern/
├── cmd/vern/           # CLI commands
│   ├── main.go       # Entry point
│   └── cmd/            # Command implementations
├── internal/          # Internal packages
│   ├── config/       # Configuration
│   ├── install/      # Installation logic
│   └── version/      # Version resolution
├── languages/          # Versioned language lists
│   ├── manifest.json  # Points to latest language list
│   └── vX.X.X.yaml    # Language definitions
├── scripts/            # Installer script
│   └── install.sh
├── .github/workflows/  # CI/CD
│   └── release.yml
├── README.md
├── AGENTS.md          # This file
└── ROADMAP.md         # Future plans
```
