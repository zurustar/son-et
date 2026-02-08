#!/bin/bash

# Build script for creating embedded executables
# Usage: 
#   Single title:   ./scripts/build-embedded.sh [--soundfont <sf2_path>] <project_dir_or_entry_file> [output_name]
#   Multiple titles: ./scripts/build-embedded.sh [--soundfont <sf2_path>] <path1> <path2> ... [output_name]
#
# Options:
#   --soundfont <path>  Path to SoundFont (.sf2) file to embed
#
# Entry point specification:
#   - Directory: samples/kuma2 (auto-detect main function)
#   - TFY file:  samples/sab2/TOKYO.TFY (explicit entry point)

set -e

# Structure to hold title info: directory and optional entry file
declare -a TITLE_DIRS
declare -a TITLE_ENTRIES
OUTPUT_NAME=""
SOUNDFONT_PATH=""

# Print usage
print_usage() {
    echo "Usage:"
    echo "  Single title:   $0 [--soundfont <sf2_path>] <project_dir_or_entry_file> [output_name]"
    echo "  Multiple titles: $0 [--soundfont <sf2_path>] <path1> <path2> ... [output_name]"
    echo ""
    echo "Options:"
    echo "  --soundfont <path>  Path to SoundFont (.sf2) file to embed"
    echo ""
    echo "Examples:"
    echo "  $0 samples/kuma2                                    # Directory (auto-detect main)"
    echo "  $0 samples/sab2/TOKYO.TFY                           # Entry file (explicit)"
    echo "  $0 samples/kuma2 samples/sab2/TOKYO.TFY             # Multiple titles"
    echo "  $0 samples/kuma2 my-app                             # With custom output name"
    echo "  $0 --soundfont GeneralUser-GS.sf2 samples/kuma2     # With embedded SoundFont"
    echo "  $0 --soundfont GeneralUser-GS.sf2 samples/kuma2 samples/sab2/TOKYO.TFY my-app"
}

if [ -z "$1" ]; then
    print_usage
    exit 1
fi

# Parse arguments
while [ $# -gt 0 ]; do
    case "$1" in
        --soundfont)
            if [ -z "$2" ]; then
                echo "Error: --soundfont requires a path argument"
                exit 1
            fi
            SOUNDFONT_PATH="$2"
            shift 2
            ;;
        --help|-h)
            print_usage
            exit 0
            ;;
        *)
            if [ -d "$1" ]; then
                # Directory specified
                TITLE_DIRS+=("$1")
                TITLE_ENTRIES+=("")  # No explicit entry file
            elif [ -f "$1" ]; then
                # File specified - check if it's a TFY file
                arg_lower=$(echo "$1" | tr '[:upper:]' '[:lower:]')
                if [[ "$arg_lower" == *.tfy ]]; then
                    # Extract directory and filename
                    dir=$(dirname "$1")
                    entry=$(basename "$1")
                    TITLE_DIRS+=("$dir")
                    TITLE_ENTRIES+=("$entry")
                else
                    echo "Error: File must be a .TFY file: $1"
                    exit 1
                fi
            else
                # Not a directory or file - assume it's the output name
                if [ -n "$OUTPUT_NAME" ]; then
                    echo "Error: Multiple output names specified or invalid path: $1"
                    exit 1
                fi
                OUTPUT_NAME="$1"
            fi
            shift
            ;;
    esac
done

if [ ${#TITLE_DIRS[@]} -eq 0 ]; then
    echo "Error: No valid project directories or entry files provided"
    exit 1
fi

# Validate SoundFont file if specified
if [ -n "$SOUNDFONT_PATH" ]; then
    if [ ! -f "$SOUNDFONT_PATH" ]; then
        echo "Error: SoundFont file not found: $SOUNDFONT_PATH"
        exit 1
    fi
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
if [ -n "$SOUNDFONT_PATH" ]; then
    echo "SoundFont: $SOUNDFONT_PATH"
fi
echo "Output name: $OUTPUT_NAME"

# Step 1: Copy project files to cmd/son-et/titles directory
echo ""
echo "Step 1: Preparing embedded files..."
TITLES_DIR="cmd/son-et/titles"
SOUNDFONTS_DIR="cmd/son-et/soundfonts"

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

# Backup existing soundfonts
SOUNDFONT_BACKUP_DIR=""
if [ -n "$SOUNDFONT_PATH" ]; then
    if [ -d "$SOUNDFONTS_DIR" ] && [ "$(ls -A $SOUNDFONTS_DIR 2>/dev/null | grep -v '.gitkeep\|README.md')" ]; then
        echo "Backing up existing soundfonts..."
        SOUNDFONT_BACKUP_DIR=".soundfonts_backup_$(date +%s)"
        mkdir -p "$SOUNDFONT_BACKUP_DIR"
        for item in "$SOUNDFONTS_DIR"/*; do
            if [ "$(basename $item)" != ".gitkeep" ] && [ "$(basename $item)" != "README.md" ]; then
                mv "$item" "$SOUNDFONT_BACKUP_DIR/"
            fi
        done
        echo "Existing soundfonts backed up to: $SOUNDFONT_BACKUP_DIR"
    fi
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

# Copy SoundFont file if specified
if [ -n "$SOUNDFONT_PATH" ]; then
    echo "Copying SoundFont to $SOUNDFONTS_DIR..."
    mkdir -p "$SOUNDFONTS_DIR"
    cp "$SOUNDFONT_PATH" "$SOUNDFONTS_DIR/GeneralUser-GS.sf2"
fi

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

# Clean up SoundFont
if [ -n "$SOUNDFONT_PATH" ]; then
    rm -f "$SOUNDFONTS_DIR/GeneralUser-GS.sf2"
fi

# Restore backup if exists
if [ -n "$BACKUP_DIR" ] && [ -d "$BACKUP_DIR" ]; then
    echo "Restoring backed up titles..."
    mv "$BACKUP_DIR"/* "$TITLES_DIR/"
    rmdir "$BACKUP_DIR"
fi

# Restore soundfont backup if exists
if [ -n "$SOUNDFONT_BACKUP_DIR" ] && [ -d "$SOUNDFONT_BACKUP_DIR" ]; then
    echo "Restoring backed up soundfonts..."
    mv "$SOUNDFONT_BACKUP_DIR"/* "$SOUNDFONTS_DIR/"
    rmdir "$SOUNDFONT_BACKUP_DIR"
fi

echo ""
echo "✓ Success! Embedded executable created: bin/$OUTPUT_NAME"
if [ -n "$SOUNDFONT_PATH" ]; then
    echo "  - SoundFont embedded: $(basename $SOUNDFONT_PATH)"
fi
echo ""
echo "To run:"
echo "  ./bin/$OUTPUT_NAME"
echo ""
echo "To run in headless mode:"
echo "  ./bin/$OUTPUT_NAME --headless --timeout 5"
