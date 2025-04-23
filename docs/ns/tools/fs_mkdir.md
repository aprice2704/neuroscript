:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-fs-mkdir-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_fs_dirs.go, docs/script_spec.md
:: relatedTo: FS.ListDirectory, FS.WriteFile, FS.DeletePath
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `FS.Mkdir` (v0.1)

* **Tool Name:** `FS.Mkdir` (v0.1)
* **Purpose:** Creates a new directory, including any necessary parent directories that do not exist, within the designated sandbox.
* **NeuroScript Syntax:** `CALL FS.Mkdir(path: <String>)`
* **Arguments:**
    * `path` (String): Required. The relative path within the sandbox of the directory to create. Intermediate parent directories will be created as needed.
* **Return Value:** (String)
    * On success: A status message string indicating success (e.g., "Successfully created directory: logs/today"). (Accessible via `LAST` after the `CALL`).
    * On error: A string describing the failure (e.g., "Mkdir path security error...", "Failed to create directory..."). (Accessible via `LAST` after the `CALL`).
* **Behavior:**
    1.  Validates that exactly one argument (`path`) of type String is provided.
    2.  Validates the `path` argument is not an empty string.
    3.  Retrieves the interpreter's configured `sandboxDir`.
    4.  Determines the parent directory of the requested `path`.
    5.  Uses the `SecureFilePath` helper to validate the *parent directory* against the `sandboxDir`. If the parent directory path is invalid or outside the sandbox, returns an error message string.
    6.  Constructs the full absolute path of the target directory to be created based on the validated parent directory and the final component of the input `path`.
    7.  Performs an additional security check to ensure the final absolute path to create does not escape the validated sandbox path (using `filepath.Clean` and prefix checks). If it escapes, returns an error message string.
    8.  If all security checks pass, attempts to create the directory (and any necessary parents) using `os.MkdirAll` with default permissions (0755).
    9.  If `os.MkdirAll` fails (e.g., permission denied, path element exists but is not a directory), returns an error message string.
    10. If directory creation is successful, returns a success message string including the original relative path.
* **Security Considerations:**
    * Confined to the interpreter's sandbox directory via `SecureFilePath` validation on the parent path and explicit checks on the final path.
    * Creates directories with permissions mode 0755 (before umask).
    * Can create multiple directory levels if they don't exist within the sandbox.
* **Examples:**
    ```neuroscript
    # Example 1: Create a simple directory
    CALL FS.Mkdir("output_data")
    SET mkdir_status = LAST
    EMIT "Mkdir status for 'output_data': " + mkdir_status

    # Example 2: Create nested directories
    CALL FS.Mkdir("project_files/src/components")
    SET nested_status = LAST
    IF String.Contains(nested_status, "Successfully created") THEN
        EMIT "Nested directories created successfully."
        # Now we can safely write a file there
        CALL FS.WriteFile("project_files/src/components/button.js", "// Button component code")
        SET write_status = LAST
        EMIT "File write status: " + write_status
    ELSE
        EMIT "Failed to create nested directories: " + nested_status
    ENDBLOCK

    # Example 3: Attempt to create directory outside sandbox (will fail)
    CALL FS.Mkdir("../../new_system_dir")
    SET bad_mkdir_status = LAST
    # Status will likely indicate a path security error
    EMIT "Status of attempting restricted mkdir: " + bad_mkdir_status
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_fs_dirs.go`
    * Function: `toolMkdir`
    * Key Go Packages: `os`, `fmt`, `path/filepath`, `strings`
    * Helpers: `core.SecureFilePath`
    * Registration: Called by `registerFsDirTools` within `pkg/core/tools_fs.go`. Returns `string, error`.