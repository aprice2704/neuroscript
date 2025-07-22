package parser

// -----------------------------------------------------------------------------
//  SIMPLE COMMENT‑ASSOCIATION ALGORITHM
// -----------------------------------------------------------------------------
//
// Goal
// -----
// Associate every stand‑alone comment with the *last* code node that appeared
// earlier in the file.  No heuristics, no blank‑line math, no special cases.
// Deterministic, repeatable, and trivial to reason about.
//
// Definitions
// -----------
// • Code node  = *ast.Program, *ast.Procedure, *ast.Step, *ast.CommandNode,
//                *ast.OnEventDecl    (extend if new node types are added)
//
// • “Stand‑alone comment” = any LINE_COMMENT token (#, //, --) that ANTLR
//   emits on the hidden channel, **regardless of whether it shares a line
//   with code or sits on its own line**.  Inline metadata comments are
//   ignored here; they stay with the metadata token.
//
// High‑level rule
// ---------------
//   For each comment encountered while scanning the source in order,
//   attach it to `lastCodeNode`, where `lastCodeNode` is the most recent
//   code node we have already seen.  At file start, `lastCodeNode` is the
//   `*ast.Program`.
//
// Step‑by‑step procedure
// ----------------------
// 1. Build a map   startLine → ast.Node
//      • Walk the finished AST, record every code node’s starting line.
//      • Always map line 1 → Program so header comments have a home.
//
// 2. Obtain a complete token stream (`*antlr.CommonTokenStream`).
//
// 3. Iterate over the tokens **in original source order**:
//
//      lastCodeNode := *ast.Program   // initial value
//
//      for each token t:
//          if t is on DEFAULT channel {        // i.e. code
//              if nodeAtLine := startLine[t.Line]; nodeAtLine != nil {
//                  lastCodeNode = nodeAtLine
//              }
//          }
//
//          if t.TokenType == LINE_COMMENT {     // hidden‑channel comment
//              comment := &ast.Comment{Text: t.Text}
//              attach(comment, lastCodeNode)
//          }
//
//      // done
//
// 4. `attach(comment, node)` simply appends the comment to the node’s
//    `Comments` slice.  No further processing.
//
// Consequences / guarantees
// -------------------------
// • All comments are preserved—none are lost or duplicated.
// • Comment ownership is *totally independent* of spacing or blank lines.
// • The algorithm is O(N) in tokens, negligible memory.
// • Future maintainers can re‑implement it in a few minutes.
//
// If more nuance is required later (e.g. blank‑line counts, “attach to next
// node when adjacent”), those layers can be added *on top* of this minimal
// baseline.
//
// -----------------------------------------------------------------------------
