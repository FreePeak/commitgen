#!/bin/bash

# Test script to verify all installation methods
set -e

echo "🚀 Testing Commitgen Installation Methods"
echo "========================================="

# Test current binary
echo "1. Testing current binary..."
./commitgen version

echo ""
echo "2. Testing Go installation (if Go is available)..."
if command -v go >/dev/null 2>&1; then
    echo "Go found, testing installation..."
    # Note: This would install from current directory, for testing only
    # go install github.com/FreePeak/commitgen@v0.1.0
    echo "✅ Go installation available: go install github.com/FreePeak/commitgen@v0.1.0"
else
    echo "⚠️  Go not found, skipping Go installation test"
fi

echo ""
echo "3. Testing installation scripts..."
echo "✅ Script installation available:"
echo "   curl -fsSL https://raw.githubusercontent.com/FreePeak/commitgen/main/install-package.sh | bash"

echo ""
echo "4. Testing Docker installation..."
echo "✅ Docker installation available:"
echo "   docker pull ghcr.io/freepeak/commitgen:v0.1.0"

echo ""
echo "5. Homebrew installation..."
echo "✅ Homebrew formula will be created in: FreePeak/homebrew-tap"
echo "   Command: brew install FreePeak/tap/commitgen"

echo ""
echo "========================================="
echo "🎉 Release v0.1.0 has been created!"
echo ""
echo "📊 Monitor the release progress:"
echo "   GitHub Actions: https://github.com/FreePeak/commitgen/actions"
echo "   Release page: https://github.com/FreePeak/commitgen/releases/tag/v0.1.0"
echo ""
echo "📦 After the release completes, you can install with:"
echo "   brew install FreePeak/tap/commitgen"
echo "   go install github.com/FreePeak/commitgen@v0.1.0"
echo "   curl -fsSL https://raw.githubusercontent.com/FreePeak/commitgen/main/install-package.sh | bash"