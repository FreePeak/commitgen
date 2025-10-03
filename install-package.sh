#!/bin/bash

# Commitgen Package Manager Installation Script
# This script helps install commitgen using various package managers

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

REPO="FreePeak/commitgen"

print_header() {
    echo -e "${BLUE}🚀 Commitgen Package Manager Installation${NC}"
    echo -e "${BLUE}======================================${NC}"
    echo
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if command -v apt >/dev/null 2>&1; then
            echo "debian"
        elif command -v pacman >/dev/null 2>&1; then
            echo "arch"
        elif command -v yum >/dev/null 2>&1; then
            echo "rhel"
        elif command -v zypper >/dev/null 2>&1; then
            echo "suse"
        else
            echo "linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    else
        echo "unknown"
    fi
}

# Homebrew installation
install_homebrew() {
    print_info "Installing with Homebrew..."

    if ! command -v brew >/dev/null 2>&1; then
        print_error "Homebrew not found. Installing Homebrew first..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi

    # Add tap and install
    brew tap FreePeak/tap 2>/dev/null || print_info "Tap already exists"
    brew install commitgen

    print_success "Installed with Homebrew!"
}

# Go installation
install_go() {
    print_info "Installing with Go..."

    if ! command -v go >/dev/null 2>&1; then
        print_error "Go not found. Please install Go first:"
        print_info "  https://golang.org/dl/"
        return 1
    fi

    go install github.com/${REPO}@latest

    # Check GOPATH
    GOBIN=$(go env GOBIN)
    if [[ -z "$GOBIN" ]]; then
        GOBIN=$(go env GOPATH)/bin
    fi

    if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
        print_warning "$GOBIN is not in your PATH"
        print_info "Add to your shell profile: export PATH=\"\$PATH:$GOBIN\""
    fi

    print_success "Installed with Go!"
}

# APT installation (Debian/Ubuntu)
install_apt() {
    print_info "Installing with APT..."

    # Download .deb package
    API_URL="https://api.github.com/repos/${REPO}/releases/latest"
    DOWNLOAD_URL=$(curl -s "$API_URL" | grep -E '"browser_download_url":.*\.deb"' | cut -d '"' -f 4)

    if [[ -z "$DOWNLOAD_URL" ]]; then
        print_error "No .deb package found"
        return 1
    fi

    TEMP_DEB=$(mktemp)
    curl -L -o "$TEMP_DEB" "$DOWNLOAD_URL"

    sudo dpkg -i "$TEMP_DEB" || sudo apt-get install -f -y
    rm -f "$TEMP_DEB"

    print_success "Installed with APT!"
}

# Pacman installation (Arch Linux)
install_pacman() {
    print_info "Installing with Pacman..."

    # Try AUR helpers
    if command -v yay >/dev/null 2>&1; then
        yay -S commitgen
    elif command -v paru >/dev/null 2>&1; then
        paru -S commitgen
    else
        print_warning "No AUR helper found. Installing yay first..."
        sudo pacman -S --needed git base-devel
        git clone https://aur.archlinux.org/yay.git /tmp/yay
        cd /tmp/yay
        makepkg -si --noconfirm
        cd - && rm -rf /tmp/yay

        yay -S commitgen
    fi

    print_success "Installed with Pacman/AUR!"
}

# Snap installation
install_snap() {
    print_info "Installing with Snap..."

    if ! command -v snap >/dev/null 2>&1; then
        print_error "Snap not found. Please install Snapd first."
        return 1
    fi

    sudo snap install commitgen

    print_success "Installed with Snap!"
}

# Main installation
main() {
    print_header

    OS=$(detect_os)
    print_info "Detected OS: $OS"
    echo

    # Show available options
    echo "Available installation methods:"
    echo "1) Go install (cross-platform, recommended)"

    case $OS in
        macos)
            echo "2) Homebrew"
            ;;
        debian)
            echo "2) APT (Debian package)"
            echo "3) Snap"
            ;;
        arch)
            echo "2) Pacman (AUR)"
            ;;
        linux)
            echo "2) Snap"
            ;;
    esac

    echo "4) Manual binary download"
    echo

    # Get user choice
    if [[ -n "$METHOD" ]]; then
        CHOICE="$METHOD"
    else
        read -p "Choose installation method (1-4): " CHOICE
    fi

    case $CHOICE in
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
                    exit 1
                    ;;
            esac
            ;;
        3)
            if [[ "$OS" == "debian" ]]; then
                install_snap
            else
                print_error "Invalid choice for $OS"
                exit 1
            fi
            ;;
        4)
            print_info "Downloading binary directly..."
            curl -fsSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash
            ;;
        *)
            print_error "Invalid choice"
            exit 1
            ;;
    esac

    echo
    if command -v commitgen >/dev/null 2>&1; then
        print_success "Commitgen installed successfully!"
        print_info "Run 'commitgen --help' to get started."
    else
        print_error "Installation failed. Please check your PATH."
        exit 1
    fi
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--method)
            METHOD="$2"
            shift 2
            ;;
        -h|--help)
            echo "Commitgen Package Manager Installation"
            echo
            echo "Usage: $0 [options]"
            echo
            echo "Options:"
            echo "  -m, --method METHOD    Installation method (1=go, 2=package, 3=snap, 4=binary)"
            echo "  -h, --help            Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

main