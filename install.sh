#!/usr/bin/env bash
# ======================================================
# UMMC (UNDERTALE macOS Mod CLI) Quick Installer
# Can be run locally:  ./install.sh
# Can be run remotely: curl -fsSL https://raw.githubusercontent.com/siemvk/UMMC/main/install.sh | bash
# ======================================================

set -e

# Repository details
REPO_RAW_URL="https://github.com/siemvk/UMMC/archive/refs/heads/main.tar.gz"

# Formatting & Colors
BOLD='\033[1m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Default settings
INSTALL_DIR="$HOME/.local/bin"
SKIP_DEPS=false

usage() {
    echo -e "${BOLD}UMMC Quick Installer${NC}"
    echo "Usage: ./install.sh [options]"
    echo "       curl -fsSL https://raw.githubusercontent.com/siemvk/UMMC/main/install.sh | bash -s -- [options]"
    echo ""
    echo "Options:"
    echo "  -b, --bin-dir <dir>   Installation directory for UMMC binary (default: ~/.local/bin)"
    echo "  -s, --skip-deps       Skip checking and installing dependencies"
    echo "  -h, --help            Show this help message"
    exit 0
}

# Parse command line flags
while [[ $# -gt 0 ]]; do
    case "$1" in
        -b|--bin-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        -s|--skip-deps)
            SKIP_DEPS=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            error "Unknown option: $1"
            ;;
    esac
done

echo -e "${BOLD}${CYAN}------------------------------------------${NC}"
echo -e "${BOLD}${CYAN}   Installing UMMC (Undertale macOS Mod CLI)   ${NC}"
echo -e "${BOLD}${CYAN}------------------------------------------${NC}"

# Check OS
OS="$(uname -s)"
if [[ "$OS" != "Darwin" && "$OS" != "Linux" ]]; then
    warn "UMMC is designed primarily for macOS (and Linux). Current OS ($OS) may not be fully supported."
fi

# Ensure we have the source code
if [ -f "main.go" ] && [ -f "go.mod" ]; then
    info "Building from local repository source..."
else
    info "Downloading latest source code from GitHub..."
    TMP_DIR="$(mktemp -d)"
    trap 'rm -rf "$TMP_DIR"' EXIT
    
    if command -v curl >/dev/null 2>&1; then
        curl -sSL "$REPO_RAW_URL" | tar -xz -C "$TMP_DIR" --strip-components=1
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$REPO_RAW_URL" | tar -xz -C "$TMP_DIR" --strip-components=1
    else
        error "Neither curl nor wget is available to download source archive."
    fi

    cd "$TMP_DIR"
fi

# Dependency Management
if [ "$SKIP_DEPS" = false ]; then
    info "Checking required dependencies..."

    MISSING_BREW_FORMULAE=()
    MISSING_BREW_CASKS=()

    # Check for 'go'
    if ! command -v go >/dev/null 2>&1; then
        warn "'go' compiler is missing."
        MISSING_BREW_FORMULAE+=("go")
    else
        success "Go installed: $(go version | awk '{print $3}')"
    fi

    # Check for 'xdelta3'
    if ! command -v xdelta3 >/dev/null 2>&1; then
        warn "'xdelta3' is missing."
        MISSING_BREW_FORMULAE+=("xdelta3")
    else
        success "xdelta3 found: $(command -v xdelta3)"
    fi

    # Check for 'steamcmd'
    if ! command -v steamcmd >/dev/null 2>&1; then
        warn "'steamcmd' is missing."
        MISSING_BREW_CASKS+=("steamcmd")
    else
        success "steamcmd found: $(command -v steamcmd)"
    fi

    # Handle missing dependencies via Homebrew if available
    if [ ${#MISSING_BREW_FORMULAE[@]} -gt 0 ] || [ ${#MISSING_BREW_CASKS[@]} -gt 0 ]; then
        if command -v brew >/dev/null 2>&1; then
            info "Homebrew detected. Attempting to install missing dependencies..."
            if [ ${#MISSING_BREW_FORMULAE[@]} -gt 0 ]; then
                info "Installing formula(s): ${MISSING_BREW_FORMULAE[*]}"
                brew install "${MISSING_BREW_FORMULAE[@]}"
            fi
            if [ ${#MISSING_BREW_CASKS[@]} -gt 0 ]; then
                info "Installing cask(s): ${MISSING_BREW_CASKS[*]}"
                brew install --cask "${MISSING_BREW_CASKS[@]}"
            fi
        else
            warn "Homebrew is not installed. Please manually install the missing dependencies:"
            [ ${#MISSING_BREW_FORMULAE[@]} -gt 0 ] && echo "  - Formulae: ${MISSING_BREW_FORMULAE[*]}"
            [ ${#MISSING_BREW_CASKS[@]} -gt 0 ] && echo "  - Casks: ${MISSING_BREW_CASKS[*]}"
            if ! command -v go >/dev/null 2>&1; then
                error "Go compiler is required to build UMMC. Please install Go and re-run this script."
            fi
        fi
    fi
fi

# Build step
info "Downloading Go dependencies..."
go mod download

info "Building UMMC binary..."
go build -o UMMC .

success "Build successful!"

# Installation step
info "Installing UMMC binary to ${INSTALL_DIR}..."
mkdir -p "$INSTALL_DIR"
cp UMMC "$INSTALL_DIR/UMMC"
chmod +x "$INSTALL_DIR/UMMC"

success "UMMC binary successfully installed to ${INSTALL_DIR}/UMMC"

# Check if INSTALL_DIR is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    warn "${INSTALL_DIR} is not currently in your PATH environment variable."
    echo ""
    echo -e "To run ${BOLD}UMMC${NC} from anywhere, add the following line to your shell configuration file:"
    
    SHELL_NAME="$(basename "$SHELL")"
    if [ "$SHELL_NAME" = "zsh" ]; then
        echo -e "  ${CYAN}echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.zshrc && source ~/.zshrc${NC}"
    elif [ "$SHELL_NAME" = "bash" ]; then
        echo -e "  ${CYAN}echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.bashrc && source ~/.bashrc${NC}"
    elif [ "$SHELL_NAME" = "fish" ]; then
        echo -e "  ${CYAN}fish_add_path $INSTALL_DIR${NC}"
    else
        echo -e "  ${CYAN}export PATH=\"$INSTALL_DIR:\$PATH\"${NC}"
    fi
    echo ""
else
    echo ""
    echo -e "${GREEN}${BOLD}You can now run UMMC from your terminal using:${NC} ${CYAN}UMMC --help${NC}"
fi

echo -e "${BOLD}${GREEN}Installation complete!${NC}"
