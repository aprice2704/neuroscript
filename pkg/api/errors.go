// filename: pkg/api/errors.go
// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Re-exports key sentinel errors from the lang package for the public API.
// nlines: 38
// risk_rating: LOW

package api

import "github.com/aprice2704/neuroscript/pkg/lang"

// Re-exported sentinel errors allow consumers of the API to programmatically
// check for specific, stable error conditions without importing internal packages.
//
// Example:
//
//	_, err := api.Parse(source)
//	if errors.Is(err, api.ErrSyntax) {
//	    // Handle the specific error of a parsing failure.
//	}
var (
	// --- Parsing Errors ---
	ErrSyntax = lang.ErrSyntax

	// --- Execution & Procedure Errors ---
	ErrProcedureNotFound   = lang.ErrProcedureNotFound
	ErrArgumentMismatch    = lang.ErrArgumentMismatch
	ErrMustConditionFailed = lang.ErrMustConditionFailed
	ErrFailStatement       = lang.ErrFailStatement
	ErrDivisionByZero      = lang.ErrDivisionByZero

	// --- Policy & Security Errors ---
	ErrToolNotAllowed    = lang.ErrToolNotAllowed
	ErrToolDenied        = lang.ErrToolDenied
	ErrSecurityViolation = lang.ErrSecurityViolation
	ErrPathViolation     = lang.ErrPathViolation

	// --- Tool & Provider Errors ---
	ErrToolNotFound     = lang.ErrToolNotFound
	ErrProviderNotFound = lang.ErrProviderNotFound

	// --- Filesystem Errors ---
	ErrFileNotFound = lang.ErrFileNotFound
)
