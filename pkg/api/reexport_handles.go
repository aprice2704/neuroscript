// NeuroScript Version: 1
// File version: 2
// Purpose: Re-exports types and functions for the opaque object handle system (HandleValue, HandleRegistry). Added TypeHandle constant string value to fix compiler error.
// filename: pkg/api/reexport_handles.go
// nlines: 26

package api

import (
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// Handle-related types and interfaces for the public API.
type (
	// HandleValue is the opaque reference to a host object.
	HandleValue = interfaces.HandleValue

	// HandleRegistry provides the public interface for creating, retrieving,
	// and deleting host object handles within an interpreter runtime.
	HandleRegistry = interfaces.HandleRegistry
)

// Handle-related constants
const (
	// TypeHandle is the string representation of the Handle value type in NeuroScript (e.g., 'typeof h').
	// FIX: Defining the string value here to resolve the UndeclaredImportedName error.
	TypeHandle = "handle"
)
