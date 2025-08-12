// NeuroScript Version: 0.5.2
// File version: 1.0.0
// Purpose: Test-only file to ensure all tool libraries are imported, triggering their init() functions for tool registration before tests are run. This resolves issues where tests would fail because the tool registry was empty.
// filename: pkg/interpreter/z_imports_test.go
// nlines: 10
// risk_rating: LOW
package interpreter

import (
	// This blank import is the key. It forces the Go compiler to include the
	// 'all' tool bundle in the test binary, which in turn triggers the init()
	// functions of every individual tool package, populating the global
	// tool registry.
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all"
)
