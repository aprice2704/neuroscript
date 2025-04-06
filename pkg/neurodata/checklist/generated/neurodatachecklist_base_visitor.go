// Code generated from NeuroDataChecklist.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated // NeuroDataChecklist
import "github.com/antlr4-go/antlr/v4"

type BaseNeuroDataChecklistVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseNeuroDataChecklistVisitor) VisitChecklistFile(ctx *ChecklistFileContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroDataChecklistVisitor) VisitItemLine(ctx *ItemLineContext) interface{} {
	return v.VisitChildren(ctx)
}
