:: type: NSproject  
:: subtype: documentation  
:: version: 0.1.0  
:: id: tool-spec-index-v0.1  
:: status: draft  
:: dependsOn: docs/ns/tools/tool_spec_structure.md, docs/ns/tools/query_table.md, docs/ns/tools/move_file.md, docs/ns/tools/go_update_imports_for_moved_package.md  
:: howToUpdate: Update the list below when tool specification documents are added, removed, or renamed in this directory.  

# NeuroScript Tool Specifications Index

This directory contains detailed specifications for the built-in `TOOL.*` functions available within the NeuroScript language. Each specification follows a standard format to ensure clarity and consistency.

## Specification Format

* **[Tool Specification Structure Template](./tool_spec_structure.md):** Defines the standard structure used for all tool specification documents in this directory.

## Available Tool Specifications

* **[TOOL.GoUpdateImportsForMovedPackage](./go_update_imports_for_moved_package.md):** Describes the tool for automatically updating Go import paths after refactoring.
* **[TOOL.MoveFile](./move_file.md):** Describes the tool for securely moving/renaming files within the sandbox.
* **[TOOL.QueryTable](./query_table.md):** Describes the tool for querying NeuroData Table (`.ndtable`) files using selection and filtering criteria.

*(This list should be updated as more tool specifications are created.)*
