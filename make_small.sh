source_dir="neuroscript"
dest_dir="neuroscript_small"
rsync -av --include='*.go' --include='*.txt' --include='*.md' --include='*/' --exclude='*' --exclude='.*/**' "$source_dir/" "$dest_dir/"