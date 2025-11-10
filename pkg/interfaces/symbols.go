// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Defines SymbolProvider using 'any' for all external types to prevent import cycles.
// filename: pkg/interfaces/symbols.go
// nlines: 41

package interfaces

// Note: This package has NO dependencies on other 'neuroscript' packages
// (like 'ast' or 'lang') to prevent import cycles.

// SymbolProviderKey is the key used to retrieve the SymbolProvider
// from the HostContext ServiceRegistry.
const SymbolProviderKey = "SymbolProviderService"

// SymbolProvider defines a contract for a host service to provide
// a foundational, read-only set of symbols to an interpreter.
// These symbols cannot be overridden by loaded scripts.
//
// It uses 'any' for all AST and Language-defined types to prevent
// import cycles. The interpreter implementation is responsible for
// type-asserting these values back to their concrete types.
type SymbolProvider interface {
	// GetProcedure checks if the provider owns a procedure.
	// The returned 'any' is expected to be an *ast.Procedure.
	GetProcedure(name string) (any, bool)

	// GetEventHandlers checks if the provider owns handlers for an event.
	// The returned '[]any' is expected tobe a []*ast.OnEventDecl.
	GetEventHandlers(eventName string) ([]any, bool)

	// GetGlobalConstant checks if the provider owns a global constant.
	// The returned 'any' is expected to be a lang.Value.
	GetGlobalConstant(name string) (any, bool)

	// ListProcedures returns a map of all procedures provided by the host.
	// The map's 'any' values are expected to be *ast.Procedure.
	ListProcedures() map[string]any

	// ListEventHandlers returns a map of all event handlers provided by the host.
	// The map's '[]any' values are expected to be []*ast.OnEventDecl.
	ListEventHandlers() map[string][]any

	// ListGlobalConstants returns a map of all global constants provided by the host.
	// The map's 'any' values are expected to be lang.Value.
	ListGlobalConstants() map[string]any
}
