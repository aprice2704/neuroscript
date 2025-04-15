:: type: NSproject
:: subtype: spec
:: version: 0.1.0
:: id: spec-structure-guideline
:: status: draft
:: dependsOn: docs/metadata.md
:: howToUpdate: Review existing specs ensure they align or update this guideline.

# NeuroScript Specification Document Structure Guideline

## 1. Purpose

This document defines a standard structure for all specification files (files with `:: subtype: spec` metadata) within the NeuroScript project. The goal is to ensure consistency, improve readability for all users (human and AI), and make it easier to locate key information quickly.

## 2. Example Specification Structure

All specification documents should generally follow the structure demonstrated below. The Example section (Section 2) should always provide a concise, illustrative example of the format or concept being specified.

```markdown
# --- Start Example Spec File ---

:: type: NSproject # Or more specific like NeuroData
:: subtype: spec
:: version: 0.1.0
:: id: example-format-spec
:: status: draft
:: dependsOn: docs/metadata.md, ... # Other dependencies
:: howToUpdate: ...

# Title of the Specification (e.g., NeuroData Widget Format Spec)

## 1. Purpose

*Briefly state the goal and scope of the format or component being specified.*

## 2. Example

*Provide a clear, concise, and illustrative example of the format or concept.*
*This should give the reader an immediate understanding of what it looks like.*
```widget-format
# Example widget data
Widget {
  id: "widget-001",
  color: "blue",
  enabled: true
}
```

## 3. Design Choices / Rationale (Optional)

*Explain the key decisions made during the design.*
*Why was this approach chosen over alternatives?*
*What trade-offs were made?*

## 4. Syntax / Format Definition / Component Breakdown

*Provide the detailed definition of the syntax, format rules, or component parts.*
*Use subsections (e.g., 4.1, 4.2) for clarity.*
*Reference other specifications or standards where appropriate.*

### 4.1 Component A
*Details...*

### 4.2 Component B
*Details...*

## 5. EBNF Grammar (Optional)

*If applicable, provide an EBNF (Extended Backus-Naur Form) or similar formal grammar.*
```ebnf
widget ::= 'Widget' '{' ... '}' ;
...
```

## 6. Tooling Requirements / Interaction (Optional)

*Describe how software tools are expected to interact with this format.*
*What parsing logic is needed?*
*Are there specific validation rules tools should enforce?*
*Are specific NeuroScript TOOLs expected to consume or produce this format?*

# --- End Example Spec File ---
```

## 3. Standard Section Definitions

The following sections should be used, in this order:

1.  **Purpose:** (Required) Clearly and concisely state what the specification defines and its intended scope.
2.  **Example:** (Required) Provide at least one clear, representative example of the format or concept being specified. This should be sufficient for a reader to get a basic understanding at a glance.
3.  **Design Choices / Rationale:** (Optional but Recommended) Explain the reasoning behind key design decisions. This helps others understand the context and potential trade-offs.
4.  **Syntax / Format Definition:** (Required) This is the core section detailing the rules, structure, components, and semantics of the item being specified. Use subsections for clarity.
5.  **EBNF Grammar:** (Optional) Include if a formal grammar aids in defining the syntax precisely.
6.  **AI Reading** This section should give clear, concise instructions for how AIs (such as LLMs) should understand the contents of the file. This may be included in prompts to the AI.
6.  **AI Writing** This section should give clear, concise additional instructions for how AIs (such as LLMs) should write contents of the file, such as cross checks to perform. This may be included in prompts to the AI.
7.  **Tooling Requirements / Interaction:** (Optional but Recommended for data formats) Describe how tools should parse, validate, or otherwise interact with the format. This section will be used primarily when building computer tools to manipulate the format.

## 4. Metadata Requirements

All specification files must begin with standard file-level metadata as defined in [[metadata.md](./metadata.md)]. This must include:
* `:: type: NSproject` (or a more specific type if applicable, like `NeuroData`)
* `:: subtype: spec`
* `:: version: <semver>`
* `:: status: <status_string>` (e.g., `draft`, `approved`)
* `:: dependsOn: ...` (List dependencies, including `docs/metadata.md` and this document)
* `:: howToUpdate: ...` (Instructions for maintenance)
* `:: id: <unique_spec_id>` (A unique identifier for the specification)

Adhering to this structure will help maintain consistency across all NeuroScript project specification documents.