# NeuroData Tree Format (.ndtree) Specification

:: type: TreeFormatSpec
:: version: 0.1.0
:: status: draft
:: grammar: graph
:: grammarVer: 0.1.0
:: dependsOn: docs/neurodata/graph.md, docs/metadata.md, docs/neurodata_and_composite_file_spec.md
:: howToUpdate: Review tree syntax, property attachment, EBNF, ensure consistency with graph spec.

## 1. Purpose

NeuroData Trees (`.ndtree`) provide a simple, human-readable plain-text format specifically for representing hierarchical tree structures. The format emphasizes readability through indentation while remaining machine-parseable. It is designed primarily to be read by humans and updated by tools or AI.

## 2. Relation to Graph Format

The `.ndtree` format is considered a specialized profile of the `.ndgraph` format (see `:: grammar: graph` metadata). It leverages the same node definition (`NODE NodeID [Props]`) and property syntax but uses indentation to represent the primary parent-child relationships instead of explicit edge markers (`->`, `<-`). Tools parsing `.ndtree` should infer directed edges from parent to child based on indentation. Cycle detection should be performed by validation tools to ensure tree structure integrity.

## 3. Syntax

A `.ndtree` file consists of:
1.  Optional file-level metadata lines (using `:: key: value` syntax [cite: uploaded:neuroscript/docs/metadata.md]).
2.  Optional comments (`#` or `--`) and blank lines.
3.  A series of `NODE` definitions, where hierarchical relationships are defined by indentation.

### 3.1 File-Level Metadata

Standard `:: key: value` lines [cite: uploaded:neuroscript/docs/metadata.md]. Recommended metadata includes:
* `:: type: Tree` (Required)
* `:: version: <semver>` (Required)
* `:: grammar: graph` (Required)
* `:: grammarVer: <semver>` (Required, refers to the graph spec version)
* `:: id: <unique_tree_id>` (Optional but recommended)
* `:: root: <NodeID>` (Optional, explicitly defines the root node)
* `:: description: <text>` (Optional)

### 3.2 Node Definition Line

* Format: `Indentation NODE NodeID [Optional Properties]`
* `Indentation`: Zero or more spaces or tabs. The level of indentation defines the node's parent in the tree (the nearest preceding node line with less indentation). Consistent indentation (e.g., 2 spaces, 4 spaces) is recommended for readability.
* `NODE`: Keyword indicating a node definition.
* `NodeID`: A unique identifier for the node within the tree. Must start with a letter or underscore, followed by letters, numbers, or underscores (`[a-zA-Z_][a-zA-Z0-9_]*`). IDs are case-sensitive.
* `[Optional Properties]`: An optional block enclosed in square brackets `[]` containing simple key-value properties for the node (see Section 3.4).

### 3.3 Complex Properties (Attached Data Block)

For nodes requiring complex or extensive properties, a standard fenced data block (e.g., JSON, YAML) can be placed immediately following the `NODE` definition line it applies to. The parser/tool should associate this data block with the preceding node.

*Example:*
```ndtree
NODE ConfigNode [label: "Configuration"]
```json
{
  "timeout": 30,
  "retry_policy": {
    "attempts": 3,
    "delay": "5s"
  },
  "enabled_features": ["featureA", "featureC"]
}
```
  NODE ChildNode ... # Next node definition starts here
```

### 3.4 Simple Property Definitions (Inline)

Simple properties can be included directly within square brackets `[]` on the `NODE` line.
* Format: `[key1: value1, key2: value2, ...]`
* Enclosed in square brackets `[]`.
* Consists of one or more comma-separated `key: value` pairs.
* `key`: A simple identifier (letters, numbers, underscore, hyphen).
* `value`: Can be:
    * A number (`123`, `4.5`, `-10`).
    * A boolean (`true`, `false`).
    * A quoted string (`"like this"`, `'or this'`). Allows standard escapes.
    * An unquoted simple string (no spaces or special characters like `[]:,"`).
* (Refer to `graph.md` specification for more detailed value parsing rules if needed).

### 3.5 Comments and Blank Lines

Lines starting with `#` or `--` (after optional whitespace) are comments and are ignored. Blank lines are also ignored.

## 4. EBNF Grammar (Draft)

(* EBNF-like notation for NeuroData Tree (.ndtree) - Focus on indentation *)
```ebnf
tree_file         ::= { metadata_line | comment_line | blank_line } node_list ;

metadata_line     ::= optional_whitespace "::" whitespace key ":" value newline ; (* As per graph spec *)
key               ::= identifier ;
value             ::= rest_of_line ;

node_list         ::= { node_definition } ;

node_definition   ::= indentation "NODE" whitespace node_id [ whitespace property_block ] newline [ fenced_data_block ] ;
node_id           ::= identifier ;
indentation       ::= { " " | "\t" } ; (* Parsed to determine level *)

property_block    ::= "[" property_list "]" ; (* As per graph spec *)
property_list     ::= property_entry { "," property_entry } ;
property_entry    ::= optional_whitespace key optional_whitespace ":" optional_whitespace property_value optional_whitespace ;
property_value    ::= number_literal | boolean_literal | string_literal | simple_string ;

fenced_data_block ::= optional_whitespace "```" [ language_tag ] newline { text_line } optional_whitespace "```" newline ;
language_tag      ::= identifier ; (* e.g., json, yaml *)
text_line         ::= any_character_except_backticks newline ;

(* Standard definitions needed: identifier, number_literal, boolean_literal, string_literal, simple_string, whitespace, newline, comment_line, blank_line *)
```
*(Note: This EBNF emphasizes the structure. A full parser would need logic to track indentation levels to build the tree hierarchy.)*

## 5. Rendering

Similar to graphs, `.ndtree` files can be rendered into various visual formats. Tools could generate:
* **Text-based tree diagrams:** Using characters like `├─`, `└─`, `│`.
* **DOT Language:** For Graphviz visualization, translating the inferred parent-child edges.
* **Other formats:** JSON, XML, etc.

## 6. Example

```ndtree
:: type: Tree
:: version: 0.1.0
:: grammar: graph
:: grammarVer: 0.1.0
:: id: file-system-example
:: root: Root

NODE Root [label: "/", type: "dir"]
  NODE Documents [label: "Documents", type: "dir"]
    NODE Resume.docx [label: "Resume.docx", size: 150kb]
    NODE Report.pdf [label: "Report.pdf", size: 2mb]
      # Example of attaching complex data to Report.pdf
      ```json
      {
        "author": "A. Price",
        "keywords": ["report", "analysis", "neurodata"],
        "revision_history": [
          {"version": "1.0", "date": "2024-01-10"},
          {"version": "1.1", "date": "2024-02-15"}
        ]
      }
      ```
  NODE Downloads [label: "Downloads", type: "dir"]
    NODE Image.jpg [label: "Image.jpg"]
  NODE Config.sys [label: "Config.sys", hidden: true]
```