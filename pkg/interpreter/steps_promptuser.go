// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Re-plumbed to use the external 'eval' package for prompt expression evaluation.
// filename: pkg/interpreter/steps_promptuser.go
// nlines: 45
// risk_rating: MEDIUM

package interpreter

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/eval"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// executePromptUser handles the "promptuser" step.
func (i *Interpreter) executePromptUser(step ast.Step) (lang.Value, error) {
	if step.PromptUserStmt == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "promptuser step is missing its PromptUserStmt node", nil).WithPosition(step.GetPos())
	}
	node := step.PromptUserStmt

	promptVal, err := eval.Expression(i, node.PromptExpr)
	if err != nil {
		return nil, lang.WrapErrorWithPosition(err, node.PromptExpr.GetPos(), "evaluating prompt for promptuser")
	}
	prompt, _ := lang.ToString(promptVal)

	response, err := i.PromptUser(prompt)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeIOFailed, "failed to get user input", err).WithPosition(step.GetPos())
	}

	responseVal, wrapErr := lang.Wrap(response)
	if wrapErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to wrap user response", wrapErr).WithPosition(step.GetPos())
	}

	if node.IntoTarget != nil {
		if err := i.setSingleLValue(node.IntoTarget, responseVal); err != nil {
			return nil, err
		}
	}

	return responseVal, nil
}
