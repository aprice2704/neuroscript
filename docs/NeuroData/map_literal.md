:: version: 0.1.0
:: type: NSproject
:: subtype: spec
:: dependsOn: docs/metadata.md, docs/script spec.md, docs/neurodata_and_composite_file_spec.md
:: howToUpdate: Review dependency specs (NS map literals) and update syntax/examples if they change.

# NeuroScript Map Literal Data Format (`ns-map-literal`) Specification

## 1. Purpose

This specification defines how to represent structured key-value data using the native NeuroScript map literal syntax, typically embedded within a fenced code block in a composite document (like Markdown). This format is intended for scenarios where structured data (like configuration, definitions, or simple object representations) needs to be associated with other content, leveraging the existing NeuroScript parser rather than introducing external formats like YAML or JSON.

## 2. Relation to NeuroScript

The content of an `ns-map-literal` block **is** a single, valid NeuroScript map literal expression. Its syntax and semantics are directly governed by the NeuroScript Language Specification's definition of map literals [[script spec.md](./script spec.md)].

## 3. Syntax

An `ns-map-literal` data block consists of:
1.  An opening fence: `` ```ns-map-literal `` or `` ```neuroscript-map `` (using `ns-map-literal` is recommended for clarity).
2.  Optional block-level metadata lines (using `:: key: value` syntax [[metadata.md](./metadata.md)]). Recommended metadata includes `:: type: MapLiteralData` and `:: version: <semver>`.
3.  Optional comments (`#` or `--`) and blank lines.
4.  A single NeuroScript map literal expression: `{ "key1": <value_expr1>, "key2": <value_expr2>, ... }`.
    * Keys **must** be string literals (`"..."` or `'...'`).
    * Values can be any valid NeuroScript expression (literals: string, number, boolean; nested lists `[...]`; nested maps `{...}`). Note that variables or function calls within these value expressions are typically *not* evaluated when simply parsing the data structure; evaluation context depends on the tool consuming the map literal.
5.  A closing fence: ```` ``` ````.

## 4. Parsing and Tooling

Tools encountering a block tagged `ns-map-literal` should:
1.  Extract the content within the fences.
2.  Parse the content using the NeuroScript expression parser, specifically targeting the `map_literal` rule [[formal script spec.md](./formal script spec.md)].
3.  The result of the parse should be an Abstract Syntax Tree (AST) representation of the map literal, or an equivalent data structure (like Go's `map[string]interface{}`) representing the nested key-value pairs and literals found within.
4.  Further interpretation or evaluation of expressions *within* the map's values depends on the consuming tool's specific requirements.

## 5. Example

```markdown
Some context...

```ns-map-literal
:: type: MapLiteralData
:: version: 1.0.0
:: id: term-definitions-example

# Example map literal holding term definitions
{
  "TermA": {
    "description": "The first term.",
    "reference": "[ref:./glossary.md#term-a]",
    "value_type": "string"
  },
  "TermB": {
    "description": "The second term, with a list.",
    "aliases": ["AliasB1", "AliasB2"],
    "value_type": "integer"
  },
  "TermC": {
    "description": "A boolean flag." , # Example comment within map
    "value_type": "boolean",
    "default": true
  }
}
```

More context...
```