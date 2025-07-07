# NeuroScript Tool ↔ Runtime Contract

Tools are **ordinary Go functions** that receive a single parameter—an object
satisfying the `tool.Runtime` interface. Everything a tool may do with the
running program is mediated by that interface; tools must **never** import
`interpreter` or `neurogo`.

```go
// abridged – see pkg/tool/tool_types.go for the authoritative spec.
package tool

type Runtime interface {
    // Logging / UI
    Println(...any)
    Ask(prompt string) (string, error)
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

---

## 1  Usage Patterns

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
ans, _ := rt.Ask("Your name?")
rt.Println("Hello,", ans)

// Invoke another tool
sum, _ := rt.CallTool("math.add", []any{int64(2), int64(3)})
```

---

## 2  Registering Tool-Sets

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

## 3  External & Dynamic Tool-Sets  *(planned roadmap)*

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

## 4  Versioning & Compatibility

1. **Adding** a method to `tool.Runtime` → minor (back-compatible).  
2. **Removing / changing** a method → major version bump.  
3. Tool authors may assume:
   * `SandboxDir()` is always writable.  
   * `Println` is line-buffered.  
   * `GetVar/SetVar` are O(1) for < 1 k variables.  

---

## 5  Test Guidelines

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
