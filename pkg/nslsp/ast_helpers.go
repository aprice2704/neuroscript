// NeuroScript Version: 0.3.1
// File version: 1
// Purpose: Provides helper functions for AST traversal and manipulation for the LSP server.
// filename: pkg/nslsp/ast_helpers.go
// nlines: 155
// risk_rating: LOW

package nslsp

import (
	"fmt"
	"os"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
)

// forceDebugf prints directly to os.Stderr if isDebug is true, bypassing other loggers.
// This is used for high-priority debug messages that should not be missed.
func forceDebugf(isDebug bool, format string, args ...interface{}) {
	if isDebug {
		prefix := fmt.Sprintf("[PID:%d EXTRACT_TOOL_DEBUG] ", os.Getpid())
		fmt.Fprintf(os.Stderr, prefix+format+"\n", args...)
	}
}

// truncateStringForLog shortens a string to a max length for cleaner logging.
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

// getTreeTextSafe safely gets the text from any ANTLR parse tree node.
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

// findInitialNodeManually walks the parse tree to find the terminal node at a specific line and character.
// This is the primary method for linking a cursor position to a specific token in the source code.
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

// getRuleNameSafe safely gets the rule name from a parse tree node.
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

// getIdentifiersTextsFromQIGeneric extracts the parts of a qualified identifier (e.g., "FS", "List" from "FS.List").
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

	if ruleIndex != gen.NeuroScriptParserRULE_qualified_identifier {
		forceDebugf(isDebug, "getIdentifiersTextsFromQIGeneric: Provided node is not a qualified_identifier. Name: %s", getRuleNameSafe(qiRuleCtx, gen.NeuroScriptParserStaticData.RuleNames))
		return identTexts
	}

	for i := 0; i < qiRuleCtx.GetChildCount(); i++ {
		child := qiRuleCtx.GetChild(i)
		termNode, okTerm := child.(antlr.TerminalNode)
		if !okTerm {
			continue
		}
		tokenType := termNode.GetSymbol().GetTokenType()
		tokenText := termNode.GetSymbol().GetText()
		if tokenType == gen.NeuroScriptLexerIDENTIFIER {
			identTexts = append(identTexts, tokenText)
		}
	}
	forceDebugf(isDebug, "getIdentifiersTextsFromQIGeneric: RESULT. Identifiers extracted from QI '%s': %v", truncateStringForLog(getTreeTextSafe(qiRuleCtx), 30), identTexts)
	return identTexts
}
