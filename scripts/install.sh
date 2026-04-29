#!/bin/bash
set -e

VERN_REPO="chris-roerig/vern"
INSTALL_DIR="$HOME/.local/bin"
DATA_DIR="$HOME/.local/share/vern"
CONFIG_DIR="$HOME/.config/vern"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() { printf "${GREEN}%s${NC}\n" "$1"; }
warn() { printf "${YELLOW}%s${NC}\n" "$1"; }
error() { printf "${RED}%s${NC}\n" "$1" >&2; }

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    if [ "$ARCH" = "x86_64" ]; then
        ARCH="amd64"
    elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
        ARCH="arm64"
    fi

    if [ "$OS" != "linux" ] && [ "$OS" != "darwin" ]; then
        error "Unsupported OS: $OS"
        exit 1
    fi
}

get_latest_version() {
    info "Fetching latest vern version..."
    LATEST=$(curl -s "https://api.github.com/repos/$VERN_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$LATEST" ]; then
        error "Failed to fetch latest version"
        exit 1
    fi
    info "Latest version: $LATEST"
}

download_vern() {
    BINARY_NAME="vern-$LATEST-$OS-$ARCH"
    DOWNLOAD_URL="https://github.com/$VERN_REPO/releases/download/$LATEST/$BINARY_NAME"

    info "Downloading vern $LATEST for $OS/$ARCH..."
    if ! curl -fsSL "$DOWNLOAD_URL" -o /tmp/vern-new; then
        error "Download failed: $DOWNLOAD_URL"
        exit 1
    fi
    chmod +x /tmp/vern-new
}

install_vern() {
    mkdir -p "$INSTALL_DIR"

    if [ -w "$INSTALL_DIR" ] || [ "$INSTALL_DIR" = "$HOME/.local/bin" ]; then
        mv /tmp/vern-new "$INSTALL_DIR/vern"
        info "Installed vern to $INSTALL_DIR/vern"
    else
        warn "Need sudo to install to $INSTALL_DIR"
        sudo mv /tmp/vern-new "$INSTALL_DIR/vern"
    fi
}

setup_dirs() {
    info "Creating config and data directories..."
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$DATA_DIR"
}

download_languages() {
    info "Downloading language list..."
    MANIFEST_URL="https://raw.githubusercontent.com/$VERN_REPO/main/languages/manifest.json"
    MANIFEST=$(curl -fsSL "$MANIFEST_URL") || {
        warn "Failed to fetch manifest, using defaults on first run"
        return
    }
    LANGS_URL=$(echo "$MANIFEST" | grep '"langs_url"' | sed -E 's/.*"langs_url"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/')
    LANGS_VERSION=$(echo "$MANIFEST" | grep '"latest_langs_version"' | sed -E 's/.*"latest_langs_version"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/')

    if [ -z "$LANGS_URL" ]; then
        warn "Failed to parse manifest, will use defaults on first run"
        return
    fi

    if curl -fsSL "$LANGS_URL" -o "$CONFIG_DIR/languages.yaml"; then
        info "Language list installed to $CONFIG_DIR/languages.yaml"
    else
        warn "Failed to download language list, will use defaults on first run"
    fi

    echo "${LANGS_VERSION:-1.0.0}" > "$CONFIG_DIR/langs_version"
}

setup_path() {
    if echo "$PATH" | grep -q "$HOME/.local/bin"; then
        info "~/.local/bin is already in PATH"
        return
    fi

    warn "~/.local/bin is not in your PATH"
    printf "Add ~/.local/bin to PATH now? [y/N] "
    read -r answer

    if [ "$answer" = "y" ] || [ "$answer" = "Y" ]; then
        SHELL_NAME=$(basename "$SHELL")
        case "$SHELL_NAME" in
            bash)
                echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
                info "Added to ~/.bashrc. Run: source ~/.bashrc"
                ;;
            zsh)
                echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.zshrc"
                info "Added to ~/.zshrc. Run: source ~/.zshrc"
                ;;
            fish)
                echo 'fish_add_path "$HOME/.local/bin"' >> "$HOME/.config/fish/config.fish"
                info "Added to fish config. Run: source ~/.config/fish/config.fish"
                ;;
            *)
                warn "Unknown shell: $SHELL_NAME"
                print_manual_path
                ;;
        esac
    else
        print_manual_path
    fi
}

print_manual_path() {
    warn "To add ~/.local/bin to PATH manually, add this to your shell config:"
    echo ""
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
    echo "For fish shell:"
    echo "  fish_add_path \"\$HOME/.local/bin\""
}

main() {
    info "Installing vern..."
    echo ""

    detect_platform
    get_latest_version
    download_vern
    install_vern
    setup_dirs
    download_languages
    setup_path

    echo ""
    info "Vern $LATEST installed successfully!"
    info "Run 'vern help' to get started."
}

main
