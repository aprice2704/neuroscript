// NeuroScript Version: 0.3.1
// File version: 1.0.0
// Purpose: Ensures all tool bundles are linked into the test binary for this package, fixing 'tool not found' errors.
// filename: pkg/nslsp/z_imports_test.go
// nlines: 10
// risk_rating: LOW

package nslsp

import (
	// This blank import is the key. It forces the Go compiler to include the
	// 'all' tool bundle in the test binary for the nslsp package. This in turn
	// triggers the init() functions of every individual tool package, populating
	// the global tool registration lists before any tests in this package are run.
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all"
)
