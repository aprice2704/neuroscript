// Code generated from FencedBlockExtractor.g4 by ANTLR 4.13.2. DO NOT EDIT.

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

type FencedBlockExtractorLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var FencedBlockExtractorLexerLexerStaticData struct {
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

func fencedblockextractorlexerLexerInit() {
	staticData := &FencedBlockExtractorLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
	staticData.LiteralNames = []string{
		"", "'```'",
	}
	staticData.SymbolicNames = []string{
		"", "FENCE_MARKER", "LANG_ID", "LINE_COMMENT", "NEWLINE", "WS", "OTHER_TEXT",
	}
	staticData.RuleNames = []string{
		"FENCE_MARKER", "LANG_ID", "LINE_COMMENT", "NEWLINE", "WS", "OTHER_TEXT",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 6, 48, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 1, 0, 1, 0, 1, 0, 1, 0, 1, 1, 4, 1, 19, 8, 1, 11,
		1, 12, 1, 20, 1, 2, 1, 2, 1, 2, 3, 2, 26, 8, 2, 1, 2, 5, 2, 29, 8, 2, 10,
		2, 12, 2, 32, 9, 2, 1, 2, 3, 2, 35, 8, 2, 1, 3, 4, 3, 38, 8, 3, 11, 3,
		12, 3, 39, 1, 4, 4, 4, 43, 8, 4, 11, 4, 12, 4, 44, 1, 5, 1, 5, 1, 30, 0,
		6, 1, 1, 3, 2, 5, 3, 7, 4, 9, 5, 11, 6, 1, 0, 3, 7, 0, 35, 35, 43, 43,
		45, 46, 48, 57, 65, 90, 95, 95, 97, 122, 2, 0, 10, 10, 13, 13, 2, 0, 9,
		9, 32, 32, 53, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0,
		7, 1, 0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0, 0, 1, 13, 1, 0, 0, 0,
		3, 18, 1, 0, 0, 0, 5, 25, 1, 0, 0, 0, 7, 37, 1, 0, 0, 0, 9, 42, 1, 0, 0,
		0, 11, 46, 1, 0, 0, 0, 13, 14, 5, 96, 0, 0, 14, 15, 5, 96, 0, 0, 15, 16,
		5, 96, 0, 0, 16, 2, 1, 0, 0, 0, 17, 19, 7, 0, 0, 0, 18, 17, 1, 0, 0, 0,
		19, 20, 1, 0, 0, 0, 20, 18, 1, 0, 0, 0, 20, 21, 1, 0, 0, 0, 21, 4, 1, 0,
		0, 0, 22, 26, 5, 35, 0, 0, 23, 24, 5, 45, 0, 0, 24, 26, 5, 45, 0, 0, 25,
		22, 1, 0, 0, 0, 25, 23, 1, 0, 0, 0, 26, 30, 1, 0, 0, 0, 27, 29, 9, 0, 0,
		0, 28, 27, 1, 0, 0, 0, 29, 32, 1, 0, 0, 0, 30, 31, 1, 0, 0, 0, 30, 28,
		1, 0, 0, 0, 31, 34, 1, 0, 0, 0, 32, 30, 1, 0, 0, 0, 33, 35, 3, 7, 3, 0,
		34, 33, 1, 0, 0, 0, 34, 35, 1, 0, 0, 0, 35, 6, 1, 0, 0, 0, 36, 38, 7, 1,
		0, 0, 37, 36, 1, 0, 0, 0, 38, 39, 1, 0, 0, 0, 39, 37, 1, 0, 0, 0, 39, 40,
		1, 0, 0, 0, 40, 8, 1, 0, 0, 0, 41, 43, 7, 2, 0, 0, 42, 41, 1, 0, 0, 0,
		43, 44, 1, 0, 0, 0, 44, 42, 1, 0, 0, 0, 44, 45, 1, 0, 0, 0, 45, 10, 1,
		0, 0, 0, 46, 47, 9, 0, 0, 0, 47, 12, 1, 0, 0, 0, 7, 0, 20, 25, 30, 34,
		39, 44, 0,
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

// FencedBlockExtractorLexerInit initializes any static state used to implement FencedBlockExtractorLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewFencedBlockExtractorLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func FencedBlockExtractorLexerInit() {
	staticData := &FencedBlockExtractorLexerLexerStaticData
	staticData.once.Do(fencedblockextractorlexerLexerInit)
}

// NewFencedBlockExtractorLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewFencedBlockExtractorLexer(input antlr.CharStream) *FencedBlockExtractorLexer {
	FencedBlockExtractorLexerInit()
	l := new(FencedBlockExtractorLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &FencedBlockExtractorLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "FencedBlockExtractor.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// FencedBlockExtractorLexer tokens.
const (
	FencedBlockExtractorLexerFENCE_MARKER = 1
	FencedBlockExtractorLexerLANG_ID      = 2
	FencedBlockExtractorLexerLINE_COMMENT = 3
	FencedBlockExtractorLexerNEWLINE      = 4
	FencedBlockExtractorLexerWS           = 5
	FencedBlockExtractorLexerOTHER_TEXT   = 6
)
