#!/bin/bash

source_dir="/home/aprice/dev/neuroscript"
dest_dir="/home/aprice/dev/neuroscript_sm2"

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
  --include='*.g4' \
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

chown -R aprice:aprice "$dest_dir/"
chmod -R 755 "$dest_dir/"

# rm commands should be unnecessary now
# rm -r "$dest_dir/.git/"
# rm -r "$dest_dir/.vscode/"
