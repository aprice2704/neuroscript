# NeuroScript Interpreter: Public API Guide

**Version:** 25

**Audience:** Developers integrating the NeuroScript interpreter into host applications.
**Version:** Reflects architecture post-October 2025 refactor.
**Purpose:** This document outlines the intended public API for instantiating, configuring, and running the NeuroScript interpreter. It focuses on the primary entry points and data structures a developer needs to understand to embed NeuroScript successfully.

---

## 1. The Core Lifecycle: Parse, Load, Run

Interacting with the interpreter follows a clear, three-step lifecycle. The API is designed to configure all external dependencies upfront, parse scripts into an Abstract Syntax Tree (AST), load that AST's definitions, and then execute code.

1.  **Instantiation & Configuration:** Create an `Interpreter` instance using `api.New()` and configure its connection to the outside world via a `HostContext` and other options.
2.  **Parsing & Loading:**
    * Parse a `.ns` script from bytes into an `AST` using `api.Parse()`.
    * For secure, signed scripts, use `api.Canonicalise()` and `api.Load()` to get a verified `LoadedUnit`.
    * Load script definitions (procedures and event handlers) into the interpreter using `api.ExecWithInterpreter()` or `api.LoadFromUnit()`.
3.  **Execution:** Run code by invoking a specific procedure with `api.RunProcedure()` or by having `api.ExecWithInterpreter()` run top-level `command` blocks.

---

## 2. Instantiation and Configuration

This is the most critical phase. An interpreter cannot function without being properly configured with its host dependencies.

### `New(opts ...Option) *Interpreter`

This is the sole entry point for creating a new interpreter instance. It returns a **root interpreter**. It takes a variable number of `Option` functions that configure the instance.

### `Option` Functions

Options are functions that modify the interpreter's configuration during creation. The most important one is `WithHostContext`.

- **`WithHostContext(hc *HostContext) Option`**: **(Mandatory)** This is the primary and essential option. It provides the interpreter with its "umbilical cord" to the host application, containing all I/O, identity, and callback functions.
- **`WithExecPolicy(policy *policy.ExecPolicy) Option`**: Applies a security policy that governs what the script is allowed to do (e.g., which tools it can call, what capabilities it has). If not provided, it defaults to a restrictive policy.
- **`WithSandboxDir(path string) Option`**: Sets the root directory for all file-based operations, preventing the script from accessing the broader filesystem.
- **`WithGlobals(globals map[string]interface{}) Option`**: Injects a map of Go values as initial global variables into the interpreter's state.
- **`WithAITranscriptWriter(w io.Writer) Option`**: Provides a writer to which the full transcript of conversations with AI providers will be written, for logging or debugging.
- **`WithoutStandardTools() Option`**: Prevents the automatic registration of the standard tool library.
- **`WithProviderRegistry(...)`**: Injects a registry of AI providers. (See Section 6.3)
- **`WithAccountStore(...)` / `WithAgentModelStore(...)`**: Injects shared, persistent stores for state. (See Section 6.3)
- **`WithCapsuleRegistry(...)` / `WithCapsuleAdminRegistry(...)`**: Configures registries for managing packaged scripts ("capsules"). (See Section 7)

### The `HostContext` Struct

This struct is the centerpiece of the configuration API. It **must** be built using the `api.NewHostContextBuilder()` fluent API. It is passed by reference and is considered immutable after the interpreter is created.

```go
// Simplified representation of the HostContext's fields
type HostContext struct {
   Logger                    interfaces.Logger // Mandatory
   Stdout                    io.Writer         // Mandatory
   Stdin                     io.Reader         // Mandatory
   Stderr                    io.Writer         // Mandatory
   Actor                     interfaces.Actor  // Optional: Injects identity
   ServiceRegistry           any // Optional: (e.g., map[string]any)
   EmitFunc                  // Optional: Callback for 'emit'
   WhisperFunc               // Optional: Callback for 'whisper'
   EventHandlerErrorCallback // Optional: Callback for event errors
   // ... and other host-provided APIs.
}
```

**Host-Level Services:** The `ServiceRegistry` field is the generic mechanism for injecting host services. It has two primary uses:
1.  **Host Symbols:** Injecting an `api.SymbolProvider` using `api.SymbolProviderKey` (see Section 6.1).
2.  **AEIOU Hook:** Injecting an `interfaces.AeiouOrchestrator` using `interfaces.AeiouServiceKey` (see Section 6.6 and `api/aeiou_hook_guide.md`).

### Convenience Helper: `NewConfigInterpreter`

**`NewConfigInterpreter(allowedTools []string, grants []capability.Capability, otherOpts ...Option) *Interpreter`**

This helper function creates a new interpreter pre-configured with a trusted "config" context policy. It is the recommended way to create interpreters that need to run administrative scripts (e.g., `tool.account.register`).

### Example: Full Setup

```go
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"[github.com/aprice2704/neuroscript/pkg/api](https://github.com/aprice2704/neuroscript/pkg/api)"
)

func main() {
	// 1. Create a HostContext with mandatory I/O and a logger.
	var output bytes.Buffer

	// The api package exports logger constructors for convenience:
	// - api.NewNoOpLogger(): A silent logger, good for production or simple examples.
	// - api.NewTestLogger(t): A logger that writes to a *testing.T object, ideal for Go tests.
	// For other logging, the host must provide its own api.Logger implementation.
	hostCtx, err := api.NewHostContextBuilder().
		WithLogger(api.NewNoOpLogger()). // Using the No-Op logger for this example
		WithStdout(&output).
		WithStdin(os.Stdin).
		WithStderr(os.Stderr).
		Build()
	if err != nil {
		panic(err)
	}

	// 2. Instantiate the interpreter with the context.
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

	// 4. Load definitions by executing the tree.
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		panic(err)
	}

	// 5. Run a procedure from the loaded library.
	result, err := api.RunProcedure(context.Background(), interp, "greet", "World")
	if err != nil {
		panic(err)
	}

    // api.Unwrap converts the NeuroScript Value back to a Go native type.
	unwrapped, _ := api.Unwrap(result)
	fmt.Println("Result from Run:", unwrapped) // "Hello, World"
}
```

---

## 3. Parsing Code

**`Parse(src []byte, mode ParseMode) (*Tree, error)`**

This is the sole entry point for parsing NeuroScript source code. It takes raw bytes and a `ParseMode` constant.

* **`api.ParseSkipComments`**: (Default) Parses the script and discards all comments. Use this for execution.
* **`api.ParsePreserveComments`**: Parses the script and attaches comments to their corresponding AST nodes. This is used by linters or formatters.

*Note: Passing a `ParseMode` of `0` is equivalent to `api.ParseSkipComments`.*

---

## 4. Loading and Executing Scripts

It is critical to understand the distinction between **definitions** (`func`/`on event`) and **imperative code** (`command` blocks).

### `ExecWithInterpreter(ctx context.Context, interp *Interpreter, tree *Tree) (Value, error)`

This is the primary method for both loading definitions and executing imperative code. It performs two actions in order:
1.  **Loads Definitions:** It walks the AST and loads all `func` and `on event` blocks into the interpreter's state.
2.  **Runs Commands:** It executes any top-level `command` blocks found in the script.

> **Symbol Conflict Rule:** The load will **fail** if the script attempts to define a `func` or `on event` with the same name as a symbol already provided by the host's `SymbolProvider`. This "No Override" rule ensures host-level functions are deterministic and secure.

### `LoadFromUnit(interp *Interpreter, unit *LoadedUnit) error`

This function is used for securely-loaded scripts. It loads definitions from a `LoadedUnit` (the output of `api.Load()`), but it does *not* execute `command` blocks. This is the recommended path for loading verified code. It is also subject to the "No Override" rule.

### `RunProcedure(ctx context.Context, interp *Interpreter, name string, args ...any) (Value, error)`

This is the primary method for *invoking* a specific piece of loaded code. It executes a named `func` block and passes the provided Go-native arguments to it (after wrapping them). It returns the value from the procedure's `return` statement.

### `ExecScript(...)`
A helper primarily for testing. It combines parsing, loading, and executing into a single call.

---

## 5. Persistence & Secure Loading

The API provides two distinct workflows for serializing and deserializing ASTs, serving two different use cases: secure script loading and service-level graph persistence.

### 5.1 Secure Loading (SignedAST Workflow)

This workflow is for verifying that a script has not been tampered with. It uses a `SignedAST` struct and the `api.Load` function.

1.  **`api.Canonicalise(tree *interfaces.Tree) ([]byte, [32]byte, error)`**:
    Takes a parsed `Tree` and returns its deterministic binary `blob` and its `sum` (a blake2b hash).
2.  **Sign:** The host signs the `sum` with a private key to produce a `sig`.
3.  **Store:** The host stores the `api.SignedAST{Blob: blob, Sum: sum, Sig: sig}`.
4.  **`api.Load(ctx context.Context, s *SignedAST, cfg LoaderConfig, pubKey ed25519.PublicKey) (*LoadedUnit, error)`**:
    This function performs the full security check:
    a. Verifies the `sig` matches the `sum` using the `pubKey`.
    b. Verifies the `sum` matches a new hash of the `blob`.
    c. Decodes the verified `blob` into an AST.
    d. Returns a `LoadedUnit` containing the trusted `Tree`.
5.  **`api.LoadFromUnit(interp, loadedUnit)`**:
    The trusted `LoadedUnit` is finally loaded into an interpreter.

### 5.2 Service Persistence (Registry Workflow)

This workflow is used by host services (like `nsinterpretersvc`) that need to persist and retrieve *full, executable ASTs* (e.g., agent definitions) from a database.

* **`api.CanonicaliseWithRegistry(tree *Tree) ([]byte, [32]byte, error)`**:
    Serializes a *full program* (which must have an `*ast.Program` root) into a binary blob. This is what you write to the graph.
* **`api.DecodeWithRegistry(blob []byte) (*Tree, error)`**:
    Deserializes the binary blob back into a *full program* `Tree`. This is what you read from the graph. The resulting `Tree` can be passed to `api.ExecWithInterpreter`.

### 5.3 Graph Node Persistence (Fragment Workflow)

This workflow is for persisting *individual AST nodes* (like a procedure or a default value) as data in a graph.

* **`api.CanonicaliseNode(node ast.Node) ([]byte, [32]byte, error)`**: Serializes a single node (e.g., `*ast.Procedure`, `*ast.StringLiteralNode`).
* **`api.DecodeNode(blob []byte) (ast.Node, error)`**: Deserializes a single node.
* **`api.ValueToNode(val lang.Value) (ast.Node, error)`**: Converts a runtime `lang.Value` (like a map or string) into its corresponding AST literal node.
* **`api.NodeToValue(node ast.Node) (lang.Value, error)`**: Converts an AST literal node back into its runtime `lang.Value`.

---

## 6. Advanced Interaction & State Management

### 6.1 Host-Provided Symbols (SymbolProvider)

To provide host-defined functions and constants, you must implement and inject a `SymbolProvider`.

1.  Implement the `api.SymbolProvider` interface.
2.  Create a map: `serviceReg := map[string]any{ api.SymbolProviderKey: myProvider }`
3.  Build the `HostContext`: `api.NewHostContextBuilder().WithServiceRegistry(serviceReg).Build()`
4.  Pass to the interpreter: `api.New(api.WithHostContext(hostCtx))`

### 6.2 Custom Tool Example

You can register your own Go functions as tools scripts can call.

```go
myTool := api.ToolImplementation{
	Spec: api.ToolSpec{
		Name:  "FormatGreeting",
		Group: "strings",
		Args:  []api.ArgSpec{{Name: "name", Type: "string"}},
	},
	Func: func(rt api.Runtime, args []any) (any, error) {
		name, _ := args[0].(string)
		return "Hello from a Go tool, " + name + "!", nil
	},
}

// Register the tool on the interpreter's registry
interp.ToolRegistry().RegisterTool(myTool)
```

### 6.3 AI & Provider Management

The host application registers AI providers and state stores by injecting them *at creation time*.

* **`api.NewProviderRegistry() *api.ProviderRegistry`**: Creates a new, empty registry for AI providers.
* **`api.WithProviderRegistry(registry *provider.Registry) Option`**: Injects the populated registry into the interpreter.
* **`RegisterAgentModel(...)`**: A method on the interpreter to register a new agent configuration.

### 6.4 Account & Model Stores

* **`api.NewAccountStore() *api.AccountStore`**: Creates a new, in-memory store for accounts.
* **`api.WithAccountStore(store *api.AccountStore) Option`**: Injects the store.
* **`api.NewAgentModelStore() *api.AgentModelStore`**: Creates a new, in-memory store for agent models.
* **`api.WithAgentModelStore(store *api.AgentModelStore) Option`**: Injects the store.

### 6.5 Variable Management (Globals)

* **`WithGlobals(globals map[string]interface{})`**: Injects Go values as global variables at creation time.

### 6.6 Host-Level Orchestration (AEIOU Hook)

To implement a custom `ask` loop (like FDM's AEIOU hook), the host injects an orchestrator service.

1.  Implement the `interfaces.AeiouOrchestrator` interface (from `pkg/interfaces/aeiou.go`).
2.  Create a map: `serviceReg := map[string]any{ interfaces.AeiouServiceKey: myService }`
3.  Inject this map via `WithServiceRegistry` (see 6.1).

When a script calls `ask`, the interpreter will call your service's `RunAskLoop` method. Your service is then responsible for calling the LLM and securely executing its code responses using `api.ExecuteSandboxedAST(...)`.

See `api/aeiou_hook_guide.md` for the full architecture.

---

## 7. Capsule Management

The API provides functions for loading, validating, and managing "capsules" (packaged scripts).

* **`api.ParseCapsule(content []byte) (*api.Capsule, error)`**:
    Parses and validates a raw capsule file. Checks for `::id`, `::version`, `::description` and calculates the SHA256 hash.
* **`api.NewAdminCapsuleRegistry() *api.AdminCapsuleRegistry`**:
    Creates a new, empty, writable registry for your custom capsules.
* **`api.DefaultCapsuleRegistry() *api.CapsuleRegistry`**:
    Returns the read-only registry of built-in NeuroScript capsules.
* **`api.NewCapsuleStore(initial ...*api.CapsuleRegistry) *api.CapsuleStore`**:
    Creates a layered, read-only store.
    `myStore := api.NewCapsuleStore(myCustomRegistry, api.DefaultCapsuleRegistry())`
* **`WithCapsuleRegistry(myStore)`**:
    Injects your store to be used by the built-in `tool.capsule.*` tools.

See `api/ns_api_capsule.md` for the full guide.

---

## 8. Data Helper Functions

The API provides helpers for converting between NeuroScript `Value` types and native Go `any` types.

* **`api.Unwrap(v Value) (any, error)`**:
    Converts an `api.Value` (e.g., `lang.StringValue`) into its corresponding Go type (e.g., `string`). Used to get results *out* of the interpreter.
* **`api.LangWrap(v any) (Value, error)`**:
    Converts a Go type (e.g., `string`, `map[string]any`) into its corresponding `api.Value` (e.g., `lang.StringValue`, `lang.MapValue`). Used to pass arguments *into* `RunProcedure`.
* **`api.LangToString(v Value) (string, bool)`**:
    Safely converts an `api.Value` to a Go `string`.

---

## 9. Shape API (Data Validation)

The `pkg/api/shape` package provides a public API for data validation.

* **`shape.ParseShape(...)`**: Compiles a validation map into a `*Shape`.
* **`shape.ValidateNSEvent(data, nil)`**: Example convenience function to validate a map against the built-in NeuroScript event shape.
* **`shape.ComposeNSEvent(...)`**: Example convenience function to build a valid event map.

See `api/shape/shape_guide.md` for details.

---

## 10. Error Handling

The `api` package re-exports key sentinel errors from the interpreter core, allowing you to use `errors.Is()` for robust error handling.

```go
_, err := api.RunProcedure(ctx, interp, "my_func")
if err != nil {
    if errors.Is(err, api.ErrProcedureNotFound) {
        // Handle a "function not found" error
    } else if errors.Is(err, api.ErrToolNotAllowed) {
        // Handle a policy violation
    }
}
```

Key errors include:
* `api.ErrSyntax`
* `api.ErrProcedureNotFound`
* `api.ErrArgumentMismatch`
* `api.ErrToolNotFound`
* `api.ErrToolNotAllowed`
* `api.ErrPolicyViolation`
* `api.ErrHandleNotFound`
* `api.ErrInvalidMagic` (from `DecodeWithRegistry`)
* `api.ErrTruncatedData` (from `DecodeWithRegistry`)
```