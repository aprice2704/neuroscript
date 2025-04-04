// Code generated from NeuroDataChecklist.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated // NeuroDataChecklist
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

type NeuroDataChecklistParser struct {
	*antlr.BaseParser
}

var NeuroDataChecklistParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func neurodatachecklistParserInit() {
	staticData := &NeuroDataChecklistParserStaticData
	staticData.LiteralNames = []string{
		"", "", "", "'-'", "'['", "']'",
	}
	staticData.SymbolicNames = []string{
		"", "METADATA_LINE", "COMMENT_LINE", "HYPHEN", "LBRACK", "RBRACK", "MARK",
		"TEXT", "NEWLINE", "WS",
	}
	staticData.RuleNames = []string{
		"checklistFile", "itemLine",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 9, 59, 2, 0, 7, 0, 2, 1, 7, 1, 1, 0, 5, 0, 6, 8, 0, 10, 0, 12, 0,
		9, 9, 0, 1, 0, 1, 0, 5, 0, 13, 8, 0, 10, 0, 12, 0, 16, 9, 0, 5, 0, 18,
		8, 0, 10, 0, 12, 0, 21, 9, 0, 1, 0, 1, 0, 1, 1, 5, 1, 26, 8, 1, 10, 1,
		12, 1, 29, 9, 1, 1, 1, 1, 1, 4, 1, 33, 8, 1, 11, 1, 12, 1, 34, 1, 1, 1,
		1, 5, 1, 39, 8, 1, 10, 1, 12, 1, 42, 9, 1, 1, 1, 1, 1, 5, 1, 46, 8, 1,
		10, 1, 12, 1, 49, 9, 1, 1, 1, 1, 1, 4, 1, 53, 8, 1, 11, 1, 12, 1, 54, 1,
		1, 1, 1, 1, 1, 0, 0, 2, 0, 2, 0, 0, 64, 0, 7, 1, 0, 0, 0, 2, 27, 1, 0,
		0, 0, 4, 6, 5, 8, 0, 0, 5, 4, 1, 0, 0, 0, 6, 9, 1, 0, 0, 0, 7, 5, 1, 0,
		0, 0, 7, 8, 1, 0, 0, 0, 8, 19, 1, 0, 0, 0, 9, 7, 1, 0, 0, 0, 10, 14, 3,
		2, 1, 0, 11, 13, 5, 8, 0, 0, 12, 11, 1, 0, 0, 0, 13, 16, 1, 0, 0, 0, 14,
		12, 1, 0, 0, 0, 14, 15, 1, 0, 0, 0, 15, 18, 1, 0, 0, 0, 16, 14, 1, 0, 0,
		0, 17, 10, 1, 0, 0, 0, 18, 21, 1, 0, 0, 0, 19, 17, 1, 0, 0, 0, 19, 20,
		1, 0, 0, 0, 20, 22, 1, 0, 0, 0, 21, 19, 1, 0, 0, 0, 22, 23, 5, 0, 0, 1,
		23, 1, 1, 0, 0, 0, 24, 26, 5, 9, 0, 0, 25, 24, 1, 0, 0, 0, 26, 29, 1, 0,
		0, 0, 27, 25, 1, 0, 0, 0, 27, 28, 1, 0, 0, 0, 28, 30, 1, 0, 0, 0, 29, 27,
		1, 0, 0, 0, 30, 32, 5, 3, 0, 0, 31, 33, 5, 9, 0, 0, 32, 31, 1, 0, 0, 0,
		33, 34, 1, 0, 0, 0, 34, 32, 1, 0, 0, 0, 34, 35, 1, 0, 0, 0, 35, 36, 1,
		0, 0, 0, 36, 40, 5, 4, 0, 0, 37, 39, 5, 9, 0, 0, 38, 37, 1, 0, 0, 0, 39,
		42, 1, 0, 0, 0, 40, 38, 1, 0, 0, 0, 40, 41, 1, 0, 0, 0, 41, 43, 1, 0, 0,
		0, 42, 40, 1, 0, 0, 0, 43, 47, 5, 6, 0, 0, 44, 46, 5, 9, 0, 0, 45, 44,
		1, 0, 0, 0, 46, 49, 1, 0, 0, 0, 47, 45, 1, 0, 0, 0, 47, 48, 1, 0, 0, 0,
		48, 50, 1, 0, 0, 0, 49, 47, 1, 0, 0, 0, 50, 52, 5, 5, 0, 0, 51, 53, 5,
		9, 0, 0, 52, 51, 1, 0, 0, 0, 53, 54, 1, 0, 0, 0, 54, 52, 1, 0, 0, 0, 54,
		55, 1, 0, 0, 0, 55, 56, 1, 0, 0, 0, 56, 57, 5, 7, 0, 0, 57, 3, 1, 0, 0,
		0, 8, 7, 14, 19, 27, 34, 40, 47, 54,
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

// NeuroDataChecklistParserInit initializes any static state used to implement NeuroDataChecklistParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewNeuroDataChecklistParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func NeuroDataChecklistParserInit() {
	staticData := &NeuroDataChecklistParserStaticData
	staticData.once.Do(neurodatachecklistParserInit)
}

// NewNeuroDataChecklistParser produces a new parser instance for the optional input antlr.TokenStream.
func NewNeuroDataChecklistParser(input antlr.TokenStream) *NeuroDataChecklistParser {
	NeuroDataChecklistParserInit()
	this := new(NeuroDataChecklistParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &NeuroDataChecklistParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "NeuroDataChecklist.g4"

	return this
}

// NeuroDataChecklistParser tokens.
const (
	NeuroDataChecklistParserEOF           = antlr.TokenEOF
	NeuroDataChecklistParserMETADATA_LINE = 1
	NeuroDataChecklistParserCOMMENT_LINE  = 2
	NeuroDataChecklistParserHYPHEN        = 3
	NeuroDataChecklistParserLBRACK        = 4
	NeuroDataChecklistParserRBRACK        = 5
	NeuroDataChecklistParserMARK          = 6
	NeuroDataChecklistParserTEXT          = 7
	NeuroDataChecklistParserNEWLINE       = 8
	NeuroDataChecklistParserWS            = 9
)

// NeuroDataChecklistParser rules.
const (
	NeuroDataChecklistParserRULE_checklistFile = 0
	NeuroDataChecklistParserRULE_itemLine      = 1
)

// IChecklistFileContext is an interface to support dynamic dispatch.
type IChecklistFileContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	AllNEWLINE() []antlr.TerminalNode
	NEWLINE(i int) antlr.TerminalNode
	AllItemLine() []IItemLineContext
	ItemLine(i int) IItemLineContext

	// IsChecklistFileContext differentiates from other interfaces.
	IsChecklistFileContext()
}

type ChecklistFileContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyChecklistFileContext() *ChecklistFileContext {
	var p = new(ChecklistFileContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroDataChecklistParserRULE_checklistFile
	return p
}

func InitEmptyChecklistFileContext(p *ChecklistFileContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroDataChecklistParserRULE_checklistFile
}

func (*ChecklistFileContext) IsChecklistFileContext() {}

func NewChecklistFileContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ChecklistFileContext {
	var p = new(ChecklistFileContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroDataChecklistParserRULE_checklistFile

	return p
}

func (s *ChecklistFileContext) GetParser() antlr.Parser { return s.parser }

func (s *ChecklistFileContext) EOF() antlr.TerminalNode {
	return s.GetToken(NeuroDataChecklistParserEOF, 0)
}

func (s *ChecklistFileContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(NeuroDataChecklistParserNEWLINE)
}

func (s *ChecklistFileContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(NeuroDataChecklistParserNEWLINE, i)
}

func (s *ChecklistFileContext) AllItemLine() []IItemLineContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IItemLineContext); ok {
			len++
		}
	}

	tst := make([]IItemLineContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IItemLineContext); ok {
			tst[i] = t.(IItemLineContext)
			i++
		}
	}

	return tst
}

func (s *ChecklistFileContext) ItemLine(i int) IItemLineContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IItemLineContext); ok {
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

	return t.(IItemLineContext)
}

func (s *ChecklistFileContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ChecklistFileContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ChecklistFileContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroDataChecklistListener); ok {
		listenerT.EnterChecklistFile(s)
	}
}

func (s *ChecklistFileContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroDataChecklistListener); ok {
		listenerT.ExitChecklistFile(s)
	}
}

func (s *ChecklistFileContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroDataChecklistVisitor:
		return t.VisitChecklistFile(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroDataChecklistParser) ChecklistFile() (localctx IChecklistFileContext) {
	localctx = NewChecklistFileContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, NeuroDataChecklistParserRULE_checklistFile)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(7)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroDataChecklistParserNEWLINE {
		{
			p.SetState(4)
			p.Match(NeuroDataChecklistParserNEWLINE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(9)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(19)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroDataChecklistParserHYPHEN || _la == NeuroDataChecklistParserWS {
		{
			p.SetState(10)
			p.ItemLine()
		}
		p.SetState(14)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == NeuroDataChecklistParserNEWLINE {
			{
				p.SetState(11)
				p.Match(NeuroDataChecklistParserNEWLINE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(16)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

		p.SetState(21)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(22)
		p.Match(NeuroDataChecklistParserEOF)
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

// IItemLineContext is an interface to support dynamic dispatch.
type IItemLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	HYPHEN() antlr.TerminalNode
	LBRACK() antlr.TerminalNode
	MARK() antlr.TerminalNode
	RBRACK() antlr.TerminalNode
	TEXT() antlr.TerminalNode
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode

	// IsItemLineContext differentiates from other interfaces.
	IsItemLineContext()
}

type ItemLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyItemLineContext() *ItemLineContext {
	var p = new(ItemLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroDataChecklistParserRULE_itemLine
	return p
}

func InitEmptyItemLineContext(p *ItemLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = NeuroDataChecklistParserRULE_itemLine
}

func (*ItemLineContext) IsItemLineContext() {}

func NewItemLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ItemLineContext {
	var p = new(ItemLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = NeuroDataChecklistParserRULE_itemLine

	return p
}

func (s *ItemLineContext) GetParser() antlr.Parser { return s.parser }

func (s *ItemLineContext) HYPHEN() antlr.TerminalNode {
	return s.GetToken(NeuroDataChecklistParserHYPHEN, 0)
}

func (s *ItemLineContext) LBRACK() antlr.TerminalNode {
	return s.GetToken(NeuroDataChecklistParserLBRACK, 0)
}

func (s *ItemLineContext) MARK() antlr.TerminalNode {
	return s.GetToken(NeuroDataChecklistParserMARK, 0)
}

func (s *ItemLineContext) RBRACK() antlr.TerminalNode {
	return s.GetToken(NeuroDataChecklistParserRBRACK, 0)
}

func (s *ItemLineContext) TEXT() antlr.TerminalNode {
	return s.GetToken(NeuroDataChecklistParserTEXT, 0)
}

func (s *ItemLineContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(NeuroDataChecklistParserWS)
}

func (s *ItemLineContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(NeuroDataChecklistParserWS, i)
}

func (s *ItemLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ItemLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ItemLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroDataChecklistListener); ok {
		listenerT.EnterItemLine(s)
	}
}

func (s *ItemLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(NeuroDataChecklistListener); ok {
		listenerT.ExitItemLine(s)
	}
}

func (s *ItemLineContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case NeuroDataChecklistVisitor:
		return t.VisitItemLine(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *NeuroDataChecklistParser) ItemLine() (localctx IItemLineContext) {
	localctx = NewItemLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, NeuroDataChecklistParserRULE_itemLine)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(27)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroDataChecklistParserWS {
		{
			p.SetState(24)
			p.Match(NeuroDataChecklistParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(29)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(30)
		p.Match(NeuroDataChecklistParserHYPHEN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(32)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == NeuroDataChecklistParserWS {
		{
			p.SetState(31)
			p.Match(NeuroDataChecklistParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(34)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(36)
		p.Match(NeuroDataChecklistParserLBRACK)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(40)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroDataChecklistParserWS {
		{
			p.SetState(37)
			p.Match(NeuroDataChecklistParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(42)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(43)
		p.Match(NeuroDataChecklistParserMARK)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(47)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == NeuroDataChecklistParserWS {
		{
			p.SetState(44)
			p.Match(NeuroDataChecklistParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(49)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(50)
		p.Match(NeuroDataChecklistParserRBRACK)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(52)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == NeuroDataChecklistParserWS {
		{
			p.SetState(51)
			p.Match(NeuroDataChecklistParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(54)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(56)
		p.Match(NeuroDataChecklistParserTEXT)
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
