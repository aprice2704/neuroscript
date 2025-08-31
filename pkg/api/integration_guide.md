# NeuroScript Integration Guide (v0.7 — 2025-08-28)

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

-   **`api.New(opts ...api.Option) *api.Interpreter`**: Creates a new interpreter.
-   **`api.WithSandboxDir(path string) api.Option`**: **Mandatory for filesystem tools.** Sets the secure root directory for all file operations for this interpreter instance. The application is responsible for creating this directory.
-   **`api.WithLogger(logger api.Logger) api.Option`**: Provides a custom logger that conforms to the `interfaces.Logger` interface.
-   **`api.WithStdout(w io.Writer) api.Option`**: Sets the standard output stream for the `emit` command.
-   **`api.WithStderr(w io.Writer) api.Option`**: Sets the standard error stream.
-   **`api.WithGlobals(globals map[string]any) api.Option`**: Sets initial global variables that can be accessed by the script.
-   **`api.WithTool(tool api.ToolImplementation) api.Option`**: Registers a custom Go function as a tool (see Section 6).

### Standard Tool and Provider Registration

**Core tools** (like `tool.math.Add`, `tool.fs.Read`, etc.) and **default AI providers** (like `google`) are automatically registered for you when you call `api.New()`. You do not need to take any special steps, such as importing a "tool bundle", to make them available. They are ready to use immediately in your NeuroScript code.

### Managing I/O

You can set the I/O streams after creation as well.

-   **`interp.SetStdout(w io.Writer)`**
-   **`interp.SetStderr(w io.Writer)`**

### Executing Code and Handling Values

-   **`api.LoadFromUnit(interp *api.Interpreter, unit *api.LoadedUnit) error`**: Loads definitions from a verified unit into an interpreter.
-   **`api.RunProcedure(ctx context.Context, interp *api.Interpreter, name string, args ...any) (api.Value, error)`**: Executes a named procedure, automatically wrapping native Go arguments (`any`) into NeuroScript values.
-   **`api.Unwrap(v api.Value) (any, error)`**: Converts a NeuroScript `api.Value` back into a standard Go `any` type.

---

## 4. Trusted Configuration Scripts

For running initialization or setup scripts that require privileged tools (e.g., `tool.agentmodel.Register` or `tool.os.Getenv`), the API provides a high-level helper to create a correctly configured, trusted interpreter.

-   **`api.NewConfigInterpreter(allowedTools []string, grants []api.Capability, otherOpts ...api.Option) *api.Interpreter`**: Creates a new interpreter pre-configured with a trusted policy.

This function handles the details of setting the execution context to `config`, which enables tools marked with `RequiresTrust = true`.

### Example: Initializing AgentModels

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/api"
)

func main() {
	// 1. Define the capabilities the startup script needs.
	// The api package re-exports helpers and constants for building capabilities.
	requiredGrants := []api.Capability{
		api.NewWithVerbs(api.ResModel, []string{api.VerbAdmin}, []string{"*"}),
		api.New(api.ResEnv, api.VerbRead, "OPENAI_API_KEY"),
	}

	// 2. Specify which tools the script is allowed to call.
	allowedTools := []string{
		"tool.agentmodel.Register",
		"tool.os.Getenv",
	}

	// 3. Create the trusted interpreter.
	interp := api.NewConfigInterpreter(allowedTools, requiredGrants, api.WithStdout(os.Stdout))

	// 4. Execute the trusted setup script.
	setupScript := `
command
    emit "Registering AgentModels..."
    set api_key = tool.os.Getenv("OPENAI_API_KEY")
    must tool.agentmodel.Register("default", {
        "provider": "openai",
        "model": "gpt-4o-mini",
        "api_key": api_key
    })
    emit "Setup complete."
endcommand
`
	// The ExecInNewInterpreter helper is not used here because we are using a
	// pre-configured interpreter instance.
	tree, _ := api.Parse([]byte(setupScript), api.ParseSkipComments)
	_, err := api.ExecWithInterpreter(context.Background(), interp, tree)
	if err != nil {
		panic(err)
	}
}
```

---

## 5. Stateless Execution

For simple, one-shot script execution where the source is trusted, you can use a more direct helper.

-   **`api.ExecInNewInterpreter(ctx context.Context, src string, opts ...api.Option) (api.Value, error)`**: Parses and runs any top-level `command` blocks in a single call.

```go
// Example: Stateless execution
src := `command { emit "This is a one-shot execution!" }`
result, err := api.ExecInNewInterpreter(context.Background(), src, api.WithStdout(os.Stdout))
```

---

## 5b. Agentic Features: `ask`, `whisper`, and the Ask-Loop

NeuroScript's most powerful features are for building AI agents. This involves a partnership between NeuroScript code (the agent's "mind") and the host Go application (the "body" that provides tools and controls the execution flow).

### The `ask` Statement

The `ask` statement is the primary way a script communicates with an AI model. It takes an `AgentModel` name and a prompt, and returns the model's response. An `AgentModel` is a configuration, registered via a trusted script using `tool.agentmodel.Register`, that specifies the AI provider, model ID, and permissions (like `tool_loop_permitted`).

```neuroscript
# Simple, single-turn query
set user_summary = ask "summarizer_agent", "Summarize this text: ..."
```

### The `whisper` Command

The `whisper` command is how the host application provides context to the script. It's a key-value store that is private to the script's execution and is not sent to the AI model unless the script explicitly includes it in a prompt. This is the primary mechanism for passing state between turns in an Ask-Loop.

### The Host-Managed Ask-Loop

When an agent with `tool_loop_permitted: true` wants to perform a multi-turn task, it doesn't loop internally. Instead, it **emits a control signal** to the host. The host application is responsible for running the actual loop, capturing the signal, and deciding whether to execute the next turn. This gives the host full control over agent execution, resource usage, and security.

#### Host Responsibilities

1.  **Manage the Loop**: Use a standard Go `for` loop to manage turns and enforce limits like `MaxTurns`.
2.  **Capture `emit` Output**: Use the `interp.SetEmitFunc` method to redirect all output from `emit` statements into a buffer.
3.  **Provide Context**: Pass information (like the previous turn's output) back to the script using the `whisper` command.
4.  **Parse Control Signals**: After each turn, use the `api.ParseLoopControl` helper to scan the captured output for the agent's `LOOP` signal (`continue`, `done`, or `abort`).

#### Ask-Loop Example

This example shows how a Go application can manage a simple, multi-turn interaction with an agent.

```go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
)

func main() {
	// 1. Create an interpreter and register a looping agent.
	// In a real app, this would be done via a trusted setup script.
	interp := api.New()
	// ... setup code to register a mock provider ...
	agentConfig := map[string]api.Value{
		"provider":            "mock_provider",
		"model":               "looper",
		"tool_loop_permitted": true,
		"max_turns":           5,
	}
	// This is a simplified registration for the example.
	// A real app would use a trusted script with tool.agentmodel.Register.
	_ = interp.RegisterAgentModel("looper_agent", agentConfig)

	// 2. The host application runs the loop.
	lastOutput := "Initial task: plan a party."
	const MAX_TURNS = 5

	for turn := 1; turn <= MAX_TURNS; turn++ {
		fmt.Printf("--- Turn %d ---\n", turn)
		var outputCollector strings.Builder
		interp.SetEmitFunc(func(v api.Value) {
			val, _ := api.Unwrap(v)
			fmt.Fprintln(&outputCollector, val)
		})

		// 3. Construct and execute the script for this turn.
		// The `lastOutput` from the previous turn is whispered as context.
		script := fmt.Sprintf(`
command
    whisper self, "last_turn_output", %q
    ask "looper_agent", "Continue the plan based on the last output." into result
    emit result
endcommand`, lastOutput)

		// Re-use the same interpreter instance for each turn.
		if _, err := api.ExecInNewInterpreter(context.Background(), script, WithInterpreter(interp)); err != nil {
			fmt.Printf("Turn failed: %v\n", err); return
		}

		lastOutput = outputCollector.String()
		fmt.Printf("Agent Output:\n%s\n", lastOutput)

		// 4. Parse the control signal from the agent's output.
		loopControl, err := api.ParseLoopControl(lastOutput)
		if err != nil {
			fmt.Println("No loop signal found. Halting."); break
		}

		// 5. Act on the signal.
		if loopControl.Control == "done" {
			fmt.Println("--- Loop Finished: Agent signaled done. ---"); return
		}
		if loopControl.Control == "abort" {
			fmt.Printf("--- Loop Aborted by Agent: %s ---\n", loopControl.Reason); return
		}
	}
	fmt.Println("--- Loop Terminated: Max turns exceeded. ---")
}

// Helper to reuse an interpreter instance with the one-shot executor.
func WithInterpreter(existing *api.Interpreter) api.Option {
    return func(i *interpreter.Interpreter) { *i = *existing.internal }
}
```









## 6. Custom Tools

You can extend NeuroScript with custom Go functions by registering them as "tools".

-   **`api.MakeToolFullName(group, name string) api.FullName`**: A helper to create the correctly formatted name for tool registry lookups.

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
	// Note the mandatory "tool." prefix.
	src := `func do_math() returns result means
  set result = tool.host.add(10, 32)
  return result
endfunc`

	tree, _ := api.Parse([]byte(src), api.ParseSkipComments)

	// ExecWithInterpreter loads definitions and can execute top-level commands.
	// It returns a result and an error, which we are ignoring here
	// as we only need to load the 'do_math' function into the interpreter.
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		panic(err) // Or handle error appropriately
	}

	// 5. Run the procedure and check the result.
	val, _ := api.RunProcedure(context.Background(), interp, "do_math")
	unwrapped, _ := api.Unwrap(val)
	fmt.Printf("Result from custom tool: %v\n", unwrapped) // Output: 42
}
```

---

## 7. Registering AI Providers

To use the `ask` statement, the interpreter must know how to communicate with an AI backend. This is done by registering an **AI Provider**. A provider is a Go object that implements the `api.AIProvider` interface.

-   **`interp.RegisterProvider(name string, p api.AIProvider)`**: Registers a provider implementation under a given name. This name is then used in AgentModel configurations to select which provider to use.

### Example: Registering a Mock Provider

```go
package main

import (
	"context"
	"github.com/aprice2704/neuroscript/pkg/api"
	"[github.com/aprice2704/neuroscript/pkg/provider](https://github.com/aprice2704/neuroscript/pkg/provider)" // Note: internal package needed for AIRequest/AIResponse
)

// 1. Create a struct that implements the api.AIProvider interface.
type MockProvider struct{}

func (m *MockProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	// In a real provider, this would make an API call.
	// Here, we just return a canned response.
	return &provider.AIResponse{
		TextContent: "This is a mock response to the prompt: " + req.Prompt,
	}, nil
}

func main() {
	// 2. Create a new interpreter.
	interp := api.New()

	// 3. Instantiate your provider and register it with the interpreter.
	mockProv := &MockProvider{}
	interp.RegisterProvider("mock_provider", mockProv)

	// Now, any AgentModel configured with "provider": "mock_provider" will use this implementation.
}
```

---

## 8. Critical Error Handling

The NeuroScript engine can identify critical, potentially security-related errors. By default, it will `panic` when one occurs. You can override this behavior to integrate with your application's logging and monitoring systems.

-   **`api.RegisterCriticalErrorHandler(h func(*api.RuntimeError))`**: Replaces the default panic handler with your custom function.

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

## 9. Implementing the Ask-Loop

The `ask` statement is designed for single-turn AI queries, but it is also the foundation of multi-turn "agentic" workflows. The **Ask-Loop** is a pattern, controlled by the host Go application, that orchestrates these conversations according to the `askloop_spec_v2.md` protocol.

The core principle is that the host application is in complete control. It runs a loop that repeatedly executes a NeuroScript `command` block. The agent's script uses the `emit` stream to send back both its results and a special `LOOP` control signal. The host captures this output, parses the signal, and decides whether to continue, stop, or abort the loop.

### Host Responsibilities

1.  **Manage the Loop**: Use a standard Go `for` loop to manage turns and enforce limits like `MaxTurns`.
2.  **Capture `emit` Output**: Use the `api.WithEmitFunc` option or `interp.SetEmitFunc` method to redirect all output from `emit` statements into a buffer.
3.  **Provide `OUTPUT` as Context**: The captured output from the previous turn serves as the `OUTPUT` section of the next turn's prompt. This is typically passed to the agent using `whisper self`.
4.  **Parse Control Signals**: After each turn, use the `api.ParseLoopControl` helper to scan the captured output for the agent's `LOOP` signal.

### Step-by-Step Example

This example shows how a Go application can manage a simple, two-turn interaction with an agent.

```go
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/api"
	"[github.com/aprice2704/neuroscript/pkg/lang](https://github.com/aprice2704/neuroscript/pkg/lang)"
)

const MAX_TURNS = 5

func main() {
	// 1. Create an interpreter.
	interp := api.New()

	// This would typically be a more complex prompt.
	initialPrompt := "Create a plan to add a user."
	lastOutput := "" // There is no output before the first turn.

	// 2. The host application runs the loop.
	for turn := 0; turn < MAX_TURNS; turn++ {
		fmt.Printf("--- Turn %d ---\n", turn+1)
		var outputBuffer bytes.Buffer

		// 3. Configure the interpreter to capture this turn's emit stream.
		interp.SetEmitFunc(func(v api.Value) {
			// This captures the raw output from the agent's script.
			fmt.Fprintln(&outputBuffer, v.(lang.Value).String())
		})

		// 4. Construct the script for this turn.
		// The `lastOutput` from the previous turn is whispered as context.
		script := fmt.Sprintf(`
command
    whisper self, %q
    ask "default_agent", %q into result
    emit result
endcommand`, lastOutput, initialPrompt)

		// 5. Execute the script.
		tree, _ := api.Parse([]byte(script))
		_, err := api.ExecWithInterpreter(context.Background(), interp, tree)
		if err != nil {
			fmt.Printf("Turn failed: %v\n", err)
			return
		}

		lastOutput = outputBuffer.String()
		fmt.Printf("Agent Output:\n%s\n", lastOutput)

		// 6. Parse the control signal from the output.
		loopControl, err := api.ParseLoopControl(lastOutput)
		if err != nil {
			fmt.Printf("Could not parse loop control: %v. Halting.\n", err)
			return
		}

		// 7. Act on the signal.
		if loopControl.Control == "done" {
			fmt.Println("--- Loop Finished: Agent signaled done. ---")
			return
		}
		if loopControl.Control == "abort" {
			fmt.Printf("--- Loop Aborted by Agent: %s ---\n", loopControl.Reason)
			return
		}
		// Otherwise, continue to the next turn.
	}

	fmt.Println("--- Loop Terminated: Max turns exceeded. ---")
}
```

---

## 10. Handling Events

NeuroScript includes a powerful event-driven model that allows your Go application to trigger behavior within a loaded script. This is ideal for building reactive systems where the script needs to respond to asynchronous events from the host.

The flow is straightforward:
1.  **Declare Handlers**: Your NeuroScript code defines one or more `on event` blocks that listen for specific event names.
2.  **Emit Events**: Your Go application calls the `EmitEvent` method on an interpreter instance to trigger these handlers.

Handlers are executed in a **sandboxed clone** of the interpreter. This is a critical security and stability feature: it means that any variables set or state changes made within an `on event` block are discarded after the handler completes and will **not** affect the main interpreter's state.

### Step-by-Step Example

This example demonstrates how to register a handler for a `user:login` event and trigger it from the host application.

```go
package main

import (
	"context"
	"os"

	"github.com/aprice2704/neuroscript/pkg/api"
	"[github.com/aprice2704/neuroscript/pkg/lang](https://github.com/aprice2704/neuroscript/pkg/lang)"
)

func main() {
	// 1. Define a script with an 'on event' handler.
	// The handler expects a payload and uses the 'emit' command to print a message.
	script := `
on event "user:login" as data do
  set user_id = data["payload"]["id"]
  emit "EVENT: User " + user_id + " has logged in."
endon
`

	// 2. Create an interpreter and load the script. This automatically
	// registers the 'on event' handler.
	interp := api.New(api.WithStdout(os.Stdout))
	tree, _ := api.Parse([]byte(script), api.ParseSkipComments)
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		panic(err)
	}

	// 3. From your application, create a payload for the event.
	// Payloads are maps of NeuroScript values.
	payloadData := map[string]lang.Value{
		"id":    lang.StringValue{Value: "usr_f81b"},
		"level": lang.NumberValue{Value: 3},
	}
	payload := lang.NewMapValue(payloadData)

	// 4. Call EmitEvent to trigger all registered handlers for "user:login".
	// The 'source' argument is for logging and traceability.
	interp.EmitEvent("user:login", "AuthService", payload)
}

// Expected output:
// EVENT: User usr_f81b has logged in.
```


## 11. Core Types Reference

The `api` package re-exports all necessary types so you don't need to import internal packages.

-   **AST & Nodes**: `api.Tree`, `api.Node`, `api.Position`, `api.Kind`
-   **Execution**: `api.Interpreter`, `api.Value`, `api.Option`, `api.RuntimeError`
-   **Security & Policy**: `api.SignedAST`, `api.LoadedUnit`, `api.ExecPolicy`, `api.Capability`
-   **Capability Helpers**:
    -   Constants for resources (`api.ResFS`, `api.ResNet`, etc.) and verbs (`api.VerbRead`, `api.VerbUse`, etc.).
    -   Constructor functions (`api.NewCapability`, `api.NewWithVerbs`).
-   **AI Providers**: `api.AIProvider`
-   **Tools**: `api.ToolImplementation`, `api.ToolSpec`, `api.ToolFunc`, `api.ArgSpec`, `api.FullName`, `api.ToolName`, `api.ToolGroup`
-   **Logging**: `api.Logger`

---

## 12. Important “Don’ts”

-   **Do not** use filesystem tools without configuring a sandbox via `api.WithSandboxDir`.
-   **Do not** import `pkg/parser`, `pkg/interpreter`, etc., directly. Use the `api` package.
-   **Do not** execute a script from an untrusted source without using the full `Parse -> Sign -> Load` workflow.
-   **Do not** re-canonicalise an AST after it has been verified by `api.Load`.
-   **Do not** construct tool names manually; always use the `api.MakeToolFullName` helper to ensure the correct `tool.group.name` format.

---

**End of file**