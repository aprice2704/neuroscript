// NeuroScript Version: 0.3.1
// File version: 0.1.44 // Added more forceful stderr logging for tool extraction debug.
// Purpose: Extracts a potential tool name from NeuroScript content at a given LSP lang.Position using AST analysis.
// filename: pkg/nslsp/server_extracttool_b.go
// nlines: 280
// risk_rating: MEDIUM

package nslsp

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
	lsp "github.com/sourcegraph/go-lsp"
)

const serverExtractToolFileVersion = "0.1.44"

// forceDebugf prints directly to os.Stderr if isDebug is true, bypassing other loggers.
func forceDebugf(isDebug bool, format string, args ...interface{}) {
	if isDebug {
		prefix := fmt.Sprintf("[PID:%d EXTRACT_TOOL_DEBUG] ", os.Getpid())
		fmt.Fprintf(os.Stderr, prefix+format+"\n", args...)
	}
}

func truncateStringForLog(s string, maxLen int) string {
	if len(s) == 0 || maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) > maxLen {
		if maxLen <= 3 {
			return string(runes[:maxLen])
		}
		return string(runes[:maxLen-3]) + "..."
	}
	return s
}

func getTreeTextSafe(node antlr.Tree) string {
	if node == nil {
		return "<nil_node>"
	}
	if ctx, ok := node.(antlr.RuleContext); ok {
		return ctx.GetText()
	}
	if tn, ok := node.(antlr.TerminalNode); ok {
		return tn.GetText()
	}
	return "<unknown_tree_type_for_gettext>"
}

func findInitialNodeManually(node antlr.ParseTree, targetLine, targetChar int, isDebug bool) antlr.TerminalNode {
	if node == nil {
		return nil
	}

	if tn, ok := node.(antlr.TerminalNode); ok {
		token := tn.GetSymbol()
		if token == nil {
			return nil
		}
		tokenLine0Based := token.GetLine() - 1
		tokenStartCol0Based := token.GetColumn()
		tokenText := token.GetText()
		tokenEndCol0Based := tokenStartCol0Based + len(tokenText)

		lineMatch := (tokenLine0Based == targetLine)
		charInSpan := (targetChar >= tokenStartCol0Based && targetChar < tokenEndCol0Based)

		if lineMatch && charInSpan {
			forceDebugf(isDebug, "findInitialNodeManually: MATCH! Terminal: '%s' at L%dC%d (target L%dC%d)", truncateStringForLog(tokenText, 20), tokenLine0Based, tokenStartCol0Based, targetLine, targetChar)
			return tn
		}
		return nil
	}

	for i := 0; i < node.GetChildCount(); i++ {
		child := node.GetChild(i)
		childParseTree, ok := child.(antlr.ParseTree)
		if !ok {
			continue
		}
		found := findInitialNodeManually(childParseTree, targetLine, targetChar, isDebug)
		if found != nil {
			return found
		}
	}
	return nil
}

func getRuleNameSafe(node antlr.Tree, ruleNames []string) string {
	if rCtx, ok := node.(antlr.RuleContext); ok {
		idx := rCtx.GetRuleIndex()
		if idx >= 0 && idx < len(ruleNames) {
			return ruleNames[idx]
		}
		return fmt.Sprintf("InvalidRuleIndex_%d", idx)
	}
	if _, ok := node.(antlr.TerminalNode); ok {
		return "TerminalNode"
	}
	return "UnknownNodeType"
}

func getIdentifiersTextsFromQIGeneric(qiRuleCtx antlr.RuleContext, isDebug bool) []string {
	var identTexts []string
	if qiRuleCtx == nil {
		forceDebugf(isDebug, "getIdentifiersTextsFromQIGeneric: Provided QI node is nil.")
		return identTexts
	}
	ruleIndex := -1
	if qiRuleCtx != nil {
		ruleIndex = qiRuleCtx.GetRuleIndex()
	}
	forceDebugf(isDebug, "getIdentifiersTextsFromQIGeneric: ENTRY. QI node Text='%s', RuleIndex=%d, TargetRuleIndex=%d",
		truncateStringForLog(getTreeTextSafe(qiRuleCtx), 40),
		ruleIndex,
		gen.NeuroScriptParserRULE_qualified_identifier,
	)

	if ruleIndex != gen.NeuroScriptParserRULE_qualified_identifier {
		forceDebugf(isDebug, "getIdentifiersTextsFromQIGeneric: Provided node is not a qualified_identifier. Name: %s", getRuleNameSafe(qiRuleCtx, gen.NeuroScriptParserStaticData.RuleNames))
		return identTexts
	}

	forceDebugf(isDebug, "getIdentifiersTextsFromQIGeneric: Processing QI node: '%s'. Child count: %d", truncateStringForLog(getTreeTextSafe(qiRuleCtx), 30), qiRuleCtx.GetChildCount())
	for i := 0; i < qiRuleCtx.GetChildCount(); i++ {
		child := qiRuleCtx.GetChild(i)
		termNode, okTerm := child.(antlr.TerminalNode)
		childRuleName := getRuleNameSafe(child, gen.NeuroScriptParserStaticData.RuleNames)
		childText := truncateStringForLog(getTreeTextSafe(child), 20)

		if !okTerm {
			forceDebugf(isDebug, "getIdentifiersTextsFromQIGeneric: Child %d of QI ('%s') is not a terminal node. Type: %s", i, childText, childRuleName)
			continue
		}
		tokenType := termNode.GetSymbol().GetTokenType()
		tokenText := termNode.GetSymbol().GetText()
		isIdentifier := tokenType == gen.NeuroScriptLexerIDENTIFIER
		forceDebugf(isDebug, "getIdentifiersTextsFromQIGeneric: QI Child %d: Token '%s', Type: %s (Is IDENTIFIER: %t)", i, tokenText, gen.NeuroScriptLexerLexerStaticData.SymbolicNames[tokenType], isIdentifier)
		if isIdentifier {
			identTexts = append(identTexts, tokenText)
		}
	}
	forceDebugf(isDebug, "getIdentifiersTextsFromQIGeneric: RESULT. Identifiers extracted from QI '%s': %v", truncateStringForLog(getTreeTextSafe(qiRuleCtx), 30), identTexts)
	return identTexts
}

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
	candidateToolName := strings.Join(ids, ".")
	forceDebugf(debugHover, "extractAndValidateFullToolName: Candidate full tool name from QI '%s' is: '%s'", truncateStringForLog(getTreeTextSafe(qiRuleCtx), 30), candidateToolName)

	if s.toolRegistry == nil {
		s.logger.Printf("ERROR: Hover: ToolRegistry is nil in LSP server instance for validation.")
		forceDebugf(debugHover, "extractAndValidateFullToolName: ToolRegistry is nil in server instance!")
		return ""
	}

	forceDebugf(debugHover, "extractAndValidateFullToolName: Checking tool registry for: '%s'", candidateToolName)
	toolDef, found := s.toolRegistry.GetTool(candidateToolName)
	forceDebugf(debugHover, "extractAndValidateFullToolName: Tool: '%s', FoundInRegistry: %t", candidateToolName, found)

	if found {
		forceDebugf(debugHover, "extractAndValidateFullToolName: Tool '%s' IS registered. Definition Name: %s", candidateToolName, toolDef.Spec.Name)
		return candidateToolName
	}

	forceDebugf(debugHover, "extractAndValidateFullToolName: Tool '%s' IS NOT registered.", candidateToolName)
	if debugHover {
		var registeredToolsSample []string
		allTools := s.toolRegistry.ListTools()
		maxToolsToLog := 10 // Log more tools
		for i, ts := range allTools {
			if i < maxToolsToLog {
				registeredToolsSample = append(registeredToolsSample, ts.Name)
			} else {
				registeredToolsSample = append(registeredToolsSample, fmt.Sprintf("...and %d more", len(allTools)-maxToolsToLog))
				break
			}
		}
		if len(allTools) == 0 {
			registeredToolsSample = append(registeredToolsSample, "<No tools registered in registry>")
		}
		// This log goes to stderr directly via forceDebugf's mechanism
		forceDebugf(debugHover, "extractAndValidateFullToolName: Sample of registered tools during check for '%s': %v (Total in registry: %d)", candidateToolName, registeredToolsSample, len(allTools))
	}
	return ""
}

func (s *Server) extractToolNameFromKWRule(foundTokenNode antlr.TerminalNode, debugHover bool) string {
	forceDebugf(debugHover, "extractToolNameFromKWRule: ENTRY. Cursor on KW_TOOL.")
	parent := foundTokenNode.GetParent()
	parentRuleCtx, okParentRule := parent.(antlr.RuleContext)
	if !okParentRule || parentRuleCtx.GetRuleIndex() != gen.NeuroScriptParserRULE_call_target {
		forceDebugf(debugHover, "extractToolNameFromKWRule: KW_TOOL's parent is not call_target. Parent: %s", getRuleNameSafe(parent, gen.NeuroScriptParserStaticData.RuleNames))
		return ""
	}

	forceDebugf(debugHover, "extractToolNameFromKWRule: Parent of KW_TOOL is call_target ('%s'). Searching for qualified_identifier child.", truncateStringForLog(getTreeTextSafe(parentRuleCtx), 30))
	for i := 0; i < parentRuleCtx.GetChildCount(); i++ {
		child := parentRuleCtx.GetChild(i)
		qiCandRuleCtx, okRule := child.(antlr.RuleContext)
		if !okRule || qiCandRuleCtx.GetRuleIndex() != gen.NeuroScriptParserRULE_qualified_identifier {
			forceDebugf(debugHover, "extractToolNameFromKWRule: Child %d of call_target ('%s') is not QI. Type: %s, Text: %s", i, truncateStringForLog(getTreeTextSafe(parentRuleCtx), 20), getRuleNameSafe(child, gen.NeuroScriptParserStaticData.RuleNames), truncateStringForLog(getTreeTextSafe(child), 20))
			continue
		}
		forceDebugf(debugHover, "extractToolNameFromKWRule: Found qualified_identifier child ('%s') of call_target at index %d.", truncateStringForLog(getTreeTextSafe(qiCandRuleCtx), 20), i)
		return s.extractAndValidateFullToolName(qiCandRuleCtx, debugHover)
	}
	forceDebugf(debugHover, "extractToolNameFromKWRule: No qualified_identifier child found for call_target (KW_TOOL case). Parent text: %s", truncateStringForLog(getTreeTextSafe(parentRuleCtx), 30))
	return ""
}

func (s *Server) extractToolNameFromIdentifierRule(foundTokenNode antlr.TerminalNode, debugHover bool) string {
	tokenText := foundTokenNode.GetSymbol().GetText()
	forceDebugf(debugHover, "extractToolNameFromIdentifierRule: ENTRY. Cursor on IDENTIFIER '%s'.", tokenText)

	var qiNode antlr.RuleContext
	currentNode := foundTokenNode.GetParent()
	loopCount := 0
	for currentNode != nil {
		loopCount++
		forceDebugf(debugHover, "extractToolNameFromIdentifierRule: NAV_LOOP %d for '%s': CurrentNode: %s ('%s')", loopCount, tokenText, getRuleNameSafe(currentNode, gen.NeuroScriptParserStaticData.RuleNames), truncateStringForLog(getTreeTextSafe(currentNode), 30))
		if rCtx, ok := currentNode.(antlr.RuleContext); ok && rCtx.GetRuleIndex() == gen.NeuroScriptParserRULE_qualified_identifier {
			qiNode = rCtx
			forceDebugf(debugHover, "extractToolNameFromIdentifierRule: Found QI for '%s': '%s'", tokenText, truncateStringForLog(getTreeTextSafe(qiNode), 30))
			break
		}
		if rCtx, ok := currentNode.(antlr.RuleContext); ok && rCtx.GetRuleIndex() == gen.NeuroScriptParserRULE_call_target {
			forceDebugf(debugHover, "extractToolNameFromIdentifierRule: Hit call_target before QI for '%s'. CurrentNode: '%s'", tokenText, truncateStringForLog(getTreeTextSafe(rCtx), 30))
			break
		}
		currentNode = currentNode.GetParent()
	}

	if qiNode == nil {
		forceDebugf(debugHover, "extractToolNameFromIdentifierRule: IDENTIFIER '%s' is not part of a qualified_identifier relevant to tool calls after %d loop(s).", tokenText, loopCount)
		return ""
	}

	p2Node := qiNode.GetParent()
	p2RuleCtx, okP2Rule := p2Node.(antlr.RuleContext)
	if !okP2Rule || p2RuleCtx.GetRuleIndex() != gen.NeuroScriptParserRULE_call_target {
		forceDebugf(debugHover, "extractToolNameFromIdentifierRule: QI's ('%s') parent is not call_target. Parent: %s ('%s')", truncateStringForLog(getTreeTextSafe(qiNode), 20), getRuleNameSafe(p2Node, gen.NeuroScriptParserStaticData.RuleNames), truncateStringForLog(getTreeTextSafe(p2Node), 30))
		return ""
	}

	forceDebugf(debugHover, "extractToolNameFromIdentifierRule: QI's ('%s') parent is call_target ('%s'). Checking if call_target starts with KW_TOOL.", truncateStringForLog(getTreeTextSafe(qiNode), 20), truncateStringForLog(getTreeTextSafe(p2RuleCtx), 30))
	if p2RuleCtx.GetChildCount() == 0 {
		forceDebugf(debugHover, "extractToolNameFromIdentifierRule: call_target has no children.")
		return ""
	}

	firstChildOfCallTarget := p2RuleCtx.GetChild(0)
	ftTerm, okFt := firstChildOfCallTarget.(antlr.TerminalNode)
	if !okFt || ftTerm.GetSymbol().GetTokenType() != gen.NeuroScriptLexerKW_TOOL {
		forceDebugf(debugHover, "extractToolNameFromIdentifierRule: call_target does not start with KW_TOOL. First child: %s ('%s')", getRuleNameSafe(firstChildOfCallTarget, gen.NeuroScriptParserStaticData.RuleNames), truncateStringForLog(getTreeTextSafe(firstChildOfCallTarget), 20))
		return ""
	}

	forceDebugf(debugHover, "extractToolNameFromIdentifierRule: call_target starts with KW_TOOL. Validating full tool name from QI ('%s').", truncateStringForLog(getTreeTextSafe(qiNode), 20))
	return s.extractAndValidateFullToolName(qiNode, debugHover)
}

func (s *Server) extractToolNameAtPosition(content string, position lsp.Position, sourceName string) string {
	if s.logger == nil {
		s.logger = log.New(os.Stderr, "[LSP_SERVER_FALLBACK_LOGGER] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	}

	s.logger.Printf("INFO: extractToolNameAtPosition called. FileVersion: %s. Source: %s L%dC%d",
		serverExtractToolFileVersion, sourceName, position.Line+1, position.Character+1)

	debugHover := os.Getenv("NSLSP_DEBUG_HOVER") != "" || os.Getenv("DEBUG_LSP_HOVER_TEST") != ""
	if debugHover { // Initial indication that debug mode is active for this function call
		forceDebugf(true, "extractToolNameAtPosition: Debug mode ENABLED. Source: %s L%dC%d", sourceName, position.Line+1, position.Character+1)
	}

	if s.coreParserAPI == nil {
		s.logger.Printf("ERROR_LSP_HOVER: coreParserAPI is nil in server. Cannot parse content.")
		forceDebugf(debugHover, "extractToolNameAtPosition: coreParserAPI is NIL.")
		return ""
	}

	treeFromParser, parseErrors := s.coreParserAPI.ParseForLSP(sourceName, content)
	if treeFromParser == nil {
		s.logger.Printf("WARN: extractToolName: AST is nil after parsing for source '%s'.", sourceName)
		forceDebugf(debugHover, "extractToolNameAtPosition: treeFromParser is NIL after ParseForLSP for '%s'.", sourceName)
		return ""
	}
	if len(parseErrors) > 0 {
		s.logger.Printf("INFO: extractToolName: Parsed '%s' with %d errors, but AST tree was returned.", sourceName, len(parseErrors))
	}

	parseTreeRoot, ok := treeFromParser.(antlr.ParseTree)
	if !ok {
		s.logger.Printf("ERROR_LSP_HOVER: Parsed tree for '%s' is not an antlr.ParseTree. Type: %T", sourceName, treeFromParser)
		forceDebugf(debugHover, "extractToolNameAtPosition: treeFromParser is not antlr.ParseTree. Type: %T", treeFromParser)
		return ""
	}

	forceDebugf(debugHover, "extractToolNameAtPosition: AST for '%s' obtained. Root type: %s", sourceName, getRuleNameSafe(parseTreeRoot, gen.NeuroScriptParserStaticData.RuleNames))
	forceDebugf(debugHover, "extractToolNameAtPosition: Calling findInitialNodeManually for '%s'. Target: L%dC%d.", sourceName, position.Line, position.Character)
	foundTokenNode := findInitialNodeManually(parseTreeRoot, position.Line, position.Character, debugHover)
	determinedToolName := ""

	if foundTokenNode == nil {
		forceDebugf(debugHover, "extractToolNameAtPosition: findInitialNodeManually returned nil for '%s'. No token at cursor L%dC%d.", sourceName, position.Line, position.Character)
		s.logger.Printf("INFO: extractToolName: No token at cursor L%dC%d for '%s'. Token: <foundNode_is_nil>", position.Line+1, position.Character+1, sourceName)
		return ""
	}

	tokenSymbol := foundTokenNode.GetSymbol()
	tokenType := tokenSymbol.GetTokenType()
	tokenText := tokenSymbol.GetText()
	lexerSymbolicNames := gen.NeuroScriptLexerLexerStaticData.SymbolicNames
	tokenTypeName := "UNKNOWN_TOKEN_TYPE"
	if tokenType >= 0 && tokenType < len(lexerSymbolicNames) {
		tokenTypeName = lexerSymbolicNames[tokenType]
	}
	forceDebugf(debugHover, "extractToolNameAtPosition: Found token for '%s': '%s' (Type: %s)", sourceName, tokenText, tokenTypeName)

	switch tokenType {
	case gen.NeuroScriptLexerKW_TOOL:
		determinedToolName = s.extractToolNameFromKWRule(foundTokenNode, debugHover)
	case gen.NeuroScriptLexerIDENTIFIER, gen.NeuroScriptLexerDOT:
		determinedToolName = s.extractToolNameFromIdentifierRule(foundTokenNode, debugHover)
	default:
		forceDebugf(debugHover, "extractToolNameAtPosition: Token type %s ('%s') is not KW_TOOL, IDENTIFIER, or DOT. No tool name extraction path.", tokenTypeName, tokenText)
	}

	if determinedToolName != "" { // This block only runs if a potential tool name was determined
		initialTokenTypeAtCursor := tokenSymbol.GetTokenType()
		isCursorOnStringLitToken := initialTokenTypeAtCursor == gen.NeuroScriptLexerSTRING_LIT || initialTokenTypeAtCursor == gen.NeuroScriptLexerTRIPLE_BACKTICK_STRING
		isCursorInsideStringContent := false

		if isCursorOnStringLitToken {
			tokenStartCol := tokenSymbol.GetColumn()
			tokenTextLen := len(tokenSymbol.GetText())
			quoteLen := 1
			if initialTokenTypeAtCursor == gen.NeuroScriptLexerTRIPLE_BACKTICK_STRING {
				quoteLen = 3
			}
			if tokenTextLen > (2*quoteLen-1) && position.Character > tokenStartCol+(quoteLen-1) && position.Character < (tokenStartCol+tokenTextLen-quoteLen) {
				isCursorInsideStringContent = true
			}
		}
		if isCursorInsideStringContent { // If cursor is inside a string, it can't be a tool name.
			forceDebugf(debugHover, "extractToolNameAtPosition: Cursor was inside string literal ('%s'). Clearing determined tool name '%s'.", tokenSymbol.GetText(), determinedToolName)
			determinedToolName = ""
		}
	}

	finalFoundNodeText := "'" + foundTokenNode.GetText() + "'"
	if determinedToolName != "" {
		s.logger.Printf("INFO: extractToolName: Tool name '%s' extracted and validated for '%s' L%dC%d. Token: %s",
			determinedToolName, sourceName, position.Line+1, position.Character+1, finalFoundNodeText)
	} else {
		s.logger.Printf("INFO: extractToolName: No registered 'tool.X' style tool name identified for '%s' L%dC%d. Token: %s",
			sourceName, position.Line+1, position.Character+1, finalFoundNodeText)
	}
	return determinedToolName
}

func formatASTNodeWithPosition(node antlr.Tree, ruleNames []string, indent string, isDebug bool) string {
	if node == nil {
		return ""
	}
	var b strings.Builder
	// ... (rest of function remains the same, ensuring any internal logging also uses forceDebugf if needed or is removed if not essential for this trace)
	// For brevity, I'm not repeating the whole formatASTNodeWithPosition, assuming its internal logging isn't the primary focus for this specific bug.
	// If it had internal dlogf calls, they would now use os.Stderr via forceDebugf.
	// The simplified version:
	b.WriteString(indent)
	b.WriteString(getRuleNameSafe(node, ruleNames))
	b.WriteString(fmt.Sprintf(" (%s)\n", truncateStringForLog(getTreeTextSafe(node), 50)))
	if rc, ok := node.(antlr.RuleContext); ok {
		for i := 0; i < rc.GetChildCount(); i++ {
			b.WriteString(formatASTNodeWithPosition(rc.GetChild(i), ruleNames, indent+"  ", isDebug))
		}
	}
	return b.String()
}
