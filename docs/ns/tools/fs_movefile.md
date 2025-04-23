:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-fs-movefile-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_fs_move.go, docs/script_spec.md
:: relatedTo: FS.WriteFile, FS.Mkdir, FS.DeletePath
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `FS.MoveFile` (v0.1)

* **Tool Name:** `FS.MoveFile` (v0.1)
* **Purpose:** Moves or renames a file or directory from a source path to a destination path within the sandbox. This operation fails if the destination path already exists.
* **NeuroScript Syntax:** `CALL FS.MoveFile(source: <String>, destination: <String>)`
* **Arguments:**
    * `source` (String): Required. The relative path (within the sandbox) of the existing file or directory to be moved/renamed.
    * `destination` (String): Required. The desired new relative path (within the sandbox) for the file or directory. This path must *not* already exist.
* **Return Value:** (Map)
    * Returns a Map containing a single key: `error`. (Accessible via `LAST` after the `CALL`).
        * On success: The value associated with the `error` key is `nil`. Example: `{"error": nil}`.
        * On failure: The value associated with the `error` key is a String describing the reason for failure (e.g., path validation error, source not found, destination already exists, OS rename error). Example: `{"error": "Destination path 'new/data.txt' already exists."}`.
* **Behavior:**
    1.  Validates that exactly two arguments (`source`, `destination`), both non-empty Strings, are provided.
    2.  Retrieves the interpreter's `sandboxDir`.
    3.  Uses `SecureFilePath` to validate both the `source` and `destination` paths against the `sandboxDir`. If either validation fails, returns a Map `{"error": "Validation error message"}`.
    4.  Checks if the validated `source` path exists using `os.Stat`. If it does not exist or another error occurs, returns a Map `{"error": "Source path error message"}`.
    5.  Checks if the validated `destination` path exists using `os.Stat`.
        * If it *does* exist, returns a Map `{"error": "Destination path '...' already exists."}`.
        * If an error *other than* "file does not exist" occurs, returns a Map `{"error": "Error checking destination path..."}`.
    6.  If the source exists and the destination does not exist, attempts to perform the move/rename operation using `os.Rename` on the validated absolute paths.
    7.  If `os.Rename` fails (e.g., permission denied, cross-device link error), returns a Map `{"error": "Failed to move/rename..."}`.
    8.  If `os.Rename` succeeds, returns the Map `{"error": nil}`.
* **Security Considerations:**
    * Restricted by the interpreter's sandbox directory (`sandboxDir`) via `SecureFilePath` validation on both source and destination paths.
    * Performs filesystem modification (move/rename).
    * Explicitly prevents overwriting existing files/directories at the destination path.
    * The underlying `os.Rename` behavior might vary slightly between operating systems or across different filesystem mounts (though usually atomic within the same filesystem).
* **Examples:**
    ```neuroscript
    # Example 1: Rename a file
    CALL FS.WriteFile("old_name.txt", "File content")
    SET write_ok = LAST

    CALL FS.MoveFile("old_name.txt", "new_name.txt")
    SET move_result = LAST

    IF move_result["error"] == nil THEN
        EMIT "File renamed successfully from old_name.txt to new_name.txt"
    ELSE
        EMIT "Error renaming file: " + move_result["error"]
    ENDBLOCK

    # Example 2: Move a file into a directory (directory must exist)
    CALL FS.Mkdir("data_files")
    SET mkdir_ok = LAST

    CALL FS.MoveFile("new_name.txt", "data_files/final_name.txt")
    SET move_into_dir_result = LAST
    IF move_into_dir_result["error"] == nil THEN
      EMIT "File moved into directory successfully."
    ELSE
      EMIT "Error moving file into directory: " + move_into_dir_result["error"]
    ENDBLOCK

    # Example 3: Attempt to move onto an existing file (will fail)
    CALL FS.WriteFile("another_file.txt", "Some other content")
    SET write_another_ok = LAST

    EMIT "Attempting to move onto existing file..."
    CALL FS.MoveFile("data_files/final_name.txt", "another_file.txt")
    SET overwrite_attempt_result = LAST
    IF overwrite_attempt_result["error"] != nil THEN
        EMIT "Move failed as expected: " + overwrite_attempt_result["error"]
        # Expected error: Destination path 'another_file.txt' already exists.
    ELSE
        EMIT "Move unexpectedly succeeded when destination existed."
    ENDBLOCK
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_fs_move.go`
    * Function: `toolMoveFile`
    * Spec Name: `MoveFile`
    * Key Go Packages: `fmt`, `os`, `errors`
    * Helpers: `core.SecureFilePath`
    * Registration: Registered by `registerFsMoveTools` within `pkg/core/tools_fs.go`. Returns `map[string]interface{}, error`.