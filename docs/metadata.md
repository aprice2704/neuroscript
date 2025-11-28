# Unified Metadata Specification for NeuroScript Projects

> Status: DRAFT
> Version: 2.2.0 (2025-11-25)

Abstract: This document defines a single, mandatory, and enforceable standard for metadata across all source files within the NeuroScript ecosystem. Its primary purpose is to ensure that metadata is unambiguously and reliably processable by computers, tooling, and automated agents. Adherence to these rules is not optional; they are a core part of the file format specifications. This standard applies to all related project files, including NeuroScript (.ns), Go (.go), NDCL (.nd*), Markdown (.md), and others.

## 1. Metadata Line Format 

All metadata is expressed as a key-value pair on a single line. 

-   Structure: `:: key : value`
-   **Prefix**: Each metadata line must begin with the `::` sigil.
-   **Whitespace**: 
    -   Parsers **must** allow a single space after the `::` sigil.
    -   Parsers **must** allow arbitrary horizontal whitespace before and after the colon (`:`) separator to facilitate alignment.
    -   **Canonical Format**: The "normal" format is `:: key: value` (one space after sigil, no space before colon, one space after colon).
-   **Key**:
    -   Must be camelCase. Parsers must treat keys as case-insensitive.
    -   Must match the character set `[a-zA-Z0-9_.-]+`.
-   **Separator**: A single colon (`:`) must follow the key.
-   **Value**: The value is the remainder of the line. Parsers must trim leading and trailing whitespace from the value.
-   Blank lines are not permitted in metadata blocks and will be considered boundaries of them.

#### Valid Formatting Examples

```text
:: key: value           (Canonical)
::key:value             (Compact - Valid)
:: longKeyName   : val  (Aligned - Valid)
```

### 1.1 Key Matching

Matching: for the purposes of **matching** a key (i.e. finding it in a list of them) the *case of the letters, and the characters underscore, dot and dash* (`_.-`) are **ignored**. Both linters and parsers must report collisions. Thus a hypothetical function MetaKeyMatch(a,b) would return:

| a | b | match? |
| snakeMeta | SNAKEMETA | true |
| snake.scope | snake_scope | true |
| snakescope | snake____scope | true |
| SNAKE_SCOPE | SNAKE#SCOPE | error |
| SNAKE1SCOPE | SNAKE0SCOPE | false |

### 1.2 Canonical Go Regex

The following Go-compatible regular expression captures the key and value from a valid metadata line, handling the flexible whitespace rules defined above.

go // Regex to capture the key (group 1) and value (group 2) from a metadata line. var metaRegex = regexp.MustCompile(`^::\s*([a-zA-Z0-9_.-]+)\s*:\s*(.*)$`) 

## 3. Placement

The placement of metadata blocks is mandatory and not a guideline. This ensures that automated tools can locate metadata without ambiguity.

1. **Default**: absolute start of file, or before the block.
2. Exception: **ns functions**: within the function
3. Exception: **Markdown** (.md): File Footer: At the absolute end of the file (EOF). This prevents metadata from cluttering the rendered view of the document on platforms like GitHub, prioritizing human readability for prose-heavy files.

##  4. Identity Keys: schema & serialization

Every file within the ecosystem must contain ::schema: and ::serialization: keys. These two keys form the fundamental identity of a file, telling any tool how to interpret its logical structure and its physical format.

- **:: schema:**: Defines the logical grammar and vocabulary of the file's contents (e.g., neuroscript, ndcl, sdi-go, spec). 
- **:: serialization:**: Defines the physical file format that wraps the content (e.g., go, md, ns, txt). 

For example, a Go source file that contains special SDI directives has a logical schema of sdi-go but a physical serializationof go. A specification document like this one has a schema of spec and a serialization of md. 

##  5.  Standard Vocabulary

To ensure consistency, all metadata **must** use keys from the standard vocabulary defined below. keys should be camelCase.  

###  5.1. File Scope

*Placed in the file header or footer as per placement rules.* | Key           | Purpose & Rules                                                                                     |
| ------------- | --------------------------------------------------------------------------------------------------- |
| schema        | **Mandatory.** Logical format (ndcl, spec, neuroscript, sdi-go).                                    |
| serialization | **Mandatory.** Physical file format (md, txt, go, ns).                                              |
| product       | The broader product suite this file belongs to (e.g., FDM/NS).                                      |
| majorVersion  | The major version integer of the product suite (e.g., 1).                                           |
| langVersion   | The version of the interpreter or grammar required (e.g.,neuroscript@0.4.1).                        |
| fileVersion   | A version for the file itself. **Must** be a monotonic integer.                                     |
| description   | A concise, one-line summary of the file's purpose.                                                  |
| latestChange  | A short description of the most recent specific change to the file.                                 |
| author        | The name of the human or agent responsible for the content.                                         |
| copyright     | Legal copyright notice (e.g., "Copyright (c) 2025 Andrew Price").                                   |
| license       | An SPDX license identifier (e.g.,MIT, Apache-2.0) or the string "Proprietary".                      |
| created       | The creation date of the file in **ISO 8601 format** (e.g.,2025-06-27 or 2025-06-27T15:58:00Z).     |
| modified      | The last modification date of the file in **ISO 8601 format**.                                      |
| tags          | Comma-separated words for indexing and search (e.g.,dataProcessing,io,experimental).                |
| type, subtype | Domain-specific classification for categorization.                                                  |
| dependsOn     | A comma-separated list of upstream source files this file logically depends on.                     |
| howToUpdate   | Brief, essential instructions for future developers on maintaining this file.                       |

###  5.2. Block, Procedure, Section, Chapter scope 

*Placed inside a function, method, or procedure block.* | Key          | Purpose & Rules                                                                       |
| ------------ | ------------------------------------------------------------------------------------- |
| description  | One-line summary of the procedure's function.                                         |
| purpose      | Multi-line explanation of the rationale behind the procedure.                         |
| param:<name> | Description of a specific parameter.                                                  |
| return:<nameIDX> | Description of a specific return value, identified by name or zero-based index. |
| algorithm    | A multi-line, high-level description of the steps the procedure follows.              |
| exampleUsage | A concrete code snippet demonstrating how to call the procedure.                      |
| caveats      | Important limitations, edge cases, or "gotchas" to be aware of.                       |
| requiresTool | A comma-separated list of required external tools (e.g.,tool.compiler,tool.database). |
| requiresLlm  | Boolean (true/false) indicating if the procedure depends on an LLM.                   |
| timeout      | A duration string (e.g.,30s, 5m) specifying the expected execution timeout.           |
| pure         | Boolean (true/false) indicating if the function is pure (no side effects).            |

###  5.3. Inline Scope 

*Placed inline, immediately preceding a specific line or block of code.* | Key             | Purpose & Rules                                                 |
| --------------- | --------------------------------------------------------------- |
| reason          | Explains *why* this specific step or line of code exists.       |
| todo            | Note for a future improvement or feature to be added.           |
| fixme           | Acknowledges a known bug or issue that needs to be fixed.       |
| securityNote    | Highlights a potential security vulnerability or consideration. |
| performanceNote | Comments on the performance implications of the code.           |

###  5.4. NeuroData Block Scope 

*Placed inside a fenced code or data block, right after the opening fence.* | Key     | Purpose & Rules                                                             |
| ------- | --------------------------------------------------------------------------- |
| id      | A unique identifier for this specific block within the file or system.      |
| version | A version string for the content of the block.                              |
| type    | Explicitly declares the content type of the block (e.g.,json, sql, prompt). |
| grammar | Specifies the grammar and version required to parse the block's content.    |

### 5.5. Verification & Trust Scope

*Used to stamp a file as verified by a trusted agent or tool. If present, these keys allow metadata to be treated as First Class Data.*

| Key           | Purpose & Rules                                                                                       |
| ------------- | ----------------------------------------------------------------------------------------------------- |
| metaValidator | The ID and version of the agent/tool that verified this file (e.g., `neuro-lint@1.2.0`, `gpt-4o`).    |
| metaValidated | The timestamp (**ISO 8601**) when the validation occurred.                                           |
| contentHash   | A cryptographic hash (e.g., `sha256:<hex>`) of the file contents.                                    |

*Logic:*
1. **Hashing Rule:** The hash is calculated on the entire byte stream of the file, **excluding** only the line containing the `:: contentHash:` key (and its trailing newline). All other metadata (author, policy, etc.) is **included** in the hash.
2. **Verification:** A tool parsing this file removes the `contentHash` line, hashes the result, and compares it to the claimed hash.
3. **Trust:** If the hash matches, the metadata is considered current and trusted as of `metaValidated`. If it mismatches, the metadata is flagged as `INVALID/STALE`.

### 5.6 Standard-Defining Files

A file which **defines** a standard **must** contain at least the keys in this example (this file, in fact):

:: name: NeuroData MetaData Standard
:: standardID: ndinmeta
:: standardName: NeuroData In-Situ Metadata
:: standardVersion: 2.2.0
:: canonicalFileLocation: github.com/aprice2704/fdm/code/docs/metadata.md
:: fileName: metadata.md
:: schema: spec
:: serialization: md
:: product: FDM/NS
:: majorVersion: 1
:: fileVersion: 6
:: author: Andrew Price
:: copyright: Copyright (c) 2025 Andrew Price
:: latestChange: Relaxed whitespace rules for formatting alignment.
:: modified: 2025-11-25
:: howToUpdate: Update vocab or placement rules first, then bump file_version.
:: dependsOn: none

### SDI

SDI keys may be added to any meta block

`:: sdiSpec: <specID>` in the file header, declares that this file is associated with a named specification
`:: sdiDesign: <designElementID,...>` Documents the parts of the Design this scope is part of
`:: sdiImpl: <implElementID,...>` Documents the parts of the implementation plan this scope is part of

##  5.7. Go Source File Scope with SDI

For Go files with a ::schema: sdi-go, special sdi:prefixed comments are used to declare architectural design and implementation details. 

`// :: sdiSpec: <specID>` in the file header, declares that this file is associated with a named specification
`// :: sdiDesign: <designElementID,...>` Above a type or func documents the parts of the Design this function is part of

##  7. Examples

###  7.1. NeuroScript Example

```neuroscript 
:: schema: neuroscript 
:: serialization: ns 
:: langVersion: neuroscript@0.2.0 
:: fileVersion: 1.1.0 
:: author: Alice Price 
:: created: 2025-04-30 
:: license: MIT 
:: description: Example script demonstrating metadata placement. 
:: tags: example,metadata 
func ProcessData(needs inputData, optional threshold returns processedCount, errorMsg) means 
  :: purpose: Processes input data according to a threshold. This is a multi-line field to explain the rationale in depth. 
  :: param: inputData: The raw data list to process. 
  :: param: threshold: Optional numeric threshold for filtering. 
  :: return: processedCount: Number of items successfully processed. 
  :: return: errorMsg: Any error message encountered, or "" on success. 
  :: algorithm: 
  ::  1. Initialize counters. 
  ::  2. Iterate through inputData. 
  ::  3. Apply threshold filter if provided. 
  ::  4. Increment counter. 
  ::  5. Return count and empty error string. 
  :: caveats: Does not handle non-numeric data gracefully yet. 
  :: requiresLlm: false 
  set count = 0 
  set err = "" 
  # Iterate and process 
  for each item in inputData 
    :: reason: This is the main processing loop for the function. 
    # ... processing logic ... 
    set count = count + 1 
  endfor 
  return count, err 
endfunc 
``` 

###  7.2. <a name='GoSDIExample'></a>Go + SDI Example

```go 
// Package memorystore persists fractal detail memories. 
// 
// :: schema: sdi-go 
// :: serialization: go 
// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 3
// :: langVersion: neuroscript@0.4.1 
// :: description: Core snapshot store with time-travel telemetry. 
// :: latestChange: Added product key.
// :: author: Andrew Price 
// :: copyright: Copyright (c) 2025 Andrew Price
// :: license: MIT 
// :: sdiSpec: memoryStore 
// :: contract: valueWrapping 
// sdi:design The store uses an immutable, content-addressed blob+tree model. 
package memorystore 
import "crypto/sha256" 
// Store is the main object for memory persistence. 
type Store struct { 
  // ... fields 
} 

// sdi:impl memoryStore 
// sdi:design Each write creates a new root commit pointing to a tree of content-addressed chunks. 
func (s *Store) Put(data []byte) ([32]byte, error) { 
    // :: performanceNote: SHA256 was chosen over faster hashes for content integrity. 
    h := sha256.Sum256(data) 
    // ... implementation logic ... 
    return h, nil 
} 
``` 
--- 

##  8. <a name='ToolingCIEnforcement'></a>8. Tooling & CI Enforcement

Adherence to this specification **must** be enforced via automated tooling and Continuous Integration (CI) checks. Linters and pre-commit hooks should be configured to perform the following validations: 

-   **Presence Check:** Fail if any file is missing the mandatory :: schema: or :: serialization: keys. 
-   **Format Check:** Fail if any metadata line does not conform to the :: key: value format and regex. 
-   **Placement Check:** Fail if file-level metadata is not in the correct location (header/footer) for its serialization type. 
-   **Vocabulary Check:** Warn on any metadata keys that are not part of the standard vocabulary.
-   **Date Format Check:** Fail if :: created: or :: modified: values are not valid ISO 8601 dates.
-   **Version Check:** Fail a build if a :: fileVersion: is not greater than the version in the main branch (to prevent regressions). 
-   **SDI Link Check (for Go):** Fail if a file contains // sdi:impl `specID` but no corresponding :: sdiSpec: 
-   **Hash Check:** If `contentHash` is present, the tool must verify the hash of the file content matches; if not, the metadata is flagged as `INVALID/STALE`.