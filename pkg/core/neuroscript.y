/* neuroscript.y - Parser definition for NeuroScript (v24 - Add comparison ops, number literal) */

%{
package core // Must match package in lexer and ast

import (
	"fmt"
	"strings"
)

// Helper function used in grammar actions
func newStep(typ string, target string, cond string, value interface{}, args []string) Step {
	return Step{Type: typ, Target: target, Cond: cond, Value: value, Args: args}
}

// No globals needed here, result managed via lexer state
%}

// Union definition for semantic values
%union {
	str        string
	step       Step
	steps      []Step
	proc       Procedure
	procs      []Procedure // Type for lists of procedures
	params     []string
	args       []string // For CALL arguments (Consider using exprs instead?)
	expr       string
	exprs      []string // For expression lists (arguments, list literals)
	mapEntries []string // For map entries string representations
}

/* Token declarations */
// Tokens with values
%token <str> IDENTIFIER STRING_LIT DOC_COMMENT_CONTENT NUMBER_LIT KW_LAST_CALL_RESULT // Added NUMBER_LIT

// Keywords
%token KW_DEFINE KW_PROCEDURE KW_COMMENT KW_END KW_SET KW_CALL KW_RETURN KW_IF KW_THEN KW_ELSE KW_WHILE KW_DO KW_FOR KW_EACH KW_IN KW_TOOL KW_LLM

// Operators and Delimiters
%token ASSIGN PLUS LPAREN RPAREN COMMA LBRACK RBRACK LBRACE RBRACE COLON DOT PLACEHOLDER_START PLACEHOLDER_END

// Comparison Operators
%token EQ NEQ GT LT GTE LTE // Added GT, LT, GTE, LTE

// Control Tokens
%token NEWLINE INVALID

/* Type declarations for non-terminals */
%type <procs> program procedure_list
%type <proc> procedure_definition
%type <params> param_list_opt param_list
%type <steps> statement_list
%type <step> statement simple_statement block_statement if_statement while_statement for_each_statement set_statement call_statement return_statement
%type <str> call_target
%type <str> expression
%type <str> term
%type <str> literal
%type <str> placeholder
%type <str> condition // Changed from expression to condition for clarity
%type <str> list_literal
%type <str> map_literal
%type <str> map_entry
%type <exprs> expression_list_opt expression_list_final // Renamed for clarity
%type <mapEntries> map_entry_list_opt map_entry_list_final // Renamed for clarity
%type <str> comment_block


/* Operator precedence (Define if complex expressions arise, simple for now) */
// %left PLUS
// %left GT LT GTE LTE // Comparison ops usually lower than arithmetic
// %left EQ NEQ        // Equality usually lower than comparison

/* Start symbol */
%start program

%% /* Grammar rules start */

// Top level rule - Use lexer object for result
program: procedure_list {
			if l, ok := yylex.(*lexer); ok {
				l.SetResult($1)
			} else {
				fmt.Println("Error: Could not access lexer object to set result.")
			}
		}
       ;

procedure_list: /* empty */ { $$ = []Procedure{} }
	| procedure_list procedure_definition { $$ = append($1, $2) }
	| procedure_list NEWLINE { $$ = $1 } // Allow blank lines between procedures
	;

procedure_definition:
	KW_DEFINE KW_PROCEDURE IDENTIFIER LPAREN param_list_opt RPAREN NEWLINE comment_block statement_list KW_END NEWLINE {
		var proc Procedure; proc.Name = $3; proc.Params = $5
		// TODO: Parse comment_block ($8) into proc.Docstring struct
		proc.Steps = $9; $$ = proc
	} ;


param_list_opt: /* empty */ { $$ = []string{} } | param_list { $$ = $1 } ;
param_list: IDENTIFIER { $$ = []string{$1} } | param_list COMMA IDENTIFIER { $$ = append($1, $3) } ;

comment_block: KW_COMMENT DOC_COMMENT_CONTENT KW_END NEWLINE { $$ = $2 } ; // Returns content as single string

statement_list: /* empty */ { $$ = []Step{} }
	| statement_list statement { if $2.Type != "" { $$ = append($1, $2) } else { $$ = $1 } }
	| statement_list NEWLINE { $$ = $1 } // Handles blank lines between statements
	;

// Statement requires NEWLINE terminator (back to v20 style, seems more robust)
statement:
      simple_statement NEWLINE { $$ = $1 }
    | block_statement NEWLINE { $$ = $1 } // Block statements also end before final NEWLINE in definition
    ;

// Simple statements are just the base rule
simple_statement: set_statement | call_statement | return_statement ;

// Block statements define structure up to their END keyword
block_statement: if_statement | while_statement | for_each_statement ;

// Definitions DO NOT consume terminating NEWLINEs
set_statement: KW_SET IDENTIFIER ASSIGN expression { $$ = newStep("SET", $2, "", $4, nil) } ;
call_statement: KW_CALL call_target LPAREN expression_list_opt RPAREN { $$ = newStep("CALL", $2, "", nil, $4) } ;
return_statement: KW_RETURN { $$ = newStep("RETURN", "", "", "", nil) } | KW_RETURN expression { $$ = newStep("RETURN", "", "", $2, nil) } ;

// Block statement rules include NEWLINE after THEN/DO and end simply with KW_END.
// The statement rule adds the final NEWLINE for the block.
if_statement: KW_IF condition KW_THEN NEWLINE statement_list KW_END { $$ = newStep("IF", "", $2, $5, nil) } ;
// TODO: Add IF/ELSE rule later
while_statement: KW_WHILE condition KW_DO NEWLINE statement_list KW_END { $$ = newStep("WHILE", "", $2, $5, nil) } ;
for_each_statement: KW_FOR KW_EACH IDENTIFIER KW_IN expression KW_DO NEWLINE statement_list KW_END { $$ = newStep("FOR", $3, $5, $8, nil) } ;


call_target: IDENTIFIER { $$ = $1 } | KW_TOOL DOT IDENTIFIER { $$ = "TOOL." + $3 } | KW_LLM { $$ = "LLM" } ;

// Updated condition rule
condition:
      expression EQ expression  { $$ = $1 + "==" + $3 }
    | expression NEQ expression { $$ = $1 + "!=" + $3 }
    | expression GT expression  { $$ = $1 + ">" + $3 }    // Added
    | expression LT expression  { $$ = $1 + "<" + $3 }    // Added
    | expression GTE expression { $$ = $1 + ">=" + $3 }   // Added
    | expression LTE expression { $$ = $1 + "<=" + $3 }   // Added
    | expression                { $$ = $1 }               // For single values like 'true', '{{var}}'
    ;

expression: term { $$ = $1 } | expression PLUS term { $$ = $1 + " + " + $3 } ; // Only string concat for now

term:
      literal
    | placeholder
    | IDENTIFIER         // Variable access
    | KW_LAST_CALL_RESULT { $$ = "__last_call_result" }
    | LPAREN expression RPAREN { $$ = "(" + $2 + ")" }
    ;

placeholder: PLACEHOLDER_START IDENTIFIER PLACEHOLDER_END { $$ = "{{" + $2 + "}}" } ; // Simple placeholder for now

literal:
      STRING_LIT   // Includes quotes
    | NUMBER_LIT   // Added
    | list_literal
    | map_literal
    ;

list_literal: LBRACK expression_list_opt RBRACK { $$ = "[" + strings.Join($2, ", ") + "]" } ;
map_literal: LBRACE map_entry_list_opt RBRACE { $$ = "{" + strings.Join($2, ", ") + "}" } ;

expression_list_opt: /* empty */ { $$ = []string{} } | expression_list_final { $$ = $1 } ;

expression_list_final: expression { $$ = []string{$1} } | expression_list_final COMMA expression { $$ = append($1, $3) } ;

map_entry_list_opt: /* empty */ { $$ = []string{} } | map_entry_list_final { $$ = $1 } ;

map_entry_list_final: map_entry { $$ = []string{$1} } | map_entry_list_final COMMA map_entry { $$ = append($1, $3) } ;

map_entry: STRING_LIT COLON expression { $$ = $1 + ":" + $3 } ; // Key must be string literal

%% /* Go code section */

// yyError function remains the same
func yyError(s string) {
	// Maybe enhance this later (e.g., integrate with lexer line/pos)
	fmt.Printf("Syntax Error: %s\n", s)
}