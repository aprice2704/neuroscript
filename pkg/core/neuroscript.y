/* neuroscript.y - Parser definition for NeuroScript (v9 - Separate %type lines) */

%{
package core // Must match package in lexer and ast

import (
	"fmt"
	"strings"
)

var parsedResult []Procedure
func newStep(typ string, target string, cond string, value interface{}, args []string) Step {
	// Assuming Step struct is defined in this package (ast.go)
	return Step{Type: typ, Target: target, Cond: cond, Value: value, Args: args}
}

// Global variable to hold the parsed docstring content temporarily
// TODO: Improve this mechanism, maybe pass lexer state to actions
var currentParsedDocstring string

%}

%union {
	str     string
	step    Step
	steps   []Step
	proc    Procedure
	procs   []Procedure
	params  []string
	args    []string // For CALL arguments
	expr    string
    exprs   []string // For expression lists (arguments, list literals)
    mapEntries []string // For map entries string representations
}

/* Token declarations */
%token <str> IDENTIFIER STRING_LIT DOC_COMMENT_CONTENT
%token KW_DEFINE KW_PROCEDURE KW_COMMENT KW_END KW_SET KW_CALL KW_RETURN KW_IF KW_THEN KW_WHILE KW_DO KW_FOR KW_EACH KW_IN KW_TOOL KW_LLM KW_ELSE
%token ASSIGN PLUS LPAREN RPAREN COMMA LBRACK RBRACK LBRACE RBRACE COLON DOT PLACEHOLDER_START PLACEHOLDER_END KW_LAST_CALL_RESULT
%token EQ NEQ
%token NEWLINE INVALID

/* Type declarations - One per line */
%type <procs> script_file
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


/* Operator precedence */
%left EQ NEQ
%left PLUS
// ...

%% /* Grammar rules start */

script_file: /* empty */ { parsedResult = []Procedure{} }
	| script_file procedure_definition { $$ = append($1, $2); parsedResult = $$ }
	| script_file NEWLINE { $$ = $1 } ;

procedure_definition:
	KW_DEFINE KW_PROCEDURE IDENTIFIER LPAREN param_list_opt RPAREN NEWLINE comment_block statement_list KW_END NEWLINE {
        var proc Procedure; proc.Name = $3; proc.Params = $5
        // TODO: Parse docstring $8
        proc.Steps = $9; $$ = proc
    } ;

param_list_opt: /* empty */ { $$ = []string{} } | param_list { $$ = $1 } ;
param_list: IDENTIFIER { $$ = []string{$1} } | param_list COMMA IDENTIFIER { $$ = append($1, $3) } ;

comment_block: KW_COMMENT NEWLINE DOC_COMMENT_CONTENT KW_END NEWLINE { $$ = $3 } ; // Returns raw docstring string

statement_list: /* empty */ { $$ = []Step{} }
    | statement_list statement { if $2.Type != "" { $$ = append($1, $2) } else { $$ = $1 } }
    | statement_list NEWLINE { $$ = $1 } ;

statement: // Statements DO NOT include their terminating NEWLINE here
	simple_statement { $$ = $1 }
	| block_statement { $$ = $1 }
	;

// Simple statements must be followed by NEWLINE in the context they are used
simple_statement: set_statement | call_statement | return_statement ;
block_statement: if_statement | while_statement | for_each_statement ;

// Definitions now DO NOT consume NEWLINE directly, handled by context (e.g. statement rule)
set_statement: KW_SET IDENTIFIER ASSIGN expression { $$ = newStep("SET", $2, "", $4, nil) } ;
call_statement: KW_CALL call_target LPAREN expression_list_opt RPAREN { $$ = newStep("CALL", $2, "", nil, $4) } ; // $4 is exprs ([]string) now
return_statement: KW_RETURN { $$ = newStep("RETURN", "", "", "", nil) } | KW_RETURN expression { $$ = newStep("RETURN", "", "", $2, nil) } ;

// Block statements consume NEWLINE between header/body and after END
if_statement: KW_IF condition KW_THEN NEWLINE statement_list KW_END { $$ = newStep("IF", "", $2, $5, nil) } ;
while_statement: KW_WHILE condition KW_DO NEWLINE statement_list KW_END { $$ = newStep("WHILE", "", $2, $5, nil) } ;
for_each_statement: KW_FOR KW_EACH IDENTIFIER KW_IN expression KW_DO NEWLINE statement_list KW_END { $$ = newStep("FOR", $3, $5, $8, nil) } ;

// call_target rule unchanged
call_target: IDENTIFIER { $$ = $1 } | KW_TOOL DOT IDENTIFIER { $$ = "TOOL." + $3 } | KW_LLM { $$ = "LLM" } ;

condition: expression { $$ = $1 } | expression EQ expression { $$ = $1 + "==" + $3 } | expression NEQ expression { $$ = $1 + "!=" + $3 } ;
expression: term { $$ = $1 } | expression PLUS term { $$ = $1 + " + " + $3 } ;
term: literal { $$ = $1 } | placeholder { $$ = $1 } | IDENTIFIER { $$ = $1 } | KW_LAST_CALL_RESULT { $$ = "__last_call_result" } | LPAREN expression RPAREN { $$ = "(" + $2 + ")" } ;
placeholder: PLACEHOLDER_START IDENTIFIER PLACEHOLDER_END { $$ = "{{" + $2 + "}}" } ;
literal: STRING_LIT { $$ = $1 } | list_literal { $$ = $1 } | map_literal { $$ = $1 } ;

// List/Map rules use the _opt rules which now have types
list_literal: LBRACK expression_list_opt RBRACK { $$ = "[" + strings.Join($2, ", ") + "]" } ; // $2 is exprs ([]string)
map_literal: LBRACE map_entry_list_opt RBRACE { $$ = "{" + strings.Join($2, ", ") + "}" } ; // $2 is mapEntries ([]string)

expression_list_opt: /* empty */ { $$ = []string{} } | expression_list_final { $$ = $1 } ; // Returns exprs
expression_list_final: expression { $$ = []string{$1} } | expression_list_final COMMA expression { $$ = append($1, $3) } ; // Returns exprs

map_entry_list_opt: /* empty */ { $$ = []string{} } | map_entry_list_final { $$ = $1 } ; // Returns mapEntries
map_entry_list_final: map_entry { $$ = []string{$1} } | map_entry_list_final COMMA map_entry { $$ = append($1, $3) } ; // Returns mapEntries
map_entry: STRING_LIT COLON expression { $$ = $1 + ":" + $3 } ; // Returns string representation of entry

%% /* Go code section */

// Removed duplicate yyLexer interface def
// Removed duplicate yyParse func def

// yyError is the error reporting function required by goyacc
func yyError(s string) {
	fmt.Printf("Syntax Error: %s\n", s)
}

// Add SetResult to lexer struct if results are stored there
// func (l *lexer) SetResult(res []Procedure) { l.result = res }