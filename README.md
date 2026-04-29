```
                    ▄▄         ▄▄   
▄▄ ▄▄ ▄▄▄▄ ▄▄▄▄     █   ▄▄  ▄▄  █   
██▄██ ██▄▄  ██▄█▄  ██   ███▄██  ██  
 ▀█▀  ██▄▄▄ ██ ██   █   ██ ▀██  █   
                    ▀▀         ▀▀   
```

# vern{n}
 
**The version manager that just makes sense.**

> A programming language version installation manager that's intuitive, fast, and actually works the way you expect.

---

## QUICK START

Install vern with a single command:

    curl -fsSL https://raw.githubusercontent.com/chris-roerig/vern/main/scripts/install.sh | bash

The installer will:
  • Detect your OS and architecture
  • Download the latest vern binary
  • Install to ~/.local/bin (or /usr/local/bin with sudo)
  • Create config directories
  • Download the supported languages list
  • Optionally add ~/.local/bin to your PATH

---

## COMMANDS

    vern                    Print version and help
    vern help               Print detailed help

    vern install <lang>                Install latest version
    vern install <lang> 3              Install latest 3.x.x
    vern install <lang> 3.11           Install latest 3.11.x
    vern install <lang> 3.11.2         Install exact version

    vern list                          List all installed languages
    vern list <lang>                   List installed versions for language

    vern default <lang>                Interactively select default
    vern default <lang> <version>      Set default directly

    vern remove <lang>                 Multi-select versions to remove

    vern update                        Update vern + language list
    vern update --only-self            Update vern binary only
    vern update --only-langs           Update language list only

    vern setup                         Create shims for version switching

---

## CURRENT OFFICIALLY SUPPORTED LANGUAGES

    • Go
    • Python
    • Node.js
    • Ruby

### ADDING PROGRAMMING LANGUAGES

Vern uses a simple YAML config file at ~/.config/vern/languages.yaml

To add a new language, add an entry to the languages list:

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

#### Fields explained:
    name           - Language identifier (used in commands)
    binary_name    - The executable name (for shims and .vern files)
    version_source
      url          - URL to scrape for available versions
      version_regex - Regex to extract version numbers from the page
    install
      download_template - Go template for download URL
                          Available variables:
                            {{.Version}}    - Full version (e.g., 1.21.0)
                            {{.MajorMinor}} - Major.Minor (e.g., 1.21)
                            {{.OS}}         - Operating system (linux, darwin)
                            {{.Arch}}       - Architecture (amd64, arm64)
                            {{.ArchAlt}}    - Alt architecture name (x64, arm64)
      extract_type      - Archive type: "tar.gz" or "tar.xz"
      bin_rel_path      - Path to binary relative to install directory
      build_config      - Optional: configure command (e.g., ./configure --prefix={{.InstallDir}})
      build_command     - Optional: build command (e.g., make -j$(nproc) && make install)

---

## VERSION SWITCHING

Vern resolves versions in this order:
    1. .vern file in current or parent directory
    2. Global default for the language

Create a .vern file in your project root:

    echo "python 3.11.2" > .vern

Set a global default:

    vern default python 3.11.2

For version switching to work, ensure vern shims are in your PATH:

    vern setup

Or add manually:

    export PATH="$HOME/.local/share/vern/shims:$PATH"

---

## UPDATING THE OFFICIALLY SUPPORTED LANGUAGES LIST

Vern maintains a versioned language list that can be updated independently of the binary.

To update to the latest supported languages:

    vern update --only-langs

This fetches the latest languages.yaml from the repository and updates your local config.
The language list version is tracked in ~/.config/vern/langs_version.

To update everything (vern binary + languages):

    vern update

---

## SHELL SETUP

If ~/.local/bin is not in your PATH, add it:

    # For bash:
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
    source ~/.bashrc

    # For zsh:
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
    source ~/.zshrc

    # For fish:
    fish_add_path "$HOME/.local/bin"

---

GitHub: https://github.com/chris-roerig/vern  
Issues: https://github.com/chris-roerig/vern/issues
