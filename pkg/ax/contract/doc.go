// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Provides package-level documentation for the ax/contract package.
// filename: pkg/ax/contract/doc.go
// nlines: 12
// risk_rating: LOW

/*
Package contract defines the stable, minimal interfaces for the Runner Parcel
model.

Its purpose is to provide a neutral ground that both the high-level `ax`
package and the low-level `interpreter` package can depend on without creating
import cycles. It contains only the core contracts and their simple, concrete
implementations.
*/
package contract
