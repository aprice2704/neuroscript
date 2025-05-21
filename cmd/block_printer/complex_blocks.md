# Complex Test Fixture

Block one immediately followed by block two.
:: id: complex-ns-1
```neuroscript
CALL TOOL.DoSomething()
```
:: id: complex-py-adjacent
```python
import os
```

A block with only metadata:
:: id: metadata-only-block
:: version: 1.1
```text
```

A block with no ID:
```javascript
console.log("No ID here");
```

A block with content before the ID (Metadata is now *before* fence, so this comment doesn't interfere):
# Some Go comment before metadata - this is now just regular text before the metadata/fence
:: version: 0.2
:: id: go-block-late-id
```go
package main
import "fmt"
func main(){ fmt.Println("Go!") }
```

A block using double-hyphen comments for metadata (These will be ignored as they aren't '::' and aren't adjacent to fence):
-- id: checklist-hyphen-meta
-- version: 1.0
```neurodata-checklist
- [x] Item A
- [ ] Item B
```

An unclosed block at the very end:
:: id: unclosed-markdown-block
```markdown
This block never ends...