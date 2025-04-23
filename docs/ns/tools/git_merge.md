:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-git-gitmerge-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_git_register.go, pkg/core/tools_git.go, docs/script_spec.md
:: relatedTo: Git.GitNewBranch, Git.GitCheckout, Git.GitStatus
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `Git.GitMerge` (v0.1)

* **Tool Name:** `Git.GitMerge` (v0.1)
* **Purpose:** Merges the history of a named branch into the current branch. This is equivalent to running `git merge <branch_name>` within the project's sandbox directory. It reports merge conflicts via the error return value.
* **NeuroScript Syntax:** `CALL Git.GitMerge(branch_name: <String>)`
* **Arguments:**
    * `branch_name` (String): Required. The name of the branch whose history should be merged into the currently checked-out branch. It cannot be an empty string.
* **Return Value:** (String)
    * On success (clean merge): A string indicating the merge was successful, potentially including output from the `git merge` command. Example: `"Successfully merged branch 'feature/login'.\nOutput:\nMerge made by the 'recursive' strategy...\n"` (Accessible via `LAST` after the `CALL`).
    * On failure or conflict: A string describing the failure. This could be due to an empty `branch_name`, or errors from the underlying `git merge` command execution (e.g., "failed to merge branch '...': <output from git merge>"). The output from `git merge` will be included in the error message if conflicts occur (e.g., mentioning "Automatic merge failed; fix conflicts and then commit the result."). (Accessible via `LAST` after the `CALL`).
* **Behavior:**
    1.  Validates that exactly one argument (`branch_name`) of type String is provided.
    2.  Validates that the `branch_name` is not empty. If it is, returns an error message string.
    3.  Retrieves the interpreter's `sandboxDir`.
    4.  Constructs a command equivalent to `git merge "<branch_name>"`.
    5.  Executes the `git merge` command within the `sandboxDir` using the `toolExec` helper function.
    6.  If the command execution fails (e.g., `git` not found, branch does not exist, merge conflicts occur), returns an error message string containing details from `toolExec` (which includes Git's standard error/output for conflicts).
    7.  If the command execution succeeds (a clean merge), returns a success message string, including the merged branch name and any standard output from the `git merge` command.
* **Security Considerations:**
    * Executes the `git merge` command within the configured sandbox directory.
    * Modifies the Git repository state: potentially creates a merge commit, updates the current branch reference, and modifies files in the working directory and index.
    * Does not automatically resolve merge conflicts; it relies on the user or subsequent script steps to handle conflicts reported in the error output.
    * Relies on the `toolExec` helper for command execution.
    * Assumes the `git` executable is available in the environment's PATH where the NeuroScript interpreter is running.
* **Examples:**
    ```neuroscript
    # Setup: Create a main branch commit, then a feature branch with another commit
    CALL FS.WriteFile("base.txt", "Line 1")
    CALL Git.GitAdd(["base.txt"])
    CALL Git.GitCommit("Initial commit on main")
    CALL Git.GitNewBranch("feature/add-line2")
    CALL FS.WriteFile("base.txt", "Line 1\nLine 2") # Modify on feature branch
    CALL Git.GitAdd(["base.txt"])
    CALL Git.GitCommit("Add line 2 on feature branch")
    CALL Git.GitCheckout("main") # Switch back to main

    # Example 1: Perform a clean merge (fast-forward or simple merge)
    EMIT "Attempting to merge 'feature/add-line2' into 'main'..."
    CALL Git.GitMerge("feature/add-line2")
    SET merge_status = LAST
    EMIT "Merge status: " + merge_status
    # Expected output should indicate success

    # Example 2: Attempt to merge a non-existent branch
    EMIT "Attempting to merge non-existent branch..."
    CALL Git.GitMerge("no-such-branch")
    SET non_exist_merge_status = LAST
    EMIT "Status for merging non-existent branch: " + non_exist_merge_status
    # Expect error like: "failed to merge branch 'no-such-branch': fatal: 'no-such-branch' does not point to a commit"

    # Example 3: Simulate a conflict scenario (manual setup needed outside script usually)
    # Assume 'main' has a change conflicting with 'feature/add-line2' before merge attempt
    # CALL Git.GitMerge("feature/add-line2")
    # SET conflict_status = LAST
    # EMIT "Status for conflicting merge: " + conflict_status
    # Expect error containing text like: "Automatic merge failed; fix conflicts and then commit the result."
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_git.go`
    * Function: `toolGitMerge`
    * Spec Name: `GitMerge` (in `pkg/core/tools_git_register.go`)
    * Key Go Packages: `fmt`
    * Helpers: `core.toolExec`, `core.ErrValidationArgValue`
    * Registration: Registered by `registerGitTools` within `pkg/core/tools_register.go`. Returns `string, error`. Merge conflicts are reported via the wrapped error from `toolExec`.