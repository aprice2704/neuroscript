:: type: NSproject  
:: subtype: spec  
:: version: 0.1.0  
:: id: spec-tool-git-status-v0.1  
:: status: draft  
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_types.go, pkg/core/tools_git_status.go  
:: howToUpdate: Review parsing logic and return map structure against implementation in tools_git_status.go. Update version if behavior changes.  

# Tool Specification: `TOOL.GitStatus` (v0.1)

* **Tool Name:** `TOOL.GitStatus` (v0.1)
* **Purpose:** Gets the current Git repository status by executing `git status --porcelain -b --untracked-files=all`, parsing the output, and returning a structured map summarizing the status.
* **NeuroScript Syntax:** `CALL TOOL.GitStatus()`
* **Arguments:** None.
* **Return Value:** (`ArgTypeAny` - Represents a Map)
    * A NeuroScript Map containing the following keys:
        * `branch` (String | null): The current local branch name (e.g., "main", "(detached HEAD)", "(unknown unborn)"). Null if parsing fails severely.
        * `remote_branch` (String | null): The remote tracking branch name (e.g., "origin/main"), if tracked. Null otherwise.
        * `ahead` (Integer): Number of commits the local branch is ahead of the remote branch. Defaults to 0 if not applicable.
        * `behind` (Integer): Number of commits the local branch is behind the remote branch. Defaults to 0 if not applicable.
        * `files` (List[Map]): A list of files with status changes. Each map has:
            * `path` (String): The file path relative to the repository root. Handles quoted paths correctly.
            * `index_status` (String): Single character code for the index/staging area status (e.g., 'M', 'A', 'D', 'R', 'C', 'U', '?'). Space ' ' indicates unmodified in index.
            * `worktree_status` (String): Single character code for the working tree status (e.g., 'M', 'D', '?'). Space ' ' indicates unmodified in worktree.
            * `original_path` (String | null): The original path for renamed ('R') or copied ('C') files. Null otherwise.
        * `untracked_files_present` (Boolean): True if any untracked files (status `??`) were detected.
        * `is_clean` (Boolean): True if there are no changes to *tracked* files in the index or working tree (i.e., `index_status` and `worktree_status` are both ' ' for all tracked files). Note: Presence of untracked files does not make `is_clean` false.
        * `error` (String | null): An error message string if the `git` command failed (e.g., not a git repository, command not found) or critical parsing failed, otherwise `null`.
* **Behavior:**
    1. Validates that no arguments are provided. Returns an error if arguments are present.
    2. Determines the execution directory (respecting sandbox settings if active).
    3. Executes the command `git status --porcelain -b --untracked-files=all` in the determined directory.
    4. Checks for execution errors. If the command fails (e.g., not a git repository), populates the `error` field in the return map and returns the map immediately.
    5. If the command succeeds, parses the standard output line by line:
        * Parses the first line (`## ...`) to extract branch name, remote tracking branch, and ahead/behind counts using regular expressions. Handles detached HEAD and unborn branch states.
        * Parses subsequent lines representing file statuses (`XY path`, `?? path`, `R new -> old`).
        * Extracts the two-character status code (`XY`), the file path (handling quotes and the `->` separator for renames/copies).
        * Populates the `files` list with maps for each file entry.
        * Sets the `untracked_files_present` flag if any `??` status is found.
        * Calculates the `is_clean` flag based on whether any tracked files have non-space status codes.
    6. Returns the populated map. If parsing the branch line fails critically, the `error` field may be set.
* **Security Considerations:**
    * Executes an external `git` command. The execution environment and command inputs are managed internally by the `toolExec` helper, which should respect sandboxing defined by the interpreter's `sandboxDir`.
    * Does not directly modify files, only reads status information.
* **Examples:**
    ```
```text?code_stderr&code_event_index=3
Traceback (most recent call last):
  File "<string>", line 1, in <module>
ModuleNotFoundError: No module named 'google.cloud'

```neuroscript
    # Get the current status
    VAR status_map = CALL TOOL.GitStatus()

    # Check for errors first
    IF status_map["error"] != null THEN
        EMIT "Error getting git status: " + status_map["error"]
    ELSE
        # Print branch and clean status
        EMIT "Current Branch: " + status_map["branch"]
        IF status_map["is_clean"] == true THEN
            EMIT "Repository is clean."
        ELSE
            EMIT "Repository has changes."
        END
        IF status_map["untracked_files_present"] == true THEN
            EMIT "Untracked files are present."
        END

        # List modified files
        VAR changed_files = status_map["files"]
        IF CALL List.Length(changed_files) > 0 THEN
            EMIT "Changed Files:"
            FOR EACH file_info IN changed_files DO
                # Example: Print files modified in the worktree
                IF file_info["worktree_status"] != " " THEN
                   EMIT "- [" + file_info["index_status"] + file_info["worktree_status"] + "] " + file_info["path"]
                   IF file_info["original_path"] != null THEN
                       EMIT "    (from " + file_info["original_path"] + ")"
                   END
                END
            ENDBLOCK
        END
    END
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_git_status.go`
    * Core Go Packages: `os/exec`, `strings`, `regexp`, `strconv`, `bytes`.
    * Helper: Relies on `toolExec` for command execution.
    * Registration: Registered in `pkg/core/tools_git.go` (or a dedicated registration file) via `registerGitTools`. Return type specified as `ArgTypeAny`. Argument validation uses `if len(args) != 0`.
"""

# Create the new Markdown file using file_manager
print(file_manager.upsert_files([
    file_manager.File(path='docs/ns/tools/git_status.md', content=git_status_md_content, state=file_manager.FileState.UPDATED) # Use UPDATED to create or overwrite
]))

# Plan next steps
# 1. Add tests: pkg/core/tools_git_test.go
# 2. Update checklist: docs/development checklist.md
print("Plan: Next step is to add tests for TOOL.GitStatus in pkg/core/tools_git_test.go.")