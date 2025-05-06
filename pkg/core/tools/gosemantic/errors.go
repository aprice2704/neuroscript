// NeuroScript Version: 0.3.1
// File version: 0.0.1
// Define sentinel errors for gosemantic package.
// filename: pkg/core/tools/gosemantic/errors.go

package gosemantic

import "errors"

var (
	// ErrSymbolNotFound indicates that the requested symbol could not be located via semantic query.
	ErrSymbolNotFound = errors.New("semantic symbol not found")
	// ErrPackageNotFound indicates that the package specified in a query was not found in the index.
	ErrPackageNotFound = errors.New("package not found in index")
	// ErrInvalidQueryFormat indicates that the semantic query string is malformed.
	ErrInvalidQueryFormat = errors.New("invalid semantic query format")
	// ErrWrongKind indicates that a symbol was found but it was not of the expected kind (e.g., looking for a function but found a type).
	ErrWrongKind = errors.New("symbol found but has wrong kind")
	// ErrAmbiguousQuery indicates the query matched multiple symbols and requires more specificity.
	ErrAmbiguousQuery = errors.New("ambiguous semantic query")
	// ErrIndexNotReady indicates the SemanticIndex is missing required components (e.g., nil Fset or Packages map).
	ErrIndexNotReady = errors.New("semantic index is incomplete or not ready")
)

// Note: Continue using core errors like core.ErrInvalidArgument, core.ErrHandleNotFound,
// core.ErrHandleWrongType, core.ErrInternal, core.ErrInvalidPath where appropriate.
