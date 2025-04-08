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
		"", "FENCE_MARKER", "LANG_ID", "METADATA_LINE", "LINE_COMMENT", "WS",
		"NEWLINE", "OTHER_TEXT",
	}
	staticData.RuleNames = []string{
		"FENCE_MARKER", "LANG_ID", "METADATA_LINE", "LINE_COMMENT", "WS", "NEWLINE",
		"OTHER_TEXT",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 7, 68, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 1, 0, 1, 0, 1, 0, 1, 0, 1, 1, 4, 1, 21,
		8, 1, 11, 1, 12, 1, 22, 1, 2, 1, 2, 1, 2, 1, 2, 4, 2, 29, 8, 2, 11, 2,
		12, 2, 30, 1, 2, 5, 2, 34, 8, 2, 10, 2, 12, 2, 37, 9, 2, 1, 3, 1, 3, 1,
		3, 3, 3, 42, 8, 3, 1, 3, 5, 3, 45, 8, 3, 10, 3, 12, 3, 48, 9, 3, 1, 4,
		4, 4, 51, 8, 4, 11, 4, 12, 4, 52, 1, 5, 3, 5, 56, 8, 5, 1, 5, 1, 5, 4,
		5, 60, 8, 5, 11, 5, 12, 5, 61, 1, 6, 4, 6, 65, 8, 6, 11, 6, 12, 6, 66,
		0, 0, 7, 1, 1, 3, 2, 5, 3, 7, 4, 9, 5, 11, 6, 13, 7, 1, 0, 4, 5, 0, 45,
		46, 48, 57, 65, 90, 95, 95, 97, 122, 2, 0, 9, 9, 32, 32, 2, 0, 10, 10,
		13, 13, 6, 0, 9, 10, 13, 13, 32, 32, 35, 35, 58, 58, 96, 96, 77, 0, 1,
		1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 7, 1, 0, 0, 0, 0, 9,
		1, 0, 0, 0, 0, 11, 1, 0, 0, 0, 0, 13, 1, 0, 0, 0, 1, 15, 1, 0, 0, 0, 3,
		20, 1, 0, 0, 0, 5, 24, 1, 0, 0, 0, 7, 41, 1, 0, 0, 0, 9, 50, 1, 0, 0, 0,
		11, 59, 1, 0, 0, 0, 13, 64, 1, 0, 0, 0, 15, 16, 5, 96, 0, 0, 16, 17, 5,
		96, 0, 0, 17, 18, 5, 96, 0, 0, 18, 2, 1, 0, 0, 0, 19, 21, 7, 0, 0, 0, 20,
		19, 1, 0, 0, 0, 21, 22, 1, 0, 0, 0, 22, 20, 1, 0, 0, 0, 22, 23, 1, 0, 0,
		0, 23, 4, 1, 0, 0, 0, 24, 25, 5, 58, 0, 0, 25, 26, 5, 58, 0, 0, 26, 28,
		1, 0, 0, 0, 27, 29, 7, 1, 0, 0, 28, 27, 1, 0, 0, 0, 29, 30, 1, 0, 0, 0,
		30, 28, 1, 0, 0, 0, 30, 31, 1, 0, 0, 0, 31, 35, 1, 0, 0, 0, 32, 34, 8,
		2, 0, 0, 33, 32, 1, 0, 0, 0, 34, 37, 1, 0, 0, 0, 35, 33, 1, 0, 0, 0, 35,
		36, 1, 0, 0, 0, 36, 6, 1, 0, 0, 0, 37, 35, 1, 0, 0, 0, 38, 42, 5, 35, 0,
		0, 39, 40, 5, 45, 0, 0, 40, 42, 5, 45, 0, 0, 41, 38, 1, 0, 0, 0, 41, 39,
		1, 0, 0, 0, 42, 46, 1, 0, 0, 0, 43, 45, 8, 2, 0, 0, 44, 43, 1, 0, 0, 0,
		45, 48, 1, 0, 0, 0, 46, 44, 1, 0, 0, 0, 46, 47, 1, 0, 0, 0, 47, 8, 1, 0,
		0, 0, 48, 46, 1, 0, 0, 0, 49, 51, 7, 1, 0, 0, 50, 49, 1, 0, 0, 0, 51, 52,
		1, 0, 0, 0, 52, 50, 1, 0, 0, 0, 52, 53, 1, 0, 0, 0, 53, 10, 1, 0, 0, 0,
		54, 56, 5, 13, 0, 0, 55, 54, 1, 0, 0, 0, 55, 56, 1, 0, 0, 0, 56, 57, 1,
		0, 0, 0, 57, 60, 5, 10, 0, 0, 58, 60, 5, 13, 0, 0, 59, 55, 1, 0, 0, 0,
		59, 58, 1, 0, 0, 0, 60, 61, 1, 0, 0, 0, 61, 59, 1, 0, 0, 0, 61, 62, 1,
		0, 0, 0, 62, 12, 1, 0, 0, 0, 63, 65, 8, 3, 0, 0, 64, 63, 1, 0, 0, 0, 65,
		66, 1, 0, 0, 0, 66, 64, 1, 0, 0, 0, 66, 67, 1, 0, 0, 0, 67, 14, 1, 0, 0,
		0, 11, 0, 22, 30, 35, 41, 46, 52, 55, 59, 61, 66, 0,
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
	FencedBlockExtractorLexerFENCE_MARKER  = 1
	FencedBlockExtractorLexerLANG_ID       = 2
	FencedBlockExtractorLexerMETADATA_LINE = 3
	FencedBlockExtractorLexerLINE_COMMENT  = 4
	FencedBlockExtractorLexerWS            = 5
	FencedBlockExtractorLexerNEWLINE       = 6
	FencedBlockExtractorLexerOTHER_TEXT    = 7
)
