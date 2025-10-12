 # NeuroScript Interpreter: Public API Guide
 
 **Audience:** Developers integrating the NeuroScript interpreter into host applications.
 **Version:** Reflects architecture post-October 2025 refactor.
 **Purpose:** This document outlines the intended public API for instantiating, configuring, and running the NeuroScript interpreter. It focuses on the primary entry points and data structures a developer needs to understand to embed NeuroScript successfully.
 
 ---
 
 ## 1. The Core Lifecycle: Create, Configure, Load, Run
 
 Interacting with the interpreter follows a clear, four-step lifecycle. The API is designed to configure all external dependencies upfront, load script logic, and then execute it.
 
 
 
 1.  **Instantiation & Configuration:** Create an `Interpreter` instance using `NewInterpreter()` and configure its connection to the outside world via a `HostContext` and other options.
 2.  **Loading Code:** Load an initial script AST into the interpreter using `Load()`. This populates the interpreter with procedures, event handlers, and commands, clearing any previous state.
 3.  **Appending Code (Optional):** Merge additional scripts into the interpreter's state using `AppendScript()`. This is useful for loading libraries or modules without overwriting the main script.
 4.  **Execution:** Run the loaded code, either by invoking a specific procedure with `Run()` or by executing all top-level `command` blocks with `ExecuteCommands()`.
 
 ---
 
 ## 2. Instantiation and Configuration
 
 This is the most critical phase. An interpreter cannot function without being properly configured with its host dependencies.
 
 ### `NewInterpreter(opts ...InterpreterOption) *Interpreter`
 
 This is the sole entry point for creating a new interpreter instance. It returns a **root interpreter**. It takes a variable number of `InterpreterOption` functions that configure the instance.
 
 ### `InterpreterOption` Functions
 
 Options are functions that modify the interpreter's configuration during creation. The most important one is `WithHostContext`.
 
 - **`WithHostContext(hc *HostContext) InterpreterOption`**: **(Mandatory)** This is the primary and essential option. It provides the interpreter with its "umbilical cord" to the host application, containing all I/O and callback functions.
 - **`WithExecPolicy(policy *policy.ExecPolicy) InterpreterOption`**: Applies a security policy that governs what the script is allowed to do (e.g., which tools it can call, what capabilities it has). If not provided, it defaults to a restrictive policy.
 - **`WithSandboxDir(path string) InterpreterOption`**: Sets the root directory for all file-based operations, preventing the script from accessing the broader filesystem.
 - **`WithGlobals(globals map[string]interface{}) InterpreterOption`**: Injects a map of Go values as initial global variables into the interpreter's state.
 - **`WithoutStandardTools() InterpreterOption`**: Prevents the automatic registration of the standard tool library, useful for creating highly restricted or specialized runtimes.
 - **`WithCapsuleRegistry(...)` / `WithCapsuleAdminRegistry(...)`**: Configures registries for managing packaged scripts ("capsules").
 
 ### The `HostContext` Struct
 
 This struct is the centerpiece of the configuration API. It's a plain data struct that you, the host developer, must populate. It is passed by reference and is considered immutable after the interpreter is created.
 
 ```go
 // from interpreter/hostcontext.go
 type HostContext struct {
     // A structured logger is mandatory.
     Logger                    interfaces.Logger
 
     // Standard I/O streams are mandatory.
     Stdout                    io.Writer
     Stdin                     io.Reader
     Stderr                    io.Writer
 
     // Callback for the 'emit' statement.
     EmitFunc                  func(lang.Value)
 
     // Callback for the 'whisper' statement.
     WhisperFunc               func(handle, data lang.Value)
 
     // Optional: Callback for unhandled errors within event handlers.
     EventHandlerErrorCallback func(eventName, source string, err *lang.RuntimeError)
 
     // Host-provided APIs and other dependencies.
     FileAPI                   interfaces.FileAPI
     Emitter                   interfaces.Emitter
     AITranscript              io.Writer
 }
 ```

 
 **Key Takeaway:** A minimal, functioning interpreter requires calling `NewInterpreter` with at least `WithHostContext`, where the context has `Logger`, `Stdout`, `Stdin`, and `Stderr` populated.
 
 ---
 
 ## 3. Loading and Executing Scripts
 
 Once configured, you can load and run your NeuroScript code.
 
 ### `Load(tree *interfaces.Tree) error`
 
 This method loads a parsed AST (`interfaces.Tree` which contains an `*ast.Program`) into the interpreter. **This is a destructive action**: it completely replaces any procedures, event handlers, or commands that were previously loaded.
 
 ### `AppendScript(tree *interfaces.Tree) error`
 
 This method merges a new AST into the *existing* state of the interpreter. It will add new procedures and event handlers but will return an error if you try to define a procedure that already exists. It appends new top-level `command` blocks to the existing queue. This is ideal for loading modular libraries.
 
 ### `Run(procName string, args ...lang.Value) (lang.Value, error)`
 
 This is the primary method for executing a specific piece of loaded code. It invokes a named procedure (a `func` block in the script) and passes the provided arguments to it. It returns the value from the procedure's `return` statement.
 
 ### `ExecuteCommands() (lang.Value, error)`
 
 This method executes all the top-level `command` blocks in the loaded scripts in the order they were defined. `command` blocks are intended for initialization or top-level script logic that isn't part of a callable procedure.
 
 ### `ExecuteScriptString(...)`
 
 While available, this is primarily a helper for testing and simple use cases. It combines parsing, loading, and executing into a single call. For production use, it's recommended to handle parsing as a separate step so you can cache the AST.
 
 ---
 
 ## 4. Advanced Interaction & State Management
 
 The API also provides access to the interpreter's various registries and state stores. These are often used for setup, or for tools that need to interact with the interpreter's core.
 
 ### AI & Provider Management
 
 - **`RegisterProvider(name string, p provider.AIProvider)`**: Allows the host application to register a concrete AI provider (e.g., a connection to OpenAI, Anthropic, etc.) under a specific name that scripts can reference.
 - **`AgentModelsAdmin() interfaces.AgentModelAdmin` / `AgentModels() interfaces.AgentModelReader`**: Provides administrative (`Register`, `Update`, `Delete`) and read-only (`Get`, `List`) access to the store of "AgentModels," which are configurations that bundle a provider, model name, and other settings.
 
 ### Tool Management
 
 - **`ToolRegistry() tool.ToolRegistry`**: Returns the interpreter's tool registry. You can use this to programmatically register custom Go functions as tools that scripts can call.
 
 ### Account Management
 
 - **`SetAccountStore(store *account.Store)`**: Allows the host to replace the default in-memory account store with a persistent one.
 - **`AccountsAdmin() interfaces.AccountAdmin` / `Accounts() interfaces.AccountReader`**: Provides administrative and read-only access to the account store, which is used to manage API keys and other secrets that AgentModels can reference.
 
 ### Variable Management
 
 - **`SetInitialVariable(name string, value any) error`**: Sets a variable in the interpreter's **global scope** before any script is run. This is the primary way to inject data into a script from the host.
 
 ```