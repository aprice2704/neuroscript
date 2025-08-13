# Tool Author's Guide to Policy & Integrity

The NeuroScript interpreter now includes a powerful **Policy Gate** that makes runtime decisions about whether to permit or deny a tool's execution. To function correctly, every tool must be annotated with metadata that declares its trust level, required permissions, and side effects.

Your responsibility is to provide this metadata accurately for every `ToolImplementation` you create.

## 1. Annotating Your `ToolImplementation`

For each tool, you must correctly populate the following four fields in the `ToolImplementation` struct. The tool registry will automatically handle generating the `SignatureChecksum` for you during registration.

```go
// From pkg/tool/tool_types.go
type ToolImplementation struct {
    Spec              ToolSpec
    Func              ToolFunc
    // --- FIELDS YOU MUST SET ---
    RequiresTrust     bool
    RequiredCaps      []capability.Capability
    Effects           []string

    // --- FIELDS HANDLED BY THE REGISTRY ---
    FullName          types.FullName
    SignatureChecksum string
}
```

#### `RequiresTrust: bool`

This flag marks tools that can perform sensitive or system-altering operations.

* **Set to `true` if your tool:**
    * Modifies system state outside its sandbox (e.g., `os.Setenv`).
    * Performs administrative actions (e.g., `agentmodel.Register`).
    * Reads sensitive secrets (e.g., `os.Getenv`).
* **Set to `false` if your tool:**
    * Is a pure function (e.g., `str.Contains`, `math.Add`).
    * Only interacts with the filesystem within its designated sandbox.
    * Is generally safe to run in a low-privilege context.

#### `RequiredCaps: []capability.Capability`

This is a list of specific permissions the tool needs to function. If the calling script doesn't have a grant that satisfies *every* capability in this list, the call will be blocked.

* **Structure**: `capability.Capability{Resource: string, Verbs: []string, Scopes: []string}`
* **Common Examples**:
    * Reading a specific environment variable:
        ```go
        RequiredCaps: []capability.Capability{
            {Resource: "env", Verbs: []string{"read"}, Scopes: []string{"OPENAI_API_KEY"}},
        },
        ```
    * Writing to a temporary directory:
        ```go
        RequiredCaps: []capability.Capability{
            {Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"/tmp/*"}},
        },
        ```
    * Using a specific class of AI models:
         ```go
        RequiredCaps: []capability.Capability{
            {Resource: "model", Verbs: []string{"use"}, Scopes: []string{"gpt-4-*"}},
        },
        ```
    * No specific capabilities needed:
        ```go
        RequiredCaps: nil,
        ```

#### `Effects: []string`

This declares the side effects of your tool, which helps the interpreter with optimizations like caching.

* **`"idempotent"`**: The tool can be called multiple times with the same arguments and will produce the same result without changing system state beyond the first call. (e.g., `fs.Write`, `agentmodel.Register`).
* **`"readsNet"`**: The tool reads from the network.
* **`"readsFS"`**: The tool reads from the filesystem.
* **`"readsClock"`**: The tool's output depends on the current time (non-deterministic).
* **`"readsRand"`**: The tool's output depends on a random source (non-deterministic).

### 2. Complete Example

Here is a correctly annotated implementation for a hypothetical `fs.Read` tool:

```go
var FsReadTool = tool.ToolImplementation{
    Spec: tool.ToolSpec{
        Name:        "Read",
        Group:       "fs",
        Description: "Reads the content of a file.",
        Args: []tool.ArgSpec{{Name: "path", Type: tool.ArgTypeString, Required: true}},
        ReturnType:  tool.ArgTypeString,
    },
    Func: fsReadFunc, // The actual Go function implementation

    // --- Policy Metadata ---
    RequiresTrust: false,
    RequiredCaps: []capability.Capability{
        {Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"*"}}, // Requires read access to any file
    },
    Effects: []string{"readsFS"},
}
```
By providing this metadata, you ensure your tools integrate smoothly and securely with the NeuroScript runtime.