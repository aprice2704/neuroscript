# NeuroData and Composite Files

NeuroData may appear without code fences by itself in a file, or may appear as code-fenced blocks within a file containing many such blocks -- such a file is called a "composite file" and is generally markdown with embedded NeuroData.

## Metadata

All NeuroData and NeuroScript files and blocks should have metadata.

Prefix: Metadata lines must start with :: (colon, colon, followed by at least one space). Optional whitespace is allowed before the ::.

Structure: The line follows a key: value pattern after the :: prefix.
Key: The key comes immediately after the :: prefix (and before the colon). Based on the regex used ([a-zA-Z0-9_.-]+), keys can contain letters, numbers, underscores, periods, and hyphens. 

Whitespace immediately around the key (but after the required space following ::) is tolerated.

Separator: A colon (:) separates the key from the value. Whitespace around the colon is tolerated.

Value: The value consists of everything following the first colon on the line, with leading/trailing whitespace trimmed from the value itself.

Location: All metadata lines must appear at the beginning of the file, before any checklist item lines (lines starting with - [ ] or - [x]). Metadata lines cannot be interspersed with checklist items.

Comments: Regular comment lines (starting with #, potentially indented) are ignored and can appear before, after, or between metadata lines (but still before the first checklist item).
Example:

```markdown

:: id: project-alpha
:: version : 1.2.0   # This is a standard comment, ignored by parser
:: requires-review : true

# Another standard comment

- [ ] First item...
```
