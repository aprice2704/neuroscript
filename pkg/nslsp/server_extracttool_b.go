// NeuroScript Version: 0.3.1
// File version: 52
// Purpose: Made tool name lookup case-insensitive. FIX: Plumbed a real logger through all hover-related functions. FIX: Corrected typo (o -> 0).
// filename: pkg/nslsp/server_extracttool_b.go
// nlines: 153
// risk_rating: HIGH

package nslsp

import (
	"log"
	"os"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/types"
	lsp "github.com/sourcegraph/go-lsp"
)

const serverExtractToolFileVersion = "0.1.50"

func (s *Server) extractAndValidateFullToolName(qiRuleCtx antlr.RuleContext, log loggerFunc) string {
	if qiRuleCtx == nil {
		log("extractAndValidateFullToolName: Called with nil qiRuleCtx.")
		return ""
	}
	log("extractAndValidateFullToolName: ENTRY. Validating QI: '%s'", truncateStringForLog(getTreeTextSafe(qiRuleCtx), 30))

	ids := getIdentifiersTextsFromQIGeneric(qiRuleCtx, log)

	if len(ids) == 0 {
		log("extractAndValidateFullToolName: QI '%s' yielded no IDENTIFIER children.", truncateStringForLog(getTreeTextSafe(qiRuleCtx), 30))
		return ""
	}

	nameWithoutPrefix := strings.Join(ids, ".")
	candidateToolName := "tool." + nameWithoutPrefix
	log("extractAndValidateFullToolName: Candidate partial name from QI '%s' is: '%s'. Prepended prefix to form full name: '%s'", truncateStringForLog(getTreeTextSafe(qiRuleCtx), 30), nameWithoutPrefix, candidateToolName)

	if s.interpreter == nil || s.interpreter.ToolRegistry() == nil {
		s.logger.Printf("ERROR: Hover: Interpreter or its ToolRegistry is nil in LSP server instance.")
		log("extractAndValidateFullToolName: Interpreter or ToolRegistry is nil in server instance!")
		return ""
	}

	lookupName := types.FullName(strings.ToLower(candidateToolName))
	log("extractAndValidateFullToolName: Checking tool registry for (case-insensitive): '%s'", lookupName)
	_, found := s.interpreter.ToolRegistry().GetTool(lookupName)
	if !found {
		if s.externalTools != nil {
			_, found = s.externalTools.GetTool(lookupName)
		}
	}
	log("extractAndValidateFullToolName: Tool: '%s', FoundInRegistry: %t", lookupName, found)

	if found {
		return candidateToolName
	}

	return ""
}

func (s *Server) extractToolNameFromKWRule(foundTokenNode antlr.TerminalNode, log loggerFunc) string {
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
		return s.extractAndValidateFullToolName(qiCandRuleCtx, log)
	}
	return ""
}

func (s *Server) extractToolNameFromIdentifierRule(foundTokenNode antlr.TerminalNode, log loggerFunc) string {
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

	return s.extractAndValidateFullToolName(qiNode, log)
}

func (s *Server) extractToolNameAtPosition(content string, position lsp.Position, sourceName string) string {
	if s.logger == nil {
		s.logger = log.New(os.Stderr, "[LSP_SERVER_FALLBACK_LOGGER] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	}

	debugHover := os.Getenv("NSLSP_DEBUG_HOVER") != "" || os.Getenv("DEBUG_LSP_HOVER_TEST") != ""
	var log loggerFunc = noOpLogger
	if debugHover {
		log = func(format string, args ...interface{}) {
			s.logger.Printf("[HOVER_TRACE] "+format, args...)
		}
	}

	if s.coreParserAPI == nil {
		s.logger.Printf("ERROR_LSP_HOVER: coreParserAPI is nil in server.")
		return ""
	}

	treeFromParser, _ := s.coreParserAPI.ParseForLSP(sourceName, content)
	if treeFromParser == nil {
		log("extractToolNameAtPosition: Parser returned a nil tree. Cannot find token.")
		s.logger.Printf("WARN: extractToolName: AST is nil after parsing for source '%s'.", sourceName)
		return ""
	}

	parseTreeRoot, ok := treeFromParser.(antlr.ParseTree)
	if !ok {
		s.logger.Printf("ERROR_LSP_HOVER: Parsed tree is not an antlr.ParseTree.")
		return ""
	}

	log("extractToolNameAtPosition: Starting token search at L%d:C%d", position.Line, position.Character)
	foundTokenNode := findInitialNodeManually(parseTreeRoot, position.Line, position.Character, log)
	if foundTokenNode == nil {
		log("extractToolNameAtPosition: findInitialNodeManually did not find a token at the specified position.")
		return ""
	}

	log("extractToolNameAtPosition: Found token '%s' of type %s.", foundTokenNode.GetText(), gen.NeuroScriptParserStaticData.SymbolicNames[foundTokenNode.GetSymbol().GetTokenType()])
	tokenType := foundTokenNode.GetSymbol().GetTokenType()
	switch tokenType {
	case gen.NeuroScriptLexerKW_TOOL:
		return s.extractToolNameFromKWRule(foundTokenNode, log)
	case gen.NeuroScriptLexerIDENTIFIER, gen.NeuroScriptLexerDOT:
		return s.extractToolNameFromIdentifierRule(foundTokenNode, log)
	default:
		return ""
	}
}
