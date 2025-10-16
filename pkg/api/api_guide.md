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
- **`WithAITranscriptWriter(w io.Writer) Option`**: Provides a writer to which the full transcript of conversations with AI providers will be written, for logging or debugging.
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

   // Optional writer for AI conversation logs.
   AITranscript              io.Writer

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
	var aiTranscript bytes.Buffer

	hostCtx, err := api.NewHostContextBuilder().
		WithLogger(logging.NewSimpleLogger(os.Stderr)).
		WithStdout(&output).
		WithStdin(os.Stdin).
		WithStderr(os.Stderr).
		Build()
	if err != nil {
		panic(err)
	}

	// 2. Instantiate the interpreter with the context and other options.
	interp := api.New(
		api.WithHostContext(hostCtx),
		api.WithAITranscriptWriter(&aiTranscript),
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

# NeuroScript Host Guide: Managing Context & Identity

**Revision:** 2025-Oct-15

**Audience:** Developers integrating the NeuroScript interpreter into host applications.

---

## 1. Guiding Principle: The Host is the Source of Truth

The NeuroScript interpreter is designed as a guest within a host application. The host application is therefore the definitive source for all contextual information. The API provides two distinct channels for the host to provide this context:

1.  **The `HostContext`:** For **stable, long-lived** identity and resources that persist for the entire session.

2.  **The `context.Context`:** For **ephemeral, request-scoped** data that applies only to a single operation.

Properly managing these two channels is essential for security, logging, and enabling advanced tool behavior.

---

## 2. The `HostContext`: Session-Wide Identity & Resources

The `HostContext` is the "umbilical cord" connecting the interpreter to its environment. It is configured **once** during the interpreter's instantiation and is considered immutable thereafter.

* **Purpose:** To provide foundational I/O, services, and the primary identity of the user or system running the script.

* **Configuration:** You must create it using the `api.NewHostContextBuilder()` and pass it to the interpreter via the `api.WithHostContext()` option.

* **Key Fields:**

    * `Logger`, `Stdout`, `Stdin`, `Stderr`: Mandatory I/O and logging channels.

    * `Actor`: An object representing the **long-term identity** (e.g., a user's DID) that is available to all scripts and tools throughout the session.

This context should be used for information that does not change from one command to the next.

#### Example: Instantiation

```go

// In your host application's setup

// 1. Define the long-term identity.

myActor := &MyUser{DID: "did:example:user-1234"}

// 2. Build the HostContext.

hostCtx, _ := api.NewHostContextBuilder().

    WithLogger(myLogger).

    WithStdout(os.Stdout).

    // ... other I/O ...

    WithActor(myActor). // Set the session-wide identity

    Build()

// 3. Create the interpreter, injecting the context.

interp := api.New(api.WithHostContext(hostCtx))

```

---

## 3. The `context.Context`: Per-Operation Data

For information that changes with each execution, such as a transaction ID or the specific agent invoking a tool, the standard Go `context.Context` is used. This context is passed as the first argument to every execution function (`RunProcedure`, `ExecWithInterpreter`, etc.).

* **Purpose:** To provide ephemeral, request-scoped data that is only relevant for the duration of a single, specific script execution.

* **Configuration:** The host application creates and populates this context immediately before calling an interpreter execution method.

* **Best Practice:** The host should define its own custom context key types to avoid collisions.

The interpreter ensures this context is propagated through all internal forks and made available to tools.

---

## 4. Complete Example: Tying It All Together

This example demonstrates how a host provides both types of context and how a tool accesses them.

#### Host Application Code

```go

package main

import (

    "context"

    "[github.com/aprice2704/neuroscript/pkg/api](https://github.com/aprice2704/neuroscript/pkg/api)"

    // ... other imports

)

// 1. Define a custom key for your request-scoped data.

type contextKey string

const transactionIDKey = contextKey("transactionID")

func main() {

    // 2. Set up the long-lived HostContext (as before).

    interp := setupInterpreter() // Assumes a function that returns a configured interpreter

    // 3. For a specific operation, create and populate a context.Context.

    transactionID := "txn-abc-987"

    ctx := context.WithValue(context.Background(), transactionIDKey, transactionID)

    // 4. Pass the context into the execution call.

    // The interpreter will now have access to both the HostContext's Actor

    // and this new context.Context with the transaction ID.

    api.RunProcedure(ctx, interp, "processTransaction")

}

```

#### NeuroScript Tool Code

A tool running inside the `processTransaction` procedure can then access both contexts via the `Runtime` interface it receives.

```go

// A tool's Go implementation

var MyTool = api.ToolImplementation{

    Spec: api.ToolSpec{ Name: "log_context", Group: "host" },

    Func: func(rt api.Runtime, args []any) (any, error) {

        // Access the long-lived Actor from the HostContext

        actorProvider, _ := rt.(interfaces.ActorProvider)

        actor, _ := actorProvider.Actor()

        rt.GetLogger().Info("Executing on behalf of Actor", "did", actor.DID())

        // Access the ephemeral data from the context.Context

        ctxProvider, _ := rt.(api.TurnContextProvider)

        turnCtx := ctxProvider.GetTurnContext()

        txID, _ := turnCtx.Value(transactionIDKey).(string)

        rt.GetLogger().Info("Processing transaction", "id", txID)

        return nil, nil

    },

}

