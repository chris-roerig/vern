```
                    ▄▄         ▄▄   
▄▄ ▄▄ ▄▄▄▄ ▄▄▄▄     █   ▄▄  ▄▄  █   
██▄██ ██▄▄  ██▄█▄  ██   ███▄██  ██  
 ▀█▀  ██▄▄▄ ██ ██   █   ██ ▀██  █   
                    ▀▀         ▀▀   
```

**ver{n}** is the version manager that works like you think it should.

---

## Quick Start

Install vern with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/chris-roerig/vern/main/scripts/install.sh | bash
```

The installer will:

- Detect your OS and architecture
- Download the latest vern binary
- Install to `~/.local/bin` (or `/usr/local/bin` with sudo)
- Create config directories
- Download the supported languages list
- Optionally add `~/.local/bin` to your PATH

---

## Commands

```
vern                                   Print version and help
vern help                              Print detailed help

vern install <lang>                    Install latest version
vern install <lang> 3                  Install latest 3.x.x
vern install <lang> 3.11               Install latest 3.11.x
vern install <lang> 3.11.2             Install exact version
vern install <lang> --verbose          Show build output

vern search <lang>                     List available versions
vern search <lang> 3.11                List matching versions

vern list                              List all installed languages
vern list <lang>                       List installed versions for language

vern which <lang>                      Show resolved version for current directory

vern default <lang>                    Interactively select default
vern default <lang> <version>          Set default directly

vern remove <lang>                     Multi-select versions to remove

vern init                              Create empty .vern file
vern init <lang> <version>             Create .vern with language and version

vern stats                             Show disk usage per language/version

vern update                            Update vern + language list
vern update --only-self                Update vern binary only
vern update --only-langs               Update language list only

vern setup                             Create shims for version switching
vern implode                           Uninstall vern (select what to remove)
```

---

## Supported Languages

- Go
- Python
- Node.js
- Ruby
- Rust
- Zig

### Adding Programming Languages

Vern uses a simple YAML config file at `~/.config/vern/languages.yaml`.

To add a new language, add an entry to the languages list:

```yaml
languages:
  - name: rust
    binary_name: rustc
    version_source:
      url: "https://forge.rust-lang.org/infra/other-installation-methods.html"
      version_regex: "(\\d+\\.\\d+\\.\\d+)"
    install:
      download_template: "https://static.rust-lang.org/dist/rust-{{.Version}}-{{.Arch}}-unknown-linux-gnu.tar.gz"
      extract_type: "tar.gz"
      bin_rel_path: "rustc/bin/rustc"
```

#### Fields

| Field | Description |
|-------|-------------|
| `name` | Language identifier (used in commands) |
| `binary_name` | The executable name (for shims and `.vern` files) |
| `version_source.url` | URL to scrape for available versions |
| `version_source.version_regex` | Regex to extract version numbers from the page |
| `install.download_template` | Go template for download URL (see variables below) |
| `install.extract_type` | Archive type: `tar.gz` or `tar.xz` |
| `install.bin_rel_path` | Path to binary relative to install directory |
| `install.build_config` | Optional: configure command (e.g., `./configure --prefix={{.InstallDir}}`) |
| `install.build_command` | Optional: build command (e.g., `make -j$(nproc) && make install`) |

#### Template Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `{{.Version}}` | Full version | `1.21.0` |
| `{{.MajorMinor}}` | Major.Minor | `1.21` |
| `{{.OS}}` | Operating system | `linux`, `darwin` |
| `{{.Arch}}` | Architecture | `amd64`, `arm64` |
| `{{.ArchAlt}}` | Alt architecture name | `x64`, `arm64` |
| `{{.InstallDir}}` | Install directory path | (used in build commands) |

---

## Version Switching

Vern resolves versions in this order:

1. `.vern` file in current or parent directory
2. Global default for the language

Create a `.vern` file in your project root:

```bash
echo "python 3.11.2" > .vern
```

Set a global default:

```bash
vern default python 3.11.2
```

For version switching to work, ensure vern shims are in your PATH:

```bash
vern setup
```

Or add manually:

```bash
export PATH="$HOME/.local/share/vern/shims:$PATH"
```

---

## Updating the Language List

Vern maintains a versioned language list that can be updated independently of the binary.

To update to the latest supported languages:

```bash
vern update --only-langs
```

This fetches the latest `languages.yaml` from the repository and updates your local config.
The language list version is tracked in `~/.config/vern/langs_version`.

To update everything (vern binary + languages):

```bash
vern update
```

---

## Shell Setup

If `~/.local/bin` is not in your PATH, add it:

```bash
# For bash:
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# For zsh:
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# For fish:
fish_add_path "$HOME/.local/bin"
```

---

## Shell Completions

vern supports tab completions for bash, zsh, fish, and PowerShell:

```bash
# Bash
vern completion bash >> ~/.bashrc

# Zsh
vern completion zsh >> ~/.zshrc

# Fish
vern completion fish > ~/.config/fish/completions/vern.fish
```

---

GitHub: https://github.com/chris-roerig/vern
Issues: https://github.com/chris-roerig/vern/issues
