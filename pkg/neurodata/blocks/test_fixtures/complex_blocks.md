# Complex Test Fixture

Block one immediately followed by block two.
```neuroscript
# id: complex-ns-1
CALL TOOL.DoSomething()
```
```python
# id: complex-py-adjacent
import os
```

A block with only metadata:
```text
# id: metadata-only-block
# version: 1.1
```

A block with no ID:
```javascript
console.log("No ID here");
```

A block with content before the ID (should still find ID):
```go
# version: 0.2
package main
# id: go-block-late-id
import "fmt"
func main(){ fmt.Println("Go!") }
```

A block using double-hyphen comments for metadata:
```neurodata-checklist
-- id: checklist-hyphen-meta
-- version: 1.0
- [x] Item A
- [ ] Item B
```

An unclosed block at the very end:
```markdown
# id: unclosed-markdown-block
This block never ends...