# Integration Guide Addendum: Injecting Host Context with a Custom Runtime

---

## 12. Advanced Usage: Injecting Host Context via a Custom Runtime

For sophisticated host applications like FDM, there is often a need to make host-level information, such as the identity of the currently authenticated user, available to NeuroScript tools. This allows tools to perform actions on behalf of a specific user while maintaining security and auditability.

The recommended pattern for this is to create a custom Go struct that embeds the `api.Interpreter` but adds host-specific fields. This custom struct, which satisfies the `api.Runtime` interface, can then be injected into the interpreter instance.

This feature is enabled by two key methods:

-   **`interp.SetRuntime(rt api.Runtime)`**: Sets the runtime context that will be passed to all tool functions.
-   **`api.WithRuntime(rt api.Runtime) api.Option`**: A functional option to set the custom runtime during interpreter creation with `api.New()`.

### Step-by-Step Example

This example demonstrates how a host application can define a custom runtime to provide an "actor ID" to a tool.

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"[github.com/aprice2704/neuroscript/pkg/api](https://github.com/aprice2704/neuroscript/pkg/api)"
)

// 1. Define your custom runtime struct.
// It must embed *api.Interpreter to satisfy the api.Runtime interface implicitly.
type AppRuntime struct {
	*api.Interpreter
	ActorID string // Your custom, host-specific data.
}

// 2. Define a tool that requires the custom context.
func WhoAmITool(rt api.Runtime, args []any) (any, error) {
	// Type-assert the runtime to access your custom fields.
	appRT, ok := rt.(*AppRuntime)
	if !ok {
		// This is a critical error; the wrong runtime was passed.
		return nil, errors.New("tool executed with incorrect runtime context")
	}
	// Now the tool can access the host-provided identity.
	return "This tool is being executed by actor: " + appRT.ActorID, nil
}

func main() {
	// 3. Create the interpreter and your custom runtime instance.
	interp := api.New(api.WithStdout(os.Stdout))
	customRT := &AppRuntime{
		Interpreter: interp,
		ActorID:     "user-xyz-123",
	}

	// 4. Inject the custom runtime into the interpreter.
	interp.SetRuntime(customRT)

	// 5. Register your context-aware tool.
	whoamiImpl := api.ToolImplementation{
		Spec: api.ToolSpec{Name: "whoami", Group: "host"},
		Func: WhoAmITool,
	}
	if _, err := interp.ToolRegistry().RegisterTool(whoamiImpl); err != nil {
		panic(err)
	}

	// 6. Execute a script that calls the tool. The interpreter will now
	// automatically pass `customRT` to the tool's function.
	script := `command emit tool.host.whoami() endcommand`
	tree, _ := api.Parse([]byte(script), api.ParseSkipComments)
	_, err := api.ExecWithInterpreter(context.Background(), interp, tree)
	if err != nil {
		panic(err)
	}
}

// Expected output:
// This tool is being executed by actor: user-xyz-123
```

This pattern ensures that host-specific context is cleanly and safely passed through the NeuroScript engine to the tools that require it, without polluting the NeuroScript language itself with host-specific concepts.