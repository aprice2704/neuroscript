:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.1 # Incremented version due to example correction
:: id: tool-spec-fs-writefile-v0.1.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_fs_write.go, docs/script_spec.md
:: relatedTo: FS.ReadFile
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `FS.WriteFile` (v0.1.1)

* **Tool Name:** `FS.WriteFile` (v0.1.1)
* **Purpose:** Writes provided string content to a specified file within the designated sandbox directory. Creates parent directories if they don't exist and overwrites the file if it already exists.
* **NeuroScript Syntax:** `CALL FS.WriteFile(filepath: <String>, content: <String>)`
* **Arguments:**
    * `filepath` (String): Required. The relative path to the file within the designated sandbox directory. Absolute paths or paths attempting to traverse outside the sandbox are forbidden.
    * `content` (String): Required. The string content to be written to the file.
* **Return Value:** (String)
    * On success: The literal string "OK". (Accessible via `LAST` after the `CALL`).
    * On error: A string describing the failure (e.g., "WriteFile path error...", "WriteFile mkdir failed...", "WriteFile failed for..."). (Accessible via `LAST` after the `CALL`). The underlying Go error is logged internally if a logger is configured.
* **Behavior:**
    1.  Validates that exactly two arguments (`filepath`, `content`), both of type String, are provided.
    2.  Retrieves the interpreter's configured `sandboxDir`.
    3.  Uses the `SecureFilePath` helper to validate the `filepath` argument against the `sandboxDir`. Checks if the path is relative and stays within the sandbox.
    4.  If `SecureFilePath` returns an error (invalid path, outside sandbox), the tool returns an error message string.
    5.  If the path is secure, the tool determines the parent directory of the validated absolute path.
    6.  It attempts to create all necessary parent directories using `os.MkdirAll` with default permissions (0755).
    7.  If creating directories fails, the tool returns an error message string.
    8.  If directories exist or are created successfully, the tool attempts to write the `content` string to the validated absolute file path using `os.WriteFile`, overwriting any existing file content. Files are written with permissions 0644.
    9.  If writing the file fails, the tool returns an error message string.
    10. If the file is written successfully, the tool returns the string "OK".
* **Security Considerations:**
    * This tool is restricted by the interpreter's sandbox directory (`sandboxDir`). It cannot write files outside this directory.
    * Relies on `SecureFilePath` for path validation. Flaws in validation could permit writing to unintended locations within the sandbox.
    * Has the ability to overwrite existing files within the sandbox. Use with caution, especially in automated scripts.
    * Creates directories within the sandbox as needed.
    * File write permissions are ultimately governed by the OS user running the interpreter, though the tool attempts to set mode 0644.
* **Examples:** (Updated to conform to `script_spec.md` v1.1.0)
    ```neuroscript
    # Example 1: Write a simple status message to a file
    SET log_message = "Script finished successfully."
    CALL FS.WriteFile("logs/script_run.log", log_message)
    SET write_status = LAST
    IF write_status != "OK" THEN
        EMIT "Error writing log file: " + write_status
    ELSE
        EMIT "Log file write successful."
    ENDBLOCK

    # Example 2: Create a config file in a subdirectory
    SET config_data = "{ \"setting\": \"value\", \"enabled\": true }" # JSON-like string content
    CALL FS.WriteFile("config/app_settings.json", config_data)
    SET config_status = LAST
    IF config_status == "OK" THEN
        EMIT "Config file created."
    ELSE
        EMIT "Failed to create config: " + config_status
    ENDBLOCK

    # Example 3: Attempt to write outside the sandbox (will fail)
    CALL FS.WriteFile("../../system_file.conf", "hacked_content")
    SET bad_write_status = LAST
    # The status will contain an error message from SecureFilePath
    EMIT "Result of attempting restricted write: " + bad_write_status
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_fs_write.go`
    * Function: `toolWriteFile`
    * Key Go Packages: `os`, `fmt`, `path/filepath`
    * Helpers: `core.SecureFilePath`
    * Registration: Called by `registerFsFileTools` within `pkg/core/tools_fs.go` (which itself is called by `registerCoreTools` in `pkg/core/tools_register.go`).