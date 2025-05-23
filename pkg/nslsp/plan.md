 # NeuroScript Diagnostics & Language Server (NSLSP) - Project Plan & Checklist
 
 **Version:** 0.1.0
 **Date:** May 22, 2025
 
 ## 1. Goals
 
 - Provide robust, real-time diagnostics for NeuroScript in IDEs.
 - Enhance the NeuroScript development experience with modern IDE features (completion, hover, goto definition).
 - Serve as a foundational piece of the NeuroScript tooling ecosystem.
 - Offer an opportunity to gain experience with LSP implementation, potentially informing future Go LSP client integration within NeuroScript itself.
 
 ## 2. Architecture
 
 - **Language:** Go
 - **Protocol:** Language Server Protocol (LSP) v3.17 (or latest stable)
 - **Core Dependencies:**
     - NeuroScript Parser (`pkg/core/parser_api.go`, ANTLR-generated code in `pkg/core/generated/`)
     - NeuroScript AST (`pkg/core/ast.go`, `pkg/core/ast_builder_main.go`)
     - JSON-RPC library for LSP communication (e.g., `sourcegraph/go-lsp` or similar, to be evaluated)
     - NeuroScript Tool Registry (`pkg/core/tools_registry.go`) for tool-related features.
 - **Proposed Directory Structure:**
     - `cmd/nslsp/main.go`: LSP server executable entry point.
     - `pkg/nslsp/`: Core LSP server implementation.
         - `server.go`: Main server logic, request routing.
         - `protocol.go`: LSP message type definitions (or use a library).
         - `documents.go`: Document manager (tracks open files, content).
         - `diagnostics.go`: Syntax and semantic error reporting.
         - `features.go`: Implementation for hover, completion, definition, etc. (can be broken into sub-packages).
         - `astutils.go`: Utilities for working with the NeuroScript AST.
         - `config.go`: Server configuration.
 
 ## 3. Project Phases & Checklist
 
 ### Phase 1: Basic LSP Server & Syntax Diagnostics
 
 - **P1.1: Setup & Communication**
     - [ ] Research and select a Go JSON-RPC/LSP library.
     - [ ] Create `cmd/nslsp/main.go` entry point.
     - [ ] Implement basic server lifecycle (initialize, initialized, shutdown, exit).
     - [ ] Establish I/O handling (stdio by default for LSP).
     - [ ] Set up logging for the LSP server itself.
 - **P1.2: Document Management**
     - [ ] Implement `textDocument/didOpen`: Store document content.
     - [ ] Implement `textDocument/didChange`: Update document content (initially full content, consider incremental later).
     - [ ] Implement `textDocument/didClose`: Remove document from active tracking.
     - [ ] Implement `textDocument/didSave`.
 - **P1.3: Parsing & Syntax Error Diagnostics**
     - [ ] Integrate existing NeuroScript ANTLR parser (`pkg/core/parser_api.go`).
     - [ ] On `didOpen` and `didChange`, parse the document content.
     - [ ] Adapt ANTLR ErrorListener to collect detailed syntax error information (line, column, offending token, message).
     - [ ] Convert parser errors to LSP `Diagnostic` objects.
     - [ ] Send diagnostics to the client via `textDocument/publishDiagnostics`.
     - [ ] **Goal:** Syntax errors are underlined in the IDE with messages.
 
 ### Phase 2: Core Language Features
 
 - **P2.1: Document Symbols (Outline View)**
     - [ ] Implement `textDocument/documentSymbol`.
     - [ ] Traverse the AST to find procedure definitions (`func` blocks) and their locations.
     - [ ] **Goal:** IDEs can display a symbol outline for `.ns` files.
 - **P2.2: Hover Information (Basic)**
     - [ ] Implement `textDocument/hover`.
     - [ ] On hover, identify the token/AST node under the cursor.
     - [ ] For procedure names, display their `:: description` or other key metadata.
     - [ ] For tool calls, display tool `ToolSpec.Description`.
     - [ ] **Goal:** Basic information appears on hover for key elements.
 - **P2.3: Completion (Basic Keywords)**
     - [ ] Implement `textDocument/completion`.
     - [ ] Provide completion items for all NeuroScript keywords.
     - [ ] **Goal:** Basic keyword completion is available.
 
 ### Phase 3: Enhanced Language Features & Semantics
 
 - **P3.1: Go to Definition (Procedures)**
     - [ ] Implement `textDocument/definition`.
     - [ ] For a procedure call, find the `func` definition within the current document.
     - [ ] (Advanced: Extend to workspace-wide definition search if procedures can be in other files).
     - [ ] **Goal:** Can jump from a procedure call to its definition.
 - **P3.2: Enhanced Completion**
     - [ ] Add completion for defined variables in the current scope.
     - [ ] Add completion for defined procedure names in the current file/workspace.
     - [ ] Add completion for registered tool names (e.g., `tool.FS.Read`) by querying the tool registry.
     - [ ] Add completion for standard metadata keys (e.g., `:: description:`, `:: param:`).
 - **P3.3: Semantic Diagnostics (Initial)**
     - [ ] Implement basic semantic checks post-parsing:
         - [ ] Check for calls to undefined procedures (within the current file initially).
         - [ ] Check for incorrect number of arguments to known procedures/tools (requires tool specs).
         - [ ] Check for use of undefined variables (basic scope analysis).
     - [ ] Report these as `Diagnostic` objects.
     - [ ] **Goal:** More than just syntax errors are reported.
 
 ### Phase 4: Advanced Features & Refinements
 
 - **P4.1: Find References (Procedures, Variables)**
     - [ ] Implement `textDocument/references`.
     - [ ] Requires more advanced symbol table management and potentially indexing if workspace-wide.
 - **P4.2: Code Formatting (Optional)**
     - [ ] Implement `textDocument/formatting` and/or `textDocument/rangeFormatting`.
     - [ ] Requires defining a canonical NeuroScript code style.
 - **P4.3: Signature Help**
     - [ ] Implement `textDocument/signatureHelp` for procedure and tool calls.
 - **P4.4: Workspace Symbols**
     - [ ] Implement `workspace/symbol` to search for symbols across the workspace.
 - **P4.5: Performance Optimization**
     - [ ] Profile and optimize parsing and analysis for large files/projects.
     - [ ] Consider incremental parsing/analysis for `textDocument/didChange`.
 
 ### Phase 5: Packaging & Distribution
 
 - **P5.1: VS Code Extension Integration**
     - [ ] Update the `vscode-neuroscript` extension's `package.json` to declare the LSP client.
     - [ ] The client part in the VS Code extension will be responsible for starting and communicating with the `nslsp` executable.
 - **P5.2: Testing**
     - [ ] Implement unit tests for individual components of the LSP server.
     - [ ] Implement integration tests for LSP features.
 - **P5.3: Documentation**
     - [ ] Document the LSP server's features and how to use it.
     - [ ] Document how to integrate with VS Code (and potentially other IDEs).
 
 ## 4. Key Considerations & Challenges
 
 - **Scope Management:** Implementing reliable scope analysis for variables will be key for accurate diagnostics and completion.
 - **Tool Registry Access:** The LSP server will need an efficient way to access information about available tools and their specifications (argument types, return types, descriptions).
 - **Workspace Awareness:** Handling multi-file projects, cross-file definitions, and library paths.
 - **Performance:** Ensuring the server is responsive, especially during typing (`didChange` events).
 - **Error Recovery in Parser:** A parser that can recover from errors and still produce a partial AST is highly beneficial for providing diagnostics on incomplete or erroneous code.
 - **Incremental Updates:** For performance, handling incremental text changes (`TextDocumentContentChangeEvent`) rather than re-parsing the whole file every time is ideal but more complex.
 
 ## 5. Open Questions / Decisions
 
 - Which Go LSP library/framework to use (e.g., `sourcegraph/go-lsp`, `golang.org/x/tools/gopls` structure for inspiration, or build from scratch with JSON-RPC)?
 - Initial focus: single-file analysis or immediate workspace awareness? (Single-file is simpler to start).
 - Strategy for accessing tool definitions (e.g., does the LSP server re-use parts of the `core.Interpreter`'s tool registry, or have its own way to load `tooldefs_*.go` files?).
 - How to handle NeuroScript versions if the language evolves? (The LSP might need to be aware of the `:: lang_version:`).
