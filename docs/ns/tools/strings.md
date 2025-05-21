:: type: NSproject
:: subtype: tool_spec_summary
:: version: 0.1.0
:: id: tool-spec-summary-string-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_string.go, docs/script_spec.md
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Summary: String Tools (v0.1)

This document provides an abbreviated overview of the built-in String tools.

---

### Tool: `String.StringLength`
* **Purpose:** Returns the number of UTF-8 characters (runes) in a string.
* **Syntax:** `CALL String.StringLength(input: <String>)`
* **Arguments:**
    * `input` (String): Required. The string to measure.
* **Returns:** (Number) The rune count.
* **Example:** `CALL String.StringLength("你好")` -> `LAST` will be `2`.

---

### Tool: `String.Substring`
* **Purpose:** Returns a portion of the string based on rune indices.
* **Syntax:** `CALL String.Substring(input: <String>, start: <Number>, end: <Number>)`
* **Arguments:**
    * `input` (String): Required. The source string.
    * `start` (Number): Required. 0-based start index (inclusive).
    * `end` (Number): Required. 0-based end index (exclusive).
* **Returns:** (String) The extracted substring. Returns empty string if indices are invalid/out of bounds.
* **Example:** `CALL String.Substring("abcdef", 1, 4)` -> `LAST` will be `"bcd"`.

---

### Tool: `String.ToUpper`
* **Purpose:** Converts a string to uppercase.
* **Syntax:** `CALL String.ToUpper(input: <String>)`
* **Arguments:**
    * `input` (String): Required. The string to convert.
* **Returns:** (String) The uppercase version of the string.
* **Example:** `CALL String.ToUpper("Hello World")` -> `LAST` will be `"HELLO WORLD"`.

---

### Tool: `String.ToLower`
* **Purpose:** Converts a string to lowercase.
* **Syntax:** `CALL String.ToLower(input: <String>)`
* **Arguments:**
    * `input` (String): Required. The string to convert.
* **Returns:** (String) The lowercase version of the string.
* **Example:** `CALL String.ToLower("Hello World")` -> `LAST` will be `"hello world"`.

---

### Tool: `String.TrimSpace`
* **Purpose:** Removes leading and trailing whitespace from a string.
* **Syntax:** `CALL String.TrimSpace(input: <String>)`
* **Arguments:**
    * `input` (String): Required. The string to trim.
* **Returns:** (String) The string with leading/trailing whitespace removed.
* **Example:** `CALL String.TrimSpace("  some text  ")` -> `LAST` will be `"some text"`.

---

### Tool: `String.SplitString`
* **Purpose:** Splits a string into a list of substrings based on a specified delimiter.
* **Syntax:** `CALL String.SplitString(input: <String>, delimiter: <String>)`
* **Arguments:**
    * `input` (String): Required. The string to split.
    * `delimiter` (String): Required. The string to use as a separator.
* **Returns:** (List of Strings) A list containing the substrings.
* **Example:** `CALL String.SplitString("apple,banana,orange", ",")` -> `LAST` will be `["apple", "banana", "orange"]`.

---

### Tool: `String.SplitWords`
* **Purpose:** Splits a string into a list of words, using whitespace as the delimiter.
* **Syntax:** `CALL String.SplitWords(input: <String>)`
* **Arguments:**
    * `input` (String): Required. The string to split into words.
* **Returns:** (List of Strings) A list containing the words.
* **Example:** `CALL String.SplitWords("Split these words")` -> `LAST` will be `["Split", "these", "words"]`.

---

### Tool: `String.JoinStrings`
* **Purpose:** Joins the elements of a list into a single string, placing a specified separator between elements. Elements are converted to strings if necessary.
* **Syntax:** `CALL String.JoinStrings(input_slice: <List>, separator: <String>)`
* **Arguments:**
    * `input_slice` (List): Required. The list of items to join.
    * `separator` (String): Required. The string to insert between elements.
* **Returns:** (String) The joined string.
* **Example:** `SET my_list = ["one", 2, "three"]` -> `CALL String.JoinStrings(my_list, " - ")` -> `LAST` will be `"one - 2 - three"`.

---

### Tool: `String.ReplaceAll`
* **Purpose:** Replaces all occurrences of a specified substring with another substring.
* **Syntax:** `CALL String.ReplaceAll(input: <String>, old: <String>, new: <String>)`
* **Arguments:**
    * `input` (String): Required. The original string.
    * `old` (String): Required. The substring to be replaced.
    * `new` (String): Required. The substring to replace with.
* **Returns:** (String) The string with all replacements made.
* **Example:** `CALL String.ReplaceAll("this is a test is", "is", "XX")` -> `LAST` will be `"thXX XX a test XX"`.

---

### Tool: `String.Contains`
* **Purpose:** Checks if a string contains a specified substring.
* **Syntax:** `CALL String.Contains(input: <String>, substring: <String>)`
* **Arguments:**
    * `input` (String): Required. The string to search within.
    * `substring` (String): Required. The substring to search for.
* **Returns:** (Boolean) `true` if the substring is found, `false` otherwise.
* **Example:** `CALL String.Contains("Hello world", "world")` -> `LAST` will be `true`.

---

### Tool: `String.HasPrefix`
* **Purpose:** Checks if a string starts with a specified prefix.
* **Syntax:** `CALL String.HasPrefix(input: <String>, prefix: <String>)`
* **Arguments:**
    * `input` (String): Required. The string to check.
    * `prefix` (String): Required. The prefix to check for.
* **Returns:** (Boolean) `true` if the string starts with the prefix, `false` otherwise.
* **Example:** `CALL String.HasPrefix("filename.txt", "file")` -> `LAST` will be `true`.

---

### Tool: `String.HasSuffix`
* **Purpose:** Checks if a string ends with a specified suffix.
* **Syntax:** `CALL String.HasSuffix(input: <String>, suffix: <String>)`
* **Arguments:**
    * `input` (String): Required. The string to check.
    * `suffix` (String): Required. The suffix to check for.
* **Returns:** (Boolean) `true` if the string ends with the suffix, `false` otherwise.
* **Example:** `CALL String.HasSuffix("filename.txt", ".txt")` -> `LAST` will be `true`.

---

### Tool: `String.LineCountString`
* **Purpose:** Counts the number of lines within a given string.
* **Syntax:** `CALL String.LineCountString(content: <String>)`
* **Arguments:**
    * `content` (String): Required. The string content in which to count lines.
* **Returns:** (Number) The number of lines (similar logic to `FS.LineCountFile`: counts `\n`, adds 1 if non-empty and no trailing `\n`).
* **Example:** `SET multi_line = "Line 1\nLine 2\nLine 3"` -> `CALL String.LineCountString(multi_line)` -> `LAST` will be `3`.