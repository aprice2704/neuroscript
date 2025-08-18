# Guide for nslsp: Ingesting External Tool Metadata

This guide explains how the NeuroScript Language Server (`nslsp`) should discover, load, and use external tool definitions to provide features like autocomplete and signature validation for project-specific tools (e.g., from FDM).

---

## 1. The Problem: Opaque Tools

The `nslsp` is a standalone binary that only knows about the standard NeuroScript toolset it was compiled with. It has no visibility into the custom tools that other programs, like FDM, register in their own interpreter instances at runtime.

To solve this, we are introducing a declarative metadata system. Projects with custom tools will generate a `tools.json` file that describes the signatures of their tools. The `nslsp` will be configured to read these files.

---

## 2. The Solution: The `tools.json` Metadata File

A new command, `ns-meta export-tools`, generates a JSON file containing an array of `ToolSpec` objects. This file is the bridge between the FDM project and the language server.

### JSON Structure

The file will contain a JSON array where each object conforms to the `tool.ToolSpec` structure.

```json
[
  {
    "Name": "SaveMemory",
    "Group": "fdm.core",
    "Description": "Saves an FDM node to persistent storage.",
    "Category": "FDM",
    "Args": [
      {
        "Name": "node_handle",
        "Type": "string",
        "Required": true,
        "Description": "The handle of the node to save."
      }
    ],
    "ReturnType": "bool",
    "Variadic": false
  }
]
```

---

## 3. nslsp Implementation Plan

The language server needs to be updated to find, load, and use these external metadata files.

### A. Configuration

The `nslsp` should look for a configuration setting in the workspace that specifies the paths to one or more tool metadata files. A good approach is to check for a `.vscode/settings.json` file or a project-level `.nslsp.json`.

**Example (`.vscode/settings.json`):**
```json
{
  "nslsp.externalToolMetadata": [
    "./tools/fdm-tools.json",
    "../shared-libs/common-tools.json"
  ]
}
```

### B. Loading and Reloading

1.  **Initial Load**: On startup, the LSP should resolve the paths from its configuration relative to the workspace root, read each JSON file, and parse the tool specs.

2.  **Storing Definitions**: These external tool definitions should be loaded into a separate, in-memory registry within the LSP. This keeps them distinct from the standard, built-in tools.

3.  **Handling Updates (CRITICAL)**: The LSP **must** support reloading these definitions. When a `tools.json` file is saved, the LSP's file watcher should trigger a reload. The correct behavior is to **completely overwrite the existing external tool definitions** for that file with the new content. This is crucial because a developer might have removed a tool or changed its signature, and the LSP must not retain the old, stale definition.

    -   **Key:** Use the file path as the key for the set of tool definitions.
    -   **Action:** On file change, remove all tools associated with that path and load the new ones.

### C. Integrating with LSP Features

Once loaded, the external tool definitions should be seamlessly integrated into the LSP's core features:

-   **Autocomplete**: When a user types `tool.fdm.`, the LSP should suggest `SaveMemory` and `FindNeighbor`.
-   **Signature Help**: When a user types `tool.fdm.SaveMemory(`, the LSP should show a popup indicating the `node_handle: string` argument.
-   **Diagnostics/Validation**: If a user calls a non-existent FDM tool or provides the wrong number/type of arguments, the LSP should generate a diagnostic error, just as it does for standard tools.






