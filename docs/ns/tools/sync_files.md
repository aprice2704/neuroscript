 :: type: NSproject
 :: subtype: tool_specification
 :: version: 0.1.0
 :: id: tool-spec-sync-files-v0.1
 :: status: draft
 :: dependsOn: [docs/ns/tools/tool_spec_structure.md](./tool_spec_structure.md), [pkg/core/tools_file_api.go](../../pkg/core/tools_file_api.go), [docs/ns/tools/list_api_files.md](./list_api_files.md), [docs/ns/tools/upload_file.md](./upload_file.md), [docs/ns/tools/delete_api_file.md](./delete_api_file.md)
 :: howToUpdate: Update when adding support for other directions ("down", "both") or changing behavior/return value.

 # Tool Specification: TOOL.SyncFiles (v0.1)

 * **Tool Name:** TOOL.SyncFiles (v0.1)
 * **Purpose:** Synchronizes files between a specified local directory within the sandbox and the Gemini File API storage. This version only supports the "up" direction (local to API).
 * **NeuroScript Syntax:**
   ```neuroscript
   syncStatsMap := TOOL.SyncFiles(direction, localDirectory, [filterGlobPattern])
   ```
 * **Arguments:**
    * `direction` (String): Required. Specifies the synchronization direction. Currently, only `"up"` is supported. Case-insensitive.
    * `local_dir` (String): Required. The path to the local directory within the NeuroScript sandbox environment to synchronize. Subject to `SecureFilePath` validation. Must be a directory.
    * `filter_pattern` (String): Optional. A glob pattern (e.g., `"*.txt"`, `"images/*"`, `"main.go"`) used to filter files based on their basename within the `local_dir`. If omitted or null, all files are considered for synchronization.
 * **Return Value:** (Map)
    * On success or partial success: A map containing statistics about the sync operation. Keys include:
        * `files_scanned` (Number): Total local files encountered during the walk.
        * `files_filtered` (Number): Local files skipped due to the `filter_pattern`.
        * `files_uploaded` (Number): Files successfully uploaded (new or changed).
        * `files_deleted` (Number): API files successfully deleted (because they were removed locally).
        * `files_up_to_date` (Number): Local files found to be identical to their API counterparts (based on hash).
        * `upload_errors` (Number): Count of errors encountered during upload attempts.
        * `delete_errors` (Number): Count of errors encountered during delete attempts.
        * `list_errors` (Number): Count of errors encountered when initially listing API files (sync aborts if > 0).
        * `walk_errors` (Number): Count of non-critical errors during local directory walk (e.g., permission denied on a sub-item).
        * `hash_errors` (Number): Count of errors calculating local file hashes.
    * On critical failure: An error (e.g., invalid `local_dir`, API client not initialized, invalid `filter_pattern`, failure to list initial API files). A Go-level error will also be returned to the interpreter. Non-critical errors during individual uploads/deletes are reflected in the stats map but do not cause a Go error return.
 * **Behavior (`direction="up"`):**
    1. Validates arguments (`direction`, `local_dir` type/security, `filter_pattern` type). Aborts with an error if invalid.
    2. Calls `TOOL.ListAPIFiles` internally to get the current state of files in the API. Creates a map of these files keyed by their `displayName`. Aborts with an error if listing fails.
    3. Recursively walks the specified `local_dir`. Keeps track of local relative paths seen.
    4. For each local file encountered:
        a. Skips if it doesn't match the `filter_pattern` (if provided). Uses glob matching on the file's base name.
        b. Calculates the SHA256 hash of the local file content (handles zero-byte files).
        c. Looks up the file's relative path (used as display name) in the API file map.
        d. Compares the local hash with the `sha256Hash` from the API file info.
        e. If the file is not in the API map, or if the hashes differ, calls `TOOL.UploadFile` internally to upload the local file (using the relative path as the display name). Logs errors and increments stats but continues.
        f. If the file exists in the API map and hashes match, increments the `files_up_to_date` stat.
    5. After the walk, iterates through the API file map obtained in step 2.
    6. For each API file, checks if its `displayName` was seen during the local walk AND if it would *not* have been skipped by the filter pattern.
    7. If an API file's `displayName` was not seen locally (and wouldn't be filtered), calls `TOOL.DeleteAPIFile` internally to delete the file from the API. Logs errors and increments stats but continues.
    8. Returns the final statistics map.
 * **Security Considerations:**
    * Requires a valid `GEMINI_API_KEY` with permissions for List, Upload, Get, and Delete file operations.
    * `local_dir` is validated using `SecureFilePath`. The walk is confined within this directory.
    * Network access to Google Cloud services is required.
    * Reads all files within the specified `local_dir` (subject to filtering) and uploads changed/new ones to Google Cloud storage.
    * Deletes files from Google Cloud storage if they are removed locally (subject to filtering).
 * **Examples:**
   ```neuroscript
   // Sync everything in the 'project_files' subdir up to the API
   stats1 := TOOL.SyncFiles("up", "project_files")
   IO.Print("Sync 1 Stats:", stats1)

   // Sync only Go files in the 'src' subdir up to the API
   stats2 := TOOL.SyncFiles("up", "src", "*.go")
   IO.Print("Sync 2 Stats (Go files):", stats2)
   IF stats2["upload_errors"] > 0 OR stats2["delete_errors"] > 0 THEN
       IO.Print("  WARNING: Sync encountered errors.")
   END
   ```
 * **Go Implementation Notes:**
    * Implemented in `pkg/core/tools_file_api.go` as `toolSyncFiles`.
    * Uses `filepath.WalkDir` for local traversal.
    * Uses `filepath.Match` for glob pattern matching.
    * Internally calls `toolListAPIFiles`, `toolUploadFile`, `toolDeleteAPIFile`.
    * Needs helper `calculateFileHash` (adapted from `gensync`).
    * Registered via `registerFileAPITools`.
    * Currently only supports `direction="up"`.