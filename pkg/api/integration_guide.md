# NeuroScript Integration Guide (v0.7 — 2025-07-23)

This guide provides a comprehensive overview of how to embed and interact with the NeuroScript engine in an external Go application using **only** the public `neuroscript/pkg/api` package. The `api` package is a stable facade that wraps the underlying parser, interpreter, and cryptographic components.

---

## 1. Versioning

You can access both the overall program version and the specific grammar version directly from the API. This is useful for logging, debugging, and ensuring compatibility.

```go
import "github.com/aprice2704/neuroscript/pkg/api"
import "fmt"

func main() {
  fmt.Printf("NeuroScript Program Version: %s\n", api.ProgramVersion)
  fmt.Printf("NeuroScript Grammar Version: %s\n", api.GrammarVersion)
}
```

---

## 2. Core Workflow: The Golden Path (Verified Execution)

For security-sensitive contexts, the "golden path" ensures that only cryptographically signed and verified scripts are executed. This workflow prevents unauthorized code execution and tampering.

> **Golden Path:** `Parse → Canonicalise → Sign → Load → Create Interpreter → Execute`

### Step-by-Step Example

```go
package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/api"
)

func main() {
	// === Part 1: Signing Authority (e.g., a build server) ===

	src := `
func greet(name) means
  return "Hello, " + name + "!"
endfunc
`
	pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)

	tree, err := api.Parse([]byte(src), api.ParseSkipComments)
	if err != nil {
		panic(err)
	}

	blob, sum, err := api.Canonicalise(tree)
	if err != nil {
		panic(err)
	}

	sig := ed25519.Sign(privKey, sum[:])
	signedAST := &api.SignedAST{Blob: blob, Sum: sum, Sig: sig}

	// === Part 2: Execution Environment (e.g., a production service) ===

	// 1. Load the signed AST. This verifies the signature and runs static analysis.
	loadedUnit, err := api.Load(context.Background(), signedAST, api.LoaderConfig{}, pubKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("Script verified and loaded successfully!")

	// 2. Create a persistent interpreter instance.
	// The sandbox directory MUST be provided if using file-based tools.
	sandboxPath := "./safe_to_run_in"
	os.MkdirAll(sandboxPath, 0750)
	defer os.RemoveAll(sandboxPath)

	interp := api.New(
		api.WithSandboxDir(sandboxPath),
		api.WithStdout(os.Stdout),
	)

	// 3. Load the verified code into the interpreter.
	if err := api.LoadFromUnit(interp, loadedUnit); err != nil {
		panic(err)
	}

	// 4. Execute a specific procedure from the loaded code.
	result, err := api.RunProcedure(context.Background(), interp, "greet", "FDM Team")
	if err != nil {
		panic(err)
	}

	// 5. Unwrap the result back to a native Go type.
	goResult, _ := api.Unwrap(result)
	fmt.Printf("Go Result: %v (Type: %T)\n", goResult, goResult)
}
```

---

## 3. Interpreter Facade API

For stateful applications, you can create a long-running interpreter instance and interact with it over time.

### Interpreter Configuration

Use `api.New()` with functional options to configure the interpreter.

- **`api.New(opts ...api.Option) *api.Interpreter`**: Creates a new interpreter.
- **`api.WithSandboxDir(path string) api.Option`**: **Mandatory for filesystem tools.** Sets the secure root directory for all file operations for this interpreter instance. The application is responsible for creating this directory.
- **`api.WithLogger(logger api.Logger) api.Option`**: Provides a custom logger that conforms to the `interfaces.Logger` interface.
- **`api.WithStdout(w io.Writer) api.Option`**: Sets the standard output stream for the `emit` command.
- **`api.WithStderr(w io.Writer) api.Option`**: Sets the standard error stream.
- **`api.WithTool(tool api.ToolImplementation) api.Option`**: Registers a custom Go function as a tool (see Section 5).

### Managing I/O

You can set the I/O streams after creation as well.

- **`interp.SetStdout(w io.Writer)`**
- **`interp.SetStderr(w io.Writer)`**

### Executing Code and Handling Values

- **`api.LoadFromUnit(interp *api.Interpreter, unit *api.LoadedUnit) error`**: Loads definitions from a verified unit into an interpreter.
- **`api.RunProcedure(ctx context.Context, interp *api.Interpreter, name string, args ...any) (api.Value, error)`**: Executes a named procedure, automatically wrapping native Go arguments (`any`) into NeuroScript values.
- **`api.Unwrap(v api.Value) (any, error)`**: Converts a NeuroScript `api.Value` back into a standard Go `any` type.

---

## 4. Stateless Execution

For simple, one-shot script execution where the source is trusted, you can use a more direct helper.

- **`api.ExecInNewInterpreter(ctx context.Context, src string, opts ...api.Option) (api.Value, error)`**: Parses and runs any top-level `command` blocks in a single call.

```go
// Example: Stateless execution
src := `command { emit "This is a one-shot execution!" }`
result, err := api.ExecInNewInterpreter(context.Background(), src, api.WithStdout(os.Stdout))
```

---

## 5. Custom Tools

You can extend NeuroScript with custom Go functions by registering them as "tools".

- **`api.MakeToolFullName(group, name string) api.FullName`**: A helper to create the correctly formatted name for tool registry lookups.

```go
package main

import (
	"context"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// 1. Define the tool's function. It must match the api.ToolFunc signature.
func AddTool(rt api.Runtime, args []any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("adder tool expects 2 arguments")
	}
	a, okA := args[0].(float64)
	b, okB := args[1].(float64)
	if !okA || !okB {
		return nil, fmt.Errorf("arguments must be numbers")
	}
	return a + b, nil
}

func main() {
	// 2. Define the tool's implementation using re-exported API types.
	adderImpl := api.ToolImplementation{
		// Use the helper for clarity and correctness.
		FullName: api.MakeToolFullName("host", "add"),
		Spec: api.ToolSpec{
			Name:  "add",
			Group: "host",
			Args: []api.ArgSpec{
				{Name: "a", Type: "number", Required: true},
				{Name: "b", Type: "number", Required: true},
			},
			ReturnType: "number",
		},
		Func: AddTool,
	}

	// 3. Create an interpreter with the custom tool registered.
	toolOpt := api.WithTool(adderImpl)
	interp := api.New(toolOpt)

	// 4. Run NeuroScript code that uses the tool.
	src := `func do_math() returns result means
  set result = host.add(10, 32)
  return result
endfunc`

	tree, _ := api.Parse([]byte(src), api.ParseSkipComments)
	api.ExecWithInterpreter(context.Background(), interp, tree) // This loads the function

	// 5. Run the procedure and check the result.
	val, _ := api.RunProcedure(context.Background(), interp, "do_math")
	unwrapped, _ := api.Unwrap(val)
	fmt.Printf("Result from custom tool: %v\n", unwrapped) // Output: 42
}
```

---

## 6. Critical Error Handling

The NeuroScript engine can identify critical, potentially security-related errors. By default, it will `panic` when one occurs. You can override this behavior to integrate with your application's logging and monitoring systems.

- **`api.RegisterCriticalErrorHandler(h func(*api.RuntimeError))`**: Replaces the default panic handler with your custom function.

```go
import (
	"log"
	"os"

	"github.com/aprice2704/neuroscript/pkg/api"
)

func init() {
	// Register the handler during your application's startup.
	api.RegisterCriticalErrorHandler(func(e *api.RuntimeError) {
		log.Printf("[CRITICAL-ERROR] NeuroScript engine reported a fatal error: %v", e)
		os.Exit(1)
	})
	log.Println("Custom NeuroScript critical error handler registered.")
}
```

---

## 7. Core Types Reference

The `api` package re-exports all necessary types so you don't need to import internal packages.

- **AST & Nodes**: `api.Tree`, `api.Node`, `api.Position`, `api.Kind`
- **Execution**: `api.Interpreter`, `api.Value`, `api.Option`, `api.RuntimeError`
- **Security**: `api.SignedAST`, `api.LoadedUnit`
- **Tools**: `api.ToolImplementation`, `api.ToolSpec`, `api.ToolFunc`, `api.ArgSpec`, `api.FullName`
- **Logging**: `api.Logger`

---

## 8. Important “Don’ts”

- **Do not** use filesystem tools without configuring a sandbox via `api.WithSandboxDir`.
- **Do not** import `pkg/parser`, `pkg/interpreter`, etc., directly. Use the `api` package.
- **Do not** execute a script from an untrusted source without using the full `Parse -> Sign -> Load` workflow.
- **Do not** re-canonicalise an AST after it has been verified by `api.Load`.
- **Do not** roll-your-own tool name, use the MakeToolFullName helper.

---

**End of file**