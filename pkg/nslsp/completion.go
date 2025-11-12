// NeuroScript Version: 0.7.0
// File version: 10
// Purpose: Implements textDocument/completion. FEAT: Merges built-in function completions and snippet completions.
// filename: pkg/nslsp/completion.go
// nlines: 206

package nslsp

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/tool"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (s *Server) handleTextDocumentCompletion(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.CompletionParams
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}

	content, found := s.documentManager.Get(params.TextDocument.URI)
	if !found {
		return nil, nil
	}

	lines := strings.Split(content, "\n")
	if params.Position.Line < 0 || params.Position.Line >= len(lines) {
		return nil, nil
	}
	line := lines[params.Position.Line]

	if params.Position.Character > len(line) {
		params.Position.Character = len(line)
	}
	// Get the text on the line from the beginning to the cursor.
	linePrefix := line[:params.Position.Character]

	// --- Tool Completion Logic ---
	// Isolate the specific `tool.` expression the user is typing.
	lastToolIndex := strings.LastIndex(linePrefix, "tool.")
	if lastToolIndex != -1 {
		// User is typing a tool, so *only* return tool completions.
		toolPrefix := strings.TrimSpace(linePrefix[lastToolIndex:])
		parts := strings.Split(toolPrefix, ".")

		// Case 1: User typed `tool.` and expects a list of groups.
		if len(parts) == 2 && parts[0] == "tool" && parts[1] == "" {
			return s.getToolGroupCompletions(), nil
		}

		// Case 2: User typed `tool.group.` and expects a list of tool names.
		if len(parts) >= 3 && parts[0] == "tool" && parts[len(parts)-1] == "" {
			return s.getToolNameCompletions(toolPrefix), nil
		}
		// If they are in the middle of typing a tool name, don't show other suggestions.
		return nil, nil
	}

	// --- General Completion Logic ---
	// If we are not in a tool expression, return snippets and built-in functions.
	snippets := s.getSnippetCompletions()
	builtIns := s.getBuiltInCompletions()

	// Merge the two lists
	snippets.Items = append(snippets.Items, builtIns.Items...)
	return snippets, nil
}

func (s *Server) getToolGroupCompletions() *lsp.CompletionList {
	groupSet := make(map[string]string) // lowercase -> original case

	collectGroups := func(impls []tool.ToolImplementation) {
		for _, impl := range impls {
			if impl.Spec.Group != "" {
				groupName := string(impl.Spec.Group)
				lowerGroupName := strings.ToLower(groupName)
				if _, exists := groupSet[lowerGroupName]; !exists {
					groupSet[lowerGroupName] = groupName
				}
			}
		}
	}

	if s.interpreter != nil && s.interpreter.ToolRegistry() != nil {
		collectGroups(s.interpreter.ToolRegistry().ListTools())
	}
	if s.externalTools != nil {
		collectGroups(s.externalTools.ListTools())
	}

	items := make([]lsp.CompletionItem, 0, len(groupSet))
	for _, originalCaseGroup := range groupSet {
		items = append(items, lsp.CompletionItem{
			Label:  originalCaseGroup,
			Kind:   lsp.CIKModule,
			Detail: "Tool Group",
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })

	return &lsp.CompletionList{IsIncomplete: false, Items: items}
}

func (s *Server) getToolNameCompletions(prefix string) *lsp.CompletionList {
	parts := strings.Split(prefix, ".")
	if len(parts) < 3 || parts[0] != "tool" {
		return nil
	}
	// The user is typing, so match the group name case-insensitively.
	// For "tool.fdm.core.", parts is ["tool", "fdm", "core", ""]. We join parts[1:3] -> "fdm.core"
	groupPrefix := strings.ToLower(strings.Join(parts[1:len(parts)-1], "."))

	items := make([]lsp.CompletionItem, 0)
	seenTools := make(map[string]struct{})

	collectTools := func(impls []tool.ToolImplementation) {
		for _, impl := range impls {
			if strings.ToLower(string(impl.Spec.Group)) == groupPrefix {
				toolName := string(impl.Spec.Name)
				if _, exists := seenTools[toolName]; !exists {
					items = append(items, createCompletionItemFromSpec(impl.Spec))
					seenTools[toolName] = struct{}{}
				}
			}
		}
	}

	if s.interpreter != nil && s.interpreter.ToolRegistry() != nil {
		collectTools(s.interpreter.ToolRegistry().ListTools())
	}
	if s.externalTools != nil {
		collectTools(s.externalTools.ListTools())
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })

	return &lsp.CompletionList{IsIncomplete: false, Items: items}
}

func createCompletionItemFromSpec(spec tool.ToolSpec) lsp.CompletionItem {
	var params []string
	for _, arg := range spec.Args {
		typeString := string(arg.Type)
		if !arg.Required {
			typeString += "?"
		}
		params = append(params, fmt.Sprintf("%s: %s", arg.Name, typeString))
	}
	paramSignature := strings.Join(params, ", ")
	fullSignatureForDetail := fmt.Sprintf("(%s) -> %s", paramSignature, spec.ReturnType)

	// Build a markdown documentation string to help the client with syntax highlighting.
	var docBuilder strings.Builder
	docBuilder.WriteString(fmt.Sprintf("```neuroscript\n(tool) %s%s\n```\n", spec.Name, fullSignatureForDetail))
	if spec.Description != "" {
		docBuilder.WriteString("---\n")
		docBuilder.WriteString(spec.Description)
	}

	return lsp.CompletionItem{
		Label:         string(spec.Name),
		Kind:          lsp.CIKFunction,
		Detail:        fullSignatureForDetail,
		Documentation: docBuilder.String(),
	}
}
