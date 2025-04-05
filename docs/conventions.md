# NeuroScript Conventions

Version: 0.2.2
DependsOn: docs/neuroscript overview.md
HowToUpdate: Manually review overview and update README summary/goals.

## File-Level Metadata

All project files (NeuroScript `.ns.txt`, NeuroData `.nd<type>`, Go `.go`, Markdown `.md`, etc.) should ideally contain the following metadata near the beginning of the file, formatted appropriately as comments if necessary for the file type (e.g., `// key: value` for Go, `# key: value` for scripts/data/markdown).

1.  **`Version:`** (Required)
    * Format: `Version: <semver_string>` (e.g., `Version: 0.1.0`)
    * Purpose: Semantic version of the *content* within this specific file. Tooling should aim to increment the patch number on change.

2.  **`DependsOn:`** (Optional)
    * Format: `DependsOn: <path_or_identifier_list>` (e.g., `DependsOn: pkg/core/parser_api.go, docs/spec.md`)
    * Purpose: Lists specific files or resources this file's content is directly derived from or depends on for correctness (excluding language-implicit dependencies).

3.  **`HowToUpdate:`** (Required if `DependsOn` is present)
    * Format: `HowToUpdate: <description_or_script_call>` (e.g., `HowToUpdate: Run 'go generate ./...'`, `HowToUpdate: Manually sync with spec.md changes.`)
    * Purpose: Describes how to update this file if its dependencies change.

## NeuroScript Files (`.ns`, `.ns.txt`, `.neuro`)

In addition to file-level metadata:

1.  **`FILE_VERSION` Declaration:** (Optional, distinct from `Version:` comment)
    * Syntax: `FILE_VERSION "semver_string"`
    * Purpose: Parsed by the NeuroScript interpreter to potentially influence execution or compatibility checks (as specified in `docs/script spec.md` [cite: uploaded:neuroscript/docs/script spec.md]). May be deprecated in favor of the comment convention.

2.  **Procedure `COMMENT:` Block:** (Required per procedure)
    * Contains structured metadata like `PURPOSE:`, `INPUTS:`, `OUTPUT:`, `ALGORITHM:`, `LANG_VERSION:`, `CAVEATS:`, `EXAMPLES:` (as specified in `docs/script spec.md` [cite: uploaded:neuroscript/docs/script spec.md]).
    * **`LANG_VERSION:`** (Optional but Recommended): Semantic version indicating the targeted NeuroScript language specification version.

## NeuroData Files (`.ndcl`, `.ndtb`, etc.)

In addition to file-level metadata:

1.  **`Type:`** (Required)
    * Format: `Type: <DataType>` (e.g., `Type: Checklist`, `Type: Table`)
    * Purpose: Specifies the kind of NeuroData contained within the file.

2.  **`Template:`** (Optional)
    * Format: `Template: <path_or_identifier>`
    * Purpose: Identifies the template file or definition this data instance is based on.

## Embedded Blocks in Composite Documents (e.g., Markdown `.md`)

When embedding NeuroScript code or NeuroData within other documents (primarily Markdown), use standard code fences with language identifiers and include metadata as comments immediately following the opening fence line.

1.  **Fencing:**
    * Use ``` followed by a language tag (e.g., `neuroscript`, `neurodata-checklist`, `neurodata-table`).

2.  **Metadata Comments:** Use standard comment markers appropriate for the embedded language immediately after the opening fence.
    * **Recommendation for NeuroScript within Markdown:** To avoid Markdown renderers potentially misinterpreting `#` as a heading, prefer using the ` -- ` (double-hyphen) comment marker for metadata lines within embedded ` ```neuroscript ` blocks. The `#` marker remains valid NeuroScript but is best avoided for metadata in this specific context. For `neurodata-*` blocks, `#` is generally safe.
    * Example:
        ```markdown
        ```neuroscript
        -- id: proc-example-1
        -- version: 1.0.0
        -- lang_version: 1.1.0
        DEFINE PROCEDURE Example()
        COMMENT: ... ENDCOMMENT
        EMIT "Hello"
        END
        ```
        ```neurodata-checklist
        # id: checklist-example-1
        # version: 0.1.0
        - [ ] Item 1
        ```
        ```

3.  **Standard Metadata Tags (using appropriate comment marker):**
    * **`id:`** (Recommended; Required for Tooling)
        * Format: `id: <unique_block_identifier>` (e.g., `id: update-proc-v1`, `id: task-checklist-main`)
        * Purpose: A unique identifier for this specific block *within the containing document*. Essential for extraction tools (`TOOL.ExtractFencedBlock`). Use descriptive, URL-safe characters (letters, numbers, hyphen, underscore).
    * **`version:`** (Required)
        * Format: `version: <semver_string>` (e.g., `version: 1.1.0`)
        * Purpose: Semantic version of the *content* of this specific embedded block.
    * **`template:`** (Optional)
        * Format: `template: <path_or_identifier>`
        * Purpose: Identifies the source template if this block instance was generated from one.
    * **`template_version:`** (Optional, Recommended if `template:` is used)
        * Format: `template_version: <semver_string>`
        * Purpose: The version of the source template used.
    * **`rendering_hint:`** (Optional, Primarily for NeuroData)
        * Format: `rendering_hint: <format_identifier>` (e.g., `rendering_hint: markdown-list`, `rendering_hint: simple-table`)
        * Purpose: Indicates how this specific embedded block is formatted or should be interpreted visually.
    * **`canonical_format:`** (Optional, Primarily for NeuroData)
        * Format: `canonical_format: <format_identifier>` (e.g., `canonical_format: structured-kv`, `canonical_format: relational-tuples`)
        * Purpose: Specifies the identifier for the underlying, potentially more structured, data format if this block is just one possible rendering of it.







