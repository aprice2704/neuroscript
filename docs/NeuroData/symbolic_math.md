:: type: NSproject
:: subtype: spec
:: version: 0.1.3
:: status: draft
:: dependsOn: docs/metadata.md, docs/references.md, docs/neurodata_and_composite_file_spec.md, docs/NeuroData/map_literal.md, docs/NeuroData/map_schema.md
:: howToUpdate: Refine function naming conventions, EBNF, specify supported functions/operators, detail tool behaviors and CAS integration strategy. Update attached block example if map_literal or map_schema specs change.

# NeuroData Symbolic Math Format (.ndmath) Specification

## 1. Purpose

NeuroData Symbolic Math (`.ndmath`) provides a format for representing mathematical expressions in a structured, unambiguous way suitable for symbolic manipulation by computer algebra systems (CAS) integrated via NeuroScript tools. It prioritizes structural clarity for machine processing over visual similarity to traditional mathematical notation.

## 2. Example `.ndmath` with Attached Definitions and Schema Reference

```ndmath
:: type: SymbolicMath
:: version: 0.1.2 # Version of this specific ndmath block content
:: notation: Functional
:: id: gr-field-eq-annotated
:: description: Conceptual GR Field Equations with term descriptions.

# The schema for the ns-map-literal block below is defined in Section 5.2

```funcmath
Equals(
  Add(
    Subtract(
      RicciTensor(mu, nu),
      Multiply(
        Divide(1, 2),
        ScalarCurvature(),
        MetricTensor(mu, nu)
      )
    ),
    Multiply(
      Lambda(),
      MetricTensor(mu, nu)
    )
  ),
  Multiply(
    Divide(
      Multiply(8, Pi(), G()), # Constants as functions
      Power(c(), 4)
    ),
    StressEnergyTensor(mu, nu)
  )
)
```
```ns-map-literal
# Attached block defining terms used in the expression above.
# It conforms to the schema defined in Section 5.2
:: type: MapLiteralData
:: version: 1.0.0
:: schema: [ref:this#math-term-def-schema] # Reference the schema block defined in Sec 5.2
{
  "RicciTensor": {
    "description": "Tensor representing curvature derived from the Riemann tensor.",
    "reference": "https://en.wikipedia.org/wiki/Ricci_curvature"
  },
  "MetricTensor": {
    "description": "Fundamental tensor defining spacetime geometry (g_µν).",
    "reference": "[ref:./tensors.md#metric]" # Link to another project file/block
  },
  "Lambda": {
    "description": "Cosmological Constant."
    # No reference needed/provided here, which is allowed by the schema
  },
  "mu": {
    "description": "Spacetime index (typically 0-3)",
    "reference": "[ref:./glossary.md#indices]"
  },
  "StressEnergyTensor": {
    "description": "Tensor describing density/flux of energy/momentum (T_µν)."
    # Missing reference (optional)
  }
  # ... definitions for nu, ScalarCurvature, Pi, G, c, Equals, Add, etc.
}
```
```

## 3. Design Choices

* **Functional Notation:** Chosen over S-expressions for improved readability for users familiar with programming language function calls. Chosen over presentational formats (like LaTeX or MathML Presentation) because the primary goal is representing the mathematical structure for computation, not visual layout. Chosen over semantic formats like Content MathML for relative simplicity in syntax and parsing, assuming a dedicated NeuroScript parser.
* **Pure Functional Form:** Operators (like `+`, `*`, `^`) are represented as functions (`Add`, `Multiply`, `Power`) to ensure an unambiguous tree structure suitable for parsing.
* **Term Definitions:** Descriptions, links, or other metadata about symbols or functions used within the expression can be provided in an attached `ns-map-literal` block [[map_literal.md](./NeuroData/map_literal.md)], conforming to the schema defined in Section 5.2.
* **Tool-Centric:** The format relies heavily on NeuroScript tools (`TOOL.Math*`) to perform actual symbolic manipulation (simplification, differentiation, etc.) and conversion to/from other formats (LaTeX, S-expressions, Infix). These tools would typically wrap external CAS libraries.

## 4. Syntax (`.ndmath`)

An `.ndmath` file or block consists of:
1.  **File-Level Metadata:** Optional `:: key: value` lines [[metadata.md](./metadata.md)].
2.  **Optional Term Definition Schema Block:** An optional fenced block tagged `ndmap_schema` defining the structure for term definitions. See Section 5.2.
3.  **Expression Block:** A single fenced block containing the mathematical expression represented in Functional Notation. Tag should be `funcmath` or similar. See Section 5.3.
4.  **Optional Attached Term Definitions Block:** An optional fenced block tagged `ns-map-literal` containing definitions for terms used in the expression block, conforming to the defined schema. See Section 5.4.

### 4.1 File-Level Metadata

Standard `:: key: value` lines. Recommended metadata includes:
* `:: type: SymbolicMath` (Required)
* `:: version: <semver>` (Required, version of the .ndmath content itself)
* `:: notation: Functional` (Required)
* `:: id: <unique_expr_id>` (Optional if referenced)
* `:: description: <text>` (Optional)

## 5. Detailed Syntax Components

### 5.1 Overview
(Section added to group detailed syntax elements previously under top-level Syntax)

### 5.2 Term Definition Schema Block (`ndmap_schema`)

* An optional block defining the structure of the term definitions map. Its syntax follows the `.ndmap_schema` format [[map_schema.md](./NeuroData/map_schema.md)].
* It's recommended to include this if using an attached term definitions block (Section 5.4).
* **Example Schema Definition:**
```ndmap_schema
# Embedded schema defining the structure for the value associated with each term key
# in the attached ns-map-literal block (e.g., the value for "RicciTensor", "mu").

:: type: NSMapSchema
:: version: 0.1.0
:: id: math-term-def-schema # ID for referencing within this file
:: description: Defines the structure expected for the definition map of a single term (symbol or function) used in .ndmath expressions.

KEY "description" # Key name must be literal string in schema
  DESC "A human-readable explanation of the term."
  TYPE string
  REQUIRED true # Description is required for clarity

KEY "reference"
  DESC "A URL or [ref:...] link to more detailed documentation or definition."
  TYPE string   # Type is string; content validation (URL/ref format) is separate
  REQUIRED false # Reference is optional
```

### 5.3 Expression Block (Functional Notation)

* The main mathematical expression is stored within a fenced block, typically with language tag `funcmath`.
* **Syntax:** Expressions are represented using a prefix functional notation: `FunctionName(arg1, arg2, ...)`
    * `FunctionName`: Represents a mathematical function (e.g., `Sin`, `Log`), operator (e.g., `Add`), or structural element (e.g., `Equals`, `Integrate`).
    * `arg1, arg2, ...`: Arguments (Literals, Symbols, Nested Calls).
* **Mapping Examples:**
    * `x + y` -> `Add(x, y)`
    * `2 * x` -> `Multiply(2, x)`
    * `x^2` -> `Power(x, 2)`
    * `sin(x)` -> `Sin(x)`
    * `df/dx` -> `Differentiate(f, x)`
    * `integrate(f(x), x)` -> `Integrate(f(x), x)` (Indefinite)
    * `integrate(f(x), x, 0, 1)` -> `Integrate(f(x), List(x, 0, 1))` (Definite)

### 5.4 Attached Term Definitions Block (`ns-map-literal`)

* Optionally, immediately following the `funcmath` block, an attached block tagged `ns-map-literal` can be used.
* The content **must** be a single, valid NeuroScript map literal [[script spec.md](./script spec.md)], conforming to the schema defined (usually in Section 5.2 or referenced via `:: schema:` metadata within this block, see [[map_literal.md](./NeuroData/map_literal.md)])
* Keys are term names (strings), values are maps usually containing `description` (string, required by schema in Sec 5.2) and `reference` (string, optional).
* Tools should parse this block using a NeuroScript parser and link definitions to terms in the main expression.
* See the example in Section 2.

## 6. EBNF Grammar (Draft - Needs Update for Attached Blocks & Schema)

```ebnf
math_file          ::= { metadata_line | comment_line | blank_line }
                      [ term_definition_schema_block ]
                      expression_block
                      [ term_definition_data_block ] ;

metadata_line      ::= optional_whitespace "::" whitespace key ":" value newline ;

term_definition_schema_block ::= optional_whitespace "```" "ndmap_schema" newline map_schema_content optional_whitespace "```" newline ;
expression_block   ::= optional_whitespace "```" language_tag? newline functional_expression optional_whitespace "```" newline ;
term_definition_data_block ::= optional_whitespace "```" "ns-map-literal" newline map_literal_content optional_whitespace "```" newline ;

functional_expression ::= function_call | symbol | literal ;

function_call     ::= identifier "(" [ argument_list ] ")" ;
argument_list     ::= functional_expression { "," functional_expression } ;

symbol            ::= identifier ;
literal           ::= number_literal | string_literal | boolean_literal ;

map_schema_content ::= (* Content parsed according to .ndmap_schema syntax *) ;
map_literal_content ::= (* Content parsed according to NeuroScript map literal syntax *) ;

identifier        ::= letter { letter | digit | "_" } ;

(* Define other terms *)
```
*(Note: This EBNF needs significant refinement.)*

## 7. Tooling Requirements

Effective use requires CAS-wrapping tools plus parsing capabilities for `.ndmap_schema` and `ns-map-literal`.

* **Core Manipulation Tools:** (Remain the same) `TOOL.MathSimplify`, `TOOL.MathExpand`, etc.
* **Conversion Tools:** (Remain the same) `TOOL.MathToLatex`, etc.
* **Parsing Requirement:** Tools interacting with `.ndmath` must be capable of:
    * Optionally parsing an `ndmap_schema` block (Section 5.2).
    * Parsing the primary `funcmath` block (Section 5.3).
    * Optionally detecting and parsing a subsequent `ns-map-literal` block (Section 5.4) using the NeuroScript expression parser.
    * Optionally validating the parsed map literal against the parsed schema (using the schema block's ID if referenced with `this#`).
    * Optionally using the extracted term definitions for validation, display, or further processing.