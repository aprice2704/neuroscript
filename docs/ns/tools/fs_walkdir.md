:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-fs-walkdir-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_fs_walk.go, pkg/core/tools_fs_utils.go, docs/script_spec.md
:: relatedTo: FS.ListDirectory
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `FS.WalkDir` (v0.1)

* **Tool Name:** `FS.WalkDir` (v0.1) (Note: Registered internally via spec named `TOOL.WalkDir`)
* **Purpose:** Recursively walks a directory tree starting from a specified path within the sandbox. It returns a flat list containing information about every file and subdirectory encountered during the traversal (excluding the starting directory itself).
* **NeuroScript Syntax:** `CALL FS.WalkDir(path: <String>)`
* **Arguments:**
    * `path` (String): Required. The relative path within the sandbox of the directory where the recursive walk should begin.
* **Return Value:** (List | nil)
    * On success: A List where each element is a Map representing a file or subdirectory found within the specified `path`. Each Map contains the following keys:
        * `name` (String): The base name of the file or directory.
        * `path` (String): The relative path of the entry *from* the directory specified in the `path` argument. Uses forward slashes (`/`) as separators.
        * `isDir` (Boolean): `true` if the entry is a directory, `false` otherwise.
        * `size` (Number): The size of the file in bytes. Typically 0 for directories.
        * `modTime` (String): The last modification time in RFC3339 format (e.g., `"2025-04-22T20:36:00Z"`).
    * On error (e.g., path validation fails, start path is not a directory, permission error during walk): Returns `nil`. Error details are logged internally. (Accessible via `LAST` after the `CALL`).
    * Special Case: If the provided start `path` does not exist, the tool returns `nil` without logging an internal error (consistent with no files found).
* **Behavior:**
    1.  Validates that exactly one argument (`path`) of type String is provided and is not empty.
    2.  Retrieves the interpreter's `sandboxDir`.
    3.  Uses `SecureFilePath` to validate the input `path` against the `sandboxDir`. Returns `nil` if validation fails.
    4.  Checks if the validated path exists using `os.Stat`. If it doesn't exist, returns `nil`. If other stat errors occur, returns `nil`.
    5.  Checks if the validated path is a directory. Returns `nil` if it's not.
    6.  Initializes an empty list to store results.
    7.  Uses `filepath.WalkDir` to recursively traverse the directory tree starting from the validated path.
    8.  For each item encountered during the walk (file or directory):
        * Skips the entry if it's the starting directory itself.
        * Retrieves file metadata (`name`, `isDir`, `size`, `modTime`).
        * Calculates the item's path relative to the starting directory.
        * Creates a Map containing the retrieved metadata (`name`, `path`, `isDir`, `size`, `modTime`).
        * Appends this Map to the results list.
        * If an error occurs while accessing an item or its metadata during the walk (e.g., permission denied), the walk is likely terminated, and the tool proceeds to step 9.
    9.  If `filepath.WalkDir` completes without returning an error, returns the populated list of Maps.
    10. If `filepath.WalkDir` returns an error (due to errors in the walking process itself or errors returned by the callback function, like permission errors), the tool logs the error and returns `nil`.
* **Security Considerations:**
    * The walk is confined to the directory tree starting within the validated sandbox path due to the initial `SecureFilePath` check.
    * Reads only file/directory metadata, not content.
    * `filepath.WalkDir` does not follow symbolic links, preventing escapes via symlinks outside the initial validated path.
    * Permission errors encountered during the walk can prevent parts of the directory tree from being listed and may cause the tool to return `nil` prematurely.
* **Examples:**
    ```neuroscript
    # Prerequisite: Create some nested files/dirs
    CALL FS.Mkdir("walktest/subdir")
    CALL FS.WriteFile("walktest/file1.txt", "Content A")
    CALL FS.WriteFile("walktest/subdir/file2.js", "Content B")

    # Example 1: Walk the directory
    EMIT "Walking directory 'walktest'..."
    CALL FS.WalkDir("walktest")
    SET walk_results = LAST

    IF walk_results == nil THEN
      EMIT "Error walking 'walktest' or directory is empty/doesn't exist."
    ELSE
      EMIT "Walk Results:"
      FOR EACH entry IN walk_results DO
        EMIT " - Path: " + entry["path"] + ", IsDir: " + entry["isDir"] + ", Size: " + entry["size"]
        # Expected output (order may vary):
        # - Path: file1.txt, IsDir: false, Size: 9
        # - Path: subdir, IsDir: true, Size: 0
        # - Path: subdir/file2.js, IsDir: false, Size: 9
      ENDBLOCK
    ENDBLOCK

    # Example 2: Walk a non-existent directory
    EMIT "Walking non-existent directory 'no_such_dir'..."
    CALL FS.WalkDir("no_such_dir")
    SET non_existent_results = LAST
    IF non_existent_results == nil THEN
      EMIT "WalkDir returned nil for non-existent path (expected)."
    ELSE
      EMIT "WalkDir unexpectedly returned results for non-existent path."
    ENDBLOCK
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_fs_walk.go`
    * Function: `toolWalkDir`
    * Spec Name: `TOOL.WalkDir` (in `tools_fs_utils.go`)
    * Key Go Packages: `os`, `fmt`, `path/filepath`, `io/fs`, `time`, `errors`
    * Helpers: `core.SecureFilePath`
    * Registration: Registered via `registerFsUtilTools` slice in `pkg/core/tools_fs_utils.go`. Returns `[]map[string]interface{}, error`. Handles errors to return `nil` list.