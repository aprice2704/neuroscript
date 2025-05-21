:: type: NeuroData
:: subtype: spec
:: version: 0.1.0
:: id: ndpatch-json-spec-0.1.0
:: status: draft
:: dependsOn: docs/metadata.md, docs/specification_structure.md
:: howToUpdate: Review format based on usage by patching tools or AI. Update version for non-trivial changes.

# NeuroScript JSON Patch Format Specification (`ndpatch.json`)

## 1. Purpose

This specification defines the `ndpatch.json` format, a JSON-based structure for representing individual file modification operations (replacements, insertions, deletions) using line numbers. It is intended as an alternative to standard diff/patch formats, particularly where robustness against intermediate formatting changes (e.g., in UI transfers) is desired, while still allowing for precise, location-based changes. Each object in the top-level array represents a single operation on a specific file. The recommended file extension is `.ndpatch.json`.

## 2. Example

*This example shows three separate operations: a replace, an insert, and a delete, potentially targeting the same file but represented as individual objects in the array.*

```json
[
  {
    "file": "path/to/your/file.txt",
    "line": 8,
    "op": "replace",
    "old": "This is the original line 8 content.",
    "new": "This is the new content replacing line 8."
  },
  {
    "file": "path/to/your/file.txt",
    "line": 5,
    "op": "insert",
    "new": "This new line will be inserted before the original line 5."
  },
  {
    "file": "path/to/your/file.txt",
    "line": 10,
    "op": "delete",
    "old": "This original line 10 content will be deleted."
  }
]
```

## 3. Design Choices / Rationale (Optional)

* **JSON Structure:** Chosen for its widespread support and robustness in representing string content.
* **Array of Operations:** Each object represents a single change, simplifying generation but requiring the applying tool to handle operations sequentially and manage state (like line number shifts) across operations targeting the same file.
* **Line Numbers:** Included for precise location of changes. Using 1-based indexing is conventional.
* **`old` field for Verification:** Provides an optional safety check for `replace` and `delete`.
* **Alternative to Standard Patch:** Addresses potential corruption issues of standard diff formats during transfer.
* **Tool Incompatibility:** Acknowledged trade-off is incompatibility with standard `patch` and `git apply` tools. Requires custom application logic.

## 4. Syntax / Format Definition

The format consists of a single top-level JSON array `[...]`. Each element within the array is an "Operation Object" representing a single modification to a file.

### 4.1 Operation Object Structure

Each object within the top-level array represents a single modification and has the following keys:

* `file`: (String, Required)
    * The relative path to the target file that needs modification. Path should be relative to a common root (e.g., project root).
* `line`: (Integer, Required)
    * The 1-based line number in the target file where the operation should occur, relative to the state of the file *before* this specific operation is applied within the overall sequence of operations in the array.
    * For `replace`: The line number to be replaced.
    * For `insert`: The line number *before* which the `new` content should be inserted. The inserted content will become the new line at this number.
    * For `delete`: The line number to be deleted.
* `op`: (String, Required)
    * Specifies the type of modification. Must be one of: `"replace"`, `"insert"`, `"delete"`.
* `old`: (String, Optional but Recommended for `replace`/`delete`)
    * The expected original content of the line identified by `line`. Used by applying tools for verification before modifying the file. Should be omitted or null for `insert`. Example name in test data: `"original_line_for_reference"`.
* `new`: (String, Required for `replace`/`insert`)
    * Contains the full text, including any leading/trailing whitespace and line ending characters (typically `\n`), for the line that should replace the existing line (`replace`) or be inserted (`insert`). Should be omitted or null for `delete`. Example name in test data: `"new_line_content"`.

## 5. EBNF Grammar (Optional)

*Not applicable. The structure is defined by JSON syntax.*

## 6. AI Reading

* Understand this format describes a sequence of individual changes to files using line numbers.
* Each object in the top-level array is a self-contained operation specifying the `file`, `line`, `op`, optional `old` content, and required `new` content (for replace/insert).
* Recognize that operations should be applied sequentially as they appear in the array.
* Crucially, understand that line numbers (`line`) refer to the file state *before* the specific operation is applied. When multiple operations target the same file, the applying tool must track line number shifts.

## 7. AI Writing

* When generating patches in this format:
    * Ensure the top-level structure is a JSON array `[...]`.
    * Each element in the array must be an object representing a single operation.
    * Each operation object must have `file` (string), `line` (integer >= 1), and `operation` (string: "replace", "insert", or "delete").
    * For `replace`, include `old` (string, recommended) and `new` (string).
    * For `insert`, include `new` (string) and omit `old`.
    * For `delete`, include `old` (string, recommended) and omit `new`.
    * Verify `old` content matches the target line when providing it.
    * Ensure `new` content is the complete desired line.
    * Calculate line numbers based on the original file state relative to where the change needs to happen, considering the logical sequence of operations already added to the patch array.

## 8. Tooling Requirements / Interaction (Optional)

* **Incompatibility:** Standard tools like `patch` or `git apply` **cannot** parse or apply this format.
* **Parsing:** Requires a standard JSON parser.
* **Application Logic:** Custom tooling (e.g., `nspatch.go` [cite: uploaded:neuroscript_small/pkg/nspatch/nspatch.go]) is required. The tool must:
    * Process the operation objects in the top-level array **strictly sequentially**.
    * For each operation:
        * Identify the target `file`.
        * Read the file content if not already in memory for that file (or maintain the modified state in memory if multiple operations target the same file).
        * Maintain an internal `line_offset` counter *per file*, initialized to 0 when the file is first encountered. This offset tracks how insertions (+) or deletions (-) have shifted subsequent line numbers *relative to the original file*.
        * Calculate the `adjusted_line_number = operation.line + line_offset`. This is the actual index (0-based) in the *current* state of the line list being modified for that specific file.
        * **Verification (Optional but Recommended):** If `old` is provided for `replace` or `delete`, check if the content at `adjusted_line_number` matches `old`. If not, report an error and potentially skip the operation or halt processing.
        * **Perform Operation:**
            * `replace`: Modify the line at `adjusted_line_number` with `new`. (No change to `line_offset`).
            * `insert`: Insert `new` content *before* the line at `adjusted_line_number`. Increment the `line_offset` for that file by 1.
            * `delete`: Remove the line at `adjusted_line_number`. Decrement the `line_offset` for that file by 1.
        * **Error Handling:** Handle cases where `adjusted_line_number` is out of bounds for the current state of the file's line list.
    * After processing all operations, write the final modified content back to all affected files.
* **Whitespace/Line Endings:** Preservation relies on `old`/`new` strings being correctly represented in JSON and handled correctly by the applying tool. Assume `\n` line endings unless specified otherwise.