#!/usr/bin/env bash
set -euo pipefail

# tengo installer
# Usage: curl -fsSL https://raw.githubusercontent.com/helloWorld44-89/tengo/main/install.sh | bash

REPO="helloWorld44-89/tengo"
BINARY="tengo"
INSTALL_DIR=""

# ── Helpers ──────────────────────────────────────────────────────────────────

info()  { printf '  \033[34m•\033[0m %s\n' "$*"; }
ok()    { printf '  \033[32m✓\033[0m %s\n' "$*"; }
err()   { printf '  \033[31m✗\033[0m %s\n' "$*" >&2; }
die()   { err "$*"; exit 1; }

need() {
    command -v "$1" >/dev/null 2>&1 || die "Required command not found: $1"
}

# ── Detect OS and arch ───────────────────────────────────────────────────────

detect_platform() {
    local os arch

    case "$(uname -s)" in
        Linux)  os="linux"  ;;
        Darwin) os="darwin" ;;
        *)      die "Unsupported OS: $(uname -s). Install manually from https://github.com/$REPO/releases" ;;
    esac

    case "$(uname -m)" in
        x86_64)          arch="amd64" ;;
        aarch64|arm64)   arch="arm64" ;;
        *)               die "Unsupported architecture: $(uname -m)" ;;
    esac

    echo "${BINARY}-${os}-${arch}"
}

# ── Pick install directory ────────────────────────────────────────────────────

pick_install_dir() {
    # Prefer /usr/local/bin if writable or sudo is available.
    if [ -w /usr/local/bin ]; then
        echo "/usr/local/bin"
    elif command -v sudo >/dev/null 2>&1; then
        echo "/usr/local/bin"
    else
        # Fall back to user-local bin.
        local dir="$HOME/.local/bin"
        mkdir -p "$dir"
        echo "$dir"
    fi
}

# ── Download ─────────────────────────────────────────────────────────────────

download() {
    local url="$1" dest="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL --progress-bar "$url" -o "$dest"
    elif command -v wget >/dev/null 2>&1; then
        wget -q --show-progress "$url" -O "$dest"
    else
        die "Neither curl nor wget found. Please install one and try again."
    fi
}

# ── Main ─────────────────────────────────────────────────────────────────────

main() {
    printf '\n  \033[1mtengo installer\033[0m\n\n'

    local asset
    asset="$(detect_platform)"
    info "Detected platform: $asset"

    INSTALL_DIR="$(pick_install_dir)"
    info "Install directory:  $INSTALL_DIR"

    local url="https://github.com/$REPO/releases/latest/download/$asset"
    local tmp
    tmp="$(mktemp)"
    trap 'rm -f "${tmp:-}"' EXIT

    info "Downloading latest release..."
    download "$url" "$tmp"
    chmod +x "$tmp"

    # Verify the binary runs.
    local ver
    ver="$("$tmp" -version 2>/dev/null || true)"
    [ -n "$ver" ] || die "Downloaded binary failed to run — please report this at https://github.com/$REPO/issues"

    # Install (use sudo only if the directory isn't writable).
    local dest="$INSTALL_DIR/$BINARY"
    if [ -w "$INSTALL_DIR" ]; then
        mv "$tmp" "$dest"
    else
        info "Requesting sudo to write to $INSTALL_DIR..."
        sudo mv "$tmp" "$dest"
    fi

    ok "Installed $ver to $dest"

    # Warn if the install dir isn't in PATH.
    if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
        printf '\n  \033[33m!\033[0m %s is not in your PATH.\n' "$INSTALL_DIR"
        printf '    Add this to your shell profile:\n'
        printf '    \033[2mexport PATH="%s:$PATH"\033[0m\n' "$INSTALL_DIR"
    fi

    # Shell completion hint.
    printf '\n  \033[2mOptional — add shell completions:\033[0m\n'
    printf '    bash:  echo '\''source <(tengo -completion bash)'\'' >> ~/.bashrc\n'
    printf '    zsh:   tengo -completion zsh > ~/.zfunc/_tengo\n'
    printf '    fish:  tengo -completion fish > ~/.config/fish/completions/tengo.fish\n'
    printf '\n'
}

main "$@"
