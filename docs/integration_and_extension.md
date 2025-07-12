# NeuroScript Tool ↔ Runtime Contract (Expanded and Corrected)

This guide details the contract between NeuroScript tools and the runtime environment. It expands upon the original `tool_contract.md` with corrections and patterns observed in the actual codebase to provide a complete and accurate reference for developers.

---

## 1. The Core Principle: A Tool is a Go Function

At its heart, a NeuroScript tool is a Go function with a specific signature. It receives a `tool.Runtime` interface, which is its sole gateway for interacting with the interpreter, and a slice of arguments.

```go
// The required signature for any tool implementation.
type ToolFunc func(rt Runtime, args []interface{}) (interface{}, error)
```

Tools must **never** import the `interpreter` or `neurogo` packages directly. All interactions must be mediated through the `Runtime` interface to ensure they are sandboxed, portable, and testable.

---

## 2. The `tool.Runtime` Interface: Your Gateway to the Engine

The `tool.Runtime` interface provides all the necessary methods for a tool to perform its function securely and effectively.

```go
// As defined in pkg/tool/tools_types.go
type Runtime interface {
    // Logging and Interactive I/O
    Println(...any)
    Ask(prompt string) string
    GetLogger() interfaces.Logger

    // Program State Management
    GetVar(name string) (any, bool)
    SetVar(name string, val any)

    // Sandboxed File Access
    SandboxDir() string

    // Introspection and Composition
    CallTool(name types.FullName, args []any) (any, error)
    ToolRegistry() ToolRegistry // Essential for introspection
    LLM() interfaces.LLMClient
}
```

- **Logging:** Use `GetLogger()` to get a structured logger instance. Example from `fs/tools_fs_write.go`: `interpreter.GetLogger().Debug("...", "key", value)`.
- **Console I/O:** Use `Println` to output information and `Ask` to prompt for user input.
- **State Sharing:** Use `GetVar` and `SetVar` for simple, mutable state sharing between tools. This is intended for basic data like tokens or flags, not complex state.
- **Secure File Access:** Always use `SandboxDir()` as the root for any file operations. The `fs` tools heavily rely on this, combined with helpers like `security.ResolveAndSecurePath`, to prevent directory traversal attacks.
- **Tool Composition:** A tool can invoke another using `CallTool`, passing the fully qualified tool name (e.g., `"math.add"`) and its arguments.
- **Introspection:** The `ToolRegistry()` method is a powerful addition not in the original contract. It allows a tool to get information about other available tools, which is how the `meta.ListTools` and `meta.ToolsHelp` tools work.

---

## 3. Defining a Tool: The `ToolImplementation` Struct

A tool is more than just its function; it requires a detailed specification for the interpreter to understand its capabilities. This is defined in a `tool.ToolImplementation` struct.

```go
// Found in pkg/tool/tool_types.go
type ToolImplementation struct {
	FullName types.FullName
	Spec     ToolSpec
	Func     ToolFunc
}

type ToolSpec struct {
	Name            types.ToolName  `json:"name"`
	Group           types.ToolGroup `json:"groupname"`
	FullName        types.FullName  `json:"fullname"`
	Description     string          `json:"description"`
	Category        string          `json:"category,omitempty"`
	Args            []ArgSpec       `json:"args,omitempty"`
	ReturnType      ArgType         `json:"returnType"`
	ReturnHelp      string          `json:"returnHelp,omitempty"`
	Example         string          `json:"example,omitempty"`
	ErrorConditions string          `json:"errorConditions,omitempty"`
}

type ArgSpec struct {
	Name        string      `json:"name"`
	Type        ArgType     `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
}
```

- **Group vs. Name:** A tool belongs to a `Group` (e.g., "fs", "git", "str") and has a `Name` (e.g., "Read", "Commit", "Split"). The `FullName` (e.g., "fs.Read") is generated from these.
- **Metadata is Key:** The `Description`, `Category`, `Args`, `ReturnHelp`, `Example`, and `ErrorConditions` fields are not just comments. They are used by the `meta.ToolsHelp` tool to provide rich, runtime documentation to the user. Keep them clear and accurate.

---

## 4. Registering a Toolset

Tool registration is decentralized and automatic. Each tool package is responsible for registering its own tools.

1.  **Create a `register.go` file** in your tool package (e.g., `pkg/tool/mytools/register.go`).
2.  **Define an `init()` function** inside `register.go`. This Go function will be executed automatically when the package is imported.
3.  **Call `tool.AddToolsetRegistration`** from within `init()`.

This pattern is used consistently across all toolsets like `fs`, `git`, `strtools`, `list`, and `time`.

**Example `pkg/tool/mytools/register.go`:**
```go
package mytools

import "github.com/aprice2704/neuroscript/pkg/tool"

// myToolsToRegister would be a []tool.ToolImplementation defined in another file
// in this package, like tooldefs_mytools.go.

func init() {
    tool.AddToolsetRegistration(
        "mytools", // The group name for this toolset
        tool.CreateRegistrationFunc("mytools", myToolsToRegister),
    )
}
```

4.  **Ensure your package is imported.** The host binary uses a file like `toolbundles/all/all.go` to perform blank imports (`_ "..."`) of all tool packages, which guarantees their `init()` functions run.

---

## 5. Authoring and Testing Tools

### Best Practices
- **Argument Handling:** Your `ToolFunc` receives `[]interface{}`. You are responsible for validating the argument count and performing type assertions.
- **Error Handling:** Don't return generic Go errors. Construct and return a `*lang.RuntimeError` using `lang.NewRuntimeError(code, message, underlyingError)`. This provides structured error information to the interpreter and the user script.
- **Advanced Interfaces:** For tools that need deeper integration with the interpreter (like `script.LoadScript`), you can define a more specific interface and use a type assertion on the `tool.Runtime` object. This is an advanced pattern but allows for powerful extensions.

### Testing
- **Unit Tests:** For private helpers within your tool package, use standard `package <toolname>` tests.
- **Integration Tests:** For testing the tool against a live interpreter, use `package <toolname>_test`.
- **Test Setup:**
    - Create a new interpreter instance for your test: `interp := interpreter.NewInterpreter()`.
    - Manually register your tool(s) with the interpreter's registry: `interp.ToolRegistry().RegisterTool(myToolImplementation)`.
    - Create helper functions (e.g., `testMyToolHelper`) to reduce boilerplate in table-driven tests. This is a common pattern seen in `tools_string_basic_test.go` and `tools_fs_helpers_test.go`.

---

## 6. External & Dynamic Tool-Sets (Roadmap)

The contract is designed for future extensibility without changing the `Runtime` surface.

| Level | Mechanism | Status |
|---|---|---|
| **A** | **Out-of-tree, compiled-in:** Any Go module that imports `tool` can self-register via `init()`. The host binary just needs to add a blank-import. | **Works Today** |
| **B** | **Optional bundles via build tags:** Guard registration files with `//go:build`. | **Planned** |
| **C** | **Runtime-loadable plugins:** Load `.so` files at runtime. | **Design Drafted** |

By following this contract, tools remain **sandboxed, portable, and swappable**, while the host **Engine** retains full control over logging, security, dependency injection, and future dynamic loading capabilities.
