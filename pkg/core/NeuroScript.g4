// NeuroScript.g4 - Added SLASH rule

grammar NeuroScript;

options {
    language = Go;
}

// --- Parser Rules ---
// ... (Parser rules remain unchanged from V18 - COMMENT_BLOCK on default channel) ...

program             : optional_newlines procedure_definition* optional_newlines EOF;

optional_newlines   : NEWLINE*;

procedure_definition: KW_STARTPROC KW_PROCEDURE IDENTIFIER LPAREN param_list_opt RPAREN NEWLINE
                      COMMENT_BLOCK? // Optional COMMENT block (token matched by lexer)
                      statement_list
                      KW_END NEWLINE?;

param_list_opt      : param_list?;
param_list          : IDENTIFIER (COMMA IDENTIFIER)*;

statement_list      : body_line* ; // Uses body_line structure

body_line           : statement NEWLINE // A line with a statement
                    | NEWLINE         // An empty line
                    ;

statement           : simple_statement | block_statement ; // Statement no longer consumes NEWLINE

simple_statement    : set_statement | call_statement | return_statement | emit_statement;

block_statement     : if_statement | while_statement | for_each_statement;

set_statement       : KW_SET IDENTIFIER ASSIGN expression;
call_statement      : KW_CALL call_target LPAREN expression_list_opt RPAREN;
return_statement    : KW_RETURN expression?;
emit_statement      : KW_EMIT expression;

if_statement        : KW_IF condition KW_THEN NEWLINE statement_list KW_ENDBLOCK;
while_statement     : KW_WHILE condition KW_DO NEWLINE statement_list KW_ENDBLOCK;
for_each_statement  : KW_FOR KW_EACH IDENTIFIER KW_IN expression KW_DO NEWLINE statement_list KW_ENDBLOCK;

call_target         : IDENTIFIER | KW_TOOL DOT IDENTIFIER | KW_LLM;
condition           : expression ( (EQ | NEQ | GT | LT | GTE | LTE) expression )? ;
expression          : term (PLUS term)*;
term                : literal | placeholder | IDENTIFIER | KW_LAST_CALL_RESULT | LPAREN expression RPAREN;
placeholder         : PLACEHOLDER_START IDENTIFIER PLACEHOLDER_END;
literal             : STRING_LIT | NUMBER_LIT | list_literal | map_literal;
list_literal        : LBRACK expression_list_opt RBRACK;
map_literal         : LBRACE map_entry_list_opt RBRACE;
expression_list_opt : expression_list?;
expression_list     : expression (COMMA expression)*;
map_entry_list_opt  : map_entry_list?;
map_entry_list      : map_entry (COMMA map_entry)*;
map_entry           : STRING_LIT COLON expression;


// --- Lexer Rules ---

// Keywords
KW_STARTPROC        : 'SPLAT';
KW_PROCEDURE        : 'PROCEDURE';
KW_END              : 'END';
KW_ENDBLOCK         : 'ENDBLOCK';
KW_COMMENT_START    : 'COMMENT:';
KW_ENDCOMMENT       : 'ENDCOMMENT';
KW_SET              : 'SET';
KW_CALL             : 'CALL';
// ... (all other keywords remain the same) ...
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


// COMMENT_BLOCK on default channel
COMMENT_BLOCK       : KW_COMMENT_START .*? KW_ENDCOMMENT;

// Literals, Operators, etc.
NUMBER_LIT          : [0-9]+ ;
STRING_LIT          : '"' ( ~["\\\r\n] | EscapeSequence )* '"' | '\'' ( ~['\\\r\n] | EscapeSequence )* '\'';
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
SLASH               : '/' -> channel(HIDDEN); // ** V19: Added rule for SLASH, send to hidden **
PLACEHOLDER_START   : '{{';
PLACEHOLDER_END     : '}}';
EQ                  : '==';
NEQ                 : '!=';
GT                  : '>';
LT                  : '<';
GTE                 : '>=';
LTE                 : '<=';

// Identifier rule MUST come AFTER all keywords
IDENTIFIER          : [a-zA-Z_] [a-zA-Z0-9_]* ;

// Comments, Whitespace, Newlines
LINE_COMMENT        : ('#' | '--') ~[\r\n]* -> channel(HIDDEN) ;
NEWLINE             : '\r'? '\n' | '\r';
WS                  : [ \t]+ -> channel(HIDDEN) ;
LINE_CONTINUATION   : '\\' '\r'? '\n' -> skip;

// Fragments
fragment EscapeSequence : '\\' (["'\\nrt] | UNICODE_ESC) ;
fragment UNICODE_ESC : 'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT ;
fragment HEX_DIGIT   : [0-9a-fA-F] ;