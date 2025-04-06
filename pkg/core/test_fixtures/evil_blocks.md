# More Evil Test Fixture

Block starting at the very beginning:
```
# id: block-at-start
Content at start.
```

A block with fences inside its content:
```yaml
# id: fences-inside
key: value
example: |
  ```bash
  echo "This inner fence should be captured."
  ```
another_key: true
```

Block with unusual language tag:
```neuroscript-v1.1
# id: weird-tag
DEFINE PROCEDURE TestWeirdTag()
  EMIT "Tag test"
END
```

An empty block:
```
```

A block with only whitespace:
```
   
	  

```

Block immediately followed by EOF (no trailing newline):
```javascript
# id: block-at-eof
console.log("EOF");
```
```
