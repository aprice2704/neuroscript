// filename: pkg/api/errors.go
// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Re-exports additional key sentinel errors for handles, validation, and policy.
// nlines: 50
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

// --- ErrorCode Type ---
type ErrorCode lang.ErrorCode

var (
	// --- Parsing & Validation Errors ---
	ErrSyntax       = lang.ErrSyntax
	ErrInvalidInput = lang.ErrInvalidInput
	ErrInvalidUTF8  = lang.ErrInvalidUTF8

	// --- Execution & Procedure Errors ---
	ErrProcedureNotFound   = lang.ErrProcedureNotFound
	ErrArgumentMismatch    = lang.ErrArgumentMismatch
	ErrMustConditionFailed = lang.ErrMustConditionFailed
	ErrFailStatement       = lang.ErrFailStatement
	ErrDivisionByZero      = lang.ErrDivisionByZero
	ErrProcedureExists     = lang.ErrProcedureExists

	// --- Policy & Security Errors ---
	ErrToolNotAllowed    = lang.ErrToolNotAllowed
	ErrToolDenied        = lang.ErrToolDenied
	ErrSecurityViolation = lang.ErrSecurityViolation
	ErrPathViolation     = lang.ErrPathViolation
	ErrPolicyViolation   = lang.ErrPolicyViolation

	// --- Tool & Provider Errors ---
	ErrToolNotFound     = lang.ErrToolNotFound
	ErrProviderNotFound = lang.ErrProviderNotFound

	// --- Handle Errors ---
	ErrHandleInvalid   = lang.ErrHandleInvalid
	ErrHandleNotFound  = lang.ErrHandleNotFound
	ErrHandleWrongType = lang.ErrHandleWrongType

	// --- Filesystem Errors ---
	ErrFileNotFound = lang.ErrFileNotFound
)

const SecurityBase ErrorCode = 99900

const (
	_                             ErrorCode = SecurityBase + iota // 99900 reserved (placeholder)
	ErrorCodeAttackPossible                                       // 99901
	ErrorCodeAttackProbable                                       // 99902
	ErrorCodeAttackCertain                                        // 99903
	ErrorCodeSubsystemCompromised                                 // 99904
	ErrorCodeSubsystemQuarantined                                 // 99905
	ErrorCodeEscapePossible                                       // 99906
	ErrorCodeEscapeProbable                                       // 99907
	ErrorCodeEscapeCertain                                        // 99908
	ErrorCodeSecretDecryption                                     // 99909
)
