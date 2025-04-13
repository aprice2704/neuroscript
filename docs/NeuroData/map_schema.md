:: type: NSproject
:: subtype: spec
:: version: 0.1.0
:: id: ndmap-schema-spec
:: status: draft
:: dependsOn: docs/metadata.md, docs/script spec.md, docs/NeuroData/map_literal.md, docs/NeuroData/references.md
:: howToUpdate: Review syntax, especially enum definitions and TYPE references, for clarity and consistency with other specs.

# NeuroData Map Schema Format (.ndmap_schema) Specification

## 1. Purpose

NeuroData Map Schema (`.ndmap_schema`) defines the expected structure, constraints, and documentation for key-value data represented using the NeuroScript Map Literal Data Format [[map_literal.md](./NeuroData/map_literal.md)]. It acts as a schema, allowing tools to validate map literal instances, understand the purpose of different keys, enforce required fields, and manage named, reusable enumerated value lists with descriptions. It prioritizes simplicity and readability, borrowing syntax conventions from other NeuroData formats like `.ndform` and `.ndtable`.

## 2. Example

```ndmap_schema
# File: docs/NeuroData/schemas/metadata_schema.ndmap_schema

:: type: NSMapSchema
:: version: 0.1.0
:: id: basic-metadata-schema
:: description: Defines common metadata keys using named enums.
:: depth: 1 # Default depth limit for values

# --- Enum Definitions ---

DEFINE ENUM LifecycleStatus
  VALUE "draft"
    DESC "Initial, non-final version."
  VALUE "review"
    DESC "Ready for review."
  VALUE "approved"
    DESC "Finalized and approved."
  VALUE "deprecated"
    DESC "No longer recommended."

DEFINE ENUM ContentType
  VALUE "NSproject"
    DESC "General project file."
  VALUE "NeuroScript"
    DESC "Executable NeuroScript code."
  # ... other types ...

# --- Key Definitions ---

KEY "version" # Keys are strings
  DESC "Semantic version."
  TYPE string
  REQUIRED true

KEY "type"
  DESC "Primary type of the file/block content."
  TYPE enum(ContentType) # Reference the defined enum
  REQUIRED true

KEY "status"
  DESC "Lifecycle status."
  TYPE enum(LifecycleStatus) # Reference the defined enum
  REQUIRED false

KEY "author_details"
  DESC "Information about the author."
  TYPE [ref:author-schema] # Reference another schema for nested structure
  REQUIRED false
  # DEPTH constraint could be defined in author-schema instead

```

```ns-map-literal
# Example data instance conforming to the schema above
{
  "version": "1.2.3",
  "type": "NeuroScript",
  "status": "approved",
  "author_details": { "name": "A. Turing", "email": "alan@example.com" } # Assumes author-schema defines name & email
}
```

## 3. Syntax

An `.ndmap_schema` file consists of the following sections:
1.  **File-Level Metadata:** Optional `:: key: value` lines [[metadata.md](./metadata.md)]. Recommended metadata includes `:: type: NSMapSchema`, `:: version: <semver>`, `:: id: <schema_id>`, and optionally `:: depth: <number>` to limit overall nesting.
2.  **Enum Definitions (Optional):** Zero or more `DEFINE ENUM` blocks, each defining a named, reusable list of allowed values and their descriptions.
3.  **Key Definitions:** One or more `KEY` definitions describing the expected keys in the map literal data and the constraints on their associated values.
4.  **Comments/Blank Lines:** Allowed between metadata and definitions, and between definitions (`#` or `--`).

### 3.1 File-Level Metadata

Standard `:: key: value` lines [[metadata.md](./metadata.md)]. Recommended:
* `:: type: NSMapSchema` (Required)
* `:: version: <semver>` (Required)
* `:: id: <unique_schema_id>` (Required if referenced)
* `:: description: <text>` (Optional)
* `:: depth: <number>` (Optional) - If present, suggests a maximum nesting depth for the entire map structure defined by this schema. Validation tools may use this.

### 3.2 Enum Definition (`DEFINE ENUM`)

* Format:
  ```ndmap_schema
  DEFINE ENUM <enum_name>
    VALUE "<enum_value_1>"
      DESC "<Description of enum_value_1>"
    VALUE "<enum_value_2>"
      DESC "<Description of enum_value_2>"
    # ... more VALUE/DESC pairs
  ```
* `DEFINE ENUM <enum_name>`: Starts the definition. `<enum_name>` must be a unique identifier within the schema file (`[a-zA-Z_][a-zA-Z0-9_]*`).
* `VALUE "<enum_value_string>"`: (Indented) Defines an allowed literal value for the enum (typically a string).
* `DESC "<description>"`: (Indented further) Provides the human-readable description for the preceding `VALUE`. This is crucial for understanding the enum's meaning.

### 3.3 Key Definition (`KEY`)

* Format:
  ```ndmap_schema
  KEY "<key_name_string>"
    # Indented attribute lines
    DESC "<description>"
    TYPE <type_spec>
    REQUIRED <true|false>
    DEPTH <number> # Optional depth limit for this key's value
  ```
* `KEY "<key_name_string>"`: Starts the definition. `<key_name_string>` is the literal string key expected in the map data instance.
* **Attribute Lines** (Indented): Define properties and constraints for the *value* associated with this key.
    * `DESC "<description>"`: (Recommended) Human-readable description of the key's purpose.
    * `TYPE <type_spec>`: (Required) Specifies the expected data type or enum reference for the value. Supported types:
        * Base types: `string`, `int`, `float`, `bool`, `list`, `map`, `any`.
        * Enum reference: `enum(<enum_name>)` - References a named enum defined elsewhere in the file using `DEFINE ENUM <enum_name>`. The value must match one of the `VALUE`s defined in that enum.
        * Schema reference: `[ref:<schema_id>]` or `[ref:path/to/schema.ndmap_schema]` - Indicates the value should be a map conforming to another `.ndmap_schema` schema (used for defining nested structures).
    * `REQUIRED <true|false>`: (Optional) Specifies if the key must be present in the map data instance. Defaults to `false`.
    * `DEPTH <number>`: (Optional) Specifies the maximum nesting depth allowed specifically for the *value* of this key (relevant if `TYPE` is `list`, `map`, `any`, or references another schema). Overrides file-level `:: depth` for this key.

## 4. Tooling Interaction

Tools interacting with `ns-map-literal` data and `.ndmap_schema` schemas should:
1.  Parse the `.ndmap_schema` schema file to understand the expected structure, types, required keys, and enum definitions.
2.  Parse the `ns-map-literal` data block using the NeuroScript expression parser.
3.  **Validation:** Compare the parsed map literal data against the parsed schema:
    * Check for missing required keys.
    * Check if all present keys are defined in the schema (optional strict mode).
    * Validate the data type of each value against the `TYPE` specified in the schema.
    * If `TYPE` is `enum(<name>)`, verify the value exists in the named `DEFINE ENUM` block.
    * If `TYPE` is `[ref:...]`, recursively validate the nested map value against the referenced schema.
    * Enforce `DEPTH` constraints if specified.
4.  **Documentation/UI:** Use the `DESC` attributes from the schema (for both keys and enum values) to provide context to users or generate documentation.