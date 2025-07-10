// NeuroScript Version: 0.3.1
// File version: 46
// Purpose: Refactored to use helper functions from ast_helpers.go for improved modularity.
// filename: pkg/nslsp/server_extracttool_b.go
// nlines: 140
// risk_rating: MEDIUM

package nslsp

import (
	"log"
	"os"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
	"github.com/aprice2704/neuroscript/pkg/types"
	lsp "github.com/sourcegraph/go-lsp"
)

const serverExtractToolFileVersion = "0.1.46"

func (s *Server) extractAndValidateFullToolName(qiRuleCtx antlr.RuleContext, debugHover bool) string {
	if qiRuleCtx == nil {
		forceDebugf(debugHover, "extractAndValidateFullToolName: Called with nil qiRuleCtx.")
		return ""
	}
	forceDebugf(debugHover, "extractAndValidateFullToolName: ENTRY. Validating QI: '%s'", truncateStringForLog(getTreeTextSafe(qiRuleCtx), 30))

	ids := getIdentifiersTextsFromQIGeneric(qiRuleCtx, debugHover)

	if len(ids) == 0 {
		forceDebugf(debugHover, "extractAndValidateFullToolName: QI '%s' yielded no IDENTIFIER children.", truncateStringForLog(getTreeTextSafe(qiRuleCtx), 30))
		return ""
	}

	nameWithoutPrefix := strings.Join(ids, ".")
	candidateToolName := "tool." + nameWithoutPrefix
	forceDebugf(debugHover, "extractAndValidateFullToolName: Candidate partial name from QI '%s' is: '%s'. Prepended prefix to form full name: '%s'", truncateStringForLog(getTreeTextSafe(qiRuleCtx), 30), nameWithoutPrefix, candidateToolName)

	if s.toolRegistry == nil {
		s.logger.Printf("ERROR: Hover: ToolRegistry is nil in LSP server instance for validation.")
		forceDebugf(debugHover, "extractAndValidateFullToolName: ToolRegistry is nil in server instance!")
		return ""
	}

	forceDebugf(debugHover, "extractAndValidateFullToolName: Checking tool registry for: '%s'", candidateToolName)
	_, found := s.toolRegistry.GetTool(types.FullName(candidateToolName))
	forceDebugf(debugHover, "extractAndValidateFullToolName: Tool: '%s', FoundInRegistry: %t", candidateToolName, found)

	if found {
		return candidateToolName
	}

	return ""
}

func (s *Server) extractToolNameFromKWRule(foundTokenNode antlr.TerminalNode, debugHover bool) string {
	parent := foundTokenNode.GetParent()
	parentRuleCtx, okParentRule := parent.(antlr.RuleContext)
	if !okParentRule || parentRuleCtx.GetRuleIndex() != gen.NeuroScriptParserRULE_call_target {
		return ""
	}

	for i := 0; i < parentRuleCtx.GetChildCount(); i++ {
		child := parentRuleCtx.GetChild(i)
		qiCandRuleCtx, okRule := child.(antlr.RuleContext)
		if !okRule || qiCandRuleCtx.GetRuleIndex() != gen.NeuroScriptParserRULE_qualified_identifier {
			continue
		}
		return s.extractAndValidateFullToolName(qiCandRuleCtx, debugHover)
	}
	return ""
}

func (s *Server) extractToolNameFromIdentifierRule(foundTokenNode antlr.TerminalNode, debugHover bool) string {
	var qiNode antlr.RuleContext
	currentNode := foundTokenNode.GetParent()
	for currentNode != nil {
		if rCtx, ok := currentNode.(antlr.RuleContext); ok && rCtx.GetRuleIndex() == gen.NeuroScriptParserRULE_qualified_identifier {
			qiNode = rCtx
			break
		}
		if rCtx, ok := currentNode.(antlr.RuleContext); ok && rCtx.GetRuleIndex() == gen.NeuroScriptParserRULE_call_target {
			break
		}
		currentNode = currentNode.GetParent()
	}

	if qiNode == nil {
		return ""
	}

	p2Node := qiNode.GetParent()
	p2RuleCtx, okP2Rule := p2Node.(antlr.RuleContext)
	if !okP2Rule || p2RuleCtx.GetRuleIndex() != gen.NeuroScriptParserRULE_call_target {
		return ""
	}

	if p2RuleCtx.GetChildCount() == 0 {
		return ""
	}

	firstChildOfCallTarget := p2RuleCtx.GetChild(0)
	ftTerm, okFt := firstChildOfCallTarget.(antlr.TerminalNode)
	if !okFt || ftTerm.GetSymbol().GetTokenType() != gen.NeuroScriptLexerKW_TOOL {
		return ""
	}

	return s.extractAndValidateFullToolName(qiNode, debugHover)
}

func (s *Server) extractToolNameAtPosition(content string, position lsp.Position, sourceName string) string {
	if s.logger == nil {
		s.logger = log.New(os.Stderr, "[LSP_SERVER_FALLBACK_LOGGER] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	}

	debugHover := os.Getenv("NSLSP_DEBUG_HOVER") != "" || os.Getenv("DEBUG_LSP_HOVER_TEST") != ""

	if s.coreParserAPI == nil {
		s.logger.Printf("ERROR_LSP_HOVER: coreParserAPI is nil in server.")
		return ""
	}

	treeFromParser, _ := s.coreParserAPI.ParseForLSP(sourceName, content)
	if treeFromParser == nil {
		s.logger.Printf("WARN: extractToolName: AST is nil after parsing for source '%s'.", sourceName)
		return ""
	}

	parseTreeRoot, ok := treeFromParser.(antlr.ParseTree)
	if !ok {
		s.logger.Printf("ERROR_LSP_HOVER: Parsed tree is not an antlr.ParseTree.")
		return ""
	}

	foundTokenNode := findInitialNodeManually(parseTreeRoot, position.Line, position.Character, debugHover)
	if foundTokenNode == nil {
		return ""
	}

	tokenType := foundTokenNode.GetSymbol().GetTokenType()
	switch tokenType {
	case gen.NeuroScriptLexerKW_TOOL:
		return s.extractToolNameFromKWRule(foundTokenNode, debugHover)
	case gen.NeuroScriptLexerIDENTIFIER, gen.NeuroScriptLexerDOT:
		return s.extractToolNameFromIdentifierRule(foundTokenNode, debugHover)
	default:
		return ""
	}
}
