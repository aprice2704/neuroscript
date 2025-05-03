 :: version: 0.6.0
 :: updated: 2025-05-02
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
 * **Status Symbol (Inside Delimiters):** A single character representing the current state. The parser handles normalization (e.g., `X` to `x`). See Section 2.2 for standard symbols.
 * **Description Text:** Any text following the closing delimiter (`]` or `|`). Leading/trailing whitespace around the description is trimmed by the parser.
 
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
 | `-`                 | `"partial"`                     | At least one child is done or partial/skipped. | Automatic|
 | `>`                 | `"inprogress"`                  | Rollup state if relevant child is in progress. | Automatic|
 | `!`                 | `"blocked"`                     | Rollup state if relevant child is blocked.     | Automatic|
 | `?`                 | `"question"`                    | Rollup state if relevant child is a question.  | Automatic|
 | *any other single* | `"special"`                     | Rollup state from a non-standard child status. | Automatic|
 
 **Notes:**
 * The parser normalizes `X` to `x`.
 * The "Status String" is the value expected in the `"status"` attribute of a `checklist_item` node in the `GenericTree` representation.
 * For automatic items (`| |`), the symbol inside the pipes represents the *calculated* status based on children (see Rollup Logic, Section 4.2), but the `GenericTree` status string reflects the meaning (e.g., `|-|` results in `"partial"` status). If a "Special" symbol rolls up, the `GenericTree` status string should match the rolled-up symbol's category (e.g., `|?|` gives `"question"`).
 * If the content inside `[]` or `||` is malformed (e.g., more than one character like `[xx]`), the parser returns `ErrMalformedItem`.
 
 ### 2.3 Other Allowed Lines
 * **File-Level Metadata:** `:: key: value` at the very beginning. Parsed using the `metadata` package.
 * **Markdown Headings:** Lines starting with `#`. Skipped by the checklist parser.
 * **Comments:** Lines starting with `#` or `--` (after optional whitespace). Skipped.
 * **Blank Lines:** Skipped.
 
 ### 2.4 End of Checklist
 The checklist parsing stops at the first line encountered that is *not* a valid checklist item (manual or automatic), a heading, a comment, a blank line, or a valid file-level metadata line (if before any items).
 
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
 * **Partial:** Symbol `-`. Written as `|-|` (automatic, calculated). Corresponds to `"partial"` status string. *Note: The same symbol `-` has different meanings based on context.*
 * **InProgress:** Symbol `>`. Written as `[>]` (manual) or `|>|` (automatic, calculated). Corresponds to `"inprogress"` status string.
 * **Blocked:** Symbol `!`. Written as `[! ]` (manual) or `|!|` (automatic, calculated). Corresponds to `"blocked"` status string.
 * **Question:** Symbol `?`. Written as `[?]` (manual) or `|?|` (automatic, calculated). Corresponds to `"question"` status string.
 * **Special:** Symbol is any *other* single character. Corresponds to `"special"` status string.
 * **Automatic:** Indicated by the use of `| |` delimiters instead of `[ ]`. The symbol *inside* `||` represents the *calculated* status based on children. The parser sets an `IsAutomatic` flag.
 
 ### 4.2 Status Rollup Logic (for Automatic `| |` Items)
 When a tool updates or reformats a checklist containing automatic items (parsed with `||`), it should determine the status symbol for that item by examining its **direct children**. The rules are applied in order of priority:
 
 1.  **Special Wins:** If *any* direct child has a status corresponding to `!`, `?`, `>`, or `"special"`, the automatic parent item takes on the status symbol of the *first* such special child encountered (e.g., `!`, `?`, `>`).
 2.  **Partial if Any Partial/Done/Skipped:** Else if *any* direct child has status corresponding to `"skipped"`, `"partial"`, OR `"done"`, the automatic parent item's status symbol becomes `-` (Partial).
 3.  **Done if All Done:** Else if *all* direct children are `"done"`, the automatic parent item's status symbol becomes `x` (Done).
 4.  **Open if All Open:** Else (must mean all direct children are `"open"`), the automatic parent item's status symbol remains ` ` (Open).
 5.  **No Children:** If an automatic item has no direct children, its status symbol is ` ` (Open).
 
 ## 5. Parser Output & Errors
 
 The `ParseChecklist` function returns a `ParsedChecklist` struct containing `Metadata map[string]string` and `Items []ChecklistItem`, or an error.
 
 ### 5.1 ChecklistItem Struct
 * `Text`: string (Trimmed description)
 * `Status`: string (Standardized: `"open"`, `"done"`, `"skipped"`, `"partial"`, `"inprogress"`, `"blocked"`, `"question"`, `"special"`)
 * `Symbol`: rune (The *original* single char symbol found: ' ', 'x', '-', '>', '!', '?', etc.)
 * `Indent`: int (Number of leading runes before '-')
 * `LineNumber`: int (1-based line number in the original input)
 * `IsAutomatic`: bool (True if `| |` was used, false if `[ ]` was used)
 
 ### 5.2 Defined Errors
 * `ErrMalformedItem`: Returned if delimiter content is invalid (e.g., `[xx]`).
 * `ErrNoContent`: Returned if the input contains no valid metadata or checklist items after skipping comments/blanks/headings.
 * `ErrScannerFailed`: Returned if an underlying error occurs during line scanning (wraps the original scanner error).
 
 ## 6. Canonical Formatting
 Tools that reformat checklists should aim for this output:
 
 * Manual Open items: `- [ ] Description`
 * Manual Done items: `- [x] Description`
 * Manual Skipped items: `- [-] Description`
 * Manual InProgress items: `- [>] Description`
 * Manual Blocked items: `- [!] Description`
 * Manual Question items: `- [?] Description`
 * Manual Special items: `- [*] Description` (using the specific symbol)
 * Automatic items: Written using `| |` delimiters with the *calculated* status symbol inside (e.g., `- |-| Description` if calculated as partial, `- |!| Description` if calculated as blocked, `- | | Description` if open).
 * Indentation: Preserved or normalized (typically 2 spaces per level is recommended).
 
 ## 7. Examples
 
 ### Example 1: Manual Statuses
 (Uses `[]`)
 ```plaintext
 - [ ] Open task
 - [x] Completed task
 - [-] Skipped task
 - [>] In-progress task
 - [!] Blocked task
 - [?] Question task
 - [*] Special status task
 ```
 
 ### Example 2: Automatic Rollup (Using `| |` Marker)
 
 ```plaintext
 # Input with Automatic Markers (using | | initially)
 - | | Overall Project
   - [x] Phase 1 Done
   - [-] Phase 2 Skipped (Manual)
   - [ ] Phase 3 Open
 - | | Feature A
   - [?] Sub-task A1 (Needs Info - Question)
   - [ ] Sub-task A2
 - | | Feature B
   - [x] Sub-task B1
   - [x] Sub-task B2
 - | | Feature C (No Children)
 - | | Feature D
   - [ ] Step D1
   - [ ] Step D2
 
 # Output After Tool Reformats/Updates
 - |-| Overall Project # Partial: Contains Skipped/Done children
   - [x] Phase 1 Done
   - [-] Phase 2 Skipped (Manual)
   - [ ] Phase 3 Open
 - |?| Feature A # Question: First child is '?'
   - [?] Sub-task A1 (Needs Info - Question)
   - [ ] Sub-task A2
 - |x| Feature B # Done: All children done
   - [x] Sub-task B1
   - [x] Sub-task B2
 - | | Feature C (No Children) # Open: No children
 - | | Feature D # Open: All children open
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
   - [-] GET /users/{id} (Skipped status)
   - [🔥] PUT /users/{id} (Special status: on fire!)
   - [ ] DELETE /users/{id}
 - [ ] Order Processing # Manual item
 ```
 *After reformatting, the `User Management` line would become `|🔥| User Management` because the first "Special" status (`🔥`) encountered among its direct children takes precedence.*