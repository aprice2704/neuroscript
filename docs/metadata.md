# Unified Metadata Specification for NeuroScript Projects

> Status: DRAFT
> Version: 2.0.0 (2025-06-27)

Abstract: This document defines a single, mandatory, and enforceable standard for metadata across all source files within the NeuroScript ecosystem. Its primary purpose is to ensure that metadata is unambiguously and reliably processable by computers, tooling, and automated agents. Adherence to these rules is not optional; they are a core part of the file format specifications. This standard applies to all related project files, including NeuroScript (.ns), Go (.go), NDCL (.nd*), Markdown (.md), and others.

## 1. Metadata Line Format 

All metadata is expressed as a simple key-value pair on a single line, intended for trivial parsing.

-   Structure: ::key: value
-   Prefix: Each metadata line must begin with the :: sigil.
-   Key:
    -   Must be camelCase. Parsers must treat keys as case-insensitive.
    -   Must match the character set [a-zA-Z0-9_.-]+.-   Separator: A single colon (:) must follow the key.
-   Value: The value is the remainder of the line. Parsers must trim leading and trailing whitespace from the value.

### 1.1 Key Matching

Matching: for the purposes of **matching** a key (i.e. finding it in a list of them) the *case of the letters, and the characters underscore, dot and dash* (`_.-`) are **ignored**. Both linters and parsers must report collisions. Thus a hypothetical function MetaKeyMatch(a,b) would return:

| a | b | match? |
| snakeMeta | SNAKEMETA | true |
| snake.scope | snake_scope | true |
| snakescope | snake____scope | true |
| SNAKE_SCOPE | SNAKE#SCOPE | error |
| SNAKE1SCOPE | SNAKE0SCOPE | false |

### 1.2 Canonical Go Regex

The following Go-compatible regular expression captures the key and value from a valid metadata line. Tooling should use this or an equivalent parser.

go // Regex to capture the key (group 1) and value (group 2) from a metadata line. var metaRegex = regexp.MustCompile(`^::\s*([a-zA-Z0-9_.-]+):\s*(.*)$`) 

## 3. Placement

The placement of metadata blocks is mandatory and not a guideline. This ensures that automated tools can locate metadata without ambiguity.

1. **Default**: absolute start of file, or before the block.
2. Exception: **ns functions**: within the function
3. Exception: **Markdown** (.md): File Footer: At the absolute end of the file (EOF). This prevents metadata from cluttering the rendered view of the document on platforms like GitHub, prioritizing human readability for prose-heavy files.

##  4. Identity Keys: schema & serialization

Every file within the ecosystem must contain ::schema: and ::serialization: keys. These two keys form the fundamental identity of a file, telling any tool how to interpret its logical structure and its physical format.

- **::schema:**: Defines the logical grammar and vocabulary of the file's contents (e.g., neuroscript, ndcl, sdi-go, spec). 
- **::serialization:**: Defines the physical file format that wraps the content (e.g., go, md, ns, txt). 

For example, a Go source file that contains special SDI directives has a logical schema of sdi-go but a physical serializationof go. A specification document like this one has a schema of spec and a serialization of md. 

##  5.  Standard Vocabulary

To ensure consistency, all metadata **must** use keys from the standard vocabulary defined below. keys should be camelCase.  

###  5.1. File Scope

*Placed in the file header or footer as per placement rules.* 

| Key           | Purpose & Rules                                                                                     |
| ------------- | --------------------------------------------------------------------------------------------------- |
| schema        | **Mandatory.** Logical format (ndcl, spec, neuroscript, sdi-go).                                    |
| serialization | **Mandatory.** Physical file format (md, txt, go, ns).                                              |
| langVersion   | The version of the interpreter or grammar required (e.g.,neuroscript@0.4.1).                        |
| fileVersion   | A version for the file itself. **Must** be a monotonic integer.|
| description   | A concise, one-line summary of the file's purpose.                                                  |
| author        | The name of the human or agent responsible for the content.                                         |
| license       | An SPDX license identifier (e.g.,MIT, Apache-2.0) or the string "Proprietary".                      |
| created       | The creation date of the file in **ISO 8601 format** (e.g.,2025-06-27 or 2025-06-27T15:58:00Z).       |
| modified      | The last modification date of the file in **ISO 8601 format**.                                      |
| tags          | Comma-separated words for indexing and search (e.g.,dataProcessing,io,experimental).     |
| type, subtype | Domain-specific classification for categorization.                                                  |
| dependsOn     | A comma-separated list of upstream source files this file logically depends on.                     |
| howToUpdate   | Brief, essential instructions for future developers on maintaining this file.                       |

###  5.2. Block, Procedure, Section, Chapter scope 

*Placed inside a function, method, or procedure block.* 

| Key          | Purpose & Rules                                                                       |
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

*Placed inline, immediately preceding a specific line or block of code.* 
| Key             | Purpose & Rules                                                 |
| --------------- | --------------------------------------------------------------- |
| reason          | Explains *why* this specific step or line of code exists.       |
| todo            | Note for a future improvement or feature to be added.           |
| fixme           | Acknowledges a known bug or issue that needs to be fixed.       |
| securityNote    | Highlights a potential security vulnerability or consideration. |
| performanceNote | Comments on the performance implications of the code.           |

###  5.4. NeuroData Block Scope 

*Placed inside a fenced code or data block, right after the opening fence.* 

| Key     | Purpose & Rules                                                             |
| ------- | --------------------------------------------------------------------------- |
| id      | A unique identifier for this specific block within the file or system.      |
| version | A version string for the content of the block.                              |
| type    | Explicitly declares the content type of the block (e.g.,json, sql, prompt). |
| grammar | Specifies the grammar and version required to parse the block's content.    |

### 5.5 Standard-Defining Files

A file which **defines** a standard **must** contain at least the keys in this example (this file, in fact):

:: name: NeuroData MetaData Standard
:: standardID: ndinmeta
:: standardName: NeuroData In-Situ Metadata
:: standardVersion: 2.0.0
:: canonicalFileLocation: github.com/aprice2704/fdm/code/docs/metadata.md
:: fileName: metadata.md
:: schema: spec
:: serialization: md
:: fileVersion: 2
:: author: Andrew Price
:: modified: 2025-06-27
:: howToUpdate: Update vocab or placement rules first, then bump file_version.
:: dependsOn: none

### SDI

SDI keys may be added to any meta block

`:: sdiSpec: <specID>` in the file header, declares that this file is associated with a named specification
`:: sdiDesign <designElementID,...>` Documents the parts of the Design this scope is part of
`:: sdiImpl <implElementID,...>` Documents the parts of the implementation plan this scope is part of

##  5.6. Go Source File Scope with SDI

For Go files with a ::schema: sdi-go, special sdi:prefixed comments are used to declare architectural design and implementation details. 

`// :: sdiSpec: <specID>` in the file header, declares that this file is associated with a named specification
`// :: sdiDesign <designElementID,...>` Above a type or func documents the parts of the Design this function is part of

##  7. Examples

###  7.1. NeuroScript Example

```neuroscript 
::schema: neuroscript 
::serialization: ns 
::langVersion: neuroscript@0.2.0 
::fileVersion: 1.1.0 
::author: Alice Price 
::created: 2025-04-30 
::license: MIT 
::description: Example script demonstrating metadata placement. 
::tags: example,metadata 
func ProcessData(needs inputData, optional threshold returns processedCount, errorMsg) means 
  ::purpose: Processes input data according to a threshold. This is a multi-line field to explain the rationale in depth. 
  ::param:inputData: The raw data list to process. 
  ::param:threshold: Optional numeric threshold for filtering. 
  ::return:processedCount: Number of items successfully processed. 
  ::return:errorMsg: Any error message encountered, or "" on success. 
  ::algorithm: 
  ::  1. Initialize counters. 
  ::  2. Iterate through inputData. 
  ::  3. Apply threshold filter if provided. 
  ::  4. Increment counter. 
  ::  5. Return count and empty error string. 
  ::caveats: Does not handle non-numeric data gracefully yet. 
  ::requiresLlm: false 
  set count = 0 
  set err = "" 
  # Iterate and process 
  for each item in inputData 
    ::reason: This is the main processing loop for the function. 
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
// ::schema: sdi-go 
// ::serialization: go 
// ::fileVersion: 0.3.0 
// ::langVersion: neuroscript@0.4.1 
// ::description: Core snapshot store with time-travel telemetry. 
// ::author: Andrew Price 
// ::license: MIT 
// ::sdi_spec: memoryStore 
// ::contract: valueWrapping 
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
    // ::performanceNote: SHA256 was chosen over faster hashes for content integrity. 
    h := sha256.Sum256(data) 
    // ... implementation logic ... 
    return h, nil 
} 
``` 
--- 

##  8. <a name='ToolingCIEnforcement'></a>7. Tooling & CI Enforcement

Adherence to this specification **must** be enforced via automated tooling and Continuous Integration (CI) checks. Linters and pre-commit hooks should be configured to perform the following validations: 

-   **Presence Check:** Fail if any file is missing the mandatory::schema: or ::serialization:keys. 
-   **Format Check:** Fail if any metadata line does not conform to the::key: valueformat and regex. 
-   **Placement Check:** Fail if file-level metadata is not in the correct location (header/footer) for itsserializationtype. 
-   **Vocabulary Check:** Warn on any metadata keys that are not part of the standard vocabulary. -   **Date Format Check:** Fail if::created:or::modified:values are not valid ISO 8601 dates.
-   **Version Check:** Fail a build if a::fileVersion:is not greater than the version in the main branch (to prevent regressions). 
-   **SDI Link Check (for Go):** Fail if a file contains// sdi:impl `specID` but no corresponding::sdi_spec: 

# NeuroScript Metadata Specification — Extended (Policy, Capabilities, Effects)

## Overview
Metadata lines appear at the top of a script or function, each beginning with `::key: value`.
They are part of the NS file format and are parsed by the interpreter before execution.
All metadata is optional unless otherwise noted; unknown keys are ignored with a warning.

This extension adds explicit **policy**, **capability grants**, **limits**, and **effects** keys.
These allow trusted configuration scripts to declare and constrain what they can do,
and allow the runtime to enforce least privilege in normal scripts.

## Existing Core Keys (unchanged)
::schema: neuroscript
::serialization: ns
::description: Short human-readable description of the script/function.
::pure: true|false  # Declares the script/function as having no external side effects.
::requiresTool: tool.name[, tool.name...]  # Declares tools that must be available to run.

## New Policy Keys
::policyContext: config|normal|test
    # Execution context for this script or function.
    # 'config' → trusted bootstrap/init scripts; can call requires_trust tools if granted.
    # 'normal' → untrusted user/business logic; requires_trust tools blocked.
    # 'test' → controlled deterministic test runs.

::policyAllow: tool.name[, pattern...] 
    # Allowlisted tools for this script; patterns may be used (e.g., tool.agentmodel.*).

::policyDeny: tool.name[, pattern...]
    # Denylisted tools; overrides allow.

## Capability Grants
# Form: grant.<resource>.<verb>: <scope-list>
# Scope list format is resource-specific; comma-separated entries; wildcards allowed.

::grant.env.read: ENV_KEY1, ENV_KEY2
::grant.secrets.read: SECRET_NAME1, SECRET_NAME2
::grant.net.read: host[:port], hostpattern[:port]
::grant.fs.read: /path/glob1, /path/glob2
::grant.fs.write: /path/glob1
::grant.model.admin: modelname[, modelname...]
::grant.model.use: modelname[, modelname...]
::grant.sandbox.admin: profileID[, profileID...]
::grant.proc.exec: procID[, procID...]
::grant.clock.read: true
::grant.rand.read: true|seed:<int>

## Capability Limits
# Form: limit.<resource>.<limit-name>: <value>

::limit.budget.CAD.max: 500     # Max CAD spend per run
::limit.budget.CAD.perCall: 50  # Max CAD spend per tool call
::limit.net.maxBytes: 1048576   # Bytes allowed across all net ops
::limit.net.maxCalls: 100
::limit.fs.maxBytes: 5242880
::limit.fs.maxCalls: 50
::limit.tool.toolname.maxCalls: 5

## Function/Block Requirements (additional to file-level policy)
::requiresCapability: resource:verb:scope[; resource:verb:scope...]
    # Required capabilities to run this function.

::requiresContext: config|normal|test
    # Required execution context.

## Effects
::effects: idempotent[, readsClock][, readsRand][, readsNet][, readsFS]
    # Declares side-effect behaviour; for caching, lints, deterministic test control.

## Example — Trusted Config Script
::schema: neuroscript
::serialization: ns
::description: Init agent models + sandbox
::policyContext: config
::policyAllow: tool.agentmodel.Register, tool.agentmodel.Delete, tool.os.Getenv, tool.sandbox.SetProfile
::grant.env.read: OPENAI_API_KEY
::grant.model.admin: *
::grant.net.read: *.openai.com:443
::limit.budget.CAD.max: 500

must tool.agentmodel.Register("mini", {
  "provider": "openai",
  "model": "gpt-4o-mini",
  "api_key": tool.os.Getenv("OPENAI_API_KEY")
})

## Example — Normal Script
::schema: neuroscript
::serialization: ns
::description: Summarize docs with 'mini'
::policyContext: normal
::grant.model.use: mini
::grant.net.read: *.openai.com:443
::limit.budget.CAD.max: 60

ask "mini", "Summarize: {{txt}}" with {"json": true} into out




---------------

:: name: NeuroData MetaData Standard
:: standardID: ndinmeta
:: standardName: NeuroData In-Situ Metadata
:: standardVersion: 2.1.0
:: canonicalFileLocation: github.com/aprice2704/fdm/code/docs/metadata.md
:: fileName: metadata.md
:: schema: spec
:: serialization: md
:: fileVersion: 3
:: author: Andrew Price
:: modified: 2025-08-10
:: howToUpdate: Update vocab or placement rules first, then bump file_version.
:: dependsOn: none