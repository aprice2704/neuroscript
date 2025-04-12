# NeuroData Symbolic Math Format (.ndmath) Specification

:: type: SymbolicMathFormatSpec
:: version: 0.1.0
:: status: draft
:: dependsOn: docs/metadata.md, docs/references.md, docs/neurodata_and_composite_file_spec.md
:: howToUpdate: Refine function naming conventions, EBNF, specify supported functions/operators, detail tool behaviors and CAS integration strategy.

## 1. Purpose

NeuroData Symbolic Math (`.ndmath`) provides a format for representing mathematical expressions in a structured, unambiguous way suitable for symbolic manipulation by computer algebra systems (CAS) integrated via NeuroScript tools. It prioritizes structural clarity for machine processing over visual similarity to traditional mathematical notation.

## 2. Design Choices

* **Functional Notation:** Chosen over S-expressions for improved readability for users familiar with programming language function calls. Chosen over presentational formats (like LaTeX or MathML Presentation) because the primary goal is representing the mathematical structure for computation, not visual layout. Chosen over semantic formats like Content MathML for relative simplicity in syntax and parsing, assuming a dedicated NeuroScript parser.
* **Pure Functional Form:** Operators (like `+`, `*`, `^`) are represented as functions (`Add`, `Multiply`, `Power`) to ensure an unambiguous tree structure suitable for parsing.
* **Tool-Centric:** The format relies heavily on NeuroScript tools (`TOOL.Math*`) to perform actual symbolic manipulation (simplification, differentiation, etc.) and conversion to/from other formats (LaTeX, S-expressions, Infix). These tools would typically wrap external CAS libraries.

## 3. Syntax (`.ndmath`)

An `.ndmath` file or block consists of:
1.  **File/Block-Level Metadata:** Optional `:: key: value` lines [cite: uploaded:neuroscript/docs/metadata.md].
2.  **Expression Block:** A single fenced block containing the mathematical expression represented in Functional Notation.

### 3.1 File/Block-Level Metadata

Standard `:: key: value` lines. Recommended metadata includes:
* `:: type: SymbolicMath` (Required)
* `:: version: <semver>` (Required)
* `:: notation: Functional` (Required)
* `:: id: <unique_expr_id>` (Optional if referenced)
* `:: description: <text>` (Optional)

### 3.2 Expression Block (Functional Notation)

* The expression is stored within a fenced block, typically with no language tag or potentially `funcmath`.
* **Syntax:** Expressions are represented using a prefix functional notation: `FunctionName(arg1, arg2, ...)`
    * `FunctionName`: Represents a mathematical function (e.g., `Sin`, `Log`, `Factorial`), operator (e.g., `Add`, `Subtract`, `Multiply`, `Divide`, `Power`), or structural element (e.g., `Equals`, `Integrate`, `Differentiate`, `List`, `Matrix`, tensor functions like `RicciTensor`). Function names should be chosen consistently (e.g., CamelCase).
    * `arg1, arg2, ...`: Arguments to the function/operator. Arguments can be:
        * **Literals:** Numbers (`123`, `3.14`, `-5`), strings (`"text"`).
        * **Symbols:** Variable names (`x`, `y`, `alpha`).
        * **Nested Function Calls:** `FunctionName(...)`.
* **Mapping Examples:**
    * `x + y` -> `Add(x, y)`
    * `2 * x` -> `Multiply(2, x)`
    * `x^2` -> `Power(x, 2)`
    * `sin(x)` -> `Sin(x)`
    * `df/dx` -> `Differentiate(f, x)`
    * `integrate(f(x), x)` -> `Integrate(f(x), x)` (Indefinite)
    * `integrate(f(x), x, 0, 1)` -> `Integrate(f(x), List(x, 0, 1))` (Definite - using List for variable/bounds)

## 4. EBNF Grammar (Draft)

```ebnf
math_file          ::= { metadata_line | comment_line | blank_line } expression_block ;

metadata_line      ::= optional_whitespace "::" whitespace key ":" value newline ;

expression_block   ::= optional_whitespace "```" [ language_tag ] newline functional_expression optional_whitespace "```" newline ;

functional_expression ::= function_call | symbol | literal ;

function_call     ::= identifier "(" [ argument_list ] ")" ;
argument_list     ::= functional_expression { "," functional_expression } ;

symbol            ::= identifier ; (* Rules for valid variable names *)
literal           ::= number_literal | string_literal | boolean_literal ; (* Bool maybe less common here *)

identifier        ::= letter { letter | digit | "_" } ;

(* Define: letter, digit, number_literal, string_literal, boolean_literal, language_tag, whitespace, newline, comment_line, blank_line, etc. Needs refinement for operator precedence if infix is ever allowed, but pure prefix avoids this. *)
```

## 5. Tooling Requirements

Effective use requires **new** NeuroScript tools wrapping a Computer Algebra System (CAS).

* **Core Manipulation Tools:**
    * `TOOL.MathSimplify(math_ref_or_content)` -> `ndmath_content`
    * `TOOL.MathExpand(math_ref_or_content)` -> `ndmath_content`
    * `TOOL.MathFactor(math_ref_or_content)` -> `ndmath_content`
    * `TOOL.MathDifferentiate(math_ref_or_content, variable_symbol)` -> `ndmath_content`
    * `TOOL.MathIntegrate(math_ref_or_content, variable_symbol)` -> `ndmath_content` (Indefinite)
    * `TOOL.MathIntegrateDefinite(math_ref_or_content, list_var_lower_upper)` -> `ndmath_content_or_number`
    * `TOOL.MathSolve(equation_ref_or_content, variable_symbol)` -> `list_of_solutions`
    * `TOOL.MathSubstitute(math_ref_or_content, substitution_map)` -> `ndmath_content`
    * `TOOL.MathEvalNumeric(math_ref_or_content, variable_map)` -> `number`
* **Conversion Tools:**
    * `TOOL.MathToLatex(math_ref_or_content)` -> `latex_string`
    * `TOOL.MathToSExpression(math_ref_or_content)` -> `s_expression_string`
    * `TOOL.MathToFunctional(math_ref_or_content)` -> `functional_string` (Identity if input is functional)
    * `TOOL.MathFromFunctional(functional_string)` -> `ndmath_content` (Parser)
    * `TOOL.MathFromSExpression(s_expression_string)` -> `ndmath_content`
    * `TOOL.MathFromInfix(infix_string)` -> `ndmath_content` (Requires robust infix parser)
    * `TOOL.MathFromLatex(latex_string)` -> `ndmath_content` (**Note:** Parsing LaTeX reliably is notoriously difficult and often ambiguous; this tool may be limited).

## 6. Example `.ndmath`

```ndmath
:: type: SymbolicMath
:: version: 0.1.0
:: notation: Functional
:: id: simple-polynomial
:: description: Representation of x^2 + 2*x + 1

```funcmath
Add(Power(x, 2), Multiply(2, x), 1)
```
```

```ndmath
:: type: SymbolicMath
:: version: 0.1.0
:: notation: Functional
:: id: definite-integral-example
:: description: Integral of Sin(x) from 0 to Pi

```funcmath
Integrate(Sin(x), List(x, 0, Pi())) # Assuming Pi() is a known constant function
```
```

```ndmath
:: type: SymbolicMath
:: version: 0.1.0
:: notation: Functional
:: id: gr-field-eq-sketch
:: description: Conceptual structure of GR Field Equations

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
