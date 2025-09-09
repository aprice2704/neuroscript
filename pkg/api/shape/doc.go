// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides package-level documentation for the public 'shape' API.
// filename: pkg/api/shape/doc.go
// nlines: 12
// risk_rating: LOW

/*
Package shape provides a stable, public API for NeuroScript's Path-Lite and Shape-Lite
functionality.

It serves as a facade over the internal `json_lite` implementation, exposing a
consistent and versioned set of types, functions, and errors for other packages
to consume. This allows the underlying implementation to evolve without breaking
dependent code.
*/
package shape
