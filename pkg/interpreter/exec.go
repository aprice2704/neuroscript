// NeuroScript Version: 0.8.0
// File version: 91
// Purpose: This file now compiles correctly as Interpreter now satisfies the eval.Runtime interface.
// filename: pkg/interpreter/exec.go
// nlines: 268
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/eval"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// Execute runs the command blocks from a given AST program in a sandboxed environment.
// It returns a NumberValue of 0 on success, or the error code on failure.
func (i *Interpreter) Execute(program *ast.Program) (lang.Value, error) {
	if program == nil {
		return lang.NumberValue{Value: 0}, nil
	}
	// Fork the interpreter to create a sandboxed environment for command execution.
	cmdInterpreter := i.fork()
	cmdInterpreter.state.commands = program.Commands
	_, err := cmdInterpreter.executeCommands()

	if err != nil {
		var rtErr *lang.RuntimeError
		if errors.As(err, &rtErr) {
			// Return the error code as the value, plus the error itself.
			return lang.NumberValue{Value: float64(rtErr.Code)}, err
		}
		// Return a generic error code for non-runtime errors.
		return lang.NumberValue{Value: 1}, err
	}
	// Per language design, success is signaled with a 0 return.
	return lang.NumberValue{Value: 0}, nil
}

func (i *Interpreter) executeSteps(steps []ast.Step, isInHandler bool, activeError *lang.RuntimeError) (lang.Value, bool, bool, error) {
	return i.recExecuteSteps(steps, isInHandler, activeError, 0)
}

// ... getStepSubjectForLogging implementation remains the same ...
func getStepSubjectForLogging(step ast.Step) string {
	switch strings.ToLower(step.Type) {
	case "set", "assign":
		if len(step.LValues) > 0 {
			var parts []string
			for _, lval := range step.LValues {
				if lval != nil {
					parts = append(parts, lval.String())
				}
			}
			return strings.Join(parts, ", ")
		}
	case "call":
		if step.Call != nil {
			return step.Call.Target.String()
		}
	case "ask":
		if step.AskStmt != nil && step.AskStmt.PromptExpr != nil {
			return step.AskStmt.PromptExpr.String()
		}
		return "<ask>"
	case "promptuser":
		if step.PromptUserStmt != nil && step.PromptUserStmt.PromptExpr != nil {
			return step.PromptUserStmt.PromptExpr.String()
		}
		return "<promptuser>"

	case "for_each", "for":
		return fmt.Sprintf("loopVar: %s, collection: %s", step.LoopVarName, step.Collection.String())
	case "must", "mustbe":
		if step.Cond != nil {
			return fmt.Sprintf("must %s", step.Cond.String())
		}
		if step.Call != nil {
			return fmt.Sprintf("mustbe %s", step.Call.String())
		}
		return "must last"
	case "return":
		if len(step.Values) > 0 {
			var parts []string
			for _, v := range step.Values {
				if v != nil {
					parts = append(parts, v.String())
				}
			}
			return fmt.Sprintf("returning %d values: %s", len(step.Values), strings.Join(parts, ", "))
		}
		return "return (nil)"
	case "emit":
		if len(step.Values) > 0 {
			var parts []string
			for _, v := range step.Values {
				if v != nil {
					parts = append(parts, v.String())
				}
			}
			return strings.Join(parts, ", ")
		}
		return "emit (empty)"
	}
	return "<no specific subject>"
}

func (i *Interpreter) recExecuteSteps(steps []ast.Step, isInHandler bool, activeError *lang.RuntimeError, depth int) (finalResult lang.Value, wasReturn bool, wasCleared bool, finalError error) {
	finalResult = &lang.NilValue{}

	for _, step := range steps {
		var stepResult lang.Value
		var stepErr error
		stepTypeLower := strings.ToLower(step.Type)

		switch stepTypeLower {
		case "set", "assign":
			stepResult, stepErr = i.executeSet(step)
		case "call":
			stepResult, stepErr = i.executeCall(step)
		case "return":
			if isInHandler {
				stepErr = lang.NewRuntimeError(lang.ErrorCodeReturnViolation, "'return' is not permitted inside an on_error block", lang.ErrReturnViolation).WithPosition(step.GetPos())
			} else {
				var returnValue lang.Value
				returnValue, wasReturn, stepErr = i.executeReturn(step)
				if stepErr == nil && wasReturn {
					finalResult = returnValue
					i.lastCallResult = finalResult
					return finalResult, true, false, nil
				}
			}
		case "emit":
			stepResult, stepErr = i.executeEmit(step)
		case "whisper":
			stepResult, stepErr = i.executeWhisper(step)
		case "if":
			var ifReturned, ifCleared bool
			var ifBlockResult lang.Value
			ifBlockResult, ifReturned, ifCleared, stepErr = i.executeIf(step, isInHandler, activeError)
			if stepErr == nil {
				stepResult = ifBlockResult
				if ifReturned {
					i.lastCallResult = stepResult
					return stepResult, true, false, nil
				}
				if ifCleared {
					wasCleared = true
					activeError = nil
				}
			}
		case "while":
			var whileReturned, whileCleared bool
			var whileBlockResult lang.Value
			whileBlockResult, whileReturned, whileCleared, stepErr = i.executeWhile(step, isInHandler, activeError)
			if stepErr == nil {
				stepResult = whileBlockResult
				if whileReturned {
					i.lastCallResult = stepResult
					return stepResult, true, wasCleared, nil
				}
				if whileCleared {
					wasCleared = true
					activeError = nil
				}
			}
		case "for", "for_each":
			var forReturned, forCleared bool
			var forResult lang.Value
			forResult, forReturned, forCleared, stepErr = i.executeFor(step, isInHandler, activeError)
			if stepErr == nil {
				stepResult = forResult
				if forReturned {
					return forResult, true, false, nil
				}
				if forCleared {
					wasCleared = true
					activeError = nil
				}
			}
		case "must", "mustbe":
			stepResult, stepErr = i.executeMust(step)
		case "fail":
			stepErr = i.executeFail(step)
		case "on_error":
			continue
		case "clear_error":
			var clearedNow bool
			clearedNow, stepErr = i.executeClearError(step, isInHandler)
			if stepErr == nil && clearedNow {
				wasCleared = true
				activeError = nil
			}
		case "ask":
			stepResult, stepErr = i.executeAsk(step)
		case "promptuser":
			stepResult, stepErr = i.executePromptUser(step)
		case "break":
			stepErr = i.executeBreak(step)
		case "continue":
			stepErr = i.executeContinue(step)
		case "expression_statement":
			if len(step.Values) > 0 {
				stepResult, stepErr = eval.Expression(i, step.Values[0])
			} else {
				stepResult = &lang.NilValue{}
			}
		default:
			errMsg := fmt.Sprintf("unknown step type '%s'", step.Type)
			stepErr = lang.NewRuntimeError(lang.ErrorCodeUnknownKeyword, errMsg, lang.ErrUnknownKeyword).WithPosition(step.GetPos())
		}

		if stepErr != nil {
			rtErr := ensureRuntimeError(stepErr, step.GetPos(), stepTypeLower)

			if errors.Is(rtErr.Unwrap(), lang.ErrBreak) || errors.Is(rtErr.Unwrap(), lang.ErrContinue) {
				return nil, false, wasCleared, rtErr
			}

			if isInHandler {
				return nil, false, false, rtErr
			}

			if len(i.state.errorHandlerStack) > 0 {
				handlerBlock := i.state.errorHandlerStack[len(i.state.errorHandlerStack)-1]
				handlerToExecute := handlerBlock[0]
				i.SetVariable("system_error_message", lang.StringValue{Value: rtErr.Message})
				_, _, handlerCleared, handlerErr := i.executeSteps(handlerToExecute.Body, true, rtErr)
				if handlerErr != nil {
					return nil, false, false, ensureRuntimeError(handlerErr, handlerToExecute.GetPos(), "ON_ERROR_HANDLER")
				}
				if handlerCleared {
					stepErr = nil
					continue
				} else {
					return nil, false, false, rtErr
				}
			} else {
				return nil, false, wasCleared, rtErr
			}
		}

		if stepErr == nil {
			if shouldUpdateLastResult(stepTypeLower) {
				finalResult = stepResult
				i.lastCallResult = stepResult
			}
		}
	}

	return finalResult, false, wasCleared, nil
}
