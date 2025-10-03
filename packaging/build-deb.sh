#!/bin/bash

set -e

# Build script for creating Debian packages
# Usage: ./packaging/build-deb.sh [version]

VERSION=${1:-1.0.0}
PACKAGE_NAME="commitgen"
ARCH="amd64"
BUILD_DIR="build/deb"

echo "Building Debian package for ${PACKAGE_NAME} v${VERSION}"

# Clean previous builds
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}/${PACKAGE_NAME}/DEBIAN
mkdir -p ${BUILD_DIR}/${PACKAGE_NAME}/opt/${PACKAGE_NAME}/bin

# Copy control files
cp packaging/deb/DEBIAN/* ${BUILD_DIR}/${PACKAGE_NAME}/DEBIAN/

# Update version in control file
sed -i "s/Version: .*/Version: ${VERSION}/" ${BUILD_DIR}/${PACKAGE_NAME}/DEBIAN/control

# Build binary for Linux amd64
echo "Building binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ${BUILD_DIR}/${PACKAGE_NAME}/opt/${PACKAGE_NAME}/bin/${PACKAGE_NAME} main.go

# Calculate installed size
INSTALLED_SIZE=$(du -s ${BUILD_DIR}/${PACKAGE_NAME}/opt | cut -f1)
sed -i "s/Installed-Size: .*/Installed-Size: ${INSTALLED_SIZE}/" ${BUILD_DIR}/${PACKAGE_NAME}/DEBIAN/control

# Build the package
echo "Creating .deb package..."
dpkg-deb --build ${BUILD_DIR}/${PACKAGE_NAME}

# Create final package name
PACKAGE_FILE="${BUILD_DIR}/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
mv ${BUILD_DIR}/${PACKAGE_NAME}.deb ${PACKAGE_FILE}

echo "✅ Debian package created: ${PACKAGE_FILE}"
echo "Install with: sudo dpkg -i ${PACKAGE_FILE}"