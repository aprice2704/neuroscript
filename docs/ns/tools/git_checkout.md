:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-git-gitcheckout-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_git_register.go, pkg/core/tools_git.go, docs/script_spec.md
:: relatedTo: Git.GitNewBranch, Git.GitStatus
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `Git.GitCheckout` (v0.1)

* **Tool Name:** `Git.GitCheckout` (v0.1)
* **Purpose:** Switches the current working branch or restores working tree files. Typically used to switch to an existing branch, but can also check out specific commits or tags. This is equivalent to running `git checkout <branch_name>` within the project's sandbox directory.
* **NeuroScript Syntax:** `CALL Git.GitCheckout(branch_name: <String>)`
* **Arguments:**
    * `branch_name` (String): Required. The name of the existing branch, tag, or commit hash to check out. It cannot be an empty string.
* **Return Value:** (String)
    * On success: A string indicating the checkout was successful, potentially including output from the `git checkout` command. Example: `"Successfully checked out branch/ref 'main'.\nOutput:\nSwitched to branch 'main'\nYour branch is up to date with 'origin/main'.\n"`. (Accessible via `LAST` after the `CALL`).
    * On error: A string describing the failure. This could be due to an empty `branch_name`, or errors from the underlying `git checkout` command execution (e.g., "failed to checkout branch/ref '...': error: pathspec '...' did not match any file(s) known to git", or errors about uncommitted changes). (Accessible via `LAST` after the `CALL`).
* **Behavior:**
    1.  Validates that exactly one argument (`branch_name`) of type String is provided.
    2.  Validates that the `branch_name` is not empty. If it is, returns an error message string.
    3.  Retrieves the interpreter's `sandboxDir`.
    4.  Constructs a command equivalent to `git checkout "<branch_name>"`.
    5.  Executes the `git checkout` command within the `sandboxDir` using the `toolExec` helper function.
    6.  If the command execution fails (e.g., `git` not found, branch/ref does not exist, uncommitted changes prevent checkout), returns an error message string containing details from `toolExec`.
    7.  If the command execution succeeds, returns a success message string, including the target branch/ref name and any standard output from the `git checkout` command.
* **Security Considerations:**
    * Executes the `git checkout` command within the configured sandbox directory.
    * Modifies the Git repository state by updating the HEAD reference.
    * Can potentially modify files in the working directory to match the state of the checked-out branch/commit. Uncommitted changes in the working directory may be lost or cause the command to fail, depending on the state and the specific `git checkout` behavior.
    * Relies on the `toolExec` helper for command execution.
    * Assumes the `git` executable is available in the environment's PATH where the NeuroScript interpreter is running.
* **Examples:**
    ```neuroscript
    # Example 1: Checkout an existing branch (e.g., 'main')
    CALL Git.GitCheckout("main")
    SET checkout_main_status = LAST
    EMIT "Checkout 'main' status: " + checkout_main_status

    # Example 2: Checkout the previously created feature branch
    # Assume 'feature/add-user-auth' was created successfully before
    CALL Git.GitCheckout("feature/add-user-auth")
    SET checkout_feature_status = LAST
    EMIT "Checkout 'feature/add-user-auth' status: " + checkout_feature_status

    # Example 3: Attempt to checkout a non-existent branch
    CALL Git.GitCheckout("no-such-branch-exists")
    SET non_existent_status = LAST
    EMIT "Status for checking out non-existent branch: " + non_existent_status
    # Expect error like: "failed to checkout branch/ref 'no-such-branch-exists': error: pathspec 'no-such-branch-exists' did not match any file(s) known to git."
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_git.go`
    * Function: `toolGitCheckout`
    * Spec Name: `GitCheckout` (in `pkg/core/tools_git_register.go`)
    * Key Go Packages: `fmt`
    * Helpers: `core.toolExec`, `core.ErrValidationArgValue`
    * Registration: Registered by `registerGitTools` within `pkg/core/tools_register.go`. Returns `string, error`.