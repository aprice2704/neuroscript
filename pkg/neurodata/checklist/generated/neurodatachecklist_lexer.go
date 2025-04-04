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
		"", "", "", "'-'", "'['", "']'",
	}
	staticData.SymbolicNames = []string{
		"", "METADATA_LINE", "COMMENT_LINE", "HYPHEN", "LBRACK", "RBRACK", "MARK",
		"TEXT", "NEWLINE", "WS",
	}
	staticData.RuleNames = []string{
		"METADATA_LINE", "COMMENT_LINE", "HYPHEN", "LBRACK", "RBRACK", "MARK",
		"TEXT", "NEWLINE", "WS",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 9, 66, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 1, 0, 1, 0, 5,
		0, 22, 8, 0, 10, 0, 12, 0, 25, 9, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1,
		5, 1, 33, 8, 1, 10, 1, 12, 1, 36, 9, 1, 1, 1, 1, 1, 1, 2, 1, 2, 1, 3, 1,
		3, 1, 4, 1, 4, 1, 5, 1, 5, 1, 6, 4, 6, 49, 8, 6, 11, 6, 12, 6, 50, 1, 7,
		3, 7, 54, 8, 7, 1, 7, 1, 7, 4, 7, 58, 8, 7, 11, 7, 12, 7, 59, 1, 8, 4,
		8, 63, 8, 8, 11, 8, 12, 8, 64, 0, 0, 9, 1, 1, 3, 2, 5, 3, 7, 4, 9, 5, 11,
		6, 13, 7, 15, 8, 17, 9, 1, 0, 3, 2, 0, 10, 10, 13, 13, 3, 0, 32, 32, 88,
		88, 120, 120, 2, 0, 9, 9, 32, 32, 72, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0,
		0, 0, 5, 1, 0, 0, 0, 0, 7, 1, 0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0,
		0, 0, 13, 1, 0, 0, 0, 0, 15, 1, 0, 0, 0, 0, 17, 1, 0, 0, 0, 1, 19, 1, 0,
		0, 0, 3, 28, 1, 0, 0, 0, 5, 39, 1, 0, 0, 0, 7, 41, 1, 0, 0, 0, 9, 43, 1,
		0, 0, 0, 11, 45, 1, 0, 0, 0, 13, 48, 1, 0, 0, 0, 15, 57, 1, 0, 0, 0, 17,
		62, 1, 0, 0, 0, 19, 23, 5, 35, 0, 0, 20, 22, 8, 0, 0, 0, 21, 20, 1, 0,
		0, 0, 22, 25, 1, 0, 0, 0, 23, 21, 1, 0, 0, 0, 23, 24, 1, 0, 0, 0, 24, 26,
		1, 0, 0, 0, 25, 23, 1, 0, 0, 0, 26, 27, 6, 0, 0, 0, 27, 2, 1, 0, 0, 0,
		28, 29, 5, 45, 0, 0, 29, 30, 5, 45, 0, 0, 30, 34, 1, 0, 0, 0, 31, 33, 8,
		0, 0, 0, 32, 31, 1, 0, 0, 0, 33, 36, 1, 0, 0, 0, 34, 32, 1, 0, 0, 0, 34,
		35, 1, 0, 0, 0, 35, 37, 1, 0, 0, 0, 36, 34, 1, 0, 0, 0, 37, 38, 6, 1, 0,
		0, 38, 4, 1, 0, 0, 0, 39, 40, 5, 45, 0, 0, 40, 6, 1, 0, 0, 0, 41, 42, 5,
		91, 0, 0, 42, 8, 1, 0, 0, 0, 43, 44, 5, 93, 0, 0, 44, 10, 1, 0, 0, 0, 45,
		46, 7, 1, 0, 0, 46, 12, 1, 0, 0, 0, 47, 49, 8, 0, 0, 0, 48, 47, 1, 0, 0,
		0, 49, 50, 1, 0, 0, 0, 50, 48, 1, 0, 0, 0, 50, 51, 1, 0, 0, 0, 51, 14,
		1, 0, 0, 0, 52, 54, 5, 13, 0, 0, 53, 52, 1, 0, 0, 0, 53, 54, 1, 0, 0, 0,
		54, 55, 1, 0, 0, 0, 55, 58, 5, 10, 0, 0, 56, 58, 5, 13, 0, 0, 57, 53, 1,
		0, 0, 0, 57, 56, 1, 0, 0, 0, 58, 59, 1, 0, 0, 0, 59, 57, 1, 0, 0, 0, 59,
		60, 1, 0, 0, 0, 60, 16, 1, 0, 0, 0, 61, 63, 7, 2, 0, 0, 62, 61, 1, 0, 0,
		0, 63, 64, 1, 0, 0, 0, 64, 62, 1, 0, 0, 0, 64, 65, 1, 0, 0, 0, 65, 18,
		1, 0, 0, 0, 8, 0, 23, 34, 50, 53, 57, 59, 64, 1, 6, 0, 0,
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
	NeuroDataChecklistLexerMETADATA_LINE = 1
	NeuroDataChecklistLexerCOMMENT_LINE  = 2
	NeuroDataChecklistLexerHYPHEN        = 3
	NeuroDataChecklistLexerLBRACK        = 4
	NeuroDataChecklistLexerRBRACK        = 5
	NeuroDataChecklistLexerMARK          = 6
	NeuroDataChecklistLexerTEXT          = 7
	NeuroDataChecklistLexerNEWLINE       = 8
	NeuroDataChecklistLexerWS            = 9
)
