// pkg/core/NeuroScript.g4
// Corrected grammar to allow optional newlines around PLUS in expressions

grammar NeuroScript;

// --- Parser Rules ---
program             : optional_newlines (procedure_definition)* optional_newlines EOF;

optional_newlines   : NEWLINE*; // Matches zero or more newlines

procedure_definition: KW_STARTPROC KW_PROCEDURE IDENTIFIER LPAREN param_list_opt RPAREN NEWLINE
                      COMMENT_BLOCK? statement_list KW_END NEWLINE?;

param_list_opt      : param_list?;
param_list          : IDENTIFIER (COMMA IDENTIFIER)*;

statement_list      : (body_line)*;
body_line           : statement NEWLINE | NEWLINE; // Statement must end with NEWLINE

statement           : simple_statement | block_statement;

simple_statement    : set_statement | call_statement | return_statement | emit_statement;
block_statement     : if_statement | while_statement | for_each_statement;

set_statement       : KW_SET IDENTIFIER ASSIGN expression;
call_statement      : KW_CALL call_target LPAREN expression_list_opt RPAREN;
return_statement    : KW_RETURN expression?;
emit_statement      : KW_EMIT expression;

if_statement        : KW_IF condition KW_THEN NEWLINE statement_list KW_ENDBLOCK;
while_statement     : KW_WHILE condition KW_DO NEWLINE statement_list KW_ENDBLOCK;
for_each_statement  : KW_FOR KW_EACH IDENTIFIER KW_IN expression KW_DO NEWLINE statement_list KW_ENDBLOCK;

call_target         : IDENTIFIER
                    | KW_TOOL DOT IDENTIFIER
                    | KW_LLM;

condition           : expression ( ( EQ | NEQ | GT | LT | GTE | LTE ) expression )?;

// --- Expression Hierarchy (Revised for Newlines) ---
// Allow optional newlines around the PLUS operator for multi-line concatenation
expression          : term ( optional_newlines PLUS optional_newlines term )*; // <<< MODIFIED

// Term is a primary followed by zero or more element accesses
term                : primary ( LBRACK expression RBRACK )*;

// Primary represents the base, non-recursive parts of an expression
primary             : literal
                    | placeholder
                    | IDENTIFIER             // Includes variables and true/false
                    | KW_LAST_CALL_RESULT
                    | LPAREN expression RPAREN;

// Placeholder for {{var}} syntax
placeholder         : PLACEHOLDER_START IDENTIFIER PLACEHOLDER_END;

// Literal types
literal             : STRING_LIT
                    | NUMBER_LIT
                    | list_literal
                    | map_literal;

// List Literal (e.g., [1, "a", {{v}}])
list_literal        : LBRACK expression_list_opt RBRACK;

// Map Literal (e.g., {"key": "value", "num": 10})
map_literal         : LBRACE map_entry_list_opt RBRACE;

// Optional list of expressions (for calls and list literals)
expression_list_opt : expression_list?;
expression_list     : expression (COMMA expression)*;

// Optional list of map entries
map_entry_list_opt  : map_entry_list?;
map_entry_list      : map_entry (COMMA map_entry)*;
map_entry           : STRING_LIT COLON expression; // Map keys must be string literals


// --- Lexer Rules ---
KW_STARTPROC        : 'SPLAT';
KW_PROCEDURE        : 'PROCEDURE';
KW_END              : 'END';
KW_ENDBLOCK         : 'ENDBLOCK';
KW_COMMENT_START    : 'COMMENT:';
KW_ENDCOMMENT       : 'ENDCOMMENT';

KW_SET              : 'SET';
KW_CALL             : 'CALL';
KW_RETURN           : 'RETURN';
KW_IF               : 'IF';
KW_THEN             : 'THEN';
KW_ELSE             : 'ELSE';
KW_WHILE            : 'WHILE';
KW_DO               : 'DO';
KW_FOR              : 'FOR';
KW_EACH             : 'EACH';
KW_IN               : 'IN';
KW_TOOL             : 'TOOL';
KW_LLM              : 'LLM';
KW_LAST_CALL_RESULT : '__last_call_result';
KW_EMIT             : 'EMIT';

COMMENT_BLOCK       : KW_COMMENT_START .*? KW_ENDCOMMENT -> skip;

NUMBER_LIT          : [0-9]+ ('.' [0-9]+)?;
STRING_LIT          : '"' ( EscapeSequence | ~('\\'|'"') )* '"'
                    | '\'' ( EscapeSequence | ~('\\'|'\'') )* '\''
                    ;

ASSIGN              : '=';
PLUS                : '+';
LPAREN              : '(';
RPAREN              : ')';
COMMA               : ',';
LBRACK              : '[';
RBRACK              : ']';
LBRACE              : '{';
RBRACE              : '}';
COLON               : ':';
DOT                 : '.';
SLASH               : '/';
PLACEHOLDER_START   : '{{';
PLACEHOLDER_END     : '}}';
EQ                  : '==';
NEQ                 : '!=';
GT                  : '>';
LT                  : '<';
GTE                 : '>=';
LTE                 : '<=';

IDENTIFIER          : [a-zA-Z_] [a-zA-Z0-9_]*;

LINE_COMMENT        : ('#'|'--') ~[\r\n]* -> skip;
// NEWLINE stays on default channel - parser rules now handle where it's allowed/required
NEWLINE             : '\r'? '\n';
WS                  : [ \t]+ -> skip;

fragment EscapeSequence: '\\' (['"\\nrt] | UNICODE_ESC );
fragment UNICODE_ESC : 'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT;
fragment HEX_DIGIT   : [0-9a-fA-F];