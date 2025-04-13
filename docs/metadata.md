:: version: 0.4.2
:: type: NSproject
:: subtype: spec
:: dependsOn: docs/neuroscript overview.md, docs/neurodata_and_composite_file_spec.md, pkg/neurodata/metadata/metadata.go, pkg/core/NeuroScript.g4
:: howToUpdate: Review the referenced documents and ensure this file accurately reflects the current metadata standards (file/block level), parser behavior regarding metadata, and extraction tool logic.

# MetaData in NeuroScript Objects

## Metadata Standard (`:: key: value`)

All project files (NeuroScript `.ns.txt`, NeuroData `.nd*`, Go `.go`, Markdown `.md`, etc.) and embedded code/data blocks should use the following metadata format where applicable for file-level or block-level information.

* **Prefix:** Metadata lines must start with `::` (colon, colon) followed by at least one space. Optional leading whitespace before `::` is allowed [cite: uploaded:neuroscript_small/docs/neurodata_and_composite_file_spec.md].
* **Structure:** `:: key: value` [cite: uploaded:neuroscript_small/docs/neurodata_and_composite_file_spec.md].
* **Key:** Immediately follows `:: ` and precedes the first `:`. Valid characters are letters, numbers, underscore, period, hyphen (`[a-zA-Z0-9_.-]+`). Whitespace around the key (after `:: `) is tolerated [cite: uploaded:neuroscript_small/docs/neurodata_and_composite_file_spec.md].
* **Separator:** A single colon `:` separates the key and value. Whitespace around the colon is tolerated [cite: uploaded:neuroscript_small/docs/neurodata_and_composite_file_spec.md].
* **Value:** Everything after the first colon, with leading/trailing whitespace trimmed [cite: uploaded:neuroscript_small/docs/neurodata_and_composite_file_spec.md].
* **Location (File Level):** For metadata applying to the entire file (like `.ns.txt`, `.md`), these lines MUST appear consecutively at the **very beginning** of the file, before any functional content (e.g., before `DEFINE PROCEDURE`, before Markdown text, before NeuroData items like `- [ ]`). [cite: uploaded:neuroscript_small/docs/neurodata_and_composite_file_spec.md].
* **Location (Block Level):** For metadata applying only to a fenced code block within a composite document (like `.md`), place `:: key: value` lines immediately *inside* the block, after the opening fence (e.g., ```go\n:: id: my-block\n...```), before the block's main content [cite: uploaded:neuroscript_small/docs/neurodata_and_composite_file_spec.md].
* **Comments/Blank Lines:** Standard comment lines (`#` or `--`) and blank lines *are* permitted between metadata lines (both file-level and block-level), but they must still appear *before* the main content begins [cite: uploaded:neuroscript_small/docs/neurodata_and_composite_file_spec.md].
* **Parser Skipping (`.ns.txt`):** The NeuroScript ANTLR grammar's lexer is configured to *skip* file-level `::` metadata lines (treating them like comments) [cite: uploaded:neuroscript_small/pkg/core/generated/neuroscript_lexer.go]. This prevents them from interfering with script parsing and execution by `neurogo`.
* **Metadata Extraction:** Tools like `TOOL.ExtractMetadata` operate on string content and are designed to parse these `:: key: value` lines from the beginning of the provided text [cite: uploaded:neuroscript_small/pkg/neurodata/metadata/metadata.go, uploaded:neuroscript_small/pkg/core/tools_metadata.go].

## Standard Metadata Tags

### Required Tags

* **`:: version: <semver_string>`**
    * Applies to: All files and blocks with meaningful content.
    * Value: A valid semantic version string (e.g., `0.1.0`, `1.2.3-alpha`).
    * Purpose: Tracks the version of the *content* within the specific file or block. Tooling should aim to increment the patch number on change.

* **`:: type: <TypeName>`**
    * Applies to: All NeuroScript project files and blocks.
    * Value: A string identifying the primary type. Recommended base types include:
        * `NSproject`: For general project files (Go code, Markdown docs, config files, etc. that aren't more specific types below).
        * `NeuroScript`: For `.ns.txt` files or blocks containing NeuroScript code.
        * `NeuroData`: For files or blocks containing specific NeuroData formats (Checklist, Table, Graph, Form, Tree, etc.). Use the specific type name here (e.g., `Checklist`, `Table`) instead of the generic `NeuroData`.
        * `Template`: For template files/blocks (e.g., Handlebars).
    * Purpose: Allows tools to identify the format and apply appropriate processing. NeuroData files/blocks should use their specific format name (e.g., `:: type: Graph`, `:: type: Form`).

### Recommended Tags

* **`:: subtype: <SubTypeName>`**
    * Applies to: Files or blocks where further classification beyond the primary `:: type:` is useful.
    * Value: A string indicating the subtype. Examples:
        * `spec`: For specification documents (like this one, or NeuroData format specs). Used with `:: type: NSproject` or potentially `:: type: NeuroData` if the spec itself *is* the data format.
        * `test_fixture`: For files specifically used as test inputs.
        * `library`: For NeuroScript files intended as reusable libraries.
        * `example`: For example code or data files.
        * `config`: For configuration files.
    * Purpose: Provides more granular classification for tools or organization.

* **`:: id: <unique_identifier>`**
    * Applies to: Embedded blocks within composite documents (e.g., Markdown). Can also apply to files.
    * Value: A unique identifier for the block or file, preferably using URL-safe characters (letters, numbers, underscore, hyphen). IDs should be unique within their containing file.
    * Purpose: Essential for referencing specific blocks using the `[ref:<location>#<block_id>]` syntax [cite: uploaded:neuroscript_small/docs/NeuroData/references.md].

### Optional Tags (Examples)

* **`:: grammar: <grammar_name>[@<semver_string>]`**
    * Applies to: Files or blocks whose content adheres to a defined external or internal grammar specification (e.g., a NeuroScript file referencing the NS grammar, a NeuroData file referencing its own format spec, or another format like JSON).
    * Value: The name of the grammar, optionally followed by `@` and a specific semantic version string (e.g., `graph@0.1.0`, `neuroscript@1.1.0`, `hbars@1.0.0`, `json`).
    * Purpose: Indicates the syntax rules the content follows. Allows tools to select appropriate parsers or validators.
    * Versioning: If the `@<semver>` is omitted, tools processing or rewriting the content *should* assume or default to the latest known stable version of the specified grammar.

* **`:: status: <status_string>`**
    * Applies to: Files or blocks, especially specifications or checklists.
    * Value: A string indicating the current status (e.g., `draft`, `review`, `approved`, `deprecated`, `pending`, `active`).
    * Purpose: Provides context about the lifecycle state of the content.

* **`:: dependsOn: <path_or_identifier_list>`**
    * Applies to: Files or blocks that have dependencies on other specific resources.
    * Value: A comma-separated list of file paths or reference identifiers (`[ref:...]`).
    * Purpose: Explicitly documents dependencies, useful for understanding impact of changes and for tooling.

* **`:: howToUpdate: <description_or_script_call>`**
    * Applies to: Files or blocks with `dependsOn`.
    * Value: A natural language description or potentially a `CALL` statement indicating how to update this content if its dependencies change.
    * Purpose: Provides guidance for maintenance. Required if `dependsOn` is present.

* **`:: template: <path_or_identifier>`**
    * Applies to: NeuroData instances (e.g., `.ndobj`) derived from a template (`.ndform`).
    * Value: A reference (`[ref:...]`) to the source template file or block.
    * Purpose: Links instance data to its defining schema.

* **`:: templateFor: <format_id>`**
    * Applies to: NeuroScript templates (e.g., Handlebars blocks).
    * Value: An identifier for the target output format (e.g., `markdown`, `json`, `html`, `neuroscript`).
    * Purpose: Guides validation and potential context-aware escaping during template rendering [cite: uploaded:neuroscript_small/docs/NeuroData/templates.md].

* *(Other tags like `author`, `description`, `rendering_hint`, `canonical_format` can be added as needed).*

## File-Specific Conventions

### NeuroScript Files (`.ns`, `.ns.txt`, `.neuro`)

* **File Metadata:** Use the `:: key: value` standard at the top (e.g., `:: version: ...`, `:: type: NeuroScript`, `:: subtype: library`, `:: grammar: neuroscript@1.1.0`). The parser skips these lines.
* **`FILE_VERSION` Declaration:** ***Deprecated.*** Use `:: version:` instead for file content versioning.
* **Procedure `COMMENT:` Block:** Required per procedure. Contains structured metadata like `PURPOSE:`, `INPUTS:`, `OUTPUT:`, `ALGORITHM:`, `LANG_VERSION:`, `CAVEATS:`, `EXAMPLES:` (as specified in `docs/script spec.md` [cite: uploaded:neuroscript_small/docs/script spec.md]).
    * **`LANG_VERSION:`** (Optional but Recommended): Semantic version indicating the targeted NeuroScript language specification version.

### NeuroData Files (`.ndcl`, `.ndtb`, etc.)

* **File Metadata:** Use the `:: key: value` standard at the top (e.g., `:: version: ...`, `:: type: Checklist`, `:: grammar: ndcl@0.5.0`). Note that `:: type:` uses the specific NeuroData format name.

### Specification Files (`.md`, etc.)

* **File Metadata:** Use the `:: key: value` standard at the top (e.g., `:: version: ...`, `:: type: NSproject`, `:: subtype: spec`).

### Markdown Files (`.md`) - General / Composite Documents

* **File Metadata:** Use the `:: key: value` standard at the **very top** (e.g., `:: version: ...`, `:: type: NSproject`).
* **Rendering:** Be aware that standard Markdown viewers *may* render these `::` lines as plain text unless filtered.

### Embedded Blocks in Composite Documents (e.g., Markdown `.md`)

* **Fencing:** Use ``` followed by a language tag (e.g., `neuroscript`, `neurodata-checklist`).
* **Block Metadata:** Place `:: key: value` lines immediately *inside* the block, after the opening fence and before the block's main content [cite: uploaded:neuroscript_small/docs/neurodata_and_composite_file_spec.md]. Include relevant tags like `:: id:`, `:: version:`, `:: type:` (if applicable, e.g., `:: type: Checklist`), `:: grammar:`.
* **Example:**
    ```markdown
    Some introductory text.

    ```neuroscript
    :: id: proc-example-1
    :: version: 1.0.0
    :: type: NeuroScript
    :: grammar: neuroscript@1.1.0
    DEFINE PROCEDURE Example()
    COMMENT: ... ENDCOMMENT
    EMIT "Hello"
    END
    ```

    More text.

    ```neurodata-checklist
    :: id: checklist-example-1
    :: version: 0.1.0
    :: type: Checklist
    :: grammar: ndcl@0.5.0
    - [ ] Item 1
    ```
    ```