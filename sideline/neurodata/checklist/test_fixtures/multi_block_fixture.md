:: version: 1.0.0
:: author: Test Suite
:: description: Fixture with multiple blocks including checklists.

# Document Title

This document contains multiple fenced code blocks to test extraction.

## Section 1: NeuroScript Example

Here is a NeuroScript block:

```neuroscript
:: id: setup-proc
:: version: 0.1
DEFINE PROCEDURE InitialSetup(param1)
COMMENT:
    PURPOSE: Does initial setup.
    INPUTS: - param1: Some value
    OUTPUT: Status string
    ALGORITHM: Emit message, return OK.
ENDCOMMENT
    EMIT "Performing setup with " + param1
    RETURN "Setup OK"
END
```

## Section 2: First Checklist

This section has the first checklist.

```neurodata-checklist
:: id: checklist-alpha
:: status: pending
# Phase 1 Tasks
- [ ] Task A
- [x] Task B
  - [-] Subtask B.1 (Partial)
# Phase 2 Tasks
- [?] Task C (Needs Info)
```

Some regular Markdown text between the blocks.

## Section 3: Second Checklist and Other Blocks

Another checklist follows immediately.

```neurodata-checklist
:: id: checklist-beta
:: priority: high
- [[]] Automatic Item
  - [x] Child Done 1
  - [ ] Child Pending 2
- [!] Blocked Item (Special)
```

Then a block without a language ID:

```
This content has no language specified.
It should be extracted with an empty language ID.
```

Finally, a Python block.

```python
:: id: simple-py
# A simple python script block
import sys

def main():
  print("Hello from multi-block fixture!")

if __name__ == "__main__":
  main()
```

End of document.