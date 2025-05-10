 # NeuroScript Gopls Integration Design (gopls_integration_design.md)

 ## 1. Overview

 This document outlines the design for integrating `gopls`, the official Go Language Server, into the NeuroScript ecosystem. The primary objective is to provide NeuroScript (and thereby AI agents using NeuroScript) with robust, real-time diagnostics and rich semantic understanding of Go code. This integration is pivotal for enabling an AI to write, analyze, and modify Go code that compiles reliably and efficiently, moving beyond brittle line/column addressing towards more stable semantic references.

 The focus for the initial v0.5.0 implementation will be on:
 - Accurate diagnostic reporting (`textDocument/publishDiagnostics`).
 - Workspace and document synchronization with `gopls` (`didOpen`, `didChange`, `didSave`, `didClose`).
 - Tools to retrieve semantic context for diagnostics and code elements (e.g., `textDocument/hover`, `textDocument/definition`, `textDocument/documentSymbol`).
 - Establishing a stable `GoplsClient` for managing communication.

 This system will complement existing static analysis tools in `pkg/core/tools/gosemantic` and `pkg/core/tools/goast` by providing a live, incremental view of the codebase as understood by the Go compiler and associated tooling.

 ## 2. Core Components

 ### 2.1. `GoplsClient` (`pkg/core/tools/gopls_client.go` - new)
 **Role**: Manages the lifecycle of a `gopls` server subprocess and handles all Language Server Protocol (LSP) JSON-RPC 2.0 communication.
 **Responsibilities**:
 -   **Process Management**: Starting, monitoring, and stopping the `gopls` server executable.
 -   **LSP Communication**:
     -   Establishing and managing the JSON-RPC 2.0 stream over stdio with the `gopls` subprocess.
     -   Sending LSP requests (e.g., `initialize`, `textDocument/hover`) and notifications (e.g., `initialized`, `textDocument/didOpen`).
     -   Receiving and dispatching LSP responses and notifications (e.g., `textDocument/publishDiagnostics`).
 -   **State Management**: Tracking client/server capabilities, open documents, and pending requests.
 -   **Interface**: Provides methods for higher-level services (like DiagnosticManager, SemanticInfoProvider) to interact with `gopls`.
 **Implementation Notes**:
 -   Consider using `golang.org/x/tools/gopls/internal/protocol` for LSP type definitions and potentially `golang.org/x/tools/internal/jsonrpc2` for the RPC mechanism if its license and stability are suitable. Otherwise, a minimal custom implementation for JSON-RPC message framing and dispatch will be needed.

 ### 2.2. Workspace/Document Synchronizer (Part of `GoplsClient`)
 **Role**: Keeps `gopls` informed about the state of the workspace and relevant Go files.
 **Responsibilities**:
 -   Handling `textDocument/didOpen` notifications when a file is first accessed or made relevant.
 -   Handling `textDocument/didChange` notifications when a file's content is modified by NeuroScript or an AI. Initially, full content sync is acceptable; incremental sync can be a future optimization.
 -   Handling `textDocument/didSave` notifications if specific save actions need to be signaled.
 -   Handling `textDocument/didClose` notifications when a file is no longer actively managed.
 -   Managing the concept of a "workspace root" (`rootUri` in LSP `initialize` parameters).

 ### 2.3. Diagnostic Manager (Conceptually part of `GoplsClient` or a close collaborator)
 **Role**: Receives, stores, and provides access to diagnostics published by `gopls`.
 **Responsibilities**:
 -   Processing incoming `textDocument/publishDiagnostics` notifications from `gopls`.
 -   Storing diagnostics per file URI, replacing previous diagnostics for that file.
 -   Making these stored diagnostics queryable by NeuroScript tools.

 ### 2.4. Semantic Information Provider (Methods on `GoplsClient`)
 **Role**: Fetches and provides semantic information about code elements.
 **Responsibilities**:
 -   Sending `textDocument/hover` requests and parsing responses.
 -   Sending `textDocument/definition` requests and parsing responses.
 -   Sending `textDocument/documentSymbol` requests and parsing responses.
 -   Sending `workspace/symbol` requests and parsing responses.
 -   Translating LSP responses into NeuroScript's `SemanticReference` and other relevant data structures.

 ## 3. Data Structures (for NeuroScript internal representation & tool outputs)

 These will reside likely in a new `pkg/core/tools/gopls_types.go` or similar.

 -   **`LSPPosition`**:
     -   `Line` (int): 0-indexed line number.
     -   `Character` (int): 0-indexed UTF-16 code unit offset (gopls typically uses UTF-8 byte offsets, which needs careful handling or negotiation; for Go, byte offsets are often more natural). **Decision: Prioritize byte offsets for internal consistency if possible, map to/from LSP's character definition as needed.**
 -   **`LSPRange`**:
     -   `Start` (`LSPPosition`)
     -   `End` (`LSPPosition`)
 -   **`LSPLocation`**:
     -   `URI` (string): Document URI.
     -   `Range` (`LSPRange`)
 -   **`DiagnosticInfo`**:
     -   `SourceFileURI` (string)
     -   `Range` (`LSPRange`): Precise byte-offset range of the diagnostic.
     -   `Severity` (int): e.g., Error, Warning, Info, Hint (maps to LSP `DiagnosticSeverity`).
     -   `Code` (string | int, optional): Diagnostic code, if any.
     -   `Source` (string, optional): Source of the diagnostic (e.g., "compiler", "vet", "staticcheck").
     -   `Message` (string): The diagnostic message.
     -   `SemanticReferenceGuess` (`SemanticReference`, optional): A best-effort attempt to identify the primary symbol related to the diagnostic.
 -   **`SemanticReference` (AI-Focused Identifier)**:
     -   `ResourceURI` (string): The URI of the file containing the element.
     -   `FullyQualifiedName` (string, optional): The fully qualified Go symbol name (e.g., `github.com/aprice2704/neuroscript/pkg/core.Interpreter.ExecuteTool`). Key for stable referencing.
     -   `SymbolKind` (string, optional): Type of symbol (e.g., "function", "method", "variable", "type", "field", "interface", maps from LSP `SymbolKind` or `go/ast` node types).
     -   `DeclarationLocation` (`LSPLocation`, optional): Location of the symbol's declaration.
     -   `ByteOffsetRange` (`LSPRange`, optional): Precise byte-offset range of the symbol's identifier or relevant span in the `ResourceURI`.
     -   `Signature` (string, optional): For functions/methods.
     -   `PackagePath` (string, optional): Go package import path.
 -   **`HoverInfo`**:
     -   `Contents` (string): Markdown content from `gopls` hover.
     -   `Range` (`LSPRange`, optional): The range of the symbol hovered over.

 ## 4. Workflow for AI-Driven Code Editing & Diagnostics

 1.  **Initialization**:
     a.  NeuroScript (via `ng` or `Interpreter`) ensures `GoplsClient` is running and initialized for the target Go workspace/module.
     b.  Relevant project files are signaled to `gopls` via `textDocument/didOpen`.
 2.  **AI Code Generation/Modification**:
     a.  AI generates or modifies Go code for a specific file.
 3.  **Synchronization**:
     a.  NeuroScript tool (`Gopls.NotifyDidChange`) sends the updated file content to `GoplsClient`, which forwards it to `gopls`.
 4.  **Diagnostic Retrieval**:
     a.  `GoplsClient` receives `textDocument/publishDiagnostics` from `gopls`.
     b.  AI uses NeuroScript tool `Gopls.GetDiagnostics(filePath)` or `Gopls.GetAllProjectDiagnostics()`.
 5.  **AI Error Analysis & Correction**:
     a.  For each diagnostic, the AI examines the `DiagnosticInfo` (message, severity, range).
     b.  To understand the context of an error or a symbol, AI uses:
         i.  `Gopls.GetSymbolInfoAt(filePath, diagnostic.Range.Start.ByteOffset)`: Retrieves `SemanticReference` (including FQN if possible), hover information, and definition location for the symbol at/near the diagnostic.
         ii. `Gopls.GetSymbolInfoByName(fullyQualifiedName)`: If a symbol's FQN is known or inferred.
     c.  AI may also use `Gopls.ListSymbolsInFile` or `Gopls.FindWorkspaceSymbols` to explore available symbols.
     d.  The AI uses this combined diagnostic and semantic information to formulate a correction.
     e.  This information can also feed into tools from `pkg/core/tools/gosemantic` (e.g., using an FQN to find all usages before a rename).
 6.  **Iteration**: Loop back to step 2.

 ## 5. NeuroScript Tool Integration (`pkg/core/tools_gopls.go` - new)

 The following tools will be exposed to NeuroScript:

 -   **Workspace/Initialization**:
     -   `Gopls.SetWorkspaceRoot(workspacePath string) (success bool, error string)`: Initializes or re-initializes `gopls` for a given workspace. This would typically be called once per session or project. The `GoplsClient` should handle the `initialize` handshake.
 -   **Document Synchronization**:
     -   `Gopls.NotifyDidOpen(filePath string, content string) (error string)`
     -   `Gopls.NotifyDidChange(filePath string, newContent string) (error string)`
     -   `Gopls.NotifyDidSave(filePath string) (error string)`
     -   `Gopls.NotifyDidClose(filePath string) (error string)`
 -   **Diagnostics**:
     -   `Gopls.GetDiagnostics(filePath string) (diagnostics []DiagnosticInfo, error string)`
     -   `Gopls.GetAllProjectDiagnostics() (diagnosticsByFile map[string][]DiagnosticInfo, error string)`
 -   **Semantic Information & Context (AI-Focused)**:
     -   `Gopls.GetSymbolInfoAt(filePath string, byteOffset int) (symbolInfo map[string]interface{}, error string)`: Returns a map containing fields like `semanticReference` (map), `hoverContent` (string), `definitionLocation` (map, representing a `SemanticReference`).
     -   `Gopls.GetSymbolInfoByName(fullyQualifiedName string) (symbolInfo map[string]interface{}, error string)`: Similar return to `GetSymbolInfoAt`.
     -   `Gopls.ListSymbolsInFile(filePath string) (symbols []map[string]interface{}, error string)`: Returns a list of symbol maps (each representing a `SemanticReference`).
     -   `Gopls.FindWorkspaceSymbols(queryString string) (symbols []map[string]interface{}, error string)`: Returns a list of symbol maps.

 All tools should return structured error information if the underlying LSP call fails or `gopls` is not responsive.

 ## 6. LSP Communication Details

 -   **Protocol**: JSON-RPC 2.0 over stdio with the `gopls` server.
 -   **Key LSP Methods for v0.5.0**:
     -   Lifecycle: `initialize`, `initialized`, `shutdown`, `exit`.
     -   Synchronization: `textDocument/didOpen`, `textDocument/didChange`, `textDocument/didSave`, `textDocument/didClose`.
     -   Diagnostics: `textDocument/publishDiagnostics` (Notification from server).
     -   Semantic Info: `textDocument/hover`, `textDocument/definition`, `textDocument/documentSymbol`, `workspace/symbol`.
 -   **Content Format**: `gopls` expects UTF-8 for Go files. Position/Range information from LSP can be character-based or UTF-16 offset-based. NeuroScript's `GoplsClient` must handle this:
     -   Negotiate UTF-8 if possible during `initialize`.
     -   If character offsets are used by LSP, convert to/from byte offsets carefully when interacting with NeuroScript tools or file contents, as Go strings and file I/O are typically byte-oriented. For Go source, assuming UTF-8 allows byte offsets to be a reliable internal representation.

 ## 7. Semantic Addressing Strategy (AI-Centric)

 The goal is to provide AI with stable and meaningful ways to refer to code elements, beyond simple line/column numbers which are highly volatile during code modification.

 -   **Primary Identifiers for AI**:
     -   **Fully Qualified Names (FQN)**: Where available (e.g., for package-level functions, types, methods, global variables). This is the most robust identifier. Example: `github.com/project/pkg.MyType.MyMethod`.
     -   **File URI + Precise Byte Offset Range**: For local variables, literals, expressions, or any element where an FQN is not applicable or easily determined. This range is sourced directly from `gopls` responses (e.g., diagnostic ranges, hover ranges, symbol ranges).
 -   **`SemanticReference` Struct**: This struct (defined in Section 3) will be the standard way NeuroScript tools return information about code symbols/elements. It will prioritize FQN and precise byte ranges.
 -   **Tool Inputs**: NeuroScript tools that require a location (e.g., `Gopls.GetSymbolInfoAt`) will accept `filePath` and `byteOffset`. Tools querying by name will accept an FQN (e.g., `Gopls.GetSymbolInfoByName`).
 -   **Interaction with `pkg/core/tools/gosemantic`**:
     -   The FQNs or precise location info obtained from `gopls` tools can be used as reliable inputs for the more advanced static analysis and refactoring tools in `pkg/core/tools/gosemantic`. For example, after `gopls` identifies a symbol at an offset, its FQN (if available from hover/definition) can be passed to `toolGoFindUsages`.

 ## 8. Error Handling

 -   The `GoplsClient` must handle errors related to:
     -   `gopls` server process management (e.g., failing to start).
     -   JSON-RPC communication (e.g., malformed messages, timeouts).
     -   LSP-specific errors returned in responses.
 -   NeuroScript tools wrapping `GoplsClient` methods must translate these errors into standard NeuroScript `RuntimeError`s with appropriate error codes and messages.
 -   Graceful degradation: If `gopls` is unavailable, the tools should fail clearly, allowing the AI to potentially fall back to other analysis methods (e.g., `pkg/core/tools/gosemantic` for non-live analysis, or simpler regex-based checks as a last resort).

 ## 9. Configuration

 -   **`gopls` Executable Path**: NeuroScript will need to know the path to the `gopls` executable. This could be configurable, or it could assume `gopls` is in the system PATH.
 -   **Workspace Root**: The `Gopls.SetWorkspaceRoot` tool will manage this dynamically per session/project. The `Interpreter` or `ng` application context might store the current gopls workspace root.

 ## 10. Future Considerations (Beyond v0.5.0)

 -   Support for `textDocument/codeAction` to get suggested fixes from `gopls`.
 -   Support for `textDocument/rename` and other refactoring capabilities.
 -   Advanced incremental `textDocument/didChange` notifications.
 -   More sophisticated mapping between LSP ranges and NeuroScript's internal AST nodes if deeper AST-LSP correlation is needed.
 -   Support for `gopls` settings/configuration via `workspace/didChangeConfiguration`.

 ---
 ## 11. Document Metadata

 :: version: 0.1.0
 :: type: NSproject
 :: subtype: design_document
 :: project: NeuroScript
 :: purpose: Design for integrating gopls into NeuroScript for advanced Go code diagnostics and semantic understanding, primarily for AI-driven development.
 :: status: draft
 :: author: Gemini (Contributor), AJP (Lead)
 :: created: 2025-05-10
 :: modified: 2025-05-10
 :: dependsOn: docs/gopls_integration.md (feasibility study), pkg/core/tools/gosemantic/*, pkg/core/ai_rules.md
 :: reviewCycle: 1
 :: nextReviewDate: 2025-05-17
 ---