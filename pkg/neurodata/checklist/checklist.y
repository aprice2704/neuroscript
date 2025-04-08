// --- Goyacc Grammar for Checklist (v7 - Metadata First, ':: ' Prefix) ---

// 1. Go Code Block (Top) - Define shared types here
%{
package checklist

import "fmt"

// CheckItem struct includes Indent
type CheckItem struct {
	Text   string
	Status string
	Indent int
}
// Type alias for metadata string slice (used by lexer & parser actions)
type MetadataLines = []string
%}

// 2. Declarations Block
%union {
	item  CheckItem
	items []CheckItem
	str   string         // For METADATA_STRING token value (raw line)
	strs  MetadataLines  // For metadataBlock result (list of raw lines)
}

// Declare tokens - COMMENT_LINE removed
%token <item> ITEM
%token <str>  METADATA_STRING // Represents ':: key: value' line
%token NEWLINE_TOK

// Declare types returned by non-terminal rules
%type <items> itemList
%type <strs>  metadataBlock
%type <item>  itemLine
// checklistFile implicitly handled via lexer fields

%% // 3. Grammar Rules Block

// Unified checklistFile rule: metadata (possibly empty) then items (possibly empty)
checklistFile:
	metadataBlock itemList
	{
		// Store results in the lexer instance
		lexerInstance, ok := yylex.(*ChecklistLexer) // Ensure ChecklistLexer type matches your lexer file
		if ok {
			 lexerInstance.MetadataLines = $1 // Store raw metadata lines ($1 from metadataBlock)
			 lexerInstance.Result = $2        // Store item list ($2 from itemList)
			 // Optional: Keep debug print or remove for production
			 fmt.Printf("Parsing finished. Stored %d metadata lines and %d items.\n", len($1), len($2))
		} else {
			 fmt.Println("ERROR: Could not assert yylex to *ChecklistLexer in checklistFile action.")
		}
	}
	;

// metadataBlock collects METADATA_STRING tokens (raw lines).
metadataBlock:
	  /* empty */ { $$ = make(MetadataLines, 0) } // Base case: empty slice
	| metadataBlock METADATA_STRING { $$ = append($1, $2) } // Append raw string ($2)
	;

// itemList - Allows empty list. Lexer skips '#' comments.
itemList:
	  /* empty */ { $$ = []CheckItem{} }
	| itemList itemLine { $$ = append($1, $2) }
	;

// itemLine remains the same
itemLine:
	ITEM NEWLINE_TOK { $$ = $1 }
	;

%% // 4. Go Code Block (Bottom)
// Lexer implementation is in lexer.go