  # MetaData in NeuroScript Objects (Revised 2025-04-30)
 
 ## Metadata Standard (`:: key: value`)
 
 All project files (NeuroScript `.ns`, Go `.go`, Markdown `.md`, NeuroData `.nd*`, etc.) and embedded code/data blocks should use the following metadata format where applicable for file-level, procedure-level, step-level, or block-level information.
 
 * **Prefix:** Metadata lines must start with `::` (colon, colon) followed by at least one space. Optional leading whitespace before `::` is allowed.
 * **Structure:** `:: key: value`.
 * **Key:** Immediately follows `:: ` and precedes the first `:`. Valid characters are letters, numbers, underscore, period, hyphen (`[a-zA-Z0-9_.-]+`). Whitespace around the key (after `:: `) is tolerated.
 * **Separator:** A single colon `:` separates the key and value. Whitespace around the colon is tolerated.
 * **Value:** The rest of the line after the first colon (`:`), stripped of leading/trailing whitespace. It can contain any characters.
 * **Storage:** Metadata is typically stored as a `map[string]string` associated with the relevant program element (file, procedure, step, block).
 
 ## Standard Metadata Vocabulary
 
 While any valid key can be used, the following keys are recommended for standardization and potential use by tooling or the interpreter.
 
 ### 1. File-Level Scope 
 
 * `:: lang_version:` *String*. **(Renamed)** The minimum language/interpreter version required (e.g., `:: lang_version: neuroscript@0.2.0`). Helps ensure compatibility.
 * `:: file_version:` *String*. **(New)** The semantic version of this specific file (e.g., `:: file_version: 1.0.0`). Distinct from `lang_version`.
 * `:: description:` *String*. A brief description of the file's overall purpose.
 * `:: author:` *String*. The name or handle of the file's author(s).
 * `:: license:` *String*. An SPDX license identifier (e.g., `:: license: MIT`, `:: license: Apache-2.0`) or "Proprietary".
 * `:: created:` *String (ISO 8601 Date)*. The date the file was initially created (e.g., `:: created: 2025-04-30`).
 * `:: modified:` *String (ISO 8601 Date)*. The date the file was last significantly modified (e.g., `:: modified: 2025-04-30`).
 * `:: tags:` *String (Comma-separated)*. Keywords describing the file's domain or function (e.g., `:: tags: fileio, text-processing, refactoring`).
 * `:: source:` *String (URI/Path)*. If the file is derived or copied, the original source location.
 * `:: type:` *String*. For NeuroData or composite files, indicates the primary data type (e.g., `:: type: Checklist`, `:: type: NSproject`).
 * `:: subtype:` *String*. Further classification (e.g., `:: subtype: spec`, `:: subtype: example`).
 * `:: grammar:` *String*. For embedded blocks, specifies the grammar required (e.g., `:: grammar: neuroscript@1.1.0`). *(Note: Less common at file level, more for blocks).*
 * `:: dependsOn:` *String (Comma-separated URI/Path)*. Lists files this file depends on conceptually.
 * `:: howToUpdate:` *String*. Instructions on how to keep this specification file current.
 
 ### 2. Procedure-Level Scope (Inside `func`/`endfunc`)
 
 * `:: description:` *String*. A brief summary of what the procedure does.
 * `:: purpose:` *String (Multiline)*. A more detailed explanation of the procedure's goal and rationale.
 * `:: param:<param_name>:` *String*. Describes a specific required or optional parameter (e.g., `:: param:input_path: Path to the file that needs processing.`). Use one line per parameter. *Preferred way to document parameters.*
 * `:: return:<index_or_name>:` *String*. Describes a specific return value by its 0-based index or, if named in the `returns` clause, by its name (e.g., `:: return:0: The number of lines processed.`). Use one line per return value. *Preferred way to document return values.*
 * `:: algorithm:` *String (Multiline)*. Describes the high-level steps, logic, or approach used within the procedure.
 * `:: example_usage:` *String*. A short example showing how to call the procedure (e.g., `:: example_usage: processFile(needs="input.txt", optional=true)`).
 * `:: caveats:` *String (Multiline)*. Lists potential issues, limitations, assumptions, or non-obvious behaviors.
 * `:: requires_tool:` *String (Comma-separated)*. Lists specific tools (`tool.<name>`) used by this procedure (e.g., `:: requires_tool: tool.ReadFile, tool.WriteFile`). Aids dependency checking.
 * `:: requires_llm:` *Boolean (`true`/`false`)*. Indicates if the procedure uses the `ask` statement (e.g., `:: requires_llm: true`).
 * `:: timeout:` *String (Duration)*. Suggests a maximum execution time for this procedure (e.g., `:: timeout: 30s`, `:: timeout: 1m`). *(Interpreter support needed)*.
 * `:: pure:` *Boolean (`true`/`false`)*. Indicates if the function is pure (output depends only on input, no side effects). Default is `false`. *(Interpreter/Tooling support needed)*.
 * *(Note: Informal keys like `:: inputs:` or `:: output:` might be seen, reflecting older styles, but prefer `:: param:<name>:` and `:: return:<index_or_name>:` for structured documentation).*
 
 ### 3. Step-Level Scope (Immediately preceding the step)
 
 *Use sparingly for clarity on specific lines.*
 
 * `:: reason:` *String*. Explains *why* a particular step is necessary or done in a specific way.
 * `:: todo:` *String*. A reminder for future improvement related to this step (e.g., `:: todo: Extract this logic into a helper procedure.`).
 * `:: fixme:` *String*. Indicates a known issue or bug related to this step that needs fixing.
 * `:: security_note:` *String*. Highlights a security consideration for this step (e.g., `:: security_note: Ensure input is sanitized before passing to shell command.`).
 * `:: performance_note:` *String*. Comments on the performance implications of this step.
 
 ### 4. Embedded Block Scope (Inside fenced code blocks ```)
 
 * `:: id:` *String*. A unique identifier for the block within its container file (e.g., `:: id: proc-example-1`).
 * `:: version:` *String*. Version of the content within the block. *(Consider using `:: file_version:` for consistency if the block represents a whole logical file).*
 * `:: type:` *String*. Specifies the type of content if not clear from the fence (e.g., `:: type: Checklist`, `:: type: NeuroScript`).
 * `:: grammar:` *String*. Specifies the exact grammar and version required to parse the block content (e.g., `:: grammar: neuroscript@1.1.0`, `:: grammar: neurodata-checklist@0.9.0`). *(May overlap with `:: lang_version:`)*.
 
 ## Metadata Placement Guidelines
 
 The general principle is to place metadata as close as possible to the element it describes.
 
 * **NeuroScript Files (`.ns`):**
     * **File-Level:** Place all file-level `::` metadata lines **at the start of the file (SOF)**. The parser is designed to find this metadata regardless of its position.
     * **Procedure-Level:** Place procedure-level `::` metadata *inside* the `func`/`endfunc` block, immediately after the `func ... means` line and before the first executable step.
     * **Step-Level:** Place step-level `::` metadata on the line(s) *immediately preceding* the step it refers to.
     * *(AST Mapping: File-level -> `Program.Metadata`, Procedure-level -> `Procedure.Metadata`, Step-level -> `Step.Metadata`)*
 * **Go Files (`.go`):**
     * File-level metadata (like build tags `//go:build ...` or license headers) conventionally appears at the **start of the file (SOF)**. Use standard Go comments for this unless a specific `::` key is needed for NeuroScript tooling.
 * **NeuroData Files (`.nd*`):**
     * File-level metadata should generally appear at the **start of the file (SOF)**.
 * **Markdown & Specification Files (`.md`, etc.):**
     * File-level metadata (`:: key: value`) should appear at the **end of the file (EOF)**. Each metadata line should end in double space to be rendered on separate lines.
 * **Embedded Blocks (in Markdown, etc.):**
     * Block-level metadata (`:: key: value`) should be placed immediately *inside* the block, after the opening fence (e.g., ```neuroscript) and before the block's main content. Include relevant tags like `:: id:`, `:: version:` (or `:: file_version:`), `:: type:`, `:: grammar:` (or `:: lang_version:`).
 
 ## Example (NeuroScript `.ns` File)
 
 ```neuroscript
 :: lang_version: neuroscript@0.2.0
 :: file_version: 1.1.0 
 :: author: Alice Price
 :: created: 2025-04-30
 :: license: MIT
 :: description: Example script demonstrating metadata placement.
 :: tags: example, metadata

 func ProcessData(needs inputData, optional threshold returns processedCount, errorMsg) means
   :: purpose: Processes input data according to a threshold.
   :: param:inputData: The raw data list to process.
   :: param:threshold: Optional numeric threshold for filtering.
   :: return:processedCount: Number of items successfully processed.
   :: return:errorMsg: Any error message encountered, or "" on success.
   :: algorithm: 
   ::   1. Initialize counters.
   ::   2. Iterate through inputData.
   ::   3. Apply threshold filter if provided.
   ::   4. Increment counter.
   ::   5. Return count and empty error string.
   :: caveats: Does not handle non-numeric data gracefully yet.
   :: requires_llm: false
 
   set count = 0
   set err = "" 
   
   # Iterate and process
   for each item in inputData
     :: reason: This is the main processing loop.
     # ... processing logic using item and threshold ...
     set count = count + 1 
   endfor
   
   return count, err
 endfunc
 
 ```

 :: version: 0.5.1  
 :: type: NSproject  
 :: subtype: spec  
 :: created: 2025-04-30   
 :: modified: 2025-04-30   
 :: dependsOn: docs/neuroscript_overview.md, docs/neurodata_and_composite_file_spec.md, pkg/neurodata/metadata/metadata.go, pkg/core/NeuroScript.g4, pkg/core/ast.go  
 :: howToUpdate: Review the referenced documents/code and ensure this file accurately reflects the current metadata standards (format, standard keys, placement), parser behavior, and AST storage.  
