# Simple Test Fixture

Some text before any blocks.

:: id: simple-ns-block
:: version: 1.0
```neuroscript
DEFINE PROCEDURE Simple()
  EMIT "Hello from simple NS"
END
```

Just plain text in between.

:: id: simple-py-block
:: version: 0.1
```python
print("Hello from simple Python")
```

:: id: simple-empty-block
```text
```

:: id: simple-comment-block
```text
# This is a comment inside.
-- So is this.

# Even with a blank line.
```

Some text after all blocks.