:: version: 0.4.0
:: dependsOn: docs/neuroscript overview.md, docs/neurodata_and_composite_file_spec.md, pkg/neurodata/metadata/metadata.go, pkg/core/NeuroScript.g4
:: howToUpdate: Review the referenced documents and ensure this file accurately reflects the current metadata standards (file/block level), parser behavior regarding metadata, and extraction tool logic.

# MetaData in NeuroScript Objects

## Metadata Standard (`:: key: value`)

All project files (NeuroScript `.ns.txt`, NeuroData `.nd*`, Go `.go`, Markdown `.md`, etc.) and embedded code/data blocks should use the following metadata format where applicable for file-level or block-level information.

* **Prefix:** Metadata lines must start with `::` (colon, colon) followed by at least one space. Optional leading whitespace before `::` is allowed [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md].
* **Structure:** `:: key: value` [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md].
* **Key:** Immediately follows `:: ` and precedes the first `:`. Valid characters are letters, numbers, underscore, period, hyphen (`[a-zA-Z0-9_.-]+`). Whitespace around the key (after `:: `) is tolerated [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md].
* **Separator:** A single colon `:` separates the key and value. Whitespace around the colon is tolerated [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md].
* **Value:** Everything after the first colon, with leading/trailing whitespace trimmed [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md].
* **Location (File Level):** For metadata applying to the entire file (like `.ns.txt`, `.md`), these lines MUST appear consecutively at the **very beginning** of the file, before any functional content (e.g., before `DEFINE PROCEDURE`, before Markdown text, before NeuroData items like `- [ ]`). [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md].
* **Location (Block Level):** For metadata applying only to a fenced code block within a composite document (like `.md`), place `:: key: value` lines immediately *inside* the block, after the opening fence (e.g., ```go\n:: id: my-block\n...```), before the block's main content [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md].
* **Comments/Blank Lines:** Standard comment lines (`#` or `--`) and blank lines *are* permitted between metadata lines (both file-level and block-level), but they must still appear *before* the main content begins [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md].
* **Parser Skipping (`.ns.txt`):** The NeuroScript ANTLR grammar's lexer is configured to *skip* file-level `::` metadata lines (treating them like comments) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_lexer.go]. This prevents them from interfering with script parsing and execution by `neurogo`.
* **Metadata Extraction:** Tools like `TOOL.ExtractMetadata` operate on string content and are designed to parse these `:: key: value` lines from the beginning of the provided text [cite: uploaded:neuroscript/pkg/neurodata/metadata/metadata.go, uploaded:neuroscript/pkg/core/tools_metadata.go].
* **Required Tags (File/Block Level):**
    * `:: version: <semver_string>` (e.g., `:: version: 0.1.0`) - Semantic version of the *content* of this specific file or block. Tooling should aim to increment the patch number on change.
* **Optional Tags (Examples):**
    * `:: id: <unique_identifier>` - Unique ID for embedded blocks. Essential for tools like `TOOL.BlocksExtractAll`. Use URL-safe characters.
    * `:: dependsOn: <path_or_identifier_list>` - Files/resources this content directly depends on.
    * `:: howToUpdate: <description_or_script_call>` - How to update if dependencies change (required if `dependsOn` is present).
    * `:: template: <path_or_identifier>` - Source template for NeuroData instances.
    * `:: template_version: <semver_string>` - Version of the source template.
    * `:: rendering_hint: <format_identifier>` - Visual format hint for NeuroData.
    * `:: canonical_format: <format_identifier>` - Underlying structured format ID for NeuroData.
    * `:: status: <status_string>` - e.g., `:: status: draft`, `:: status: approved`

## File-Specific Conventions

### NeuroScript Files (`.ns`, `.ns.txt`, `.neuro`)

* **File Metadata:** Use the `:: key: value` standard at the top of the file (e.g., `:: version: ...`). The parser skips these lines.
* **`FILE_VERSION` Declaration:** ***Deprecated.*** Use `:: version:` instead for file content versioning.
* **Procedure `COMMENT:` Block:** Required per procedure. Contains structured metadata like `PURPOSE:`, `INPUTS:`, `OUTPUT:`, `ALGORITHM:`, `LANG_VERSION:`, `CAVEATS:`, `EXAMPLES:` (as specified in `docs/script spec.md` [cite: uploaded:neuroscript/docs/script spec.md]).
    * **`LANG_VERSION:`** (Optional but Recommended): Semantic version indicating the targeted NeuroScript language specification version.

### NeuroData Files (`.ndcl`, `.ndtb`, etc.)

* **File Metadata:** Use the `:: key: value` standard at the top (e.g., `:: version: ...`, `:: type: Checklist`).
* **`:: type:`** (Required) - Specifies the kind of NeuroData (e.g., `Checklist`, `Table`).

### Markdown Files (`.md`)

* **File Metadata:** Use the `:: key: value` standard at the **very top** of the file, before any Markdown content. This is consistent with other file types and allows tools like `TOOL.ExtractMetadata` to read them.
* **Rendering:** Be aware that standard Markdown viewers *may* render these `::` lines as plain text. UI-specific workarounds (like prepending `@@@` for display) might be necessary depending on the viewing tool.

### Embedded Blocks in Composite Documents (e.g., Markdown `.md`)

* **Fencing:** Use ``` followed by a language tag (e.g., `neuroscript`, `neurodata-checklist`).
* **Block Metadata:** Place `:: key: value` lines immediately *inside* the block, after the opening fence and before the block's main content [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md].
* **Example:**
    ```markdown
    Some introductory text.

    ```neuroscript
    :: id: proc-example-1
    :: version: 1.0.0
    :: lang_version: 1.1.0
    DEFINE PROCEDURE Example()
    COMMENT: ... ENDCOMMENT
    EMIT "Hello"
    END
    ```

    More text.

    ```neurodata-checklist
    :: id: checklist-example-1
    :: version: 0.1.0
    - [ ] Item 1
    ```
    ```