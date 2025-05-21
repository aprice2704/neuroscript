:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-git-diff-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_git_register.go, pkg/core/tools_git.go, docs/script_spec.md
:: relatedTo: git.Status, git.Add
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `git.Diff` (v0.1)

* **Tool Name:** `git.Diff` (v0.1)
* **Purpose:** Shows the differences between the files in the working directory and the Git index (staging area). It highlights changes that have been made but not yet staged for commit. This is equivalent to running `git diff` (with no arguments) within the project's sandbox directory.
* **NeuroScript Syntax:** `CALL git.Diff()`
* **Arguments:** None.
* **Return Value:** (String)
    * On success:
        * If changes exist between the working tree and the index: Returns a string containing the textual diff output in the standard `git diff` format. (Accessible via `LAST` after the `CALL`).
        * If there are no changes between the working tree and the index: Returns the specific string `"GitDiff: No changes detected in the working tree."`. (Accessible via `LAST` after the `CALL`).
    * On failure: A string describing the failure, typically indicating a fatal error from the underlying `git diff` command execution (e.g., "GitDiff command failed: fatal: not a git repository..."). (Accessible via `LAST` after the `CALL`).
* **Behavior:**
    1.  Validates that no arguments are provided.
    2.  Retrieves the interpreter's `sandboxDir`.
    3.  Constructs a command equivalent to `git diff`.
    4.  Executes the `git diff` command within the `sandboxDir` using the `toolExec` helper function. Note that `git diff` typically exits successfully (exit code 0) even when differences are found.
    5.  If the command execution fails (e.g., `git` not found, not a Git repository), returns an error message string containing details from `toolExec`.
    6.  If the command execution succeeds, checks the standard output returned by `toolExec`:
        * If the output string is empty, returns the specific message `"GitDiff: No changes detected in the working tree."`.
        * If the output string is not empty, returns the raw diff output string directly.
* **Security Considerations:**
    * Executes the `git diff` command within the configured sandbox directory.
    * Reads Git repository metadata and file content from the working directory and index to generate the diff; does not modify the repository state or files.
    * Relies on the `toolExec` helper for command execution.
    * Assumes the `git` executable is available in the environment's PATH where the NeuroScript interpreter is running.
* **Examples:**
    ```neuroscript
    # Example 1: Show diff after modifying a tracked file
    # Assume README.md is tracked
    CALL FS.WriteFile("README.md", "Add a new line to the README.")
    SET write_ok = LAST

    EMIT "Showing diff after modifying README..."
    CALL git.Diff()
    SET diff_output = LAST
    EMIT "Git Diff Output:"
    EMIT diff_output
    # Expected output will show the diff for README.md

    # Example 2: Show diff when working tree is clean (matches index)
    # Assume previous change was added and committed, or no changes made
    CALL git.Diff()
    SET clean_diff_output = LAST
    EMIT "Git Diff Output (Clean): " + clean_diff_output
    # Expect: "GitDiff: No changes detected in the working tree."

    # Example 3: Add the change, then show diff (should be clean again)
    CALL git.GitAdd(["README.md"]) # Stage the change from Example 1
    SET add_ok = LAST

    CALL git.Diff()
    SET diff_after_add = LAST
    EMIT "Git Diff Output (After Add): " + diff_after_add
    # Expect: "GitDiff: No changes detected in the working tree."
    # Note: To see staged changes vs HEAD, one would need 'git diff --staged' functionality
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_git.go`
    * Function: `toolGitDiff`
    * Spec Name: `GitDiff` (in `pkg/core/tools_git_register.go`)
    * Key Go Packages: `fmt`
    * Helpers: `core.toolExec`
    * Registration: Registered by `registerGitTools` within `pkg/core/tools_register.go`. Returns `string, error`. Handles the "no changes" case specifically.