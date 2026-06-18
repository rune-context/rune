#!/bin/sh
set -eu

# Rune installer
# Usage: curl -fsSL https://raw.githubusercontent.com/rune-context/rune/main/install.sh | sh

REPO="rune-context/rune"
BINARY="rune"

main() {
    os=$(detect_os)
    arch=$(detect_arch)

    echo "Rune installer"
    echo "  OS:   $os"
    echo "  Arch: $arch"
    echo ""

    # Determine download URL
    if [ "$os" = "windows" ]; then
        ext="zip"
    else
        ext="tar.gz"
    fi

    url="https://github.com/${REPO}/releases/latest/download/${BINARY}-${os}-${arch}.${ext}"
    echo "Downloading: $url"

    # Create temp directory
    tmp=$(mktemp -d)
    trap "rm -rf $tmp" EXIT

    # Download
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$tmp/archive.$ext"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$tmp/archive.$ext" "$url"
    else
        echo "Error: curl or wget required"
        exit 1
    fi

    # Extract
    cd "$tmp"
    if [ "$ext" = "zip" ]; then
        unzip -q "archive.zip"
    else
        tar -xzf "archive.tar.gz"
    fi

    # Install
    install_dir=$(detect_install_dir "$os")
    echo "Installing to: $install_dir"

    if [ -w "$install_dir" ]; then
        install -m 755 "$BINARY" "$install_dir/$BINARY"
    else
        echo "Need elevated permissions for $install_dir"
        sudo install -m 755 "$BINARY" "$install_dir/$BINARY"
    fi

    # Verify
    if command -v "$BINARY" >/dev/null 2>&1; then
        echo ""
        echo "✓ $($BINARY --version)"
        echo ""
        echo "Get started:"
        echo "  cd your-project"
        echo "  rune init"
        echo "  rune index"
    else
        echo ""
        echo "✓ Installed to $install_dir/$BINARY"
        echo ""
        echo "Make sure $install_dir is in your PATH:"
        echo "  export PATH=\"$install_dir:\$PATH\""
    fi
}

detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *)
            echo "Unsupported OS: $(uname -s)" >&2
            exit 1
            ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *)
            echo "Unsupported architecture: $(uname -m)" >&2
            exit 1
            ;;
    esac
}

detect_install_dir() {
    os="$1"
    case "$os" in
        linux)
            if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
                echo "/usr/local/bin"
            elif [ -d "$HOME/.local/bin" ]; then
                echo "$HOME/.local/bin"
            else
                mkdir -p "$HOME/.local/bin"
                echo "$HOME/.local/bin"
            fi
            ;;
        darwin)
            if [ -d "/usr/local/bin" ]; then
                echo "/usr/local/bin"
            elif [ -d "/opt/homebrew/bin" ]; then
                echo "/opt/homebrew/bin"
            else
                mkdir -p "$HOME/.local/bin"
                echo "$HOME/.local/bin"
            fi
            ;;
        windows)
            dir="$USERPROFILE/bin"
            mkdir -p "$dir" 2>/dev/null || true
            echo "$dir"
            ;;
    esac
}

main
