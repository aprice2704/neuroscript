# NeuroData Simple List Format (.ndlist) Specification

:: type: ListFormatSpec
:: version: 0.1.0
:: status: draft
:: dependsOn: docs/metadata.md, docs/references.md, docs/NeuroData/checklist.md, docs/neurodata_and_composite_file_spec.md
:: howToUpdate: Refine syntax, parsing options, tooling descriptions, EBNF, examples.

## 1. Purpose

NeuroData Simple Lists (`.ndlist`) provide a minimal, human-readable plain-text format for representing ordered or hierarchical lists of items, primarily intended for simple text entries or references. It serves as a simpler alternative to the NeuroData Checklist format [cite: uploaded:neuroscript/docs/NeuroData/checklist.md] when status tracking (`[ ]`, `[x]`) is not required.

## 2. Relation to Checklist Format (`.ndcl`)

The `.ndlist` format uses a syntax visually similar to `.ndcl` but removes the status marker (`[...]` or `|...|`). It focuses purely on the list items and their hierarchical structure (if any), indicated by indentation.

## 3. Syntax

An `.ndlist` file or block consists of:
1.  Optional file-level metadata lines (using `:: key: value` syntax [cite: uploaded:neuroscript/docs/metadata.md]).
2.  Optional comments (`#` or `--`) and blank lines.
3.  A series of list item lines.

### 3.1 File/Block-Level Metadata

Standard `:: key: value` lines [cite: uploaded:neuroscript/docs/metadata.md]. Recommended metadata includes:
* `:: type: SimpleList` (or `ItemList`) (Required)
* `:: version: <semver>` (Required)
* `:: id: <unique_list_id>` (Optional if referenced)
* `:: description: <text>` (Optional)

### 3.2 List Item Line

* Format: `Indentation - Item Text`
* `Indentation`: Zero or more spaces or tabs. The level of indentation defines the item's parent in the hierarchy (the nearest preceding item line with less indentation). Consistent indentation is recommended.
* `- `: A literal hyphen followed by a single space MUST precede the item text.
* `Item Text`: The content of the list item. This is treated as a raw string and can include any characters, including NeuroScript References (`[ref:...]` [cite: generated previously in `docs/references.md`]). Leading/trailing whitespace *after* the required space is part of the item text.

### 3.3 Comments and Blank Lines

Lines starting with `#` or `--` (after optional whitespace) are comments and are ignored by the parser. Blank lines are also ignored and do not typically contribute to the structure, although tools *could* optionally preserve them if needed for specific rendering.

## 4. EBNF Grammar (Draft)

```ebnf
list_file         ::= { metadata_line | comment_line | blank_line } { list_item_line } ;

metadata_line     ::= optional_whitespace "::" whitespace key ":" value newline ; (* As per references spec *)

list_item_line    ::= indentation "-" whitespace item_text newline ;
indentation       ::= { " " | "\t" } ; (* Parsed to determine level *)
item_text         ::= rest_of_line ; (* Raw text content *)

(* Define: key, value, rest_of_line, whitespace, newline, comment_line, blank_line *)
```

## 5. Tool Interaction

Interaction with `.ndlist` data primarily involves parsing the content into a usable structure.

* **Parsing (`TOOL.ParseList` - hypothetical):** A dedicated tool would parse the `.ndlist` content (provided as a string or via a `[ref:...]`).
    * **Input:** `list_content_or_ref` (String or Reference)
    * **Options (Optional):**
        * `output_format` (String): `"hierarchy"` (default), `"flat_text"`, `"flat_indent"`.
    * **Output:**
        * **Default (`"hierarchy"`):** Returns a nested structure representing the tree inferred from indentation (e.g., a list of maps, where each map has `"text"` and `"children"` keys).
        * **`"flat_text"`:** Returns a simple flat list of strings, containing the text of each item in document order.
        * **`"flat_indent"`:** Returns a flat list of maps, where each map contains `{"text": string, "indent": int}`.
    * **Reference Handling:** The parser reads `Item Text` literally. It does *not* automatically resolve `[ref:...]` strings found within item text. Resolution must be handled by subsequent steps in the calling NeuroScript procedure if needed.
* **Hierarchy Check (`TOOL.IsListHierarchical` - hypothetical):** A potential helper tool, as you suggested (perhaps named like this), could take the *parsed data* (e.g., the output from `TOOL.ParseList` in `"flat_indent"` format) and return `true` if items have varying indentation levels, `false` otherwise. This confirms if a list uses nesting.

## 6. Example

```ndlist
:: type: SimpleList
:: version: 0.1.0
:: description: List of relevant specification documents.

- Core Language
  - [ref:docs/script spec.md]
  - [ref:docs/formal script spec.md]
- NeuroData Formats
  - [ref:docs/NeuroData/checklist.md]
  - [ref:docs/neurodata/table.md]
  - [ref:docs/neurodata/graph.md]
  - [ref:docs/neurodata/tree.md]
  - [ref:docs/neurodata/form.md]
  - [ref:this] # Reference to this list spec itself
- Supporting Concepts
  - [ref:docs/metadata.md]
  - [ref:docs/references.md]
- Tooling
  - [ref:pkg/core/interpreter.go]
  - [ref:pkg/neurogo/app.go]
```