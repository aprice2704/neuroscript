// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 6
// :: description: FEAT: Added hover support for user-defined workspace procedures.
// :: latestChange: Updated hover links to display short filenames while linking to full URIs.
// :: filename: pkg/nslsp/hover.go
// :: serialization: go
package nslsp

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
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
	uriStr := string(params.TextDocument.URI)

	// --- Step 1: Check for a tool name (e.g., tool.FS.Read) ---
	toolName := s.extractToolNameAtPosition(content, params.Position, uriStr)
	if toolName != "" {
		lookupName := types.FullName(strings.ToLower(toolName))

		var impl tool.ToolImplementation
		var foundTool bool
		if s.interpreter.ToolRegistry() != nil {
			impl, foundTool = s.interpreter.ToolRegistry().GetTool(lookupName)
		}
		if !foundTool && s.externalTools != nil {
			impl, foundTool = s.externalTools.GetTool(lookupName)
		}

		if foundTool {
			return s.formatToolHover(toolName, impl), nil
		}
	}

	// --- Step 2: Check for a built-in function name (e.g., sin) ---
	builtInName := s.extractBuiltInNameAtPosition(content, params.Position, uriStr)
	if builtInName != "" {
		return s.getBuiltInHover(builtInName), nil
	}

	// --- Step 3: Check for a structural keyword (e.g., func) ---
	keywordTokenType := s.extractKeywordAtPosition(content, params.Position, uriStr)
	if keywordTokenType != 0 {
		return s.getKeywordHover(keywordTokenType), nil
	}

	// --- Step 4: Check for a workspace-defined procedure (e.g. MyFunc) ---
	procName := s.extractProcedureNameAtPosition(content, params.Position, uriStr)
	if procName != "" {
		if info, found := s.symbolManager.GetSymbolInfo(procName); found {
			var hoverContent strings.Builder
			hoverContent.WriteString(fmt.Sprintf("```neuroscript\nfunc %s%s\n```\n", procName, info.Signature))
			hoverContent.WriteString("---\n")

			// Clean up the link
			uriString := string(info.URI)
			parsedURI, err := url.Parse(uriString)
			displayName := uriString
			if err == nil {
				// Use the base filename for display (e.g. "startup_utils.ns")
				displayName = filepath.Base(parsedURI.Path)
			}

			// Markdown Link: [display_text](link_target)
			hoverContent.WriteString(fmt.Sprintf("Defined in: [%s](%s)", displayName, uriString))

			return &lsp.Hover{
				Contents: []lsp.MarkedString{
					{
						Language: "markdown",
						Value:    hoverContent.String(),
					},
				},
			}, nil
		}
	}

	// --- Step 5: Nothing found ---
	return nil, nil
}

// formatToolHover builds the hover response for a NeuroScript tool.
func (s *Server) formatToolHover(toolName string, impl tool.ToolImplementation) *lsp.Hover {
	spec := impl.Spec
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

	return &lsp.Hover{
		Contents: []lsp.MarkedString{
			{
				Language: "markdown",
				Value:    hoverContent.String(),
			},
		},
	}
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
