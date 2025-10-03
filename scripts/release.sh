#!/bin/bash

# Commitgen Release Script
# Automates semantic versioning and release process

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
REPO="FreePeak/commitgen"
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

print_header() {
    echo -e "${BLUE}🚀 Commitgen Release Script${NC}"
    echo -e "${BLUE}===========================${NC}"
    echo
    print_info "Current version: ${CURRENT_VERSION}"
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

# Validate version format
validate_version() {
    local version=$1
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        print_error "Invalid version format. Use semantic versioning: v1.2.3"
        return 1
    fi
}

# Bump version based on commit messages
bump_version() {
    local type=$1
    local current=${CURRENT_VERSION#v}  # Remove 'v' prefix

    # Parse current version
    IFS='.' read -ra PARTS <<< "$current"
    local major=${PARTS[0]}
    local minor=${PARTS[1]}
    local patch=${PARTS[2]}

    case $type in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            print_error "Invalid bump type. Use: major, minor, patch"
            exit 1
            ;;
    esac

    echo "v$major.$minor.$patch"
}

# Check if working directory is clean
check_git_status() {
    if [[ -n $(git status --porcelain) ]]; then
        print_error "Working directory is not clean. Commit or stash changes first."
        exit 1
    fi
}

# Run tests and build
run_tests_and_build() {
    print_info "Running tests..."
    go test ./...

    print_info "Building binary..."
    make build

    print_success "Tests passed and build successful!"
}


# Create release commit and tag
create_release() {
    local new_version=$1

    print_info "Creating release commit and tag..."

    # Update version in main.go if needed
    if grep -q "version.*=" main.go; then
        sed -i.bak "s/version.*= \".*\"/version = \"$new_version\"/" main.go
        rm main.go.bak
    fi

    # Commit changes
    git add .
    git commit -m "chore: release $new_version"

    # Create tag
    git tag -a "$new_version" -m "Release $new_version"

    print_success "Created release commit and tag: $new_version"
    print_info "GoReleaser will automatically update Homebrew formula when tag is pushed"
}

# Push to GitHub
push_to_github() {
    local version=$1

    print_info "Pushing to GitHub..."
    git push origin main
    git push origin "$version"

    print_success "Pushed to GitHub!"
}

# Create GitHub release
create_github_release() {
    local version=$1

    print_info "Creating GitHub release..."

    # Get changelog since last release
    local changelog=$(git log --pretty=format:"- %s" ${CURRENT_VERSION}..HEAD)

    # Create release using gh CLI
    if command -v gh >/dev/null 2>&1; then
        gh release create "$version" \
            --title "Commitgen $version" \
            --notes "## Changes since ${CURRENT_VERSION}

$changelog

## Installation

### One-command installation:
\`\`\`bash
curl -fsSL https://raw.githubusercontent.com/FreePeak/commitgen/main/install-package.sh | bash
\`\`\`

### Package managers:
- **Homebrew**: \`brew install FreePeak/tap/commitgen\`
- **Go**: \`go install github.com/FreePeak/commitgen@${version}\`
- **APT**: Download .deb from releases
- **AUR**: \`yay -S commitgen\`

### Docker:
\`\`\`bash
docker pull ghcr.io/freepeak/commitgen:${version}
\`\`\`"

        print_success "GitHub release created!"
    else
        print_warning "GitHub CLI not found. Create release manually:"
        print_info "  https://github.com/${REPO}/releases/new?tag=${version}"
    fi
}

# Main function
main() {
    print_header

    # Check git status
    check_git_status

    # Parse arguments
    local version=""
    local bump_type=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                version="$2"
                shift 2
                ;;
            -t|--type)
                bump_type="$2"
                shift 2
                ;;
            -h|--help)
                echo "Commitgen Release Script"
                echo
                echo "Usage: $0 [options]"
                echo
                echo "Options:"
                echo "  -v, --version VERSION    Specific version to release (e.g., v1.2.3)"
                echo "  -t, --type TYPE         Auto-bump type: major, minor, patch"
                echo "  -h, --help              Show this help message"
                echo
                echo "Features:"
                echo "  - Updates version in main.go"
                echo "  - Pushes tag to trigger GoReleaser"
                echo "  - GoReleaser handles Homebrew formula automatically"
                echo "  - Creates GitHub release"
                echo "  - Runs tests and builds"
                echo
                echo "Examples:"
                echo "  $0 -t patch             # Bump patch version"
                echo "  $0 -t minor             # Bump minor version"
                echo "  $0 -v v1.2.3           # Release specific version"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    # Determine new version
    if [[ -n "$version" ]]; then
        validate_version "$version"
    elif [[ -n "$bump_type" ]]; then
        version=$(bump_version "$bump_type")
    else
        print_error "Please specify either --version or --type"
        print_info "Use --help for usage information"
        exit 1
    fi

    print_info "New version: $version"
    echo

    # Confirm release
    read -p "Do you want to proceed with release $version? [y/N] " confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        print_info "Release cancelled."
        exit 0
    fi

    echo

    # Run release process
    run_tests_and_build
    create_release "$version"
    push_to_github "$version"
    create_github_release "$version"

    echo
    print_success "Release $version completed successfully! 🎉"
    print_info "Release page: https://github.com/${REPO}/releases/tag/${version}"
}

# Check dependencies
for cmd in git go; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        print_error "$cmd is required but not installed."
        exit 1
    fi
done

# Run main function
main "$@"