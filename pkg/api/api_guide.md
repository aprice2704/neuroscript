# NeuroScript Interpreter: Public API Guide

**Audience:** Developers integrating the NeuroScript interpreter into host applications.
**Version:** Reflects architecture post-October 2025 refactor.
**Purpose:** This document outlines the intended public API for instantiating, configuring, and running the NeuroScript interpreter. It focuses on the primary entry points and data structures a developer needs to understand to embed NeuroScript successfully.

---

## 1. The Core Lifecycle: Create, Configure, Load, Run

Interacting with the interpreter follows a clear, four-step lifecycle. The API is designed to configure all external dependencies upfront, load script logic, and then execute it.

1.  **Instantiation & Configuration:** Create an `Interpreter` instance using `New()` and configure its connection to the outside world via a `HostContext` and other options.
2.  **Loading Code:** Load script definitions (procedures and event handlers) from an AST into the interpreter using `Load()`. This populates the interpreter's library of callable code.
3.  **Appending Code (Optional):** Merge additional script definitions into the interpreter's state using `AppendScript()`. This is useful for loading libraries or modules without overwriting the main script.
4.  **Execution:** Run code, either by invoking a specific procedure with `RunProcedure()` or by executing a script containing top-level `command` blocks with `ExecWithInterpreter()`.

---

## 2. Instantiation and Configuration

This is the most critical phase. An interpreter cannot function without being properly configured with its host dependencies.

### `New(opts ...Option) *Interpreter`

This is the sole entry point for creating a new interpreter instance. It returns a **root interpreter**. It takes a variable number of `Option` functions that configure the instance.

### `Option` Functions

Options are functions that modify the interpreter's configuration during creation. The most important one is `WithHostContext`.

- **`WithHostContext(hc *HostContext) Option`**: **(Mandatory)** This is the primary and essential option. It provides the interpreter with its "umbilical cord" to the host application, containing all I/O and callback functions.
- **`WithExecPolicy(policy *policy.ExecPolicy) Option`**: Applies a security policy that governs what the script is allowed to do (e.g., which tools it can call, what capabilities it has). If not provided, it defaults to a restrictive policy.
- **`WithSandboxDir(path string) Option`**: Sets the root directory for all file-based operations, preventing the script from accessing the broader filesystem.
- **`WithGlobals(globals map[string]interface{}) Option`**: Injects a map of Go values as initial global variables into the interpreter's state.
- **`WithoutStandardTools() Option`**: Prevents the automatic registration of the standard tool library, useful for creating highly restricted or specialized runtimes.
- **`WithCapsuleRegistry(...)` / `WithCapsuleAdminRegistry(...)`**: Configures registries for managing packaged scripts ("capsules").

### The `HostContext` Struct

This struct is the centerpiece of the configuration API. It's built using the `NewHostContextBuilder()` fluent API. It is passed by reference and is considered immutable after the interpreter is created.

```go
// Simplified representation of the builder's purpose
type HostContext struct {
   // A structured logger is mandatory.
   Logger                    interfaces.Logger

   // Standard I/O streams are mandatory.
   Stdout                    io.Writer
   Stdin                     io.Reader
   Stderr                    io.Writer

   // Callback for the 'emit' statement.
   EmitFunc                  func(api.Value)

   // ... and other callbacks and host-provided APIs.
}
```

**Key Takeaway:** A minimal, functioning interpreter requires calling `New` with at least `WithHostContext`, where the context has `Logger`, `Stdout`, `Stdin`, and `Stderr` populated.

### Example: Full Setup

```go
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"[github.com/aprice2704/neuroscript/pkg/api](https://github.com/aprice2704/neuroscript/pkg/api)"
	"[github.com/aprice2704/neuroscript/pkg/logging](https://github.com/aprice2704/neuroscript/pkg/logging)"
)

func main() {
	// 1. Create a HostContext with mandatory I/O and a logger.
	var output bytes.Buffer
	hostCtx, err := api.NewHostContextBuilder().
		WithLogger(logging.NewSimpleLogger(os.Stderr)).
		WithStdout(&output).
		WithStdin(os.Stdin).
		WithStderr(os.Stderr).
		Build()
	if err != nil {
		panic(err)
	}

	// 2. Instantiate the interpreter with the context and initial globals.
	interp := api.New(
		api.WithHostContext(hostCtx),
		api.WithGlobals(map[string]interface{}{
			"host_version": "1.0.0",
		}),
	)

	// 3. Parse and load a script containing definitions.
	libScript := `func greet(needs name) means return "Hello, " + name endfunc`
	tree, err := api.Parse([]byte(libScript), api.ParseSkipComments)
   if err != nil {
       panic(err)
   }
	if err := interp.Load(tree); err != nil {
       panic(err)
   }

	// 4. Run a procedure from the loaded library.
	result, err := api.RunProcedure(context.Background(), interp, "greet", "World")
	if err != nil {
		panic(err)
	}

   unwrapped, _ := api.Unwrap(result)
	fmt.Println("Result from Run:", unwrapped) // "Hello, World"
}
```

---

## 3. Execution Model: Root vs. Forks

The interpreter uses a "root vs. fork" model to ensure sandboxing and prevent unintended side effects.

- **Root Interpreter:** The instance you create with `New()`. It holds all persistent state, including loaded procedures, global variables, agent models, and accounts.
- **Forked Interpreter (Child):** Whenever a procedure is called (via `Run`) or an event handler is triggered, the interpreter creates a lightweight, temporary *fork*.

A **fork** inherits read-only access to everything from the root (procedures, globals, etc.) but has its own separate, isolated memory for local variables. **Crucially, a fork cannot modify the state of its parent.** This means a function can't change a global variable, and local variables set inside a function disappear after it returns. This is the core of NeuroScript's safety model.

---

## 4. Loading and Executing Scripts

Once configured, you can load your definitions and execute scripts. It is critical to understand the distinction between **definitions** (`func`/`on event`) and **imperative code** (`command` blocks).

**Rule:** A single script or file should contain *either* definitions *or* command blocks, but **never both**.

### `Load(tree *interfaces.Tree) error`

This method loads definitions (`func` and `on event` blocks) from a parsed AST into the interpreter. **This is a destructive action**: it completely replaces any procedures and event handlers that were previously loaded. It does not execute any code.

### `AppendScript(tree *interfaces.Tree) error`

This method merges new definitions from an AST into the *existing* state of the interpreter. It will add new procedures and event handlers but will return an error if you try to define a procedure that already exists. This is ideal for loading modular libraries of functions. It does not execute any code.

### `RunProcedure(ctx context.Context, interp *Interpreter, name string, args ...any) (Value, error)`

This is the primary method for executing a specific piece of loaded code. It invokes a named procedure (a `func` block) and passes the provided Go-native arguments to it. It returns the value from the procedure's `return` statement.

### `ExecWithInterpreter(ctx context.Context, interp *Interpreter, tree *Tree) (Value, error)`

This method is used to execute a script consisting of top-level `command` blocks. `command` blocks contain imperative code that runs immediately. They are sandboxed and cannot modify the state of the parent interpreter. This is the entry point for running the "main" logic of a script file.

### `ExecScript(...)`

While available, this is primarily a helper for testing and simple use cases. It combines parsing, loading, and executing into a single call. For production use, it's recommended to handle parsing as a separate step so you can cache the AST.

---

## 5. Advanced Interaction & State Management

The API also provides access to the interpreter's various registries and state stores. These are often used for setup, or for tools that need to interact with the interpreter's core.

### Custom Tool Example

You can register your own Go functions as tools that scripts can call.

```go
// In your Go application setup:
myTool := api.ToolImplementation{
	Spec: api.ToolSpec{
		Name:  "FormatGreeting",
		Group: "strings",
		Args:  []api.ArgSpec{{Name: "name", Type: "string"}},
		ReturnType: "string",
	},
	Func: func(rt api.Runtime, args []any) (any, error) {
		name, _ := args[0].(string)
		return "Hello from a Go tool, " + name + "!", nil
	},
}

// Create interpreter with api.WithTool(myTool)

// Now, in a NeuroScript file:
/*
command
   emit tool.strings.FormatGreeting("developer")
endcommand
*/
```

### AI & Provider Management

- **`RegisterProvider(name string, p AIProvider)`**: Allows the host application to register a concrete AI provider (e.g., a connection to OpenAI, Anthropic, etc.) under a specific name that scripts can reference.
- **`RegisterAgentModel(name string, config map[string]any) error`**: Provides administrative (`Register`, `Update`, `Delete`) access to the store of "AgentModels," which are configurations that bundle a provider, model name, and other settings.

### Account Management

- **`WithAccountStore(store *account.Store)`**: Allows the host to replace the default in-memory account store with a persistent one at creation time.

### Variable Management

- **`WithGlobals(globals map[string]interface{})`**: Sets a map of variables in the interpreter's **global scope** at creation time. This is the primary way to inject data into a script from the host.