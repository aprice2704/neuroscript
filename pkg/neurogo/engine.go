// filename: pkg/neurogo/engine.go
// File version: 1.1
// Corrected undefined types to use interfaces package.
package neurogo

import (
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
)

type Engine struct {
	interp *interpreter.Interpreter
	llm    interfaces.LLMClient
	logger interfaces.Logger
	// … other pluggables
}
