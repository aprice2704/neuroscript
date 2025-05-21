:: type: NeuroData
:: subtype: spec
:: version: 0.1.0
:: id: ndenum-spec-v0.1
:: status: draft
:: dependsOn: docs/metadata.md, docs/neurodata_and_composite_file_spec.md, docs/NeuroData/map_schema.md
:: howToUpdate: Review attributes, examples, EBNF. Ensure consistency with map_schema enum usage.

# NeuroData Enum Definition Format (.ndenum) Specification

## 1. Purpose

NeuroData Enum Definitions (`.ndenum`) provide a standalone, reusable format for defining named enumerated types (controlled vocabularies). Each member (value) of the enumeration can have optional associated metadata, including a human-readable label, a description, a numeric value, and an attached block for arbitrary structured data. This allows for consistent use of controlled values across different NeuroData files (like `.ndtable`, `.ndform`, `.ndmap_schema`) and NeuroScript procedures.

## 2. Example

```ndenum
:: type: EnumDefinition
:: version: 0.1.0
:: id: task-status-enum
:: description: Defines standard statuses for tasks in the system.

# Enum Values Defined Below

VALUE "pending"
  LABEL "Pending"
  DESC "Task has been created but not yet started."
  NUMERIC 1
  DATA ```json
  {
    "ui_color": "orange",
    "is_active": false
  }
  ```

VALUE "in-progress"
  LABEL "In Progress"
  DESC "Task is actively being worked on."
  NUMERIC 2
  DATA ```json
  {
    "ui_color": "blue",
    "is_active": true
  }
  ```

VALUE "completed"
  LABEL "Completed"
  DESC "Task finished successfully."
  NUMERIC 3
  DATA ```json
  {
    "ui_color": "green",
    "is_active": false
  }
  ```

VALUE "blocked"
  # LABEL defaults to "blocked" if omitted
  DESC "Task cannot proceed due to an issue."
  # NUMERIC is optional

VALUE "archived"
  LABEL "Archived"
  DESC "Task is closed and hidden from active views."
  NUMERIC -1
  # No DATA block needed here
```

## 3. Design Choices / Rationale

* **Consistency:** The syntax (`VALUE`, indented attributes) mirrors the `DEFINE ENUM` block used within the `.ndmap_schema` format. This promotes consistency and leverages existing parsing patterns.
* **Readability:** Keyword-driven attributes (`LABEL`, `DESC`, `NUMERIC`) enhance human readability.
* **Flexibility:** Optional attributes and the optional attached `DATA` block allow for simple or rich enum definitions as needed.
* **Reusability:** Standalone files allow enums to be defined once and referenced from multiple other NeuroData files or scripts using standard `[ref:<id>]` syntax.

## 4. Syntax / Format Definition

An `.ndenum` file consists of:
1.  Optional file-level metadata lines (`:: key: value`).
2.  Optional comments (`#` or `--`) and blank lines.
3.  One or more `VALUE` definition blocks.

### 4.1 File-Level Metadata

Standard `:: key: value` lines at the beginning of the file. Recommended metadata includes:
* `:: type: EnumDefinition` (Required)
* `:: version: <semver>` (Required, version of this enum definition content)
* `:: id: <unique_enum_id>` (Required if this enum will be referenced from elsewhere)
* `:: description: <text>` (Optional)

### 4.2 Enum Member Definition (`VALUE`)

Each member of the enumeration is defined by a block starting with `VALUE`.

* Format:
    ```ndenum
    VALUE "<keyword_string>"
      # Optional indented attribute lines
      [LABEL "<display_text>"]
      [DESC "<description_text>"]
      [NUMERIC <number_literal>]
    # Optional attached data block
    [DATA ```<format_tag>
    ... data content ...
    ```]
    ```
* `VALUE "<keyword_string>"`: Starts the definition. The `<keyword_string>` is the required, unique identifier for this enum member within the file (e.g., `"pending"`, `"active"`). It must be a valid string literal.
* **Indentation:** Attribute lines (`LABEL`, `DESC`, `NUMERIC`) and the `DATA` block fence MUST be indented relative to the `VALUE` line. Consistent indentation (e.g., 2 or 4 spaces) is recommended.

### 4.3 Optional Attributes

These attributes are defined on lines indented relative to the `VALUE` line:

* `LABEL "<display_text>"`: (Optional) A string literal providing a human-friendly label for the enum member. If omitted, tools should default to using the `<keyword_string>` from the `VALUE` line.
* `DESC "<description_text>"`: (Optional) A string literal providing a detailed description of the enum member's meaning or usage.
* `NUMERIC <number_literal>`: (Optional) Associates a numeric value (integer or float) with the enum member. The `<number_literal>` should be a valid NeuroScript number (e.g., `1`, `-10`, `3.14`).

### 4.4 Attached Data Block (`DATA`)

* (Optional) Immediately following the `VALUE` line and its indented attribute lines, an optional standard fenced data block can be attached using the `DATA` keyword on the line before the opening fence.
* Format:
    ```ndenum
    VALUE "..."
      ... attributes ...
    DATA ```<format_tag>
    { "structured": "data", "value": 123 }
    ```
    ```
* The `<format_tag>` (e.g., `json`, `yaml`) indicates the format of the content within the block.
* This block allows associating arbitrary structured data with an enum member.

## 5. EBNF Grammar (Draft)

```ebnf
enum_file         ::= { metadata_line | comment_line | blank_line } { value_definition_block } ;

metadata_line     ::= optional_whitespace "::" whitespace key ":" value newline ; (* As per metadata spec *)

value_definition_block ::= value_line { attribute_line } [ data_block ] ;

value_line        ::= optional_whitespace "VALUE" whitespace string_literal newline ;

attribute_line    ::= indentation ("LABEL"|"DESC"|"NUMERIC") whitespace attribute_value newline ;
indentation       ::= whitespace+ ;
attribute_value   ::= string_literal | number_literal ; (* Value type depends on keyword *)

data_block        ::= indentation "DATA" whitespace fenced_block ;
fenced_block      ::= "```" [ language_tag ] newline { text_line } optional_whitespace "```" newline ;

(* Define: string_literal, number_literal, key, value, whitespace, newline, comment_line, blank_line, language_tag, text_line, optional_whitespace *)
```
*(Note: This EBNF needs refinement, especially regarding indentation parsing)*

## 6. Tooling Requirements / Interaction

* **Parsing:** Tools need to parse the file structure, recognizing `VALUE` blocks and their indented attributes (`LABEL`, `DESC`, `NUMERIC`). They must also handle the optional attached `DATA` block, potentially parsing its content based on the format tag (e.g., using a JSON parser if ```json).
* **Validation:** Tools consuming `.ndenum` references (e.g., in `.ndtable` or `.ndform` schemas) should validate that provided values match one of the defined `<keyword_string>` values in the referenced `.ndenum` file.
* **Lookup:** Tools might provide functions to look up associated data (label, description, numeric value, data block content) based on a given keyword string.
* **UI Generation:** The `LABEL` and `DESC` attributes can be used by tools to generate more user-friendly interfaces (e.g., dropdown lists with descriptions in forms).
* **Formatting (`TOOL.fmt`):** A formatter should ensure consistent indentation and spacing.
