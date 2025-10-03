# Commitgen - AI-Powered Git Commit Message Generator

A standalone CLI tool that generates conventional commit messages using AI. Commitgen analyzes your git changes and creates descriptive, properly formatted commit messages automatically.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
  - [From Source](#from-source)
  - [Quick Install](#quick-install)
- [Usage](#usage)
  - [Basic Usage](#basic-usage)
  - [Using Different AI Providers](#using-different-ai-providers)
  - [Examples](#examples)
- [Configuration](#configuration)
- [Output Format](#output-format)
- [Development](#development)
  - [Building](#building)
  - [Testing](#testing)
  - [Installation for Development](#installation-for-development)
- [Advanced Usage](#advanced-usage)
  - [Custom Git Diff Analysis](#custom-git-diff-analysis)
  - [Troubleshooting](#troubleshooting)
  - [Integration with Git Hooks](#integration-with-git-hooks)
- [Contributing](#contributing)
- [Requirements](#requirements)
- [License](#license)

## Features

- **Multiple Analysis Modes**: Staged files, all changes, or untracked files
- **Conventional Commits**: Follows `type(scope): description` format
- **Smart Scope Detection**: Automatically extracts service/module from file paths
- **Multiple AI Providers**: Support for Claude, Gemini, and Copilot
- **Interactive**: Preview and confirm commit messages before committing

## Installation

### 🚀 One-Command Installation (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/FreePeak/commitgen/main/install-package.sh | bash
```

This script automatically detects your OS and installs commitgen using the best available package manager.

### 📦 Package Manager Installation

#### macOS (Homebrew)
```bash
brew tap FreePeak/tap
brew install commitgen
```

#### Go Install (Cross-platform)
```bash
go install github.com/FreePeak/commitgen@latest
```

#### Debian/Ubuntu (APT)
```bash
# Download and install .deb package
curl -fsSL https://raw.githubusercontent.com/FreePeak/commitgen/main/install-package.sh | METHOD=2 bash
```

#### Arch Linux (Pacman/AUR)
```bash
# Using yay (recommended)
yay -S commitgen

# Or using paru
paru -S commitgen

# Or manually
git clone https://aur.archlinux.org/commitgen.git
cd commitgen
makepkg -si
```

#### Snap (Cross-platform)
```bash
sudo snap install commitgen
```

### 🐳 Docker Installation
```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/freepeak/commitgen:latest

# Run with volume mount
docker run --rm -v $(pwd):/app -w /app ghcr.io/freepeak/commitgen:latest commit staged
```

### 📥 Manual Binary Download
```bash
# Download the latest binary for your platform
curl -fsSL https://raw.githubusercontent.com/FreePeak/commitgen/main/install.sh | bash
```

### 🔧 From Source

```bash
git clone https://github.com/FreePeak/commitgen.git
cd commitgen
go build -o commitgen main.go
./commitgen install
```

### ⚡ Quick Install (Power users)

```bash
# Clone and build in one command
git clone https://github.com/FreePeak/commitgen.git && cd commitgen && go build -o commitgen main.go && ./commitgen install
```

## Usage

### Basic Usage

```bash
# Generate commit message from staged files (default)
commitgen

# Or explicitly
commitgen commit staged
commitgen commit s

# Generate from all changes (stages + commits)
commitgen commit all
commitgen commit a

# Generate from untracked files only
commitgen commit untracked
commitgen commit u
```

### Using Different AI Providers

```bash
# Use Claude (default)
commitgen
commitgen --provider claude

# Use Gemini
commitgen --provider gemini

# Use Copilot
commitgen --provider copilot

# Use provider with specific subcommand
commitgen commit staged --provider gemini
commitgen commit all --provider copilot
```

### Examples

```bash
# Stage your changes first
git add .

# Generate and commit with AI
commitgen

# Output:
# Generated commit message:
# "feat(service:rating): add get RestaurantQuickReview with caching"
#
# Do you want to use this commit message? [y/N] y
# Committed successfully!

# Using different AI provider
commitgen --provider gemini commit staged

# Output:
# Generated commit message:
# "fix(api:user): resolve null pointer in validation"
#
# Do you want to use this commit message? [y/N] y
# Committed successfully!
```

## Configuration

Commitgen uses your existing AI CLI commands and configuration. Make sure you have:

1. AI CLI commands available in your PATH:
   - `claude` for Claude
   - `gemini` for Gemini
   - `copilot` for Copilot
2. Proper API keys configured for your chosen AI provider
3. Git repository initialized

## Output Format

Commitgen generates conventional commit messages following this format:

```
type(scope): description
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `chore`: Maintenance tasks

**Examples:**
- `feat(service:rating): add get RestaurantQuickReview with caching`
- `fix(api:user): resolve null pointer in validation`
- `docs(readme): update setup instructions`

## Development

### Building

```bash
go build -o commitgen main.go
```

### Testing

```bash
go test ./...
```

### Installation for Development

```bash
go build -o commitgen main.go
./commitgen install
```

## Requirements

- Go 1.19+
- Git
- AI CLI commands with proper API configuration:
  - `claude-glm` for Claude
  - `gemini` for Gemini
  - `copilot` for Copilot
- Unix-like system (Linux, macOS)

## Advanced Usage

### Custom Git Diff Analysis

Commitgen automatically analyzes git changes and provides context to the AI model:

- **Staged changes**: Uses `git diff --cached` to review staged modifications
- **All changes**: Combines modified files and untracked files for comprehensive analysis
- **Untracked files**: Reads content of new files that haven't been added to git yet

### Troubleshooting

#### Common Issues

**"not in a git repository"**
- Make sure you're in a directory with a `.git` folder
- Run `git init` if starting a new repository

**"no changes found to analyze"**
- Stage some files with `git add <files>`
- Or use `commitgen commit all` to include unstaged changes
- Use `commitgen commit untracked` for new files

**"failed to call [provider] API"**
- Ensure the AI CLI command is available in your PATH
- Verify API keys are properly configured
- Test the AI CLI command directly: `claude-glm` or `gemini` or `copilot`

#### Provider-Specific Setup

**Claude (Default)**
```bash
# Install claude-glm
pip install claude-glm
# Configure API key in environment
export ANTHROPIC_API_KEY="your-key-here"
```

**Gemini**
```bash
# Install Gemini CLI
npm install -g @google-cloud/vertexai
# Configure authentication
gcloud auth application-default login
```

**Copilot**
```bash
# Install GitHub CLI with Copilot extension
gh extension install github/gh-copilot
# Login to GitHub
gh auth login
```

### Integration with Git Hooks

You can integrate commitgen into your git workflow using hooks:

```bash
# Set up prepare-commit-msg hook
echo '#!/bin/bash
if [ -z "$2" ]; then
  ./commitgen --provider claude > .git/COMMIT_EDITMSG.tmp
  if [ -f .git/COMMIT_EDITMSG.tmp ]; then
    cat .git/COMMIT_EDITMSG.tmp > "$1"
    rm .git/COMMIT_EDITMSG.tmp
  fi
fi' > .git/hooks/prepare-commit-msg

chmod +x .git/hooks/prepare-commit-msg
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`commitgen commit staged`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details.
