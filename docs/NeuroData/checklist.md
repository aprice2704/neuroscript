:: version: 0.8.1
:: updated: 2025-08-24
:: dependsOn: pkg/neurodata/checklist/scanner_parser.go, pkg/neurodata/checklist/defined_errors.go, docs/conventions.md
:: howToUpdate: Review spec for clarity on auto-generated vs. human-authored IDs, the generation example, and the universal uniqueness validation rule.

# NeuroData Checklist Format (.ndcl) Specification

## 1. Purpose

NeuroData Checklists provide a simple, human-readable format for tracking tasks, requirements, or states. They are designed to be easily parsed and manipulated by tools while remaining clear in plain text. They use a syntax based on Markdown task lists, with extensions for automatic status rollup and stable item referencing.

## 2. Syntax

A checklist file or block primarily consists of checklist item lines, optionally preceded by file-level metadata. Comments and blank lines are also permitted. The parser uses string manipulation (not regular expressions) to identify items.

### 2.1 Checklist Item Line

Each checklist item starts with an optional indentation, followed by a hyphen (`-`), one or more spaces, status delimiters, a description, and an optional reference ID.

```
Indent + "- " + Delimiter + Status + Delimiter + Description + Optional ID
```

* **Indentation, Marker, Delimiter, Status, Description:** See sections below.
* **Optional ID:** An optional, stable reference ID for the item. See Section 2.5 for details. Example: `- [ ] My item #(3k7x)`

### 2.2 Status Symbols

The following table defines the standard status symbols recognized by the parser and the corresponding string representation used for the `status` attribute in the `GenericTree` representation.

| Symbol(s) in File | Status String (`GenericTree`) | Meaning                                        | Type     |
| ------------------- | ------------------------------- | ---------------------------------------------- | -------- |
| ` ` (space), empty  | `"open"`                        | Task is not started or is pending.             | Manual   |
| `x`, `X`            | `"done"`                        | Task is completed.                             | Manual   |
| `-`                 | `"skipped"`                     | Task is skipped or intentionally not done.   | Manual   |
| `>`                 | `"inprogress"`                  | Task is currently being worked on.             | Manual   |
| `!`                 | `"blocked"`                     | Task cannot proceed due to external factors.   | Manual   |
| `?`                 | `"question"`                    | Task requires clarification or information.    | Manual   |
| *any other single* | `"special"`                     | Any other non-standard single char status.     | Manual   |
| ` ` (space), empty  | `"open"`                        | Default state, or all children are open.       | Automatic|
| `x`                 | `"done"`                        | All direct children are done.                  | Automatic|
| `>`                 | `"partial"`                     | At least one child is done or partial/skipped. | Automatic|
| `>`                 | `"inprogress"`                  | Rollup state if relevant child is in progress. | Automatic|
| `!`                 | `"blocked"`                     | Rollup state if relevant child is blocked.     | Automatic|
| `?`                 | `"question"`                    | Rollup state if relevant child is a question.  | Automatic|
| *any other single* | `"special"`                     | Rollup state from a non-standard child status. | Automatic|

**Notes:**
* The parser normalizes `X` to `x`.
* The "Status String" is the value expected in the `"status"` attribute of a `checklist_item` node in the `GenericTree` representation.
* For automatic items (`| |`), the symbol inside the pipes represents the *calculated* status based on children (see Rollup Logic, Section 4.2), but the `GenericTree` status string reflects the meaning (e.g., `|>|` results in `"partial"` status). If a "Special" symbol rolls up, the `GenericTree` status string should match the rolled-up symbol's category (e.g., `|?|` gives `"question"`).
* If the content inside `[]` or `||` is malformed (e.g., more than one character like `[xx]`), the parser returns `ErrMalformedItem`.

### 2.3 Other Allowed Lines
* **File-Level Metadata:** `:: key: value` at the very beginning. Parsed using the `metadata` package.
* **Markdown Headings:** Lines starting with `#`. Skipped by the checklist parser.
* **Comments:** Lines starting with `#` or `--` (after optional whitespace). Skipped.
* **Blank Lines:** Skipped.

### 2.4 End of Checklist
The checklist parsing stops at the first line encountered that is *not* a valid checklist item (manual or automatic), a heading, a comment, a blank line, or a valid file-level metadata line (if before any items).

### 2.5 Item Referencing (Optional)

To provide a stable reference to an item that is immune to reordering or content edits, an optional ID tag may be appended to the description. Once assigned, an ID is considered "sticky" and should only be changed or removed by the author.

* **Syntax**: The ID is enclosed in `#(...)`, for example, `#(c1k4)` or `#(milestone-1)`. It should be the last element on the line after the description.
* **Format**: The ID is a string containing alphanumeric characters (`a-z`, `0-9`) and hyphens (`-`).

#### ID Generation and Authoring

The system supports both machine-generated IDs for convenience and human-authored IDs for clarity.

##### Auto-Generation (Semantic Hash)
When a tool needs to create an ID, it should generate it deterministically from the item's content. The recommended approach is a "semantic hash":
1.  Extract all words and numbers from the item's description.
2.  Identify the **three longest words**.
3.  Create a single string by concatenating the three longest words and any numbers found.
4.  Hash this combined string (e.g., using SHA-1) and encode the result as a **4-digit base36 string**. This is the standard format for auto-generated IDs.

##### Generation Example
This example demonstrates how an ID is generated for a given line item.
1.  **Input Item:** `- [ ] Fix the memory leak in the v2.4 rendering pipeline`
2.  **Extraction:** Filter out common "stop words" (`the`, `in`) to get key terms.
    * **Words:** `Fix`, `memory`, `leak`, `rendering`, `pipeline`
    * **Numbers:** `2`, `4`
3.  **Selection & Concatenation:** Identify the three longest words and combine them with the numbers.
    * **Selected Words:** `rendering`, `pipeline`, `memory`
    * **Semantic String:** `"renderingpipelinememory24"`
4.  **Hashing & Encoding:** The semantic string is hashed and encoded.
    * `hash("renderingpipelinememory24")` → `#(c1k4)` (hypothetical result)
5.  **Final Line:** `- [ ] Fix the memory leak in the v2.4 rendering pipeline #(c1k4)`

##### Human-Authored IDs
An author may choose to write their own meaningful ID for any item. This is especially useful for key milestones, features, or tasks that need to be easily referenced.
* Example: `- |x| Complete project Alpha release #(milestone-1)`

#### Uniqueness Validation

Regardless of whether an ID is auto-generated or human-authored, its most critical property is uniqueness within the document.

* **Tooling Requirement**: Any tool that parses or edits a checklist **must** validate that no two items in the document share the same ID.
* **On Collision**: If a duplicate ID is detected (either during generation or on validation of a human-authored ID), the tool must alert the user, indicating the line numbers of the conflicting items.

## 3. Metadata
### 3.1 File-Level Metadata
Standard `:: key: value` lines at the very beginning of the content, parsed by the `metadata.Extract` function.

### 3.2 Block-Level Metadata (in Composite Files)
Standard `:: key: value` lines immediately preceding a fenced code block (e.g., ```neurodata-checklist). Handled by block extraction tools.

## 4. Status Interpretation and Rollup
(Section remains the same as previous version, describes the intended *semantics*.)
### 4.1 Item Status Categories
* **Open:** Symbol ` ` (space). Written as `[ ]` (manual) or `| |` (automatic, calculated). Corresponds to `"open"` status string.
* **Done:** Symbol `x` (normalized from `x` or `X`). Written as `[x]` (manual) or `|x|` (automatic, calculated). Corresponds to `"done"` status string.
* **Skipped:** Symbol `-`. Written as `[-]` (manual). Corresponds to `"skipped"` status string.
* **Partial:** Symbol `>`. Written as `|>|` (automatic, calculated). Corresponds to `"partial"` status string.
* **InProgress:** Symbol `>`. Written as `[>]` (manual) or `|>|` (automatic, calculated). Corresponds to `"inprogress"` status string.
* **Blocked:** Symbol `!`. Written as `[! ]` (manual) or `|!|` (automatic, calculated). Corresponds to `"blocked"` status string.
* **Question:** Symbol `?`. Written as `[?]` (manual) or `|?|` (automatic, calculated). Corresponds to `"question"` status string.
* **Special:** Symbol is any *other* single character. Corresponds to `"special"` status string.
* **Automatic:** Indicated by the use of `| |` delimiters instead of `[ ]`. The symbol *inside* `||` represents the *calculated* status based on children. The parser sets an `IsAutomatic` flag.

### 4.2 Status Rollup Logic (for Automatic `| |` Items)
When a tool updates or reformats a checklist containing automatic items (parsed with `||`), it should determine the status symbol for that item by examining its **direct children**. The rules are applied in order of priority:

1.  **Special Wins:** If *any* direct child has a status corresponding to `!`, `?`, `>`, or `"special"`, the automatic parent item takes on the status symbol of the *first* such special child encountered (e.g., `!`, `?`, `>`).
2.  **Partial if Any Partial/Done/Skipped:** Else if *any* direct child has status corresponding to `"skipped"`, `"partial"`, OR `"done"`, the automatic parent item's status symbol becomes `>` (Partial).
3.  **Done if All Done:** Else if *all* direct children are `"done"`, the automatic parent item's status symbol becomes `x` (Done).
4.  **Open if All Open:** Else (must mean all direct children are `"open"`), the automatic parent item's status symbol remains ` ` (Open).
5.  **No Children:** If an automatic item has no direct children, its status symbol is ` ` (Open).

## 5. Parser Output & Errors

The `ParseChecklist` function returns a `ParsedChecklist` struct containing `Metadata map[string]string` and `Items []ChecklistItem`, or an error.

### 5.1 ChecklistItem Struct
* `Text`: string (Trimmed description, with ID tag removed)
* `Status`: string (Standardized: `"open"`, `"done"`, `"skipped"`, `"partial"`, `"inprogress"`, `"blocked"`, `"question"`, `"special"`)
* `Symbol`: rune (The *original* single char symbol found: ' ', 'x', '-', '>', '!', '?', etc.)
* `Indent`: int (Number of leading runes before '-')
* `LineNumber`: int (1-based line number in the original input)
* `IsAutomatic`: bool (True if `| |` was used, false if `[ ]` was used)
* `ID`: string (The ID string, if present, e.g., "c1k4" or "milestone-1")

### 5.2 Defined Errors
* `ErrMalformedItem`: Returned if delimiter content is invalid (e.g., `[xx]`).
* `ErrNoContent`: Returned if the input contains no valid metadata or checklist items after skipping comments/blanks/headings.
* `ErrScannerFailed`: Returned if an underlying error occurs during line scanning (wraps the original scanner error).
* `ErrDuplicateID`: Returned by tools if a duplicate ID is detected within a single document.

## 6. Canonical Formatting
Tools that reformat checklists should aim for this output:

* Manual Open items: `- [ ] Description`
* Manual Done items: `- [x] Description`
* ...and so on for other manual statuses.
* Automatic items: Written using `| |` delimiters with the *calculated* status symbol inside.
* Indentation: Preserved or normalized (typically 2 spaces per level is recommended).
* ID Tags: Must be preserved at the end of the line if they exist.

## 7. Examples

### Example 1: Manual Statuses
(Uses `[]`)
```plaintext
- [ ] Open task #(2x4a1)
- [x] Completed task #(4g7h)
- [-] Skipped task #(b31z0)
- [>] In-progress task #(p2r8)
- [!] Blocked task #(b9k1)
- [?] Question task #(q5s2)
- [*] Special status task #(s3t4)
```

### Example 2: Automatic Rollup (Using `| |` Marker)

```plaintext
# Output After Tool Reformats/Updates
- |>| Overall Project #(proj-alpha)
  - [x] Phase 1 Done #(p1d9)
  - [-] Phase 2 Skipped (Manual) #(p2s5)
  - [ ] Phase 3 Open #(p3o2)
- |?| Feature A #(feat-a)
  - [?] Sub-task A1 (Needs Info - Question) #(a1q4)
  - [ ] Sub-task A2 #(a2o9)
- |x| Feature B #(feat-b)
  - [x] Sub-task B1 #(b1d3)
  - [x] Sub-task B2 #(b2d6)
- | | Feature C (No Children) #(feat-c)
```

### Example 3: File With Metadata and Special Rollup

```plaintext
:: version: 0.2.0
:: component: Backend

# API Endpoints
- | | User Management #(api-users)
  - [x] GET /users #(g1u7)
  - [x] POST /users #(p8u2)
  - [-] GET /users/{id} (Skipped status) #(g6u4)
  - [🔥] PUT /users/{id} (Special status: on fire!) #(p9u5)
  - [ ] DELETE /users/{id} #(d3u8)
- [ ] Order Processing #(o7p1)
```

## 8. Implementation Notes

-- TODO: Implement tooling to automatically generate and validate the semantic hash IDs for item referencing as defined in section 2.5. The current focus is on parsing existing checklist formats; generation can be added later.