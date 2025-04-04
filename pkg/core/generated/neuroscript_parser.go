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
		"", "'FILE_VERSION'", "'DEFINE'", "'PROCEDURE'", "'END'", "'ENDBLOCK'",
		"'COMMENT:'", "'ENDCOMMENT'", "'SET'", "'CALL'", "'RETURN'", "'IF'",
		"'THEN'", "'ELSE'", "'WHILE'", "'DO'", "'FOR'", "'EACH'", "'IN'", "'TOOL'",
		"'LLM'", "'LAST'", "'EVAL'", "'EMIT'", "'true'", "'false'", "", "",
		"", "'='", "'+'", "'('", "')'", "','", "'['", "']'", "'{'", "'}'", "':'",
		"'.'", "'/'", "'{{'", "'}}'", "'=='", "'!='", "'>'", "'<'", "'>='",
		"'<='",
	}
	staticData.SymbolicNames = []string{
		"", "KW_FILE_VERSION", "KW_DEFINE", "KW_PROCEDURE", "KW_END", "KW_ENDBLOCK",
		"KW_COMMENT_START", "KW_ENDCOMMENT", "KW_SET", "KW_CALL", "KW_RETURN",
		"KW_IF", "KW_THEN", "KW_ELSE", "KW_WHILE", "KW_DO", "KW_FOR", "KW_EACH",
		"KW_IN", "KW_TOOL", "KW_LLM", "KW_LAST", "KW_EVAL", "KW_EMIT", "KW_TRUE",
		"KW_FALSE", "COMMENT_BLOCK", "NUMBER_LIT", "STRING_LIT", "ASSIGN", "PLUS",
		"LPAREN", "RPAREN", "COMMA", "LBRACK", "RBRACK", "LBRACE", "RBRACE",
		"COLON", "DOT", "SLASH", "PLACEHOLDER_START", "PLACEHOLDER_END", "EQ",
		"NEQ", "GT", "LT", "GTE", "LTE", "IDENTIFIER", "LINE_COMMENT", "HASH_BANG",
		"NEWLINE", "WS",
	}
	staticData.RuleNames = []string{
		"program", "optional_newlines", "file_version_decl", "procedure_definition",
		"param_list_opt", "param_list", "statement_list", "body_line", "statement",
		"simple_statement", "block_statement", "set_statement", "call_statement",
		"return_statement", "emit_statement", "if_statement", "while_statement",
		"for_each_statement", "call_target", "condition", "expression", "term",
		"primary", "placeholder", "literal", "boolean_literal", "list_literal",
		"map_literal", "expression_list_opt", "expression_list", "map_entry_list_opt",
		"map_entry_list", "map_entry",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 53, 286, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7, 20, 2,
		21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25, 2, 26,
		7, 26, 2, 27, 7, 27, 2, 28, 7, 28, 2, 29, 7, 29, 2, 30, 7, 30, 2, 31, 7,
		31, 2, 32, 7, 32, 1, 0, 1, 0, 3, 0, 69, 8, 0, 1, 0, 1, 0, 5, 0, 73, 8,
		0, 10, 0, 12, 0, 76, 9, 0, 1, 0, 1, 0, 1, 0, 1, 1, 5, 1, 82, 8, 1, 10,
		1, 12, 1, 85, 9, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3, 1, 3, 1,
		3, 1, 3, 1, 3, 1, 3, 3, 3, 99, 8, 3, 1, 3, 1, 3, 1, 3, 3, 3, 104, 8, 3,
		1, 4, 3, 4, 107, 8, 4, 1, 5, 1, 5, 1, 5, 5, 5, 112, 8, 5, 10, 5, 12, 5,
		115, 9, 5, 1, 6, 5, 6, 118, 8, 6, 10, 6, 12, 6, 121, 9, 6, 1, 7, 1, 7,
		1, 7, 1, 7, 3, 7, 127, 8, 7, 1, 8, 1, 8, 3, 8, 131, 8, 8, 1, 9, 1, 9, 1,
		9, 1, 9, 3, 9, 137, 8, 9, 1, 10, 1, 10, 1, 10, 3, 10, 142, 8, 10, 1, 11,
		1, 11, 1, 11, 1, 11, 1, 11, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1,
		13, 1, 13, 3, 13, 157, 8, 13, 1, 14, 1, 14, 1, 14, 1, 15, 1, 15, 1, 15,
		1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 3, 15, 170, 8, 15, 1, 15, 1, 15, 1,
		16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 17, 1, 17, 1, 17, 1, 17,
		1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 1, 18, 1, 18, 1, 18, 1, 18, 1,
		18, 3, 18, 196, 8, 18, 1, 19, 1, 19, 1, 19, 3, 19, 201, 8, 19, 1, 20, 1,
		20, 1, 20, 1, 20, 1, 20, 1, 20, 5, 20, 209, 8, 20, 10, 20, 12, 20, 212,
		9, 20, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 5, 21, 219, 8, 21, 10, 21, 12,
		21, 222, 9, 21, 1, 22, 1, 22, 1, 22, 1, 22, 1, 22, 1, 22, 1, 22, 1, 22,
		1, 22, 1, 22, 1, 22, 1, 22, 1, 22, 3, 22, 237, 8, 22, 1, 23, 1, 23, 1,
		23, 1, 23, 1, 24, 1, 24, 1, 24, 1, 24, 1, 24, 3, 24, 248, 8, 24, 1, 25,
		1, 25, 1, 26, 1, 26, 1, 26, 1, 26, 1, 27, 1, 27, 1, 27, 1, 27, 1, 28, 3,
		28, 261, 8, 28, 1, 29, 1, 29, 1, 29, 5, 29, 266, 8, 29, 10, 29, 12, 29,
		269, 9, 29, 1, 30, 3, 30, 272, 8, 30, 1, 31, 1, 31, 1, 31, 5, 31, 277,
		8, 31, 10, 31, 12, 31, 280, 9, 31, 1, 32, 1, 32, 1, 32, 1, 32, 1, 32, 0,
		0, 33, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34,
		36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60, 62, 64, 0, 3, 1, 0,
		43, 48, 2, 0, 21, 21, 49, 49, 1, 0, 24, 25, 287, 0, 66, 1, 0, 0, 0, 2,
		83, 1, 0, 0, 0, 4, 86, 1, 0, 0, 0, 6, 90, 1, 0, 0, 0, 8, 106, 1, 0, 0,
		0, 10, 108, 1, 0, 0, 0, 12, 119, 1, 0, 0, 0, 14, 126, 1, 0, 0, 0, 16, 130,
		1, 0, 0, 0, 18, 136, 1, 0, 0, 0, 20, 141, 1, 0, 0, 0, 22, 143, 1, 0, 0,
		0, 24, 148, 1, 0, 0, 0, 26, 154, 1, 0, 0, 0, 28, 158, 1, 0, 0, 0, 30, 161,
		1, 0, 0, 0, 32, 173, 1, 0, 0, 0, 34, 180, 1, 0, 0, 0, 36, 195, 1, 0, 0,
		0, 38, 197, 1, 0, 0, 0, 40, 202, 1, 0, 0, 0, 42, 213, 1, 0, 0, 0, 44, 236,
		1, 0, 0, 0, 46, 238, 1, 0, 0, 0, 48, 247, 1, 0, 0, 0, 50, 249, 1, 0, 0,
		0, 52, 251, 1, 0, 0, 0, 54, 255, 1, 0, 0, 0, 56, 260, 1, 0, 0, 0, 58, 262,
		1, 0, 0, 0, 60, 271, 1, 0, 0, 0, 62, 273, 1, 0, 0, 0, 64, 281, 1, 0, 0,
		0, 66, 68, 3, 2, 1, 0, 67, 69, 3, 4, 2, 0, 68, 67, 1, 0, 0, 0, 68, 69,
		1, 0, 0, 0, 69, 70, 1, 0, 0, 0, 70, 74, 3, 2, 1, 0, 71, 73, 3, 6, 3, 0,
		72, 71, 1, 0, 0, 0, 73, 76, 1, 0, 0, 0, 74, 72, 1, 0, 0, 0, 74, 75, 1,
		0, 0, 0, 75, 77, 1, 0, 0, 0, 76, 74, 1, 0, 0, 0, 77, 78, 3, 2, 1, 0, 78,
		79, 5, 0, 0, 1, 79, 1, 1, 0, 0, 0, 80, 82, 5, 52, 0, 0, 81, 80, 1, 0, 0,
		0, 82, 85, 1, 0, 0, 0, 83, 81, 1, 0, 0, 0, 83, 84, 1, 0, 0, 0, 84, 3, 1,
		0, 0, 0, 85, 83, 1, 0, 0, 0, 86, 87, 5, 1, 0, 0, 87, 88, 5, 28, 0, 0, 88,
		89, 5, 52, 0, 0, 89, 5, 1, 0, 0, 0, 90, 91, 5, 2, 0, 0, 91, 92, 5, 3, 0,
		0, 92, 93, 5, 49, 0, 0, 93, 94, 5, 31, 0, 0, 94, 95, 3, 8, 4, 0, 95, 96,
		5, 32, 0, 0, 96, 98, 5, 52, 0, 0, 97, 99, 5, 26, 0, 0, 98, 97, 1, 0, 0,
		0, 98, 99, 1, 0, 0, 0, 99, 100, 1, 0, 0, 0, 100, 101, 3, 12, 6, 0, 101,
		103, 5, 4, 0, 0, 102, 104, 5, 52, 0, 0, 103, 102, 1, 0, 0, 0, 103, 104,
		1, 0, 0, 0, 104, 7, 1, 0, 0, 0, 105, 107, 3, 10, 5, 0, 106, 105, 1, 0,
		0, 0, 106, 107, 1, 0, 0, 0, 107, 9, 1, 0, 0, 0, 108, 113, 5, 49, 0, 0,
		109, 110, 5, 33, 0, 0, 110, 112, 5, 49, 0, 0, 111, 109, 1, 0, 0, 0, 112,
		115, 1, 0, 0, 0, 113, 111, 1, 0, 0, 0, 113, 114, 1, 0, 0, 0, 114, 11, 1,
		0, 0, 0, 115, 113, 1, 0, 0, 0, 116, 118, 3, 14, 7, 0, 117, 116, 1, 0, 0,
		0, 118, 121, 1, 0, 0, 0, 119, 117, 1, 0, 0, 0, 119, 120, 1, 0, 0, 0, 120,
		13, 1, 0, 0, 0, 121, 119, 1, 0, 0, 0, 122, 123, 3, 16, 8, 0, 123, 124,
		5, 52, 0, 0, 124, 127, 1, 0, 0, 0, 125, 127, 5, 52, 0, 0, 126, 122, 1,
		0, 0, 0, 126, 125, 1, 0, 0, 0, 127, 15, 1, 0, 0, 0, 128, 131, 3, 18, 9,
		0, 129, 131, 3, 20, 10, 0, 130, 128, 1, 0, 0, 0, 130, 129, 1, 0, 0, 0,
		131, 17, 1, 0, 0, 0, 132, 137, 3, 22, 11, 0, 133, 137, 3, 24, 12, 0, 134,
		137, 3, 26, 13, 0, 135, 137, 3, 28, 14, 0, 136, 132, 1, 0, 0, 0, 136, 133,
		1, 0, 0, 0, 136, 134, 1, 0, 0, 0, 136, 135, 1, 0, 0, 0, 137, 19, 1, 0,
		0, 0, 138, 142, 3, 30, 15, 0, 139, 142, 3, 32, 16, 0, 140, 142, 3, 34,
		17, 0, 141, 138, 1, 0, 0, 0, 141, 139, 1, 0, 0, 0, 141, 140, 1, 0, 0, 0,
		142, 21, 1, 0, 0, 0, 143, 144, 5, 8, 0, 0, 144, 145, 5, 49, 0, 0, 145,
		146, 5, 29, 0, 0, 146, 147, 3, 40, 20, 0, 147, 23, 1, 0, 0, 0, 148, 149,
		5, 9, 0, 0, 149, 150, 3, 36, 18, 0, 150, 151, 5, 31, 0, 0, 151, 152, 3,
		56, 28, 0, 152, 153, 5, 32, 0, 0, 153, 25, 1, 0, 0, 0, 154, 156, 5, 10,
		0, 0, 155, 157, 3, 40, 20, 0, 156, 155, 1, 0, 0, 0, 156, 157, 1, 0, 0,
		0, 157, 27, 1, 0, 0, 0, 158, 159, 5, 23, 0, 0, 159, 160, 3, 40, 20, 0,
		160, 29, 1, 0, 0, 0, 161, 162, 5, 11, 0, 0, 162, 163, 3, 38, 19, 0, 163,
		164, 5, 12, 0, 0, 164, 165, 5, 52, 0, 0, 165, 169, 3, 12, 6, 0, 166, 167,
		5, 13, 0, 0, 167, 168, 5, 52, 0, 0, 168, 170, 3, 12, 6, 0, 169, 166, 1,
		0, 0, 0, 169, 170, 1, 0, 0, 0, 170, 171, 1, 0, 0, 0, 171, 172, 5, 5, 0,
		0, 172, 31, 1, 0, 0, 0, 173, 174, 5, 14, 0, 0, 174, 175, 3, 38, 19, 0,
		175, 176, 5, 15, 0, 0, 176, 177, 5, 52, 0, 0, 177, 178, 3, 12, 6, 0, 178,
		179, 5, 5, 0, 0, 179, 33, 1, 0, 0, 0, 180, 181, 5, 16, 0, 0, 181, 182,
		5, 17, 0, 0, 182, 183, 5, 49, 0, 0, 183, 184, 5, 18, 0, 0, 184, 185, 3,
		40, 20, 0, 185, 186, 5, 15, 0, 0, 186, 187, 5, 52, 0, 0, 187, 188, 3, 12,
		6, 0, 188, 189, 5, 5, 0, 0, 189, 35, 1, 0, 0, 0, 190, 196, 5, 49, 0, 0,
		191, 192, 5, 19, 0, 0, 192, 193, 5, 39, 0, 0, 193, 196, 5, 49, 0, 0, 194,
		196, 5, 20, 0, 0, 195, 190, 1, 0, 0, 0, 195, 191, 1, 0, 0, 0, 195, 194,
		1, 0, 0, 0, 196, 37, 1, 0, 0, 0, 197, 200, 3, 40, 20, 0, 198, 199, 7, 0,
		0, 0, 199, 201, 3, 40, 20, 0, 200, 198, 1, 0, 0, 0, 200, 201, 1, 0, 0,
		0, 201, 39, 1, 0, 0, 0, 202, 210, 3, 42, 21, 0, 203, 204, 3, 2, 1, 0, 204,
		205, 5, 30, 0, 0, 205, 206, 3, 2, 1, 0, 206, 207, 3, 42, 21, 0, 207, 209,
		1, 0, 0, 0, 208, 203, 1, 0, 0, 0, 209, 212, 1, 0, 0, 0, 210, 208, 1, 0,
		0, 0, 210, 211, 1, 0, 0, 0, 211, 41, 1, 0, 0, 0, 212, 210, 1, 0, 0, 0,
		213, 220, 3, 44, 22, 0, 214, 215, 5, 34, 0, 0, 215, 216, 3, 40, 20, 0,
		216, 217, 5, 35, 0, 0, 217, 219, 1, 0, 0, 0, 218, 214, 1, 0, 0, 0, 219,
		222, 1, 0, 0, 0, 220, 218, 1, 0, 0, 0, 220, 221, 1, 0, 0, 0, 221, 43, 1,
		0, 0, 0, 222, 220, 1, 0, 0, 0, 223, 237, 3, 48, 24, 0, 224, 237, 3, 46,
		23, 0, 225, 237, 5, 49, 0, 0, 226, 237, 5, 21, 0, 0, 227, 228, 5, 22, 0,
		0, 228, 229, 5, 31, 0, 0, 229, 230, 3, 40, 20, 0, 230, 231, 5, 32, 0, 0,
		231, 237, 1, 0, 0, 0, 232, 233, 5, 31, 0, 0, 233, 234, 3, 40, 20, 0, 234,
		235, 5, 32, 0, 0, 235, 237, 1, 0, 0, 0, 236, 223, 1, 0, 0, 0, 236, 224,
		1, 0, 0, 0, 236, 225, 1, 0, 0, 0, 236, 226, 1, 0, 0, 0, 236, 227, 1, 0,
		0, 0, 236, 232, 1, 0, 0, 0, 237, 45, 1, 0, 0, 0, 238, 239, 5, 41, 0, 0,
		239, 240, 7, 1, 0, 0, 240, 241, 5, 42, 0, 0, 241, 47, 1, 0, 0, 0, 242,
		248, 5, 28, 0, 0, 243, 248, 5, 27, 0, 0, 244, 248, 3, 52, 26, 0, 245, 248,
		3, 54, 27, 0, 246, 248, 3, 50, 25, 0, 247, 242, 1, 0, 0, 0, 247, 243, 1,
		0, 0, 0, 247, 244, 1, 0, 0, 0, 247, 245, 1, 0, 0, 0, 247, 246, 1, 0, 0,
		0, 248, 49, 1, 0, 0, 0, 249, 250, 7, 2, 0, 0, 250, 51, 1, 0, 0, 0, 251,
		252, 5, 34, 0, 0, 252, 253, 3, 56, 28, 0, 253, 254, 5, 35, 0, 0, 254, 53,
		1, 0, 0, 0, 255, 256, 5, 36, 0, 0, 256, 257, 3, 60, 30, 0, 257, 258, 5,
		37, 0, 0, 258, 55, 1, 0, 0, 0, 259, 261, 3, 58, 29, 0, 260, 259, 1, 0,
		0, 0, 260, 261, 1, 0, 0, 0, 261, 57, 1, 0, 0, 0, 262, 267, 3, 40, 20, 0,
		263, 264, 5, 33, 0, 0, 264, 266, 3, 40, 20, 0, 265, 263, 1, 0, 0, 0, 266,
		269, 1, 0, 0, 0, 267, 265, 1, 0, 0, 0, 267, 268, 1, 0, 0, 0, 268, 59, 1,
		0, 0, 0, 269, 267, 1, 0, 0, 0, 270, 272, 3, 62, 31, 0, 271, 270, 1, 0,
		0, 0, 271, 272, 1, 0, 0, 0, 272, 61, 1, 0, 0, 0, 273, 278, 3, 64, 32, 0,
		274, 275, 5, 33, 0, 0, 275, 277, 3, 64, 32, 0, 276, 274, 1, 0, 0, 0, 277,
		280, 1, 0, 0, 0, 278, 276, 1, 0, 0, 0, 278, 279, 1, 0, 0, 0, 279, 63, 1,
		0, 0, 0, 280, 278, 1, 0, 0, 0, 281, 282, 5, 28, 0, 0, 282, 283, 5, 38,
		0, 0, 283, 284, 3, 40, 20, 0, 284, 65, 1, 0, 0, 0, 24, 68, 74, 83, 98,
		103, 106, 113, 119, 126, 130, 136, 141, 156, 169, 195, 200, 210, 220, 236,
		247, 260, 267, 271, 278,
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
	NeuroScriptParserEOF               = antlr.TokenEOF
	NeuroScriptParserKW_FILE_VERSION   = 1
	NeuroScriptParserKW_DEFINE         = 2
	NeuroScriptParserKW_PROCEDURE      = 3
	NeuroScriptParserKW_END            = 4
	NeuroScriptParserKW_ENDBLOCK       = 5
	NeuroScriptParserKW_COMMENT_START  = 6
	NeuroScriptParserKW_ENDCOMMENT     = 7
	NeuroScriptParserKW_SET            = 8
	NeuroScriptParserKW_CALL           = 9
	NeuroScriptParserKW_RETURN         = 10
	NeuroScriptParserKW_IF             = 11
	NeuroScriptParserKW_THEN           = 12
	NeuroScriptParserKW_ELSE           = 13
	NeuroScriptParserKW_WHILE          = 14
	NeuroScriptParserKW_DO             = 15
	NeuroScriptParserKW_FOR            = 16
	NeuroScriptParserKW_EACH           = 17
	NeuroScriptParserKW_IN             = 18
	NeuroScriptParserKW_TOOL           = 19
	NeuroScriptParserKW_LLM            = 20
	NeuroScriptParserKW_LAST           = 21
	NeuroScriptParserKW_EVAL           = 22
	NeuroScriptParserKW_EMIT           = 23
	NeuroScriptParserKW_TRUE           = 24
	NeuroScriptParserKW_FALSE          = 25
	NeuroScriptParserCOMMENT_BLOCK     = 26
	NeuroScriptParserNUMBER_LIT        = 27
	NeuroScriptParserSTRING_LIT        = 28
	NeuroScriptParserASSIGN            = 29
	NeuroScriptParserPLUS              = 30
	NeuroScriptParserLPAREN            = 31
	NeuroScriptParserRPAREN            = 32
	NeuroScriptParserCOMMA             = 33
	NeuroScriptParserLBRACK            = 34
	NeuroScriptParserRBRACK            = 35
	NeuroScriptParserLBRACE            = 36
	NeuroScriptParserRBRACE            = 37
	NeuroScriptParserCOLON             = 38
	NeuroScriptParserDOT               = 39
	NeuroScriptParserSLASH             = 40
	NeuroScriptParserPLACEHOLDER_START = 41
	NeuroScriptParserPLACEHOLDER_END   = 42
	NeuroScriptParserEQ                = 43
	NeuroScriptParserNEQ               = 44
	NeuroScriptParserGT                = 45
	NeuroScriptParserLT                = 46
	NeuroScriptParserGTE               = 47
	NeuroScriptParserLTE               = 48
	NeuroScriptParserIDENTIFIER        = 49
	NeuroScriptParserLINE_COMMENT      = 50
	NeuroScriptParserHASH_BANG         = 51
	NeuroScriptParserNEWLINE           = 52
	NeuroScriptParserWS                = 53
)

// NeuroScriptParser rules.
const (
	NeuroScriptParserRULE_program              = 0
	NeuroScriptParserRULE_optional_newlines    = 1
	NeuroScriptParserRULE_file_version_decl    = 2
	NeuroScriptParserRULE_procedure_definition = 3
	NeuroScriptParserRULE_param_list_opt       = 4
	NeuroScriptParserRULE_param_list           = 5
	NeuroScriptParserRULE_statement_list       = 6
	NeuroScriptParserRULE_body_line            = 7
	NeuroScriptParserRULE_statement            = 8
	NeuroScriptParserRULE_simple_statement     = 9
	NeuroScriptParserRULE_block_statement      = 10
	NeuroScriptParserRULE_set_statement        = 11
	NeuroScriptParserRULE_call_statement       = 12
	NeuroScriptParserRULE_return_statement     = 13
	NeuroScriptParserRULE_emit_statement       = 14
	NeuroScriptParserRULE_if_statement         = 15
	NeuroScriptParserRULE_while_statement      = 16
	NeuroScriptParserRULE_for_each_statement   = 17
	NeuroScriptParserRULE_call_target          = 18
	NeuroScriptParserRULE_condition            = 19
	NeuroScriptParserRULE_expression           = 20
	NeuroScriptParserRULE_term                 = 21
	NeuroScriptParserRULE_primary              = 22
	NeuroScriptParserRULE_placeholder          = 23
	NeuroScriptParserRULE_literal              = 24
	NeuroScriptParserRULE_boolean_literal      = 25
	NeuroScriptParserRULE_list_literal         = 26
	NeuroScriptParserRULE_map_literal          = 27
	NeuroScriptParserRULE_expression_list_opt  = 28
	NeuroScriptParserRULE_expression_list      = 29
	NeuroScriptParserRULE_map_entry_list_opt   = 30
	NeuroScriptParserRULE_map_entry_list       = 31
	NeuroScriptParserRULE_map_entry            = 32
)

// IProgramContext is an interface to support dynamic dispatch.
type IProgramContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllOptional_newlines() []IOptional_newlinesContext
	Optional_newlines(i int) IOptional_newlinesContext
	EOF() antlr.TerminalNode
	File_version_decl() IFile_version_declContext
	AllProcedure_definition() []IProcedure_definitionContext
	Procedure_definition(i int) IProcedure_definitionContext

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

func (s *ProgramContext) AllOptional_newlines() []IOptional_newlinesContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IOptional_newlinesContext); ok {
			len++
		}
	}

	tst := make([]IOptional_newlinesContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IOptional_newlinesContext); ok {
			tst[i] = t.(IOptional_newlinesContext)
			i++
		}
	}

	return tst
}

func (s *ProgramContext) Optional_newlines(i int) IOptional_newlinesContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOptional_newlinesContext); ok {
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

	return t.(IOptional_newlinesContext)
}

func (s *ProgramContext) EOF() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserEOF, 0)
}

func (s *ProgramContext) File_version_decl() IFile_version_declContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFile_version_declContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFile_version_declContext)
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
		p.SetState(66)
		p.Optional_newlines()
	}
	p.SetState(68)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NeuroScriptParserKW_FILE_VERSION {
		{
			p.SetState(67)
			p.File_version_decl()
		}

	}
	{
		p.SetState(70)
		p.Optional_newlines()
	}
	p.SetState(74)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserKW_DEFINE {
		{
			p.SetState(71)
			p.Procedure_definition()
		}

		p.SetState(76)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(77)
		p.Optional_newlines()
	}
	{
		p.SetState(78)
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

// IOptional_newlinesContext is an interface to support dynamic dispatch.
type IOptional_newlinesContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllNEWLINE() []antlr.TerminalNode
	NEWLINE(i int) antlr.TerminalNode

	// IsOptional_newlinesContext differentiates from other interfaces.
	IsOptional_newlinesContext()
}

type Optional_newlinesContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOptional_newlinesContext() *Optional_newlinesContext {
	var p = new(Optional_newlinesContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_optional_newlines
	return p
}

func InitEmptyOptional_newlinesContext(p *Optional_newlinesContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_optional_newlines
}

func (*Optional_newlinesContext) IsOptional_newlinesContext() {}

func NewOptional_newlinesContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Optional_newlinesContext {
	var p = new(Optional_newlinesContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_optional_newlines

	return p
}

func (s *Optional_newlinesContext) GetParser() antlr.Parser { return s.parser }

func (s *Optional_newlinesContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserNEWLINE)
}

func (s *Optional_newlinesContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, i)
}

func (s *Optional_newlinesContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Optional_newlinesContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Optional_newlinesContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterOptional_newlines(s)
	}
}

func (s *Optional_newlinesContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitOptional_newlines(s)
	}
}

func (s *Optional_newlinesContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitOptional_newlines(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Optional_newlines() (localctx IOptional_newlinesContext) {
	localctx = NewOptional_newlinesContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, NeuroScriptParserRULE_optional_newlines)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(83)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(80)
				p.Match(NeuroScriptParserNEWLINE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(85)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext())
		if p.HasError() {
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

// IFile_version_declContext is an interface to support dynamic dispatch.
type IFile_version_declContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_FILE_VERSION() antlr.TerminalNode
	STRING_LIT() antlr.TerminalNode
	NEWLINE() antlr.TerminalNode

	// IsFile_version_declContext differentiates from other interfaces.
	IsFile_version_declContext()
}

type File_version_declContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFile_version_declContext() *File_version_declContext {
	var p = new(File_version_declContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_file_version_decl
	return p
}

func InitEmptyFile_version_declContext(p *File_version_declContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_file_version_decl
}

func (*File_version_declContext) IsFile_version_declContext() {}

func NewFile_version_declContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *File_version_declContext {
	var p = new(File_version_declContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_file_version_decl

	return p
}

func (s *File_version_declContext) GetParser() antlr.Parser { return s.parser }

func (s *File_version_declContext) KW_FILE_VERSION() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_FILE_VERSION, 0)
}

func (s *File_version_declContext) STRING_LIT() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserSTRING_LIT, 0)
}

func (s *File_version_declContext) NEWLINE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, 0)
}

func (s *File_version_declContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *File_version_declContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *File_version_declContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterFile_version_decl(s)
	}
}

func (s *File_version_declContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitFile_version_decl(s)
	}
}

func (s *File_version_declContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitFile_version_decl(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) File_version_decl() (localctx IFile_version_declContext) {
	localctx = NewFile_version_declContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, NeuroScriptParserRULE_file_version_decl)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(86)
		p.Match(NeuroScriptParserKW_FILE_VERSION)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(87)
		p.Match(NeuroScriptParserSTRING_LIT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(88)
		p.Match(NeuroScriptParserNEWLINE)
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

// IProcedure_definitionContext is an interface to support dynamic dispatch.
type IProcedure_definitionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_DEFINE() antlr.TerminalNode
	KW_PROCEDURE() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode
	LPAREN() antlr.TerminalNode
	Param_list_opt() IParam_list_optContext
	RPAREN() antlr.TerminalNode
	AllNEWLINE() []antlr.TerminalNode
	NEWLINE(i int) antlr.TerminalNode
	Statement_list() IStatement_listContext
	KW_END() antlr.TerminalNode
	COMMENT_BLOCK() antlr.TerminalNode

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

func (s *Procedure_definitionContext) KW_DEFINE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_DEFINE, 0)
}

func (s *Procedure_definitionContext) KW_PROCEDURE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_PROCEDURE, 0)
}

func (s *Procedure_definitionContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserIDENTIFIER, 0)
}

func (s *Procedure_definitionContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLPAREN, 0)
}

func (s *Procedure_definitionContext) Param_list_opt() IParam_list_optContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IParam_list_optContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IParam_list_optContext)
}

func (s *Procedure_definitionContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserRPAREN, 0)
}

func (s *Procedure_definitionContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserNEWLINE)
}

func (s *Procedure_definitionContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, i)
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

func (s *Procedure_definitionContext) KW_END() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_END, 0)
}

func (s *Procedure_definitionContext) COMMENT_BLOCK() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserCOMMENT_BLOCK, 0)
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
	p.EnterRule(localctx, 6, NeuroScriptParserRULE_procedure_definition)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(90)
		p.Match(NeuroScriptParserKW_DEFINE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(91)
		p.Match(NeuroScriptParserKW_PROCEDURE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(92)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(93)
		p.Match(NeuroScriptParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(94)
		p.Param_list_opt()
	}
	{
		p.SetState(95)
		p.Match(NeuroScriptParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(96)
		p.Match(NeuroScriptParserNEWLINE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(98)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NeuroScriptParserCOMMENT_BLOCK {
		{
			p.SetState(97)
			p.Match(NeuroScriptParserCOMMENT_BLOCK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(100)
		p.Statement_list()
	}
	{
		p.SetState(101)
		p.Match(NeuroScriptParserKW_END)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(103)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(102)
			p.Match(NeuroScriptParserNEWLINE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	} else if p.HasError() { // JIM
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

// IParam_list_optContext is an interface to support dynamic dispatch.
type IParam_list_optContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Param_list() IParam_listContext

	// IsParam_list_optContext differentiates from other interfaces.
	IsParam_list_optContext()
}

type Param_list_optContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyParam_list_optContext() *Param_list_optContext {
	var p = new(Param_list_optContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_param_list_opt
	return p
}

func InitEmptyParam_list_optContext(p *Param_list_optContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_param_list_opt
}

func (*Param_list_optContext) IsParam_list_optContext() {}

func NewParam_list_optContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Param_list_optContext {
	var p = new(Param_list_optContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_param_list_opt

	return p
}

func (s *Param_list_optContext) GetParser() antlr.Parser { return s.parser }

func (s *Param_list_optContext) Param_list() IParam_listContext {
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

func (s *Param_list_optContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Param_list_optContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Param_list_optContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterParam_list_opt(s)
	}
}

func (s *Param_list_optContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitParam_list_opt(s)
	}
}

func (s *Param_list_optContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitParam_list_opt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Param_list_opt() (localctx IParam_list_optContext) {
	localctx = NewParam_list_optContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, NeuroScriptParserRULE_param_list_opt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(106)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NeuroScriptParserIDENTIFIER {
		{
			p.SetState(105)
			p.Param_list()
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
	p.EnterRule(localctx, 10, NeuroScriptParserRULE_param_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(108)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(113)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserCOMMA {
		{
			p.SetState(109)
			p.Match(NeuroScriptParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(110)
			p.Match(NeuroScriptParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(115)
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
	p.EnterRule(localctx, 12, NeuroScriptParserRULE_statement_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(119)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&4503599635844864) != 0 {
		{
			p.SetState(116)
			p.Body_line()
		}

		p.SetState(121)
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
	p.EnterRule(localctx, 14, NeuroScriptParserRULE_body_line)
	p.SetState(126)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_SET, NeuroScriptParserKW_CALL, NeuroScriptParserKW_RETURN, NeuroScriptParserKW_IF, NeuroScriptParserKW_WHILE, NeuroScriptParserKW_FOR, NeuroScriptParserKW_EMIT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(122)
			p.Statement()
		}
		{
			p.SetState(123)
			p.Match(NeuroScriptParserNEWLINE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserNEWLINE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(125)
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
	p.EnterRule(localctx, 16, NeuroScriptParserRULE_statement)
	p.SetState(130)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_SET, NeuroScriptParserKW_CALL, NeuroScriptParserKW_RETURN, NeuroScriptParserKW_EMIT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(128)
			p.Simple_statement()
		}

	case NeuroScriptParserKW_IF, NeuroScriptParserKW_WHILE, NeuroScriptParserKW_FOR:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(129)
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
	p.EnterRule(localctx, 18, NeuroScriptParserRULE_simple_statement)
	p.SetState(136)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_SET:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(132)
			p.Set_statement()
		}

	case NeuroScriptParserKW_CALL:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(133)
			p.Call_statement()
		}

	case NeuroScriptParserKW_RETURN:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(134)
			p.Return_statement()
		}

	case NeuroScriptParserKW_EMIT:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(135)
			p.Emit_statement()
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
	p.EnterRule(localctx, 20, NeuroScriptParserRULE_block_statement)
	p.SetState(141)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_IF:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(138)
			p.If_statement()
		}

	case NeuroScriptParserKW_WHILE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(139)
			p.While_statement()
		}

	case NeuroScriptParserKW_FOR:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(140)
			p.For_each_statement()
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
	p.EnterRule(localctx, 22, NeuroScriptParserRULE_set_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(143)
		p.Match(NeuroScriptParserKW_SET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(144)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(145)
		p.Match(NeuroScriptParserASSIGN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(146)
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
	Call_target() ICall_targetContext
	LPAREN() antlr.TerminalNode
	Expression_list_opt() IExpression_list_optContext
	RPAREN() antlr.TerminalNode

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

func (s *Call_statementContext) Call_target() ICall_targetContext {
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

func (s *Call_statementContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLPAREN, 0)
}

func (s *Call_statementContext) Expression_list_opt() IExpression_list_optContext {
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

func (s *Call_statementContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserRPAREN, 0)
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
	p.EnterRule(localctx, 24, NeuroScriptParserRULE_call_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(148)
		p.Match(NeuroScriptParserKW_CALL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(149)
		p.Call_target()
	}
	{
		p.SetState(150)
		p.Match(NeuroScriptParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(151)
		p.Expression_list_opt()
	}
	{
		p.SetState(152)
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

// IReturn_statementContext is an interface to support dynamic dispatch.
type IReturn_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_RETURN() antlr.TerminalNode
	Expression() IExpressionContext

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

func (s *Return_statementContext) Expression() IExpressionContext {
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
	p.EnterRule(localctx, 26, NeuroScriptParserRULE_return_statement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(154)
		p.Match(NeuroScriptParserKW_RETURN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(156)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&565237482782720) != 0 {
		{
			p.SetState(155)
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
	p.EnterRule(localctx, 28, NeuroScriptParserRULE_emit_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(158)
		p.Match(NeuroScriptParserKW_EMIT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(159)
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

// IIf_statementContext is an interface to support dynamic dispatch.
type IIf_statementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetIf_body returns the if_body rule contexts.
	GetIf_body() IStatement_listContext

	// GetElse_body returns the else_body rule contexts.
	GetElse_body() IStatement_listContext

	// SetIf_body sets the if_body rule contexts.
	SetIf_body(IStatement_listContext)

	// SetElse_body sets the else_body rule contexts.
	SetElse_body(IStatement_listContext)

	// Getter signatures
	KW_IF() antlr.TerminalNode
	Condition() IConditionContext
	KW_THEN() antlr.TerminalNode
	AllNEWLINE() []antlr.TerminalNode
	NEWLINE(i int) antlr.TerminalNode
	KW_ENDBLOCK() antlr.TerminalNode
	AllStatement_list() []IStatement_listContext
	Statement_list(i int) IStatement_listContext
	KW_ELSE() antlr.TerminalNode

	// IsIf_statementContext differentiates from other interfaces.
	IsIf_statementContext()
}

type If_statementContext struct {
	antlr.BaseParserRuleContext
	parser    antlr.Parser
	if_body   IStatement_listContext
	else_body IStatement_listContext
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

func (s *If_statementContext) GetIf_body() IStatement_listContext { return s.if_body }

func (s *If_statementContext) GetElse_body() IStatement_listContext { return s.else_body }

func (s *If_statementContext) SetIf_body(v IStatement_listContext) { s.if_body = v }

func (s *If_statementContext) SetElse_body(v IStatement_listContext) { s.else_body = v }

func (s *If_statementContext) KW_IF() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_IF, 0)
}

func (s *If_statementContext) Condition() IConditionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *If_statementContext) KW_THEN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_THEN, 0)
}

func (s *If_statementContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserNEWLINE)
}

func (s *If_statementContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, i)
}

func (s *If_statementContext) KW_ENDBLOCK() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ENDBLOCK, 0)
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
	p.EnterRule(localctx, 30, NeuroScriptParserRULE_if_statement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(161)
		p.Match(NeuroScriptParserKW_IF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(162)
		p.Condition()
	}
	{
		p.SetState(163)
		p.Match(NeuroScriptParserKW_THEN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(164)
		p.Match(NeuroScriptParserNEWLINE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(165)

		var _x = p.Statement_list()

		localctx.(*If_statementContext).if_body = _x
	}
	p.SetState(169)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NeuroScriptParserKW_ELSE {
		{
			p.SetState(166)
			p.Match(NeuroScriptParserKW_ELSE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(167)
			p.Match(NeuroScriptParserNEWLINE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(168)

			var _x = p.Statement_list()

			localctx.(*If_statementContext).else_body = _x
		}

	}
	{
		p.SetState(171)
		p.Match(NeuroScriptParserKW_ENDBLOCK)
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
	Condition() IConditionContext
	KW_DO() antlr.TerminalNode
	NEWLINE() antlr.TerminalNode
	Statement_list() IStatement_listContext
	KW_ENDBLOCK() antlr.TerminalNode

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

func (s *While_statementContext) Condition() IConditionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *While_statementContext) KW_DO() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_DO, 0)
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

func (s *While_statementContext) KW_ENDBLOCK() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ENDBLOCK, 0)
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
	p.EnterRule(localctx, 32, NeuroScriptParserRULE_while_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(173)
		p.Match(NeuroScriptParserKW_WHILE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(174)
		p.Condition()
	}
	{
		p.SetState(175)
		p.Match(NeuroScriptParserKW_DO)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(176)
		p.Match(NeuroScriptParserNEWLINE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(177)
		p.Statement_list()
	}
	{
		p.SetState(178)
		p.Match(NeuroScriptParserKW_ENDBLOCK)
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
	KW_DO() antlr.TerminalNode
	NEWLINE() antlr.TerminalNode
	Statement_list() IStatement_listContext
	KW_ENDBLOCK() antlr.TerminalNode

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

func (s *For_each_statementContext) KW_DO() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_DO, 0)
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

func (s *For_each_statementContext) KW_ENDBLOCK() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ENDBLOCK, 0)
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
	p.EnterRule(localctx, 34, NeuroScriptParserRULE_for_each_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(180)
		p.Match(NeuroScriptParserKW_FOR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(181)
		p.Match(NeuroScriptParserKW_EACH)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(182)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(183)
		p.Match(NeuroScriptParserKW_IN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(184)
		p.Expression()
	}
	{
		p.SetState(185)
		p.Match(NeuroScriptParserKW_DO)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(186)
		p.Match(NeuroScriptParserNEWLINE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(187)
		p.Statement_list()
	}
	{
		p.SetState(188)
		p.Match(NeuroScriptParserKW_ENDBLOCK)
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

// ICall_targetContext is an interface to support dynamic dispatch.
type ICall_targetContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	KW_TOOL() antlr.TerminalNode
	DOT() antlr.TerminalNode
	KW_LLM() antlr.TerminalNode

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

func (s *Call_targetContext) KW_LLM() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_LLM, 0)
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
	p.EnterRule(localctx, 36, NeuroScriptParserRULE_call_target)
	p.SetState(195)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(190)
			p.Match(NeuroScriptParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_TOOL:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(191)
			p.Match(NeuroScriptParserKW_TOOL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(192)
			p.Match(NeuroScriptParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(193)
			p.Match(NeuroScriptParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_LLM:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(194)
			p.Match(NeuroScriptParserKW_LLM)
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

// IConditionContext is an interface to support dynamic dispatch.
type IConditionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	EQ() antlr.TerminalNode
	NEQ() antlr.TerminalNode
	GT() antlr.TerminalNode
	LT() antlr.TerminalNode
	GTE() antlr.TerminalNode
	LTE() antlr.TerminalNode

	// IsConditionContext differentiates from other interfaces.
	IsConditionContext()
}

type ConditionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyConditionContext() *ConditionContext {
	var p = new(ConditionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_condition
	return p
}

func InitEmptyConditionContext(p *ConditionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_condition
}

func (*ConditionContext) IsConditionContext() {}

func NewConditionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ConditionContext {
	var p = new(ConditionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_condition

	return p
}

func (s *ConditionContext) GetParser() antlr.Parser { return s.parser }

func (s *ConditionContext) AllExpression() []IExpressionContext {
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

func (s *ConditionContext) Expression(i int) IExpressionContext {
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

func (s *ConditionContext) EQ() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserEQ, 0)
}

func (s *ConditionContext) NEQ() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEQ, 0)
}

func (s *ConditionContext) GT() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserGT, 0)
}

func (s *ConditionContext) LT() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLT, 0)
}

func (s *ConditionContext) GTE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserGTE, 0)
}

func (s *ConditionContext) LTE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLTE, 0)
}

func (s *ConditionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ConditionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterCondition(s)
	}
}

func (s *ConditionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitCondition(s)
	}
}

func (s *ConditionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitCondition(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Condition() (localctx IConditionContext) {
	localctx = NewConditionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, NeuroScriptParserRULE_condition)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(197)
		p.Expression()
	}
	p.SetState(200)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&554153860399104) != 0 {
		{
			p.SetState(198)
			_la = p.GetTokenStream().LA(1)

			if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&554153860399104) != 0) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(199)
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

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllTerm() []ITermContext
	Term(i int) ITermContext
	AllOptional_newlines() []IOptional_newlinesContext
	Optional_newlines(i int) IOptional_newlinesContext
	AllPLUS() []antlr.TerminalNode
	PLUS(i int) antlr.TerminalNode

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

func (s *ExpressionContext) AllTerm() []ITermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITermContext); ok {
			len++
		}
	}

	tst := make([]ITermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITermContext); ok {
			tst[i] = t.(ITermContext)
			i++
		}
	}

	return tst
}

func (s *ExpressionContext) Term(i int) ITermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
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

	return t.(ITermContext)
}

func (s *ExpressionContext) AllOptional_newlines() []IOptional_newlinesContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IOptional_newlinesContext); ok {
			len++
		}
	}

	tst := make([]IOptional_newlinesContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IOptional_newlinesContext); ok {
			tst[i] = t.(IOptional_newlinesContext)
			i++
		}
	}

	return tst
}

func (s *ExpressionContext) Optional_newlines(i int) IOptional_newlinesContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOptional_newlinesContext); ok {
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

	return t.(IOptional_newlinesContext)
}

func (s *ExpressionContext) AllPLUS() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserPLUS)
}

func (s *ExpressionContext) PLUS(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserPLUS, i)
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
	p.EnterRule(localctx, 40, NeuroScriptParserRULE_expression)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(202)
		p.Term()
	}
	p.SetState(210)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 16, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(203)
				p.Optional_newlines()
			}
			{
				p.SetState(204)
				p.Match(NeuroScriptParserPLUS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(205)
				p.Optional_newlines()
			}
			{
				p.SetState(206)
				p.Term()
			}

		}
		p.SetState(212)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 16, p.GetParserRuleContext())
		if p.HasError() {
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

// ITermContext is an interface to support dynamic dispatch.
type ITermContext interface {
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

	// IsTermContext differentiates from other interfaces.
	IsTermContext()
}

type TermContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTermContext() *TermContext {
	var p = new(TermContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_term
	return p
}

func InitEmptyTermContext(p *TermContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroScriptParserRULE_term
}

func (*TermContext) IsTermContext() {}

func NewTermContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TermContext {
	var p = new(TermContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroScriptParserRULE_term

	return p
}

func (s *TermContext) GetParser() antlr.Parser { return s.parser }

func (s *TermContext) Primary() IPrimaryContext {
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

func (s *TermContext) AllLBRACK() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserLBRACK)
}

func (s *TermContext) LBRACK(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLBRACK, i)
}

func (s *TermContext) AllExpression() []IExpressionContext {
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

func (s *TermContext) Expression(i int) IExpressionContext {
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

func (s *TermContext) AllRBRACK() []antlr.TerminalNode {
	return s.GetTokens(NeuroScriptParserRBRACK)
}

func (s *TermContext) RBRACK(i int) antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserRBRACK, i)
}

func (s *TermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TermContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TermContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.EnterTerm(s)
	}
}

func (s *TermContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroScriptListener); ok {
		listenerT.ExitTerm(s)
	}
}

func (s *TermContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroScriptVisitor:
		return t.VisitTerm(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroScriptParser) Term() (localctx ITermContext) {
	localctx = NewTermContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 42, NeuroScriptParserRULE_term)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(213)
		p.Primary()
	}
	p.SetState(220)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserLBRACK {
		{
			p.SetState(214)
			p.Match(NeuroScriptParserLBRACK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(215)
			p.Expression()
		}
		{
			p.SetState(216)
			p.Match(NeuroScriptParserRBRACK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(222)
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
	p.EnterRule(localctx, 44, NeuroScriptParserRULE_primary)
	p.SetState(236)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_TRUE, NeuroScriptParserKW_FALSE, NeuroScriptParserNUMBER_LIT, NeuroScriptParserSTRING_LIT, NeuroScriptParserLBRACK, NeuroScriptParserLBRACE:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(223)
			p.Literal()
		}

	case NeuroScriptParserPLACEHOLDER_START:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(224)
			p.Placeholder()
		}

	case NeuroScriptParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(225)
			p.Match(NeuroScriptParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_LAST:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(226)
			p.Match(NeuroScriptParserKW_LAST)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_EVAL:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(227)
			p.Match(NeuroScriptParserKW_EVAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(228)
			p.Match(NeuroScriptParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(229)
			p.Expression()
		}
		{
			p.SetState(230)
			p.Match(NeuroScriptParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserLPAREN:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(232)
			p.Match(NeuroScriptParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(233)
			p.Expression()
		}
		{
			p.SetState(234)
			p.Match(NeuroScriptParserRPAREN)
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
	p.EnterRule(localctx, 46, NeuroScriptParserRULE_placeholder)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(238)
		p.Match(NeuroScriptParserPLACEHOLDER_START)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(239)
		_la = p.GetTokenStream().LA(1)

		if !(_la == NeuroScriptParserKW_LAST || _la == NeuroScriptParserIDENTIFIER) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	{
		p.SetState(240)
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
	NUMBER_LIT() antlr.TerminalNode
	List_literal() IList_literalContext
	Map_literal() IMap_literalContext
	Boolean_literal() IBoolean_literalContext

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
	p.EnterRule(localctx, 48, NeuroScriptParserRULE_literal)
	p.SetState(247)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserSTRING_LIT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(242)
			p.Match(NeuroScriptParserSTRING_LIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserNUMBER_LIT:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(243)
			p.Match(NeuroScriptParserNUMBER_LIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserLBRACK:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(244)
			p.List_literal()
		}

	case NeuroScriptParserLBRACE:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(245)
			p.Map_literal()
		}

	case NeuroScriptParserKW_TRUE, NeuroScriptParserKW_FALSE:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(246)
			p.Boolean_literal()
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
	p.EnterRule(localctx, 50, NeuroScriptParserRULE_boolean_literal)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(249)
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
	p.EnterRule(localctx, 52, NeuroScriptParserRULE_list_literal)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(251)
		p.Match(NeuroScriptParserLBRACK)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(252)
		p.Expression_list_opt()
	}
	{
		p.SetState(253)
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
	p.EnterRule(localctx, 54, NeuroScriptParserRULE_map_literal)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(255)
		p.Match(NeuroScriptParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(256)
		p.Map_entry_list_opt()
	}
	{
		p.SetState(257)
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
	p.EnterRule(localctx, 56, NeuroScriptParserRULE_expression_list_opt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(260)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&565237482782720) != 0 {
		{
			p.SetState(259)
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
	p.EnterRule(localctx, 58, NeuroScriptParserRULE_expression_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(262)
		p.Expression()
	}
	p.SetState(267)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserCOMMA {
		{
			p.SetState(263)
			p.Match(NeuroScriptParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(264)
			p.Expression()
		}

		p.SetState(269)
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
	p.EnterRule(localctx, 60, NeuroScriptParserRULE_map_entry_list_opt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(271)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NeuroScriptParserSTRING_LIT {
		{
			p.SetState(270)
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
	p.EnterRule(localctx, 62, NeuroScriptParserRULE_map_entry_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(273)
		p.Map_entry()
	}
	p.SetState(278)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserCOMMA {
		{
			p.SetState(274)
			p.Match(NeuroScriptParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(275)
			p.Map_entry()
		}

		p.SetState(280)
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
	p.EnterRule(localctx, 64, NeuroScriptParserRULE_map_entry)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(281)
		p.Match(NeuroScriptParserSTRING_LIT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(282)
		p.Match(NeuroScriptParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(283)
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
