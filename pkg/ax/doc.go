// Package ax defines a tiny, composable "ports" layer that downstream
// consumers (FDM/Zadeh/Lotfi/tools) can rely on without importing NS internals.
//
// Goals:
//   - Distinct naming to avoid collisions with existing "Interpreter/Runtime" names
//   - Compile-time contracts (no `any` for identity, stores, etc.)
//   - Interfaces sized to consumer needs (FDM’s 6 responsibilities)
//   - Stable API surface: additive changes preferred, no init() magic
//
// This package is safe to import from outside NS and is intended to be re-used
// across FDM and Zadeh without causing import cycles or leaking private types.
package ax
