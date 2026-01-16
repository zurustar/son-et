#!/bin/bash
# Build script for creating embedded son-et executables
# Usage: ./build_embedded.sh <project_name>
# Example: ./build_embedded.sh kuma2

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <project_name>"
    echo ""
    echo "Available projects:"
    echo "  kuma2    - Kuma2 sample game"
    echo ""
    echo "Example:"
    echo "  $0 kuma2"
    exit 1
fi

PROJECT=$1
PROJECT_DIR="samples/${PROJECT}"
OUTPUT="${PROJECT}"

echo "Building embedded executable for: ${PROJECT}"
echo "Project directory: ${PROJECT_DIR}"
echo "Output: ${OUTPUT}"
echo ""

# Check if project directory exists
if [ ! -d "${PROJECT_DIR}" ]; then
    echo "Error: Project directory not found: ${PROJECT_DIR}"
    exit 1
fi

# Create a temporary Go file with the embed directive in the project directory
TEMP_FILE="${PROJECT_DIR}/embed_generated.go"

cat > "${TEMP_FILE}" << EOF
package main

import "embed"

//go:embed *
var embeddedFSData embed.FS

func init() {
	embeddedFS = embeddedFSData
	projectDir = "${PROJECT_DIR}"
	projectName = "${PROJECT}"
}
EOF

echo "Created temporary embed file"

# Copy main.go to the project directory temporarily
cp cmd/son-et-embedded/main.go "${PROJECT_DIR}/main_generated.go"

# Build from the project directory
echo "Building..."
go build -o "${OUTPUT}" "./${PROJECT_DIR}"

# Clean up
rm -f "${TEMP_FILE}" "${PROJECT_DIR}/main_generated.go"

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Build successful!"
    echo ""
    echo "Run the embedded executable:"
    echo "  ./${OUTPUT}"
    echo ""
    echo "The executable contains all assets and can be distributed standalone."
    
    # Show file size
    SIZE=$(ls -lh "${OUTPUT}" | awk '{print $5}')
    echo "Executable size: ${SIZE}"
else
    echo ""
    echo "✗ Build failed"
    rm -f "${TEMP_FILE}" "${PROJECT_DIR}/main_generated.go"
    exit 1
fi
