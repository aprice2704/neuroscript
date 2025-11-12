// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: FEAT: Defines all built-in code snippets for LSP completion. Added ask and whisper. FIX: Removed non-existent InsertTextFormat field. FIX: Expanded func snippet to include optional and returns.
// filename: pkg/nslsp/completion_snippets.go
// nlines: 99

package nslsp

import (
	"strings"

	lsp "github.com/sourcegraph/go-lsp"
)

// A note on snippet syntax:
// ${1:placeholder} is a tab-stop with a placeholder.
// ${0} is the final tab-stop (where the cursor ends up).
// \t is a literal tab character.
const (
	snippetFunc = `
func ${1:MyFunction}(needs ${2:param} optional ${3:opt_param} returns ${4:ret_val}) means
	${0}
endfunc
`
	snippetIf = `
if ${1:condition}
	${0}
endif
`
	snippetIfElse = `
if ${1:condition}
	${2}
else
	${0}
endif
`
	snippetFor = `
for each ${1:item} in ${2:list}
	${0}
endfor
`
	snippetWhile = `
while ${1:condition}
	${0}
endwhile
`
	snippetOnEvent = `
on event ${1:"event_name"} as ${2:evt} do
	${0}
endon
`
	snippetCommand = `
command
	${0}
on error do

endon
endcommand
`
	snippetAsk = `
ask ${1:"model"}, ${2:"prompt"} into ${3:result}
`
	snippetWhisper = `
whisper ${1:"handle"}, ${2:value}
`
)

// getSnippetCompletions returns a list of all built-in code snippets.
func (s *Server) getSnippetCompletions() *lsp.CompletionList {
	items := []lsp.CompletionItem{
		{
			Label:         "func",
			Kind:          lsp.CIKSnippet,
			Detail:        "Function block snippet",
			Documentation: "Creates a new function definition block.",
			InsertText:    strings.TrimSpace(snippetFunc),
		},
		{
			Label:         "if",
			Kind:          lsp.CIKSnippet,
			Detail:        "If block snippet",
			Documentation: "Creates an if...endif block.",
			InsertText:    strings.TrimSpace(snippetIf),
		},
		{
			Label:         "ifelse",
			Kind:          lsp.CIKSnippet,
			Detail:        "If/Else block snippet",
			Documentation: "Creates an if...else...endif block.",
			InsertText:    strings.TrimSpace(snippetIfElse),
		},
		{
			Label:         "for",
			Kind:          lsp.CIKSnippet,
			Detail:        "For...each block snippet",
			Documentation: "Creates a for each...endfor loop.",
			InsertText:    strings.TrimSpace(snippetFor),
		},
		{
			Label:         "while",
			Kind:          lsp.CIKSnippet,
			Detail:        "While block snippet",
			Documentation: "Creates a while...endwhile loop.",
			InsertText:    strings.TrimSpace(snippetWhile),
		},
		{
			Label:         "on event",
			Kind:          lsp.CIKSnippet,
			Detail:        "Event handler snippet",
			Documentation: "Creates an on event...endon block.",
			InsertText:    strings.TrimSpace(snippetOnEvent),
		},
		{
			Label:         "command",
			Kind:          lsp.CIKSnippet,
			Detail:        "Command block snippet",
			Documentation: "Creates a command...endcommand block.",
			InsertText:    strings.TrimSpace(snippetCommand),
		},
		{
			Label:         "ask",
			Kind:          lsp.CIKSnippet,
			Detail:        "Ask snippet",
			Documentation: "Creates an ask...into statement.",
			InsertText:    strings.TrimSpace(snippetAsk),
		},
		{
			Label:         "whisper",
			Kind:          lsp.CIKSnippet,
			Detail:        "Whisper snippet",
			Documentation: "Creates a whisper statement.",
			InsertText:    strings.TrimSpace(snippetWhisper),
		},
	}

	return &lsp.CompletionList{IsIncomplete: false, Items: items}
}
