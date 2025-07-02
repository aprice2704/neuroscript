// NeuroScript Version: 0.3.1
// File version: 0.1.4
// Purpose: Defines core logic for NS syntax analysis tool. Returns a list of error maps.
// filename: pkg/tool/syntax/tools_syntax_analyzer.go
// nlines: 80 // Approximate
// risk_rating: LOW

package syntax

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	// "encoding/json" // No longer needed as we return a map/slice directly
)

// GrammarVersion is assumed to be a package-level variable (e.g., in interpreter.go or utils.go),
// injected via ldflags at build time, holding the NeuroScript grammar version.
// var GrammarVersion string // This is defined elsewhere in package core

const (
	analyzerMaxErrorsToReportInternal	= 20
	analyzerSourceNameInternal		= "nsSyntaxAnalysisToolInput"
)

// SyntaxAnalysisReport struct is removed as the tool now returns a []map[string]interface{} directly.

// AnalyzeNSSyntaxInternal is the core logic for the syntax analysis tool.
// It's called by the wrapper function defined in tooldefs_syntax.go.
// It now returns a slice of maps (each map representing a StructuredSyntaxError),
// or an empty slice if no errors. The error return is for unexpected internal issues.
func AnalyzeNSSyntaxInternal(interpreter *Interpreter, nsScriptContent string) (interface{}, error) {
	if interpreter == nil {
		return nil, fmt.Errorf("interpreter cannot be nil: %w", ErrInvalidArgument)	//
	}
	logger := interpreter.Logger()
	if logger == nil {
		logger = &adapters.NewNoOpLogger{}	//
	}

	parserAPI := NewParserAPI(logger)	//
	// ParseForLSP returns all structured errors found.
	_, structuredErrors := parserAPI.ParseForLSP(analyzerSourceNameInternal, nsScriptContent)

	if len(structuredErrors) == 0 {
		return []map[string]interface{}{}, nil	// Return empty list for no errors
	}

	// Determine how many errors to report (cap at maxErrorsToReportInternal)
	numToReport := len(structuredErrors)
	if numToReport > analyzerMaxErrorsToReportInternal {
		numToReport = analyzerMaxErrorsToReportInternal
	}

	errorList := make([]map[string]interface{}, numToReport)

	for i := 0; i < numToReport; i++ {
		sErr := structuredErrors[i]
		errorList[i] = map[string]interface{}{
			"Line":			sErr.Line,
			"Column":		sErr.Column,
			"Msg":			sErr.Msg,
			"OffendingSymbol":	sErr.OffendingSymbol,
			"SourceName":		sErr.SourceName,	// This will be analyzerSourceNameInternal
		}
	}

	// No separate "totalErrorsFound" or "grammarVersion" in this direct return type.
	// The calling NeuroScript can get the count from len(errorList).
	// Grammar version is an ambient property of the interpreter/system.
	return errorList, nil
}