# Integration Guide Addendum: Sharing Function Libraries

---

## 11. Advanced Usage: Sharing Function Libraries

For complex applications, you may have a "library" script containing common helper functions that need to be available to many different, isolated script executions. Instead of parsing and loading this library into every new interpreter, you can load it once into a "base" interpreter and then copy the functions to new interpreters as needed.

This pattern is explicitly supported by the `CopyFunctionsFrom` method.

-   **`dest.CopyFunctionsFrom(source *Interpreter) error`**: Iterates through the function definitions in the `source` interpreter and copies them to the `dest` interpreter.

This method is highly specific:

* ✅ **It copies:** Only the function definitions (`func...endfunc`).
* ❌ **It does not copy:** Event handlers, top-level command blocks, variables, `emit` handlers, or any other runtime state.
* It will return an error if a function being copied already exists in the destination interpreter, preventing accidental overwrites.

### Step-by-Step Example

This example shows how to create a base interpreter with library functions and then use them in a separate, sandboxed runtime interpreter.

```go
package main

import (
    "context"
    "fmt"
    "[github.com/aprice2704/neuroscript/pkg/api](https://github.com/aprice2704/neuroscript/pkg/api)"
)

func main() {
    // --- Phase 1: Create and load the library interpreter ---
    libraryScript := `
        func get_greeting(returns string) means
            return "Hello from the library!"
        endfunc
    `
    libraryInterp := api.New()
    tree, _ := api.Parse([]byte(libraryScript), api.ParseSkipComments)
    api.LoadFromUnit(libraryInterp, &api.LoadedUnit{Tree: tree})

    // --- Phase 2: Create a runtime interpreter and copy functions ---
    runtimeInterp := api.New()
    if err := runtimeInterp.CopyFunctionsFrom(libraryInterp); err != nil {
        panic(err)
    }

    // --- Phase 3: Use the copied function in the runtime interpreter ---
    mainScript := `
        func main(returns string) means
            # This function was not defined in this script, but was copied.
            return get_greeting()
        endfunc
    `
    tree, _ = api.Parse([]byte(mainScript), api.ParseSkipComments)
    runtimeInterp.AppendScript(tree) // Use AppendScript to add the new function

    result, err := api.RunProcedure(context.Background(), runtimeInterp, "main")
    if err != nil {
        panic(err)
    }

    unwrapped, _ := api.Unwrap(result)
    fmt.Println(unwrapped) // Output: Hello from the library!
}
```