 # NeuroScript API: Interpreter Configuration
 
 **Audience:** Developers integrating the NeuroScript interpreter.
 **Purpose:** This document specifies the official pattern for instantiating and configuring the NeuroScript interpreter, focusing on the clear separation between building dependencies and configuring the runtime.
 
 
 ## The Configuration Pattern: Build, Then Configure
 
 The public API is designed around a simple, two-step process to ensure clarity and type safety.
 
 
 ### Step 1: Build the `HostContext`
 
 All external dependencies—such as I/O streams, loggers, and callbacks—are bundled into a single `HostContext` object. Because this object has several mandatory fields, it **must** be created using the `HostContextBuilder`.
 
 The builder provides a fluent, readable API and guarantees at compile time that a valid `HostContext` is created. It will return an error if any mandatory fields are missing.
 
 **Example:**
 ```go
 hc, err := api.NewHostContextBuilder().
     WithLogger(myLogger).
     WithStdout(os.Stdout).
     WithStdin(os.Stdin).
     WithStderr(os.Stderr).
     WithEmitFunc(myEmitCallback).
     Build()
 if err != nil {
     // Handle error
 }
 ```
 
@&zwnj;`
 
 ### Step 2: Configure the Interpreter
 
 Once you have a valid `HostContext`, you create an interpreter instance using `api.New()` and pass the context in using the `WithHostContext` option. Other options for configuring policies, stores, and tools are provided separately.
 
 **Example:**
 ```go
 interp, err := api.New(
     api.WithHostContext(hc),
     api.WithExecPolicy(myPolicy),
     api.WithGlobals(map[string]any{"debug": true}),
 )
 if err != nil {
     // Handle error
 }
 ```
 
 This pattern cleanly separates the concern of *dependency construction* from *interpreter configuration*, leading to a more robust and maintainable integration. The internal `interpreter` package is responsible for mapping the public `api.HostContext` to its internal representation.