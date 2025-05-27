# GoIndexer Enhancement Plan: Richer, Component-Based, Accessible Index

**Overall Goal:** Evolve `goindexer` to produce a more detailed, multi-file code index for the NeuroScript Go project. The index will focus on exported Go language entities and NeuroScript tool specifications to improve AI-assisted development, provide a foundation for advanced tooling, and enable programmatic access to the index from NeuroScript itself (via a NeuroScript tool or direct AI function calls). GoDoc processing is deferred.

**Key Output Artifacts:**
1.  **Project Index File (e.g., `project_index.json`):** An overall file listing defined "components" (e.g., "core", "nslsp"), their source paths, and pointers to their individual component index files.
2.  **Component Index Files (e.g., `core_index.json`, `nslsp_index.json`):** Detailed JSON files for each logical component of the codebase, containing rich information about exported entities and NeuroScript tools.

---
## Phase 1: Enhance `goindexer` for Detailed Exported Entity Parsing

**Goal:** Enable `goindexer` to parse and extract detailed information about exported Go entities from source files, focusing on information density and type accuracy.

* **P1.1: Define Rich Data Structures for Index (JSON Output)**
    * [X] In `main.go` (or a new `types.go` within the `main` package), define new/updated Go structs to represent the rich index information. These will replace or augment the current `Index`, `PackageInfo`, `FileInfo`, `MethodInfo` structs. (Completed with "plan-aligned version 2.0.0" of `pkg/goindex/types.go`)
        * **`ProjectIndex`**: Structure for `project_index.json`.
        * **`ComponentIndex`**: Structure for component-specific files (e.g., `core_index.json`).
        * **`PackageDetail`**: (Within `ComponentIndex`) Contains lists of `FunctionDetail`, `MethodDetail`, `StructDetail`, `InterfaceDetail`, etc.
        * **`ParamDetail`**: `{"name": "string", "type": "string"}`.
        * **`FunctionDetail`**: `name` (fully qualified), `sourceFile` (relative to component), `parameters: []ParamDetail`, `returns: []string` (type strings).
        * **`MethodDetail`**: `receiverName`, `receiverType` (fully qualified), `name`, `sourceFile`, `parameters: []ParamDetail`, `returns: []string`.
        * **`FieldDetail`**: `{"name": "string", "type": "string", "tags": "string", "exported": true}`.
        * **`StructDetail`**: `name`, `sourceFile`, `fields: []FieldDetail`.
        * **`InterfaceDetail`**: `name`, `sourceFile`, `methods: []MethodDetail` (for the interface method set).
        * **`GlobalVarDetail` / `GlobalConstDetail`**: `name`, `type`, `sourceFile`, `value` (for simple literal consts).
        * **`TypeAliasDetail`**: `name`, `underlyingType`, `sourceFile`.
    * _Note:_ All type strings should aim for parseable Go type representations (e.g., `[]*mypkg.MyType`, `map[string]error`). Explicitly capture `error` types in return signatures. Ensure `FunctionDetail.name` stores fully qualified Go function names.
* **P1.2: Implement Enhanced Go AST Parsing (primarily in `parser.go`)**
    * [X] Modify `processFile` in `parser.go` to traverse the `ast.File` node. (Refactored to populate new structures)
    * [X] For `ast.FuncDecl`:
        * Extract full signature details (parameters, return types) for functions and methods. Use/enhance `formatters.go`'s `formatFieldList` and `formatNode` for accurate type stringification. (Done, with `formatters.go` updated)
        * Distinguish between functions and methods based on `decl.Recv`. (Done)
        * Populate `FunctionDetail` and `MethodDetail` structs. (Done)
    * [X] For `ast.GenDecl` (with `tok == token.TYPE`):
        * Iterate `decl.Specs`. For each `*ast.TypeSpec`:
            * Extract struct fields (exported only) for `*ast.StructType`, populating `StructDetail`. (Done)
            * Extract interface methods (with signatures) for `*ast.InterfaceType`, populating `InterfaceDetail`. (Done, using `MethodDetail`)
            * Identify type aliases/definitions, populating `TypeAliasDetail`. (Done, `PackageDetail` now includes `TypeAliases`)
            * Use/enhance `formatters.go`'s `determineKind` and `formatNode` for type information. (Done)
    * [X] For `ast.GenDecl` (with `tok == token.VAR` or `tok == token.CONST`):
        * Iterate `decl.Specs`. For each `*ast.ValueSpec`:
            * Extract names, types, and simple literal values (for consts), populating `GlobalVarDetail` / `GlobalConstDetail`. (Done)
    * [X] Ensure all extracted entities are checked for export status (capitalized names). (Handled by `ast.IsExported` checks)
    * [X] **Decision Point:** Start with `go/ast` and `formatNode` from `formatters.go` for type representation. (Maintained this approach)
* **P1.3: Update `main.go` for Initial Rich Output Generation**
    * [X] Modify the top-level `Index` struct in `main.go` (or replace it) to hold the new detailed entity structures (now using `ProjectIndex` and `ComponentIndex`).
    * [X] Ensure `main.go` can serialize this richer index to a single JSON output file for debugging and validation of P1.1 and P1.2. (Now generates component and project files).
    * [X] Add modified time (of index), git branch etc. to index so staleness is detectable. (Added to `ProjectIndex` and `ComponentIndex` in `types.go` and populated in `main.go`)

---
## Phase 2: Implement Component-Based Indexing & Programmatic Access

**Goal:** Structure the index into manageable components, create a project-level index, and design for programmatic access to the indexed data by other Go code.

* **P2.1: Define Project "Components" and Configuration**
    * [X] In `main.go`, implement a mechanism to define components. (Using `defaultComponentDefs` map/slice).
    * [X] `finder.go`'s `findRepoPaths` will continue to establish the project `repoRootPath` and `repoModulePath`. (Assumed working as `main.go` uses these).
* **P2.2: Implement Component-Aware File Processing in `main.go`**
    * [X] Modify `main.go`'s directory scanning logic (the `walkFunc`). (Done).
    * [X] For each processed `.go` file, determine which component it belongs to based on its path relative to `repoRootPath` and the component definitions from P2.1. (Using `assignFileToComponent`).
    * [X] Adapt data collection to aggregate indexed information into separate data structures, one per component. (Using `componentIndexes` map).
* **P2.3: Generate Component-Specific and Project Index Files**
    * [X] Modify `main.go` to write the detailed index for each component into its own JSON file (e.g., `core_index.json`). Paths like `sourceFile` within these component indexes should be relative to that component's root directory. (Done).
    * [X] `main.go` to generate the `project_index.json` file. This file will list:
        * `schemaVersion` (for the project index).
        * `repoModulePath` (overall module path).
        * A map/list of `components`, each entry containing: `name` (e.g., "core"), `path` (e.g., "pkg/core"), `indexFile` (e.g., "core_index.json").
        * `lastIndexedTimestamp`. (Done).
* **P2.4: Design for Index Accessibility (for `ns` Tool / AI Function Calls) - "ns Tool version"**
    * [X] Define Go structs (e.g., in a new `goindexer/reader` package or `pkg/goindexreader` if `goindexer` becomes a library module) that mirror the structure of `project_index.json` and a `component_index.json` file for easy unmarshalling. (Done in `pkg/goindex/types.go`).
    * [X] Create functions in this new reader package: (`pkg/goindex/reader.go` now contains these)
        * `LoadProjectIndex(projectIndexFilePath string) (*ProjectIndexData, error)`
        * `LoadComponentIndex(componentMeta ComponentMeta, projectRootPath string) (*ComponentIndexData, error)` (Now `(r *IndexReader) GetComponentIndex(componentName string)`)
        * Provide query methods on `ComponentIndexData` (e.g., `GetComponentFunction(packageName, funcName)`, `GetComponentStruct(packageName, structName)`). (Implemented as `FindFunction`, `FindStruct`, `FindMethod` on `IndexReader`).
    * [X] This reader package must be importable by `pkg/core` (where the NeuroScript tool will be implemented) without creating circular dependencies with `goindexer`'s `main` package. (Achieved by `pkg/goindex` being a separate package).

---
## Phase 3: Index NeuroScript Tool Specifications (Revised Dynamic Approach)

**Goal:** Enrich the index with detailed information about available NeuroScript tools by linking their `ToolSpec` definitions to their implementing Go functions using dynamic introspection.

* **P3.1: Enhance `core.ToolSpec` and Implement JSON Meta-Tool for Dynamic Access**
    * [X] Define `Category string` and `Example string` fields (also `ReturnHelp`, `Variadic`, `ErrorConditions`) in the `core.ToolSpec` struct (`pkg/core/tools_types.go`).
    * [~] Update existing `ToolSpec` definitions in *other* `tooldefs_*.go` files throughout the `pkg/core` and other relevant packages to populate these new `Category`, `Example`, `ReturnHelp`, `Variadic`, and `ErrorConditions` fields. (This is your pending task, "swarm of agents" - User indicates this is now mostly done).
    * [X] Implement a new meta-tool in `pkg/core` (e.g., `Meta.GetToolSpecificationsJSON()`):
        * This tool, when called on a `core.Interpreter` instance, will retrieve all registered tools. (Done in `tooldefs_meta.go` and `tools_meta.go`)
        * It will collect their `ToolSpec` data (now including `Category`, `Example`, etc.). (Done)
        * It will serialize this list of `ToolSpec` objects into a single JSON string. (Done)
* **P3.2: Link Tool Specs to Go Implementation Details via Dynamic Analysis by `goindexer`**
    * [X] `goindexer` will be enhanced to initialize a "no-op" `core.Interpreter` instance. This involves ensuring all necessary core and extended toolsets are registered with this interpreter. (Done in `cmd/goindexer/main.go`)
    * [X] `goindexer` will call the new `Meta.GetToolSpecificationsJSON()` meta-tool (from P3.1) using this no-op interpreter to retrieve the `ToolSpec` data for all registered NeuroScript tools. (Done via `indexReader.GenerateEnhancedToolDetails` in `cmd/goindexer/main.go`)
    * [X] For each `ToolSpec` obtained:
        * `goindexer` will use `interpreter.GetTool(toolSpec.Name)` on its no-op interpreter to get the corresponding `core.ToolImplementation`. (Done via `indexReader.GenerateEnhancedToolDetails`)
        * From the `ToolImplementation.Func` field (which is a `core.ToolFunc`), `goindexer` will use Go's runtime reflection (e.g., `runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()`) to determine the fully qualified Go function name of the tool's implementing function. (Done via `indexReader.GenerateEnhancedToolDetails`)
    * [X] `goindexer` will then use this fully qualified Go function name to look up its `FunctionDetail` from the data indexed in Phase 1. This `FunctionDetail` provides critical information like the Go function's `sourceFile` (relative to its component), its specific parameters, and its return types. (Done via `indexReader.GenerateEnhancedToolDetails`, results used in `cmd/goindexer/main.go`)
* **P3.3: Integrate Tool Data into Component Index**
    * [X] Define a `neuroscriptTools` field (e.g., an array of structured tool information objects, `NeuroScriptToolDetail`) in the JSON schema for relevant component indexes (Completed in `pkg/goindex/types.go` within `ComponentIndex`).
    * [X] Populate this with:
        * The `ToolSpec` data (`Name`, `Description`, `Args`, `ReturnType`, `Category`, `Example`, `ErrorConditions`) obtained from the JSON meta-tool (P3.1).
        * The `implementingGoFunctionFullName`: The fully qualified name of the Go function that implements the tool (obtained in P3.2).
        * The `implementingGoFunctionSourceFile`: The source file path (relative to its component) of the Go function (obtained from Phase 1 data via the lookup in P3.2).
        * A flag or field indicating Go-level error handling, e.g., `goImplementingFunctionReturnsError: true/false` (derived from the `returns` field of the `FunctionDetail` looked up in P3.2).
        (Done in `cmd/goindexer/main.go` using data from `indexReader.GenerateEnhancedToolDetails`)

---
## Phase 4: Schema Definition & Finalization

**Goal:** Formalize and document the new JSON schema(s) for all index files.

* **P4.1: Define and Document JSON Schemas**
    * [ ] Create/update schema definition files (e.g., `project_index_schema_vX.Y.json`, `component_index_schema_vX.Y.json`) reflecting all new fields and structures. The version should be incremented (e.g., to 2.0).
* **P4.2: Schema Versioning in Output Files**
    * [X] Ensure `goindexer` embeds the correct `IndexSchemaVersion` string (e.g., "project_index_v2.0.0", "component_index_v2.0.0") into each generated JSON file. (Added to `types.go` and `main.go` initializations).
* **P4.3: (Optional) Index Validation**
    * [ ] Consider creating a separate script or adding a mode to `goindexer` to validate existing index files against their declared schemas.

---
## Phase 5: Deferred (Future Enhancements)
* [ ] GoDoc comment processing (summarization/rationalization).
* [ ] Advanced call graph analysis (more structured "calls" and "calledBy" information in the index, potentially leveraging `resolvers.go`'s current capabilities).
* [ ] Incremental indexing capabilities for `goindexer` to improve performance on subsequent runs.