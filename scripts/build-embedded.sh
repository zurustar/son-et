#!/bin/bash

# Build script for creating embedded executables
# Usage: 
#   Single title:   ./scripts/build-embedded.sh <project_dir> [output_name]
#   Multiple titles: ./scripts/build-embedded.sh <project_dir1> <project_dir2> ... [output_name]

set -e

if [ -z "$1" ]; then
    echo "Usage:"
    echo "  Single title:   $0 <project_dir> [output_name]"
    echo "  Multiple titles: $0 <project_dir1> <project_dir2> ... [output_name]"
    echo ""
    echo "Examples:"
    echo "  $0 samples/kuma2"
    echo "  $0 samples/kuma2 samples/ftile400"
    echo "  $0 samples/kuma2 samples/ftile400 my-filly-collection"
    exit 1
fi

# Collect all project directories (all args except possibly the last one)
PROJECT_DIRS=()
OUTPUT_NAME=""

for arg in "$@"; do
    if [ -d "$arg" ]; then
        PROJECT_DIRS+=("$arg")
    else
        # If not a directory, assume it's the output name
        OUTPUT_NAME="$arg"
    fi
done

if [ ${#PROJECT_DIRS[@]} -eq 0 ]; then
    echo "Error: No valid project directories provided"
    exit 1
fi

# Default output name
if [ -z "$OUTPUT_NAME" ]; then
    if [ ${#PROJECT_DIRS[@]} -eq 1 ]; then
        OUTPUT_NAME="$(basename ${PROJECT_DIRS[0]})"
    else
        OUTPUT_NAME="filly-titles"
    fi
fi

echo "Building embedded executable with ${#PROJECT_DIRS[@]} title(s)"
for dir in "${PROJECT_DIRS[@]}"; do
    echo "  - $dir"
done
echo "Output name: $OUTPUT_NAME"

# Step 1: Copy project files to cmd/son-et/titles directory
echo ""
echo "Step 1: Preparing embedded files..."
TITLES_DIR="cmd/son-et/titles"

# Backup existing titles
if [ -d "$TITLES_DIR" ] && [ "$(ls -A $TITLES_DIR 2>/dev/null | grep -v '.gitkeep\|README.md')" ]; then
    echo "Backing up existing titles..."
    BACKUP_DIR=".titles_backup_$(date +%s)"
    mkdir -p "$BACKUP_DIR"
    for item in "$TITLES_DIR"/*; do
        if [ "$(basename $item)" != ".gitkeep" ] && [ "$(basename $item)" != "README.md" ]; then
            mv "$item" "$BACKUP_DIR/"
        fi
    done
    echo "Existing titles backed up to: $BACKUP_DIR"
fi

# Copy all projects to titles directory
for PROJECT_DIR in "${PROJECT_DIRS[@]}"; do
    TARGET_DIR="$TITLES_DIR/$(basename $PROJECT_DIR)"
    echo "Copying $(basename $PROJECT_DIR) to $TARGET_DIR..."
    rm -rf "$TARGET_DIR"
    mkdir -p "$TARGET_DIR"
    cp -r "$PROJECT_DIR"/* "$TARGET_DIR/"
done

# Step 2: Build the executable
echo "Step 2: Building executable..."
mkdir -p bin
go build -o "bin/$OUTPUT_NAME" ./cmd/son-et

# Step 3: Clean up
echo "Step 3: Cleaning up..."
for PROJECT_DIR in "${PROJECT_DIRS[@]}"; do
    TARGET_DIR="$TITLES_DIR/$(basename $PROJECT_DIR)"
    rm -rf "$TARGET_DIR"
done

# Restore backup if exists
if [ -n "$BACKUP_DIR" ] && [ -d "$BACKUP_DIR" ]; then
    echo "Restoring backed up titles..."
    mv "$BACKUP_DIR"/* "$TITLES_DIR/"
    rmdir "$BACKUP_DIR"
fi

echo ""
echo "✓ Success! Embedded executable created: bin/$OUTPUT_NAME"
echo ""
echo "To run:"
echo "  ./bin/$OUTPUT_NAME"
echo ""
echo "To run in headless mode:"
echo "  ./bin/$OUTPUT_NAME --headless --timeout 5"

