# NeuroData Form Format (.ndform) Specification

:: type: FormFormatSpec
:: version: 0.1.0
:: status: draft
:: dependsOn: docs/metadata.md, docs/references.md, docs/neurodata/table.md, docs/neurodata_and_composite_file_spec.md, pkg/neurodata/blocks/blocks_extractor.go
:: howToUpdate: Refine field attributes, types, validation rules, EBNF, NS fragment scope/allowlist, examples. Define .ndobj format.

## 1. Purpose

NeuroData Forms (`.ndform`) define the structure, presentation, validation rules, and associated metadata for a data entry form. They act as a schema or template for data capture. The `.ndform` file itself does not contain the filled-in instance data, but rather describes the fields and their properties. It aims to be human-readable while providing enough structure for tools and AI to render, validate, and process form instances.

## 2. Relation to Form Data (`.ndobj`)

The `.ndform` file defines the *schema* of the form. The actual *data* entered into an instance of the form should be stored separately, potentially using a simple key-value format designated `.ndobj` (NeuroData Object).

An `.ndobj` instance should reference the `.ndform` it corresponds to via metadata (e.g., `:: formRef: [ref:path/to/form.ndform#form-id]`).

It is expected that `.ndform` definitions and `.ndobj` data instances will often be bundled together within composite documents (e.g., Markdown files), identifiable via `:: type: Form` and `:: type: Object` (or similar) metadata and extracted using block processing tools [cite: uploaded:neuroscript/pkg/neurodata/blocks/blocks_extractor.go].

## 3. Syntax (`.ndform`)

An `.ndform` file or block consists of:
1.  **File/Block-Level Metadata:** Optional `:: key: value` lines [cite: uploaded:neuroscript/docs/metadata.md].
2.  **Field Definitions:** A series of field definitions describing the form structure.
3.  **Comments/Blank Lines:** Allowed between metadata and fields, and between fields.

### 3.1 File/Block-Level Metadata

Standard `:: key: value` lines. Recommended metadata includes:
* `:: type: Form` (Required)
* `:: version: <semver>` (Required)
* `:: id: <unique_form_id>` (Required if referenced)
* `:: title: "<Form Title>"` (Optional, human-readable title)
* `:: description: <text>` (Optional)

### 3.2 Field Definition

Each field is defined using a `FIELD` line followed by indented attribute lines:
* `FIELD <field_id>`: Starts the definition. `<field_id>` must be a unique identifier within the form (`[a-zA-Z_][a-zA-Z0-9_]*`).
* Attribute Lines (Indentation optional but recommended for readability): Define properties using keywords followed by their value. Common attributes include:
    * `LABEL "<text>"`: The human-readable label or question for the field.
    * `TYPE <type_spec>`: Data type. Supported types: `string`, `text` (multi-line string), `int`, `float`, `bool`, `timestamp`, `email`, `url`, `block_ref`, `enum("val1", "val2")`.
    * `VALUE <literal | template_var | block>`: The default or current value. For multi-line text, use a nested fenced block (see 3.3). May contain template variables (`{{...}}`) if the form is processed by a templating engine.
    * `HELP <text | block>`: Explanatory text for the user. For multi-line help, use a nested fenced block (see 3.3).
    * `VALIDATION <rules...>`: Space-separated standard validation rules (see Section 3.4).
    * `DEFAULT <value>`: Default value (static literal or `NOW` for timestamp).
    * `READONLY <true|false>`: Indicates if the field value can be edited.
    * `VISIBLE <true|false>`: Indicates if the field should be initially visible.
    * (Future attributes for NS fragments: `VALIDATE_NS`, `DEFAULT_NS`, `CALCULATE_NS`, `READONLY_NS`, `VISIBLE_NS` - see Section 4).
* **Attached Blocks:** A standard fenced block (` ```...``` `) placed immediately after all attribute lines for a `FIELD` is considered attached to that field (e.g., for complex properties via `TYPE: block_ref` or just providing context).

### 3.3 Multi-line Values (VALUE, HELP)

To define multi-line text for `VALUE` or `HELP` attributes, use a nested fenced block immediately following the attribute line:
```ndform
FIELD notes
  LABEL "Additional Notes"
  TYPE text
  VALUE ```text
Line 1 of the value.
Line 2 of the value.
```
  HELP ```markdown
Please provide any relevant details.
* Use bullet points if needed.
* Markdown is supported here.
```
```

### 3.4 Standard Validation Rules

Similar to `.ndtable`:
* `NOT NULL`: Field must have a non-empty value.
* `UNIQUE`: (Context-dependent) Value should be unique relative to other instances processed together.
* `REGEX("pattern")`: String value must match the Go regex pattern.
* `MIN(value)` / `MAX(value)`: For numeric/timestamp types.

## 4. NeuroScript Fragment Integration (Future v0.1.0+)

*(Note: The following attributes are planned features and not part of the v0.1.0 specification.)*

Future versions may allow embedding restricted NeuroScript expressions for dynamic behavior:
* `VALIDATE_NS "<expression>"`: Expression must evaluate to true for the field value (using `{{value}}`) to be valid.
* `DEFAULT_NS "<expression>"`: Expression result provides the default value.
* `CALCULATE_NS "<expression>"`: Field value is dynamically calculated (e.g., `CALCULATE_NS("'{{row.first}}'+' '+'{{row.last}}'"`).
* `READONLY_NS "<expression>"`: Field is readonly if the expression evaluates to true.
* `VISIBLE_NS "<expression>"`: Field is visible if the expression evaluates to true.

**Execution Context:** These NS fragments would execute in a highly restricted sandbox:
* **Scope:** Access only to the current form instance data (e.g., via `data.field_id` or `row.field_id` for calculated fields) and the specific field's value (`value`). No access to global interpreter state or `LAST`.
* **Allowlist:** A minimal allowlist of safe, pure functions/operators (basic math, string manipulation, boolean logic). **No I/O** (`ReadFile`, `WriteFile`), **no `ExecuteCommand`**, **no `CALL LLM`**, etc.

## 5. EBNF Grammar (Draft)

```ebnf
form_file         ::= { metadata_line | comment_line | blank_line } { field_definition } ;

metadata_line     ::= optional_whitespace "::" whitespace key ":" value newline ; (* As per references spec *)

field_definition  ::= optional_whitespace "FIELD" whitespace field_id newline { field_attribute_line } [ fenced_block ] ;
field_id          ::= identifier ;

field_attribute_line ::= indentation attribute_keyword whitespace attribute_value newline | multi_line_attribute ;

attribute_keyword ::= "LABEL" | "TYPE" | "VALUE" | "HELP" | "VALIDATION" | "DEFAULT" | "READONLY" | "VISIBLE" | "VALIDATE_NS" | "DEFAULT_NS" | "CALCULATE_NS" | "READONLY_NS" | "VISIBLE_NS" ; (* NS ones are future *)
attribute_value   ::= rest_of_line ; (* Specific parsing depends on keyword, includes literals, type specs, validation rules etc. *)

multi_line_attribute ::= indentation ("VALUE" | "HELP") whitespace "```" [ language_tag ] newline { text_line } optional_whitespace "```" newline ;

fenced_block      ::= optional_whitespace "```" [ language_tag ] newline { text_line } optional_whitespace "```" newline ;

(* Define: identifier, key, value, rest_of_line, type_spec, validation_rules, language_tag, text_line etc. *)
```

## 6. Tool Interaction

Tools would interact with forms:
* **Rendering:** A tool could take an `.ndform` definition and an optional `.ndobj` data instance to render an interactive form (e.g., in a terminal UI or web page).
* **Validation:** `TOOL.ValidateFormData(form_ref, data_obj)` could validate the data in an `.ndobj` against the rules in its corresponding `.ndform`.
* **Extraction:** `TOOL.ExtractFormData(form_ref)` could perhaps prompt a user to fill a form and return the resulting `.ndobj`.
* **Templating:** `TOOL.RenderTemplate` could potentially use form data (`.ndobj`) as input.

## 7. Example `.ndform`

```ndform
:: type: Form
:: version: 0.1.0
:: id: bug-report-form
:: title: "Bug Report"

FIELD report_id
  LABEL "Report ID"
  TYPE string
  DEFAULT_NS "CALL TOOL.GenerateUUID()" # Future example
  READONLY true
  HELP "Unique identifier for this report."

FIELD summary
  LABEL "Summary"
  TYPE string
  VALIDATION NOT NULL
  HELP "Provide a one-line summary of the issue."

FIELD component
  LABEL "Component"
  TYPE enum("Core Interpreter", "Tooling", "Agent Mode", "NeuroData", "Other")
  DEFAULT "Core Interpreter"

FIELD steps_to_reproduce
  LABEL "Steps to Reproduce"
  TYPE text
  VALIDATION NOT NULL
  HELP ```markdown
Please list the exact steps needed to trigger the bug.
1. Step one...
2. Step two...
3. ...
```

FIELD logs # Example attaching a block for context/reference
  LABEL "Relevant Logs (Optional)"
  TYPE block_ref # Value would be a [ref:...] or empty
  HELP "Attach relevant log output below or reference a block ID."
```text
:: id: sample-log-format
Paste logs here...
```

FIELD severity
  LABEL "Severity (1-5)"
  TYPE int
  DEFAULT 3
  VALIDATION NOT NULL VALIDATE_NS("{{value}} >= 1 AND {{value}} <= 5") # Future Example
