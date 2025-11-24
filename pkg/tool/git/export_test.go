// NeuroScript Version: 1
// File version: 1
// Purpose: Exposes internal git package variables to the git_test package for integration testing.
// filename: pkg/tool/git/export_test.go
package git

// GitToolsToRegister exports the internal tool definition list so
// external tests (package git_test) can register them manually
// in a test interpreter.
var GitToolsToRegister = gitToolsToRegister
