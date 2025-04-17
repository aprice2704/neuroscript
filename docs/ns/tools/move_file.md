# Tool Specification Structure Template

## Tool Specification: `TOOL.MoveFile` (v0.1)

* **Tool Name:** `TOOL.MoveFile`
* **Purpose:** Moves or renames a file or directory within the allowed filesystem sandbox defined by `SecureFilePath`.
* **NeuroScript Syntax:** `CALL TOOL.MoveFile(source, destination)`
* **Arguments:**
    * `source` (String): Required. The path to the existing file or directory to move, relative to the sandboxed root.
    * `destination` (String): Required. The new path for the file or directory, relative to the sandboxed root.
* **Return Value:** (Map)
    * `error` (String | null): A string describing the error if the move/rename fails (e.g., source not found, destination exists, permission denied), otherwise `null`.
* **Behavior:**
    1.  Validates that exactly two string arguments (`source`, `destination`) are provided. Returns `ErrValidationArgCount` if not.
    2.  Validates *both* the `source` and `destination` paths using `SecureFilePath` [cite: uploaded:neuroscript_small/pkg/core/security.go] to ensure they resolve within the designated secure working directory sandbox. Returns an appropriate error (e.g., `ErrSecurityPathViolation`) if validation fails.
    3.  Checks if the resolved `source` path exists on the filesystem. Returns an error (e.g., `"Source path does not exist"`) if it does not.
    4.  Checks if the resolved `destination` path *already exists* on the filesystem. To prevent accidental overwrites or ambiguous moves into directories, this tool **returns an error** (e.g., `"Destination path already exists"`) if the destination path exists. It does not support overwriting.
    5.  If all validations pass and the destination does not exist, attempts to move/rename the `source` to the `destination` using the operating system's rename functionality (e.g., `os.Rename` in Go).
    6.  Returns a map `{"error": null}` on success.
    7.  On failure (permission error, I/O error, etc. during `os.Rename`), returns a map `{"error": "descriptive message"}`.
* **Security Considerations:**
    * Relies entirely on the `SecureFilePath` validation to prevent moving files outside the intended sandbox. The implementation and configuration of `SecureFilePath` must be robust.
* **Examples:**
    ```neuroscript
    # Rename a file
    VAR result = CALL TOOL.MoveFile("old_name.txt", "new_name.txt")
    IF result.error != null THEN
        EMIT "Error renaming file: ", result.error
    END

    # Move a file into a subdirectory (assuming subdir exists)
    VAR result_move = CALL TOOL.MoveFile("report.txt", "reports/final_report.txt")
    IF result_move.error != null THEN
        EMIT "Error moving file: ", result_move.error
    END

    # Attempt to overwrite (will fail)
    CALL TOOL.WriteFile("existing.txt", "content")
    VAR overwrite_result = CALL TOOL.MoveFile("some_other_file.txt", "existing.txt")
    # overwrite_result.error should be non-null here
    EMIT "Overwrite attempt error: ", overwrite_result.error
    ```
* **Go Implementation Notes:**
    * Location: Likely `pkg/core/tools_fs.go`.
    * Use Go's `os.Rename(secureSourcePath, secureDestPath)` after performing checks.
    * Remember to check for source existence (`os.Stat`) and destination non-existence (`os.Stat` returning `fs.ErrNotExist`) before calling `os.Rename`.
    * Register the function (e.g., `toolFSMoveFile`) in `pkg/core/tools_register.go`.