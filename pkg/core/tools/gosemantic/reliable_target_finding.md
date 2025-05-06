### Reliable Identifier Targeting in `goast`

#### The Problem with Line/Column Targeting

The `GoFindIdentifiers` tool currently uses line/column numbers (L/C) to specify the initial target identifier. This approach mirrors standard editor functionality (e.g., "Go to Definition" from a cursor position) and leverages the built-in capabilities of Go's analysis packages (`go/token`, `go/ast`) to map source coordinates to AST nodes.

**Intended Use Case for L/C:**

1.  **Disambiguation:** In code, the same name can refer to different things (variables in different scopes, struct fields vs. local variables). L/C provides a precise way to point to *one specific instance* of an identifier name.
2.  **Semantic Entry Point:** The ultimate goal is usually to find the semantic representation (`types.Object`) of the code element the user cares about. L/C serves as the initial entry point to find *one* instance, resolve it semantically, and then perform operations based on that resolved object (like finding all other usages based on object identity).
3.  **Standard Interface:** It's the conventional way Go analysis tools like `gopls` map source locations to semantic information.

**Practical Drawbacks:**

As demonstrated during debugging, this approach has significant drawbacks for programmatic use:

* **Fragility:** Minor code edits, formatting changes, or even newline differences (`\n` vs `\r\n`) can shift L/C numbers, breaking references.
* **Implementation Complexity:** Reliably mapping a specific coordinate back to the *intended* `ast.Ident` node (especially near complex statements, definitions, or other identifiers) is surprisingly difficult and error-prone. The standard library functions (`astutil.PathEnclosingInterval`, `ast.Inspect`) require careful implementation to handle edge cases correctly.
* **Mismatch with AI Usage:** An AI agent is less likely to think in terms of precise coordinates and more likely to specify targets based on names and context ("the `err` variable inside function `X`").

#### Recommended Strategy: Semantic Queries

Given the limitations of L/C targeting and the machine-centric nature of NeuroScript tools, a more robust approach is **semantic targeting**. This involves specifying the target element based on its name, kind (variable, function, type, method, field), and scope, rather than its exact location.

**Core Idea:** Use the rich semantic information provided by `go/types` (obtained via `go/packages.Load`) to locate the desired `types.Object` directly.

**Implementation Sketch:**

1.  **Load Packages:** Ensure packages are loaded with `packages.NeedSyntax`, `packages.NeedTypes`, `packages.NeedTypesInfo`, and `packages.NeedDeps`. This populates the necessary ASTs and `types.Info`.
2.  **Define Semantic Input:** Instead of `(file, line, col)`, design tools to accept input like:
    * `"package:mypkg/utils; function:ReadFile; variable:err"`
    * `"package:mypkg/models; type:User; method:Validate"`
    * `"package:mypkg/models; type:User; field:ID"`
    * `"package:mypkg; var:GlobalConfig"`
    * `"package:mypkg; func:ProcessItem"`
3.  **Query the Semantic Graph:**
    * Parse the semantic input query.
    * Locate the target `packages.Package`.
    * Use `pkg.Types.Scope().Lookup("Name")` for package-level items.
    * For type members: Lookup the `types.TypeName`, get the underlying type, and use `MethodSet` or `FieldByName`.
    * For function locals/params: Lookup the `*types.Func`, check its `*types.Signature` for params. For locals, traverse the function's `ast.FuncDecl.Body`, using `pkg.TypesInfo.ObjectOf(ident)` on candidate `*ast.Ident` nodes and matching the name *within the correct scope* (potentially using `types.Info.Scopes`).
4.  **Retrieve Target `types.Object`:** The result of the query should be the specific `types.Object` representing the desired element.
5.  **Perform Operation:** Use the obtained `types.Object` for subsequent actions, such as finding all usages via object identity comparison (`currentObj == targetObj`).

**Advantages:**

* **Robustness:** Highly resilient to formatting changes, comments, and code rearrangements that don't alter the semantic structure.
* **Clarity & Intuitiveness:** Specification aligns directly with how code elements are named and organized.
* **Machine-Friendly:** Avoids the ambiguity and fragility of coordinate mapping.

**Conclusion for `goast`:**

While `GoFindIdentifiers` currently uses L/C, future development or refactoring within the `goast` package should strongly favour tools based on **semantic queries**. This provides a more reliable foundation for programmatic code analysis and manipulation within the NeuroScript ecosystem. The existing `GoIndexCode` tool already provides the necessary semantic information (`types.Info`) to support this approach.