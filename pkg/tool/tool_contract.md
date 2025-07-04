# NeuroScript Tool ↔ Runtime Contract

Tools are **ordinary Go functions** that receive a single parameter—an object
satisfying the `tool.Runtime` interface. Everything a tool may do with the
running program is mediated by that interface; tools must **never** import
`interpreter` or `neurogo`.

```go
// abridged – see pkg/tool/runtime.go for the authoritative spec.
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
}
```

---

## 1. Usage Patterns

### Logging

```go
logger := rt.GetLogger()           // ALWAYS use the host's logger
logger.Debug("hashing file", "path", p)
```

*Never* `fmt.Printf` from a tool—central logging keeps TUI/CLI and servers in
sync.

### Secure file access

```go
sandbox := rt.SandboxDir()                         // e.g. /tmp/ns-run-123/
abs, err := security.ResolveAndSecurePath(rel, sandbox)
```

### Sharing state with the script

```go
tokens, ok := rt.GetVar("remaining_api_tokens")
if ok && tokens.(int64) <= 0 {
    return nil, errors.New("quota exceeded")
}
rt.SetVar("remaining_api_tokens", tokens.(int64)-1)
```

### Interactive I/O and AI calls

```go
name, _ := rt.Ask("Your name?")
rt.Println("Hello,", name)
```

### Composing tools

```go
sum, err := rt.CallTool("math.add", []any{int64(2), int64(3)})
if err != nil { return nil, err }
```

---

## 2. Registering a Tool-Set

Each tool package provides an **`init()`** that registers its implementations.
Example for the *math* tool-set:

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

## 3. Versioning & Compatibility Rules

1. **Adding** a method to `tool.Runtime` → minor, back-compatible.
2. **Removing / changing** a method → breaking; bump major module version.
3. Tools may assume:
   * `SandboxDir()` is always writable.
   * `Println` is line-buffered.
   * `GetVar/SetVar` are O(1) for small scopes.

---

## 4. Test Guidelines

| need                                   | recommended package        | cycle-safe |
|----------------------------------------|----------------------------|------------|
| Unit-test private helpers in a tool    | `package <tool>`           | ✅          |
| Integration-test via Engine            | `package <tool>_test`      | ✅          |

For interpreter-level white-box tests use the build-tagged
`interpreter.NewTestInterpreter`.

---

By following this contract tools remain **sandboxed, portable, and swap-able**,
while the host **Engine** retains full control over logging, security, and
dependency injection.
