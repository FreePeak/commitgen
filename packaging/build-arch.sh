#!/bin/bash

set -e

# Build script for creating Arch Linux packages
# Usage: ./packaging/build-arch.sh [version]

VERSION=${1:-1.0.0}
PACKAGE_NAME="commitgen"
BUILD_DIR="build/arch"

echo "Building Arch Linux package for ${PACKAGE_NAME} v${VERSION}"

# Clean previous builds
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

# Copy packaging files to build directory
cp -r packaging/arch/* ${BUILD_DIR}/

# Update version in PKGBUILD
sed -i "s/pkgver=.*/pkgver=${VERSION}/" ${BUILD_DIR}/PKGBUILD
sed -i "s/pkgrel=.*/pkgrel=1/" ${BUILD_DIR}/PKGBUILD

# Update source URL with version
sed -i "s|source=.*|source=(\"$pkgname-\$pkgver.tar.gz::https://github.com/FreePeak/commitgen/archive/refs/tags/v\$pkgver.tar.gz\")|" ${BUILD_DIR}/PKGBUILD

# Update .SRCINFO
cd ${BUILD_DIR}
makepkg --printsrcinfo > .SRCINFO

# Build the package
echo "Creating .pkg.tar.zst package..."
makepkg -f

# Find the built package
PACKAGE_FILE=$(ls ${PACKAGE_NAME}-${VERSION}-*.pkg.tar.zst | head -n1)

echo "✅ Arch Linux package created: ${BUILD_DIR}/${PACKAGE_FILE}"
echo "Install with: sudo pacman -U ${BUILD_DIR}/${PACKAGE_FILE}"
echo "Or using yay: yay -U ${BUILD_DIR}/${PACKAGE_FILE}"