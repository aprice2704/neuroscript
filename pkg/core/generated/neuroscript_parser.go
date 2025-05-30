// Code generated from NeuroScript.g4 by ANTLR 4.13.2. DO NOT EDIT.

package core // NeuroScript
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type NeuroScriptParser struct {
	*antlr.BaseParser
}

var NeuroScriptParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func neuroscriptParserInit() {
	staticData := &NeuroScriptParserStaticData
	staticData.LiteralNames = []string{
		"", "'call'", "'func'", "'needs'", "'optional'", "'returns'", "'means'",
		"'endfunc'", "'set'", "'return'", "'emit'", "'if'", "'else'", "'endif'",
		"'while'", "'endwhile'", "'for'", "'each'", "'in'", "'endfor'", "'on_error'",
		"'endon'", "'clear_error'", "'must'", "'mustbe'", "'fail'", "'no'",
		"'some'", "'tool'", "'last'", "'eval'", "'true'", "'false'", "'and'",
		"'or'", "'not'", "'ln'", "'log'", "'sin'", "'cos'", "'tan'", "'asin'",
		"'acos'", "'atan'", "'ask'", "'into'", "'break'", "'continue'", "'nil'",
		"'typeof'", "", "", "", "'='", "'+'", "'-'", "'*'", "'/'", "'%'", "'**'",
		"'&'", "'|'", "'^'", "'~'", "'('", "')'", "','", "'['", "']'", "'{'",
		"'}'", "':'", "'.'", "'{{'", "'}}'", "'=='", "'!='", "'>'", "'<'", "'>='",
		"'<='",
	}
	staticData.SymbolicNames = []string{
		"", "KW_CALL", "KW_FUNC", "KW_NEEDS", "KW_OPTIONAL", "KW_RETURNS", "KW_MEANS",
		"KW_ENDFUNC", "KW_SET", "KW_RETURN", "KW_EMIT", "KW_IF", "KW_ELSE",
		"KW_ENDIF", "KW_WHILE", "KW_ENDWHILE", "KW_FOR", "KW_EACH", "KW_IN",
		"KW_ENDFOR", "KW_ON_ERROR", "KW_ENDON", "KW_CLEAR_ERROR", "KW_MUST",
		"KW_MUSTBE", "KW_FAIL", "KW_NO", "KW_SOME", "KW_TOOL", "KW_LAST", "KW_EVAL",
		"KW_TRUE", "KW_FALSE", "KW_AND", "KW_OR", "KW_NOT", "KW_LN", "KW_LOG",
		"KW_SIN", "KW_COS", "KW_TAN", "KW_ASIN", "KW_ACOS", "KW_ATAN", "KW_ASK",
		"KW_INTO", "KW_BREAK", "KW_CONTINUE", "KW_NIL", "KW_TYPEOF", "NUMBER_LIT",
		"STRING_LIT", "TRIPLE_BACKTICK_STRING", "ASSIGN", "PLUS", "MINUS", "STAR",
		"SLASH", "PERCENT", "STAR_STAR", "AMPERSAND", "PIPE", "CARET", "TILDE",
		"LPAREN", "RPAREN", "COMMA", "LBRACK", "RBRACK", "LBRACE", "RBRACE",
		"COLON", "DOT", "PLACEHOLDER_START", "PLACEHOLDER_END", "EQ", "NEQ",
		"GT", "LT", "GTE", "LTE", "IDENTIFIER", "METADATA_LINE", "LINE_COMMENT",
		"NEWLINE", "WS",
	}
	staticData.RuleNames = []string{
		"program", "file_header", "procedure_definition", "signature_part",
		"needs_clause", "optional_clause", "returns_clause", "param_list", "metadata_block",
		"statement_list", "body_line", "statement", "simple_statement", "block_statement",
		"expressionStatement", "set_statement", "call_statement", "return_statement",
		"emit_statement", "must_statement", "fail_statement", "clearErrorStmt",
		"ask_stmt", "break_statement", "continue_statement", "if_statement",
		"while_statement", "for_each_statement", "onErrorStmt", "qualified_identifier",
		"call_target", "expression", "logical_or_expr", "logical_and_expr",
		"bitwise_or_expr", "bitwise_xor_expr", "bitwise_and_expr", "equality_expr",
		"relational_expr", "additive_expr", "multiplicative_expr", "unary_expr",
		"power_expr", "accessor_expr", "primary", "callable_expr", "placeholder",
		"literal", "nil_literal", "boolean_literal", "list_literal", "map_literal",
		"expression_list_opt", "expression_list", "map_entry_list_opt", "map_entry_list",
		"map_entry",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 85, 494, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7, 20, 2,
		21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25, 2, 26,
		7, 26, 2, 27, 7, 27, 2, 28, 7, 28, 2, 29, 7, 29, 2, 30, 7, 30, 2, 31, 7,
		31, 2, 32, 7, 32, 2, 33, 7, 33, 2, 34, 7, 34, 2, 35, 7, 35, 2, 36, 7, 36,
		2, 37, 7, 37, 2, 38, 7, 38, 2, 39, 7, 39, 2, 40, 7, 40, 2, 41, 7, 41, 2,
		42, 7, 42, 2, 43, 7, 43, 2, 44, 7, 44, 2, 45, 7, 45, 2, 46, 7, 46, 2, 47,
		7, 47, 2, 48, 7, 48, 2, 49, 7, 49, 2, 50, 7, 50, 2, 51, 7, 51, 2, 52, 7,
		52, 2, 53, 7, 53, 2, 54, 7, 54, 2, 55, 7, 55, 2, 56, 7, 56, 1, 0, 1, 0,
		1, 0, 5, 0, 118, 8, 0, 10, 0, 12, 0, 121, 9, 0, 5, 0, 123, 8, 0, 10, 0,
		12, 0, 126, 9, 0, 1, 0, 1, 0, 1, 1, 5, 1, 131, 8, 1, 10, 1, 12, 1, 134,
		9, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 3, 2, 142, 8, 2, 1, 2, 1, 2,
		1, 2, 1, 3, 1, 3, 3, 3, 149, 8, 3, 1, 3, 3, 3, 152, 8, 3, 1, 3, 3, 3, 155,
		8, 3, 1, 3, 1, 3, 1, 3, 3, 3, 160, 8, 3, 1, 3, 3, 3, 163, 8, 3, 1, 3, 1,
		3, 3, 3, 167, 8, 3, 1, 3, 1, 3, 3, 3, 171, 8, 3, 1, 4, 1, 4, 1, 4, 1, 5,
		1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 5, 7, 185, 8, 7, 10, 7,
		12, 7, 188, 9, 7, 1, 8, 1, 8, 5, 8, 192, 8, 8, 10, 8, 12, 8, 195, 9, 8,
		1, 9, 5, 9, 198, 8, 9, 10, 9, 12, 9, 201, 9, 9, 1, 10, 1, 10, 1, 10, 1,
		10, 3, 10, 207, 8, 10, 1, 11, 1, 11, 3, 11, 211, 8, 11, 1, 12, 1, 12, 1,
		12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 3, 12, 223, 8, 12,
		1, 13, 1, 13, 1, 13, 1, 13, 3, 13, 229, 8, 13, 1, 14, 1, 14, 1, 15, 1,
		15, 1, 15, 1, 15, 1, 15, 1, 16, 1, 16, 1, 16, 1, 17, 1, 17, 3, 17, 243,
		8, 17, 1, 18, 1, 18, 1, 18, 1, 19, 1, 19, 1, 19, 1, 19, 3, 19, 252, 8,
		19, 1, 20, 1, 20, 3, 20, 256, 8, 20, 1, 21, 1, 21, 1, 22, 1, 22, 1, 22,
		1, 22, 3, 22, 264, 8, 22, 1, 23, 1, 23, 1, 24, 1, 24, 1, 25, 1, 25, 1,
		25, 1, 25, 1, 25, 1, 25, 1, 25, 3, 25, 277, 8, 25, 1, 25, 1, 25, 1, 26,
		1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 27, 1, 27, 1, 27, 1, 27, 1, 27, 1,
		27, 1, 27, 1, 27, 1, 27, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 29,
		1, 29, 1, 29, 5, 29, 305, 8, 29, 10, 29, 12, 29, 308, 9, 29, 1, 30, 1,
		30, 1, 30, 1, 30, 3, 30, 314, 8, 30, 1, 31, 1, 31, 1, 32, 1, 32, 1, 32,
		5, 32, 321, 8, 32, 10, 32, 12, 32, 324, 9, 32, 1, 33, 1, 33, 1, 33, 5,
		33, 329, 8, 33, 10, 33, 12, 33, 332, 9, 33, 1, 34, 1, 34, 1, 34, 5, 34,
		337, 8, 34, 10, 34, 12, 34, 340, 9, 34, 1, 35, 1, 35, 1, 35, 5, 35, 345,
		8, 35, 10, 35, 12, 35, 348, 9, 35, 1, 36, 1, 36, 1, 36, 5, 36, 353, 8,
		36, 10, 36, 12, 36, 356, 9, 36, 1, 37, 1, 37, 1, 37, 5, 37, 361, 8, 37,
		10, 37, 12, 37, 364, 9, 37, 1, 38, 1, 38, 1, 38, 5, 38, 369, 8, 38, 10,
		38, 12, 38, 372, 9, 38, 1, 39, 1, 39, 1, 39, 5, 39, 377, 8, 39, 10, 39,
		12, 39, 380, 9, 39, 1, 40, 1, 40, 1, 40, 5, 40, 385, 8, 40, 10, 40, 12,
		40, 388, 9, 40, 1, 41, 1, 41, 1, 41, 1, 41, 1, 41, 3, 41, 395, 8, 41, 1,
		42, 1, 42, 1, 42, 3, 42, 400, 8, 42, 1, 43, 1, 43, 1, 43, 1, 43, 1, 43,
		5, 43, 407, 8, 43, 10, 43, 12, 43, 410, 9, 43, 1, 44, 1, 44, 1, 44, 1,
		44, 1, 44, 1, 44, 1, 44, 1, 44, 1, 44, 1, 44, 1, 44, 1, 44, 1, 44, 1, 44,
		3, 44, 426, 8, 44, 1, 45, 1, 45, 1, 45, 1, 45, 1, 45, 1, 45, 1, 45, 1,
		45, 1, 45, 3, 45, 437, 8, 45, 1, 45, 1, 45, 1, 45, 1, 45, 1, 46, 1, 46,
		1, 46, 1, 46, 1, 47, 1, 47, 1, 47, 1, 47, 1, 47, 1, 47, 1, 47, 3, 47, 454,
		8, 47, 1, 48, 1, 48, 1, 49, 1, 49, 1, 50, 1, 50, 1, 50, 1, 50, 1, 51, 1,
		51, 1, 51, 1, 51, 1, 52, 3, 52, 469, 8, 52, 1, 53, 1, 53, 1, 53, 5, 53,
		474, 8, 53, 10, 53, 12, 53, 477, 9, 53, 1, 54, 3, 54, 480, 8, 54, 1, 55,
		1, 55, 1, 55, 5, 55, 485, 8, 55, 10, 55, 12, 55, 488, 9, 55, 1, 56, 1,
		56, 1, 56, 1, 56, 1, 56, 0, 0, 57, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20,
		22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56,
		58, 60, 62, 64, 66, 68, 70, 72, 74, 76, 78, 80, 82, 84, 86, 88, 90, 92,
		94, 96, 98, 100, 102, 104, 106, 108, 110, 112, 0, 8, 2, 0, 82, 82, 84,
		84, 1, 0, 75, 76, 1, 0, 77, 80, 1, 0, 54, 55, 1, 0, 56, 58, 4, 0, 26, 27,
		35, 35, 55, 55, 63, 63, 2, 0, 29, 29, 81, 81, 1, 0, 31, 32, 511, 0, 114,
		1, 0, 0, 0, 2, 132, 1, 0, 0, 0, 4, 135, 1, 0, 0, 0, 6, 170, 1, 0, 0, 0,
		8, 172, 1, 0, 0, 0, 10, 175, 1, 0, 0, 0, 12, 178, 1, 0, 0, 0, 14, 181,
		1, 0, 0, 0, 16, 193, 1, 0, 0, 0, 18, 199, 1, 0, 0, 0, 20, 206, 1, 0, 0,
		0, 22, 210, 1, 0, 0, 0, 24, 222, 1, 0, 0, 0, 26, 228, 1, 0, 0, 0, 28, 230,
		1, 0, 0, 0, 30, 232, 1, 0, 0, 0, 32, 237, 1, 0, 0, 0, 34, 240, 1, 0, 0,
		0, 36, 244, 1, 0, 0, 0, 38, 251, 1, 0, 0, 0, 40, 253, 1, 0, 0, 0, 42, 257,
		1, 0, 0, 0, 44, 259, 1, 0, 0, 0, 46, 265, 1, 0, 0, 0, 48, 267, 1, 0, 0,
		0, 50, 269, 1, 0, 0, 0, 52, 280, 1, 0, 0, 0, 54, 286, 1, 0, 0, 0, 56, 295,
		1, 0, 0, 0, 58, 301, 1, 0, 0, 0, 60, 313, 1, 0, 0, 0, 62, 315, 1, 0, 0,
		0, 64, 317, 1, 0, 0, 0, 66, 325, 1, 0, 0, 0, 68, 333, 1, 0, 0, 0, 70, 341,
		1, 0, 0, 0, 72, 349, 1, 0, 0, 0, 74, 357, 1, 0, 0, 0, 76, 365, 1, 0, 0,
		0, 78, 373, 1, 0, 0, 0, 80, 381, 1, 0, 0, 0, 82, 394, 1, 0, 0, 0, 84, 396,
		1, 0, 0, 0, 86, 401, 1, 0, 0, 0, 88, 425, 1, 0, 0, 0, 90, 436, 1, 0, 0,
		0, 92, 442, 1, 0, 0, 0, 94, 453, 1, 0, 0, 0, 96, 455, 1, 0, 0, 0, 98, 457,
		1, 0, 0, 0, 100, 459, 1, 0, 0, 0, 102, 463, 1, 0, 0, 0, 104, 468, 1, 0,
		0, 0, 106, 470, 1, 0, 0, 0, 108, 479, 1, 0, 0, 0, 110, 481, 1, 0, 0, 0,
		112, 489, 1, 0, 0, 0, 114, 124, 3, 2, 1, 0, 115, 119, 3, 4, 2, 0, 116,
		118, 5, 84, 0, 0, 117, 116, 1, 0, 0, 0, 118, 121, 1, 0, 0, 0, 119, 117,
		1, 0, 0, 0, 119, 120, 1, 0, 0, 0, 120, 123, 1, 0, 0, 0, 121, 119, 1, 0,
		0, 0, 122, 115, 1, 0, 0, 0, 123, 126, 1, 0, 0, 0, 124, 122, 1, 0, 0, 0,
		124, 125, 1, 0, 0, 0, 125, 127, 1, 0, 0, 0, 126, 124, 1, 0, 0, 0, 127,
		128, 5, 0, 0, 1, 128, 1, 1, 0, 0, 0, 129, 131, 7, 0, 0, 0, 130, 129, 1,
		0, 0, 0, 131, 134, 1, 0, 0, 0, 132, 130, 1, 0, 0, 0, 132, 133, 1, 0, 0,
		0, 133, 3, 1, 0, 0, 0, 134, 132, 1, 0, 0, 0, 135, 136, 5, 2, 0, 0, 136,
		137, 5, 81, 0, 0, 137, 138, 3, 6, 3, 0, 138, 139, 5, 6, 0, 0, 139, 141,
		5, 84, 0, 0, 140, 142, 3, 16, 8, 0, 141, 140, 1, 0, 0, 0, 141, 142, 1,
		0, 0, 0, 142, 143, 1, 0, 0, 0, 143, 144, 3, 18, 9, 0, 144, 145, 5, 7, 0,
		0, 145, 5, 1, 0, 0, 0, 146, 148, 5, 64, 0, 0, 147, 149, 3, 8, 4, 0, 148,
		147, 1, 0, 0, 0, 148, 149, 1, 0, 0, 0, 149, 151, 1, 0, 0, 0, 150, 152,
		3, 10, 5, 0, 151, 150, 1, 0, 0, 0, 151, 152, 1, 0, 0, 0, 152, 154, 1, 0,
		0, 0, 153, 155, 3, 12, 6, 0, 154, 153, 1, 0, 0, 0, 154, 155, 1, 0, 0, 0,
		155, 156, 1, 0, 0, 0, 156, 171, 5, 65, 0, 0, 157, 159, 3, 8, 4, 0, 158,
		160, 3, 10, 5, 0, 159, 158, 1, 0, 0, 0, 159, 160, 1, 0, 0, 0, 160, 162,
		1, 0, 0, 0, 161, 163, 3, 12, 6, 0, 162, 161, 1, 0, 0, 0, 162, 163, 1, 0,
		0, 0, 163, 171, 1, 0, 0, 0, 164, 166, 3, 10, 5, 0, 165, 167, 3, 12, 6,
		0, 166, 165, 1, 0, 0, 0, 166, 167, 1, 0, 0, 0, 167, 171, 1, 0, 0, 0, 168,
		171, 3, 12, 6, 0, 169, 171, 1, 0, 0, 0, 170, 146, 1, 0, 0, 0, 170, 157,
		1, 0, 0, 0, 170, 164, 1, 0, 0, 0, 170, 168, 1, 0, 0, 0, 170, 169, 1, 0,
		0, 0, 171, 7, 1, 0, 0, 0, 172, 173, 5, 3, 0, 0, 173, 174, 3, 14, 7, 0,
		174, 9, 1, 0, 0, 0, 175, 176, 5, 4, 0, 0, 176, 177, 3, 14, 7, 0, 177, 11,
		1, 0, 0, 0, 178, 179, 5, 5, 0, 0, 179, 180, 3, 14, 7, 0, 180, 13, 1, 0,
		0, 0, 181, 186, 5, 81, 0, 0, 182, 183, 5, 66, 0, 0, 183, 185, 5, 81, 0,
		0, 184, 182, 1, 0, 0, 0, 185, 188, 1, 0, 0, 0, 186, 184, 1, 0, 0, 0, 186,
		187, 1, 0, 0, 0, 187, 15, 1, 0, 0, 0, 188, 186, 1, 0, 0, 0, 189, 190, 5,
		82, 0, 0, 190, 192, 5, 84, 0, 0, 191, 189, 1, 0, 0, 0, 192, 195, 1, 0,
		0, 0, 193, 191, 1, 0, 0, 0, 193, 194, 1, 0, 0, 0, 194, 17, 1, 0, 0, 0,
		195, 193, 1, 0, 0, 0, 196, 198, 3, 20, 10, 0, 197, 196, 1, 0, 0, 0, 198,
		201, 1, 0, 0, 0, 199, 197, 1, 0, 0, 0, 199, 200, 1, 0, 0, 0, 200, 19, 1,
		0, 0, 0, 201, 199, 1, 0, 0, 0, 202, 203, 3, 22, 11, 0, 203, 204, 5, 84,
		0, 0, 204, 207, 1, 0, 0, 0, 205, 207, 5, 84, 0, 0, 206, 202, 1, 0, 0, 0,
		206, 205, 1, 0, 0, 0, 207, 21, 1, 0, 0, 0, 208, 211, 3, 24, 12, 0, 209,
		211, 3, 26, 13, 0, 210, 208, 1, 0, 0, 0, 210, 209, 1, 0, 0, 0, 211, 23,
		1, 0, 0, 0, 212, 223, 3, 30, 15, 0, 213, 223, 3, 32, 16, 0, 214, 223, 3,
		34, 17, 0, 215, 223, 3, 36, 18, 0, 216, 223, 3, 38, 19, 0, 217, 223, 3,
		40, 20, 0, 218, 223, 3, 42, 21, 0, 219, 223, 3, 44, 22, 0, 220, 223, 3,
		46, 23, 0, 221, 223, 3, 48, 24, 0, 222, 212, 1, 0, 0, 0, 222, 213, 1, 0,
		0, 0, 222, 214, 1, 0, 0, 0, 222, 215, 1, 0, 0, 0, 222, 216, 1, 0, 0, 0,
		222, 217, 1, 0, 0, 0, 222, 218, 1, 0, 0, 0, 222, 219, 1, 0, 0, 0, 222,
		220, 1, 0, 0, 0, 222, 221, 1, 0, 0, 0, 223, 25, 1, 0, 0, 0, 224, 229, 3,
		50, 25, 0, 225, 229, 3, 52, 26, 0, 226, 229, 3, 54, 27, 0, 227, 229, 3,
		56, 28, 0, 228, 224, 1, 0, 0, 0, 228, 225, 1, 0, 0, 0, 228, 226, 1, 0,
		0, 0, 228, 227, 1, 0, 0, 0, 229, 27, 1, 0, 0, 0, 230, 231, 3, 62, 31, 0,
		231, 29, 1, 0, 0, 0, 232, 233, 5, 8, 0, 0, 233, 234, 5, 81, 0, 0, 234,
		235, 5, 53, 0, 0, 235, 236, 3, 62, 31, 0, 236, 31, 1, 0, 0, 0, 237, 238,
		5, 1, 0, 0, 238, 239, 3, 90, 45, 0, 239, 33, 1, 0, 0, 0, 240, 242, 5, 9,
		0, 0, 241, 243, 3, 106, 53, 0, 242, 241, 1, 0, 0, 0, 242, 243, 1, 0, 0,
		0, 243, 35, 1, 0, 0, 0, 244, 245, 5, 10, 0, 0, 245, 246, 3, 62, 31, 0,
		246, 37, 1, 0, 0, 0, 247, 248, 5, 23, 0, 0, 248, 252, 3, 62, 31, 0, 249,
		250, 5, 24, 0, 0, 250, 252, 3, 90, 45, 0, 251, 247, 1, 0, 0, 0, 251, 249,
		1, 0, 0, 0, 252, 39, 1, 0, 0, 0, 253, 255, 5, 25, 0, 0, 254, 256, 3, 62,
		31, 0, 255, 254, 1, 0, 0, 0, 255, 256, 1, 0, 0, 0, 256, 41, 1, 0, 0, 0,
		257, 258, 5, 22, 0, 0, 258, 43, 1, 0, 0, 0, 259, 260, 5, 44, 0, 0, 260,
		263, 3, 62, 31, 0, 261, 262, 5, 45, 0, 0, 262, 264, 5, 81, 0, 0, 263, 261,
		1, 0, 0, 0, 263, 264, 1, 0, 0, 0, 264, 45, 1, 0, 0, 0, 265, 266, 5, 46,
		0, 0, 266, 47, 1, 0, 0, 0, 267, 268, 5, 47, 0, 0, 268, 49, 1, 0, 0, 0,
		269, 270, 5, 11, 0, 0, 270, 271, 3, 62, 31, 0, 271, 272, 5, 84, 0, 0, 272,
		276, 3, 18, 9, 0, 273, 274, 5, 12, 0, 0, 274, 275, 5, 84, 0, 0, 275, 277,
		3, 18, 9, 0, 276, 273, 1, 0, 0, 0, 276, 277, 1, 0, 0, 0, 277, 278, 1, 0,
		0, 0, 278, 279, 5, 13, 0, 0, 279, 51, 1, 0, 0, 0, 280, 281, 5, 14, 0, 0,
		281, 282, 3, 62, 31, 0, 282, 283, 5, 84, 0, 0, 283, 284, 3, 18, 9, 0, 284,
		285, 5, 15, 0, 0, 285, 53, 1, 0, 0, 0, 286, 287, 5, 16, 0, 0, 287, 288,
		5, 17, 0, 0, 288, 289, 5, 81, 0, 0, 289, 290, 5, 18, 0, 0, 290, 291, 3,
		62, 31, 0, 291, 292, 5, 84, 0, 0, 292, 293, 3, 18, 9, 0, 293, 294, 5, 19,
		0, 0, 294, 55, 1, 0, 0, 0, 295, 296, 5, 20, 0, 0, 296, 297, 5, 6, 0, 0,
		297, 298, 5, 84, 0, 0, 298, 299, 3, 18, 9, 0, 299, 300, 5, 21, 0, 0, 300,
		57, 1, 0, 0, 0, 301, 306, 5, 81, 0, 0, 302, 303, 5, 72, 0, 0, 303, 305,
		5, 81, 0, 0, 304, 302, 1, 0, 0, 0, 305, 308, 1, 0, 0, 0, 306, 304, 1, 0,
		0, 0, 306, 307, 1, 0, 0, 0, 307, 59, 1, 0, 0, 0, 308, 306, 1, 0, 0, 0,
		309, 314, 5, 81, 0, 0, 310, 311, 5, 28, 0, 0, 311, 312, 5, 72, 0, 0, 312,
		314, 3, 58, 29, 0, 313, 309, 1, 0, 0, 0, 313, 310, 1, 0, 0, 0, 314, 61,
		1, 0, 0, 0, 315, 316, 3, 64, 32, 0, 316, 63, 1, 0, 0, 0, 317, 322, 3, 66,
		33, 0, 318, 319, 5, 34, 0, 0, 319, 321, 3, 66, 33, 0, 320, 318, 1, 0, 0,
		0, 321, 324, 1, 0, 0, 0, 322, 320, 1, 0, 0, 0, 322, 323, 1, 0, 0, 0, 323,
		65, 1, 0, 0, 0, 324, 322, 1, 0, 0, 0, 325, 330, 3, 68, 34, 0, 326, 327,
		5, 33, 0, 0, 327, 329, 3, 68, 34, 0, 328, 326, 1, 0, 0, 0, 329, 332, 1,
		0, 0, 0, 330, 328, 1, 0, 0, 0, 330, 331, 1, 0, 0, 0, 331, 67, 1, 0, 0,
		0, 332, 330, 1, 0, 0, 0, 333, 338, 3, 70, 35, 0, 334, 335, 5, 61, 0, 0,
		335, 337, 3, 70, 35, 0, 336, 334, 1, 0, 0, 0, 337, 340, 1, 0, 0, 0, 338,
		336, 1, 0, 0, 0, 338, 339, 1, 0, 0, 0, 339, 69, 1, 0, 0, 0, 340, 338, 1,
		0, 0, 0, 341, 346, 3, 72, 36, 0, 342, 343, 5, 62, 0, 0, 343, 345, 3, 72,
		36, 0, 344, 342, 1, 0, 0, 0, 345, 348, 1, 0, 0, 0, 346, 344, 1, 0, 0, 0,
		346, 347, 1, 0, 0, 0, 347, 71, 1, 0, 0, 0, 348, 346, 1, 0, 0, 0, 349, 354,
		3, 74, 37, 0, 350, 351, 5, 60, 0, 0, 351, 353, 3, 74, 37, 0, 352, 350,
		1, 0, 0, 0, 353, 356, 1, 0, 0, 0, 354, 352, 1, 0, 0, 0, 354, 355, 1, 0,
		0, 0, 355, 73, 1, 0, 0, 0, 356, 354, 1, 0, 0, 0, 357, 362, 3, 76, 38, 0,
		358, 359, 7, 1, 0, 0, 359, 361, 3, 76, 38, 0, 360, 358, 1, 0, 0, 0, 361,
		364, 1, 0, 0, 0, 362, 360, 1, 0, 0, 0, 362, 363, 1, 0, 0, 0, 363, 75, 1,
		0, 0, 0, 364, 362, 1, 0, 0, 0, 365, 370, 3, 78, 39, 0, 366, 367, 7, 2,
		0, 0, 367, 369, 3, 78, 39, 0, 368, 366, 1, 0, 0, 0, 369, 372, 1, 0, 0,
		0, 370, 368, 1, 0, 0, 0, 370, 371, 1, 0, 0, 0, 371, 77, 1, 0, 0, 0, 372,
		370, 1, 0, 0, 0, 373, 378, 3, 80, 40, 0, 374, 375, 7, 3, 0, 0, 375, 377,
		3, 80, 40, 0, 376, 374, 1, 0, 0, 0, 377, 380, 1, 0, 0, 0, 378, 376, 1,
		0, 0, 0, 378, 379, 1, 0, 0, 0, 379, 79, 1, 0, 0, 0, 380, 378, 1, 0, 0,
		0, 381, 386, 3, 82, 41, 0, 382, 383, 7, 4, 0, 0, 383, 385, 3, 82, 41, 0,
		384, 382, 1, 0, 0, 0, 385, 388, 1, 0, 0, 0, 386, 384, 1, 0, 0, 0, 386,
		387, 1, 0, 0, 0, 387, 81, 1, 0, 0, 0, 388, 386, 1, 0, 0, 0, 389, 390, 7,
		5, 0, 0, 390, 395, 3, 82, 41, 0, 391, 392, 5, 49, 0, 0, 392, 395, 3, 82,
		41, 0, 393, 395, 3, 84, 42, 0, 394, 389, 1, 0, 0, 0, 394, 391, 1, 0, 0,
		0, 394, 393, 1, 0, 0, 0, 395, 83, 1, 0, 0, 0, 396, 399, 3, 86, 43, 0, 397,
		398, 5, 59, 0, 0, 398, 400, 3, 84, 42, 0, 399, 397, 1, 0, 0, 0, 399, 400,
		1, 0, 0, 0, 400, 85, 1, 0, 0, 0, 401, 408, 3, 88, 44, 0, 402, 403, 5, 67,
		0, 0, 403, 404, 3, 62, 31, 0, 404, 405, 5, 68, 0, 0, 405, 407, 1, 0, 0,
		0, 406, 402, 1, 0, 0, 0, 407, 410, 1, 0, 0, 0, 408, 406, 1, 0, 0, 0, 408,
		409, 1, 0, 0, 0, 409, 87, 1, 0, 0, 0, 410, 408, 1, 0, 0, 0, 411, 426, 3,
		94, 47, 0, 412, 426, 3, 92, 46, 0, 413, 426, 5, 81, 0, 0, 414, 426, 5,
		29, 0, 0, 415, 426, 3, 90, 45, 0, 416, 417, 5, 30, 0, 0, 417, 418, 5, 64,
		0, 0, 418, 419, 3, 62, 31, 0, 419, 420, 5, 65, 0, 0, 420, 426, 1, 0, 0,
		0, 421, 422, 5, 64, 0, 0, 422, 423, 3, 62, 31, 0, 423, 424, 5, 65, 0, 0,
		424, 426, 1, 0, 0, 0, 425, 411, 1, 0, 0, 0, 425, 412, 1, 0, 0, 0, 425,
		413, 1, 0, 0, 0, 425, 414, 1, 0, 0, 0, 425, 415, 1, 0, 0, 0, 425, 416,
		1, 0, 0, 0, 425, 421, 1, 0, 0, 0, 426, 89, 1, 0, 0, 0, 427, 437, 3, 60,
		30, 0, 428, 437, 5, 36, 0, 0, 429, 437, 5, 37, 0, 0, 430, 437, 5, 38, 0,
		0, 431, 437, 5, 39, 0, 0, 432, 437, 5, 40, 0, 0, 433, 437, 5, 41, 0, 0,
		434, 437, 5, 42, 0, 0, 435, 437, 5, 43, 0, 0, 436, 427, 1, 0, 0, 0, 436,
		428, 1, 0, 0, 0, 436, 429, 1, 0, 0, 0, 436, 430, 1, 0, 0, 0, 436, 431,
		1, 0, 0, 0, 436, 432, 1, 0, 0, 0, 436, 433, 1, 0, 0, 0, 436, 434, 1, 0,
		0, 0, 436, 435, 1, 0, 0, 0, 437, 438, 1, 0, 0, 0, 438, 439, 5, 64, 0, 0,
		439, 440, 3, 104, 52, 0, 440, 441, 5, 65, 0, 0, 441, 91, 1, 0, 0, 0, 442,
		443, 5, 73, 0, 0, 443, 444, 7, 6, 0, 0, 444, 445, 5, 74, 0, 0, 445, 93,
		1, 0, 0, 0, 446, 454, 5, 51, 0, 0, 447, 454, 5, 52, 0, 0, 448, 454, 5,
		50, 0, 0, 449, 454, 3, 100, 50, 0, 450, 454, 3, 102, 51, 0, 451, 454, 3,
		98, 49, 0, 452, 454, 3, 96, 48, 0, 453, 446, 1, 0, 0, 0, 453, 447, 1, 0,
		0, 0, 453, 448, 1, 0, 0, 0, 453, 449, 1, 0, 0, 0, 453, 450, 1, 0, 0, 0,
		453, 451, 1, 0, 0, 0, 453, 452, 1, 0, 0, 0, 454, 95, 1, 0, 0, 0, 455, 456,
		5, 48, 0, 0, 456, 97, 1, 0, 0, 0, 457, 458, 7, 7, 0, 0, 458, 99, 1, 0,
		0, 0, 459, 460, 5, 67, 0, 0, 460, 461, 3, 104, 52, 0, 461, 462, 5, 68,
		0, 0, 462, 101, 1, 0, 0, 0, 463, 464, 5, 69, 0, 0, 464, 465, 3, 108, 54,
		0, 465, 466, 5, 70, 0, 0, 466, 103, 1, 0, 0, 0, 467, 469, 3, 106, 53, 0,
		468, 467, 1, 0, 0, 0, 468, 469, 1, 0, 0, 0, 469, 105, 1, 0, 0, 0, 470,
		475, 3, 62, 31, 0, 471, 472, 5, 66, 0, 0, 472, 474, 3, 62, 31, 0, 473,
		471, 1, 0, 0, 0, 474, 477, 1, 0, 0, 0, 475, 473, 1, 0, 0, 0, 475, 476,
		1, 0, 0, 0, 476, 107, 1, 0, 0, 0, 477, 475, 1, 0, 0, 0, 478, 480, 3, 110,
		55, 0, 479, 478, 1, 0, 0, 0, 479, 480, 1, 0, 0, 0, 480, 109, 1, 0, 0, 0,
		481, 486, 3, 112, 56, 0, 482, 483, 5, 66, 0, 0, 483, 485, 3, 112, 56, 0,
		484, 482, 1, 0, 0, 0, 485, 488, 1, 0, 0, 0, 486, 484, 1, 0, 0, 0, 486,
		487, 1, 0, 0, 0, 487, 111, 1, 0, 0, 0, 488, 486, 1, 0, 0, 0, 489, 490,
		5, 51, 0, 0, 490, 491, 5, 71, 0, 0, 491, 492, 3, 62, 31, 0, 492, 113, 1,
		0, 0, 0, 44, 119, 124, 132, 141, 148, 151, 154, 159, 162, 166, 170, 186,
		193, 199, 206, 210, 222, 228, 242, 251, 255, 263, 276, 306, 313, 322, 330,
		338, 346, 354, 362, 370, 378, 386, 394, 399, 408, 425, 436, 453, 468, 475,
		479, 486,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// NeuroScriptParserInit initializes any static state used to implement NeuroScriptParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewNeuroScriptParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func NeuroScriptParserInit() {
	staticData := &NeuroScriptParserStaticData
	staticData.once.Do(neuroscriptParserInit)
}

// NewNeuroScriptParser produces a new parser instance for the optional input antlr.TokenStream.
func NewNeuroScriptParser(input antlr.TokenStream) *NeuroScriptParser {
	NeuroScriptParserInit()
	this := new(NeuroScriptParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &NeuroScriptParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "NeuroScript.g4"

	return this
}

// NeuroScriptParser tokens.
const (
	NeuroScriptParserEOF                    = antlr.TokenEOF
	NeuroScriptParserKW_CALL                = 1
	NeuroScriptParserKW_FUNC                = 2
	NeuroScriptParserKW_NEEDS               = 3
	NeuroScriptParserKW_OPTIONAL            = 4
	NeuroScriptParserKW_RETURNS             = 5
	NeuroScriptParserKW_MEANS               = 6
	NeuroScriptParserKW_ENDFUNC             = 7
	NeuroScriptParserKW_SET                 = 8
	NeuroScriptParserKW_RETURN              = 9
	NeuroScriptParserKW_EMIT                = 10
	NeuroScriptParserKW_IF                  = 11
	NeuroScriptParserKW_ELSE                = 12
	NeuroScriptParserKW_ENDIF               = 13
	NeuroScriptParserKW_WHILE               = 14
	NeuroScriptParserKW_ENDWHILE            = 15
	NeuroScriptParserKW_FOR                 = 16
	NeuroScriptParserKW_EACH                = 17
	NeuroScriptParserKW_IN                  = 18
	NeuroScriptParserKW_ENDFOR              = 19
	NeuroScriptParserKW_ON_ERROR            = 20
	NeuroScriptParserKW_ENDON               = 21
	NeuroScriptParserKW_CLEAR_ERROR         = 22
	NeuroScriptParserKW_MUST                = 23
	NeuroScriptParserKW_MUSTBE              = 24
	NeuroScriptParserKW_FAIL                = 25
	NeuroScriptParserKW_NO                  = 26
	NeuroScriptParserKW_SOME                = 27
	NeuroScriptParserKW_TOOL                = 28
	NeuroScriptParserKW_LAST                = 29
	NeuroScriptParserKW_EVAL                = 30
	NeuroScriptParserKW_TRUE                = 31
	NeuroScriptParserKW_FALSE               = 32
	NeuroScriptParserKW_AND                 = 33
	NeuroScriptParserKW_OR                  = 34
	NeuroScriptParserKW_NOT                 = 35
	NeuroScriptParserKW_LN                  = 36
	NeuroScriptParserKW_LOG                 = 37
	NeuroScriptParserKW_SIN                 = 38
	NeuroScriptParserKW_COS                 = 39
	NeuroScriptParserKW_TAN                 = 40
	NeuroScriptParserKW_ASIN                = 41
	NeuroScriptParserKW_ACOS                = 42
	NeuroScriptParserKW_ATAN                = 43
	NeuroScriptParserKW_ASK                 = 44
	NeuroScriptParserKW_INTO                = 45
	NeuroScriptParserKW_BREAK               = 46
	NeuroScriptParserKW_CONTINUE            = 47
	NeuroScriptParserKW_NIL                 = 48
	NeuroScriptParserKW_TYPEOF              = 49
	NeuroScriptParserNUMBER_LIT             = 50
	NeuroScriptParserSTRING_LIT             = 51
	NeuroScriptParserTRIPLE_BACKTICK_STRING = 52
	NeuroScriptParserASSIGN                 = 53
	NeuroScriptParserPLUS                   = 54
	NeuroScriptParserMINUS                  = 55
	NeuroScriptParserSTAR                   = 56
	NeuroScriptParserSLASH                  = 57
	NeuroScriptParserPERCENT                = 58
	NeuroScriptParserSTAR_STAR              = 59
	NeuroScriptParserAMPERSAND              = 60
	NeuroScriptParserPIPE                   = 61
	NeuroScriptParserCARET                  = 62
	NeuroScriptParserTILDE                  = 63
	NeuroScriptParserLPAREN                 = 64
	NeuroScriptParserRPAREN                 = 65
	NeuroScriptParserCOMMA                  = 66
	NeuroScriptParserLBRACK                 = 67
	NeuroScriptParserRBRACK                 = 68
	NeuroScriptParserLBRACE                 = 69
	NeuroScriptParserRBRACE                 = 70
	NeuroScriptParserCOLON                  = 71
	NeuroScriptParserDOT                    = 72
	NeuroScriptParserPLACEHOLDER_START      = 73
	NeuroScriptParserPLACEHOLDER_END        = 74
	NeuroScriptParserEQ                     = 75
	NeuroScriptParserNEQ                    = 76
	NeuroScriptParserGT                     = 77
	NeuroScriptParserLT                     = 78
	NeuroScriptParserGTE                    = 79
	NeuroScriptParserLTE                    = 80
	NeuroScriptParserIDENTIFIER             = 81
	NeuroScriptParserMETADATA_LINE          = 82
	NeuroScriptParserLINE_COMMENT           = 83
	NeuroScriptParserNEWLINE                = 84
	NeuroScriptParserWS                     = 85
)

// NeuroScriptParser rules.
const (
	NeuroScriptParserRULE_program              = 0
	NeuroScriptParserRULE_file_header          = 1
	NeuroScriptParserRULE_procedure_definition = 2
	NeuroScriptParserRULE_signature_part       = 3
	NeuroScriptParserRULE_needs_clause         = 4
	NeuroScriptParserRULE_optional_clause      = 5
	NeuroScriptParserRULE_returns_clause       = 6
	NeuroScriptParserRULE_param_list           = 7
	NeuroScriptParserRULE_metadata_block       = 8
	NeuroScriptParserRULE_statement_list       = 9
	NeuroScriptParserRULE_body_line            = 10
	NeuroScriptParserRULE_statement            = 11
	NeuroScriptParserRULE_simple_statement     = 12
	NeuroScriptParserRULE_block_statement      = 13
	NeuroScriptParserRULE_expressionStatement  = 14
	NeuroScriptParserRULE_set_statement        = 15
	NeuroScriptParserRULE_call_statement       = 16
	NeuroScriptParserRULE_return_statement     = 17
	NeuroScriptParserRULE_emit_statement       = 18
	NeuroScriptParserRULE_must_statement       = 19
	NeuroScriptParserRULE_fail_statement       = 20
	NeuroScriptParserRULE_clearErrorStmt       = 21
	NeuroScriptParserRULE_ask_stmt             = 22
	NeuroScriptParserRULE_break_statement      = 23
	NeuroScriptParserRULE_continue_statement   = 24
	NeuroScriptParserRULE_if_statement         = 25
	NeuroScriptParserRULE_while_statement      = 26
	NeuroScriptParserRULE_for_each_statement   = 27
	NeuroScriptParserRULE_onErrorStmt          = 28
	NeuroScriptParserRULE_qualified_identifier = 29
	NeuroScriptParserRULE_call_target          = 30
	NeuroScriptParserRULE_expression           = 31
	NeuroScriptParserRULE_logical_or_expr      = 32
	NeuroScriptParserRULE_logical_and_expr     = 33
	NeuroScriptParserRULE_bitwise_or_expr      = 34
	NeuroScriptParserRULE_bitwise_xor_expr     = 35
	NeuroScriptParserRULE_bitwise_and_expr     = 36
	NeuroScriptParserRULE_equality_expr        = 37
	NeuroScriptParserRULE_relational_expr      = 38
	NeuroScriptParserRULE_additive_expr        = 39
	NeuroScriptParserRULE_multiplicative_expr  = 40
	NeuroScriptParserRULE_unary_expr           = 41
	NeuroScriptParserRULE_power_expr           = 42
	NeuroScriptParserRULE_accessor_expr        = 43
	NeuroScriptParserRULE_primary              = 44
	NeuroScriptParserRULE_callable_expr        = 45
	NeuroScriptParserRULE_placeholder          = 46
	NeuroScriptParserRULE_literal              = 47
	NeuroScriptParserRULE_nil_literal          = 48
	NeuroScriptParserRULE_boolean_literal      = 49
	NeuroScriptParserRULE_list_literal         = 50
	NeuroScriptParserRULE_map_literal          = 51
	NeuroScriptParserRULE_expression_list_opt  = 52
	NeuroScriptParserRULE_expression_list      = 53
	NeuroScriptParserRULE_map_entry_list_opt   = 54
	NeuroScriptParserRULE_map_entry_list       = 55
	NeuroScriptParserRULE_map_entry            = 56
)

// IProgramContext is an interface to support dynamic dispatch.
type IProgramContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	File_header() IFile_headerContext
	EOF() antlr.TerminalNode
	AllProcedure_definition() []IProcedure_definitionContext
	Procedure_definition(i int) IProcedure_definitionContext
	AllNEWLINE() []antlr.TerminalNode
	NEWLINE(i int) antlr.TerminalNode

	// IsProgramContext differentiates from other interfaces.
	IsProgramContext()
}

type ProgramContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyProgramContext() *ProgramContext {
	var p = new(ProgramContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_program
	return p
}

func InitEmptyProgramContext(p *ProgramContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_program
}

func (*ProgramContext) IsProgramContext() {}

func NewProgramContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ProgramContext {
	var p = new(ProgramContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_program

	return p
}

func (s *ProgramContext) GetParser() antlr.Parser { return s.parser }

func (s *ProgramContext) File_header() IFile_headerContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFile_headerContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFile_headerContext)
}

func (s *ProgramContext) EOF() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserEOF, 0)
}

func (s *ProgramContext) AllProcedure_definition() []IProcedure_definitionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IProcedure_definitionContext); ok {
			len++
		}
	}

	tst := make([]IProcedure_definitionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IProcedure_definitionContext); ok {
			tst[i] = t.(IProcedure_definitionContext)
			i++
		}
	}

	return tst
}

func (s *ProgramContext) Procedure_definition(i int) IProcedure_definitionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IProcedure_definitionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IProcedure_definitionContext)
}

func (s *ProgramContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserNEWLINE)
}

func (s *ProgramContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, i)
}

func (s *ProgramContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ProgramContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ProgramContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterProgram(s)
	}
}

func (s *ProgramContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitProgram(s)
	}
}

func (s *ProgramContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitProgram(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Program() (localctx IProgramContext) {
	localctx = NewProgramContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, NeuroScriptParserRULE_program)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(114)
		p.File_header()
	}
	p.SetState(124)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserKW_FUNC {
		{
			p.SetState(115)
			p.Procedure_definition()
		}
		p.SetState(119)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == NeuroScriptParserNEWLINE {
			{
				p.SetState(116)
				p.Match(NeuroScriptParserNEWLINE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(121)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

		p.SetState(126)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(127)
		p.Match(NeuroScriptParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFile_headerContext is an interface to support dynamic dispatch.
type IFile_headerContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllMETADATA_LINE() []antlr.TerminalNode
	METADATA_LINE(i int) antlr.TerminalNode
	AllNEWLINE() []antlr.TerminalNode
	NEWLINE(i int) antlr.TerminalNode

	// IsFile_headerContext differentiates from other interfaces.
	IsFile_headerContext()
}

type File_headerContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFile_headerContext() *File_headerContext {
	var p = new(File_headerContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_file_header
	return p
}

func InitEmptyFile_headerContext(p *File_headerContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_file_header
}

func (*File_headerContext) IsFile_headerContext() {}

func NewFile_headerContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *File_headerContext {
	var p = new(File_headerContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_file_header

	return p
}

func (s *File_headerContext) GetParser() antlr.Parser { return s.parser }

func (s *File_headerContext) AllMETADATA_LINE() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserMETADATA_LINE)
}

func (s *File_headerContext) METADATA_LINE(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserMETADATA_LINE, i)
}

func (s *File_headerContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserNEWLINE)
}

func (s *File_headerContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, i)
}

func (s *File_headerContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *File_headerContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *File_headerContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterFile_header(s)
	}
}

func (s *File_headerContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitFile_header(s)
	}
}

func (s *File_headerContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitFile_header(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) File_header() (localctx IFile_headerContext) {
	localctx = NewFile_headerContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, NeuroScriptParserRULE_file_header)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(132)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserMETADATA_LINE || _la == NeuroScriptParserNEWLINE {
		{
			p.SetState(129)
			_la = p.GetTokenStream().LA(1)

			if !(_la == NeuroScriptParserMETADATA_LINE || _la == NeuroScriptParserNEWLINE) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

		p.SetState(134)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IProcedure_definitionContext is an interface to support dynamic dispatch.
type IProcedure_definitionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_FUNC() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode
	Signature_part() ISignature_partContext
	KW_MEANS() antlr.TerminalNode
	NEWLINE() antlr.TerminalNode
	Statement_list() IStatement_listContext
	KW_ENDFUNC() antlr.TerminalNode
	Metadata_block() IMetadata_blockContext

	// IsProcedure_definitionContext differentiates from other interfaces.
	IsProcedure_definitionContext()
}

type Procedure_definitionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyProcedure_definitionContext() *Procedure_definitionContext {
	var p = new(Procedure_definitionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_procedure_definition
	return p
}

func InitEmptyProcedure_definitionContext(p *Procedure_definitionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_procedure_definition
}

func (*Procedure_definitionContext) IsProcedure_definitionContext() {}

func NewProcedure_definitionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Procedure_definitionContext {
	var p = new(Procedure_definitionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_procedure_definition

	return p
}

func (s *Procedure_definitionContext) GetParser() antlr.Parser { return s.parser }

func (s *Procedure_definitionContext) KW_FUNC() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_FUNC, 0)
}

func (s *Procedure_definitionContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserIDENTIFIER, 0)
}

func (s *Procedure_definitionContext) Signature_part() ISignature_partContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISignature_partContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISignature_partContext)
}

func (s *Procedure_definitionContext) KW_MEANS() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_MEANS, 0)
}

func (s *Procedure_definitionContext) NEWLINE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, 0)
}

func (s *Procedure_definitionContext) Statement_list() IStatement_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStatement_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStatement_listContext)
}

func (s *Procedure_definitionContext) KW_ENDFUNC() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ENDFUNC, 0)
}

func (s *Procedure_definitionContext) Metadata_block() IMetadata_blockContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMetadata_blockContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMetadata_blockContext)
}

func (s *Procedure_definitionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Procedure_definitionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Procedure_definitionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterProcedure_definition(s)
	}
}

func (s *Procedure_definitionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitProcedure_definition(s)
	}
}

func (s *Procedure_definitionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitProcedure_definition(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Procedure_definition() (localctx IProcedure_definitionContext) {
	localctx = NewProcedure_definitionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, NeuroScriptParserRULE_procedure_definition)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(135)
		p.Match(NeuroScriptParserKW_FUNC)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(136)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(137)
		p.Signature_part()
	}
	{
		p.SetState(138)
		p.Match(NeuroScriptParserKW_MEANS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(139)
		p.Match(NeuroScriptParserNEWLINE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(141)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(140)
			p.Metadata_block()
		}

	} else if p.HasError() { // JIM
		goto errorExit
	}
	{
		p.SetState(143)
		p.Statement_list()
	}
	{
		p.SetState(144)
		p.Match(NeuroScriptParserKW_ENDFUNC)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISignature_partContext is an interface to support dynamic dispatch.
type ISignature_partContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LPAREN() antlr.TerminalNode
	RPAREN() antlr.TerminalNode
	Needs_clause() INeeds_clauseContext
	Optional_clause() IOptional_clauseContext
	Returns_clause() IReturns_clauseContext

	// IsSignature_partContext differentiates from other interfaces.
	IsSignature_partContext()
}

type Signature_partContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySignature_partContext() *Signature_partContext {
	var p = new(Signature_partContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_signature_part
	return p
}

func InitEmptySignature_partContext(p *Signature_partContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_signature_part
}

func (*Signature_partContext) IsSignature_partContext() {}

func NewSignature_partContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Signature_partContext {
	var p = new(Signature_partContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_signature_part

	return p
}

func (s *Signature_partContext) GetParser() antlr.Parser { return s.parser }

func (s *Signature_partContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLPAREN, 0)
}

func (s *Signature_partContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserRPAREN, 0)
}

func (s *Signature_partContext) Needs_clause() INeeds_clauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INeeds_clauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INeeds_clauseContext)
}

func (s *Signature_partContext) Optional_clause() IOptional_clauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOptional_clauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOptional_clauseContext)
}

func (s *Signature_partContext) Returns_clause() IReturns_clauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReturns_clauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReturns_clauseContext)
}

func (s *Signature_partContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Signature_partContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Signature_partContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterSignature_part(s)
	}
}

func (s *Signature_partContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitSignature_part(s)
	}
}

func (s *Signature_partContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitSignature_part(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Signature_part() (localctx ISignature_partContext) {
	localctx = NewSignature_partContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, NeuroScriptParserRULE_signature_part)
	var _la int

	p.SetState(170)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserLPAREN:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(146)
			p.Match(NeuroScriptParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(148)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == NeuroScriptParserKW_NEEDS {
			{
				p.SetState(147)
				p.Needs_clause()
			}

		}
		p.SetState(151)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == NeuroScriptParserKW_OPTIONAL {
			{
				p.SetState(150)
				p.Optional_clause()
			}

		}
		p.SetState(154)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == NeuroScriptParserKW_RETURNS {
			{
				p.SetState(153)
				p.Returns_clause()
			}

		}
		{
			p.SetState(156)
			p.Match(NeuroScriptParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_NEEDS:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(157)
			p.Needs_clause()
		}
		p.SetState(159)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == NeuroScriptParserKW_OPTIONAL {
			{
				p.SetState(158)
				p.Optional_clause()
			}

		}
		p.SetState(162)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == NeuroScriptParserKW_RETURNS {
			{
				p.SetState(161)
				p.Returns_clause()
			}

		}

	case NeuroScriptParserKW_OPTIONAL:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(164)
			p.Optional_clause()
		}
		p.SetState(166)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == NeuroScriptParserKW_RETURNS {
			{
				p.SetState(165)
				p.Returns_clause()
			}

		}

	case NeuroScriptParserKW_RETURNS:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(168)
			p.Returns_clause()
		}

	case NeuroScriptParserKW_MEANS:
		p.EnterOuterAlt(localctx, 5)

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// INeeds_clauseContext is an interface to support dynamic dispatch.
type INeeds_clauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_NEEDS() antlr.TerminalNode
	Param_list() IParam_listContext

	// IsNeeds_clauseContext differentiates from other interfaces.
	IsNeeds_clauseContext()
}

type Needs_clauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNeeds_clauseContext() *Needs_clauseContext {
	var p = new(Needs_clauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_needs_clause
	return p
}

func InitEmptyNeeds_clauseContext(p *Needs_clauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_needs_clause
}

func (*Needs_clauseContext) IsNeeds_clauseContext() {}

func NewNeeds_clauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Needs_clauseContext {
	var p = new(Needs_clauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_needs_clause

	return p
}

func (s *Needs_clauseContext) GetParser() antlr.Parser { return s.parser }

func (s *Needs_clauseContext) KW_NEEDS() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_NEEDS, 0)
}

func (s *Needs_clauseContext) Param_list() IParam_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IParam_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IParam_listContext)
}

func (s *Needs_clauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Needs_clauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Needs_clauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterNeeds_clause(s)
	}
}

func (s *Needs_clauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitNeeds_clause(s)
	}
}

func (s *Needs_clauseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitNeeds_clause(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Needs_clause() (localctx INeeds_clauseContext) {
	localctx = NewNeeds_clauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, NeuroScriptParserRULE_needs_clause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(172)
		p.Match(NeuroScriptParserKW_NEEDS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(173)
		p.Param_list()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOptional_clauseContext is an interface to support dynamic dispatch.
type IOptional_clauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_OPTIONAL() antlr.TerminalNode
	Param_list() IParam_listContext

	// IsOptional_clauseContext differentiates from other interfaces.
	IsOptional_clauseContext()
}

type Optional_clauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOptional_clauseContext() *Optional_clauseContext {
	var p = new(Optional_clauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_optional_clause
	return p
}

func InitEmptyOptional_clauseContext(p *Optional_clauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_optional_clause
}

func (*Optional_clauseContext) IsOptional_clauseContext() {}

func NewOptional_clauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Optional_clauseContext {
	var p = new(Optional_clauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_optional_clause

	return p
}

func (s *Optional_clauseContext) GetParser() antlr.Parser { return s.parser }

func (s *Optional_clauseContext) KW_OPTIONAL() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_OPTIONAL, 0)
}

func (s *Optional_clauseContext) Param_list() IParam_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IParam_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IParam_listContext)
}

func (s *Optional_clauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Optional_clauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Optional_clauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterOptional_clause(s)
	}
}

func (s *Optional_clauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitOptional_clause(s)
	}
}

func (s *Optional_clauseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitOptional_clause(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Optional_clause() (localctx IOptional_clauseContext) {
	localctx = NewOptional_clauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, NeuroScriptParserRULE_optional_clause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(175)
		p.Match(NeuroScriptParserKW_OPTIONAL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(176)
		p.Param_list()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IReturns_clauseContext is an interface to support dynamic dispatch.
type IReturns_clauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_RETURNS() antlr.TerminalNode
	Param_list() IParam_listContext

	// IsReturns_clauseContext differentiates from other interfaces.
	IsReturns_clauseContext()
}

type Returns_clauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyReturns_clauseContext() *Returns_clauseContext {
	var p = new(Returns_clauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_returns_clause
	return p
}

func InitEmptyReturns_clauseContext(p *Returns_clauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_returns_clause
}

func (*Returns_clauseContext) IsReturns_clauseContext() {}

func NewReturns_clauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Returns_clauseContext {
	var p = new(Returns_clauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_returns_clause

	return p
}

func (s *Returns_clauseContext) GetParser() antlr.Parser { return s.parser }

func (s *Returns_clauseContext) KW_RETURNS() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_RETURNS, 0)
}

func (s *Returns_clauseContext) Param_list() IParam_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IParam_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IParam_listContext)
}

func (s *Returns_clauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Returns_clauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Returns_clauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterReturns_clause(s)
	}
}

func (s *Returns_clauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitReturns_clause(s)
	}
}

func (s *Returns_clauseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitReturns_clause(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Returns_clause() (localctx IReturns_clauseContext) {
	localctx = NewReturns_clauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, NeuroScriptParserRULE_returns_clause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(178)
		p.Match(NeuroScriptParserKW_RETURNS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(179)
		p.Param_list()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IParam_listContext is an interface to support dynamic dispatch.
type IParam_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIDENTIFIER() []antlr.TerminalNode
	IDENTIFIER(i int) antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsParam_listContext differentiates from other interfaces.
	IsParam_listContext()
}

type Param_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyParam_listContext() *Param_listContext {
	var p = new(Param_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_param_list
	return p
}

func InitEmptyParam_listContext(p *Param_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_param_list
}

func (*Param_listContext) IsParam_listContext() {}

func NewParam_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Param_listContext {
	var p = new(Param_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_param_list

	return p
}

func (s *Param_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Param_listContext) AllIDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserIDENTIFIER)
}

func (s *Param_listContext) IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserIDENTIFIER, i)
}

func (s *Param_listContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserCOMMA)
}

func (s *Param_listContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserCOMMA, i)
}

func (s *Param_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Param_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Param_listContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterParam_list(s)
	}
}

func (s *Param_listContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitParam_list(s)
	}
}

func (s *Param_listContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitParam_list(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Param_list() (localctx IParam_listContext) {
	localctx = NewParam_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, NeuroScriptParserRULE_param_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(181)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(186)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserCOMMA {
		{
			p.SetState(182)
			p.Match(NeuroScriptParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(183)
			p.Match(NeuroScriptParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(188)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMetadata_blockContext is an interface to support dynamic dispatch.
type IMetadata_blockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllMETADATA_LINE() []antlr.TerminalNode
	METADATA_LINE(i int) antlr.TerminalNode
	AllNEWLINE() []antlr.TerminalNode
	NEWLINE(i int) antlr.TerminalNode

	// IsMetadata_blockContext differentiates from other interfaces.
	IsMetadata_blockContext()
}

type Metadata_blockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMetadata_blockContext() *Metadata_blockContext {
	var p = new(Metadata_blockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_metadata_block
	return p
}

func InitEmptyMetadata_blockContext(p *Metadata_blockContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_metadata_block
}

func (*Metadata_blockContext) IsMetadata_blockContext() {}

func NewMetadata_blockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Metadata_blockContext {
	var p = new(Metadata_blockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_metadata_block

	return p
}

func (s *Metadata_blockContext) GetParser() antlr.Parser { return s.parser }

func (s *Metadata_blockContext) AllMETADATA_LINE() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserMETADATA_LINE)
}

func (s *Metadata_blockContext) METADATA_LINE(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserMETADATA_LINE, i)
}

func (s *Metadata_blockContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserNEWLINE)
}

func (s *Metadata_blockContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, i)
}

func (s *Metadata_blockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Metadata_blockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Metadata_blockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterMetadata_block(s)
	}
}

func (s *Metadata_blockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitMetadata_block(s)
	}
}

func (s *Metadata_blockContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitMetadata_block(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Metadata_block() (localctx IMetadata_blockContext) {
	localctx = NewMetadata_blockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, NeuroScriptParserRULE_metadata_block)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(193)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserMETADATA_LINE {
		{
			p.SetState(189)
			p.Match(NeuroScriptParserMETADATA_LINE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(190)
			p.Match(NeuroScriptParserNEWLINE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(195)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IStatement_listContext is an interface to support dynamic dispatch.
type IStatement_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllBody_line() []IBody_lineContext
	Body_line(i int) IBody_lineContext

	// IsStatement_listContext differentiates from other interfaces.
	IsStatement_listContext()
}

type Statement_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStatement_listContext() *Statement_listContext {
	var p = new(Statement_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_statement_list
	return p
}

func InitEmptyStatement_listContext(p *Statement_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_statement_list
}

func (*Statement_listContext) IsStatement_listContext() {}

func NewStatement_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Statement_listContext {
	var p = new(Statement_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_statement_list

	return p
}

func (s *Statement_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Statement_listContext) AllBody_line() []IBody_lineContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IBody_lineContext); ok {
			len++
		}
	}

	tst := make([]IBody_lineContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IBody_lineContext); ok {
			tst[i] = t.(IBody_lineContext)
			i++
		}
	}

	return tst
}

func (s *Statement_listContext) Body_line(i int) IBody_lineContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBody_lineContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBody_lineContext)
}

func (s *Statement_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Statement_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Statement_listContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterStatement_list(s)
	}
}

func (s *Statement_listContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitStatement_list(s)
	}
}

func (s *Statement_listContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitStatement_list(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Statement_list() (localctx IStatement_listContext) {
	localctx = NewStatement_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, NeuroScriptParserRULE_statement_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(199)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&228698482626306) != 0) || _la == NeuroScriptParserNEWLINE {
		{
			p.SetState(196)
			p.Body_line()
		}

		p.SetState(201)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBody_lineContext is an interface to support dynamic dispatch.
type IBody_lineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Statement() IStatementContext
	NEWLINE() antlr.TerminalNode

	// IsBody_lineContext differentiates from other interfaces.
	IsBody_lineContext()
}

type Body_lineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBody_lineContext() *Body_lineContext {
	var p = new(Body_lineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_body_line
	return p
}

func InitEmptyBody_lineContext(p *Body_lineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_body_line
}

func (*Body_lineContext) IsBody_lineContext() {}

func NewBody_lineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Body_lineContext {
	var p = new(Body_lineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_body_line

	return p
}

func (s *Body_lineContext) GetParser() antlr.Parser { return s.parser }

func (s *Body_lineContext) Statement() IStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStatementContext)
}

func (s *Body_lineContext) NEWLINE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, 0)
}

func (s *Body_lineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Body_lineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Body_lineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterBody_line(s)
	}
}

func (s *Body_lineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitBody_line(s)
	}
}

func (s *Body_lineContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitBody_line(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Body_line() (localctx IBody_lineContext) {
	localctx = NewBody_lineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, NeuroScriptParserRULE_body_line)
	p.SetState(206)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_CALL, NeuroScriptParserKW_SET, NeuroScriptParserKW_RETURN, NeuroScriptParserKW_EMIT, NeuroScriptParserKW_IF, NeuroScriptParserKW_WHILE, NeuroScriptParserKW_FOR, NeuroScriptParserKW_ON_ERROR, NeuroScriptParserKW_CLEAR_ERROR, NeuroScriptParserKW_MUST, NeuroScriptParserKW_MUSTBE, NeuroScriptParserKW_FAIL, NeuroScriptParserKW_ASK, NeuroScriptParserKW_BREAK, NeuroScriptParserKW_CONTINUE:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(202)
			p.Statement()
		}
		{
			p.SetState(203)
			p.Match(NeuroScriptParserNEWLINE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserNEWLINE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(205)
			p.Match(NeuroScriptParserNEWLINE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IStatementContext is an interface to support dynamic dispatch.
type IStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Simple_statement() ISimple_statementContext
	Block_statement() IBlock_statementContext

	// IsStatementContext differentiates from other interfaces.
	IsStatementContext()
}

type StatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStatementContext() *StatementContext {
	var p = new(StatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_statement
	return p
}

func InitEmptyStatementContext(p *StatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_statement
}

func (*StatementContext) IsStatementContext() {}

func NewStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StatementContext {
	var p = new(StatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_statement

	return p
}

func (s *StatementContext) GetParser() antlr.Parser { return s.parser }

func (s *StatementContext) Simple_statement() ISimple_statementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISimple_statementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISimple_statementContext)
}

func (s *StatementContext) Block_statement() IBlock_statementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBlock_statementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBlock_statementContext)
}

func (s *StatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StatementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterStatement(s)
	}
}

func (s *StatementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitStatement(s)
	}
}

func (s *StatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Statement() (localctx IStatementContext) {
	localctx = NewStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, NeuroScriptParserRULE_statement)
	p.SetState(210)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_CALL, NeuroScriptParserKW_SET, NeuroScriptParserKW_RETURN, NeuroScriptParserKW_EMIT, NeuroScriptParserKW_CLEAR_ERROR, NeuroScriptParserKW_MUST, NeuroScriptParserKW_MUSTBE, NeuroScriptParserKW_FAIL, NeuroScriptParserKW_ASK, NeuroScriptParserKW_BREAK, NeuroScriptParserKW_CONTINUE:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(208)
			p.Simple_statement()
		}

	case NeuroScriptParserKW_IF, NeuroScriptParserKW_WHILE, NeuroScriptParserKW_FOR, NeuroScriptParserKW_ON_ERROR:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(209)
			p.Block_statement()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISimple_statementContext is an interface to support dynamic dispatch.
type ISimple_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Set_statement() ISet_statementContext
	Call_statement() ICall_statementContext
	Return_statement() IReturn_statementContext
	Emit_statement() IEmit_statementContext
	Must_statement() IMust_statementContext
	Fail_statement() IFail_statementContext
	ClearErrorStmt() IClearErrorStmtContext
	Ask_stmt() IAsk_stmtContext
	Break_statement() IBreak_statementContext
	Continue_statement() IContinue_statementContext

	// IsSimple_statementContext differentiates from other interfaces.
	IsSimple_statementContext()
}

type Simple_statementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySimple_statementContext() *Simple_statementContext {
	var p = new(Simple_statementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_simple_statement
	return p
}

func InitEmptySimple_statementContext(p *Simple_statementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_simple_statement
}

func (*Simple_statementContext) IsSimple_statementContext() {}

func NewSimple_statementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Simple_statementContext {
	var p = new(Simple_statementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_simple_statement

	return p
}

func (s *Simple_statementContext) GetParser() antlr.Parser { return s.parser }

func (s *Simple_statementContext) Set_statement() ISet_statementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISet_statementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISet_statementContext)
}

func (s *Simple_statementContext) Call_statement() ICall_statementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICall_statementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICall_statementContext)
}

func (s *Simple_statementContext) Return_statement() IReturn_statementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReturn_statementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReturn_statementContext)
}

func (s *Simple_statementContext) Emit_statement() IEmit_statementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IEmit_statementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IEmit_statementContext)
}

func (s *Simple_statementContext) Must_statement() IMust_statementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMust_statementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMust_statementContext)
}

func (s *Simple_statementContext) Fail_statement() IFail_statementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFail_statementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFail_statementContext)
}

func (s *Simple_statementContext) ClearErrorStmt() IClearErrorStmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IClearErrorStmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IClearErrorStmtContext)
}

func (s *Simple_statementContext) Ask_stmt() IAsk_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAsk_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAsk_stmtContext)
}

func (s *Simple_statementContext) Break_statement() IBreak_statementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBreak_statementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBreak_statementContext)
}

func (s *Simple_statementContext) Continue_statement() IContinue_statementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IContinue_statementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IContinue_statementContext)
}

func (s *Simple_statementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Simple_statementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Simple_statementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterSimple_statement(s)
	}
}

func (s *Simple_statementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitSimple_statement(s)
	}
}

func (s *Simple_statementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitSimple_statement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Simple_statement() (localctx ISimple_statementContext) {
	localctx = NewSimple_statementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, NeuroScriptParserRULE_simple_statement)
	p.SetState(222)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_SET:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(212)
			p.Set_statement()
		}

	case NeuroScriptParserKW_CALL:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(213)
			p.Call_statement()
		}

	case NeuroScriptParserKW_RETURN:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(214)
			p.Return_statement()
		}

	case NeuroScriptParserKW_EMIT:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(215)
			p.Emit_statement()
		}

	case NeuroScriptParserKW_MUST, NeuroScriptParserKW_MUSTBE:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(216)
			p.Must_statement()
		}

	case NeuroScriptParserKW_FAIL:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(217)
			p.Fail_statement()
		}

	case NeuroScriptParserKW_CLEAR_ERROR:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(218)
			p.ClearErrorStmt()
		}

	case NeuroScriptParserKW_ASK:
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(219)
			p.Ask_stmt()
		}

	case NeuroScriptParserKW_BREAK:
		p.EnterOuterAlt(localctx, 9)
		{
			p.SetState(220)
			p.Break_statement()
		}

	case NeuroScriptParserKW_CONTINUE:
		p.EnterOuterAlt(localctx, 10)
		{
			p.SetState(221)
			p.Continue_statement()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBlock_statementContext is an interface to support dynamic dispatch.
type IBlock_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	If_statement() IIf_statementContext
	While_statement() IWhile_statementContext
	For_each_statement() IFor_each_statementContext
	OnErrorStmt() IOnErrorStmtContext

	// IsBlock_statementContext differentiates from other interfaces.
	IsBlock_statementContext()
}

type Block_statementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBlock_statementContext() *Block_statementContext {
	var p = new(Block_statementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_block_statement
	return p
}

func InitEmptyBlock_statementContext(p *Block_statementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_block_statement
}

func (*Block_statementContext) IsBlock_statementContext() {}

func NewBlock_statementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Block_statementContext {
	var p = new(Block_statementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_block_statement

	return p
}

func (s *Block_statementContext) GetParser() antlr.Parser { return s.parser }

func (s *Block_statementContext) If_statement() IIf_statementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIf_statementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIf_statementContext)
}

func (s *Block_statementContext) While_statement() IWhile_statementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IWhile_statementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IWhile_statementContext)
}

func (s *Block_statementContext) For_each_statement() IFor_each_statementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFor_each_statementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFor_each_statementContext)
}

func (s *Block_statementContext) OnErrorStmt() IOnErrorStmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOnErrorStmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOnErrorStmtContext)
}

func (s *Block_statementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Block_statementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Block_statementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterBlock_statement(s)
	}
}

func (s *Block_statementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitBlock_statement(s)
	}
}

func (s *Block_statementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitBlock_statement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Block_statement() (localctx IBlock_statementContext) {
	localctx = NewBlock_statementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, NeuroScriptParserRULE_block_statement)
	p.SetState(228)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_IF:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(224)
			p.If_statement()
		}

	case NeuroScriptParserKW_WHILE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(225)
			p.While_statement()
		}

	case NeuroScriptParserKW_FOR:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(226)
			p.For_each_statement()
		}

	case NeuroScriptParserKW_ON_ERROR:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(227)
			p.OnErrorStmt()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExpressionStatementContext is an interface to support dynamic dispatch.
type IExpressionStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Expression() IExpressionContext

	// IsExpressionStatementContext differentiates from other interfaces.
	IsExpressionStatementContext()
}

type ExpressionStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionStatementContext() *ExpressionStatementContext {
	var p = new(ExpressionStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_expressionStatement
	return p
}

func InitEmptyExpressionStatementContext(p *ExpressionStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_expressionStatement
}

func (*ExpressionStatementContext) IsExpressionStatementContext() {}

func NewExpressionStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionStatementContext {
	var p = new(ExpressionStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_expressionStatement

	return p
}

func (s *ExpressionStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionStatementContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ExpressionStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExpressionStatementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterExpressionStatement(s)
	}
}

func (s *ExpressionStatementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitExpressionStatement(s)
	}
}

func (s *ExpressionStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitExpressionStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) ExpressionStatement() (localctx IExpressionStatementContext) {
	localctx = NewExpressionStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, NeuroScriptParserRULE_expressionStatement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(230)
		p.Expression()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISet_statementContext is an interface to support dynamic dispatch.
type ISet_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_SET() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode
	ASSIGN() antlr.TerminalNode
	Expression() IExpressionContext

	// IsSet_statementContext differentiates from other interfaces.
	IsSet_statementContext()
}

type Set_statementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySet_statementContext() *Set_statementContext {
	var p = new(Set_statementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_set_statement
	return p
}

func InitEmptySet_statementContext(p *Set_statementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_set_statement
}

func (*Set_statementContext) IsSet_statementContext() {}

func NewSet_statementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Set_statementContext {
	var p = new(Set_statementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_set_statement

	return p
}

func (s *Set_statementContext) GetParser() antlr.Parser { return s.parser }

func (s *Set_statementContext) KW_SET() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_SET, 0)
}

func (s *Set_statementContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserIDENTIFIER, 0)
}

func (s *Set_statementContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserASSIGN, 0)
}

func (s *Set_statementContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *Set_statementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Set_statementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Set_statementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterSet_statement(s)
	}
}

func (s *Set_statementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitSet_statement(s)
	}
}

func (s *Set_statementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitSet_statement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Set_statement() (localctx ISet_statementContext) {
	localctx = NewSet_statementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, NeuroScriptParserRULE_set_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(232)
		p.Match(NeuroScriptParserKW_SET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(233)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(234)
		p.Match(NeuroScriptParserASSIGN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(235)
		p.Expression()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICall_statementContext is an interface to support dynamic dispatch.
type ICall_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_CALL() antlr.TerminalNode
	Callable_expr() ICallable_exprContext

	// IsCall_statementContext differentiates from other interfaces.
	IsCall_statementContext()
}

type Call_statementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCall_statementContext() *Call_statementContext {
	var p = new(Call_statementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_call_statement
	return p
}

func InitEmptyCall_statementContext(p *Call_statementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_call_statement
}

func (*Call_statementContext) IsCall_statementContext() {}

func NewCall_statementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Call_statementContext {
	var p = new(Call_statementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_call_statement

	return p
}

func (s *Call_statementContext) GetParser() antlr.Parser { return s.parser }

func (s *Call_statementContext) KW_CALL() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_CALL, 0)
}

func (s *Call_statementContext) Callable_expr() ICallable_exprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICallable_exprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICallable_exprContext)
}

func (s *Call_statementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Call_statementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Call_statementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterCall_statement(s)
	}
}

func (s *Call_statementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitCall_statement(s)
	}
}

func (s *Call_statementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitCall_statement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Call_statement() (localctx ICall_statementContext) {
	localctx = NewCall_statementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 32, NeuroScriptParserRULE_call_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(237)
		p.Match(NeuroScriptParserKW_CALL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(238)
		p.Callable_expr()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IReturn_statementContext is an interface to support dynamic dispatch.
type IReturn_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_RETURN() antlr.TerminalNode
	Expression_list() IExpression_listContext

	// IsReturn_statementContext differentiates from other interfaces.
	IsReturn_statementContext()
}

type Return_statementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyReturn_statementContext() *Return_statementContext {
	var p = new(Return_statementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_return_statement
	return p
}

func InitEmptyReturn_statementContext(p *Return_statementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_return_statement
}

func (*Return_statementContext) IsReturn_statementContext() {}

func NewReturn_statementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Return_statementContext {
	var p = new(Return_statementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_return_statement

	return p
}

func (s *Return_statementContext) GetParser() antlr.Parser { return s.parser }

func (s *Return_statementContext) KW_RETURN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_RETURN, 0)
}

func (s *Return_statementContext) Expression_list() IExpression_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpression_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpression_listContext)
}

func (s *Return_statementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Return_statementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Return_statementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterReturn_statement(s)
	}
}

func (s *Return_statementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitReturn_statement(s)
	}
}

func (s *Return_statementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitReturn_statement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Return_statement() (localctx IReturn_statementContext) {
	localctx = NewReturn_statementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, NeuroScriptParserRULE_return_statement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(240)
		p.Match(NeuroScriptParserKW_RETURN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(242)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64((_la-26)) & ^0x3f) == 0 && ((int64(1)<<(_la-26))&36180942607613567) != 0 {
		{
			p.SetState(241)
			p.Expression_list()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IEmit_statementContext is an interface to support dynamic dispatch.
type IEmit_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_EMIT() antlr.TerminalNode
	Expression() IExpressionContext

	// IsEmit_statementContext differentiates from other interfaces.
	IsEmit_statementContext()
}

type Emit_statementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyEmit_statementContext() *Emit_statementContext {
	var p = new(Emit_statementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_emit_statement
	return p
}

func InitEmptyEmit_statementContext(p *Emit_statementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_emit_statement
}

func (*Emit_statementContext) IsEmit_statementContext() {}

func NewEmit_statementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Emit_statementContext {
	var p = new(Emit_statementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_emit_statement

	return p
}

func (s *Emit_statementContext) GetParser() antlr.Parser { return s.parser }

func (s *Emit_statementContext) KW_EMIT() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_EMIT, 0)
}

func (s *Emit_statementContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *Emit_statementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Emit_statementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Emit_statementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterEmit_statement(s)
	}
}

func (s *Emit_statementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitEmit_statement(s)
	}
}

func (s *Emit_statementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitEmit_statement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Emit_statement() (localctx IEmit_statementContext) {
	localctx = NewEmit_statementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 36, NeuroScriptParserRULE_emit_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(244)
		p.Match(NeuroScriptParserKW_EMIT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(245)
		p.Expression()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMust_statementContext is an interface to support dynamic dispatch.
type IMust_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_MUST() antlr.TerminalNode
	Expression() IExpressionContext
	KW_MUSTBE() antlr.TerminalNode
	Callable_expr() ICallable_exprContext

	// IsMust_statementContext differentiates from other interfaces.
	IsMust_statementContext()
}

type Must_statementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMust_statementContext() *Must_statementContext {
	var p = new(Must_statementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_must_statement
	return p
}

func InitEmptyMust_statementContext(p *Must_statementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_must_statement
}

func (*Must_statementContext) IsMust_statementContext() {}

func NewMust_statementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Must_statementContext {
	var p = new(Must_statementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_must_statement

	return p
}

func (s *Must_statementContext) GetParser() antlr.Parser { return s.parser }

func (s *Must_statementContext) KW_MUST() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_MUST, 0)
}

func (s *Must_statementContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *Must_statementContext) KW_MUSTBE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_MUSTBE, 0)
}

func (s *Must_statementContext) Callable_expr() ICallable_exprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICallable_exprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICallable_exprContext)
}

func (s *Must_statementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Must_statementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Must_statementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterMust_statement(s)
	}
}

func (s *Must_statementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitMust_statement(s)
	}
}

func (s *Must_statementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitMust_statement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Must_statement() (localctx IMust_statementContext) {
	localctx = NewMust_statementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, NeuroScriptParserRULE_must_statement)
	p.SetState(251)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_MUST:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(247)
			p.Match(NeuroScriptParserKW_MUST)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(248)
			p.Expression()
		}

	case NeuroScriptParserKW_MUSTBE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(249)
			p.Match(NeuroScriptParserKW_MUSTBE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(250)
			p.Callable_expr()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFail_statementContext is an interface to support dynamic dispatch.
type IFail_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_FAIL() antlr.TerminalNode
	Expression() IExpressionContext

	// IsFail_statementContext differentiates from other interfaces.
	IsFail_statementContext()
}

type Fail_statementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFail_statementContext() *Fail_statementContext {
	var p = new(Fail_statementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_fail_statement
	return p
}

func InitEmptyFail_statementContext(p *Fail_statementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_fail_statement
}

func (*Fail_statementContext) IsFail_statementContext() {}

func NewFail_statementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Fail_statementContext {
	var p = new(Fail_statementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_fail_statement

	return p
}

func (s *Fail_statementContext) GetParser() antlr.Parser { return s.parser }

func (s *Fail_statementContext) KW_FAIL() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_FAIL, 0)
}

func (s *Fail_statementContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *Fail_statementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Fail_statementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Fail_statementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterFail_statement(s)
	}
}

func (s *Fail_statementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitFail_statement(s)
	}
}

func (s *Fail_statementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitFail_statement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Fail_statement() (localctx IFail_statementContext) {
	localctx = NewFail_statementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 40, NeuroScriptParserRULE_fail_statement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(253)
		p.Match(NeuroScriptParserKW_FAIL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(255)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64((_la-26)) & ^0x3f) == 0 && ((int64(1)<<(_la-26))&36180942607613567) != 0 {
		{
			p.SetState(254)
			p.Expression()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IClearErrorStmtContext is an interface to support dynamic dispatch.
type IClearErrorStmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_CLEAR_ERROR() antlr.TerminalNode

	// IsClearErrorStmtContext differentiates from other interfaces.
	IsClearErrorStmtContext()
}

type ClearErrorStmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyClearErrorStmtContext() *ClearErrorStmtContext {
	var p = new(ClearErrorStmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_clearErrorStmt
	return p
}

func InitEmptyClearErrorStmtContext(p *ClearErrorStmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_clearErrorStmt
}

func (*ClearErrorStmtContext) IsClearErrorStmtContext() {}

func NewClearErrorStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ClearErrorStmtContext {
	var p = new(ClearErrorStmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_clearErrorStmt

	return p
}

func (s *ClearErrorStmtContext) GetParser() antlr.Parser { return s.parser }

func (s *ClearErrorStmtContext) KW_CLEAR_ERROR() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_CLEAR_ERROR, 0)
}

func (s *ClearErrorStmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ClearErrorStmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ClearErrorStmtContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterClearErrorStmt(s)
	}
}

func (s *ClearErrorStmtContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitClearErrorStmt(s)
	}
}

func (s *ClearErrorStmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitClearErrorStmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) ClearErrorStmt() (localctx IClearErrorStmtContext) {
	localctx = NewClearErrorStmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 42, NeuroScriptParserRULE_clearErrorStmt)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(257)
		p.Match(NeuroScriptParserKW_CLEAR_ERROR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAsk_stmtContext is an interface to support dynamic dispatch.
type IAsk_stmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_ASK() antlr.TerminalNode
	Expression() IExpressionContext
	KW_INTO() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode

	// IsAsk_stmtContext differentiates from other interfaces.
	IsAsk_stmtContext()
}

type Ask_stmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAsk_stmtContext() *Ask_stmtContext {
	var p = new(Ask_stmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_ask_stmt
	return p
}

func InitEmptyAsk_stmtContext(p *Ask_stmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_ask_stmt
}

func (*Ask_stmtContext) IsAsk_stmtContext() {}

func NewAsk_stmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Ask_stmtContext {
	var p = new(Ask_stmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_ask_stmt

	return p
}

func (s *Ask_stmtContext) GetParser() antlr.Parser { return s.parser }

func (s *Ask_stmtContext) KW_ASK() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ASK, 0)
}

func (s *Ask_stmtContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *Ask_stmtContext) KW_INTO() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_INTO, 0)
}

func (s *Ask_stmtContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserIDENTIFIER, 0)
}

func (s *Ask_stmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Ask_stmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Ask_stmtContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterAsk_stmt(s)
	}
}

func (s *Ask_stmtContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitAsk_stmt(s)
	}
}

func (s *Ask_stmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitAsk_stmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Ask_stmt() (localctx IAsk_stmtContext) {
	localctx = NewAsk_stmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 44, NeuroScriptParserRULE_ask_stmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(259)
		p.Match(NeuroScriptParserKW_ASK)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(260)
		p.Expression()
	}
	p.SetState(263)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NeuroScriptParserKW_INTO {
		{
			p.SetState(261)
			p.Match(NeuroScriptParserKW_INTO)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(262)
			p.Match(NeuroScriptParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBreak_statementContext is an interface to support dynamic dispatch.
type IBreak_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_BREAK() antlr.TerminalNode

	// IsBreak_statementContext differentiates from other interfaces.
	IsBreak_statementContext()
}

type Break_statementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBreak_statementContext() *Break_statementContext {
	var p = new(Break_statementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_break_statement
	return p
}

func InitEmptyBreak_statementContext(p *Break_statementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_break_statement
}

func (*Break_statementContext) IsBreak_statementContext() {}

func NewBreak_statementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Break_statementContext {
	var p = new(Break_statementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_break_statement

	return p
}

func (s *Break_statementContext) GetParser() antlr.Parser { return s.parser }

func (s *Break_statementContext) KW_BREAK() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_BREAK, 0)
}

func (s *Break_statementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Break_statementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Break_statementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterBreak_statement(s)
	}
}

func (s *Break_statementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitBreak_statement(s)
	}
}

func (s *Break_statementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitBreak_statement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Break_statement() (localctx IBreak_statementContext) {
	localctx = NewBreak_statementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 46, NeuroScriptParserRULE_break_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(265)
		p.Match(NeuroScriptParserKW_BREAK)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IContinue_statementContext is an interface to support dynamic dispatch.
type IContinue_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_CONTINUE() antlr.TerminalNode

	// IsContinue_statementContext differentiates from other interfaces.
	IsContinue_statementContext()
}

type Continue_statementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyContinue_statementContext() *Continue_statementContext {
	var p = new(Continue_statementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_continue_statement
	return p
}

func InitEmptyContinue_statementContext(p *Continue_statementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_continue_statement
}

func (*Continue_statementContext) IsContinue_statementContext() {}

func NewContinue_statementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Continue_statementContext {
	var p = new(Continue_statementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_continue_statement

	return p
}

func (s *Continue_statementContext) GetParser() antlr.Parser { return s.parser }

func (s *Continue_statementContext) KW_CONTINUE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_CONTINUE, 0)
}

func (s *Continue_statementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Continue_statementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Continue_statementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterContinue_statement(s)
	}
}

func (s *Continue_statementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitContinue_statement(s)
	}
}

func (s *Continue_statementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitContinue_statement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Continue_statement() (localctx IContinue_statementContext) {
	localctx = NewContinue_statementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 48, NeuroScriptParserRULE_continue_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(267)
		p.Match(NeuroScriptParserKW_CONTINUE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIf_statementContext is an interface to support dynamic dispatch.
type IIf_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_IF() antlr.TerminalNode
	Expression() IExpressionContext
	AllNEWLINE() []antlr.TerminalNode
	NEWLINE(i int) antlr.TerminalNode
	AllStatement_list() []IStatement_listContext
	Statement_list(i int) IStatement_listContext
	KW_ENDIF() antlr.TerminalNode
	KW_ELSE() antlr.TerminalNode

	// IsIf_statementContext differentiates from other interfaces.
	IsIf_statementContext()
}

type If_statementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIf_statementContext() *If_statementContext {
	var p = new(If_statementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_if_statement
	return p
}

func InitEmptyIf_statementContext(p *If_statementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_if_statement
}

func (*If_statementContext) IsIf_statementContext() {}

func NewIf_statementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *If_statementContext {
	var p = new(If_statementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_if_statement

	return p
}

func (s *If_statementContext) GetParser() antlr.Parser { return s.parser }

func (s *If_statementContext) KW_IF() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_IF, 0)
}

func (s *If_statementContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *If_statementContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserNEWLINE)
}

func (s *If_statementContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, i)
}

func (s *If_statementContext) AllStatement_list() []IStatement_listContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStatement_listContext); ok {
			len++
		}
	}

	tst := make([]IStatement_listContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStatement_listContext); ok {
			tst[i] = t.(IStatement_listContext)
			i++
		}
	}

	return tst
}

func (s *If_statementContext) Statement_list(i int) IStatement_listContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStatement_listContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStatement_listContext)
}

func (s *If_statementContext) KW_ENDIF() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ENDIF, 0)
}

func (s *If_statementContext) KW_ELSE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ELSE, 0)
}

func (s *If_statementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *If_statementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *If_statementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterIf_statement(s)
	}
}

func (s *If_statementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitIf_statement(s)
	}
}

func (s *If_statementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitIf_statement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) If_statement() (localctx IIf_statementContext) {
	localctx = NewIf_statementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 50, NeuroScriptParserRULE_if_statement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(269)
		p.Match(NeuroScriptParserKW_IF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(270)
		p.Expression()
	}
	{
		p.SetState(271)
		p.Match(NeuroScriptParserNEWLINE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(272)
		p.Statement_list()
	}
	p.SetState(276)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NeuroScriptParserKW_ELSE {
		{
			p.SetState(273)
			p.Match(NeuroScriptParserKW_ELSE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(274)
			p.Match(NeuroScriptParserNEWLINE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(275)
			p.Statement_list()
		}

	}
	{
		p.SetState(278)
		p.Match(NeuroScriptParserKW_ENDIF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IWhile_statementContext is an interface to support dynamic dispatch.
type IWhile_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_WHILE() antlr.TerminalNode
	Expression() IExpressionContext
	NEWLINE() antlr.TerminalNode
	Statement_list() IStatement_listContext
	KW_ENDWHILE() antlr.TerminalNode

	// IsWhile_statementContext differentiates from other interfaces.
	IsWhile_statementContext()
}

type While_statementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyWhile_statementContext() *While_statementContext {
	var p = new(While_statementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_while_statement
	return p
}

func InitEmptyWhile_statementContext(p *While_statementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_while_statement
}

func (*While_statementContext) IsWhile_statementContext() {}

func NewWhile_statementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *While_statementContext {
	var p = new(While_statementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_while_statement

	return p
}

func (s *While_statementContext) GetParser() antlr.Parser { return s.parser }

func (s *While_statementContext) KW_WHILE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_WHILE, 0)
}

func (s *While_statementContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *While_statementContext) NEWLINE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, 0)
}

func (s *While_statementContext) Statement_list() IStatement_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStatement_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStatement_listContext)
}

func (s *While_statementContext) KW_ENDWHILE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ENDWHILE, 0)
}

func (s *While_statementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *While_statementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *While_statementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterWhile_statement(s)
	}
}

func (s *While_statementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitWhile_statement(s)
	}
}

func (s *While_statementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitWhile_statement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) While_statement() (localctx IWhile_statementContext) {
	localctx = NewWhile_statementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 52, NeuroScriptParserRULE_while_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(280)
		p.Match(NeuroScriptParserKW_WHILE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(281)
		p.Expression()
	}
	{
		p.SetState(282)
		p.Match(NeuroScriptParserNEWLINE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(283)
		p.Statement_list()
	}
	{
		p.SetState(284)
		p.Match(NeuroScriptParserKW_ENDWHILE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFor_each_statementContext is an interface to support dynamic dispatch.
type IFor_each_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_FOR() antlr.TerminalNode
	KW_EACH() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode
	KW_IN() antlr.TerminalNode
	Expression() IExpressionContext
	NEWLINE() antlr.TerminalNode
	Statement_list() IStatement_listContext
	KW_ENDFOR() antlr.TerminalNode

	// IsFor_each_statementContext differentiates from other interfaces.
	IsFor_each_statementContext()
}

type For_each_statementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFor_each_statementContext() *For_each_statementContext {
	var p = new(For_each_statementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_for_each_statement
	return p
}

func InitEmptyFor_each_statementContext(p *For_each_statementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_for_each_statement
}

func (*For_each_statementContext) IsFor_each_statementContext() {}

func NewFor_each_statementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *For_each_statementContext {
	var p = new(For_each_statementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_for_each_statement

	return p
}

func (s *For_each_statementContext) GetParser() antlr.Parser { return s.parser }

func (s *For_each_statementContext) KW_FOR() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_FOR, 0)
}

func (s *For_each_statementContext) KW_EACH() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_EACH, 0)
}

func (s *For_each_statementContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserIDENTIFIER, 0)
}

func (s *For_each_statementContext) KW_IN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_IN, 0)
}

func (s *For_each_statementContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *For_each_statementContext) NEWLINE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, 0)
}

func (s *For_each_statementContext) Statement_list() IStatement_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStatement_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStatement_listContext)
}

func (s *For_each_statementContext) KW_ENDFOR() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ENDFOR, 0)
}

func (s *For_each_statementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *For_each_statementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *For_each_statementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterFor_each_statement(s)
	}
}

func (s *For_each_statementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitFor_each_statement(s)
	}
}

func (s *For_each_statementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitFor_each_statement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) For_each_statement() (localctx IFor_each_statementContext) {
	localctx = NewFor_each_statementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 54, NeuroScriptParserRULE_for_each_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(286)
		p.Match(NeuroScriptParserKW_FOR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(287)
		p.Match(NeuroScriptParserKW_EACH)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(288)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(289)
		p.Match(NeuroScriptParserKW_IN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(290)
		p.Expression()
	}
	{
		p.SetState(291)
		p.Match(NeuroScriptParserNEWLINE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(292)
		p.Statement_list()
	}
	{
		p.SetState(293)
		p.Match(NeuroScriptParserKW_ENDFOR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOnErrorStmtContext is an interface to support dynamic dispatch.
type IOnErrorStmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_ON_ERROR() antlr.TerminalNode
	KW_MEANS() antlr.TerminalNode
	NEWLINE() antlr.TerminalNode
	Statement_list() IStatement_listContext
	KW_ENDON() antlr.TerminalNode

	// IsOnErrorStmtContext differentiates from other interfaces.
	IsOnErrorStmtContext()
}

type OnErrorStmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOnErrorStmtContext() *OnErrorStmtContext {
	var p = new(OnErrorStmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_onErrorStmt
	return p
}

func InitEmptyOnErrorStmtContext(p *OnErrorStmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_onErrorStmt
}

func (*OnErrorStmtContext) IsOnErrorStmtContext() {}

func NewOnErrorStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OnErrorStmtContext {
	var p = new(OnErrorStmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_onErrorStmt

	return p
}

func (s *OnErrorStmtContext) GetParser() antlr.Parser { return s.parser }

func (s *OnErrorStmtContext) KW_ON_ERROR() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ON_ERROR, 0)
}

func (s *OnErrorStmtContext) KW_MEANS() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_MEANS, 0)
}

func (s *OnErrorStmtContext) NEWLINE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, 0)
}

func (s *OnErrorStmtContext) Statement_list() IStatement_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStatement_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStatement_listContext)
}

func (s *OnErrorStmtContext) KW_ENDON() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ENDON, 0)
}

func (s *OnErrorStmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OnErrorStmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OnErrorStmtContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterOnErrorStmt(s)
	}
}

func (s *OnErrorStmtContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitOnErrorStmt(s)
	}
}

func (s *OnErrorStmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitOnErrorStmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) OnErrorStmt() (localctx IOnErrorStmtContext) {
	localctx = NewOnErrorStmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 56, NeuroScriptParserRULE_onErrorStmt)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(295)
		p.Match(NeuroScriptParserKW_ON_ERROR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(296)
		p.Match(NeuroScriptParserKW_MEANS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(297)
		p.Match(NeuroScriptParserNEWLINE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(298)
		p.Statement_list()
	}
	{
		p.SetState(299)
		p.Match(NeuroScriptParserKW_ENDON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IQualified_identifierContext is an interface to support dynamic dispatch.
type IQualified_identifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIDENTIFIER() []antlr.TerminalNode
	IDENTIFIER(i int) antlr.TerminalNode
	AllDOT() []antlr.TerminalNode
	DOT(i int) antlr.TerminalNode

	// IsQualified_identifierContext differentiates from other interfaces.
	IsQualified_identifierContext()
}

type Qualified_identifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyQualified_identifierContext() *Qualified_identifierContext {
	var p = new(Qualified_identifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_qualified_identifier
	return p
}

func InitEmptyQualified_identifierContext(p *Qualified_identifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_qualified_identifier
}

func (*Qualified_identifierContext) IsQualified_identifierContext() {}

func NewQualified_identifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Qualified_identifierContext {
	var p = new(Qualified_identifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_qualified_identifier

	return p
}

func (s *Qualified_identifierContext) GetParser() antlr.Parser { return s.parser }

func (s *Qualified_identifierContext) AllIDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserIDENTIFIER)
}

func (s *Qualified_identifierContext) IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserIDENTIFIER, i)
}

func (s *Qualified_identifierContext) AllDOT() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserDOT)
}

func (s *Qualified_identifierContext) DOT(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserDOT, i)
}

func (s *Qualified_identifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Qualified_identifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Qualified_identifierContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterQualified_identifier(s)
	}
}

func (s *Qualified_identifierContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitQualified_identifier(s)
	}
}

func (s *Qualified_identifierContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitQualified_identifier(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Qualified_identifier() (localctx IQualified_identifierContext) {
	localctx = NewQualified_identifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 58, NeuroScriptParserRULE_qualified_identifier)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(301)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(306)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserDOT {
		{
			p.SetState(302)
			p.Match(NeuroScriptParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(303)
			p.Match(NeuroScriptParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(308)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICall_targetContext is an interface to support dynamic dispatch.
type ICall_targetContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	KW_TOOL() antlr.TerminalNode
	DOT() antlr.TerminalNode
	Qualified_identifier() IQualified_identifierContext

	// IsCall_targetContext differentiates from other interfaces.
	IsCall_targetContext()
}

type Call_targetContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCall_targetContext() *Call_targetContext {
	var p = new(Call_targetContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_call_target
	return p
}

func InitEmptyCall_targetContext(p *Call_targetContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_call_target
}

func (*Call_targetContext) IsCall_targetContext() {}

func NewCall_targetContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Call_targetContext {
	var p = new(Call_targetContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_call_target

	return p
}

func (s *Call_targetContext) GetParser() antlr.Parser { return s.parser }

func (s *Call_targetContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserIDENTIFIER, 0)
}

func (s *Call_targetContext) KW_TOOL() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_TOOL, 0)
}

func (s *Call_targetContext) DOT() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserDOT, 0)
}

func (s *Call_targetContext) Qualified_identifier() IQualified_identifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IQualified_identifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IQualified_identifierContext)
}

func (s *Call_targetContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Call_targetContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Call_targetContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterCall_target(s)
	}
}

func (s *Call_targetContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitCall_target(s)
	}
}

func (s *Call_targetContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitCall_target(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Call_target() (localctx ICall_targetContext) {
	localctx = NewCall_targetContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 60, NeuroScriptParserRULE_call_target)
	p.SetState(313)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(309)
			p.Match(NeuroScriptParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_TOOL:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(310)
			p.Match(NeuroScriptParserKW_TOOL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(311)
			p.Match(NeuroScriptParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(312)
			p.Qualified_identifier()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Logical_or_expr() ILogical_or_exprContext

	// IsExpressionContext differentiates from other interfaces.
	IsExpressionContext()
}

type ExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_expression
	return p
}

func InitEmptyExpressionContext(p *ExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_expression
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) Logical_or_expr() ILogical_or_exprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILogical_or_exprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILogical_or_exprContext)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterExpression(s)
	}
}

func (s *ExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitExpression(s)
	}
}

func (s *ExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Expression() (localctx IExpressionContext) {
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 62, NeuroScriptParserRULE_expression)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(315)
		p.Logical_or_expr()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILogical_or_exprContext is an interface to support dynamic dispatch.
type ILogical_or_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllLogical_and_expr() []ILogical_and_exprContext
	Logical_and_expr(i int) ILogical_and_exprContext
	AllKW_OR() []antlr.TerminalNode
	KW_OR(i int) antlr.TerminalNode

	// IsLogical_or_exprContext differentiates from other interfaces.
	IsLogical_or_exprContext()
}

type Logical_or_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLogical_or_exprContext() *Logical_or_exprContext {
	var p = new(Logical_or_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_logical_or_expr
	return p
}

func InitEmptyLogical_or_exprContext(p *Logical_or_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_logical_or_expr
}

func (*Logical_or_exprContext) IsLogical_or_exprContext() {}

func NewLogical_or_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Logical_or_exprContext {
	var p = new(Logical_or_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_logical_or_expr

	return p
}

func (s *Logical_or_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Logical_or_exprContext) AllLogical_and_expr() []ILogical_and_exprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILogical_and_exprContext); ok {
			len++
		}
	}

	tst := make([]ILogical_and_exprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILogical_and_exprContext); ok {
			tst[i] = t.(ILogical_and_exprContext)
			i++
		}
	}

	return tst
}

func (s *Logical_or_exprContext) Logical_and_expr(i int) ILogical_and_exprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILogical_and_exprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILogical_and_exprContext)
}

func (s *Logical_or_exprContext) AllKW_OR() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserKW_OR)
}

func (s *Logical_or_exprContext) KW_OR(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_OR, i)
}

func (s *Logical_or_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Logical_or_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Logical_or_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterLogical_or_expr(s)
	}
}

func (s *Logical_or_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitLogical_or_expr(s)
	}
}

func (s *Logical_or_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitLogical_or_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Logical_or_expr() (localctx ILogical_or_exprContext) {
	localctx = NewLogical_or_exprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 64, NeuroScriptParserRULE_logical_or_expr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(317)
		p.Logical_and_expr()
	}
	p.SetState(322)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserKW_OR {
		{
			p.SetState(318)
			p.Match(NeuroScriptParserKW_OR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(319)
			p.Logical_and_expr()
		}

		p.SetState(324)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILogical_and_exprContext is an interface to support dynamic dispatch.
type ILogical_and_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllBitwise_or_expr() []IBitwise_or_exprContext
	Bitwise_or_expr(i int) IBitwise_or_exprContext
	AllKW_AND() []antlr.TerminalNode
	KW_AND(i int) antlr.TerminalNode

	// IsLogical_and_exprContext differentiates from other interfaces.
	IsLogical_and_exprContext()
}

type Logical_and_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLogical_and_exprContext() *Logical_and_exprContext {
	var p = new(Logical_and_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_logical_and_expr
	return p
}

func InitEmptyLogical_and_exprContext(p *Logical_and_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_logical_and_expr
}

func (*Logical_and_exprContext) IsLogical_and_exprContext() {}

func NewLogical_and_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Logical_and_exprContext {
	var p = new(Logical_and_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_logical_and_expr

	return p
}

func (s *Logical_and_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Logical_and_exprContext) AllBitwise_or_expr() []IBitwise_or_exprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IBitwise_or_exprContext); ok {
			len++
		}
	}

	tst := make([]IBitwise_or_exprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IBitwise_or_exprContext); ok {
			tst[i] = t.(IBitwise_or_exprContext)
			i++
		}
	}

	return tst
}

func (s *Logical_and_exprContext) Bitwise_or_expr(i int) IBitwise_or_exprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBitwise_or_exprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBitwise_or_exprContext)
}

func (s *Logical_and_exprContext) AllKW_AND() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserKW_AND)
}

func (s *Logical_and_exprContext) KW_AND(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_AND, i)
}

func (s *Logical_and_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Logical_and_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Logical_and_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterLogical_and_expr(s)
	}
}

func (s *Logical_and_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitLogical_and_expr(s)
	}
}

func (s *Logical_and_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitLogical_and_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Logical_and_expr() (localctx ILogical_and_exprContext) {
	localctx = NewLogical_and_exprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 66, NeuroScriptParserRULE_logical_and_expr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(325)
		p.Bitwise_or_expr()
	}
	p.SetState(330)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserKW_AND {
		{
			p.SetState(326)
			p.Match(NeuroScriptParserKW_AND)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(327)
			p.Bitwise_or_expr()
		}

		p.SetState(332)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBitwise_or_exprContext is an interface to support dynamic dispatch.
type IBitwise_or_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllBitwise_xor_expr() []IBitwise_xor_exprContext
	Bitwise_xor_expr(i int) IBitwise_xor_exprContext
	AllPIPE() []antlr.TerminalNode
	PIPE(i int) antlr.TerminalNode

	// IsBitwise_or_exprContext differentiates from other interfaces.
	IsBitwise_or_exprContext()
}

type Bitwise_or_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBitwise_or_exprContext() *Bitwise_or_exprContext {
	var p = new(Bitwise_or_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_bitwise_or_expr
	return p
}

func InitEmptyBitwise_or_exprContext(p *Bitwise_or_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_bitwise_or_expr
}

func (*Bitwise_or_exprContext) IsBitwise_or_exprContext() {}

func NewBitwise_or_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Bitwise_or_exprContext {
	var p = new(Bitwise_or_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_bitwise_or_expr

	return p
}

func (s *Bitwise_or_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Bitwise_or_exprContext) AllBitwise_xor_expr() []IBitwise_xor_exprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IBitwise_xor_exprContext); ok {
			len++
		}
	}

	tst := make([]IBitwise_xor_exprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IBitwise_xor_exprContext); ok {
			tst[i] = t.(IBitwise_xor_exprContext)
			i++
		}
	}

	return tst
}

func (s *Bitwise_or_exprContext) Bitwise_xor_expr(i int) IBitwise_xor_exprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBitwise_xor_exprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBitwise_xor_exprContext)
}

func (s *Bitwise_or_exprContext) AllPIPE() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserPIPE)
}

func (s *Bitwise_or_exprContext) PIPE(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserPIPE, i)
}

func (s *Bitwise_or_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Bitwise_or_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Bitwise_or_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterBitwise_or_expr(s)
	}
}

func (s *Bitwise_or_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitBitwise_or_expr(s)
	}
}

func (s *Bitwise_or_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitBitwise_or_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Bitwise_or_expr() (localctx IBitwise_or_exprContext) {
	localctx = NewBitwise_or_exprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 68, NeuroScriptParserRULE_bitwise_or_expr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(333)
		p.Bitwise_xor_expr()
	}
	p.SetState(338)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserPIPE {
		{
			p.SetState(334)
			p.Match(NeuroScriptParserPIPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(335)
			p.Bitwise_xor_expr()
		}

		p.SetState(340)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBitwise_xor_exprContext is an interface to support dynamic dispatch.
type IBitwise_xor_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllBitwise_and_expr() []IBitwise_and_exprContext
	Bitwise_and_expr(i int) IBitwise_and_exprContext
	AllCARET() []antlr.TerminalNode
	CARET(i int) antlr.TerminalNode

	// IsBitwise_xor_exprContext differentiates from other interfaces.
	IsBitwise_xor_exprContext()
}

type Bitwise_xor_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBitwise_xor_exprContext() *Bitwise_xor_exprContext {
	var p = new(Bitwise_xor_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_bitwise_xor_expr
	return p
}

func InitEmptyBitwise_xor_exprContext(p *Bitwise_xor_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_bitwise_xor_expr
}

func (*Bitwise_xor_exprContext) IsBitwise_xor_exprContext() {}

func NewBitwise_xor_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Bitwise_xor_exprContext {
	var p = new(Bitwise_xor_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_bitwise_xor_expr

	return p
}

func (s *Bitwise_xor_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Bitwise_xor_exprContext) AllBitwise_and_expr() []IBitwise_and_exprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IBitwise_and_exprContext); ok {
			len++
		}
	}

	tst := make([]IBitwise_and_exprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IBitwise_and_exprContext); ok {
			tst[i] = t.(IBitwise_and_exprContext)
			i++
		}
	}

	return tst
}

func (s *Bitwise_xor_exprContext) Bitwise_and_expr(i int) IBitwise_and_exprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBitwise_and_exprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBitwise_and_exprContext)
}

func (s *Bitwise_xor_exprContext) AllCARET() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserCARET)
}

func (s *Bitwise_xor_exprContext) CARET(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserCARET, i)
}

func (s *Bitwise_xor_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Bitwise_xor_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Bitwise_xor_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterBitwise_xor_expr(s)
	}
}

func (s *Bitwise_xor_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitBitwise_xor_expr(s)
	}
}

func (s *Bitwise_xor_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitBitwise_xor_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Bitwise_xor_expr() (localctx IBitwise_xor_exprContext) {
	localctx = NewBitwise_xor_exprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 70, NeuroScriptParserRULE_bitwise_xor_expr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(341)
		p.Bitwise_and_expr()
	}
	p.SetState(346)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserCARET {
		{
			p.SetState(342)
			p.Match(NeuroScriptParserCARET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(343)
			p.Bitwise_and_expr()
		}

		p.SetState(348)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBitwise_and_exprContext is an interface to support dynamic dispatch.
type IBitwise_and_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllEquality_expr() []IEquality_exprContext
	Equality_expr(i int) IEquality_exprContext
	AllAMPERSAND() []antlr.TerminalNode
	AMPERSAND(i int) antlr.TerminalNode

	// IsBitwise_and_exprContext differentiates from other interfaces.
	IsBitwise_and_exprContext()
}

type Bitwise_and_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBitwise_and_exprContext() *Bitwise_and_exprContext {
	var p = new(Bitwise_and_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_bitwise_and_expr
	return p
}

func InitEmptyBitwise_and_exprContext(p *Bitwise_and_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_bitwise_and_expr
}

func (*Bitwise_and_exprContext) IsBitwise_and_exprContext() {}

func NewBitwise_and_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Bitwise_and_exprContext {
	var p = new(Bitwise_and_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_bitwise_and_expr

	return p
}

func (s *Bitwise_and_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Bitwise_and_exprContext) AllEquality_expr() []IEquality_exprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IEquality_exprContext); ok {
			len++
		}
	}

	tst := make([]IEquality_exprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IEquality_exprContext); ok {
			tst[i] = t.(IEquality_exprContext)
			i++
		}
	}

	return tst
}

func (s *Bitwise_and_exprContext) Equality_expr(i int) IEquality_exprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IEquality_exprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IEquality_exprContext)
}

func (s *Bitwise_and_exprContext) AllAMPERSAND() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserAMPERSAND)
}

func (s *Bitwise_and_exprContext) AMPERSAND(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserAMPERSAND, i)
}

func (s *Bitwise_and_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Bitwise_and_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Bitwise_and_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterBitwise_and_expr(s)
	}
}

func (s *Bitwise_and_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitBitwise_and_expr(s)
	}
}

func (s *Bitwise_and_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitBitwise_and_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Bitwise_and_expr() (localctx IBitwise_and_exprContext) {
	localctx = NewBitwise_and_exprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 72, NeuroScriptParserRULE_bitwise_and_expr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(349)
		p.Equality_expr()
	}
	p.SetState(354)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserAMPERSAND {
		{
			p.SetState(350)
			p.Match(NeuroScriptParserAMPERSAND)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(351)
			p.Equality_expr()
		}

		p.SetState(356)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IEquality_exprContext is an interface to support dynamic dispatch.
type IEquality_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllRelational_expr() []IRelational_exprContext
	Relational_expr(i int) IRelational_exprContext
	AllEQ() []antlr.TerminalNode
	EQ(i int) antlr.TerminalNode
	AllNEQ() []antlr.TerminalNode
	NEQ(i int) antlr.TerminalNode

	// IsEquality_exprContext differentiates from other interfaces.
	IsEquality_exprContext()
}

type Equality_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyEquality_exprContext() *Equality_exprContext {
	var p = new(Equality_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_equality_expr
	return p
}

func InitEmptyEquality_exprContext(p *Equality_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_equality_expr
}

func (*Equality_exprContext) IsEquality_exprContext() {}

func NewEquality_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Equality_exprContext {
	var p = new(Equality_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_equality_expr

	return p
}

func (s *Equality_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Equality_exprContext) AllRelational_expr() []IRelational_exprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IRelational_exprContext); ok {
			len++
		}
	}

	tst := make([]IRelational_exprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IRelational_exprContext); ok {
			tst[i] = t.(IRelational_exprContext)
			i++
		}
	}

	return tst
}

func (s *Equality_exprContext) Relational_expr(i int) IRelational_exprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRelational_exprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRelational_exprContext)
}

func (s *Equality_exprContext) AllEQ() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserEQ)
}

func (s *Equality_exprContext) EQ(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserEQ, i)
}

func (s *Equality_exprContext) AllNEQ() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserNEQ)
}

func (s *Equality_exprContext) NEQ(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEQ, i)
}

func (s *Equality_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Equality_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Equality_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterEquality_expr(s)
	}
}

func (s *Equality_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitEquality_expr(s)
	}
}

func (s *Equality_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitEquality_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Equality_expr() (localctx IEquality_exprContext) {
	localctx = NewEquality_exprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 74, NeuroScriptParserRULE_equality_expr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(357)
		p.Relational_expr()
	}
	p.SetState(362)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserEQ || _la == NeuroScriptParserNEQ {
		{
			p.SetState(358)
			_la = p.GetTokenStream().LA(1)

			if !(_la == NeuroScriptParserEQ || _la == NeuroScriptParserNEQ) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(359)
			p.Relational_expr()
		}

		p.SetState(364)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IRelational_exprContext is an interface to support dynamic dispatch.
type IRelational_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllAdditive_expr() []IAdditive_exprContext
	Additive_expr(i int) IAdditive_exprContext
	AllGT() []antlr.TerminalNode
	GT(i int) antlr.TerminalNode
	AllLT() []antlr.TerminalNode
	LT(i int) antlr.TerminalNode
	AllGTE() []antlr.TerminalNode
	GTE(i int) antlr.TerminalNode
	AllLTE() []antlr.TerminalNode
	LTE(i int) antlr.TerminalNode

	// IsRelational_exprContext differentiates from other interfaces.
	IsRelational_exprContext()
}

type Relational_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRelational_exprContext() *Relational_exprContext {
	var p = new(Relational_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_relational_expr
	return p
}

func InitEmptyRelational_exprContext(p *Relational_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_relational_expr
}

func (*Relational_exprContext) IsRelational_exprContext() {}

func NewRelational_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Relational_exprContext {
	var p = new(Relational_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_relational_expr

	return p
}

func (s *Relational_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Relational_exprContext) AllAdditive_expr() []IAdditive_exprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAdditive_exprContext); ok {
			len++
		}
	}

	tst := make([]IAdditive_exprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAdditive_exprContext); ok {
			tst[i] = t.(IAdditive_exprContext)
			i++
		}
	}

	return tst
}

func (s *Relational_exprContext) Additive_expr(i int) IAdditive_exprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAdditive_exprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAdditive_exprContext)
}

func (s *Relational_exprContext) AllGT() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserGT)
}

func (s *Relational_exprContext) GT(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserGT, i)
}

func (s *Relational_exprContext) AllLT() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserLT)
}

func (s *Relational_exprContext) LT(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLT, i)
}

func (s *Relational_exprContext) AllGTE() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserGTE)
}

func (s *Relational_exprContext) GTE(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserGTE, i)
}

func (s *Relational_exprContext) AllLTE() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserLTE)
}

func (s *Relational_exprContext) LTE(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLTE, i)
}

func (s *Relational_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Relational_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Relational_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterRelational_expr(s)
	}
}

func (s *Relational_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitRelational_expr(s)
	}
}

func (s *Relational_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitRelational_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Relational_expr() (localctx IRelational_exprContext) {
	localctx = NewRelational_exprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 76, NeuroScriptParserRULE_relational_expr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(365)
		p.Additive_expr()
	}
	p.SetState(370)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64((_la-77)) & ^0x3f) == 0 && ((int64(1)<<(_la-77))&15) != 0 {
		{
			p.SetState(366)
			_la = p.GetTokenStream().LA(1)

			if !((int64((_la-77)) & ^0x3f) == 0 && ((int64(1)<<(_la-77))&15) != 0) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(367)
			p.Additive_expr()
		}

		p.SetState(372)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAdditive_exprContext is an interface to support dynamic dispatch.
type IAdditive_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllMultiplicative_expr() []IMultiplicative_exprContext
	Multiplicative_expr(i int) IMultiplicative_exprContext
	AllPLUS() []antlr.TerminalNode
	PLUS(i int) antlr.TerminalNode
	AllMINUS() []antlr.TerminalNode
	MINUS(i int) antlr.TerminalNode

	// IsAdditive_exprContext differentiates from other interfaces.
	IsAdditive_exprContext()
}

type Additive_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAdditive_exprContext() *Additive_exprContext {
	var p = new(Additive_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_additive_expr
	return p
}

func InitEmptyAdditive_exprContext(p *Additive_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_additive_expr
}

func (*Additive_exprContext) IsAdditive_exprContext() {}

func NewAdditive_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Additive_exprContext {
	var p = new(Additive_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_additive_expr

	return p
}

func (s *Additive_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Additive_exprContext) AllMultiplicative_expr() []IMultiplicative_exprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IMultiplicative_exprContext); ok {
			len++
		}
	}

	tst := make([]IMultiplicative_exprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IMultiplicative_exprContext); ok {
			tst[i] = t.(IMultiplicative_exprContext)
			i++
		}
	}

	return tst
}

func (s *Additive_exprContext) Multiplicative_expr(i int) IMultiplicative_exprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMultiplicative_exprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMultiplicative_exprContext)
}

func (s *Additive_exprContext) AllPLUS() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserPLUS)
}

func (s *Additive_exprContext) PLUS(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserPLUS, i)
}

func (s *Additive_exprContext) AllMINUS() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserMINUS)
}

func (s *Additive_exprContext) MINUS(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserMINUS, i)
}

func (s *Additive_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Additive_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Additive_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterAdditive_expr(s)
	}
}

func (s *Additive_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitAdditive_expr(s)
	}
}

func (s *Additive_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitAdditive_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Additive_expr() (localctx IAdditive_exprContext) {
	localctx = NewAdditive_exprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 78, NeuroScriptParserRULE_additive_expr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(373)
		p.Multiplicative_expr()
	}
	p.SetState(378)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserPLUS || _la == NeuroScriptParserMINUS {
		{
			p.SetState(374)
			_la = p.GetTokenStream().LA(1)

			if !(_la == NeuroScriptParserPLUS || _la == NeuroScriptParserMINUS) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(375)
			p.Multiplicative_expr()
		}

		p.SetState(380)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMultiplicative_exprContext is an interface to support dynamic dispatch.
type IMultiplicative_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllUnary_expr() []IUnary_exprContext
	Unary_expr(i int) IUnary_exprContext
	AllSTAR() []antlr.TerminalNode
	STAR(i int) antlr.TerminalNode
	AllSLASH() []antlr.TerminalNode
	SLASH(i int) antlr.TerminalNode
	AllPERCENT() []antlr.TerminalNode
	PERCENT(i int) antlr.TerminalNode

	// IsMultiplicative_exprContext differentiates from other interfaces.
	IsMultiplicative_exprContext()
}

type Multiplicative_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMultiplicative_exprContext() *Multiplicative_exprContext {
	var p = new(Multiplicative_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_multiplicative_expr
	return p
}

func InitEmptyMultiplicative_exprContext(p *Multiplicative_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_multiplicative_expr
}

func (*Multiplicative_exprContext) IsMultiplicative_exprContext() {}

func NewMultiplicative_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Multiplicative_exprContext {
	var p = new(Multiplicative_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_multiplicative_expr

	return p
}

func (s *Multiplicative_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Multiplicative_exprContext) AllUnary_expr() []IUnary_exprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IUnary_exprContext); ok {
			len++
		}
	}

	tst := make([]IUnary_exprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IUnary_exprContext); ok {
			tst[i] = t.(IUnary_exprContext)
			i++
		}
	}

	return tst
}

func (s *Multiplicative_exprContext) Unary_expr(i int) IUnary_exprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUnary_exprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUnary_exprContext)
}

func (s *Multiplicative_exprContext) AllSTAR() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserSTAR)
}

func (s *Multiplicative_exprContext) STAR(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserSTAR, i)
}

func (s *Multiplicative_exprContext) AllSLASH() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserSLASH)
}

func (s *Multiplicative_exprContext) SLASH(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserSLASH, i)
}

func (s *Multiplicative_exprContext) AllPERCENT() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserPERCENT)
}

func (s *Multiplicative_exprContext) PERCENT(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserPERCENT, i)
}

func (s *Multiplicative_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Multiplicative_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Multiplicative_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterMultiplicative_expr(s)
	}
}

func (s *Multiplicative_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitMultiplicative_expr(s)
	}
}

func (s *Multiplicative_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitMultiplicative_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Multiplicative_expr() (localctx IMultiplicative_exprContext) {
	localctx = NewMultiplicative_exprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 80, NeuroScriptParserRULE_multiplicative_expr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(381)
		p.Unary_expr()
	}
	p.SetState(386)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&504403158265495552) != 0 {
		{
			p.SetState(382)
			_la = p.GetTokenStream().LA(1)

			if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&504403158265495552) != 0) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(383)
			p.Unary_expr()
		}

		p.SetState(388)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IUnary_exprContext is an interface to support dynamic dispatch.
type IUnary_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Unary_expr() IUnary_exprContext
	MINUS() antlr.TerminalNode
	KW_NOT() antlr.TerminalNode
	KW_NO() antlr.TerminalNode
	KW_SOME() antlr.TerminalNode
	TILDE() antlr.TerminalNode
	KW_TYPEOF() antlr.TerminalNode
	Power_expr() IPower_exprContext

	// IsUnary_exprContext differentiates from other interfaces.
	IsUnary_exprContext()
}

type Unary_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUnary_exprContext() *Unary_exprContext {
	var p = new(Unary_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_unary_expr
	return p
}

func InitEmptyUnary_exprContext(p *Unary_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_unary_expr
}

func (*Unary_exprContext) IsUnary_exprContext() {}

func NewUnary_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Unary_exprContext {
	var p = new(Unary_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_unary_expr

	return p
}

func (s *Unary_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Unary_exprContext) Unary_expr() IUnary_exprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUnary_exprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUnary_exprContext)
}

func (s *Unary_exprContext) MINUS() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserMINUS, 0)
}

func (s *Unary_exprContext) KW_NOT() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_NOT, 0)
}

func (s *Unary_exprContext) KW_NO() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_NO, 0)
}

func (s *Unary_exprContext) KW_SOME() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_SOME, 0)
}

func (s *Unary_exprContext) TILDE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserTILDE, 0)
}

func (s *Unary_exprContext) KW_TYPEOF() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_TYPEOF, 0)
}

func (s *Unary_exprContext) Power_expr() IPower_exprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPower_exprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPower_exprContext)
}

func (s *Unary_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Unary_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Unary_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterUnary_expr(s)
	}
}

func (s *Unary_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitUnary_expr(s)
	}
}

func (s *Unary_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitUnary_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Unary_expr() (localctx IUnary_exprContext) {
	localctx = NewUnary_exprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 82, NeuroScriptParserRULE_unary_expr)
	var _la int

	p.SetState(394)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_NO, NeuroScriptParserKW_SOME, NeuroScriptParserKW_NOT, NeuroScriptParserMINUS, NeuroScriptParserTILDE:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(389)
			_la = p.GetTokenStream().LA(1)

			if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&-9187343205274746880) != 0) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(390)
			p.Unary_expr()
		}

	case NeuroScriptParserKW_TYPEOF:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(391)
			p.Match(NeuroScriptParserKW_TYPEOF)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(392)
			p.Unary_expr()
		}

	case NeuroScriptParserKW_TOOL, NeuroScriptParserKW_LAST, NeuroScriptParserKW_EVAL, NeuroScriptParserKW_TRUE, NeuroScriptParserKW_FALSE, NeuroScriptParserKW_LN, NeuroScriptParserKW_LOG, NeuroScriptParserKW_SIN, NeuroScriptParserKW_COS, NeuroScriptParserKW_TAN, NeuroScriptParserKW_ASIN, NeuroScriptParserKW_ACOS, NeuroScriptParserKW_ATAN, NeuroScriptParserKW_NIL, NeuroScriptParserNUMBER_LIT, NeuroScriptParserSTRING_LIT, NeuroScriptParserTRIPLE_BACKTICK_STRING, NeuroScriptParserLPAREN, NeuroScriptParserLBRACK, NeuroScriptParserLBRACE, NeuroScriptParserPLACEHOLDER_START, NeuroScriptParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(393)
			p.Power_expr()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPower_exprContext is an interface to support dynamic dispatch.
type IPower_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Accessor_expr() IAccessor_exprContext
	STAR_STAR() antlr.TerminalNode
	Power_expr() IPower_exprContext

	// IsPower_exprContext differentiates from other interfaces.
	IsPower_exprContext()
}

type Power_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPower_exprContext() *Power_exprContext {
	var p = new(Power_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_power_expr
	return p
}

func InitEmptyPower_exprContext(p *Power_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_power_expr
}

func (*Power_exprContext) IsPower_exprContext() {}

func NewPower_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Power_exprContext {
	var p = new(Power_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_power_expr

	return p
}

func (s *Power_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Power_exprContext) Accessor_expr() IAccessor_exprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAccessor_exprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAccessor_exprContext)
}

func (s *Power_exprContext) STAR_STAR() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserSTAR_STAR, 0)
}

func (s *Power_exprContext) Power_expr() IPower_exprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPower_exprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPower_exprContext)
}

func (s *Power_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Power_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Power_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterPower_expr(s)
	}
}

func (s *Power_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitPower_expr(s)
	}
}

func (s *Power_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitPower_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Power_expr() (localctx IPower_exprContext) {
	localctx = NewPower_exprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 84, NeuroScriptParserRULE_power_expr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(396)
		p.Accessor_expr()
	}
	p.SetState(399)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NeuroScriptParserSTAR_STAR {
		{
			p.SetState(397)
			p.Match(NeuroScriptParserSTAR_STAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(398)
			p.Power_expr()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAccessor_exprContext is an interface to support dynamic dispatch.
type IAccessor_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Primary() IPrimaryContext
	AllLBRACK() []antlr.TerminalNode
	LBRACK(i int) antlr.TerminalNode
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	AllRBRACK() []antlr.TerminalNode
	RBRACK(i int) antlr.TerminalNode

	// IsAccessor_exprContext differentiates from other interfaces.
	IsAccessor_exprContext()
}

type Accessor_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAccessor_exprContext() *Accessor_exprContext {
	var p = new(Accessor_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_accessor_expr
	return p
}

func InitEmptyAccessor_exprContext(p *Accessor_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_accessor_expr
}

func (*Accessor_exprContext) IsAccessor_exprContext() {}

func NewAccessor_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Accessor_exprContext {
	var p = new(Accessor_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_accessor_expr

	return p
}

func (s *Accessor_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Accessor_exprContext) Primary() IPrimaryContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPrimaryContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPrimaryContext)
}

func (s *Accessor_exprContext) AllLBRACK() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserLBRACK)
}

func (s *Accessor_exprContext) LBRACK(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLBRACK, i)
}

func (s *Accessor_exprContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *Accessor_exprContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *Accessor_exprContext) AllRBRACK() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserRBRACK)
}

func (s *Accessor_exprContext) RBRACK(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserRBRACK, i)
}

func (s *Accessor_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Accessor_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Accessor_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterAccessor_expr(s)
	}
}

func (s *Accessor_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitAccessor_expr(s)
	}
}

func (s *Accessor_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitAccessor_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Accessor_expr() (localctx IAccessor_exprContext) {
	localctx = NewAccessor_exprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 86, NeuroScriptParserRULE_accessor_expr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(401)
		p.Primary()
	}
	p.SetState(408)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserLBRACK {
		{
			p.SetState(402)
			p.Match(NeuroScriptParserLBRACK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(403)
			p.Expression()
		}
		{
			p.SetState(404)
			p.Match(NeuroScriptParserRBRACK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(410)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPrimaryContext is an interface to support dynamic dispatch.
type IPrimaryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Literal() ILiteralContext
	Placeholder() IPlaceholderContext
	IDENTIFIER() antlr.TerminalNode
	KW_LAST() antlr.TerminalNode
	Callable_expr() ICallable_exprContext
	KW_EVAL() antlr.TerminalNode
	LPAREN() antlr.TerminalNode
	Expression() IExpressionContext
	RPAREN() antlr.TerminalNode

	// IsPrimaryContext differentiates from other interfaces.
	IsPrimaryContext()
}

type PrimaryContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPrimaryContext() *PrimaryContext {
	var p = new(PrimaryContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_primary
	return p
}

func InitEmptyPrimaryContext(p *PrimaryContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_primary
}

func (*PrimaryContext) IsPrimaryContext() {}

func NewPrimaryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PrimaryContext {
	var p = new(PrimaryContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_primary

	return p
}

func (s *PrimaryContext) GetParser() antlr.Parser { return s.parser }

func (s *PrimaryContext) Literal() ILiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteralContext)
}

func (s *PrimaryContext) Placeholder() IPlaceholderContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPlaceholderContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPlaceholderContext)
}

func (s *PrimaryContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserIDENTIFIER, 0)
}

func (s *PrimaryContext) KW_LAST() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_LAST, 0)
}

func (s *PrimaryContext) Callable_expr() ICallable_exprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICallable_exprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICallable_exprContext)
}

func (s *PrimaryContext) KW_EVAL() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_EVAL, 0)
}

func (s *PrimaryContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLPAREN, 0)
}

func (s *PrimaryContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *PrimaryContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserRPAREN, 0)
}

func (s *PrimaryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PrimaryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PrimaryContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterPrimary(s)
	}
}

func (s *PrimaryContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitPrimary(s)
	}
}

func (s *PrimaryContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitPrimary(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Primary() (localctx IPrimaryContext) {
	localctx = NewPrimaryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 88, NeuroScriptParserRULE_primary)
	p.SetState(425)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 37, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(411)
			p.Literal()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(412)
			p.Placeholder()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(413)
			p.Match(NeuroScriptParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(414)
			p.Match(NeuroScriptParserKW_LAST)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(415)
			p.Callable_expr()
		}

	case 6:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(416)
			p.Match(NeuroScriptParserKW_EVAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(417)
			p.Match(NeuroScriptParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(418)
			p.Expression()
		}
		{
			p.SetState(419)
			p.Match(NeuroScriptParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 7:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(421)
			p.Match(NeuroScriptParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(422)
			p.Expression()
		}
		{
			p.SetState(423)
			p.Match(NeuroScriptParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICallable_exprContext is an interface to support dynamic dispatch.
type ICallable_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LPAREN() antlr.TerminalNode
	Expression_list_opt() IExpression_list_optContext
	RPAREN() antlr.TerminalNode
	Call_target() ICall_targetContext
	KW_LN() antlr.TerminalNode
	KW_LOG() antlr.TerminalNode
	KW_SIN() antlr.TerminalNode
	KW_COS() antlr.TerminalNode
	KW_TAN() antlr.TerminalNode
	KW_ASIN() antlr.TerminalNode
	KW_ACOS() antlr.TerminalNode
	KW_ATAN() antlr.TerminalNode

	// IsCallable_exprContext differentiates from other interfaces.
	IsCallable_exprContext()
}

type Callable_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCallable_exprContext() *Callable_exprContext {
	var p = new(Callable_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_callable_expr
	return p
}

func InitEmptyCallable_exprContext(p *Callable_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_callable_expr
}

func (*Callable_exprContext) IsCallable_exprContext() {}

func NewCallable_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Callable_exprContext {
	var p = new(Callable_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_callable_expr

	return p
}

func (s *Callable_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Callable_exprContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLPAREN, 0)
}

func (s *Callable_exprContext) Expression_list_opt() IExpression_list_optContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpression_list_optContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpression_list_optContext)
}

func (s *Callable_exprContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserRPAREN, 0)
}

func (s *Callable_exprContext) Call_target() ICall_targetContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICall_targetContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICall_targetContext)
}

func (s *Callable_exprContext) KW_LN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_LN, 0)
}

func (s *Callable_exprContext) KW_LOG() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_LOG, 0)
}

func (s *Callable_exprContext) KW_SIN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_SIN, 0)
}

func (s *Callable_exprContext) KW_COS() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_COS, 0)
}

func (s *Callable_exprContext) KW_TAN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_TAN, 0)
}

func (s *Callable_exprContext) KW_ASIN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ASIN, 0)
}

func (s *Callable_exprContext) KW_ACOS() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ACOS, 0)
}

func (s *Callable_exprContext) KW_ATAN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ATAN, 0)
}

func (s *Callable_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Callable_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Callable_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterCallable_expr(s)
	}
}

func (s *Callable_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitCallable_expr(s)
	}
}

func (s *Callable_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitCallable_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Callable_expr() (localctx ICallable_exprContext) {
	localctx = NewCallable_exprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 90, NeuroScriptParserRULE_callable_expr)
	p.EnterOuterAlt(localctx, 1)
	p.SetState(436)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_TOOL, NeuroScriptParserIDENTIFIER:
		{
			p.SetState(427)
			p.Call_target()
		}

	case NeuroScriptParserKW_LN:
		{
			p.SetState(428)
			p.Match(NeuroScriptParserKW_LN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_LOG:
		{
			p.SetState(429)
			p.Match(NeuroScriptParserKW_LOG)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_SIN:
		{
			p.SetState(430)
			p.Match(NeuroScriptParserKW_SIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_COS:
		{
			p.SetState(431)
			p.Match(NeuroScriptParserKW_COS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_TAN:
		{
			p.SetState(432)
			p.Match(NeuroScriptParserKW_TAN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_ASIN:
		{
			p.SetState(433)
			p.Match(NeuroScriptParserKW_ASIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_ACOS:
		{
			p.SetState(434)
			p.Match(NeuroScriptParserKW_ACOS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_ATAN:
		{
			p.SetState(435)
			p.Match(NeuroScriptParserKW_ATAN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	{
		p.SetState(438)
		p.Match(NeuroScriptParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(439)
		p.Expression_list_opt()
	}
	{
		p.SetState(440)
		p.Match(NeuroScriptParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPlaceholderContext is an interface to support dynamic dispatch.
type IPlaceholderContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	PLACEHOLDER_START() antlr.TerminalNode
	PLACEHOLDER_END() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode
	KW_LAST() antlr.TerminalNode

	// IsPlaceholderContext differentiates from other interfaces.
	IsPlaceholderContext()
}

type PlaceholderContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPlaceholderContext() *PlaceholderContext {
	var p = new(PlaceholderContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_placeholder
	return p
}

func InitEmptyPlaceholderContext(p *PlaceholderContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_placeholder
}

func (*PlaceholderContext) IsPlaceholderContext() {}

func NewPlaceholderContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PlaceholderContext {
	var p = new(PlaceholderContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_placeholder

	return p
}

func (s *PlaceholderContext) GetParser() antlr.Parser { return s.parser }

func (s *PlaceholderContext) PLACEHOLDER_START() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserPLACEHOLDER_START, 0)
}

func (s *PlaceholderContext) PLACEHOLDER_END() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserPLACEHOLDER_END, 0)
}

func (s *PlaceholderContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserIDENTIFIER, 0)
}

func (s *PlaceholderContext) KW_LAST() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_LAST, 0)
}

func (s *PlaceholderContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PlaceholderContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PlaceholderContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterPlaceholder(s)
	}
}

func (s *PlaceholderContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitPlaceholder(s)
	}
}

func (s *PlaceholderContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitPlaceholder(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Placeholder() (localctx IPlaceholderContext) {
	localctx = NewPlaceholderContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 92, NeuroScriptParserRULE_placeholder)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(442)
		p.Match(NeuroScriptParserPLACEHOLDER_START)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(443)
		_la = p.GetTokenStream().LA(1)

		if !(_la == NeuroScriptParserKW_LAST || _la == NeuroScriptParserIDENTIFIER) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	{
		p.SetState(444)
		p.Match(NeuroScriptParserPLACEHOLDER_END)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILiteralContext is an interface to support dynamic dispatch.
type ILiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING_LIT() antlr.TerminalNode
	TRIPLE_BACKTICK_STRING() antlr.TerminalNode
	NUMBER_LIT() antlr.TerminalNode
	List_literal() IList_literalContext
	Map_literal() IMap_literalContext
	Boolean_literal() IBoolean_literalContext
	Nil_literal() INil_literalContext

	// IsLiteralContext differentiates from other interfaces.
	IsLiteralContext()
}

type LiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLiteralContext() *LiteralContext {
	var p = new(LiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_literal
	return p
}

func InitEmptyLiteralContext(p *LiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_literal
}

func (*LiteralContext) IsLiteralContext() {}

func NewLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LiteralContext {
	var p = new(LiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_literal

	return p
}

func (s *LiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *LiteralContext) STRING_LIT() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserSTRING_LIT, 0)
}

func (s *LiteralContext) TRIPLE_BACKTICK_STRING() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserTRIPLE_BACKTICK_STRING, 0)
}

func (s *LiteralContext) NUMBER_LIT() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNUMBER_LIT, 0)
}

func (s *LiteralContext) List_literal() IList_literalContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IList_literalContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IList_literalContext)
}

func (s *LiteralContext) Map_literal() IMap_literalContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMap_literalContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMap_literalContext)
}

func (s *LiteralContext) Boolean_literal() IBoolean_literalContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBoolean_literalContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBoolean_literalContext)
}

func (s *LiteralContext) Nil_literal() INil_literalContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INil_literalContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INil_literalContext)
}

func (s *LiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterLiteral(s)
	}
}

func (s *LiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitLiteral(s)
	}
}

func (s *LiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Literal() (localctx ILiteralContext) {
	localctx = NewLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 94, NeuroScriptParserRULE_literal)
	p.SetState(453)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserSTRING_LIT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(446)
			p.Match(NeuroScriptParserSTRING_LIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserTRIPLE_BACKTICK_STRING:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(447)
			p.Match(NeuroScriptParserTRIPLE_BACKTICK_STRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserNUMBER_LIT:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(448)
			p.Match(NeuroScriptParserNUMBER_LIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserLBRACK:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(449)
			p.List_literal()
		}

	case NeuroScriptParserLBRACE:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(450)
			p.Map_literal()
		}

	case NeuroScriptParserKW_TRUE, NeuroScriptParserKW_FALSE:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(451)
			p.Boolean_literal()
		}

	case NeuroScriptParserKW_NIL:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(452)
			p.Nil_literal()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// INil_literalContext is an interface to support dynamic dispatch.
type INil_literalContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_NIL() antlr.TerminalNode

	// IsNil_literalContext differentiates from other interfaces.
	IsNil_literalContext()
}

type Nil_literalContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNil_literalContext() *Nil_literalContext {
	var p = new(Nil_literalContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_nil_literal
	return p
}

func InitEmptyNil_literalContext(p *Nil_literalContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_nil_literal
}

func (*Nil_literalContext) IsNil_literalContext() {}

func NewNil_literalContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Nil_literalContext {
	var p = new(Nil_literalContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_nil_literal

	return p
}

func (s *Nil_literalContext) GetParser() antlr.Parser { return s.parser }

func (s *Nil_literalContext) KW_NIL() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_NIL, 0)
}

func (s *Nil_literalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Nil_literalContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Nil_literalContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterNil_literal(s)
	}
}

func (s *Nil_literalContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitNil_literal(s)
	}
}

func (s *Nil_literalContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitNil_literal(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Nil_literal() (localctx INil_literalContext) {
	localctx = NewNil_literalContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 96, NeuroScriptParserRULE_nil_literal)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(455)
		p.Match(NeuroScriptParserKW_NIL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBoolean_literalContext is an interface to support dynamic dispatch.
type IBoolean_literalContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_TRUE() antlr.TerminalNode
	KW_FALSE() antlr.TerminalNode

	// IsBoolean_literalContext differentiates from other interfaces.
	IsBoolean_literalContext()
}

type Boolean_literalContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBoolean_literalContext() *Boolean_literalContext {
	var p = new(Boolean_literalContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_boolean_literal
	return p
}

func InitEmptyBoolean_literalContext(p *Boolean_literalContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_boolean_literal
}

func (*Boolean_literalContext) IsBoolean_literalContext() {}

func NewBoolean_literalContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Boolean_literalContext {
	var p = new(Boolean_literalContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_boolean_literal

	return p
}

func (s *Boolean_literalContext) GetParser() antlr.Parser { return s.parser }

func (s *Boolean_literalContext) KW_TRUE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_TRUE, 0)
}

func (s *Boolean_literalContext) KW_FALSE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_FALSE, 0)
}

func (s *Boolean_literalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Boolean_literalContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Boolean_literalContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterBoolean_literal(s)
	}
}

func (s *Boolean_literalContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitBoolean_literal(s)
	}
}

func (s *Boolean_literalContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitBoolean_literal(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Boolean_literal() (localctx IBoolean_literalContext) {
	localctx = NewBoolean_literalContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 98, NeuroScriptParserRULE_boolean_literal)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(457)
		_la = p.GetTokenStream().LA(1)

		if !(_la == NeuroScriptParserKW_TRUE || _la == NeuroScriptParserKW_FALSE) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IList_literalContext is an interface to support dynamic dispatch.
type IList_literalContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LBRACK() antlr.TerminalNode
	Expression_list_opt() IExpression_list_optContext
	RBRACK() antlr.TerminalNode

	// IsList_literalContext differentiates from other interfaces.
	IsList_literalContext()
}

type List_literalContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyList_literalContext() *List_literalContext {
	var p = new(List_literalContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_list_literal
	return p
}

func InitEmptyList_literalContext(p *List_literalContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_list_literal
}

func (*List_literalContext) IsList_literalContext() {}

func NewList_literalContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *List_literalContext {
	var p = new(List_literalContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_list_literal

	return p
}

func (s *List_literalContext) GetParser() antlr.Parser { return s.parser }

func (s *List_literalContext) LBRACK() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLBRACK, 0)
}

func (s *List_literalContext) Expression_list_opt() IExpression_list_optContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpression_list_optContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpression_list_optContext)
}

func (s *List_literalContext) RBRACK() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserRBRACK, 0)
}

func (s *List_literalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *List_literalContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *List_literalContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterList_literal(s)
	}
}

func (s *List_literalContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitList_literal(s)
	}
}

func (s *List_literalContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitList_literal(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) List_literal() (localctx IList_literalContext) {
	localctx = NewList_literalContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 100, NeuroScriptParserRULE_list_literal)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(459)
		p.Match(NeuroScriptParserLBRACK)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(460)
		p.Expression_list_opt()
	}
	{
		p.SetState(461)
		p.Match(NeuroScriptParserRBRACK)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMap_literalContext is an interface to support dynamic dispatch.
type IMap_literalContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LBRACE() antlr.TerminalNode
	Map_entry_list_opt() IMap_entry_list_optContext
	RBRACE() antlr.TerminalNode

	// IsMap_literalContext differentiates from other interfaces.
	IsMap_literalContext()
}

type Map_literalContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMap_literalContext() *Map_literalContext {
	var p = new(Map_literalContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_map_literal
	return p
}

func InitEmptyMap_literalContext(p *Map_literalContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_map_literal
}

func (*Map_literalContext) IsMap_literalContext() {}

func NewMap_literalContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Map_literalContext {
	var p = new(Map_literalContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_map_literal

	return p
}

func (s *Map_literalContext) GetParser() antlr.Parser { return s.parser }

func (s *Map_literalContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLBRACE, 0)
}

func (s *Map_literalContext) Map_entry_list_opt() IMap_entry_list_optContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMap_entry_list_optContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMap_entry_list_optContext)
}

func (s *Map_literalContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserRBRACE, 0)
}

func (s *Map_literalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Map_literalContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Map_literalContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterMap_literal(s)
	}
}

func (s *Map_literalContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitMap_literal(s)
	}
}

func (s *Map_literalContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitMap_literal(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Map_literal() (localctx IMap_literalContext) {
	localctx = NewMap_literalContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 102, NeuroScriptParserRULE_map_literal)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(463)
		p.Match(NeuroScriptParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(464)
		p.Map_entry_list_opt()
	}
	{
		p.SetState(465)
		p.Match(NeuroScriptParserRBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExpression_list_optContext is an interface to support dynamic dispatch.
type IExpression_list_optContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Expression_list() IExpression_listContext

	// IsExpression_list_optContext differentiates from other interfaces.
	IsExpression_list_optContext()
}

type Expression_list_optContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpression_list_optContext() *Expression_list_optContext {
	var p = new(Expression_list_optContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_expression_list_opt
	return p
}

func InitEmptyExpression_list_optContext(p *Expression_list_optContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_expression_list_opt
}

func (*Expression_list_optContext) IsExpression_list_optContext() {}

func NewExpression_list_optContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Expression_list_optContext {
	var p = new(Expression_list_optContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_expression_list_opt

	return p
}

func (s *Expression_list_optContext) GetParser() antlr.Parser { return s.parser }

func (s *Expression_list_optContext) Expression_list() IExpression_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpression_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpression_listContext)
}

func (s *Expression_list_optContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expression_list_optContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Expression_list_optContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterExpression_list_opt(s)
	}
}

func (s *Expression_list_optContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitExpression_list_opt(s)
	}
}

func (s *Expression_list_optContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitExpression_list_opt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Expression_list_opt() (localctx IExpression_list_optContext) {
	localctx = NewExpression_list_optContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 104, NeuroScriptParserRULE_expression_list_opt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(468)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64((_la-26)) & ^0x3f) == 0 && ((int64(1)<<(_la-26))&36180942607613567) != 0 {
		{
			p.SetState(467)
			p.Expression_list()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExpression_listContext is an interface to support dynamic dispatch.
type IExpression_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsExpression_listContext differentiates from other interfaces.
	IsExpression_listContext()
}

type Expression_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpression_listContext() *Expression_listContext {
	var p = new(Expression_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_expression_list
	return p
}

func InitEmptyExpression_listContext(p *Expression_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_expression_list
}

func (*Expression_listContext) IsExpression_listContext() {}

func NewExpression_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Expression_listContext {
	var p = new(Expression_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_expression_list

	return p
}

func (s *Expression_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Expression_listContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *Expression_listContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *Expression_listContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserCOMMA)
}

func (s *Expression_listContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserCOMMA, i)
}

func (s *Expression_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expression_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Expression_listContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterExpression_list(s)
	}
}

func (s *Expression_listContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitExpression_list(s)
	}
}

func (s *Expression_listContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitExpression_list(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Expression_list() (localctx IExpression_listContext) {
	localctx = NewExpression_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 106, NeuroScriptParserRULE_expression_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(470)
		p.Expression()
	}
	p.SetState(475)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserCOMMA {
		{
			p.SetState(471)
			p.Match(NeuroScriptParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(472)
			p.Expression()
		}

		p.SetState(477)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMap_entry_list_optContext is an interface to support dynamic dispatch.
type IMap_entry_list_optContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Map_entry_list() IMap_entry_listContext

	// IsMap_entry_list_optContext differentiates from other interfaces.
	IsMap_entry_list_optContext()
}

type Map_entry_list_optContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMap_entry_list_optContext() *Map_entry_list_optContext {
	var p = new(Map_entry_list_optContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_map_entry_list_opt
	return p
}

func InitEmptyMap_entry_list_optContext(p *Map_entry_list_optContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_map_entry_list_opt
}

func (*Map_entry_list_optContext) IsMap_entry_list_optContext() {}

func NewMap_entry_list_optContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Map_entry_list_optContext {
	var p = new(Map_entry_list_optContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_map_entry_list_opt

	return p
}

func (s *Map_entry_list_optContext) GetParser() antlr.Parser { return s.parser }

func (s *Map_entry_list_optContext) Map_entry_list() IMap_entry_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMap_entry_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMap_entry_listContext)
}

func (s *Map_entry_list_optContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Map_entry_list_optContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Map_entry_list_optContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterMap_entry_list_opt(s)
	}
}

func (s *Map_entry_list_optContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitMap_entry_list_opt(s)
	}
}

func (s *Map_entry_list_optContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitMap_entry_list_opt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Map_entry_list_opt() (localctx IMap_entry_list_optContext) {
	localctx = NewMap_entry_list_optContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 108, NeuroScriptParserRULE_map_entry_list_opt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(479)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NeuroScriptParserSTRING_LIT {
		{
			p.SetState(478)
			p.Map_entry_list()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMap_entry_listContext is an interface to support dynamic dispatch.
type IMap_entry_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllMap_entry() []IMap_entryContext
	Map_entry(i int) IMap_entryContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsMap_entry_listContext differentiates from other interfaces.
	IsMap_entry_listContext()
}

type Map_entry_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMap_entry_listContext() *Map_entry_listContext {
	var p = new(Map_entry_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_map_entry_list
	return p
}

func InitEmptyMap_entry_listContext(p *Map_entry_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_map_entry_list
}

func (*Map_entry_listContext) IsMap_entry_listContext() {}

func NewMap_entry_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Map_entry_listContext {
	var p = new(Map_entry_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_map_entry_list

	return p
}

func (s *Map_entry_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Map_entry_listContext) AllMap_entry() []IMap_entryContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IMap_entryContext); ok {
			len++
		}
	}

	tst := make([]IMap_entryContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IMap_entryContext); ok {
			tst[i] = t.(IMap_entryContext)
			i++
		}
	}

	return tst
}

func (s *Map_entry_listContext) Map_entry(i int) IMap_entryContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMap_entryContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMap_entryContext)
}

func (s *Map_entry_listContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserCOMMA)
}

func (s *Map_entry_listContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserCOMMA, i)
}

func (s *Map_entry_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Map_entry_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Map_entry_listContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterMap_entry_list(s)
	}
}

func (s *Map_entry_listContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitMap_entry_list(s)
	}
}

func (s *Map_entry_listContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitMap_entry_list(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Map_entry_list() (localctx IMap_entry_listContext) {
	localctx = NewMap_entry_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 110, NeuroScriptParserRULE_map_entry_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(481)
		p.Map_entry()
	}
	p.SetState(486)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserCOMMA {
		{
			p.SetState(482)
			p.Match(NeuroScriptParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(483)
			p.Map_entry()
		}

		p.SetState(488)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMap_entryContext is an interface to support dynamic dispatch.
type IMap_entryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING_LIT() antlr.TerminalNode
	COLON() antlr.TerminalNode
	Expression() IExpressionContext

	// IsMap_entryContext differentiates from other interfaces.
	IsMap_entryContext()
}

type Map_entryContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMap_entryContext() *Map_entryContext {
	var p = new(Map_entryContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_map_entry
	return p
}

func InitEmptyMap_entryContext(p *Map_entryContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_map_entry
}

func (*Map_entryContext) IsMap_entryContext() {}

func NewMap_entryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Map_entryContext {
	var p = new(Map_entryContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_map_entry

	return p
}

func (s *Map_entryContext) GetParser() antlr.Parser { return s.parser }

func (s *Map_entryContext) STRING_LIT() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserSTRING_LIT, 0)
}

func (s *Map_entryContext) COLON() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserCOLON, 0)
}

func (s *Map_entryContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *Map_entryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Map_entryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Map_entryContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterMap_entry(s)
	}
}

func (s *Map_entryContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitMap_entry(s)
	}
}

func (s *Map_entryContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitMap_entry(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Map_entry() (localctx IMap_entryContext) {
	localctx = NewMap_entryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 112, NeuroScriptParserRULE_map_entry)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(489)
		p.Match(NeuroScriptParserSTRING_LIT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(490)
		p.Match(NeuroScriptParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(491)
		p.Expression()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}
