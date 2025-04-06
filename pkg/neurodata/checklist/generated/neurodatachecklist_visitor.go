// Code generated from NeuroDataChecklist.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated // NeuroDataChecklist
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by NeuroDataChecklistParser.
type NeuroDataChecklistVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by NeuroDataChecklistParser#checklistFile.
	VisitChecklistFile(ctx *ChecklistFileContext) interface{}

	// Visit a parse tree produced by NeuroDataChecklistParser#itemLine.
	VisitItemLine(ctx *ItemLineContext) interface{}
}
