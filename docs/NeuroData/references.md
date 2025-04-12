# NeuroScript References Specification

:: type: Specification
:: version: 0.1.2
:: status: draft
:: dependsOn: docs/metadata.md, docs/neurodata_and_composite_file_spec.md, pkg/core/security.go, pkg/neurodata/blocks/blocks_extractor.go
:: howToUpdate: Ensure syntax definitions for both file and block refs are clear, examples are accurate, path restrictions are explicit.

## 1. Purpose

This document defines a standard, consistent syntax for referencing specific resources within a NeuroScript project. This includes referencing entire files or specific fenced code/data blocks within those files. This allows for reliable linking between documentation, metadata, scripts, and data, promoting portability and maintainability.

## 2. Syntax

The standard format for a reference is enclosed in square brackets and begins with `ref:`:

1.  **File Reference:** `[ref:<location>]`
2.  **Block Reference:** `[ref:<location>#<block_id>]`

Components:
* **`[` and `]`:** Square brackets enclose the entire reference.
* **`ref:`:** A mandatory literal prefix indicating that this is a NeuroScript reference, distinguishing it from other uses of square brackets (like properties or list literals).
* **`<location>`:** Specifies the file containing the target resource. This **MUST** be one of the following:
    * `this`: A special keyword referring to the *current* file where the reference itself resides.
    * A relative file path (e.g., `../data/users.ndtable`, `sibling_script.ns.txt`). Paths **MUST** use forward slashes (`/`) as separators. Relative paths are strongly recommended for portability. **Absolute paths are disallowed** in this syntax. Path validation using security routines (like `SecureFilePath` [cite: uploaded:neuroscript/pkg/core/security.go]) should still be performed by tools resolving these references based on their context (e.g., current working directory or sandbox root).
* **`#<block_id>` (Optional):** If present, this part indicates a reference to a specific block within the file.
    * **`#`:** A mandatory separator character when referencing a block. Its absence indicates a reference to the entire file.
    * **`<block_id>`:** The unique identifier of the target block within its file. This identifier **MUST** correspond to the value defined in the block's `:: id:` metadata tag [cite: uploaded:neuroscript/docs/metadata.md]. Block IDs should be unique within their containing file. Tools resolving block references will rely on the `:: id:` metadata found by block extraction mechanisms [cite: uploaded:neuroscript/pkg/neurodata/blocks/blocks_extractor.go].

## 3. Nesting

This specification currently only defines references to top-level blocks within a file. A syntax or convention for referencing blocks nested within other blocks is not defined at this time and may require extensions to block extraction tools. File references (`[ref:<location>]`) naturally do not involve nesting.

## 4. Tooling

Tools interacting with NeuroScript files and NeuroData formats need to be able to parse this reference syntax and resolve it appropriately based on the presence or absence of the `#<block_id>` component.
* References without `#<block_id>` (e.g., `[ref:config.yaml]`) typically imply reading or identifying the entire file content. Tools like `TOOL.ReadFile` might implicitly accept this format, or a dedicated `TOOL.ResolveReference` could be used.
* References with `#<block_id>` (e.g., `[ref:this#setup]`) require tools to read the file, extract blocks (using logic similar to `TOOL.BlocksExtractAll` [cite: uploaded:neuroscript/pkg/neurodata/blocks/blocks_tool.go]), and find the specific block by its `:: id:`.
* A potential future tool, `TOOL.GetBlockFromString(content_string, block_id)`, could extract a block by ID directly from string content held in a variable.
* Path validation using security routines [cite: uploaded:neuroscript/pkg/core/security.go] must always be applied when the `<location>` refers to a file path.

## 5. Usage Examples

References can be used in various contexts:

* **Metadata:**
    ```
    :: dependsOn: [ref:this#section-2], [ref:../schemas/user.ndtable], [ref:../schemas/common.ns.txt#validation-rules]
    :: template: [ref:templates/base.md]
    ```
* **NeuroScript Code:**
    ```neuroscript
    SET config_content = CALL TOOL.ReadFile("[ref:config.yaml]")
    SET template_code = CALL TOOL.GetBlockFromFile("[ref:this#template-code]") # Hypothetical tool
    CALL TOOL.ApplySchema("[ref:schemas/data_schema.json]", input_data)
    ```
* **Documentation (Markdown Links):**
    ```markdown
    See the [API Specification](ref:../api/spec_v1.md) or the specific [User Endpoint details](ref:../api/spec_v1.md#user-endpoint).
    (Alternative: Source file is [ref:main.go], main logic is [ref:this#main-logic])
    ```
* **NeuroData Files:**
    ```ndgraph
    NODE ScriptRunner [script: "[ref:../scripts/run.ns.txt#main-proc]"]
    NODE DataLoader [source_file: "[ref:data/input.csv]"]
    ```

## 6. Examples of Reference Strings

* **File References:**
    * `[ref:this]` - References the current file.
    * `[ref:config/production.yaml]` - References the file `production.yaml` in the `config` subdirectory.
    * `[ref:../LICENSE]` - References the `LICENSE` file in the parent directory.
* **Block References:**
    * `[ref:this#data-validation-rules]` - References block `:: id: data-validation-rules` in the current file.
    * `[ref:schemas/user.ndtable#unique-email-rule]` - References block `:: id: unique-email-rule` in the specified table file.
    * `[ref:../docs/api.md#get-user-example]` - References block `:: id: get-user-example` in a relative documentation file.
