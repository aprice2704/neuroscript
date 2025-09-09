// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Re-exports the public API of the internal json_lite package.
// filename: pkg/api/shape/shape.go
// nlines: 55
// risk_rating: LOW

package shape

import (
	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// --- Types ---

// Path is the parsed, executable representation of a path-lite string.
type Path = json_lite.Path

// PathSegment represents one component of a parsed path, either a map key or list index.
type PathSegment = json_lite.PathSegment

// Shape is a compiled representation of a validation rule for JSON-like data.
type Shape = json_lite.Shape

// FieldSpec defines the validation rules for a single field within a Shape.
type FieldSpec = json_lite.FieldSpec

// SelectOptions provides options for the Select function.
type SelectOptions = json_lite.SelectOptions

// ValidateOptions provides options for the Validate method.
type ValidateOptions = json_lite.ValidateOptions

// --- Functions ---

// ParsePath compiles a path-lite string (e.g., "a.b[0].c") into a reusable Path structure.
var ParsePath = json_lite.ParsePath

// ParseShape compiles a shape definition map into a reusable Shape structure.
var ParseShape = json_lite.ParseShape

// Select retrieves a value from a nested data structure using a pre-parsed Path.
var Select = json_lite.Select

// --- Errors ---

var (
	ErrInvalidPath                  = json_lite.ErrInvalidPath
	ErrNestingDepthExceeded         = json_lite.ErrNestingDepthExceeded
	ErrMapKeyNotFound               = json_lite.ErrMapKeyNotFound
	ErrListIndexOutOfBounds         = json_lite.ErrListIndexOutOfBounds
	ErrListInvalidIndexType         = json_lite.ErrListInvalidIndexType
	ErrCannotAccessType             = json_lite.ErrCannotAccessType
	ErrCollectionIsNil              = json_lite.ErrCollectionIsNil
	ErrInvalidArgument              = json_lite.ErrInvalidArgument
	ErrValidationTypeMismatch       = json_lite.ErrValidationTypeMismatch
	ErrValidationRequiredArgMissing = json_lite.ErrValidationRequiredArgMissing
	ErrValidationFailed             = json_lite.ErrValidationFailed
)

var Unwrap = lang.UnwrapForShapeValidation
