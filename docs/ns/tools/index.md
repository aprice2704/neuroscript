 :: type: NSproject
 :: subtype: documentation
 :: version: 0.1.2  // Version incremented
 :: id: tool-spec-index-v0.1.2 // Version incremented
 :: status: draft
 :: dependsOn: [docs/ns/tools/tool_spec_structure.md](./tool_spec_structure.md), [docs/ns/tools/query_table.md](./query_table.md), [docs/ns/tools/move_file.md](./move_file.md), [docs/ns/tools/go_update_imports_for_moved_package.md](./go_update_imports_for_moved_package.md), [docs/ns/tools/git_pull.md](./git_pull.md), [docs/ns/tools/git_push.md](./git_push.md), [docs/ns/tools/git_diff.md](./git_diff.md) // Added git_push.md, git_diff.md
 :: howToUpdate: Update the list below when tool specification documents are added, removed, or renamed in this directory.

 # NeuroScript Tool Specifications Index

 This directory contains detailed specifications for the built-in `TOOL.*` functions available within the NeuroScript language. Each specification follows a standard format to ensure clarity and consistency.

 ## Specification Format

 * **[Tool Specification Structure Template](./tool_spec_structure.md):** Defines the standard structure used for all tool specification documents in this directory.

 ## Available Tool Specifications

 * **[TOOL.GitDiff](./git_diff.md):** Describes the tool for showing unstaged changes in the working directory. (NEW)
 * **[TOOL.GitPull](./git_pull.md):** Describes the tool for fetching and integrating changes from a remote repository.
 * **[TOOL.GitPush](./git_push.md):** Describes the tool for pushing local commits to a remote repository. (NEW)
 * **[TOOL.GoUpdateImportsForMovedPackage](./go_update_imports_for_moved_package.md):** Describes the tool for automatically updating Go import paths after refactoring.
 * **[TOOL.MoveFile](./move_file.md):** Describes the tool for securely moving/renaming files within the sandbox.
 * **[TOOL.QueryTable](./query_table.md):** Describes the tool for querying NeuroData Table (`.ndtable`) files using selection and filtering criteria.

 *(This list should be updated as more tool specifications are created.)*