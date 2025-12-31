// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 82
// :: description: Enhanced ensureRuntimeError to lookup and display procedure definition locations in stack traces.
// :: latestChange: Added definition location lookup in stack trace generation.
// :: filename: pkg/interpreter/exec_more.go
// :: serialization: go

package interpreter

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/eval"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (i *Interpreter) executeCall(step ast.Step) (lang.Value, error) {
	if step.Call == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "call step is missing call expression", nil).WithPosition(step.GetPos())
	}
	// The entire call logic is now handled by the evaluator.
	return eval.Expression(i, step.Call)
}

func (i *Interpreter) executeWhisper(step ast.Step) (lang.Value, error) {
	if step.WhisperStmt == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "whisper step is missing whisper statement details", nil).WithPosition(step.GetPos())
	}

	handleVal, err := eval.Expression(i, step.WhisperStmt.Handle)
	if err != nil {
		return nil, err
	}
	dataVal, err := eval.Expression(i, step.WhisperStmt.Value)
	if err != nil {
		return nil, err
	}

	if i.hostContext != nil && i.hostContext.WhisperFunc != nil {
		i.hostContext.WhisperFunc(handleVal, dataVal)
	} else {
		// This default behavior should likely be removed in favor of a mandatory handler.
		i.Logger().Debug("Whisper statement executed without a custom handler", "handle", handleVal.String(), "data", dataVal.String())
	}
	return dataVal, nil
}

func (i *Interpreter) executeBlock(blockValue interface{}, parentPos *types.Position, blockType string, isInHandler bool, activeError *lang.RuntimeError, depth int) (result lang.Value, wasReturn bool, wasCleared bool, err error) {
	steps, ok := blockValue.([]ast.Step)
	if !ok {
		if blockValue == nil {
			return &lang.NilValue{}, false, false, nil
		}
		errMsg := fmt.Sprintf("internal error: invalid block format for %s - expected []Step, got %T", blockType, blockValue)
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, errMsg, lang.ErrInternal).WithPosition(parentPos)
	}
	return i.recExecuteSteps(steps, isInHandler, activeError, depth)
}

func shouldUpdateLastResult(stepTypeLower string) bool {
	switch stepTypeLower {
	case "set", "assign", "emit", "ask", "promptuser", "call", "expression_statement", "must", "mustbe", "whisper":
		return true
	default:
		return false
	}
}

// ensureRuntimeError wraps an error into a proper RuntimeError if it isn't one already.
// It also enriches the error message with position information and a stack trace.
func (i *Interpreter) ensureRuntimeError(err error, pos *types.Position, context string) *lang.RuntimeError {
	var rtErr *lang.RuntimeError
	if asRtErr, ok := err.(*lang.RuntimeError); ok {
		rtErr = asRtErr
	} else {
		rtErr = lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("internal error during %s", context), err)
	}

	// 1. Ensure Position
	if rtErr.Position == nil && pos != nil {
		rtErr = rtErr.WithPosition(pos)
	}

	// 2. Format Message with Context & Stack Trace
	// Avoid re-appending if we catch the error higher up the stack.
	if !strings.Contains(rtErr.Message, "Stack Trace:") {
		var sb strings.Builder
		sb.WriteString(rtErr.Message)

		// Append readable position if available
		if rtErr.Position != nil {
			sb.WriteString(fmt.Sprintf("\n  at %s", rtErr.Position.String()))
		} else if pos != nil {
			sb.WriteString(fmt.Sprintf("\n  at %s", pos.String()))
		}

		// Append Stack Trace
		// Stack frames are stored [root -> leaf]. We print reverse [leaf -> root].
		stackFrames := i.state.stackFrames
		if len(stackFrames) > 0 {
			sb.WriteString("\nStack Trace:")
			for idx := len(stackFrames) - 1; idx >= 0; idx-- {
				name := stackFrames[idx]
				sb.WriteString(fmt.Sprintf("\n  %s", name))

				// Lookup definition location for procedures
				// Note: We use the shared knownProcedures map.
				if proc, ok := i.state.knownProcedures[name]; ok && proc != nil {
					if p := proc.GetPos(); p != nil {
						sb.WriteString(fmt.Sprintf(" (defined at %s)", p.String()))
					}
				}
			}
		} else if i.state.currentProcName != "" {
			// Fallback for single procedure execution or contexts (like events/commands)
			// where we manually set currentProcName but didn't push a stack frame.
			sb.WriteString(fmt.Sprintf("\nStack Trace:\n  %s", i.state.currentProcName))
		}

		rtErr.Message = sb.String()
	}

	return rtErr
}
