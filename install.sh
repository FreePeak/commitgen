#!/bin/bash

# Commitgen Installation Script
# Supports multiple installation methods across different platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="FreePeak/commitgen"
BINARY_NAME="commitgen"
INSTALL_DIR="/usr/local/bin"
VERSION=${COMMITGEN_VERSION:-"latest"}

# Helper functions
print_header() {
    echo -e "${BLUE}🚀 Commitgen Installation Script${NC}"
    echo -e "${BLUE}=================================${NC}"
    echo
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Detect operating system
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if command -v apt >/dev/null 2>&1; then
            echo "debian"
        elif command -v pacman >/dev/null 2>&1; then
            echo "arch"
        elif command -v yum >/dev/null 2>&1; then
            echo "rhel"
        else
            echo "linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
        echo "windows"
    else
        echo "unknown"
    fi
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install using Homebrew (macOS)
install_homebrew() {
    print_info "Installing using Homebrew..."

    if ! command_exists brew; then
        print_error "Homebrew not found. Please install Homebrew first:"
        print_info "  /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        return 1
    fi

    # Check if formula exists in custom tap, otherwise install from source
    if brew list --repo | grep -q "FreePeak/tap" 2>/dev/null; then
        brew install FreePeak/tap/commitgen
    else
        print_info "Adding custom tap..."
        brew tap FreePeak/tap
        brew install commitgen
    fi
}

# Install using Go
install_go() {
    print_info "Installing using Go..."

    if ! command_exists go; then
        print_error "Go not found. Please install Go first:"
        print_info "  https://golang.org/dl/"
        return 1
    fi

    if [[ "$VERSION" == "latest" ]]; then
        go install github.com/${REPO}@latest
    else
        go install github.com/${REPO}@v${VERSION}
    fi

    # Check if GOPATH/bin is in PATH
    GOBIN=$(go env GOBIN)
    if [[ -z "$GOBIN" ]]; then
        GOBIN=$(go env GOPATH)/bin
    fi

    if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
        print_warning "$GOBIN is not in your PATH"
        print_info "Add the following to your shell profile:"
        print_info "  export PATH=\"\$PATH:$GOBIN\""
    fi
}

# Install using apt (Debian/Ubuntu)
install_apt() {
    print_info "Installing using apt..."

    if ! command_exists apt; then
        print_error "apt not found. This method only works on Debian/Ubuntu systems."
        return 1
    fi

    # Try to install from PPA or download .deb package
    print_info "Downloading .deb package..."

    # Get the latest release
    API_URL="https://api.github.com/repos/${REPO}/releases/${VERSION}"
    if [[ "$VERSION" == "latest" ]]; then
        API_URL="https://api.github.com/repos/${REPO}/releases/latest"
    fi

    DOWNLOAD_URL=$(curl -s "$API_URL" | grep -E '"browser_download_url":.*\.deb"' | cut -d '"' -f 4)

    if [[ -z "$DOWNLOAD_URL" ]]; then
        print_error "No .deb package found for version $VERSION"
        return 1
    fi

    TEMP_DEB=$(mktemp)
    curl -L -o "$TEMP_DEB" "$DOWNLOAD_URL"

    print_info "Installing .deb package..."
    sudo dpkg -i "$TEMP_DEB" || sudo apt-get install -f -y

    rm -f "$TEMP_DEB"
}

# Install using pacman (Arch Linux)
install_pacman() {
    print_info "Installing using pacman..."

    if ! command_exists pacman; then
        print_error "pacman not found. This method only works on Arch Linux systems."
        return 1
    fi

    # Try to install from AUR
    if command_exists yay; then
        print_info "Installing using yay (AUR helper)..."
        yay -S commitgen
    elif command_exists paru; then
        print_info "Installing using paru (AUR helper)..."
        paru -S commitgen
    else
        print_warning "No AUR helper found. Manual installation required."
        print_info "Please install an AUR helper (yay or paru) first:"
        print_info "  yay -S yay"
        print_info "  or"
        print_info "  sudo pacman -S --needed git base-devel"
        print_info "  git clone https://aur.archlinux.org/yay.git"
        print_info "  cd yay && makepkg -si"
        return 1
    fi
}

# Install from binary
install_binary() {
    print_info "Installing from binary..."

    # Detect OS and architecture
    OS=$(detect_os)
    ARCH=$(uname -m)

    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            return 1
            ;;
    esac

    case $OS in
        macos)
            OS="darwin"
            ;;
        linux)
            OS="linux"
            ;;
        windows)
            OS="windows"
            ;;
        *)
            print_error "Unsupported operating system: $OS"
            return 1
            ;;
    esac

    # Get the latest release
    API_URL="https://api.github.com/repos/${REPO}/releases/${VERSION}"
    if [[ "$VERSION" == "latest" ]]; then
        API_URL="https://api.github.com/repos/${REPO}/releases/latest"
    fi

    DOWNLOAD_URL=$(curl -s "$API_URL" | grep -E "\"browser_download_url\":.*${OS}_${ARCH}" | cut -d '"' -f 4)

    if [[ -z "$DOWNLOAD_URL" ]]; then
        print_error "No binary found for ${OS}_${ARCH} version $VERSION"
        return 1
    fi

    # Download and extract
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"

    print_info "Downloading from $DOWNLOAD_URL..."
    curl -L -o "commitgen.tar.gz" "$DOWNLOAD_URL"
    tar -xzf "commitgen.tar.gz"

    # Find and copy binary
    BINARY_PATH=$(find . -name "$BINARY_NAME" -type f | head -n1)

    if [[ -z "$BINARY_PATH" ]]; then
        print_error "Binary not found in archive"
        rm -rf "$TEMP_DIR"
        return 1
    fi

    print_info "Installing binary to $INSTALL_DIR..."
    if [[ -w "$INSTALL_DIR" ]]; then
        cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    else
        sudo cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi

    rm -rf "$TEMP_DIR"
}

# Main installation logic
main() {
    print_header

    OS=$(detect_os)
    print_info "Detected OS: $OS"
    print_info "Requested version: $VERSION"
    echo

    # Ask for installation method
    echo "Available installation methods:"
    echo "1) Go install (recommended for developers)"
    if [[ "$OS" == "macos" ]]; then
        echo "2) Homebrew"
    elif [[ "$OS" == "debian" ]]; then
        echo "2) apt (Debian package)"
    elif [[ "$OS" == "arch" ]]; then
        echo "2) pacman/AUR"
    fi
    echo "3) Binary download"
    echo

    if [[ -n "$INSTALL_METHOD" ]]; then
        METHOD="$INSTALL_METHOD"
    else
        read -p "Choose installation method (1-3): " METHOD
    fi

    case $METHOD in
        1)
            install_go
            ;;
        2)
            case $OS in
                macos)
                    install_homebrew
                    ;;
                debian)
                    install_apt
                    ;;
                arch)
                    install_pacman
                    ;;
                *)
                    print_error "Package installation not supported on $OS"
                    print_info "Try method 1 (Go install) or 3 (Binary download)"
                    exit 1
                    ;;
            esac
            ;;
        3)
            install_binary
            ;;
        *)
            print_error "Invalid method selected"
            exit 1
            ;;
    esac

    # Verify installation
    if command_exists commitgen; then
        print_success "Commitgen installed successfully!"
        echo
        print_info "Run 'commitgen --help' to get started."
        print_info "Visit https://github.com/${REPO} for documentation."
    else
        print_error "Installation verification failed"
        print_info "Please try a different installation method or check your PATH"
        exit 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -m|--method)
            INSTALL_METHOD="$2"
            shift 2
            ;;
        -h|--help)
            echo "Commitgen Installation Script"
            echo
            echo "Usage: $0 [options]"
            echo
            echo "Options:"
            echo "  -v, --version VERSION  Install specific version (default: latest)"
            echo "  -m, --method METHOD    Installation method (1=go, 2=package, 3=binary)"
            echo "  -h, --help            Show this help message"
            echo
            echo "Environment variables:"
            echo "  COMMITGEN_VERSION      Set version to install"
            echo "  INSTALL_METHOD         Set installation method"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main function
main