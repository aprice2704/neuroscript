# Plan: Refactor pkg/core into Sub-packages using NeuroScript

**Goal:** Automate the process of splitting the `pkg/core` Go package into logical sub-packages (e.g., `tools/fs`, `tools/go/ast`, `internal`) using an interactive NeuroScript script.

**Required NeuroScript Tools:**

* **Existing:**
    * `TOOL.GitNewBranch`, `TOOL.GitAdd`, `TOOL.GitCommit`
    * `TOOL.ListDirectory`
    * `TOOL.StringSplit`, `TOOL.StringContains`, `TOOL.StringPrefix`, etc. (for filename analysis)
    * `TOOL.Mkdir`
    * `TOOL.GoParseFile`, `TOOL.GoModifyAST`, `TOOL.GoFormatASTNode`
    * `TOOL.GoModTidy`, `TOOL.GoBuild`, `TOOL.GoTest`
    * `IO.Print`
    * List/Map manipulation functions
* **Needed (To be added/implemented):**
    * `IO.Input(prompt)`: To get user confirmation/input.
    * `TOOL.MoveFile(src, dest)`: To move files reliably.
    * `TOOL.GoUpdateImportsForMovedPackage(oldImportPath, newImportPathPrefix, scope)`: To intelligently update import paths across the project after files are moved.

**NeuroScript (`refactor-to-sub-packages.ns.txt`) Logic:**

1.  **Setup:**
    * Define constants: source directory (`pkg/core`), target base (`pkg/core/tools`), default internal dir (`pkg/core/internal`), project module path (`neuroscript`).
    * Generate a unique branch name (e.g., `refactor/core-split-<timestamp>`).
    * Create and checkout the new Git branch using `TOOL.GitNewBranch` and `TOOL.GitCheckout`.

2.  **Analyze Files and Propose Mapping:**
    * List all `.go` files in `pkg/core` using `TOOL.ListDirectory`.
    * Iterate through the file list:
        * Use string manipulation to propose a target sub-package based on filename conventions (e.g., `tools_fs*.go` -> `tools/fs`, `tools_go_ast*.go` -> `tools/go/ast`, `interpreter*.go` -> `internal`, `ast*.go` -> `internal`, etc.).
        * Files that don't match specific tool patterns could default to `internal` or prompt the user.
    * Construct a map (`proposed_moves`) where keys are current paths and values are proposed new paths.

3.  **User Confirmation:**
    * Format the `proposed_moves` map into a human-readable string.
    * Use `IO.Print` to display the proposed moves.
    * **(Use `IO.Input`)** Prompt the user to confirm (`y/n`) or provide a path to an edited mapping file (e.g., JSON/ND). (*Requires `IO.Input` tool*).
    * If rejected, exit. If edits are provided, load and use the edited map.

4.  **Execute File Moves:**
    * Create all necessary target subdirectories using `TOOL.Mkdir` based on the values in the confirmed `proposed_moves` map.
    * Iterate through the `proposed_moves` map:
        * **(Use `TOOL.MoveFile`)** Move the file from the source path (key) to the destination path (value). (*Requires `TOOL.MoveFile` tool*).

5.  **Update Package Declarations:**
    * Iterate through the *new* locations of the moved files (values in `proposed_moves` map):
        * Determine the correct new package name from the directory path (e.g., `pkg/core/tools/fs` -> `fs`, `pkg/core/tools/go/ast` -> `ast`, `pkg/core/internal` -> `internal`).
        * Parse the Go file using `TOOL.GoParseFile`.
        * Use `TOOL.GoModifyAST` with the 'Change Package Declaration' directive to set the new package name.
        * Format the modified AST using `TOOL.GoFormatASTNode`.
        * Write the updated content back to the file using `TOOL.WriteFile`.

6.  **Update Import Paths (Complex Step):**
    * **(Use `TOOL.GoUpdateImportsForMovedPackage` or AI Logic)** This is the most complex step requiring the new/enhanced tool or significant AI reasoning.
    * The goal is to scan *all* `.go` files in the project (`.`).
    * For each file, identify imports of the original `neuroscript/pkg/core`.
    * Analyze the code to see which symbols from `core` are actually used.
    * Determine the new sub-package where each used symbol now resides (e.g., if `core.ReadFile` was used, its new import path is `neuroscript/pkg/core/tools/fs`).
    * Update the import statements accordingly, potentially adding multiple new import lines for different sub-packages if symbols from different sub-packages were used.
    * *Initial Sketch Expectation:* The NeuroScript might just call the placeholder tool: `TOOL.GoUpdateImportsForMovedPackage("neuroscript/pkg/core", "neuroscript/pkg/core", ".")`. The tool itself (or the AI executing the script) needs to handle the detailed logic.

7.  **Verification:**
    * Run `TOOL.GoModTidy` to clean up `go.mod`/`go.sum`.
    * Run `TOOL.GoBuild("./...")` to check for compilation errors.
    * Run `TOOL.GoTest("./...")` to check if tests pass.

8.  **Commit:**
    * If Build and Test succeed:
        * Use `TOOL.GitAdd .` to stage all changes.
        * Use `TOOL.GitCommit` with a descriptive message (e.g., "Refactor: Split pkg/core into sub-packages").
        * `IO.Print` success message with branch name.
    * If Build or Test fail:
        * `IO.Print` failure message, indicating the branch contains the broken state for debugging.

9.  **Error Handling:** Throughout the script, check for errors returned by tools and provide informative messages using `IO.Print`. Potentially offer to stop or revert using Git commands if critical steps fail.