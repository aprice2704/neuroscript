# Integration Guide Addendum: Managing Persistent Capsules (Revised)

**Audience:** Host Application Developers (e.g., FDM Team)
**Purpose:** Explains the advanced pattern for creating and persisting custom capsules using a trusted configuration script.

---

### 1. Overview: The Two-Phase State Model

Just like `accounts` and `agentmodels`, custom `capsules` defined in a configuration script must persist for later, unprivileged scripts to use. The NeuroScript API supports this via a two-phase model orchestrated by the host application.

1.  **Phase 1 (Configuration):** Your application creates a `CapsuleRegistry` object in its own memory. It then creates a short-lived, trusted interpreter and gives it **write access** to that registry. It runs a script that calls the privileged `tool.capsule.Add` to populate the registry.
2.  **Phase 2 (Runtime):** For all subsequent operations, your application creates new, untrusted interpreters. It gives them **read-only access** to the *same* capsule registry that was populated in Phase 1.

The host application is responsible for creating and holding onto the capsule registry between these two phases.

---

### 2. API for Host-Managed Capsules

To enable this pattern, the `api` package exposes two key functions:

* **`api.WithCapsuleAdminRegistry(registry *api.CapsuleRegistry) api.Option`**: A trusted option used only for configuration interpreters. It injects a *writable* registry that the `tool.capsule.Add` tool can modify.
* **`api.WithCapsuleRegistry(registry *api.CapsuleRegistry) api.Option`**: A standard option that adds a registry as a *read-only* layer to the interpreter's capsule store.

---

### 3. Step-by-Step Implementation

Here is the complete workflow for a host application to manage persistent, script-defined capsules.

### Phase 1: Run the Configuration Script

```go
package main

import (
    "context"
    "fmt"
    "github.com/aprice2704/neuroscript/pkg/api"
)

func main() {
    // --- Phase 1: Configuration ---

    // 1. Host application creates and holds the live registry.
    liveCapsuleRegistry := api.NewCapsuleRegistry()

    // 2. Define the trusted script that will add a new capsule.
    configScript := `
command
    must tool.capsule.Add({
        "name": "capsule/fdm-prompt",
        "version": "1",
        "content": "This is a custom prompt persisted from config."
    })
endcommand
`
    // 3. Create a privileged policy for the config interpreter.
    allowedTools := []string{"tool.capsule.Add"}
    requiredGrants := []api.Capability{
        api.NewCapability(api.ResCapsule, api.VerbWrite, "*"),
    }

    // 4. Create a special config interpreter, injecting the LIVE registry.
    configInterp := api.NewConfigInterpreter(
        allowedTools,
        requiredGrants,
        api.WithCapsuleAdminRegistry(liveCapsuleRegistry), // <-- Give it write access
    )

    // 5. Run the script to populate the liveCapsuleRegistry.
    tree, _ := api.Parse([]byte(configScript), api.ParseSkipComments)
    _, err := api.ExecWithInterpreter(context.Background(), configInterp, tree)
    if err != nil {
        panic(err)
    }

    // The configInterpreter can now be discarded. The populated
    // liveCapsuleRegistry persists in the host application.

    // --- Phase 2: Runtime ---

    // 1. Define a normal, unprivileged script that reads the capsule.
    runtimeScript := `
func main(returns string) means
    set my_cap = tool.capsule.GetLatest("capsule/fdm-prompt")
    return my_cap["content"]
endfunc
`
    // 2. Create a standard, unprivileged interpreter.
    runtimeInterp := api.New(
        // 3. Add the populated registry as a new, read-only layer.
        api.WithCapsuleRegistry(liveCapsuleRegistry),
    )

    // 4. Load and run the script.
    tree, _ = api.Parse([]byte(runtimeScript), api.ParseSkipComments)
    api.ExecWithInterpreter(context.Background(), runtimeInterp, tree)

    result, _ := api.RunProcedure(context.Background(), runtimeInterp, "main")
    unwrapped, _ := api.Unwrap(result)

    fmt.Println(unwrapped)
    // Expected Output: This is a custom prompt persisted from config.
}
```