# NeuroScript Tool ↔ Runtime Contract

This document defines the two-way contract between a tool and the NeuroScript runtime.

1.  **Tool $\rightarrow$ Runtime:** The tool provides a `ToolImplementation` (Spec, Func, Policy).
2.  **Runtime $\rightarrow$ Tool:** The runtime provides a `tool.Runtime` interface to the tool's function.

---

## 1. Defining & Implementing Tools

A tool is defined by a `tool.ToolImplementation` struct, which combines its specification, its Go function, and its policy requirements.

```go
var myTool = tool.ToolImplementation{
    Spec: tool.ToolSpec{
        Name: "my_tool",
        Group: "example",
        Args: []tool.ArgSpec{
            {Name: "path", Type: tool.ArgTypeString, Required: true},
            {Name: "tags", Type: tool.ArgTypeSliceString, Required: false},
        },
        ReturnType: tool.ArgTypeBool,
    },
    Func: myToolFunc,
}
```

### The `ToolSpec` Contract: Validation & Coercion

The most critical part of the contract is the `ToolSpec`, specifically the `Args` slice. This enables centralized argument validation.

**Tool Author's Responsibility (The "Spec"):**
* You **MUST** define the type of every argument as precisely as possible (e.g., `tool.ArgTypeSliceString`, `tool.ArgTypeInt`, `tool.ArgTypeMap`).
* You **MUST** correctly flag `Required` arguments.
* You **MUST NOT** use `tool.ArgTypeAny` unless it is unavoidable.

**Runtime's Responsibility (The "Validation"):**
In exchange, the NeuroScript runtime **guarantees** that *before* your `ToolFunc` is ever called, it will:
1.  Check the total argument count.
2.  Return an error if any `Required` arguments are missing or `nil`.
3.  **Coerce** all provided arguments to the exact Go types specified in your `ArgSpec`.

This centralizes all parameter validation. Your tool function no longer needs to do boilerplate type-checking and can trust its inputs.

#### Example: The New Contract in Practice

**OLD WAY (DEPRECATED):** Using `ArgTypeAny` and manual type-checking inside the tool.

```go
// Spec:
// {Name: "tags", Type: tool.ArgTypeAny, Required: false}

// Func:
func myToolFunc(rt tool.Runtime, args []any) (any, error) {
    // Manual, brittle type-checking
    tags, ok := sliceAnyToString(args[1]) // Requires helper
    if !ok {
        return nil, fmt.Errorf("tags must be a slice of strings")
    }
    // ...
}
```

**NEW WAY (REQUIRED):** Using a specific `ArgSpec` and trusting the pre-coerced types.

```go
// Spec:
// {Name: "tags", Type: tool.ArgTypeSliceString, Required: false}

// Func:
func myToolFunc(rt tool.Runtime, args []any) (any, error) {
    path := args[0].(string) // Safe: guaranteed to be string

    var tags []string
    if args[1] != nil {
        // Safe: guaranteed to be []string if not nil
        tags = args[1].([]string)
    } else {
        tags = []string{} // Handle optional arg
    }
    // ...
    return true, nil
}
```

---

## 2. Usage Patterns (Runtime Interface)

Tools are **ordinary Go functions** that receive arguments and a `tool.Runtime`
interface. Everything a tool may do with the running program is mediated by that
interface; tools must **never** import `interpreter` or `neurogo`.

```go
// abridged – see pkg/tool/tool_types.go for the authoritative spec.
package tool

type Runtime interface {
    // Logging / UI
    Println(...any)
    PromptUser(prompt string) (string, error)
    GetLogger() interfaces.Logger

    // Program state
    GetVar(name string) (any, bool)
    SetVar(name string, val any)

    // Safe file access
    SandboxDir() string

    // Tool composition
    CallTool(name string, args []any) (any, error)
    LLM() interfaces.LLMClient
}
```

*Logging, secure file access, state-sharing, interactive I/O, and composing
other tools* follow the idioms shown below.

```go
// Logging
logger := rt.GetLogger()
logger.Debug("hashing file", "path", p)

// Secure file access
sandbox := rt.SandboxDir()
abs, err := security.ResolveAndSecurePath(rel, sandbox)

// Share mutable state
if tok, ok := rt.GetVar("remaining"); ok && tok.(int64) > 0 {
    rt.SetVar("remaining", tok.(int64)-1)
}

// Chat / console I/O
ans, _ := rt.PromptUser("Your name?")
rt.Println("Hello,", ans)

// Invoke another tool
sum, _ := rt.CallTool("math.add", []any{int64(2), int64(3)})
```

---

## 3. Registering Tool-Sets

Each tool package provides an **`init()`** that registers its
implementations:

```go
package math
import "github.com/aprice2704/neuroscript/pkg/tool"

func init() {
    tool.AddToolsetRegistration(
        "math",
        tool.CreateRegistrationFunc("math", mathToolsToRegister),
    )
}
```

At start-up **`neurogo.Engine`** calls `tool.RegisterExtendedTools(reg)`,
which iterates over every registration function collected via `init()` and
builds the live registry.

---

## 4. External & Dynamic Tool-Sets  *(planned roadmap)*

We will support three concentric levels of extensibility without changing
the `Runtime` surface:

| level | mechanism | status |
|-------|-----------|--------|
| **A** | *Out-of-tree, compiled-in* – any Go module that imports **`tool`** can self-register via `init()`. Add a blank-import in the host binary. | **works today** |
| **B** | *Optional bundles via build tags* – external tool-sets guard their registration file with `//go:build ns_with_<name>`. Host chooses tags at `go build` time. | **planned**, trivial once tool-sets adopt build tags |
| **C** | *Runtime-loadable plugins* – a shared-object (`.so`) exposes a single symbol:  `var Register tool.PluginRegister`. `Engine.LoadPlugin(path)` opens the plugin, type-asserts the symbol, and calls it. | **design drafted**, implementation pending |

```go
// plugin side (future)
package main
import "github.com/aprice2704/neuroscript/pkg/tool"

var Register tool.PluginRegister = func(reg tool.ToolRegistrar) error {
    return tool.CreateRegistrationFunc("acme", acmeTools)(reg)
}
```

The same **`tool.Runtime`** contract applies regardless of how the tool-set
arrives; only the *delivery* pathway differs.

---

## 5. Versioning & Compatibility

1.  **Adding** a method to `tool.Runtime` → minor (back-compatible).
2.  **Removing / changing** a method → major version bump.
3.  Tool authors may assume:
      * `SandboxDir()` is always writable.
      * `Println` is line-buffered.
      * `GetVar/SetVar` are O(1) for \< 1 k variables.

---

## 6. Test Guidelines

| need                                   | recommended package        | cycle-safe |
|----------------------------------------|----------------------------|------------|
| Unit-test private helpers in a tool    | `package <tool>`           | ✅          |
| Integration-test via Engine            | `package <tool>_test`      | ✅          |

For interpreter white-box tests use the build-tagged
`interpreter.NewTestInterpreter`.

---

By following this contract tools remain **sandboxed, portable, and
swap-able**, while the host **Engine** retains full control over logging,
security, dependency injection, and—eventually—dynamic plugin loading.