# NeuroData Decision Table Format (.nddt) Specification

:: type: DecisionTableFormatSpec
:: version: 0.1.0
:: status: draft
:: dependsOn: docs/metadata.md, docs/references.md, docs/neurodata/table.md, docs/neurodata_and_composite_file_spec.md
:: howToUpdate: Refine condition/action syntax, cell value representations (ranges, wildcards), hit policy definitions, EBNF, examples.

## 1. Purpose

NeuroData Decision Tables (`.nddt`) provide a structured, human-readable format for representing business rules and decision logic. They map combinations of input **Conditions** to determined **Actions** or **Outcomes**. This format aims to make complex conditional logic explicit, inspectable, and manageable, separating it from procedural NeuroScript code.

While potentially inspired by logic programming concepts like Prolog, the `.nddt` format focuses on a tabular representation for readability and tool-based evaluation, rather than direct execution as Prolog clauses.

## 2. Relation to Table Format (`.ndtable`)

The `.nddt` syntax borrows heavily from the `.ndtable` format for its structure (metadata, schema definition, delimited rows). However, its *semantics* are different:
* `.ndtable` stores rows of *data records*.
* `.nddt` stores rows of *rules*, where columns represent logical conditions or resulting actions/outcomes.
* Tooling interaction is distinct: `.ndtable` supports CRUD operations, while `.nddt` is primarily used for evaluation via a dedicated tool (e.g., `TOOL.EvaluateDecisionTable`).

Therefore, `.nddt` uses its own `:: type: DecisionTable` metadata.

## 3. Syntax (`.nddt`)

An `.nddt` file or block consists of the following sections in order:
1.  **File/Block-Level Metadata:** Optional `:: key: value` lines [cite: uploaded:neuroscript/docs/metadata.md].
2.  **Schema Definition:** `CONDITION` and `ACTION`/`OUTCOME` definition lines specifying the table's columns.
3.  **Rules Separator:** A line containing exactly `--- RULES ---`.
4.  **Rule Rows:** Lines representing the decision rules, with cells delimited by pipe (`|`).

Comments (`#` or `--`) and blank lines are allowed before the schema and between schema lines.

### 3.1 File/Block-Level Metadata

Standard `:: key: value` lines. Recommended metadata includes:
* `:: type: DecisionTable` (Required)
* `:: version: <semver>` (Required)
* `:: id: <unique_table_id>` (Required if referenced)
* `:: description: <text>` (Optional)
* `:: hitPolicy: <policy>` (Optional, default: `unique` or `first` - TBD). Defines how multiple matching rules are handled. Common policies (from DMN standard):
    * `unique`: Only one rule can match. Error if multiple match. (Default?)
    * `first`: The first matching rule in document order is selected.
    * `any`: Multiple rules can match, but must all produce the same output. Error if outputs differ.
    * `collect`: All matching rules fire. Outputs are collected (e.g., into a list). Requires defining aggregation for actions if needed (e.g., `collect sum`, `collect list`). (Consider deferring complex collect policies).

### 3.2 Schema Definition Section

Defines the input conditions and output actions/outcomes as columns.
* **Condition Definition:** `CONDITION condition_id data_type [label: "<text>"] [validation_rules...]`
    * `CONDITION`: Keyword.
    * `condition_id`: Unique identifier (`[a-zA-Z_][a-zA-Z0-9_]*`) for this condition column. Used to map input data during evaluation.
    * `data_type`: Expected type for matching input data (`string`, `int`, `float`, `bool`, `timestamp`, `enum(...)`).
    * `[label: "<text>"]`: Optional human-readable description.
    * `[validation_rules...]`: Optional rules (e.g., `REGEX`, `MIN`, `MAX`) applied to *input data* before matching, if needed.
* **Action/Outcome Definition:** `ACTION action_id data_type [label: "<text>"]` or `OUTCOME action_id data_type [label: "<text>"]`
    * `ACTION` / `OUTCOME`: Keywords. (Use one consistently, e.g., `ACTION`).
    * `action_id`: Unique identifier for this output column.
    * `data_type`: Type of the output value defined in the rule rows (`string`, `int`, `float`, `bool`, `timestamp`, `enum(...)`).
    * `[label: "<text>"]`: Optional human-readable description.

### 3.3 Rules Separator

A single line containing exactly `--- RULES ---` MUST follow the schema definition.

### 3.4 Rule Rows Section

* Each line after the separator represents one rule.
* Columns are separated by the pipe character (`|`), corresponding to the `CONDITION` and `ACTION` definitions in order.
* **Condition Cells:** Contain the specific criteria for the rule to match.
    * **Literals:** Exact values (`"Gold"`, `10`, `true`). Must match the `CONDITION` data type.
    * **Wildcard:** A hyphen (`-`) indicates this condition is irrelevant ("don't care") for this rule.
    * **Ranges/Expressions (Optional - Requires Tool Support):** Simple expressions like `">10"`, `"[18-65]"`, `"!=\"Pending\""` could be supported by the evaluation tool. Define syntax clearly if added.
* **Action/Outcome Cells:** Contain the literal output values assigned if the rule matches. Must match the `ACTION` data type.
* **Escaping:** Use `\|` for literal pipe, `\\` for literal backslash within cells, as defined for `.ndtable`.
* **Rule Metadata (Optional):** Comments or `:: ruleId: <id>` can appear after the last cell on a rule line.

## 4. EBNF Grammar (Draft)

```ebnf
decision_table_file ::= { metadata_line | comment_line | blank_line }
                        schema_section
                        rules_separator newline
                        rules_section ;

metadata_line       ::= optional_whitespace "::" whitespace key ":" value newline ;
schema_section        ::= { schema_line | comment_line | blank_line } ;
schema_line         ::= condition_definition | action_definition ;
condition_definition ::= optional_whitespace "CONDITION" whitespace identifier whitespace data_type { whitespace property_block } newline ; (* property_block for label? *)
action_definition   ::= optional_whitespace ("ACTION"|"OUTCOME") whitespace identifier whitespace data_type { whitespace property_block } newline ;

rules_separator     ::= optional_whitespace "--- RULES ---" optional_whitespace ;
rules_section       ::= { rule_row | comment_line | blank_line } ;
rule_row            ::= rule_cell { optional_whitespace "|" optional_whitespace rule_cell } [ optional_whitespace rule_metadata ] newline ;
rule_cell           ::= cell_content ; (* Represents condition entry or action entry *)
cell_content        ::= { character_except_pipe_or_newline | escaped_pipe | escaped_backslash } | "-" ; (* Allow wildcard, ranges need spec *)
rule_metadata       ::= comment_line | metadata_line ; (* Allow rule IDs? *)

(* Define: identifier, key, value, data_type, property_block, validation_rule, whitespace, newline, comment_line, blank_line, etc. *)
```

## 5. Tool Interaction

The primary interaction is via an evaluation tool:
* `TOOL.EvaluateDecisionTable(table_ref_or_content, input_data_map)`
    * `table_ref_or_content` (String or Reference): The `.nddt` definition.
    * `input_data_map` (Map): A map where keys match `CONDITION` IDs and values are the inputs to check (e.g., `{"customer_type": "Gold", "order_total": 1200.50}`).
    * **Behavior:**
        1. Parses the schema and rules.
        2. Iterates through rules, comparing `input_data_map` values against the corresponding `CONDITION` cells (respecting type, wildcards, potentially ranges).
        3. Applies the `:: hitPolicy`.
        4. Collects the values from the `ACTION` cells of the winning rule(s).
    * **Returns:** (Map) A map where keys match `ACTION` IDs and values are the results from the selected rule(s) (or an error string/map if evaluation fails or hit policy violated). For `collect` policies, values might be lists.

## 6. Example

```nddt
:: type: DecisionTable
:: version: 0.1.0
:: id: discount-rules-example
:: hitPolicy: first # First matching rule wins

# Conditions - Input data expected like {"cust_type": "Gold", "total": 1500}
CONDITION cust_type  string  [label: "Customer Type"]
CONDITION total      float   [label: "Order Total"]

# Actions - Output map will be like {"discount": 15, "needs_approval": true}
ACTION discount         int     [label: "Discount %"]
ACTION needs_approval   bool    [label: "Requires Approval"]

--- RULES ---
# Cust Type | Total | Discount | Approval | :: Rule Info
"Gold"     | >1000 | 15       | true     | :: rule_gold_high
"Gold"     | -     | 10       | false    | :: rule_gold_any
"Silver"   | >500  | 7        | false    | :: rule_silver_med
"Silver"   | <=500 | 5        | false    | :: rule_silver_low
-          | >2000 | 5        | true     | :: rule_any_very_high
-          | -     | 0        | false    | :: rule_default_catchall
```