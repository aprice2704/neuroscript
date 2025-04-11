:: version: 0.5.0
:: dependsOn: pkg/neurodata/checklist/scanner_parser.go, pkg/neurodata/checklist/defined_errors.go, docs/conventions.md
:: howToUpdate: Review scanner_parser.go and defined_errors.go, ensure syntax, error handling (ErrMalformedItem, ErrNoContent), parsing logic (string manip, not regex), status conventions, rollup logic, and examples are accurate.

# NeuroData Checklist Format (.ndcl) Specification

## 1. Purpose

NeuroData Checklists provide a simple, human-readable format for tracking tasks, requirements, or states. They are designed to be easily parsed and manipulated by tools while remaining clear in plain text. They use a syntax based on Markdown task lists, with an extension for items whose status is automatically derived from children.

## 2. Syntax

A checklist file or block primarily consists of checklist item lines, optionally preceded by file-level metadata. Comments and blank lines are also permitted. The parser uses string manipulation (not regular expressions) to identify items.

### 2.1 Checklist Item Line

Each checklist item starts with an optional indentation, followed by a hyphen (`-`), one or more spaces, and then *either* square brackets `[]` for manual items or pipe symbols `||` for automatic items, enclosing a status symbol.

**Manual Item:**
```
Optional Indentation + "- " + "[" + Status Symbol + "]" + Optional Whitespace + Description Text
```

**Automatic Item:**
```
Optional Indentation + "- " + "|" + Status Symbol + "|" + Optional Whitespace + Description Text
```

* **Indentation:** Optional leading whitespace (spaces or tabs) before the `-` defines the item's nesting level (calculated as number of runes). Significant for automatic status rollup (see Section 4.2).
* **Marker:** Must start with a hyphen (`-`) followed by at least one space (`- `).
* **Delimiter:**
    * `[`...`]` (Square Brackets): Indicate a **manual** item whose status is set directly.
    * `|`...`|` (Pipe Symbols): Indicate an **automatic** item whose status should be calculated by tools based on its children (see Section 4.2). The parser identifies the closing pipe relative to the opening one.
* **Status Symbol (Inside Delimiters):** A single character representing the current state.
    * **Parsed Symbols:** The parser recognizes:
        * ` ` (space) or empty (`[]`, `||`) -> Parsed Status: "pending"
        * `x` or `X` -> Parsed Status: "done" (Normalized to 'x')
        * `-` -> Parsed Status: "partial"
        * Any other *single* non-whitespace character (e.g., `?`, `!`, `*`, `🔥`) -> Parsed Status: "special".
    * **Malformed Delimiter Content:** If the content inside `[]` or `||` is not empty and contains more than one character (e.g., `[xx]`, `|?!|`), the parser returns `ErrMalformedItem`.
* **Description Text:** Any text following the closing delimiter (`]` or `|`). Leading/trailing whitespace around the description is trimmed by the parser.

### 2.2 Other Allowed Lines
* **File-Level Metadata:** `:: key: value` at the very beginning. Parsed using the `metadata` package.
* **Markdown Headings:** Lines starting with `#`. Skipped by the checklist parser.
* **Comments:** Lines starting with `#` or `--` (after optional whitespace). Skipped.
* **Blank Lines:** Skipped.

### 2.3 End of Checklist
The checklist parsing stops at the first line encountered that is *not* a valid checklist item (manual or automatic), a heading, a comment, a blank line, or a valid file-level metadata line (if before any items).

## 3. Metadata
### 3.1 File-Level Metadata
Standard `:: key: value` lines at the very beginning of the content, parsed by the `metadata.Extract` function.

### 3.2 Block-Level Metadata (in Composite Files)
Standard `:: key: value` lines immediately preceding a fenced code block (e.g., ```neurodata-checklist). Handled by block extraction tools.

## 4. Status Interpretation and Rollup
(Section remains the same as previous version, describes the intended *semantics*.)
### 4.1 Item Status Categories
* **Pending:** Symbol ` ` (space). Written as `[ ]` (manual) or `| |` (automatic, calculated).
* **Done:** Symbol `x` (normalized from `x` or `X`). Written as `[x]` (manual) or `|x|` (automatic, calculated).
* **Partial:** Symbol `-`. Written as `[-]` (manual) or `|-|` (automatic, calculated).
* **Special:** Symbol is any single character *other than* space, `x`, `X`, or `-`. Examples: `?`, `!`, `*`, `🔥`. Written as `[?]` (manual) or `|?|` (automatic, calculated).
* **Automatic:** Indicated by the use of `| |` delimiters instead of `[ ]`. The symbol *inside* `||` represents the *calculated* status based on children. The parser sets an `IsAutomatic` flag.

### 4.2 Status Rollup Logic (for Automatic `| |` Items)
When a tool updates or reformats a checklist containing automatic items (parsed with `||`), it should determine the status symbol for that item by examining its **direct children**. The rules are applied in order of priority:

1.  **Special Wins:** If *any* direct child has a "Special" status (e.g., `[?]`, `|!|`), the automatic parent item takes on the status symbol of the *first* special child encountered (e.g., `?`, `!`).
2.  **Partial if Any Partial/Done:** Else if *any* direct child has "Partial" (`[-]`, `|-|`) OR "Done" (`[x]`, `|x|`) status, the automatic parent item's status symbol becomes `-` (Partial).
3.  **Done if All Done:** Else if *all* direct children are "Done" (`[x]`, `|x|`), the automatic parent item's status symbol becomes `x` (Done).
4.  **Pending if All Pending:** Else (must mean all direct children are "Pending" `[ ]` or `| |`), the automatic parent item's status symbol remains ` ` (Pending).
5.  **No Children:** If an automatic item has no direct children, its status symbol is ` ` (Pending).

## 5. Parser Output & Errors

The `ParseChecklist` function returns a `ParsedChecklist` struct containing `Metadata map[string]string` and `Items []ChecklistItem`, or an error.

### 5.1 ChecklistItem Struct
* `Text`: string (Trimmed description)
* `Status`: string ("pending", "done", "partial", "special")
* `Symbol`: rune (' ', 'x', '-', '?', '!', etc.)
* `Indent`: int (Number of leading runes before '-')
* `LineNumber`: int (1-based line number in the original input)
* `IsAutomatic`: bool (True if `| |` was used, false if `[ ]` was used)

### 5.2 Defined Errors
* `ErrMalformedItem`: Returned if delimiter content is invalid (e.g., `[xx]`).
* `ErrNoContent`: Returned if the input contains no valid metadata or checklist items after skipping comments/blanks/headings.
* `ErrScannerFailed`: Returned if an underlying error occurs during line scanning (wraps the original scanner error).

## 6. Canonical Formatting
Tools that reformat checklists should aim for this output:

* Manual Pending items: `- [ ] Description`
* Manual Done items: `- [x] Description`
* Manual Partial items: `- [-] Description`
* Manual Special items: `- [?] Description` (using the specific symbol)
* Automatic items: Written using `| |` delimiters with the *calculated* status symbol inside (e.g., `- |-| Description` if calculated as partial, `- |!| Description` if calculated as special '!', `- | | Description` if pending).
* Indentation: Preserved or normalized.

## 7. Examples

### Example 1: Manual and Special Statuses
(Uses `[]`)
```plaintext
- [ ] Pending task
- [x] Completed task
- [-] Partially completed task
- [?] Task needing information (Special)
- [!] Task blocked (Special)
```

### Example 2: Automatic Rollup (Using `| |` Marker)

```plaintext
# Input with Automatic Markers (using | | initially)
- | | Overall Project
  - [x] Phase 1 Done
  - [-] Phase 2 Partial
  - [ ] Phase 3 Pending
- | | Feature A
  - [?] Sub-task A1 (Needs Info - Special)
  - [ ] Sub-task A2
- | | Feature B
  - [x] Sub-task B1
  - [x] Sub-task B2
- | | Feature C (No Children)
- | | Feature D
  - [ ] Step D1
  - [ ] Step D2

# Output After Tool Reformats/Updates
- |-| Overall Project # Partial: Contains Partial/Done children
  - [x] Phase 1 Done
  - [-] Phase 2 Partial
  - [ ] Phase 3 Pending
- |?| Feature A # Special: First child is '?'
  - [?] Sub-task A1 (Needs Info - Special)
  - [ ] Sub-task A2
- |x| Feature B # Done: All children done
  - [x] Sub-task B1
  - [x] Sub-task B2
- | | Feature C (No Children) # Pending: No children
- | | Feature D # Pending: All children pending
  - [ ] Step D1
  - [ ] Step D2
```

### Example 3: File With Metadata and Special Rollup (Using `| |` Marker)

```plaintext
:: version: 0.2.0
:: type: Checklist
:: component: Backend

# API Endpoints
- | | User Management # Marked as automatic
  - [x] GET /users
  - [x] POST /users
  - [-] GET /users/{id} (Partial status)
  - [🔥] PUT /users/{id} (Special status: on fire!)
  - [ ] DELETE /users/{id}
- [ ] Order Processing # Manual item
```
*After reformatting, the `User Management` line would become `|🔥| User Management` because the first "Special" status (`🔥`) encountered among its direct children takes precedence.*