 :: type: NSproject
 :: subtype: tool_specification
 :: version: 1.0
 :: id: tool-spec-checklistloadtree-v1.0
 :: status: approved
 :: dependsOn: docs/metadata.md, docs/ns/types/handle.md, docs/ns/types/generictree.md, docs/ns/dataformats/checklist.md
 :: howToUpdate: Review when ChecklistLoadTree implementation or underlying Checklist parsing/adapter logic changes.

 # Tool Specification: ChecklistLoadTree (v1.0)

 * **Tool Name:** `ChecklistLoadTree` (v1.0)
 * **Purpose:** Parses text content formatted as a NeuroData Checklist into an internal `GenericTree` structure, returning a handle to the tree. This enables querying, modification, and potentially re-serialization of the checklist using standard Tree tools.
 * **NeuroScript Syntax:**
   ```neuroscript
   set treeHandle = ChecklistLoadTree(content=checklistString)
   ```
 * **Arguments:**
   * `content` (String): Required. The string containing the checklist content. Must adhere to the NeuroData Checklist format (metadata `:: key: value`, items `- [status] Text` or `- |status| Text`, headings `# ...`, comments `-- ...` or `# ...`, indentation for hierarchy).
 * **Return Value:** (String)
   * Description: Upon success, returns a handle ID (string) referencing the created `core.GenericTree` object stored within the interpreter's handle manager. The handle can be used with other Tree tools (e.g., `TreeGetNode`, `TreeFindNodes`). The underlying handle type is `generic_tree`.
   * Error Handling: On failure (e.g., invalid input format, empty content, internal error during parsing or tree creation), the tool returns a standard Go `error`. This error is typically caught by the interpreter's `on_error` mechanism or terminates the script execution if unhandled. It does **not** return an error message as part of the string result.
 * **Behavior:**
   1.  Validates that the `content` argument is provided and is a string. Returns a `core.ErrValidationTypeMismatch` error if not.
   2.  Invokes the internal NeuroData Checklist parser (`checklist.ParseChecklist`) on the `content` string.
   3.  If parsing fails due to genuinely empty or invalid content (containing no recognizable checklist items or metadata), a `core.ErrInvalidArgument` error is returned (wrapping `checklist.ErrNoContent`), as a tree cannot be formed.
   4.  If parsing fails due to malformed item syntax (e.g., `- [xx] Bad`), a `core.ErrInvalidArgument` error is returned (wrapping `checklist.ErrMalformedItem`). Other scanning errors also result in `core.ErrInvalidArgument`.
   5.  If parsing succeeds, invokes the internal adapter (`checklist.ChecklistToTree`) to convert the parsed items and metadata into a `core.GenericTree` structure.
       * A root node of type `checklist_root` is created. Any metadata parsed (`:: key: value`) is stored as attributes on this root node.
       * Each checklist item becomes a node of type `checklist_item`.
       * The full text of the item (as parsed, following the `- [x] ` or `- |x| ` marker) is stored in the node's `Value` field (as a string).
       * The item's status is mapped and stored in the `status` attribute (values: `"open"`, `"done"`, `"skipped"`, `"partial"`, `"inprogress"`, `"blocked"`, `"question"`, `"special"`, `"unknown"`).
       * If the item was parsed as automatic (`- | | ...`), an attribute `is_automatic: "true"` is added to the node.
       * If the status is `"special"` and the symbol used was not one of the standard ones (`>`, `!`, `?`), the actual symbol is stored in the `special_symbol` attribute.
       * Parent-child relationships in the tree reflect the indentation levels in the source checklist.
   6.  If the tree adaptation process fails internally, a `core.ErrInternalTool` error is returned.
   7.  Registers the newly created `GenericTree` with the interpreter's handle manager. If handle registration fails, a `core.ErrInternalTool` error is returned.
   8.  Upon successful creation and registration, returns the string handle ID for the tree.
 * **Security Considerations:**
   * This tool does not directly perform I/O or execute external commands.
   * Input is parsed according to defined checklist rules; malformed input generally results in errors rather than unsafe execution.
   * Processing extremely large checklist strings could consume significant memory during parsing and tree construction. Ensure input sizes are reasonable within the execution environment.
   * Handles returned by this tool consume memory in the interpreter's handle manager. While handles are typically cleaned up when an interpreter context ends, long-running processes should consider handle lifecycle management (potential future `ReleaseHandle` tool).
 * **Examples:**
   ```neuroscript
   # Example: Load a checklist and inspect it using Tree tools

   set myListContent = ```
   :: title: Project Tasks
   :: version: 0.1

   # Phase 1
   - [x] Define requirements
   - [?] Design architecture (needs review)
     - [ ] Sub-task 1.1

   # Phase 2
   - | | Implement core features (Auto)
     - |-| Sub-feature 2.1 (Auto partial)
   ```

   set handle = ChecklistLoadTree(content=myListContent)
   if handle == nil
       fail message="Failed to load checklist"
   endif

   # Get root metadata - Assuming TreeGetNode exists and root ID is predictable/discoverable
   # NOTE: Requires TreeGetNode tool to be available and root ID logic stable.
   set root = TreeGetNode(handle=handle, nodeId="node-1") # Example: Assume root is node-1
   if root != nil
       emit "Title: " + root.Attributes["title"]
   else
       emit "Could not retrieve root node."
   endif

   # Find all automatic items - Assuming TreeFindNodes exists
   # NOTE: Requires TreeFindNodes tool to be available.
   set autoItems = TreeFindNodes(handle=handle, attributes={"is_automatic":"true"})
   emit "Found automatic items:"
   set i = 0
   while i < len(autoItems)
       set node = autoItems[i]
       emit "  - " + node.Value + " (Status: " + node.Attributes["status"] + ")"
       set i = i + 1
   endwhile

   # Example of how an error from ChecklistLoadTree would be handled
   on_error means
     emit "An error occurred: " + system.error_message # Placeholder for actual error access
     fail "Script failed due to checklist processing error."
   endon

   set badContent = "- [xx] Malformed item"
   # The following call would trigger the on_error block above
   # set badHandle = ChecklistLoadTree(content=badContent)
   ```
 * **Go Implementation Notes:**
   * Location: `pkg/neurodata/checklist/checklist_tool.go`
   * Key Go Functions: `toolChecklistLoadTree`, `checklist.ParseChecklist`, `checklist.ChecklistToTree`, `interpreter.RegisterHandle`, `core.NewGenericTree`
   * Registration: `RegisterChecklistTools` in `pkg/neurodata/checklist/checklist_tool.go` adds `toolChecklistLoadTreeImpl` to the registry.
   * Depends on types and functions within `pkg/core` and `pkg/neurodata/checklist`.