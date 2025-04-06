// Code generated from NeuroDataChecklist.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated // NeuroDataChecklist
import "github.com/antlr4-go/antlr/v4"

// BaseNeuroDataChecklistListener is a complete listener for a parse tree produced by NeuroDataChecklistParser.
type BaseNeuroDataChecklistListener struct{}

var _ NeuroDataChecklistListener = &BaseNeuroDataChecklistListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseNeuroDataChecklistListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseNeuroDataChecklistListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseNeuroDataChecklistListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseNeuroDataChecklistListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterChecklistFile is called when production checklistFile is entered.
func (s *BaseNeuroDataChecklistListener) EnterChecklistFile(ctx *ChecklistFileContext) {}

// ExitChecklistFile is called when production checklistFile is exited.
func (s *BaseNeuroDataChecklistListener) ExitChecklistFile(ctx *ChecklistFileContext) {}

// EnterItemLine is called when production itemLine is entered.
func (s *BaseNeuroDataChecklistListener) EnterItemLine(ctx *ItemLineContext) {}

// ExitItemLine is called when production itemLine is exited.
func (s *BaseNeuroDataChecklistListener) ExitItemLine(ctx *ItemLineContext) {}
