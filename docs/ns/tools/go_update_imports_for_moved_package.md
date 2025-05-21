# Tool Specification Structure Template

## Tool Specification: `TOOL.GoUpdateImportsForMovedPackage` (v0.1)

* **Tool Name:** `TOOL.GoUpdateImportsForMovedPackage`
* **Purpose:** Scans Go source files within a specified scope and automatically updates import paths for symbols that were moved during a package refactoring. Specifically designed for cases where symbols from a single original package (e.g., `pkg/core`) are split into multiple new sub-packages (e.g., `pkg/core/tools/fs`, `pkg/core/internal`).
* **NeuroScript Syntax:** `CALL TOOL.GoUpdateImportsForMovedPackage(refactored_package_path, scan_scope)`
* **Arguments:**
    * `refactored_package_path` (String): Required. The full import path of the *original* package whose contents have been moved (e.g., `"neuroscript/pkg/core"`). This path is used to identify relevant import statements and as a base for finding the new sub-packages.
    * `scan_scope` (String): Required. The directory path from which to start scanning recursively for `.go` files to update (e.g., `"."` for the project root). This path will be validated using `SecureFilePath`.
* **Return Value:** (Map)
    * `modified_files` (List[String] | null): A list of the full paths of files that were successfully modified. `null` if a catastrophic error occurred.
    * `skipped_files` (Map[String]String | null): A map where keys are paths of files that were parsed but not modified (e.g., didn't import the target package, no relevant symbols used) and values are the reason. `null` if a catastrophic error occurred.
    * `failed_files` (Map[String]String | null): A map where keys are paths of files that encountered errors during processing (parsing, analysis, modification, writing) and values are the error messages. `null` if a catastrophic error occurred.
    * `error` (String | null): A general error message if the tool encounters a fatal issue preventing it from operating (e.g., invalid arguments, critical failure scanning scope, unable to build symbol map), otherwise `null`.
* **Behavior:**
    1.  Validates arguments (`refactored_package_path`, `scan_scope`). Returns `ErrValidationArgCount` or `ErrValidationTypeMismatch` if invalid.
    2.  Validates `scan_scope` using `SecureFilePath`. Returns an error if validation fails.
    3.  **Build Symbol Map (Internal Complexity):** The tool must determine the mapping between exported symbols previously in `refactored_package_path` and their *new* package locations. It attempts this automatically:
        * It constructs the expected filesystem path corresponding to `refactored_package_path` (e.g., `$GOPATH/src/neuroscript/pkg/core` or `./pkg/core` in module mode).
        * It recursively scans `.go` files within the subdirectories of this path (e.g., `./pkg/core/tools/fs`, `./pkg/core/internal`).
        * It parses these files to identify exported symbols (functions, types, vars, consts) and maps them to their new full import path (e.g., build internal map `{"ReadFile": "neuroscript/pkg/core/tools/fs", "Interpreter": "neuroscript/pkg/core/internal", ...}`).
        * If this automatic detection fails or is ambiguous, the tool should return a top-level `error`. Robust implementation of this step is critical and challenging.
    4.  Recursively finds all `.go` files within the validated `scan_scope`, excluding files within the refactored package's own new subdirectories (to avoid self-modification issues if structure changes).
    5.  For each relevant `.go` file found:
        * Attempt to parse the file (`go/parser`). Record error in `failed_files` on failure.
        * Analyze the AST for import declarations matching `refactored_package_path` (potentially with aliases). If none found, record in `skipped_files`.
        * Identify all qualified identifiers used that originate from the `refactored_package_path` import (e.g., `core.ReadFile`, `alias.SomeType`).
        * For each used identifier's symbol (e.g., `ReadFile`), look up its new package path from the internal symbol map (Step 3). Record error in `failed_files` if a used symbol cannot be mapped.
        * Determine the set of *new* distinct import paths required based on the successfully mapped symbols.
        * Modify the AST's import block: Remove the original import spec for `refactored_package_path`. Add new import specs for the required new paths (using `astutil.AddImport` or similar). Handle potential naming collisions if aliases were not used originally.
        * **Note:** This version (v0.1) focuses *only* on correcting the `import` statements. It does **not** attempt to automatically update the qualifiers in the code (e.g., changing `core.ReadFile` to `fs.ReadFile`). This would require more complex analysis and transformation, potentially as a separate tool or later version. The resulting code might require manual qualifier updates or fail to compile until qualifiers are fixed.
        * Format the potentially modified AST (`go/format`).
        * Write the formatted code back to the original file path. Record error in `failed_files` on failure.
        * If modification and write were successful, add the file path to `modified_files`. If parsing occurred but no relevant imports/symbols were found or mapping failed, ensure it's in `skipped_files` or `failed_files` respectively.
    6.  Returns the result map containing `modified_files`, `skipped_files`, `failed_files`, and `error`.
* **Security Considerations:**
    * Relies on `SecureFilePath` for the `scan_scope`.
    * Modifies Go source files in place; **critical** to run this on a clean Git branch or with backups.
    * The complexity of Go AST parsing, analysis, and modification introduces risks of incorrect transformations. Thorough testing on diverse codebases is essential.
* **Examples:**
    ```neuroscript
    # Assume pkg/core was split into pkg/core/tools/fs and pkg/core/internal
    # Module path is "neuroscript"
    EMIT "Attempting to update imports related to pkg/core refactor..."
    VAR result = CALL TOOL.GoUpdateImportsForMovedPackage(
        refactored_package_path = "neuroscript/pkg/core",
        scan_scope = "."
    )

    IF result.error != null THEN
        EMIT "Fatal error running import update tool: ", result.error
    ELSE
        EMIT "Import update process completed."
        IF result.modified_files != null AND CALL List.Length(result.modified_files) > 0 THEN
            EMIT "Modified files:"
            FOREACH f IN result.modified_files
                EMIT "- ", f
            END
        END
        IF result.failed_files != null AND CALL Map.Size(result.failed_files) > 0 THEN
            EMIT "Files with errors:"
            FOREACH f, err IN result.failed_files
                EMIT "- ", f, ": ", err
            END
            EMIT "Manual review and potentially qualifier updates needed."
        END
    END
    ```
* **Go Implementation Notes:**
    * Location: Suggest `pkg/core/tools_go_refactor.go`.
    * Core Go Packages: `go/ast`, `go/parser`, `go/token`, `go/format`, `path/filepath`, `strings`, `golang.org/x/tools/go/packages` (potentially useful for symbol resolution), `golang.org/x/tools/go/ast/astutil`.
    * Implementation Detail: The symbol mapping (Step 3) is the hardest part. It likely requires parsing the destination packages first. Consider edge cases like symbols defined in `*_test.go` files.
    * Implementation Detail: Handling import aliases correctly during both detection and addition is crucial.
    * Register the function (e.g., `toolGoUpdateImports`) in `pkg/core/tools_register.go`.