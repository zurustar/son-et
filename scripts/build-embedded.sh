#!/bin/bash

# Build script for creating embedded executables
# Usage: 
#   Single title:   ./scripts/build-embedded.sh <project_dir_or_entry_file> [output_name]
#   Multiple titles: ./scripts/build-embedded.sh <path1> <path2> ... [output_name]
#
# Entry point specification:
#   - Directory: samples/kuma2 (auto-detect main function)
#   - TFY file:  samples/sab2/TOKYO.TFY (explicit entry point)

set -e

if [ -z "$1" ]; then
    echo "Usage:"
    echo "  Single title:   $0 <project_dir_or_entry_file> [output_name]"
    echo "  Multiple titles: $0 <path1> <path2> ... [output_name]"
    echo ""
    echo "Examples:"
    echo "  $0 samples/kuma2                          # Directory (auto-detect main)"
    echo "  $0 samples/sab2/TOKYO.TFY                 # Entry file (explicit)"
    echo "  $0 samples/kuma2 samples/sab2/TOKYO.TFY   # Multiple titles"
    echo "  $0 samples/kuma2 my-app                   # With custom output name"
    exit 1
fi

# Structure to hold title info: directory and optional entry file
declare -a TITLE_DIRS
declare -a TITLE_ENTRIES
OUTPUT_NAME=""

# Parse arguments
for arg in "$@"; do
    if [ -d "$arg" ]; then
        # Directory specified
        TITLE_DIRS+=("$arg")
        TITLE_ENTRIES+=("")  # No explicit entry file
    elif [ -f "$arg" ]; then
        # File specified - check if it's a TFY file
        if [[ "${arg,,}" == *.tfy ]]; then
            # Extract directory and filename
            dir=$(dirname "$arg")
            entry=$(basename "$arg")
            TITLE_DIRS+=("$dir")
            TITLE_ENTRIES+=("$entry")
        else
            echo "Error: File must be a .TFY file: $arg"
            exit 1
        fi
    else
        # Not a directory or file - assume it's the output name
        if [ -n "$OUTPUT_NAME" ]; then
            echo "Error: Multiple output names specified or invalid path: $arg"
            exit 1
        fi
        OUTPUT_NAME="$arg"
    fi
done

if [ ${#TITLE_DIRS[@]} -eq 0 ]; then
    echo "Error: No valid project directories or entry files provided"
    exit 1
fi

# Default output name
if [ -z "$OUTPUT_NAME" ]; then
    if [ ${#TITLE_DIRS[@]} -eq 1 ]; then
        OUTPUT_NAME="$(basename ${TITLE_DIRS[0]})"
    else
        OUTPUT_NAME="filly-titles"
    fi
fi

echo "Building embedded executable with ${#TITLE_DIRS[@]} title(s)"
for i in "${!TITLE_DIRS[@]}"; do
    dir="${TITLE_DIRS[$i]}"
    entry="${TITLE_ENTRIES[$i]}"
    if [ -n "$entry" ]; then
        echo "  - $dir (entry: $entry)"
    else
        echo "  - $dir (auto-detect main)"
    fi
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

# Copy all projects to titles directory and create title.json if needed
for i in "${!TITLE_DIRS[@]}"; do
    dir="${TITLE_DIRS[$i]}"
    entry="${TITLE_ENTRIES[$i]}"
    TARGET_DIR="$TITLES_DIR/$(basename $dir)"
    
    echo "Copying $(basename $dir) to $TARGET_DIR..."
    rm -rf "$TARGET_DIR"
    mkdir -p "$TARGET_DIR"
    cp -r "$dir"/* "$TARGET_DIR/"
    
    # Create title.json if entry file is specified
    if [ -n "$entry" ]; then
        echo "Creating title.json with entry file: $entry"
        cat > "$TARGET_DIR/title.json" << EOF
{
    "entryFile": "$entry"
}
EOF
    fi
done

# Step 2: Build the executable
echo ""
echo "Step 2: Building executable..."
mkdir -p bin
go build -o "bin/$OUTPUT_NAME" ./cmd/son-et

# Step 3: Clean up
echo ""
echo "Step 3: Cleaning up..."
for i in "${!TITLE_DIRS[@]}"; do
    dir="${TITLE_DIRS[$i]}"
    TARGET_DIR="$TITLES_DIR/$(basename $dir)"
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
