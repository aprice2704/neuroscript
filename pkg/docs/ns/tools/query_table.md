:: type: NSproject
:: subtype: spec
:: version: 0.1.0
:: id: tool-query-table-spec-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, docs/NeuroData/table.md, docs/NeuroData/references.md, docs/script spec.md, docs/metadata.md
:: howToUpdate: Review against implementation in Go. Update version for non-trivial changes to arguments, behavior, or return value.

# Tool Specification: `TOOL.QueryTable` (v0.1)

* **Tool Name:** `TOOL.QueryTable`
* **Purpose:** Reads data from a specified NeuroData Table (`.ndtable`) source, filters rows based on a condition, and returns selected columns. Provides basic SQL `SELECT`/`WHERE` functionality for `.ndtable` files.
* **NeuroScript Syntax:** `CALL TOOL.QueryTable(table, [select], [where])`
* **Arguments:**
    * `table` (String | Reference): Required. Either the string content of an `.ndtable` file or a `[ref:<location>]` string pointing to an `.ndtable` file or block. The location part of a reference will be validated using `SecureFilePath`.
    * `select` (List[String] | Null, Optional): A NeuroScript list of column name strings to include in the results. Column names must match those defined in the source table's schema. If this argument is omitted, `null`, or an empty list, all columns defined in the table's schema are returned.
    * `where` (String | Null, Optional): A NeuroScript expression string that is evaluated for each row of the table. Only rows where this expression evaluates to `true` (according to NeuroScript truthiness rules) are included in the result. If omitted or `null`, all rows are included (no filtering applied). Within the expression string, values from the current row are accessed using the convention `row.<ColumnName>` (e.g., `row.Status`, `row.Cost`).
* **Return Value:** (List[Map])
    * On success, returns a NeuroScript list. Each element in the list is a map representing a single row that matched the `where` clause.
    * Each map contains only the columns specified in the `select` parameter (or all columns if `select` was omitted/null/empty). Map keys are the column names (string), and map values are the corresponding data from the row, converted to appropriate NeuroScript types (string, number, boolean).
    * Returns an empty list `[]` if no rows match the `where` clause or the table is empty.
    * On fatal error (e.g., table parsing failed, `where` clause parsing failed, invalid `select` column), returns `null` (alternative: return a map `{"error": "Error message"}` - TBD during implementation).
* **Behavior:**
    1.  Validates argument count and types. Expects 1 required (`table`) and 2 optional (`select`, `where`).
    2.  Resolves the `table` argument. If it's a `[ref:...]`, it securely resolves the path using `SecureFilePath` and reads the content.
    3.  Parses the `.ndtable` content, extracting the schema (column definitions) and data rows according to the `.ndtable` specification. Returns error if parsing fails.
    4.  Validates the list of column names provided in the optional `select` argument against the parsed schema names. Returns error if any selected column does not exist. If `select` is omitted/null/empty, uses all columns from the schema.
    5.  If a `where` clause string is provided:
        * Parses the `where` expression string into a NeuroScript Abstract Syntax Tree (AST) *once*. Returns error if `where` string is invalid syntax.
    6.  Initializes an empty list `results = []`.
    7.  Iterates through each data row parsed from the table:
        * Converts the row's string cell values into appropriate NeuroScript types based on the schema. Handles conversion errors.
        * If a `where` clause AST exists:
            * Creates a temporary evaluation scope containing the current row's values, accessible via the `row.<ColumnName>` convention (e.g., scope maps `"row"` to a map representing the current row).
            * Evaluates the `where` AST within this scope using the standard NeuroScript expression evaluator.
            * If evaluation results in an error, logs a warning for that row and skips it.
            * If evaluation results in a value that is *not* NeuroScript `true`, skips the current row.
        * (If row was not skipped) Creates a result map containing only the selected columns (with their typed values) for the current row.
        * Appends the result map to the `results` list.
    8.  Returns the final `results` list.
* **Security Considerations:**
    * File system access for `table` references must be sandboxed via `SecureFilePath`.
    * The `where` clause uses the NeuroScript expression evaluator, which is sandboxed by design (cannot perform I/O, `CALL LLM`, etc., directly within the expression). However, overly complex `where` clauses could potentially consume significant resources during evaluation for very large tables. Tool timeouts could be considered for future robustness.
* **Examples:**
    ```neuroscript
    # Assume workers.ndtable exists and is referenced correctly

    # Example 1: Find all available, proficient AI workers, select specific columns
    VAR ai_workers = CALL TOOL.QueryTable(
        table = "[ref:data/workers.ndtable]",
        select = ["WorkerID", "Name", "Endpoint"],
        where = "row.WorkerType == 'AI' AND row.Status == 'available' AND row.SkillLevel == 'proficient'"
    )
    EMIT "Available Proficient AI Workers:"
    EMIT ai_workers # Output will be a list of maps, e.g., [{"WorkerID": "...", "Name": "...", "Endpoint": "..."}]

    # Example 2: Find workers with file management skills, return all columns
    VAR file_mgrs = CALL TOOL.QueryTable(
        table = "[ref:data/workers.ndtable]",
        # 'select' parameter omitted - returns all columns
        where = "CALL TOOL.Contains(row.Skills, 'file_mgmt')" # Assumes TOOL.Contains exists and works
    )
    EMIT "File Managers (all details):"
    EMIT file_mgrs # Output will be a list of maps, each containing all columns for matching rows

    # Example 3: Get all data for a specific worker
    VAR specific_worker_list = CALL TOOL.QueryTable(
        table = "[ref:data/workers.ndtable]",
        where = "row.WorkerID == 'build-server-01'"
    )
    # Result is a list, potentially empty or with one element
    EMIT "Details for build-server-01:"
    EMIT specific_worker_list
    ```
* **Go Implementation Notes:**
    * Suggested Location: `pkg/core/tools_query.go` (new file) or potentially `tools_neurodata.go`.
    * Requires:
        * Ability to parse `.ndtable` schema and data rows (potentially a new function in `pkg/neurodata/table/` or similar).
        * Integration with the NeuroScript expression parser (`pkg/core/parser_api.go`).
        * Integration with the NeuroScript expression evaluator (`pkg/core/evaluation_main.go`), including setting up the `row.` context.
        * Integration with `SecureFilePath` for resolving `table` references.
    * Register the tool function (e.g., `toolQueryTable`) in `pkg/core/tools_register.go`.