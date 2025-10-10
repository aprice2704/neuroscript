// NeuroScript Version: 0.8.0
// File version: 86
// Purpose: FIX: Replaces direct access to the removed `turnCtx` field with calls to `GetTurnContext()`.
// filename: pkg/interpreter/interpreter_exec.go
// nlines: 268
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// Execute runs the command blocks from a given AST program.
func (i *Interpreter) Execute(program *ast.Program) (lang.Value, error) {
	// --- MORE DEBUGGING ---
	ctx := i.GetTurnContext()
	if ctx != nil {
		sid, _ := ctx.Value(aeiou.SessionIDKey).(string)
		turn, _ := ctx.Value(aeiou.TurnIndexKey).(int)
		fmt.Fprintf(os.Stderr, "[DEBUG Execute START] Interp ID: %s, SID: %q, Turn: %d\n", i.id, sid, turn)
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG Execute START] Interp ID: %s, Context is NIL\n", i.id)
	}
	// --- END DEBUGGING ---

	if program == nil {
		return &lang.NilValue{}, nil
	}
	i.state.commands = program.Commands
	return i.ExecuteCommands()
}

func (i *Interpreter) executeSteps(steps []ast.Step, isInHandler bool, activeError *lang.RuntimeError) (lang.Value, bool, bool, error) {
	return i.recExecuteSteps(steps, isInHandler, activeError, 0)
}

func getStepSubjectForLogging(step ast.Step) string {
	// ... (content unchanged)
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
		// --- NEW DEBUGGING ---
		ctx := i.GetTurnContext()
		if ctx != nil {
			sid, _ := ctx.Value(aeiou.SessionIDKey).(string)
			turn, _ := ctx.Value(aeiou.TurnIndexKey).(int)
			fmt.Fprintf(os.Stderr, "[DEBUG recExecuteSteps LOOP] Interp ID: %s, Step: %s, SID: %q, Turn: %d\n", i.id, step.Type, sid, turn)
		} else {
			fmt.Fprintf(os.Stderr, "[DEBUG recExecuteSteps LOOP] Interp ID: %s, Step: %s, Context is NIL\n", i.id, step.Type)
		}
		// --- END NEW DEBUGGING ---

		var stepResult lang.Value
		var stepErr error
		stepTypeLower := strings.ToLower(step.Type)

		// --- MORE DEBUGGING ---
		// fmt.Fprintf(os.Stderr, "[EXEC STEP] ID: %s | Step: %s | Subject: %s\n", i.id, stepTypeLower, getStepSubjectForLogging(step))
		// // --- END DEBUGGING ---

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
				stepResult, stepErr = i.evaluate.Expression(step.Values[0])
			} else {
				stepResult = &lang.NilValue{}
			}
		default:
			errMsg := fmt.Sprintf("unknown step type '%s'", step.Type)
			stepErr = lang.NewRuntimeError(lang.ErrorCodeUnknownKeyword, errMsg, lang.ErrUnknownKeyword).WithPosition(step.GetPos())
		}

		if stepErr != nil {
			rtErr := ensureRuntimeError(stepErr, step.GetPos(), stepTypeLower)
			// fmt.Fprintf(os.Stderr, "\n>>> [EXEC DEBUG %s] Error in step '%s': %v\n", i.id, stepTypeLower, rtErr)
			// fmt.Fprintf(os.Stderr, ">>> [EXEC DEBUG %s] isInHandler flag is: %t\n", i.id, isInHandler)

			if errors.Is(rtErr.Unwrap(), lang.ErrBreak) || errors.Is(rtErr.Unwrap(), lang.ErrContinue) {
				return nil, false, wasCleared, rtErr
			}

			if isInHandler {
				// fmt.Fprintf(os.Stderr, ">>> [EXEC DEBUG %s] In handler, propagating error up to caller (e.g., EmitEvent).\n\n", i.id)
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
