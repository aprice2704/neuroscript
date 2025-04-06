// Code generated from NeuroDataChecklist.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"sync"
	"unicode"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type NeuroDataChecklistLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var NeuroDataChecklistLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func neurodatachecklistlexerLexerInit() {
	staticData := &NeuroDataChecklistLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
	staticData.LiteralNames = []string{
		"", "'['", "']'", "", "'-'",
	}
	staticData.SymbolicNames = []string{
		"", "LBRACK", "RBRACK", "MARK", "HYPHEN", "TEXT", "METADATA_LINE", "COMMENT_LINE",
		"NEWLINE", "WS",
	}
	staticData.RuleNames = []string{
		"LBRACK", "RBRACK", "MARK", "HYPHEN", "TEXT", "METADATA_LINE", "COMMENT_LINE",
		"NEWLINE", "WS",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 9, 101, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 1, 0, 1, 0, 1,
		1, 1, 1, 1, 2, 1, 2, 1, 3, 1, 3, 1, 4, 4, 4, 29, 8, 4, 11, 4, 12, 4, 30,
		1, 5, 1, 5, 1, 5, 3, 5, 36, 8, 5, 1, 5, 3, 5, 39, 8, 5, 1, 5, 1, 5, 1,
		5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1,
		5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 3, 5, 64, 8, 5, 1, 5,
		1, 5, 5, 5, 68, 8, 5, 10, 5, 12, 5, 71, 9, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1,
		6, 3, 6, 78, 8, 6, 1, 6, 5, 6, 81, 8, 6, 10, 6, 12, 6, 84, 9, 6, 1, 6,
		1, 6, 1, 7, 3, 7, 89, 8, 7, 1, 7, 1, 7, 3, 7, 93, 8, 7, 1, 8, 4, 8, 96,
		8, 8, 11, 8, 12, 8, 97, 1, 8, 1, 8, 1, 30, 0, 9, 1, 1, 3, 2, 5, 3, 7, 4,
		9, 5, 11, 6, 13, 7, 15, 8, 17, 9, 1, 0, 3, 3, 0, 32, 32, 88, 88, 120, 120,
		2, 0, 10, 10, 13, 13, 2, 0, 9, 9, 32, 32, 111, 0, 1, 1, 0, 0, 0, 0, 3,
		1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 7, 1, 0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11,
		1, 0, 0, 0, 0, 13, 1, 0, 0, 0, 0, 15, 1, 0, 0, 0, 0, 17, 1, 0, 0, 0, 1,
		19, 1, 0, 0, 0, 3, 21, 1, 0, 0, 0, 5, 23, 1, 0, 0, 0, 7, 25, 1, 0, 0, 0,
		9, 28, 1, 0, 0, 0, 11, 35, 1, 0, 0, 0, 13, 77, 1, 0, 0, 0, 15, 92, 1, 0,
		0, 0, 17, 95, 1, 0, 0, 0, 19, 20, 5, 91, 0, 0, 20, 2, 1, 0, 0, 0, 21, 22,
		5, 93, 0, 0, 22, 4, 1, 0, 0, 0, 23, 24, 7, 0, 0, 0, 24, 6, 1, 0, 0, 0,
		25, 26, 5, 45, 0, 0, 26, 8, 1, 0, 0, 0, 27, 29, 8, 1, 0, 0, 28, 27, 1,
		0, 0, 0, 29, 30, 1, 0, 0, 0, 30, 31, 1, 0, 0, 0, 30, 28, 1, 0, 0, 0, 31,
		10, 1, 0, 0, 0, 32, 36, 5, 35, 0, 0, 33, 34, 5, 45, 0, 0, 34, 36, 5, 45,
		0, 0, 35, 32, 1, 0, 0, 0, 35, 33, 1, 0, 0, 0, 36, 38, 1, 0, 0, 0, 37, 39,
		3, 17, 8, 0, 38, 37, 1, 0, 0, 0, 38, 39, 1, 0, 0, 0, 39, 63, 1, 0, 0, 0,
		40, 41, 5, 105, 0, 0, 41, 64, 5, 100, 0, 0, 42, 43, 5, 118, 0, 0, 43, 44,
		5, 101, 0, 0, 44, 45, 5, 114, 0, 0, 45, 46, 5, 115, 0, 0, 46, 47, 5, 105,
		0, 0, 47, 48, 5, 111, 0, 0, 48, 64, 5, 110, 0, 0, 49, 50, 5, 114, 0, 0,
		50, 51, 5, 101, 0, 0, 51, 52, 5, 110, 0, 0, 52, 53, 5, 100, 0, 0, 53, 54,
		5, 101, 0, 0, 54, 55, 5, 114, 0, 0, 55, 56, 5, 105, 0, 0, 56, 57, 5, 110,
		0, 0, 57, 58, 5, 103, 0, 0, 58, 59, 5, 95, 0, 0, 59, 60, 5, 104, 0, 0,
		60, 61, 5, 105, 0, 0, 61, 62, 5, 110, 0, 0, 62, 64, 5, 116, 0, 0, 63, 40,
		1, 0, 0, 0, 63, 42, 1, 0, 0, 0, 63, 49, 1, 0, 0, 0, 64, 65, 1, 0, 0, 0,
		65, 69, 5, 58, 0, 0, 66, 68, 8, 1, 0, 0, 67, 66, 1, 0, 0, 0, 68, 71, 1,
		0, 0, 0, 69, 67, 1, 0, 0, 0, 69, 70, 1, 0, 0, 0, 70, 72, 1, 0, 0, 0, 71,
		69, 1, 0, 0, 0, 72, 73, 6, 5, 0, 0, 73, 12, 1, 0, 0, 0, 74, 78, 5, 35,
		0, 0, 75, 76, 5, 45, 0, 0, 76, 78, 5, 45, 0, 0, 77, 74, 1, 0, 0, 0, 77,
		75, 1, 0, 0, 0, 78, 82, 1, 0, 0, 0, 79, 81, 8, 1, 0, 0, 80, 79, 1, 0, 0,
		0, 81, 84, 1, 0, 0, 0, 82, 80, 1, 0, 0, 0, 82, 83, 1, 0, 0, 0, 83, 85,
		1, 0, 0, 0, 84, 82, 1, 0, 0, 0, 85, 86, 6, 6, 0, 0, 86, 14, 1, 0, 0, 0,
		87, 89, 5, 13, 0, 0, 88, 87, 1, 0, 0, 0, 88, 89, 1, 0, 0, 0, 89, 90, 1,
		0, 0, 0, 90, 93, 5, 10, 0, 0, 91, 93, 5, 13, 0, 0, 92, 88, 1, 0, 0, 0,
		92, 91, 1, 0, 0, 0, 93, 16, 1, 0, 0, 0, 94, 96, 7, 2, 0, 0, 95, 94, 1,
		0, 0, 0, 96, 97, 1, 0, 0, 0, 97, 95, 1, 0, 0, 0, 97, 98, 1, 0, 0, 0, 98,
		99, 1, 0, 0, 0, 99, 100, 6, 8, 0, 0, 100, 18, 1, 0, 0, 0, 11, 0, 30, 35,
		38, 63, 69, 77, 82, 88, 92, 97, 1, 6, 0, 0,
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

// NeuroDataChecklistLexerInit initializes any static state used to implement NeuroDataChecklistLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewNeuroDataChecklistLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func NeuroDataChecklistLexerInit() {
	staticData := &NeuroDataChecklistLexerLexerStaticData
	staticData.once.Do(neurodatachecklistlexerLexerInit)
}

// NewNeuroDataChecklistLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewNeuroDataChecklistLexer(input antlr.CharStream) *NeuroDataChecklistLexer {
	NeuroDataChecklistLexerInit()
	l := new(NeuroDataChecklistLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &NeuroDataChecklistLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "NeuroDataChecklist.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// NeuroDataChecklistLexer tokens.
const (
	NeuroDataChecklistLexerLBRACK        = 1
	NeuroDataChecklistLexerRBRACK        = 2
	NeuroDataChecklistLexerMARK          = 3
	NeuroDataChecklistLexerHYPHEN        = 4
	NeuroDataChecklistLexerTEXT          = 5
	NeuroDataChecklistLexerMETADATA_LINE = 6
	NeuroDataChecklistLexerCOMMENT_LINE  = 7
	NeuroDataChecklistLexerNEWLINE       = 8
	NeuroDataChecklistLexerWS            = 9
)
