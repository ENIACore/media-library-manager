#!/usr/bin/env bash
# Usage: ./copy_tree.sh /source/path /destination/path

SRC="$1"
DEST="$2"

# Check inputs
if [[ -z "$SRC" || -z "$DEST" ]]; then
    echo "Usage: $0 <source_directory> <destination_directory>"
    exit 1
fi

if [[ ! -d "$SRC" ]]; then
    echo "Source directory does not exist: $SRC"
    exit 1
fi

# Create destination root if it doesn't exist
mkdir -p "$DEST"

# Copy directory structure
cd "$SRC" || exit
find . -type d -exec mkdir -p "$DEST/{}" \;

# Create empty files
find . -type f -exec touch "$DEST/{}" \;

echo "Directory structure copied from $SRC to $DEST with empty files."
