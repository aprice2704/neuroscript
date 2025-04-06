// Code generated from NeuroDataChecklist.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated // NeuroDataChecklist
import "github.com/antlr4-go/antlr/v4"

// NeuroDataChecklistListener is a complete listener for a parse tree produced by NeuroDataChecklistParser.
type NeuroDataChecklistListener interface {
	antlr.ParseTreeListener

	// EnterChecklistFile is called when entering the checklistFile production.
	EnterChecklistFile(c *ChecklistFileContext)

	// EnterItemLine is called when entering the itemLine production.
	EnterItemLine(c *ItemLineContext)

	// ExitChecklistFile is called when exiting the checklistFile production.
	ExitChecklistFile(c *ChecklistFileContext)

	// ExitItemLine is called when exiting the itemLine production.
	ExitItemLine(c *ItemLineContext)
}
