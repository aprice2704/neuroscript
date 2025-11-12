// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: FEAT: Defines hover documentation for all major structural keywords.
// filename: pkg/nslsp/keyword_hover.go
// nlines: 104

package nslsp

import (
	"fmt"
	"strings"

	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	lsp "github.com/sourcegraph/go-lsp"
)

// KeywordDocInfo holds the documentation for a top-level keyword.
type KeywordDocInfo struct {
	Signature   string
	Description string
}

// KeywordDocs is the canonical map of all keywords and their help info.
var KeywordDocs = map[int]KeywordDocInfo{
	gen.NeuroScriptLexerKW_FUNC: {
		Signature:   "func <name>(needs ... optional ... returns ...) means ... endfunc",
		Description: "Defines a reusable procedure (function) with a list of required, optional, and return parameters.",
	},
	gen.NeuroScriptLexerKW_COMMAND: {
		Signature:   "command ... endcommand",
		Description: "Defines a top-level, executable block of code. This is the main entry point for a script that performs actions.",
	},
	gen.NeuroScriptLexerKW_ON: {
		Signature:   "on event <name_expr> as <var> do ... endon\non error do ... endon",
		Description: "Defines an event handler. It can be used to subscribe to a named event or to define a local error handler for a function or command block.",
	},
	gen.NeuroScriptLexerKW_IF: {
		Signature:   "if <condition> ... else ... endif",
		Description: "A conditional block. The `else` block is optional.",
	},
	gen.NeuroScriptLexerKW_FOR: {
		Signature:   "for each <var> in <list_or_map> ... endfor",
		Description: "Iterates over each item in a list or map. For maps, the loop variable will be the *key*.",
	},
	gen.NeuroScriptLexerKW_WHILE: {
		Signature:   "while <condition> ... endwhile",
		Description: "A loop that executes as long as the condition evaluates to true.",
	},
	gen.NeuroScriptLexerKW_ASK: {
		Signature:   "ask <model_expr>, <prompt_expr> with <options_map> into <result_var>",
		Description: "Sends a prompt to an AI model and stores the response in a variable. `with` and `into` are optional.",
	},
	gen.NeuroScriptLexerKW_WHISPER: {
		Signature:   "whisper <handle_expr>, <value_expr>",
		Description: "Sends a value to the host application's 'noetic' layer. This is used for out-of-band communication, like updating state or UI elements.",
	},
	gen.NeuroScriptLexerKW_SET: {
		Signature:   "set <variable> = <expression>",
		Description: "Assigns the result of an expression to a variable.",
	},
	gen.NeuroScriptLexerKW_CALL: {
		Signature:   "call <procedure_name>(...)\ncall tool.<group>.<name>(...)",
		Description: "Executes a procedure or a tool.",
	},
	gen.NeuroScriptLexerKW_RETURN: {
		Signature:   "return <value1>, <value2>, ...",
		Description: "Exits the current function and returns one or more values. `return` with no values is also valid.",
	},
	gen.NeuroScriptLexerKW_EMIT: {
		Signature:   "emit <event_expr>",
		Description: "Emits a named event to be caught by an `on event` handler.",
	},
	gen.NeuroScriptLexerKW_FAIL: {
		Signature:   "fail <error_message>",
		Description: "Stops execution and raises a runtime error with the given message. This will trigger an `on error` handler if one is defined in scope.",
	},
}

// getKeywordHover creates hover info for a structural keyword.
func (s *Server) getKeywordHover(tokenType int) *lsp.Hover {
	info, found := KeywordDocs[tokenType]
	if !found {
		return nil
	}

	var hoverContent strings.Builder
	hoverContent.WriteString(fmt.Sprintf("```neuroscript\n(keyword) %s\n```\n", info.Signature))
	hoverContent.WriteString("---\n")
	hoverContent.WriteString(info.Description)

	return &lsp.Hover{
		Contents: []lsp.MarkedString{
			{
				Language: "markdown",
				Value:    hoverContent.String(),
			},
		},
	}
}
