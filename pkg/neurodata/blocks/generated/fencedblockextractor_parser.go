// Code generated from FencedBlockExtractor.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated // FencedBlockExtractor
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

type FencedBlockExtractorParser struct {
	*antlr.BaseParser
}

var FencedBlockExtractorParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func fencedblockextractorParserInit() {
	staticData := &FencedBlockExtractorParserStaticData
	staticData.LiteralNames = []string{
		"", "'```'",
	}
	staticData.SymbolicNames = []string{
		"", "FENCE_MARKER", "LANG_ID", "LINE_COMMENT", "NEWLINE", "WS", "OTHER_TEXT",
	}
	staticData.RuleNames = []string{
		"document", "token",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 6, 15, 2, 0, 7, 0, 2, 1, 7, 1, 1, 0, 5, 0, 6, 8, 0, 10, 0, 12, 0,
		9, 9, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 0, 0, 2, 0, 2, 0, 1, 1, 0, 1, 6,
		13, 0, 7, 1, 0, 0, 0, 2, 12, 1, 0, 0, 0, 4, 6, 3, 2, 1, 0, 5, 4, 1, 0,
		0, 0, 6, 9, 1, 0, 0, 0, 7, 5, 1, 0, 0, 0, 7, 8, 1, 0, 0, 0, 8, 10, 1, 0,
		0, 0, 9, 7, 1, 0, 0, 0, 10, 11, 5, 0, 0, 1, 11, 1, 1, 0, 0, 0, 12, 13,
		7, 0, 0, 0, 13, 3, 1, 0, 0, 0, 1, 7,
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

// FencedBlockExtractorParserInit initializes any static state used to implement FencedBlockExtractorParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewFencedBlockExtractorParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func FencedBlockExtractorParserInit() {
	staticData := &FencedBlockExtractorParserStaticData
	staticData.once.Do(fencedblockextractorParserInit)
}

// NewFencedBlockExtractorParser produces a new parser instance for the optional input antlr.TokenStream.
func NewFencedBlockExtractorParser(input antlr.TokenStream) *FencedBlockExtractorParser {
	FencedBlockExtractorParserInit()
	this := new(FencedBlockExtractorParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &FencedBlockExtractorParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "FencedBlockExtractor.g4"

	return this
}

// FencedBlockExtractorParser tokens.
const (
	FencedBlockExtractorParserEOF          = antlr.TokenEOF
	FencedBlockExtractorParserFENCE_MARKER = 1
	FencedBlockExtractorParserLANG_ID      = 2
	FencedBlockExtractorParserLINE_COMMENT = 3
	FencedBlockExtractorParserNEWLINE      = 4
	FencedBlockExtractorParserWS           = 5
	FencedBlockExtractorParserOTHER_TEXT   = 6
)

// FencedBlockExtractorParser rules.
const (
	FencedBlockExtractorParserRULE_document = 0
	FencedBlockExtractorParserRULE_token    = 1
)

// IDocumentContext is an interface to support dynamic dispatch.
type IDocumentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	AllToken() []ITokenContext
	Token(i int) ITokenContext

	// IsDocumentContext differentiates from other interfaces.
	IsDocumentContext()
}

type DocumentContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDocumentContext() *DocumentContext {
	var p = new(DocumentContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = FencedBlockExtractorParserRULE_document
	return p
}

func InitEmptyDocumentContext(p *DocumentContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = FencedBlockExtractorParserRULE_document
}

func (*DocumentContext) IsDocumentContext() {}

func NewDocumentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DocumentContext {
	var p = new(DocumentContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = FencedBlockExtractorParserRULE_document

	return p
}

func (s *DocumentContext) GetParser() antlr.Parser { return s.parser }

func (s *DocumentContext) EOF() antlr.TerminalNode {
	return s.GetToken(FencedBlockExtractorParserEOF, 0)
}

func (s *DocumentContext) AllToken() []ITokenContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITokenContext); ok {
			len++
		}
	}

	tst := make([]ITokenContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITokenContext); ok {
			tst[i] = t.(ITokenContext)
			i++
		}
	}

	return tst
}

func (s *DocumentContext) Token(i int) ITokenContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITokenContext); ok {
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

	return t.(ITokenContext)
}

func (s *DocumentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DocumentContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DocumentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FencedBlockExtractorListener); ok {
		listenerT.EnterDocument(s)
	}
}

func (s *DocumentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FencedBlockExtractorListener); ok {
		listenerT.ExitDocument(s)
	}
}

func (s *DocumentContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case FencedBlockExtractorVisitor:
		return t.VisitDocument(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *FencedBlockExtractorParser) Document() (localctx IDocumentContext) {
	localctx = NewDocumentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, FencedBlockExtractorParserRULE_document)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(7)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&126) != 0 {
		{
			p.SetState(4)
			p.Token()
		}

		p.SetState(9)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(10)
		p.Match(FencedBlockExtractorParserEOF)
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

// ITokenContext is an interface to support dynamic dispatch.
type ITokenContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	FENCE_MARKER() antlr.TerminalNode
	LANG_ID() antlr.TerminalNode
	NEWLINE() antlr.TerminalNode
	OTHER_TEXT() antlr.TerminalNode
	LINE_COMMENT() antlr.TerminalNode
	WS() antlr.TerminalNode

	// IsTokenContext differentiates from other interfaces.
	IsTokenContext()
}

type TokenContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTokenContext() *TokenContext {
	var p = new(TokenContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = FencedBlockExtractorParserRULE_token
	return p
}

func InitEmptyTokenContext(p *TokenContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = FencedBlockExtractorParserRULE_token
}

func (*TokenContext) IsTokenContext() {}

func NewTokenContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TokenContext {
	var p = new(TokenContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = FencedBlockExtractorParserRULE_token

	return p
}

func (s *TokenContext) GetParser() antlr.Parser { return s.parser }

func (s *TokenContext) FENCE_MARKER() antlr.TerminalNode {
	return s.GetToken(FencedBlockExtractorParserFENCE_MARKER, 0)
}

func (s *TokenContext) LANG_ID() antlr.TerminalNode {
	return s.GetToken(FencedBlockExtractorParserLANG_ID, 0)
}

func (s *TokenContext) NEWLINE() antlr.TerminalNode {
	return s.GetToken(FencedBlockExtractorParserNEWLINE, 0)
}

func (s *TokenContext) OTHER_TEXT() antlr.TerminalNode {
	return s.GetToken(FencedBlockExtractorParserOTHER_TEXT, 0)
}

func (s *TokenContext) LINE_COMMENT() antlr.TerminalNode {
	return s.GetToken(FencedBlockExtractorParserLINE_COMMENT, 0)
}

func (s *TokenContext) WS() antlr.TerminalNode {
	return s.GetToken(FencedBlockExtractorParserWS, 0)
}

func (s *TokenContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TokenContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TokenContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FencedBlockExtractorListener); ok {
		listenerT.EnterToken(s)
	}
}

func (s *TokenContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FencedBlockExtractorListener); ok {
		listenerT.ExitToken(s)
	}
}

func (s *TokenContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case FencedBlockExtractorVisitor:
		return t.VisitToken(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *FencedBlockExtractorParser) Token() (localctx ITokenContext) {
	localctx = NewTokenContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, FencedBlockExtractorParserRULE_token)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(12)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&126) != 0) {
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
