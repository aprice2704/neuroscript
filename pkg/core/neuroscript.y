/* neuroscript.y - Parser definition for NeuroScript (v23 - Remove NEWLINE from simple_statement) */

%{
package core // Must match package in lexer and ast

import (
	"fmt"
	"strings"
)

// var parsedResult []Procedure // Removed global

func newStep(typ string, target string, cond string, value interface{}, args []string) Step {
	return Step{Type: typ, Target: target, Cond: cond, Value: value, Args: args}
}

var currentParsedDocstring string

%}

%union {
	str        string
	step       Step
	steps      []Step
	proc       Procedure
	procs      []Procedure // Type for lists of procedures
	params     []string
	args       []string // For CALL arguments
	expr       string
	exprs      []string // For expression lists (arguments, list literals)
	mapEntries []string // For map entries string representations
}

/* Token declarations (Unchanged) */
%token <str> IDENTIFIER STRING_LIT DOC_COMMENT_CONTENT
%token KW_DEFINE KW_PROCEDURE KW_COMMENT KW_END KW_SET KW_CALL KW_RETURN KW_IF KW_THEN KW_WHILE KW_DO KW_FOR KW_EACH KW_IN KW_TOOL KW_LLM KW_ELSE
%token ASSIGN PLUS LPAREN RPAREN COMMA LBRACK RBRACK LBRACE RBRACE COLON DOT PLACEHOLDER_START PLACEHOLDER_END KW_LAST_CALL_RESULT
%token EQ NEQ
%token NEWLINE INVALID

/* Type declarations (Unchanged) */
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
%type <str> condition
%type <str> list_literal
%type <str> map_literal
%type <str> map_entry
%type <exprs> expression_list_opt
%type <exprs> expression_list_final
%type <mapEntries> map_entry_list_opt
%type <mapEntries> map_entry_list_final
%type <str> comment_block


/* Operator precedence (Unchanged) */
%left EQ NEQ
%left PLUS
// ...

/* Start symbol (Unchanged) */
%start program

%% /* Grammar rules start */

// Top level rule - Use lexer object for result (v22)
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
	| procedure_list NEWLINE { $$ = $1 }
	;

procedure_definition:
	// Require NEWLINE after final KW_END (v18/v20 grammar)
	KW_DEFINE KW_PROCEDURE IDENTIFIER LPAREN param_list_opt RPAREN NEWLINE comment_block statement_list KW_END NEWLINE {
		var proc Procedure; proc.Name = $3; proc.Params = $5
		proc.Steps = $9; $$ = proc
	} ;

// Rules below are unchanged from v22 UNTIL 'statement'

param_list_opt: /* empty */ { $$ = []string{} } | param_list { $$ = $1 } ;
param_list: IDENTIFIER { $$ = []string{$1} } | param_list COMMA IDENTIFIER { $$ = append($1, $3) } ;

comment_block: KW_COMMENT DOC_COMMENT_CONTENT KW_END NEWLINE { $$ = $2 } ;

statement_list: /* empty */ { $$ = []Step{} }
	| statement_list statement { if $2.Type != "" { $$ = append($1, $2) } else { $$ = $1 } }
	| statement_list NEWLINE { $$ = $1 } // Handles blank lines between statements
	;

// *** MODIFIED statement rule: Remove explicit NEWLINE requirement ***
statement: simple_statement { $$ = $1 }
	| block_statement { $$ = $1 }
	;

// Simple statements are just the base rule (Unchanged)
simple_statement: set_statement | call_statement | return_statement ;

// Block statements define structure up to their END keyword (Unchanged)
block_statement: if_statement | while_statement | for_each_statement ;

// Definitions DO NOT consume terminating NEWLINEs (Unchanged)
set_statement: KW_SET IDENTIFIER ASSIGN expression { $$ = newStep("SET", $2, "", $4, nil) } ;
call_statement: KW_CALL call_target LPAREN expression_list_opt RPAREN { $$ = newStep("CALL", $2, "", nil, $4) } ;
return_statement: KW_RETURN { $$ = newStep("RETURN", "", "", "", nil) } | KW_RETURN expression { $$ = newStep("RETURN", "", "", $2, nil) } ;

// Block statement rules end simply with KW_END. (Unchanged)
if_statement: KW_IF condition KW_THEN statement_list KW_END { $$ = newStep("IF", "", $2, $4, nil) } ;
while_statement: KW_WHILE condition KW_DO statement_list KW_END { $$ = newStep("WHILE", "", $2, $4, nil) } ;
for_each_statement: KW_FOR KW_EACH IDENTIFIER KW_IN expression KW_DO statement_list KW_END { $$ = newStep("FOR", $3, $5, $7, nil) } ;

// --- Rules below remain unchanged ---
call_target: IDENTIFIER { $$ = $1 } | KW_TOOL DOT IDENTIFIER { $$ = "TOOL." + $3 } | KW_LLM { $$ = "LLM" } ;
condition: expression { $$ = $1 } | expression EQ expression { $$ = $1 + "==" + $3 } | expression NEQ expression { $$ = $1 + "!=" + $3 } ;
expression: term { $$ = $1 } | expression PLUS term { $$ = $1 + " + " + $3 } ;
term: literal { $$ = $1 } | placeholder { $$ = $1 } | IDENTIFIER { $$ = $1 } | KW_LAST_CALL_RESULT { $$ = "__last_call_result" } | LPAREN expression RPAREN { $$ = "(" + $2 + ")" } ;
placeholder: PLACEHOLDER_START IDENTIFIER PLACEHOLDER_END { $$ = "{{" + $2 + "}}" } ;
literal: STRING_LIT { $$ = $1 } | list_literal { $$ = $1 } | map_literal { $$ = $1 } ;
list_literal: LBRACK expression_list_opt RBRACK { $$ = "[" + strings.Join($2, ", ") + "]" } ;
map_literal: LBRACE map_entry_list_opt RBRACE { $$ = "{" + strings.Join($2, ", ") + "}" } ;
expression_list_opt: /* empty */ { $$ = []string{} } | expression_list_final { $$ = $1 } ;
expression_list_final: expression { $$ = []string{$1} } | expression_list_final COMMA expression { $$ = append($1, $3) } ;
map_entry_list_opt: /* empty */ { $$ = []string{} } | map_entry_list_final { $$ = $1 } ;
map_entry_list_final: map_entry { $$ = []string{$1} } | map_entry_list_final COMMA map_entry { $$ = append($1, $3) } ;
map_entry: STRING_LIT COLON expression { $$ = $1 + ":" + $3 } ;

%% /* Go code section */

// yyError function remains the same
func yyError(s string) {
	fmt.Printf("Syntax Error: %s\n", s)
}