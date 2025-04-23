:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-shell-gobuild-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_shell.go, pkg/core/tools_register.go, docs/script_spec.md
:: relatedTo: Shell.ExecuteCommand, Shell.GoCheck, Shell.GoTest
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `Shell.GoBuild` (v0.1)

* **Tool Name:** `Shell.GoBuild` (v0.1)
* **Purpose:** Compiles Go packages and their dependencies within the sandbox directory. This is primarily used to check for compilation errors or produce executable binaries. Equivalent to running `go build [target]`.
* **NeuroScript Syntax:** `CALL Shell.GoBuild([target: <String>])`
* **Arguments:**
    * `target` (String): Optional. The build target package(s) or file(s). Can be a package path (e.g., `./cmd/mytool`), a pattern (e.g., `./...` to build all packages within the current module), or a specific Go file. Defaults to `./...` if omitted. The path is validated to be within the sandbox.
* **Return Value:** (Map)
    * Returns a Map containing the results of the `go build` command execution, identical in structure to the `Shell.ExecuteCommand` result. (Accessible via `LAST` after the `CALL`). Keys:
        * `stdout` (String): Standard output (usually empty for `go build` unless errors occur related to package listing).
        * `stderr` (String): Standard error (contains compilation errors, warnings, or other messages from the build process).
        * `exit_code` (Number): The exit code of the `go build` command (0 for success, non-zero for failure).
        * `success` (Boolean): `true` if the build completed successfully (exit code 0), `false` otherwise.
* **Behavior:**
    1.  Validates that zero or one argument (`target`) of type String is provided.
    2.  If `target` is provided, validates it using `SecureFilePath` against the current working directory (which should ideally match the `sandboxDir`). Returns an error map if validation fails. If `target` is ".", it's treated as valid.
    3.  Sets the build target: uses the validated `target` if provided, otherwise defaults to `./...`.
    4.  Constructs the command: `go build <build_target>`.
    5.  Executes the `go build` command using the `Shell.ExecuteCommand` tool logic (within the interpreter's `sandboxDir`).
    6.  Captures `stdout`, `stderr`, `exit_code`, and `success` status from the execution.
    7.  Returns the result Map. Compilation errors will be present in `stderr` and reflected in `exit_code` and `success`.
* **Security Considerations:**
    * Executes the `go build` command, which involves running the Go compiler and potentially downloading dependencies via `go get`. While generally safe for valid Go code, ensure the Go toolchain itself is secure and that dependency sources are trusted.
    * The optional `target` argument is validated using `SecureFilePath`, but the build process might read files outside the immediate target (e.g., other packages in the module, module cache in `GOPATH`).
    * Runs within the interpreter's configured `sandboxDir` via the underlying call to `toolExecuteCommand`.
    * Depends on `Shell.ExecuteCommand`'s security model.
    * Assumes the `go` executable is available in the environment's PATH.
* **Examples:**
    ```neuroscript
    # Example 1: Build all packages in the current module
    EMIT "Attempting to build all packages ('./...')..."
    CALL Shell.GoBuild() # No target, defaults to './...'
    SET build_all_result = LAST

    IF build_all_result["success"] == true THEN
        EMIT "Build successful!"
        EMIT "Output (stderr should be empty): " + build_all_result["stderr"]
    ELSE
        EMIT "Build failed!"
        EMIT "Exit Code: " + build_all_result["exit_code"]
        EMIT "Error Output (stderr):"
        EMIT build_all_result["stderr"]
    ENDIF

    # Example 2: Build a specific package (assuming ./cmd/mytool exists)
    EMIT "Attempting to build specific package './cmd/mytool'..."
    CALL Shell.GoBuild("./cmd/mytool")
    SET build_specific_result = LAST

    IF build_specific_result["success"] THEN
        EMIT "Build of ./cmd/mytool successful."
    ELSE
        EMIT "Build of ./cmd/mytool failed. Stderr:"
        EMIT build_specific_result["stderr"]
    ENDIF

    # Example 3: Attempt to build invalid target path (validation failure)
    CALL Shell.GoBuild("../../../outside_project")
    SET build_invalid_result = LAST
    EMIT "Build Invalid Path Success Status: " + build_invalid_result["success"] # Expect false
    EMIT "Build Invalid Path Error: " + build_invalid_result["stderr"] # Expect path error
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_shell.go`
    * Function: `toolGoBuild`
    * Spec Name: `GoBuild`
    * Key Go Packages: `fmt`, `os`, `path/filepath`
    * Helpers: `core.SecureFilePath`, `core.toolExecuteCommand`
    * Registration: Registered by `registerShellTools` within `pkg/core/tools_register.go`. Returns the result map from the underlying `toolExecuteCommand`.