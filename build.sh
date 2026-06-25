#!/bin/bash
# MCP-Go Cross-Platform Build Script
# Usage: chmod +x build.sh && ./build.sh

set -e

OUTPUT_DIR="dist"
MODULE="./cmd/server"
APP_NAME="mcp-server"

# Clean previous builds
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

echo ""
echo "🔨 Building MCP-Go for all platforms..."
echo ""

build() {
    local goos=$1 goarch=$2 ext=$3
    local outfile="$OUTPUT_DIR/$APP_NAME-$goos-$goarch$ext"
    printf "  Building %s/%s... " "$goos" "$goarch"
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -ldflags="-s -w" -o "$outfile" "$MODULE"
    local size
    size=$(du -h "$outfile" | cut -f1)
    echo "✅ ($size)"
}

build windows amd64 .exe
build darwin  amd64 ""
build darwin  arm64 ""
build linux   amd64 ""

echo ""
echo "✅ All builds complete! Output in ./$OUTPUT_DIR/"
echo ""
ls -lh "$OUTPUT_DIR"
