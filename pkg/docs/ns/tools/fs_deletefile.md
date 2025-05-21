:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-fs-deletefile-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_fs_delete.go, docs/script_spec.md
:: relatedTo: FS.MoveFile, FS.WriteFile, FS.Mkdir
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `FS.DeleteFile` (v0.1)

* **Tool Name:** `FS.DeleteFile` (v0.1)
* **Purpose:** Deletes a specified file or an *empty* directory within the sandbox. It does *not* delete non-empty directories.
* **NeuroScript Syntax:** `CALL FS.DeleteFile(path: <String>)`
* **Arguments:**
    * `path` (String): Required. The relative path (within the sandbox) of the file or empty directory to be deleted.
* **Return Value:** (String)
    * Returns the literal string `"OK"` if: (Accessible via `LAST` after the `CALL`)
        * The file or empty directory was successfully deleted.
        * The specified path did *not* exist (considered idempotent success).
    * Returns an error message string if: (Accessible via `LAST` after the `CALL`)
        * Path validation fails.
        * The path refers to a non-empty directory.
        * A permission error or other OS-level error occurs during deletion.
* **Behavior:**
    1.  Validates that exactly one argument (`path`) of type String is provided and is not empty.
    2.  Retrieves the interpreter's `sandboxDir`.
    3.  Uses `SecureFilePath` to validate the `path` argument against the `sandboxDir`. If validation fails, returns an error message string.
    4.  Attempts to delete the file or directory at the validated absolute path using `os.Remove`.
    5.  If `os.Remove` returns an error:
        * Checks if the error indicates the path does not exist (`os.ErrNotExist`). If so, returns `"OK"`.
        * Checks if the error message indicates the directory is not empty (using string checks for common OS messages). If so, returns an error message string detailing this.
        * For any other error (e.g., permissions), returns an error message string containing the OS error details.
    6.  If `os.Remove` succeeds without error, returns `"OK"`.
* **Security Considerations:**
    * Restricted by the interpreter's sandbox directory (`sandboxDir`) via `SecureFilePath`. Cannot delete files or directories outside the sandbox.
    * Performs filesystem modification (deletion). Deletions are generally permanent.
    * Explicitly prevents deletion of non-empty directories, adding a layer of safety against accidental recursive deletion.
    * Relies on the underlying operating system's permissions for deletion capabilities.
* **Examples:**
    ```neuroscript
    # Example 1: Create and delete a file
    CALL FS.WriteFile("temp_file_to_delete.txt", "Delete me")
    SET write_ok = LAST

    CALL FS.DeleteFile("temp_file_to_delete.txt")
    SET delete_status = LAST

    IF delete_status == "OK" THEN
        EMIT "Successfully deleted temp_file_to_delete.txt"
    ELSE
        EMIT "Error deleting file: " + delete_status
    ENDBLOCK

    # Example 2: Attempt to delete a non-existent file (should return OK)
    CALL FS.DeleteFile("this_does_not_exist.tmp")
    SET non_exist_status = LAST
    EMIT "Status for deleting non-existent file: " + non_exist_status # Expect "OK"

    # Example 3: Create a directory, add a file, try to delete (will fail)
    CALL FS.Mkdir("dir_to_delete")
    CALL FS.WriteFile("dir_to_delete/cannot_delete_me.txt", "I prevent deletion")

    EMIT "Attempting to delete non-empty directory..."
    CALL FS.DeleteFile("dir_to_delete")
    SET non_empty_delete_status = LAST
    IF non_empty_delete_status != "OK" THEN
        EMIT "Deletion failed as expected: " + non_empty_delete_status
        # Expected error message like: Failed to delete 'dir_to_delete': ...directory not empty...
    ELSE
        EMIT "Deletion unexpectedly succeeded for non-empty directory."
    ENDBLOCK

    # Example 4: Delete the file inside, then delete the now-empty directory
    CALL FS.DeleteFile("dir_to_delete/cannot_delete_me.txt")
    SET delete_inner_status = LAST
    EMIT "Deleted inner file status: " + delete_inner_status

    CALL FS.DeleteFile("dir_to_delete")
    SET empty_dir_delete_status = LAST
    EMIT "Status for deleting empty directory: " + empty_dir_delete_status # Expect "OK"
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_fs_delete.go`
    * Function: `toolDeleteFile`
    * Spec Name: `DeleteFile`
    * Key Go Packages: `fmt`, `os`, `errors`, `strings`
    * Helpers: `core.SecureFilePath`
    * Registration: Registered by `registerFsDeleteTools` within `pkg/core/tools_fs.go`. Returns `string, error`. Handles `ErrNotExist` as success ("OK") and checks for directory-not-empty errors specifically.