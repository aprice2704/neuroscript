To add **semantic error checking** (beyond what the grammar captures) to your LSP support for VSCode and Vim, you need to extend your current setup in the following structured way:

---

## 🧠 Semantic Errors vs. Grammar Errors

* **Grammar errors**: Syntax-level, handled by ANTLR + LSP grammar definitions.
* **Semantic errors**: Logical/structural issues in a *syntactically valid* program (e.g., undefined variables, type mismatches, misuse of functions).

You now need an **analyzer** layer that:

1. **Consumes the AST** (produced from parsing)
2. **Runs semantic checks**
3. **Reports diagnostics** back to the editor via the LSP

---

## ✅ Strategy Overview

### 1. **Build a Semantic Analyzer**

Use your existing AST (likely Go structs from ANTLR's parse) and traverse it.

For example:

* Undefined variables
* Calling non-functions
* Type mismatches
* Invalid event handlers or tools
* Violations of `final`/`nooverride` semantics

You'll define:

```go
type SemanticAnalyzer struct {
    diagnostics []lsp.Diagnostic
    symbolTable SymbolTable
    // possibly more: scope stack, current file, etc.
}
```

Then walk the AST with visitors or a manual traversal, accumulating diagnostics:

```go
func (s *SemanticAnalyzer) VisitFunctionCall(node *ast.FuncCall) {
    if !s.symbolTable.IsDefined(node.Name) {
        s.diagnostics = append(s.diagnostics, makeDiagnostic(node.Pos, "Function not defined"))
    }
}
```

You already have most of this logic in the interpreter or runtime — extract and reuse where possible.

---

### 2. **Expose Semantic Analyzer to the LSP Server**

Your LSP server (likely using `golang.org/x/tools/gopls`-style wiring or `lsp-protocol`) should now do the following in `textDocument/didOpen` or `textDocument/didChange`:

```go
func (s *Server) onDidOpen(params lsp.DidOpenTextDocumentParams) {
    parseTree := parser.Parse(params.Text)
    ast := builder.BuildAST(parseTree)
    
    analyzer := semantic.NewAnalyzer()
    diagnostics := analyzer.Analyze(ast)

    s.conn.PublishDiagnostics(params.TextDocument.URI, diagnostics)
}
```

You report these using the LSP Diagnostic type:

```go
lsp.Diagnostic{
    Range:    lsp.Range{Start: ..., End: ...},
    Severity: lsp.SeverityError,
    Message:  "Function not defined",
}
```

---

### 3. **Connect to VSCode and Vim**

You already have the grammar hooked up, so:

* **VSCode**: In your extension, make sure the LSP client can receive diagnostics. Usually handled via the language server client setup.
* **Vim/Neovim**: `coc.nvim`, `nvim-lspconfig`, or similar can display these diagnostics in-line.

---

## 🛠 Helpful Infrastructure

* **Symbol Table with Scopes**: For resolving vars, types, tools
* **Type system**: Leverage your `TypeOf()` work
* **Error reporting utilities**: Map AST positions to line/column spans
* **Test harness**: Validate semantic checks with in-memory scripts

---

## Example Errors You Might Flag

| Error Type           | Example              | Message                                             |
| -------------------- | -------------------- | --------------------------------------------------- |
| Undefined var        | `x = y + 1`          | `Variable 'y' not defined`                          |
| Type mismatch        | `x = "hello" + 5`    | `Cannot add string and number`                      |
| Tool misuse          | `tool("unknown")`    | `Tool 'unknown' not found`                          |
| Function misuse      | `main = 5; main()`   | `Cannot call non-function 'main'`                   |
| `final` redefinition | second `func main()` | `Function 'main' is final and cannot be overridden` |

---

## 📦 Optional Bonus: Inlay Hints & Code Actions

Once semantic diagnostics work:

* **Inlay hints**: Show types of vars or function results
* **Code actions**: Suggest fixes (e.g., "Declare variable ‘x’")

---

Would you like a working `SemanticAnalyzer` Go skeleton to get started?


This is a deep and important question. You're asking how to:

1. **Extract** the semantic rules implicit in your interpreter/runtime.
2. **Codify** them into explicit form, ideally decoupled from the interpreter.
3. **Keep** them **up to date** as the language evolves.

Here’s a rigorous way to approach it:

---

## 🧱 1. **Extracting Semantic Rules from Source**

### A. **Trace Existing Behavior**

Start by reading/interpreting the *implicit rules* in your runtime/interpreter:

* When do you raise runtime errors?
* What assumptions do your function calls, tool dispatchers, etc. make?
* Where do you panic or return `ErrInvalid`?

### B. **Catalog Rules**

Build a table of semantic rules as you find them, e.g.:

| Rule ID | Description                                       | Context               |
| ------- | ------------------------------------------------- | --------------------- |
| `SR001` | Variables must be declared before use             | Expression evaluation |
| `SR002` | Tool names must resolve to a known tool impl      | Tool call             |
| `SR003` | Functions must not be redefined if marked `final` | Function declaration  |

You may already have this half-implied in your interpreter's behavior — now it's about **pulling those rules into explicit form**.

---

## 📐 2. **Codifying Rules into an Independent Layer**

Move toward a **rule-driven semantic analyzer**, i.e. a separate phase that consumes AST and applies declarative checks.

Two main strategies:

### A. **Visitor + Rule Registry**

Each rule is a function with a defined scope:

```go
type Rule struct {
    ID       string
    Name     string
    AppliesTo func(ast.Node) bool
    Check    func(node ast.Node, ctx *AnalysisContext) []Diagnostic
}
```

You iterate over AST nodes, dispatch to all rules where `AppliesTo(node)` returns true.

This makes adding new rules **composable and testable**.

### B. **Rule DSL or Metadata**

Optionally, define rule specs in a separate format (e.g. YAML or JSON) and load them:

```yaml
- id: SR002
  applies_to: ToolCall
  message: "Tool '{name}' not found"
  condition: "!ctx.IsKnownTool(node.name)"
```

This is more abstract, but harder to maintain without a mini-language or codegen tool. Might be overkill for NS/FDM.

---

## 🔁 3. **Keeping Rules Up-to-Date**

This is where rigor pays off.

### A. **Bind Rules to Tests**

Every semantic rule should be validated via a **unit test or integration test** that breaks the rule:

* Test scripts can live in `semantic_tests/`
* Each test declares which rule(s) it should trigger
* Test framework asserts expected diagnostics

### B. **Link Rules to Interpreter Commits**

Tag interpreter changes with references to affected semantic rules:

```go
// Interpreter: updated to support byte literals
// Affects: SR027, SR033
```

You can even generate a rule coverage report: "Which rules are exercised by tests? Which are untested?"

### C. **Mirror from Spec if Available**

If you maintain a language spec (like `ns_script_spec.md`), you can:

* Cross-link rule IDs to spec sections
* Auto-lint: warn if spec describes a rule not enforced in code

---

## 🛠 Tooling Helpers

| Tool             | Use                                                                      |
| ---------------- | ------------------------------------------------------------------------ |
| `grep`/`go list` | Find all `error` returns or `panic` calls to locate semantic assumptions |
| Static analysis  | Use `go vet` or custom analyzers to find unhandled conditions            |
| Doc generators   | Auto-generate a semantic rule doc from Go code comments                  |

---

## 🧠 Bottom Line

> **Semantic rules are an informal spec made formal**.

By codifying them in a rule layer separate from execution, you gain:

* Better editor integration
* Easier testing and enforcement
* Faster evolution tracking

---

If you'd like a starter template for rule definition in Go, or a test harness structure for semantic cases, I can provide one.
