// neurogo/engine.go
package neurogo

import "github.com/aprice2704/neuroscript/pkg/interpreter"

type Engine struct {
	interp *interpreter.Interpreter
	llm    llm.Client
	logger logging.Logger
	// … other pluggables
}
