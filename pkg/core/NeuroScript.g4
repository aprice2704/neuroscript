// pkg/core/NeuroScript.g4
grammar NeuroScript;

// --- PARSER RULES ---

// Program can start with optional newlines, then an optional version declaration,
// then more optional newlines, then procedure definitions, then optional newlines, then EOF.
program: optional_newlines file_version_decl? optional_newlines procedure_definition* optional_newlines EOF;

optional_newlines: NEWLINE*;

// Rule for the version declaration
file_version_decl: KW_FILE_VERSION STRING_LIT NEWLINE;

procedure_definition:
    KW_DEFINE KW_PROCEDURE IDENTIFIER
    LPAREN param_list_opt RPAREN NEWLINE
    COMMENT_BLOCK?
    statement_list
    KW_END NEWLINE?;

param_list_opt: param_list?;
param_list: IDENTIFIER (COMMA IDENTIFIER)*;

statement_list: body_line*;

body_line: statement NEWLINE | NEWLINE;

statement: simple_statement | block_statement;

simple_statement:
    set_statement
    | call_statement
    | return_statement
    | emit_statement;

block_statement:
    if_statement
    | while_statement
    | for_each_statement;

set_statement: KW_SET IDENTIFIER ASSIGN expression;
call_statement: KW_CALL call_target LPAREN expression_list_opt RPAREN;
return_statement: KW_RETURN expression?;
emit_statement: KW_EMIT expression;

if_statement:
    KW_IF condition KW_THEN NEWLINE
    if_body=statement_list
    (KW_ELSE NEWLINE else_body=statement_list)?
    KW_ENDBLOCK;

while_statement:
    KW_WHILE condition KW_DO NEWLINE
    statement_list
    KW_ENDBLOCK;

for_each_statement:
    KW_FOR KW_EACH IDENTIFIER KW_IN expression KW_DO NEWLINE
    statement_list
    KW_ENDBLOCK;

call_target: IDENTIFIER | KW_TOOL DOT IDENTIFIER | KW_LLM;

condition:
    expression ( (EQ | NEQ | GT | LT | GTE | LTE) expression )?;

expression: term ( optional_newlines PLUS optional_newlines term )*;

term: primary ( LBRACK expression RBRACK )*;

primary:
    literal
    | placeholder
    | IDENTIFIER
    | KW_LAST
    | KW_EVAL LPAREN expression RPAREN
    | LPAREN expression RPAREN;

placeholder: PLACEHOLDER_START (IDENTIFIER | KW_LAST) PLACEHOLDER_END;

literal:
    STRING_LIT
    | NUMBER_LIT
    | list_literal
    | map_literal
    | boolean_literal;

boolean_literal: KW_TRUE | KW_FALSE;

list_literal: LBRACK expression_list_opt RBRACK;
map_literal: LBRACE map_entry_list_opt RBRACE;

expression_list_opt: expression_list?;
expression_list: expression (COMMA expression)*;

map_entry_list_opt: map_entry_list?;
map_entry_list: map_entry (COMMA map_entry)*;
map_entry: STRING_LIT COLON expression;


// --- LEXER RULES ---

KW_FILE_VERSION: 'FILE_VERSION'; // Added
KW_DEFINE: 'DEFINE';
KW_PROCEDURE: 'PROCEDURE';
KW_END: 'END';
KW_ENDBLOCK: 'ENDBLOCK';
KW_COMMENT_START: 'COMMENT:';
KW_ENDCOMMENT: 'ENDCOMMENT';
KW_SET: 'SET';
KW_CALL: 'CALL';
KW_RETURN: 'RETURN';
KW_IF: 'IF';
KW_THEN: 'THEN';
KW_ELSE: 'ELSE';
KW_WHILE: 'WHILE';
KW_DO: 'DO';
KW_FOR: 'FOR';
KW_EACH: 'EACH';
KW_IN: 'IN';
KW_TOOL: 'TOOL';
KW_LLM: 'LLM';
KW_LAST: 'LAST';
KW_EVAL: 'EVAL';
KW_EMIT: 'EMIT';
KW_TRUE: 'true';
KW_FALSE: 'false';

COMMENT_BLOCK: KW_COMMENT_START .*? KW_ENDCOMMENT -> skip;

NUMBER_LIT: [0-9]+ ('.' [0-9]+)?;
STRING_LIT:
    '"' (EscapeSequence | ~["\\\r\n])* '"'
    | '\'' (EscapeSequence | ~['\\\r\n])* '\'';

ASSIGN: '=';
PLUS: '+';
LPAREN: '(';
RPAREN: ')';
COMMA: ',';
LBRACK: '[';
RBRACK: ']';
LBRACE: '{';
RBRACE: '}';
COLON: ':';
DOT: '.';
SLASH: '/';
PLACEHOLDER_START: '{{';
PLACEHOLDER_END: '}}';
EQ: '==';
NEQ: '!=';
GT: '>';
LT: '<';
GTE: '>=';
LTE: '<=';

IDENTIFIER: [a-zA-Z_] [a-zA-Z0-9_]*;

// --- Comments and Whitespace ---
// Updated LINE_COMMENT to include hash-bang #!
LINE_COMMENT: ('#' ~[!]? | '--') ~[\r\n]* -> skip; // Skips #comment, --comment BUT NOT #!
HASH_BANG: '#!' ~[\r\n]* -> skip; // Explicitly skip hash-bang lines

NEWLINE: '\r'? '\n' | '\r';
WS: [ \t]+ -> skip;

fragment EscapeSequence: '\\' (["'\\nrt] | UNICODE_ESC);
fragment UNICODE_ESC: 'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT;
fragment HEX_DIGIT: [0-9a-fA-F];