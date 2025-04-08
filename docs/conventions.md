:: version: 0.3.0
:: dependsOn: docs/neuroscript overview.md, docs/neurodata_and_composite_file_spec.md
:: howToUpdate: Review the referenced documents and ensure this file accurately reflects the current metadata standards, file/procedure structure, and embedded block conventions.

# NeuroScript Conventions

## Metadata Standard (`:: key: value`)

All project files (NeuroScript `.ns.txt`, NeuroData `.nd*`, Go `.go`, Markdown `.md`, etc.) and embedded code/data blocks should use the following metadata format where applicable.

* **Prefix:** Metadata lines must start with `::` (colon, colon) followed by at least one space. Optional leading whitespace before `::` is allowed. [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md]
* **Structure:** `:: key: value` [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md]
* **Key:** Immediately follows `:: ` and precedes the first `:`. Valid characters are letters, numbers, underscore, period, hyphen (`[a-zA-Z0-9_.-]+`). Whitespace around the key (after `:: `) is tolerated. [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md]
* **Separator:** A single colon `:` separates the key and value. Whitespace around the colon is tolerated. [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md]
* **Value:** Everything after the first colon, with leading/trailing whitespace trimmed. [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md]
* **Location:** Metadata lines must appear consecutively at the **beginning** of the file or embedded block, before any functional content (e.g., before `DEFINE PROCEDURE`, before NeuroData items like `- [ ]`). [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md]
* **Comments/Blank Lines:** Standard comment lines (`#` or `--`) and blank lines *are* permitted between metadata lines, but still must appear *before* the main content begins. [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md]
* **Required Tags (File/Block Level):**
    * `:: version: <semver_string>` (e.g., `:: version: 0.1.0`) - Semantic version of the *content* of this specific file or block. Tooling should aim to increment the patch number on change.
* **Optional Tags (Examples):**
    * `:: id: <unique_identifier>` - Unique ID for embedded blocks within a composite document. Essential for tools like `TOOL.BlocksExtractAll`. Use URL-safe characters.
    * `:: dependsOn: <path_or_identifier_list>` - Files/resources this content directly depends on.
    * `:: howToUpdate: <description_or_script_call>` - How to update if dependencies change (required if `dependsOn` is present).
    * `:: template: <path_or_identifier>` - Source template for NeuroData instances.
    * `:: template_version: <semver_string>` - Version of the source template.
    * `:: rendering_hint: <format_identifier>` - Visual format hint for NeuroData.
    * `:: canonical_format: <format_identifier>` - Underlying structured format ID for NeuroData.
    * `:: status: <status_string>` - e.g., `:: status: draft`, `:: status: approved`

## File-Specific Conventions

### NeuroScript Files (`.ns`, `.ns.txt`, `.neuro`)

* **Metadata:** Use the `:: key: value` standard at the top of the file (e.g., `:: version: ...`).
* **`FILE_VERSION` Declaration:** ***Deprecated.*** Use `:: version:` instead for file content versioning.
* **Procedure `COMMENT:` Block:** Required per procedure. Contains structured metadata like `PURPOSE:`, `INPUTS:`, `OUTPUT:`, `ALGORITHM:`, `LANG_VERSION:`, `CAVEATS:`, `EXAMPLES:` (as specified in `docs/script spec.md` [cite: uploaded:neuroscript/docs/script spec.md]).
    * **`LANG_VERSION:`** (Optional but Recommended): Semantic version indicating the targeted NeuroScript language specification version.

### NeuroData Files (`.ndcl`, `.ndtb`, etc.)

* **Metadata:** Use the `:: key: value` standard at the top (e.g., `:: version: ...`, `:: type: Checklist`).
* **`:: type:`** (Required) - Specifies the kind of NeuroData (e.g., `Checklist`, `Table`).

### Embedded Blocks in Composite Documents (e.g., Markdown `.md`)

* **Fencing:** Use ``` followed by a language tag (e.g., `neuroscript`, `neurodata-checklist`). [cite: uploaded:neuroscript/docs/conventions.md]
* **Metadata:** Place `:: key: value` lines immediately after the opening fence, before the block's main content. [cite: uploaded:neuroscript/docs/neurodata_and_composite_file_spec.md]
* **Example:**
    ```markdown
    ```neuroscript
    :: id: proc-example-1
    :: version: 1.0.0
    :: lang_version: 1.1.0
    DEFINE PROCEDURE Example()
    COMMENT: ... ENDCOMMENT
    EMIT "Hello"
    END
    ```
    ```neurodata-checklist
    :: id: checklist-example-1
    :: version: 0.1.0
    - [ ] Item 1
    ```
    ```