# Simple Test Fixture

Some text before any blocks.

```neuroscript
# id: simple-ns-block
# version: 1.0
DEFINE PROCEDURE Simple()
  EMIT "Hello from simple NS"
END
```

Just plain text in between.

```python
# id: simple-py-block
# version: 0.1
print("Hello from simple Python")
```

```text
# id: simple-empty-block
```

```text
# id: simple-comment-block
# This is a comment inside.
-- So is this.

# Even with a blank line.
```

Some text after all blocks.