// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides the textDocument/hover LSP handler and all related formatting logic, separated from the main handlers file.
// filename: pkg/nslsp/hover.go
// nlines: 125
// risk_rating: MEDIUM

package nslsp

import (
	"context"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// handleTextDocumentHover generates and returns hover information for a tool.
func (s *Server) handleTextDocumentHover(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.TextDocumentPositionParams
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}
	content, found := s.documentManager.Get(params.TextDocument.URI)
	if !found {
		return nil, nil
	}

	toolName := s.extractToolNameAtPosition(content, params.Position, string(params.TextDocument.URI))
	if toolName == "" {
		return nil, nil
	}
	lookupName := types.FullName(strings.ToLower(toolName))

	var impl tool.ToolImplementation
	var foundTool bool
	if s.interpreter.ToolRegistry() != nil {
		impl, foundTool = s.interpreter.ToolRegistry().GetTool(lookupName)
	}
	if !foundTool && s.externalTools != nil {
		impl, foundTool = s.externalTools.GetTool(lookupName)
	}

	if !foundTool {
		return nil, nil
	}
	spec := impl.Spec

	// Build the hover content using Markdown for better presentation.
	var hoverContent strings.Builder
	signature := buildSignatureString(toolName, spec)
	hoverContent.WriteString(fmt.Sprintf("```neuroscript\n%s\n```\n", signature))
	hoverContent.WriteString("---\n")
	if spec.Description != "" {
		hoverContent.WriteString(spec.Description + "\n\n")
	}
	if impl.RequiresTrust {
		hoverContent.WriteString("**Requires Trust:** `true`\n\n")
	}
	if len(impl.RequiredCaps) > 0 {
		hoverContent.WriteString("**Required Capabilities:**\n")
		hoverContent.WriteString(formatCapsForHover(impl.RequiredCaps))
		hoverContent.WriteString("\n")
	}
	if len(spec.Args) > 0 {
		hoverContent.WriteString("**Parameters:**\n")
		hoverContent.WriteString(formatParamsForHover(spec.Args))
		hoverContent.WriteString("\n")
	}
	hoverContent.WriteString(fmt.Sprintf("**Returns:** (`%s`)", spec.ReturnType))
	if spec.ReturnHelp != "" {
		hoverContent.WriteString(" " + spec.ReturnHelp)
	}
	hoverContent.WriteString("\n")

	// THE FIX IS HERE: The sourcegraph/go-lsp library uses a slice of MarkedString.
	// We create one entry with our fully formatted Markdown content.
	return &lsp.Hover{
		Contents: []lsp.MarkedString{
			{
				Language: "markdown",
				Value:    hoverContent.String(),
			},
		},
	}, nil
}

// buildSignatureString creates a compact signature for a tool.
func buildSignatureString(fullName string, spec tool.ToolSpec) string {
	var params []string
	for _, arg := range spec.Args {
		part := fmt.Sprintf("%s: %s", arg.Name, arg.Type)
		if !arg.Required {
			part += "?"
		}
		params = append(params, part)
	}
	return fmt.Sprintf("(tool) %s(%s) -> %s", fullName, strings.Join(params, ", "), spec.ReturnType)
}

// formatParamsForHover creates a Markdown list of tool parameters.
func formatParamsForHover(params []tool.ArgSpec) string {
	var mdBuilder strings.Builder
	for _, p := range params {
		requiredStr := ""
		if !p.Required {
			requiredStr = "*(optional)* "
		}
		mdBuilder.WriteString(fmt.Sprintf("* `%s` (`%s`): %s%s\n", p.Name, p.Type, requiredStr, p.Description))
	}
	return mdBuilder.String()
}

// formatCapsForHover creates a Markdown list of required capabilities.
func formatCapsForHover(caps []capability.Capability) string {
	var mdBuilder strings.Builder
	for _, c := range caps {
		scopePart := ""
		if len(c.Scopes) > 0 {
			scopePart = ":" + strings.Join(c.Scopes, ",")
		}
		mdBuilder.WriteString(fmt.Sprintf("* `%s:%s%s`\n", c.Resource, strings.Join(c.Verbs, ","), scopePart))
	}
	return mdBuilder.String()
}
