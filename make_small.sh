#!/bin/bash

# Define your source and destination directories
# It's good practice to define these at the top or pass them as arguments
# For example:

source_dir="/home/aprice/dev/neuroscript"
dest_dir="/home/aprice/dev/neuroscript_sm2"

# Make sure source_dir and dest_dir are set before running this script
if [ -z "$source_dir" ] || [ -z "$dest_dir" ]; then
  echo "Error: source_dir and dest_dir variables must be set."
  exit 1
fi

# Build the index
goindexer -root . -output pkg/codebase-indices


# Remove and recreate the .txt version of the .g4 file
# This part seems specific to your NeuroScript project, ensuring the .txt is an exact copy
rm -f "$source_dir/pkg/core/NeuroScript.g4.txt" # Use -f to avoid error if file doesn't exist
cp "$source_dir/pkg/core/NeuroScript.g4" "$source_dir/pkg/core/NeuroScript.g4.txt"

# Ensure destination directory exists
mkdir -p "$dest_dir"

echo "Syncing from $source_dir to $dest_dir..."

rsync -av \
  --delete \
  --exclude='.git/' \
  --exclude='.vscode/' \
  --exclude='.*' \
  --include='*.go' \
  --include='*.txt' \
  --include='*.md' \
  --include='*.g4' \
  --include='*.json' \
  --prune-empty-dirs \
  --include='*/' \
  --exclude='*' \
  "$source_dir/" "$dest_dir/"

# Optional: Check rsync's exit code
rsync_exit_code=$?
if [ $rsync_exit_code -eq 0 ]; then
  echo "Sync complete successfully."
else
  # The rsync command itself likely printed specific errors above
  echo "Sync may have encountered errors (exit code $rsync_exit_code)."
fi

# Setting ownership and permissions on the destination
# Ensure the user 'aprice' exists on the system where this script is run
if id "aprice" &>/dev/null; then
  chown -R aprice:aprice "$dest_dir/"
else
  echo "Warning: User 'aprice' not found. Skipping chown."
fi
chmod -R 755 "$dest_dir/"

# The rm commands for .git and .vscode are indeed unnecessary
# because --exclude='.*/' and --exclude='*' along with --delete
# should handle their removal if they were copied previously and are not in source.
# However, explicit --exclude rules like --exclude='.git/' are more robust.

echo "Script finished."
