# NeuroData Graph Format (.ndgraph) Specification

:: type: GraphFormatSpec
:: version: 0.1.1
:: status: draft
:: dependsOn: docs/metadata.md, docs/neurodata_and_composite_file_spec.md
:: howToUpdate: Review decisions, update EBNF, ensure examples match spec.

## 1. Purpose

NeuroData Graphs (`.ndgraph`) provide a simple, human-readable plain-text format for representing node-edge graph structures. The format prioritizes human readability while being machine-parseable. It is designed primarily to be read by humans and updated by tools or AI, supporting explicit bidirectional link representation for clarity.

## 2. Syntax

A `.ndgraph` file consists of:
1.  Optional file-level metadata lines (using `:: key: value` syntax [cite: uploaded:neuroscript/docs/metadata.md]).
2.  Optional comments (`#` or `--`) and blank lines.
3.  A series of `NODE` definitions.
4.  Each `NODE` definition can be followed by:
    a. An optional attached fenced data block (for complex properties).
    b. Indented `EDGE` definitions representing connections (outgoing, incoming, or undirected) associated with that node.

### 2.1 File-Level Metadata

Standard `:: key: value` lines at the very beginning of the file. Recommended metadata includes:
* `:: type: Graph` (Required)
* `:: version: <semver>` (Required)
* `:: id: <unique_graph_id>` (Optional but recommended)
* `:: directed: <true|false>` (Optional, defaults to true. Affects edge interpretation.)
* `:: description: <text>` (Optional)

### 2.2 Node Definition Line

* Format: `NODE NodeID [Optional Simple Properties]`
* `NODE`: Keyword indicating a node definition.
* `NodeID`: A unique identifier for the node within the graph. Must start with a letter or underscore, followed by letters, numbers, or underscores (`[a-zA-Z_][a-zA-Z0-9_]*`). IDs are case-sensitive.
* `[Optional Simple Properties]`: An optional block enclosed in square brackets `[]` containing simple key-value properties for the node (see Section 2.4). Complex properties should use an Attached Data Block (see Section 2.5).

### 2.3 Edge Definition Lines

* Format: `Indentation EdgeMarker TargetNodeID [Optional Simple Properties]`
* **Indentation:** One or more spaces or tabs MUST precede the edge marker. This indicates the edge belongs to the preceding `NODE` definition.
* **EdgeMarker:**
    * `->`: Indicates an outgoing directed edge from the parent `NODE` to the `TargetNodeID`.
    * `<-`: Indicates an incoming directed edge from the `TargetNodeID` to the parent `NODE`.
    * `--`: Indicates an undirected edge between the parent `NODE` and the `TargetNodeID`.
* `TargetNodeID`: The ID of the node the edge connects to (or originates from, for `<-`). Must be a valid `NodeID`.
* `[Optional Simple Properties]`: An optional block enclosed in square brackets `[]` containing simple key-value properties for the edge (see Section 2.4). Complex edge properties are discouraged; consider representing complex edge data as a separate node if necessary.

**Consistency Requirement:** For every explicitly listed directed edge (e.g., `A -> B [prop]`), the corresponding reverse edge (`B <- A [prop]`) SHOULD also be listed under the target node's definition. For every undirected edge (`A -- B [prop]`), the corresponding edge (`B -- A [prop]`) SHOULD also be listed. A formatting tool (`ndgraphfmt`) should be used to enforce this consistency.

### 2.4 Simple Property Definitions (Inline)

Simple properties can be included directly within square brackets `[]` on the `NODE` or `EDGE` line. This is suitable for labels, weights, statuses, etc.
* Format: `[key1: value1, key2: value2, ...]`
* Enclosed in square brackets `[]`.
* Consists of one or more comma-separated `key: value` pairs.
* `key`: A simple identifier (letters, numbers, underscore, hyphen).
* `value`: Can be:
    * A number (`123`, `4.5`, `-10`).
    * A boolean (`true`, `false`).
    * A quoted string (`"like this"`, `'or this'`). Allows standard escapes.
    * An unquoted simple string (no spaces or special characters like `[]:,"`).

### 2.5 Complex Properties (Attached Data Block)

For nodes requiring complex or extensive properties (e.g., nested data, lists, multi-line text), a standard fenced data block (like JSON or YAML) can be placed immediately following the `NODE` definition line it applies to.
* The block should use standard ``` syntax, optionally specifying the data format (e.g., ```json).
* Tools parsing the `.ndgraph` file should associate the content of this block with the immediately preceding `NODE`.
* This leverages existing block extraction mechanisms [cite: uploaded:neuroscript/pkg/neurodata/blocks/blocks_extractor.go].

*Example:*
```ndgraph
NODE N3 [label: "Node with Data Block"]
```json
{
  "description": "Uses JSON.",
  "config": {"attempts": 3}
}
```
  # Edges for N3 would follow here
  -> N4
NODE N4 ...
```

### 2.6 Comments and Blank Lines

Lines starting with `#` or `--` (after optional whitespace) are comments and are ignored. Blank lines are also ignored.

## 3. EBNF Grammar (Draft)

(* EBNF reflecting attached data block *)
```ebnf
graph_file        ::= { metadata_line | comment_line | blank_line } { node_definition_block } ;

metadata_line     ::= optional_whitespace "::" whitespace key ":" value newline ;
key               ::= identifier ;
value             ::= rest_of_line ;

node_definition_block ::= node_definition_line [ fenced_data_block ] { edge_definition } ;

node_definition_line  ::= optional_whitespace "NODE" whitespace node_id [ whitespace simple_property_block ] newline ;
node_id           ::= identifier ;

edge_definition   ::= indentation edge_marker whitespace node_id [ whitespace simple_property_block ] newline ;
indentation       ::= whitespace+ ;
edge_marker       ::= "->" | "<-" | "--" ;

simple_property_block ::= "[" property_list "]" ;
property_list     ::= property_entry { "," property_entry } ;
property_entry    ::= optional_whitespace key optional_whitespace ":" optional_whitespace property_value optional_whitespace ;
property_value    ::= number_literal | boolean_literal | string_literal | simple_string ;

fenced_data_block ::= optional_whitespace "```" [ language_tag ] newline { text_line } optional_whitespace "```" newline ;
language_tag      ::= identifier ;
text_line         ::= any_character_except_backticks newline ;

identifier        ::= letter { letter | digit | "_" } ;
letter            ::= "a..z" | "A..Z" | "_" ;
digit             ::= "0..9" ;
simple_string     ::= (letter | digit | "_" | "-")+ ;

(* Standard definitions needed: number_literal, boolean_literal, string_literal, whitespace, newline, comment_line, blank_line *)
```
*(Note: This EBNF is a draft.)*

## 4. Rendering

Tools can parse this format and render it into other graphical representations like DOT (Graphviz), JSON Graph Format, GML, etc. The rendering mechanism is separate from this specification.

## 5. Example

```ndgraph
:: type: Graph
:: version: 0.1.1
:: id: example-graph-props
:: directed: true
:: description: Example with simple and complex properties.

NODE StartNode [label: "Start", shape: "circle", initial_value: 0]
  # Edges for StartNode
  -> MidNode [weight: 2.1]
  -- AltNode # Undirected

NODE MidNode [label: "Processing"]
  # Complex properties via attached JSON block
  ```json
  {
    "retries": 3,
    "timeout_ms": 5000,
    "parameters": {"alpha": 0.5, "beta": 0.1}
  }
  ```
  # Edges for MidNode
  <- StartNode [weight: 2.1]
  -> EndNode [label: "Success"]

NODE EndNode [label: "End", shape: "doublecircle"]
  <- MidNode [label: "Success"]

NODE AltNode [status: "alternative"]
  -- StartNode # Undirected
```