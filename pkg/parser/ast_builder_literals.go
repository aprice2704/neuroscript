// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 14
// :: description: Added support for DOUBLE_BRACKET_STRING and string interpolation. Fixed node kinds for canon roundtrip.
// :: latestChange: Corrected node kinds to types.KindVariable and types.KindLast to fix roundtrip errors.
// :: filename: pkg/parser/ast_builder_literals.go
// :: serialization: go

package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

var interpSymbols = map[string]string{
	"@nl":  "\n",
	"@cr":  "\r",
	"@tab": "\t",
	"@bt":  "`",
	"@tbt": "```",
	"@sq":  "'",
	"@dq":  "\"",
	"@tsq": "'''",
	"@tdq": "\"\"\"",
}

func buildInterpolatedString(content string, token antlr.Token) ast.Expression {
	var parts []ast.Expression

	remaining := content
	for {
		startIdx := strings.Index(remaining, "{{")
		if startIdx == -1 {
			if len(remaining) > 0 {
				node := &ast.StringLiteralNode{Value: remaining, IsRaw: true}
				parts = append(parts, newNode(node, token, types.KindStringLiteral))
			}
			break
		}
		if startIdx > 0 {
			node := &ast.StringLiteralNode{Value: remaining[:startIdx], IsRaw: true}
			parts = append(parts, newNode(node, token, types.KindStringLiteral))
		}

		remaining = remaining[startIdx+2:]
		endIdx := strings.Index(remaining, "}}")
		if endIdx == -1 {
			node := &ast.StringLiteralNode{Value: "{{" + remaining, IsRaw: true}
			parts = append(parts, newNode(node, token, types.KindStringLiteral))
			break
		}

		inside := strings.TrimSpace(remaining[:endIdx])
		remaining = remaining[endIdx+2:]

		if inside == "" {
			continue
		}

		if strings.HasPrefix(inside, "@") {
			if val, ok := interpSymbols[inside]; ok {
				node := &ast.StringLiteralNode{Value: val, IsRaw: true}
				parts = append(parts, newNode(node, token, types.KindStringLiteral))
			} else {
				node := &ast.StringLiteralNode{Value: "{{" + inside + "}}", IsRaw: true}
				parts = append(parts, newNode(node, token, types.KindStringLiteral))
			}
		} else {
			if strings.ToUpper(inside) == "LAST" {
				// Use correct kind for LAST keyword
				parts = append(parts, newNode(&ast.LastNode{}, token, types.KindLast))
			} else {
				// Use correct kind for variable references
				parts = append(parts, newNode(&ast.VariableNode{Name: inside}, token, types.KindVariable))
			}
		}
	}

	if len(parts) == 0 {
		return newNode(&ast.StringLiteralNode{Value: "", IsRaw: true}, token, types.KindStringLiteral)
	}

	expr := parts[0]

	if _, isStr := expr.(*ast.StringLiteralNode); !isStr {
		emptyStrNode := newNode(&ast.StringLiteralNode{Value: "", IsRaw: true}, token, types.KindStringLiteral)
		expr = newNode(&ast.BinaryOpNode{
			Left:     emptyStrNode,
			Operator: "+",
			Right:    expr,
		}, token, types.KindBinaryOp)
	}

	for i := 1; i < len(parts); i++ {
		binOp := &ast.BinaryOpNode{
			Left:     expr,
			Operator: "+",
			Right:    parts[i],
		}
		expr = newNode(binOp, token, types.KindBinaryOp)
	}

	return expr
}

func (l *neuroScriptListenerImpl) ExitLiteral(ctx *gen.LiteralContext) {
	l.logDebugAST(" >> Exit Literal: %s", ctx.GetText())

	var nodeToPush ast.Expression

	if numNode := ctx.NUMBER_LIT(); numNode != nil {
		token := numNode.GetSymbol()
		val, err := parseNumber(token.GetText())
		if err != nil {
			l.addErrorf(token, "invalid number literal: %v", err)
			errorNode := &ast.ErrorNode{Message: fmt.Sprintf("invalid number: %v", err)}
			nodeToPush = newNode(errorNode, token, types.KindUnknown)
		} else {
			node := &ast.NumberLiteralNode{Value: val}
			nodeToPush = newNode(node, token, types.KindNumberLiteral)
		}
		l.push(nodeToPush)
	} else if strNode := ctx.STRING_LIT(); strNode != nil {
		token := strNode.GetSymbol()
		unescapedString, err := unescapeString(token.GetText())
		if err != nil {
			l.addErrorf(token, "invalid string literal: %v", err)
			errorNode := &ast.ErrorNode{Message: fmt.Sprintf("invalid string: %v", err)}
			nodeToPush = newNode(errorNode, token, types.KindUnknown)
		} else {
			node := &ast.StringLiteralNode{Value: unescapedString, IsRaw: false}
			nodeToPush = newNode(node, token, types.KindStringLiteral)
		}
		l.push(nodeToPush)
	} else if tripleStrNode := ctx.TRIPLE_BACKTICK_STRING(); tripleStrNode != nil {
		token := tripleStrNode.GetSymbol()
		tokenText := token.GetText()
		if len(tokenText) < 6 {
			l.addErrorf(token, "malformed triple-backtick string literal token (too short): %s", tokenText)
			errorNode := &ast.ErrorNode{Message: "malformed raw string"}
			nodeToPush = newNode(errorNode, token, types.KindUnknown)
		} else {
			rawContent := tokenText[3 : len(tokenText)-3]
			nodeToPush = buildInterpolatedString(rawContent, token)
		}
		l.push(nodeToPush)
	} else if doubleBracketNode := ctx.DOUBLE_BRACKET_STRING(); doubleBracketNode != nil {
		token := doubleBracketNode.GetSymbol()
		tokenText := token.GetText()
		if len(tokenText) < 4 {
			l.addErrorf(token, "malformed double-bracket string literal token (too short): %s", tokenText)
			errorNode := &ast.ErrorNode{Message: "malformed raw string"}
			nodeToPush = newNode(errorNode, token, types.KindUnknown)
		} else {
			rawContent := tokenText[2 : len(tokenText)-2]
			nodeToPush = buildInterpolatedString(rawContent, token)
		}
		l.push(nodeToPush)
	} else if tripleSqStrNode := ctx.TRIPLE_SQ_STRING(); tripleSqStrNode != nil {
		token := tripleSqStrNode.GetSymbol()
		tokenText := token.GetText()
		if len(tokenText) < 6 {
			l.addErrorf(token, "malformed triple-single-quote string literal token (too short): %s", tokenText)
			errorNode := &ast.ErrorNode{Message: "malformed raw string"}
			nodeToPush = newNode(errorNode, token, types.KindUnknown)
		} else {
			rawContent := tokenText[3 : len(tokenText)-3]
			node := &ast.StringLiteralNode{Value: rawContent, IsRaw: true}
			nodeToPush = newNode(node, token, types.KindStringLiteral)
		}
		l.push(nodeToPush)
	}

	l.logDebugAST("   << Exit Literal")
}

func (l *neuroScriptListenerImpl) ExitBoolean_literal(ctx *gen.Boolean_literalContext) {
	l.logDebugAST(" >> Exit BooleanLiteral: %s", ctx.GetText())
	var node ast.Expression
	var token antlr.Token
	var val bool

	if ctx.KW_TRUE() != nil {
		token = ctx.KW_TRUE().GetSymbol()
		val = true
		boolNode := &ast.BooleanLiteralNode{Value: val}
		node = newNode(boolNode, token, types.KindBooleanLiteral)
	} else if ctx.KW_FALSE() != nil {
		token = ctx.KW_FALSE().GetSymbol()
		val = false
		boolNode := &ast.BooleanLiteralNode{Value: val}
		node = newNode(boolNode, token, types.KindBooleanLiteral)
	} else {
		token = ctx.GetStart()
		l.addErrorf(token, "malformed boolean literal: missing TRUE or FALSE keyword in rule: %s", ctx.GetText())
		errorNode := &ast.ErrorNode{Message: "malformed boolean"}
		node = newNode(errorNode, token, types.KindUnknown)
	}
	l.push(node)
	l.logDebugAST("   << Exit BooleanLiteral, Pushed Node: %T", node)
}

func (l *neuroScriptListenerImpl) ExitNil_literal(ctx *gen.Nil_literalContext) {
	l.logDebugAST(" >> Exit NilLiteral: %s", ctx.GetText())
	var node ast.Expression
	var token antlr.Token

	if ctx.KW_NIL() != nil {
		token = ctx.KW_NIL().GetSymbol()
		node = newNode(&ast.NilLiteralNode{}, token, types.KindNilLiteral)
	} else {
		token = ctx.GetStart()
		l.addErrorf(token, "malformed nil literal: missing NIL keyword in rule: %s", ctx.GetText())
		errorNode := &ast.ErrorNode{Message: "malformed nil"}
		node = newNode(errorNode, token, types.KindUnknown)
	}
	l.push(node)
	l.logDebugAST("   << Exit NilLiteral, Pushed Node: %T", node)
}

func parseNumber(numStr string) (float64, error) {
	fVal, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number literal: %q", numStr)
	}
	return fVal, nil
}

func unescapeString(quotedStr string) (string, error) {
	if len(quotedStr) < 2 {
		return "", fmt.Errorf("string literal too short: %q", quotedStr)
	}

	if quotedStr[0] == '\'' {
		if quotedStr[len(quotedStr)-1] != '\'' {
			return "", fmt.Errorf("mismatched single quotes in literal: %s", quotedStr)
		}
		return strings.ReplaceAll(quotedStr[1:len(quotedStr)-1], `\'`, `'`), nil
	}

	unquoted, err := strconv.Unquote(quotedStr)
	if err != nil {
		return "", fmt.Errorf("invalid string literal %q: %w", quotedStr, err)
	}
	return unquoted, nil
}
