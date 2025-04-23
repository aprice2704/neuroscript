:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-fs-filehash-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_fs_hash.go, docs/script_spec.md
:: relatedTo: FS.ReadFile
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `FS.FileHash` (v0.1)

* **Tool Name:** `FS.FileHash` (v0.1)
* **Purpose:** Calculates the SHA256 hash of a specified file within the sandbox and returns it as a hex-encoded string. Useful for verifying file integrity or detecting changes.
* **NeuroScript Syntax:** `CALL FS.FileHash(filepath: <String>)`
* **Arguments:**
    * `filepath` (String): Required. The relative path (within the sandbox) of the file to hash.
* **Return Value:** (String)
    * On success: A string containing the lowercase hex-encoded SHA256 hash of the file's content. (Accessible via `LAST` after the `CALL`).
    * On error (e.g., path validation fails, file not found, path is a directory, read error): Returns an empty string (`""`). (Accessible via `LAST` after the `CALL`).
* **Behavior:**
    1.  Validates that exactly one argument (`filepath`) of type String is provided and is not empty.
    2.  Retrieves the interpreter's `sandboxDir`.
    3.  Uses `SecureFilePath` to validate the `filepath` argument against the `sandboxDir`. Returns `""` if validation fails.
    4.  Attempts to open the file at the validated absolute path. Returns `""` if opening fails (e.g., file not found, permission denied).
    5.  Checks if the opened path refers to a directory. If it is a directory, returns `""`.
    6.  Initializes a new SHA256 hasher.
    7.  Reads the entire file content and feeds it into the hasher using `io.Copy`. Returns `""` if there's an error during reading.
    8.  Computes the final SHA256 hash digest.
    9.  Formats the hash digest as a lowercase hexadecimal string.
    10. Returns the resulting hex string.
* **Security Considerations:**
    * Restricted by the interpreter's sandbox directory (`sandboxDir`) via `SecureFilePath`. Cannot access files outside the sandbox.
    * Reads the entire file content into memory implicitly during the `io.Copy` operation for hashing. Very large files could consume significant memory and processing time.
    * Does not modify the file system.
* **Examples:**
    ```neuroscript
    # Example 1: Hash a known file
    CALL FS.WriteFile("data.txt", "This is a test file.") # Create file
    SET write_ok = LAST

    CALL FS.FileHash("data.txt")
    SET file_hash = LAST

    IF file_hash == "" THEN
      EMIT "Error calculating hash for data.txt"
    ELSE
      EMIT "SHA256 Hash of data.txt: " + file_hash
      # Expected hash for "This is a test file." is typically:
      # c7be1ed902fb8dd4d48997c6452f5d7e509fbcdbe2808b16bcf4edce4c07d14e
    ENDBLOCK

    # Example 2: Hash a non-existent file
    CALL FS.FileHash("no_file_here.dat")
    SET missing_hash = LAST
    EMIT "Hash result for non-existent file: '" + missing_hash + "'" # Expect ""

    # Example 3: Attempt to hash a directory
    CALL FS.Mkdir("temp_dir")
    SET mkdir_ok = LAST
    CALL FS.FileHash("temp_dir")
    SET dir_hash = LAST
    EMIT "Hash result for directory: '" + dir_hash + "'" # Expect ""
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_fs_hash.go`
    * Function: `toolFileHash`
    * Spec Name: `FileHash`
    * Key Go Packages: `crypto/sha256`, `fmt`, `io`, `os`
    * Helpers: `core.SecureFilePath`
    * Registration: Registered by `registerFsHashTools` within `pkg/core/tools_fs.go`. Returns `string, error` internally, but handles errors to return the hash or `""`.