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
		"", "'SPLAT'", "'PROCEDURE'", "'END'", "'ENDBLOCK'", "'COMMENT:'", "'ENDCOMMENT'",
		"'SET'", "'CALL'", "'RETURN'", "'IF'", "'THEN'", "'ELSE'", "'WHILE'",
		"'DO'", "'FOR'", "'EACH'", "'IN'", "'TOOL'", "'LLM'", "'__last_call_result'",
		"'EMIT'", "", "", "", "'='", "'+'", "'('", "')'", "','", "'['", "']'",
		"'{'", "'}'", "':'", "'.'", "'/'", "'{{'", "'}}'", "'=='", "'!='", "'>'",
		"'<'", "'>='", "'<='",
	}
	staticData.SymbolicNames = []string{
		"", "KW_STARTPROC", "KW_PROCEDURE", "KW_END", "KW_ENDBLOCK", "KW_COMMENT_START",
		"KW_ENDCOMMENT", "KW_SET", "KW_CALL", "KW_RETURN", "KW_IF", "KW_THEN",
		"KW_ELSE", "KW_WHILE", "KW_DO", "KW_FOR", "KW_EACH", "KW_IN", "KW_TOOL",
		"KW_LLM", "KW_LAST_CALL_RESULT", "KW_EMIT", "COMMENT_BLOCK", "NUMBER_LIT",
		"STRING_LIT", "ASSIGN", "PLUS", "LPAREN", "RPAREN", "COMMA", "LBRACK",
		"RBRACK", "LBRACE", "RBRACE", "COLON", "DOT", "SLASH", "PLACEHOLDER_START",
		"PLACEHOLDER_END", "EQ", "NEQ", "GT", "LT", "GTE", "LTE", "IDENTIFIER",
		"LINE_COMMENT", "NEWLINE", "WS", "LINE_CONTINUATION",
	}
	staticData.RuleNames = []string{
		"program", "optional_newlines", "procedure_definition", "param_list_opt",
		"param_list", "statement_list", "body_line", "statement", "simple_statement",
		"block_statement", "set_statement", "call_statement", "return_statement",
		"emit_statement", "if_statement", "while_statement", "for_each_statement",
		"call_target", "condition", "expression", "term", "placeholder", "literal",
		"list_literal", "map_literal", "expression_list_opt", "expression_list",
		"map_entry_list_opt", "map_entry_list", "map_entry",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 49, 246, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7, 20, 2,
		21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25, 2, 26,
		7, 26, 2, 27, 7, 27, 2, 28, 7, 28, 2, 29, 7, 29, 1, 0, 1, 0, 5, 0, 63,
		8, 0, 10, 0, 12, 0, 66, 9, 0, 1, 0, 1, 0, 1, 0, 1, 1, 5, 1, 72, 8, 1, 10,
		1, 12, 1, 75, 9, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 3,
		2, 85, 8, 2, 1, 2, 1, 2, 1, 2, 3, 2, 90, 8, 2, 1, 3, 3, 3, 93, 8, 3, 1,
		4, 1, 4, 1, 4, 5, 4, 98, 8, 4, 10, 4, 12, 4, 101, 9, 4, 1, 5, 5, 5, 104,
		8, 5, 10, 5, 12, 5, 107, 9, 5, 1, 6, 1, 6, 1, 6, 1, 6, 3, 6, 113, 8, 6,
		1, 7, 1, 7, 3, 7, 117, 8, 7, 1, 8, 1, 8, 1, 8, 1, 8, 3, 8, 123, 8, 8, 1,
		9, 1, 9, 1, 9, 3, 9, 128, 8, 9, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 11,
		1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 12, 1, 12, 3, 12, 143, 8, 12, 1,
		13, 1, 13, 1, 13, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 15,
		1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 16, 1, 16, 1, 16, 1, 16, 1,
		16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17,
		3, 17, 177, 8, 17, 1, 18, 1, 18, 1, 18, 3, 18, 182, 8, 18, 1, 19, 1, 19,
		1, 19, 5, 19, 187, 8, 19, 10, 19, 12, 19, 190, 9, 19, 1, 20, 1, 20, 1,
		20, 1, 20, 1, 20, 1, 20, 1, 20, 1, 20, 3, 20, 200, 8, 20, 1, 21, 1, 21,
		1, 21, 1, 21, 1, 22, 1, 22, 1, 22, 1, 22, 3, 22, 210, 8, 22, 1, 23, 1,
		23, 1, 23, 1, 23, 1, 24, 1, 24, 1, 24, 1, 24, 1, 25, 3, 25, 221, 8, 25,
		1, 26, 1, 26, 1, 26, 5, 26, 226, 8, 26, 10, 26, 12, 26, 229, 9, 26, 1,
		27, 3, 27, 232, 8, 27, 1, 28, 1, 28, 1, 28, 5, 28, 237, 8, 28, 10, 28,
		12, 28, 240, 9, 28, 1, 29, 1, 29, 1, 29, 1, 29, 1, 29, 0, 0, 30, 0, 2,
		4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40,
		42, 44, 46, 48, 50, 52, 54, 56, 58, 0, 1, 1, 0, 39, 44, 245, 0, 60, 1,
		0, 0, 0, 2, 73, 1, 0, 0, 0, 4, 76, 1, 0, 0, 0, 6, 92, 1, 0, 0, 0, 8, 94,
		1, 0, 0, 0, 10, 105, 1, 0, 0, 0, 12, 112, 1, 0, 0, 0, 14, 116, 1, 0, 0,
		0, 16, 122, 1, 0, 0, 0, 18, 127, 1, 0, 0, 0, 20, 129, 1, 0, 0, 0, 22, 134,
		1, 0, 0, 0, 24, 140, 1, 0, 0, 0, 26, 144, 1, 0, 0, 0, 28, 147, 1, 0, 0,
		0, 30, 154, 1, 0, 0, 0, 32, 161, 1, 0, 0, 0, 34, 176, 1, 0, 0, 0, 36, 178,
		1, 0, 0, 0, 38, 183, 1, 0, 0, 0, 40, 199, 1, 0, 0, 0, 42, 201, 1, 0, 0,
		0, 44, 209, 1, 0, 0, 0, 46, 211, 1, 0, 0, 0, 48, 215, 1, 0, 0, 0, 50, 220,
		1, 0, 0, 0, 52, 222, 1, 0, 0, 0, 54, 231, 1, 0, 0, 0, 56, 233, 1, 0, 0,
		0, 58, 241, 1, 0, 0, 0, 60, 64, 3, 2, 1, 0, 61, 63, 3, 4, 2, 0, 62, 61,
		1, 0, 0, 0, 63, 66, 1, 0, 0, 0, 64, 62, 1, 0, 0, 0, 64, 65, 1, 0, 0, 0,
		65, 67, 1, 0, 0, 0, 66, 64, 1, 0, 0, 0, 67, 68, 3, 2, 1, 0, 68, 69, 5,
		0, 0, 1, 69, 1, 1, 0, 0, 0, 70, 72, 5, 47, 0, 0, 71, 70, 1, 0, 0, 0, 72,
		75, 1, 0, 0, 0, 73, 71, 1, 0, 0, 0, 73, 74, 1, 0, 0, 0, 74, 3, 1, 0, 0,
		0, 75, 73, 1, 0, 0, 0, 76, 77, 5, 1, 0, 0, 77, 78, 5, 2, 0, 0, 78, 79,
		5, 45, 0, 0, 79, 80, 5, 27, 0, 0, 80, 81, 3, 6, 3, 0, 81, 82, 5, 28, 0,
		0, 82, 84, 5, 47, 0, 0, 83, 85, 5, 22, 0, 0, 84, 83, 1, 0, 0, 0, 84, 85,
		1, 0, 0, 0, 85, 86, 1, 0, 0, 0, 86, 87, 3, 10, 5, 0, 87, 89, 5, 3, 0, 0,
		88, 90, 5, 47, 0, 0, 89, 88, 1, 0, 0, 0, 89, 90, 1, 0, 0, 0, 90, 5, 1,
		0, 0, 0, 91, 93, 3, 8, 4, 0, 92, 91, 1, 0, 0, 0, 92, 93, 1, 0, 0, 0, 93,
		7, 1, 0, 0, 0, 94, 99, 5, 45, 0, 0, 95, 96, 5, 29, 0, 0, 96, 98, 5, 45,
		0, 0, 97, 95, 1, 0, 0, 0, 98, 101, 1, 0, 0, 0, 99, 97, 1, 0, 0, 0, 99,
		100, 1, 0, 0, 0, 100, 9, 1, 0, 0, 0, 101, 99, 1, 0, 0, 0, 102, 104, 3,
		12, 6, 0, 103, 102, 1, 0, 0, 0, 104, 107, 1, 0, 0, 0, 105, 103, 1, 0, 0,
		0, 105, 106, 1, 0, 0, 0, 106, 11, 1, 0, 0, 0, 107, 105, 1, 0, 0, 0, 108,
		109, 3, 14, 7, 0, 109, 110, 5, 47, 0, 0, 110, 113, 1, 0, 0, 0, 111, 113,
		5, 47, 0, 0, 112, 108, 1, 0, 0, 0, 112, 111, 1, 0, 0, 0, 113, 13, 1, 0,
		0, 0, 114, 117, 3, 16, 8, 0, 115, 117, 3, 18, 9, 0, 116, 114, 1, 0, 0,
		0, 116, 115, 1, 0, 0, 0, 117, 15, 1, 0, 0, 0, 118, 123, 3, 20, 10, 0, 119,
		123, 3, 22, 11, 0, 120, 123, 3, 24, 12, 0, 121, 123, 3, 26, 13, 0, 122,
		118, 1, 0, 0, 0, 122, 119, 1, 0, 0, 0, 122, 120, 1, 0, 0, 0, 122, 121,
		1, 0, 0, 0, 123, 17, 1, 0, 0, 0, 124, 128, 3, 28, 14, 0, 125, 128, 3, 30,
		15, 0, 126, 128, 3, 32, 16, 0, 127, 124, 1, 0, 0, 0, 127, 125, 1, 0, 0,
		0, 127, 126, 1, 0, 0, 0, 128, 19, 1, 0, 0, 0, 129, 130, 5, 7, 0, 0, 130,
		131, 5, 45, 0, 0, 131, 132, 5, 25, 0, 0, 132, 133, 3, 38, 19, 0, 133, 21,
		1, 0, 0, 0, 134, 135, 5, 8, 0, 0, 135, 136, 3, 34, 17, 0, 136, 137, 5,
		27, 0, 0, 137, 138, 3, 50, 25, 0, 138, 139, 5, 28, 0, 0, 139, 23, 1, 0,
		0, 0, 140, 142, 5, 9, 0, 0, 141, 143, 3, 38, 19, 0, 142, 141, 1, 0, 0,
		0, 142, 143, 1, 0, 0, 0, 143, 25, 1, 0, 0, 0, 144, 145, 5, 21, 0, 0, 145,
		146, 3, 38, 19, 0, 146, 27, 1, 0, 0, 0, 147, 148, 5, 10, 0, 0, 148, 149,
		3, 36, 18, 0, 149, 150, 5, 11, 0, 0, 150, 151, 5, 47, 0, 0, 151, 152, 3,
		10, 5, 0, 152, 153, 5, 4, 0, 0, 153, 29, 1, 0, 0, 0, 154, 155, 5, 13, 0,
		0, 155, 156, 3, 36, 18, 0, 156, 157, 5, 14, 0, 0, 157, 158, 5, 47, 0, 0,
		158, 159, 3, 10, 5, 0, 159, 160, 5, 4, 0, 0, 160, 31, 1, 0, 0, 0, 161,
		162, 5, 15, 0, 0, 162, 163, 5, 16, 0, 0, 163, 164, 5, 45, 0, 0, 164, 165,
		5, 17, 0, 0, 165, 166, 3, 38, 19, 0, 166, 167, 5, 14, 0, 0, 167, 168, 5,
		47, 0, 0, 168, 169, 3, 10, 5, 0, 169, 170, 5, 4, 0, 0, 170, 33, 1, 0, 0,
		0, 171, 177, 5, 45, 0, 0, 172, 173, 5, 18, 0, 0, 173, 174, 5, 35, 0, 0,
		174, 177, 5, 45, 0, 0, 175, 177, 5, 19, 0, 0, 176, 171, 1, 0, 0, 0, 176,
		172, 1, 0, 0, 0, 176, 175, 1, 0, 0, 0, 177, 35, 1, 0, 0, 0, 178, 181, 3,
		38, 19, 0, 179, 180, 7, 0, 0, 0, 180, 182, 3, 38, 19, 0, 181, 179, 1, 0,
		0, 0, 181, 182, 1, 0, 0, 0, 182, 37, 1, 0, 0, 0, 183, 188, 3, 40, 20, 0,
		184, 185, 5, 26, 0, 0, 185, 187, 3, 40, 20, 0, 186, 184, 1, 0, 0, 0, 187,
		190, 1, 0, 0, 0, 188, 186, 1, 0, 0, 0, 188, 189, 1, 0, 0, 0, 189, 39, 1,
		0, 0, 0, 190, 188, 1, 0, 0, 0, 191, 200, 3, 44, 22, 0, 192, 200, 3, 42,
		21, 0, 193, 200, 5, 45, 0, 0, 194, 200, 5, 20, 0, 0, 195, 196, 5, 27, 0,
		0, 196, 197, 3, 38, 19, 0, 197, 198, 5, 28, 0, 0, 198, 200, 1, 0, 0, 0,
		199, 191, 1, 0, 0, 0, 199, 192, 1, 0, 0, 0, 199, 193, 1, 0, 0, 0, 199,
		194, 1, 0, 0, 0, 199, 195, 1, 0, 0, 0, 200, 41, 1, 0, 0, 0, 201, 202, 5,
		37, 0, 0, 202, 203, 5, 45, 0, 0, 203, 204, 5, 38, 0, 0, 204, 43, 1, 0,
		0, 0, 205, 210, 5, 24, 0, 0, 206, 210, 5, 23, 0, 0, 207, 210, 3, 46, 23,
		0, 208, 210, 3, 48, 24, 0, 209, 205, 1, 0, 0, 0, 209, 206, 1, 0, 0, 0,
		209, 207, 1, 0, 0, 0, 209, 208, 1, 0, 0, 0, 210, 45, 1, 0, 0, 0, 211, 212,
		5, 30, 0, 0, 212, 213, 3, 50, 25, 0, 213, 214, 5, 31, 0, 0, 214, 47, 1,
		0, 0, 0, 215, 216, 5, 32, 0, 0, 216, 217, 3, 54, 27, 0, 217, 218, 5, 33,
		0, 0, 218, 49, 1, 0, 0, 0, 219, 221, 3, 52, 26, 0, 220, 219, 1, 0, 0, 0,
		220, 221, 1, 0, 0, 0, 221, 51, 1, 0, 0, 0, 222, 227, 3, 38, 19, 0, 223,
		224, 5, 29, 0, 0, 224, 226, 3, 38, 19, 0, 225, 223, 1, 0, 0, 0, 226, 229,
		1, 0, 0, 0, 227, 225, 1, 0, 0, 0, 227, 228, 1, 0, 0, 0, 228, 53, 1, 0,
		0, 0, 229, 227, 1, 0, 0, 0, 230, 232, 3, 56, 28, 0, 231, 230, 1, 0, 0,
		0, 231, 232, 1, 0, 0, 0, 232, 55, 1, 0, 0, 0, 233, 238, 3, 58, 29, 0, 234,
		235, 5, 29, 0, 0, 235, 237, 3, 58, 29, 0, 236, 234, 1, 0, 0, 0, 237, 240,
		1, 0, 0, 0, 238, 236, 1, 0, 0, 0, 238, 239, 1, 0, 0, 0, 239, 57, 1, 0,
		0, 0, 240, 238, 1, 0, 0, 0, 241, 242, 5, 24, 0, 0, 242, 243, 5, 34, 0,
		0, 243, 244, 3, 38, 19, 0, 244, 59, 1, 0, 0, 0, 21, 64, 73, 84, 89, 92,
		99, 105, 112, 116, 122, 127, 142, 176, 181, 188, 199, 209, 220, 227, 231,
		238,
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
	NeuroScriptParserEOF                 = antlr.TokenEOF
	NeuroScriptParserKW_STARTPROC        = 1
	NeuroScriptParserKW_PROCEDURE        = 2
	NeuroScriptParserKW_END              = 3
	NeuroScriptParserKW_ENDBLOCK         = 4
	NeuroScriptParserKW_COMMENT_START    = 5
	NeuroScriptParserKW_ENDCOMMENT       = 6
	NeuroScriptParserKW_SET              = 7
	NeuroScriptParserKW_CALL             = 8
	NeuroScriptParserKW_RETURN           = 9
	NeuroScriptParserKW_IF               = 10
	NeuroScriptParserKW_THEN             = 11
	NeuroScriptParserKW_ELSE             = 12
	NeuroScriptParserKW_WHILE            = 13
	NeuroScriptParserKW_DO               = 14
	NeuroScriptParserKW_FOR              = 15
	NeuroScriptParserKW_EACH             = 16
	NeuroScriptParserKW_IN               = 17
	NeuroScriptParserKW_TOOL             = 18
	NeuroScriptParserKW_LLM              = 19
	NeuroScriptParserKW_LAST_CALL_RESULT = 20
	NeuroScriptParserKW_EMIT             = 21
	NeuroScriptParserCOMMENT_BLOCK       = 22
	NeuroScriptParserNUMBER_LIT          = 23
	NeuroScriptParserSTRING_LIT          = 24
	NeuroScriptParserASSIGN              = 25
	NeuroScriptParserPLUS                = 26
	NeuroScriptParserLPAREN              = 27
	NeuroScriptParserRPAREN              = 28
	NeuroScriptParserCOMMA               = 29
	NeuroScriptParserLBRACK              = 30
	NeuroScriptParserRBRACK              = 31
	NeuroScriptParserLBRACE              = 32
	NeuroScriptParserRBRACE              = 33
	NeuroScriptParserCOLON               = 34
	NeuroScriptParserDOT                 = 35
	NeuroScriptParserSLASH               = 36
	NeuroScriptParserPLACEHOLDER_START   = 37
	NeuroScriptParserPLACEHOLDER_END     = 38
	NeuroScriptParserEQ                  = 39
	NeuroScriptParserNEQ                 = 40
	NeuroScriptParserGT                  = 41
	NeuroScriptParserLT                  = 42
	NeuroScriptParserGTE                 = 43
	NeuroScriptParserLTE                 = 44
	NeuroScriptParserIDENTIFIER          = 45
	NeuroScriptParserLINE_COMMENT        = 46
	NeuroScriptParserNEWLINE             = 47
	NeuroScriptParserWS                  = 48
	NeuroScriptParserLINE_CONTINUATION   = 49
)

// NeuroScriptParser rules.
const (
	NeuroScriptParserRULE_program              = 0
	NeuroScriptParserRULE_optional_newlines    = 1
	NeuroScriptParserRULE_procedure_definition = 2
	NeuroScriptParserRULE_param_list_opt       = 3
	NeuroScriptParserRULE_param_list           = 4
	NeuroScriptParserRULE_statement_list       = 5
	NeuroScriptParserRULE_body_line            = 6
	NeuroScriptParserRULE_statement            = 7
	NeuroScriptParserRULE_simple_statement     = 8
	NeuroScriptParserRULE_block_statement      = 9
	NeuroScriptParserRULE_set_statement        = 10
	NeuroScriptParserRULE_call_statement       = 11
	NeuroScriptParserRULE_return_statement     = 12
	NeuroScriptParserRULE_emit_statement       = 13
	NeuroScriptParserRULE_if_statement         = 14
	NeuroScriptParserRULE_while_statement      = 15
	NeuroScriptParserRULE_for_each_statement   = 16
	NeuroScriptParserRULE_call_target          = 17
	NeuroScriptParserRULE_condition            = 18
	NeuroScriptParserRULE_expression           = 19
	NeuroScriptParserRULE_term                 = 20
	NeuroScriptParserRULE_placeholder          = 21
	NeuroScriptParserRULE_literal              = 22
	NeuroScriptParserRULE_list_literal         = 23
	NeuroScriptParserRULE_map_literal          = 24
	NeuroScriptParserRULE_expression_list_opt  = 25
	NeuroScriptParserRULE_expression_list      = 26
	NeuroScriptParserRULE_map_entry_list_opt   = 27
	NeuroScriptParserRULE_map_entry_list       = 28
	NeuroScriptParserRULE_map_entry            = 29
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
		p.SetState(60)
		p.Optional_newlines()
	}
	p.SetState(64)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserKW_STARTPROC {
		{
			p.SetState(61)
			p.Procedure_definition()
		}

		p.SetState(66)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(67)
		p.Optional_newlines()
	}
	{
		p.SetState(68)
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
	p.SetState(73)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(70)
				p.Match(NeuroScriptParserNEWLINE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(75)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext())
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

// IProcedure_definitionContext is an interface to support dynamic dispatch.
type IProcedure_definitionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KW_STARTPROC() antlr.TerminalNode
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

func (s *Procedure_definitionContext) KW_STARTPROC() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_STARTPROC, 0)
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
	p.EnterRule(localctx, 4, NeuroScriptParserRULE_procedure_definition)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(76)
		p.Match(NeuroScriptParserKW_STARTPROC)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(77)
		p.Match(NeuroScriptParserKW_PROCEDURE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(78)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(79)
		p.Match(NeuroScriptParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(80)
		p.Param_list_opt()
	}
	{
		p.SetState(81)
		p.Match(NeuroScriptParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(82)
		p.Match(NeuroScriptParserNEWLINE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(84)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NeuroScriptParserCOMMENT_BLOCK {
		{
			p.SetState(83)
			p.Match(NeuroScriptParserCOMMENT_BLOCK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(86)
		p.Statement_list()
	}
	{
		p.SetState(87)
		p.Match(NeuroScriptParserKW_END)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(89)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(88)
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
	p.EnterRule(localctx, 6, NeuroScriptParserRULE_param_list_opt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(92)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NeuroScriptParserIDENTIFIER {
		{
			p.SetState(91)
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
	p.EnterRule(localctx, 8, NeuroScriptParserRULE_param_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(94)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(99)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserCOMMA {
		{
			p.SetState(95)
			p.Match(NeuroScriptParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(96)
			p.Match(NeuroScriptParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(101)
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
	p.EnterRule(localctx, 10, NeuroScriptParserRULE_statement_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(105)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&140737490495360) != 0 {
		{
			p.SetState(102)
			p.Body_line()
		}

		p.SetState(107)
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
	p.EnterRule(localctx, 12, NeuroScriptParserRULE_body_line)
	p.SetState(112)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_SET, NeuroScriptParserKW_CALL, NeuroScriptParserKW_RETURN, NeuroScriptParserKW_IF, NeuroScriptParserKW_WHILE, NeuroScriptParserKW_FOR, NeuroScriptParserKW_EMIT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(108)
			p.Statement()
		}
		{
			p.SetState(109)
			p.Match(NeuroScriptParserNEWLINE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserNEWLINE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(111)
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
	p.EnterRule(localctx, 14, NeuroScriptParserRULE_statement)
	p.SetState(116)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_SET, NeuroScriptParserKW_CALL, NeuroScriptParserKW_RETURN, NeuroScriptParserKW_EMIT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(114)
			p.Simple_statement()
		}

	case NeuroScriptParserKW_IF, NeuroScriptParserKW_WHILE, NeuroScriptParserKW_FOR:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(115)
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
	p.EnterRule(localctx, 16, NeuroScriptParserRULE_simple_statement)
	p.SetState(122)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_SET:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(118)
			p.Set_statement()
		}

	case NeuroScriptParserKW_CALL:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(119)
			p.Call_statement()
		}

	case NeuroScriptParserKW_RETURN:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(120)
			p.Return_statement()
		}

	case NeuroScriptParserKW_EMIT:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(121)
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
	p.EnterRule(localctx, 18, NeuroScriptParserRULE_block_statement)
	p.SetState(127)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserKW_IF:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(124)
			p.If_statement()
		}

	case NeuroScriptParserKW_WHILE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(125)
			p.While_statement()
		}

	case NeuroScriptParserKW_FOR:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(126)
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
	p.EnterRule(localctx, 20, NeuroScriptParserRULE_set_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(129)
		p.Match(NeuroScriptParserKW_SET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(130)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(131)
		p.Match(NeuroScriptParserASSIGN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(132)
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
	p.EnterRule(localctx, 22, NeuroScriptParserRULE_call_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(134)
		p.Match(NeuroScriptParserKW_CALL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(135)
		p.Call_target()
	}
	{
		p.SetState(136)
		p.Match(NeuroScriptParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(137)
		p.Expression_list_opt()
	}
	{
		p.SetState(138)
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
	p.EnterRule(localctx, 24, NeuroScriptParserRULE_return_statement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(140)
		p.Match(NeuroScriptParserKW_RETURN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(142)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&35327340183552) != 0 {
		{
			p.SetState(141)
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
	p.EnterRule(localctx, 26, NeuroScriptParserRULE_emit_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(144)
		p.Match(NeuroScriptParserKW_EMIT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(145)
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

	// Getter signatures
	KW_IF() antlr.TerminalNode
	Condition() IConditionContext
	KW_THEN() antlr.TerminalNode
	NEWLINE() antlr.TerminalNode
	Statement_list() IStatement_listContext
	KW_ENDBLOCK() antlr.TerminalNode

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

func (s *If_statementContext) NEWLINE() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserNEWLINE, 0)
}

func (s *If_statementContext) Statement_list() IStatement_listContext {
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

func (s *If_statementContext) KW_ENDBLOCK() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_ENDBLOCK, 0)
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
	p.EnterRule(localctx, 28, NeuroScriptParserRULE_if_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(147)
		p.Match(NeuroScriptParserKW_IF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(148)
		p.Condition()
	}
	{
		p.SetState(149)
		p.Match(NeuroScriptParserKW_THEN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(150)
		p.Match(NeuroScriptParserNEWLINE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(151)
		p.Statement_list()
	}
	{
		p.SetState(152)
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
	p.EnterRule(localctx, 30, NeuroScriptParserRULE_while_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(154)
		p.Match(NeuroScriptParserKW_WHILE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(155)
		p.Condition()
	}
	{
		p.SetState(156)
		p.Match(NeuroScriptParserKW_DO)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(157)
		p.Match(NeuroScriptParserNEWLINE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(158)
		p.Statement_list()
	}
	{
		p.SetState(159)
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
	p.EnterRule(localctx, 32, NeuroScriptParserRULE_for_each_statement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(161)
		p.Match(NeuroScriptParserKW_FOR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(162)
		p.Match(NeuroScriptParserKW_EACH)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(163)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(164)
		p.Match(NeuroScriptParserKW_IN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(165)
		p.Expression()
	}
	{
		p.SetState(166)
		p.Match(NeuroScriptParserKW_DO)
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
		p.Statement_list()
	}
	{
		p.SetState(169)
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
	p.EnterRule(localctx, 34, NeuroScriptParserRULE_call_target)
	p.SetState(176)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(171)
			p.Match(NeuroScriptParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_TOOL:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(172)
			p.Match(NeuroScriptParserKW_TOOL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(173)
			p.Match(NeuroScriptParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(174)
			p.Match(NeuroScriptParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_LLM:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(175)
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
	p.EnterRule(localctx, 36, NeuroScriptParserRULE_condition)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(178)
		p.Expression()
	}
	p.SetState(181)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&34634616274944) != 0 {
		{
			p.SetState(179)
			_la = p.GetTokenStream().LA(1)

			if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&34634616274944) != 0) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(180)
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
	p.EnterRule(localctx, 38, NeuroScriptParserRULE_expression)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(183)
		p.Term()
	}
	p.SetState(188)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserPLUS {
		{
			p.SetState(184)
			p.Match(NeuroScriptParserPLUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(185)
			p.Term()
		}

		p.SetState(190)
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

// ITermContext is an interface to support dynamic dispatch.
type ITermContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Literal() ILiteralContext
	Placeholder() IPlaceholderContext
	IDENTIFIER() antlr.TerminalNode
	KW_LAST_CALL_RESULT() antlr.TerminalNode
	LPAREN() antlr.TerminalNode
	Expression() IExpressionContext
	RPAREN() antlr.TerminalNode

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

func (s *TermContext) Literal() ILiteralContext {
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

func (s *TermContext) Placeholder() IPlaceholderContext {
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

func (s *TermContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserIDENTIFIER, 0)
}

func (s *TermContext) KW_LAST_CALL_RESULT() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserKW_LAST_CALL_RESULT, 0)
}

func (s *TermContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserLPAREN, 0)
}

func (s *TermContext) Expression() IExpressionContext {
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

func (s *TermContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserRPAREN, 0)
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
	p.EnterRule(localctx, 40, NeuroScriptParserRULE_term)
	p.SetState(199)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserNUMBER_LIT, NeuroScriptParserSTRING_LIT, NeuroScriptParserLBRACK, NeuroScriptParserLBRACE:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(191)
			p.Literal()
		}

	case NeuroScriptParserPLACEHOLDER_START:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(192)
			p.Placeholder()
		}

	case NeuroScriptParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(193)
			p.Match(NeuroScriptParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserKW_LAST_CALL_RESULT:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(194)
			p.Match(NeuroScriptParserKW_LAST_CALL_RESULT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserLPAREN:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(195)
			p.Match(NeuroScriptParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(196)
			p.Expression()
		}
		{
			p.SetState(197)
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
	IDENTIFIER() antlr.TerminalNode
	PLACEHOLDER_END() antlr.TerminalNode

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

func (s *PlaceholderContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserIDENTIFIER, 0)
}

func (s *PlaceholderContext) PLACEHOLDER_END() antlr.TerminalNode {
	return s.GetToken(NeuroScriptParserPLACEHOLDER_END, 0)
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
	p.EnterRule(localctx, 42, NeuroScriptParserRULE_placeholder)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(201)
		p.Match(NeuroScriptParserPLACEHOLDER_START)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(202)
		p.Match(NeuroScriptParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(203)
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
	p.EnterRule(localctx, 44, NeuroScriptParserRULE_literal)
	p.SetState(209)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case NeuroScriptParserSTRING_LIT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(205)
			p.Match(NeuroScriptParserSTRING_LIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserNUMBER_LIT:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(206)
			p.Match(NeuroScriptParserNUMBER_LIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case NeuroScriptParserLBRACK:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(207)
			p.List_literal()
		}

	case NeuroScriptParserLBRACE:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(208)
			p.Map_literal()
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
	p.EnterRule(localctx, 46, NeuroScriptParserRULE_list_literal)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(211)
		p.Match(NeuroScriptParserLBRACK)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(212)
		p.Expression_list_opt()
	}
	{
		p.SetState(213)
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
	p.EnterRule(localctx, 48, NeuroScriptParserRULE_map_literal)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(215)
		p.Match(NeuroScriptParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(216)
		p.Map_entry_list_opt()
	}
	{
		p.SetState(217)
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
	p.EnterRule(localctx, 50, NeuroScriptParserRULE_expression_list_opt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(220)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&35327340183552) != 0 {
		{
			p.SetState(219)
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
	p.EnterRule(localctx, 52, NeuroScriptParserRULE_expression_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(222)
		p.Expression()
	}
	p.SetState(227)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserCOMMA {
		{
			p.SetState(223)
			p.Match(NeuroScriptParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(224)
			p.Expression()
		}

		p.SetState(229)
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
	p.EnterRule(localctx, 54, NeuroScriptParserRULE_map_entry_list_opt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(231)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == NeuroScriptParserSTRING_LIT {
		{
			p.SetState(230)
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
	p.EnterRule(localctx, 56, NeuroScriptParserRULE_map_entry_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(233)
		p.Map_entry()
	}
	p.SetState(238)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroScriptParserCOMMA {
		{
			p.SetState(234)
			p.Match(NeuroScriptParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(235)
			p.Map_entry()
		}

		p.SetState(240)
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
	p.EnterRule(localctx, 58, NeuroScriptParserRULE_map_entry)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(241)
		p.Match(NeuroScriptParserSTRING_LIT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(242)
		p.Match(NeuroScriptParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(243)
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
