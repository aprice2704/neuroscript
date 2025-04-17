#!/bin/bash

source_dir="neuroscript"
dest_dir="neuroscript_small"

# Ensure destination directory exists
mkdir -p "$dest_dir"

echo "Syncing from $source_dir to $dest_dir..."

rsync -av \
  --exclude='.git/' \
  --exclude='.vscode/' \
  --exclude='.*' \
  --include='*.go' \
  --include='*.txt' \
  --include='*.md' \
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

# rm commands should be unnecessary now
# rm -r "$dest_dir/.git/"
# rm -r "$dest_dir/.vscode/"