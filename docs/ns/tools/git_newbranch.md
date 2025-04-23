:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-git-gitnewbranch-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_git_register.go, pkg/core/tools_git.go, docs/script_spec.md
:: relatedTo: Git.GitCheckout, Git.GitCommit
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `Git.GitNewBranch` (v0.1)

* **Tool Name:** `Git.GitNewBranch` (v0.1)
* **Purpose:** Creates a new branch starting from the current HEAD and immediately switches to it. This is equivalent to running `git checkout -b <branch_name>` within the project's sandbox directory.
* **NeuroScript Syntax:** `CALL Git.GitNewBranch(branch_name: <String>)`
* **Arguments:**
    * `branch_name` (String): Required. The name for the new branch. It cannot be empty and must not contain characters typically disallowed in Git branch names (e.g., whitespace, `\`, `:`, `*`, `?`, `"`, `<`, `>`, `|`, `~`, `^`).
* **Return Value:** (String)
    * On success: A string indicating the new branch was created and checked out successfully, potentially including output from the `git checkout -b` command. Example: `"Successfully created and checked out new branch 'feature/login'.\nOutput:\nSwitched to a new branch 'feature/login'\n"`. (Accessible via `LAST` after the `CALL`).
    * On error: A string describing the failure. This could be due to an invalid branch name (e.g., "branch name cannot be empty", "branch name '...' contains invalid characters") or errors from the underlying `git checkout -b` command execution (e.g., "failed to create new branch '...': fatal: A branch named '...' already exists"). (Accessible via `LAST` after the `CALL`).
* **Behavior:**
    1.  Validates that exactly one argument (`branch_name`) of type String is provided.
    2.  Validates that the `branch_name` is not empty. If it is, returns an error message string.
    3.  Validates that the `branch_name` does not contain invalid characters (` \t\n\\/:*?\"<>|~^`). If it does, returns an error message string.
    4.  Retrieves the interpreter's `sandboxDir`.
    5.  Constructs a command equivalent to `git checkout -b "<branch_name>"`.
    6.  Executes the `git checkout -b` command within the `sandboxDir` using the `toolExec` helper function.
    7.  If the command execution fails (e.g., `git` not found, branch already exists), returns an error message string containing details from `toolExec`.
    8.  If the command execution succeeds, returns a success message string, including the new branch name and any standard output from the `git checkout -b` command.
* **Security Considerations:**
    * Executes the `git checkout -b` command within the configured sandbox directory.
    * Modifies the Git repository state by creating a new branch reference and updating the HEAD to point to this new branch.
    * Includes basic validation against common invalid characters in branch names.
    * Relies on the `toolExec` helper for command execution.
    * Assumes the `git` executable is available in the environment's PATH where the NeuroScript interpreter is running.
* **Examples:**
    ```neuroscript
    # Example 1: Create a new feature branch
    SET new_branch = "feature/add-user-auth"
    CALL Git.GitNewBranch(new_branch)
    SET create_status = LAST
    EMIT "Create branch status: " + create_status

    # Example 2: Attempt to create a branch that already exists (assuming 'main' exists)
    CALL Git.GitNewBranch("main")
    SET exists_status = LAST
    EMIT "Status for creating existing branch: " + exists_status
    # Expect error like: "failed to create new branch 'main': fatal: A branch named 'main' already exists"

    # Example 3: Attempt to create a branch with an invalid name
    CALL Git.GitNewBranch("invalid name with spaces")
    SET invalid_name_status = LAST
    EMIT "Status for creating branch with invalid name: " + invalid_name_status
    # Expect error like: "branch name 'invalid name with spaces' contains invalid characters..."
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_git.go`
    * Function: `toolGitNewBranch`
    * Spec Name: `GitNewBranch` (in `pkg/core/tools_git_register.go`)
    * Key Go Packages: `fmt`, `strings`
    * Helpers: `core.toolExec`, `core.ErrValidationArgValue`
    * Registration: Registered by `registerGitTools` within `pkg/core/tools_register.go`. Returns `string, error`.