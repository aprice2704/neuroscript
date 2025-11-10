# NeuroScript Interpreter: Public API Guide

**Audience:** Developers integrating the NeuroScript interpreter into host applications.
**Version:** Reflects architecture post-October 2025 refactor.
**Purpose:** This document outlines the intended public API for instantiating, configuring, and running the NeuroScript interpreter. It focuses on the primary entry points and data structures a developer needs to understand to embed NeuroScript successfully.

---

## 1. The Core Lifecycle: Create, Parse, Load, Run

Interacting with the interpreter follows a clear, four-step lifecycle. The API is designed to configure all external dependencies upfront, parse scripts into an AST, load the AST's definitions, and then execute code.

1.  **Instantiation & Configuration:** Create an `Interpreter` instance using `api.New()` and configure its connection to the outside world via a `HostContext` and other options.
2.  **Parsing Code:** Parse a `.ns` script from bytes into an `AST` using `api.Parse()`.
3.  **Loading Code:** Load script definitions (procedures and event handlers) from the parsed AST into the interpreter using `api.ExecWithInterpreter()` or `api.LoadFromUnit()`.
4.  **Execution:** Run code by invoking a specific procedure with `api.RunProcedure()` or by having `api.ExecWithInterpreter()` run top-level `command` blocks.

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
- **`WithProviderRegistry(...)`**: Injects a registry of AI providers.
- **`WithAccountStore(...)` / `WithAgentModelStore(...)`**: Injects shared, persistent stores for state.

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

   // The ServiceRegistry is the injection point for host
   // services, most notably the SymbolProvider.
   ServiceRegistry           any // (e.g., map[string]any)

   // ... and other callbacks and host-provided APIs.
}
```

**Key Takeaway:** A minimal, functioning interpreter requires calling `New` with at least `WithHostContext`, where the context has `Logger`, `Stdout`, `Stdin`, and `Stderr` populated.

**Host-Level Services:** The `ServiceRegistry` field is the generic mechanism for injecting host services. Its primary use is to provide host-defined global symbols (functions, constants) by injecting an implementation of `api.SymbolProvider` (see Section 5.1).

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
	// Load definitions by executing the tree. This loads funcs/events.
	// Since this script has no 'command' blocks, nothing else is run.
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
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
- **Forked Interpreter (Child):** Whenever a procedure is called (via `RunProcedure`) or an event handler is triggered, the interpreter creates a lightweight, temporary *fork*.

A **fork** inherits read-only access to everything from the root (procedures, globals, **host-provided symbols**, etc.) but has its own separate, isolated memory for local variables. **Crucially, a fork cannot modify the state of its parent.** This means a function can't change a global variable, and local variables set inside a function disappear after it returns. This is the core of NeuroScript's safety model.

---

## 4. Loading and Executing Scripts

Once configured, you can load your definitions and execute scripts. It is critical to understand the distinction between **definitions** (`func`/`on event`) and **imperative code** (`command` blocks).

**Rule:** A single script or file should contain *either* definitions *or* command blocks, but **never both**.

### `ExecWithInterpreter(ctx context.Context, interp *Interpreter, tree *Tree) (Value, error)`

This static helper function is the primary method for both loading definitions and executing imperative code. It performs two actions in order:
1.  **Loads Definitions:** It walks the AST and loads all `func` and `on event` blocks into the interpreter's state.
2.  **Runs Commands:** It executes any top-level `command` blocks found in the script.

If you pass a tree that *only* contains `func` definitions (like a library), this function acts as the "load" step.

> **Symbol Conflict Rule:** The load will **fail** if the script attempts to define a `func` or `on event` with the same name as a symbol already provided by the host's `SymbolProvider`. This "No Override" rule ensures host-level functions are deterministic and secure.

### `LoadFromUnit(interp *Interpreter, unit *LoadedUnit) error`

This static helper function is used for securely-loaded scripts. It loads definitions from a `LoadedUnit` (the output of `api.Load()`), but it does *not* execute `command` blocks. This is the recommended path for loading verified code.

> **Symbol Conflict Rule:** This function is also subject to the "No Override" rule described above.

### `RunProcedure(ctx context.Context, interp *Interpreter, name string, args ...any) (Value, error)`

This static helper function is the primary method for executing a specific piece of loaded code. It invokes a named procedure (a `func` block) and passes the provided Go-native arguments to it. It returns the value from the procedure's `return` statement.

### `ExecScript(...)`

While available, this is primarily a helper for testing and simple use cases. It combines parsing, loading, and executing into a single call. For production use, it's recommended to handle parsing as a separate step so you can cache the AST.

---

## 5. Advanced Interaction & State Management

The API also provides access to the interpreter's various registries and state stores. These are often used for setup, or for tools that need to interact with the interpreter's core.

### 5.1 Host-Provided Symbols (SymbolProvider)

For providing foundational, host-defined functions and constants (the "Read Path" described in `ns_globals.md`), the host application must implement and inject a `SymbolProvider`.

1.  **Implement the Interface:** Create a Go struct that implements the `api.SymbolProvider` interface (which is an alias for `interfaces.SymbolProvider`).
2.  **Inject via ServiceRegistry:** Place an instance of your provider into a map using `api.SymbolProviderKey` as the key.
3.  **Build HostContext:** Pass this map to `NewHostContextBuilder().WithServiceRegistry(...)`.

```go
// In your Go application setup:

// 1. Define your provider
type MySymbolProvider struct {
    // ... (e.g., a database connection)
}

// Implement the interface
func (p *MySymbolProvider) GetProcedure(name string) (any, bool) {
    if name == "my_host_func" {
        // Return a parsed *ast.Procedure node
        // (This is typically done in a "Config Context" boot process)
        return myParsedProc, true
    }
    return nil, false
}
func (p *MySymbolPlayer) GetGlobalConstant(name string) (any, bool) {
    if name == "MY_HOST_CONST" {
        return lang.StringValue{Value: "hello from host"}, true
    }
    return nil, false
}
// ... (implement other List* and Get* methods)

// 2. Create the provider and registry map
myProvider := &MySymbolProvider{}
serviceReg := map[string]any{
    api.SymbolProviderKey: myProvider,
}

// 3. Build HostContext and create interpreter
hostCtx, _ := api.NewHostContextBuilder().
    WithLogger(myLogger).
    WithStdout(os.Stdout).
    WithStdin(os.Stdin).
    WithStderr(os.Stderr).
    WithServiceRegistry(serviceReg). // Inject the provider
    Build()

interp := api.New(api.WithHostContext(hostCtx))

// Now, any script run in 'interp' can call 'my_host_func()'
// and access 'MY_HOST_CONST' as if they were built-in.
```

### 5.2 Custom Tool Example

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

// Register the tool on the interpreter's registry
interp.ToolRegistry().RegisterTool(myTool)

// Now, in a NeuroScript file:
/*
command
   emit tool.strings.FormatGreeting("developer")
endcommand
*/
```

### 5.3 AI & Provider Management

The host application registers AI providers by injecting a populated registry *at creation time*.

- **`WithProviderRegistry(registry *provider.Registry) Option`**: Injects a pre-populated `ProviderRegistry` into the interpreter.

```go
// In Go setup code:
myRegistry := api.NewProviderRegistry()

// You must get an admin handle to the registry to mutate it.
// This typically requires a trusted policy.
providerAdmin := provider.NewAdmin(myRegistry, myConfigPolicy) 
providerAdmin.Register("my-provider", myProviderImpl)

// At interpreter creation:
interp := api.New(
    api.WithHostContext(hostCtx),
    api.WithProviderRegistry(myRegistry),
)
```

- **`RegisterAgentModel(name string, config map[string]any) error`**: This method (available via embedding) provides administrative access to the store of "AgentModels," which are configurations that bundle a provider, model name, and other settings.

### 5.4 Account Management

- **`WithAccountStore(store *account.Store)`**: Allows the host to replace the default in-memory account store with a persistent one at creation time.

### 5.5 Variable Management

- **`WithGlobals(globals map[string]interface{})`**: Sets a map of variables in the interpreter's **global scope** at creation time. This is the primary way to inject data into a script from the host.

### 5.6 Host-Level Orchestration

For advanced integrations, such as implementing a custom 'ask' loop (like FDM's AEIOU hook), the API provides low-level functions:

- **`api.CheckScriptTools(tree *Tree, interp Runtime) error`**: Verifies that all tools required by an AST are registered in the interpreter.
- **`api.ExecuteSandboxedAST(...)`**: Runs a parsed AST in a secure, sandboxed fork of an interpreter and captures all `emit` and `whisper` output.