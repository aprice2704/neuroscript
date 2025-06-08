// File:        NeuroScript.g4
// Grammar:     NeuroScript
// Version:     0.4.1
// File version: 61
// Ver_Comment: Added new built-in type keywords (error, event, timedate, fuzzy).
// Date:        2025-06-08
// NOTES:

grammar NeuroScript;

// --- LEXER RULES (Order matters!) ---

// Global line continuation: Highest precedence to consume '\' + newline between tokens.
LINE_ESCAPE_GLOBAL : '\\' ('\r'? '\n' | '\r') -> channel(HIDDEN);

// Keywords
KW_ACOS       : 'acos';
KW_AND        : 'and';
KW_ASIN       : 'asin';
KW_ASK        : 'ask';
KW_ATAN       : 'atan';
KW_BREAK      : 'break';
KW_CALL       : 'call';
KW_CLEAR_ERROR: 'clear_error';
KW_CONTINUE   : 'continue';
KW_COS        : 'cos';
KW_EACH       : 'each';
KW_ELSE       : 'else';
KW_EMIT       : 'emit';
KW_ENDFOR     : 'endfor';
KW_ENDFUNC    : 'endfunc';
KW_ENDIF      : 'endif';
KW_ENDON      : 'endon';
KW_ENDWHILE   : 'endwhile';
KW_ERROR      : 'error';
KW_EVAL       : 'eval';
KW_EVENT      : 'event';
KW_FAIL       : 'fail';
KW_FALSE      : 'false';
KW_FOR        : 'for';
KW_FUNC       : 'func';
KW_FUZZY      : 'fuzzy';
KW_IF         : 'if';
KW_IN         : 'in';
KW_INTO       : 'into';
KW_LAST       : 'last';
KW_LN         : 'ln';
KW_LOG        : 'log';
KW_MEANS      : 'means';
KW_MUST       : 'must';
KW_MUSTBE     : 'mustbe';
KW_NEEDS      : 'needs';
KW_NIL        : 'nil';
KW_NO         : 'no';
KW_NOT        : 'not';
KW_ON_ERROR   : 'on_error';
KW_OPTIONAL   : 'optional';
KW_OR         : 'or';
KW_RETURN     : 'return';
KW_RETURNS    : 'returns';
KW_SET        : 'set';
KW_SIN        : 'sin';
KW_SOME       : 'some';
KW_TAN        : 'tan';
KW_TIMEDATE   : 'timedate';
KW_TOOL       : 'tool';
KW_TRUE       : 'true';
KW_TYPEOF     : 'typeof';
KW_WHILE      : 'while';


// Re-instated for multi-line capability within strings/metadata
fragment CONTINUED_LINE: '\\' ('\r'? '\n' | '\r') ;
fragment STRING_DQ_ATOM : EscapeSequence | CONTINUED_LINE | ~["\\\r\n] ;
fragment STRING_SQ_ATOM : EscapeSequence | CONTINUED_LINE | ~['\\\r\n] ;

STRING_LIT:
    '"' (STRING_DQ_ATOM)*? '"'
    | '\'' (STRING_SQ_ATOM)*? '\''
    ;

TRIPLE_BACKTICK_STRING: '```' .*? '```';

fragment METADATA_CONTENT_ATOM: EscapeSequence | CONTINUED_LINE | ~[\\\r\n] ;
METADATA_LINE:
    [\t ]* '::' [ \t]+
    (METADATA_CONTENT_ATOM)* ;

NUMBER_LIT: [0-9]+ ('.' [0-9]+)?;
IDENTIFIER: [a-zA-Z_] [a-zA-Z0-9_]*;

ASSIGN: '=';
PLUS: '+';
MINUS: '-';
STAR: '*';
SLASH: '/';
PERCENT: '%';
STAR_STAR: '**';
AMPERSAND: '&';
PIPE: '|';
CARET: '^';
TILDE: '~';
LPAREN: '(';
RPAREN: ')';
COMMA: ',';
LBRACK: '['; // Left square bracket
RBRACK: ']'; // Right square bracket
LBRACE: '{';
RBRACE: '}';
COLON: ':';
DOT: '.';
PLACEHOLDER_START: '{{';
PLACEHOLDER_END: '}}';
EQ: '==';
NEQ: '!=';
GT: '>';
LT: '<';
GTE: '>=';
LTE: '<=';

LINE_COMMENT: ('#' | '--' | '//') ~[\r\n]* -> channel(HIDDEN);
NEWLINE: ('\r'? '\n' | '\r');
WS: [ \t]+ -> channel(HIDDEN);

fragment EscapeSequence: '\\' ( UNICODE_ESC | HEX_ESC | OCTAL_ESC | CHAR_ESC );
fragment CHAR_ESC: ["'\\btnfrv~`]; // Note: Added ~ and ` as escapable characters based on some contexts
fragment UNICODE_ESC: 'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT;
fragment HEX_ESC: 'x' HEX_DIGIT HEX_DIGIT;
fragment OCTAL_ESC: [0-3]? [0-7] [0-7]; // Allows up to 377 octal
fragment HEX_DIGIT: [0-9a-fA-F];

// --- PARSER RULES ---
program: file_header (procedure_definition NEWLINE*)* EOF;

file_header: (METADATA_LINE | NEWLINE)*; // Optional metadata lines at the start

procedure_definition:
    KW_FUNC IDENTIFIER signature_part KW_MEANS NEWLINE
    metadata_block?
    statement_list
    KW_ENDFUNC;

signature_part:
    LPAREN needs_clause? optional_clause? returns_clause? RPAREN
    | needs_clause optional_clause? returns_clause?
    | optional_clause returns_clause?
    | returns_clause
    | /* empty */ ; // Allows func foo means ...

needs_clause: KW_NEEDS param_list;
optional_clause: KW_OPTIONAL param_list;
returns_clause: KW_RETURNS param_list;
param_list: IDENTIFIER (COMMA IDENTIFIER)*;

metadata_block: (METADATA_LINE NEWLINE)*; // Metadata specific to a function

statement_list: body_line*;
body_line: statement NEWLINE | NEWLINE;

statement: simple_statement | block_statement ;

simple_statement:
    set_statement
    | call_statement
    | return_statement
    | emit_statement
    | must_statement
    | fail_statement
    | clearErrorStmt
    | ask_stmt
    | break_statement
    | continue_statement
    ;

block_statement:
    if_statement
    | while_statement
    | for_each_statement
    | onErrorStmt
    ;

// --- MODIFIED RULE for set_statement and NEW lvalue rule ---
lvalue: IDENTIFIER ( LBRACK expression RBRACK | DOT IDENTIFIER )* ;

set_statement: KW_SET lvalue ASSIGN expression;
// --- END OF MODIFIED RULES ---

expressionStatement: expression ; // For potential future use or expression-only lines if allowed

call_statement: KW_CALL callable_expr;

return_statement: KW_RETURN expression_list?;

emit_statement: KW_EMIT expression;

must_statement: KW_MUST expression | KW_MUSTBE callable_expr;

fail_statement: KW_FAIL expression?;

clearErrorStmt: KW_CLEAR_ERROR;

ask_stmt: KW_ASK expression (KW_INTO IDENTIFIER)? ;

break_statement: KW_BREAK;
continue_statement: KW_CONTINUE;

if_statement:
    KW_IF expression NEWLINE
    statement_list
    (KW_ELSE NEWLINE statement_list)?
    KW_ENDIF;

while_statement:
    KW_WHILE expression NEWLINE
    statement_list
    KW_ENDWHILE;

for_each_statement:
    KW_FOR KW_EACH IDENTIFIER KW_IN expression NEWLINE
    statement_list
    KW_ENDFOR;

onErrorStmt:
    KW_ON_ERROR KW_MEANS NEWLINE
    statement_list
    KW_ENDON;

qualified_identifier: IDENTIFIER (DOT IDENTIFIER)*;

call_target: IDENTIFIER | KW_TOOL DOT qualified_identifier ;

expression: logical_or_expr; // Precedence: OR is lowest for binary logical
logical_or_expr: logical_and_expr (KW_OR logical_and_expr)*;
logical_and_expr: bitwise_or_expr (KW_AND bitwise_or_expr)*;
bitwise_or_expr: bitwise_xor_expr (PIPE bitwise_xor_expr)*;
bitwise_xor_expr: bitwise_and_expr (CARET bitwise_and_expr)*;
bitwise_and_expr: equality_expr (AMPERSAND equality_expr)*;
equality_expr: relational_expr ((EQ | NEQ) relational_expr)*;
relational_expr: additive_expr ((GT | LT | GTE | LTE) additive_expr)*;
additive_expr: multiplicative_expr ((PLUS | MINUS) multiplicative_expr)*;
multiplicative_expr: unary_expr ((STAR | SLASH | PERCENT) unary_expr)*;
unary_expr:
    (MINUS | KW_NOT | KW_NO | KW_SOME | TILDE) unary_expr
    | KW_TYPEOF unary_expr
    | power_expr
    ;
power_expr: accessor_expr (STAR_STAR power_expr)?; // Right-associative for power
accessor_expr: primary ( LBRACK expression RBRACK )* ; // Handles x[y][z] type access
primary:
    literal
    | placeholder
    | IDENTIFIER
    | KW_LAST
    | callable_expr
    | KW_EVAL LPAREN expression RPAREN
    | LPAREN expression RPAREN
    ;

callable_expr:
    ( call_target | KW_LN | KW_LOG | KW_SIN | KW_COS | KW_TAN | KW_ASIN | KW_ACOS | KW_ATAN )
    LPAREN expression_list_opt RPAREN;

placeholder: PLACEHOLDER_START (IDENTIFIER | KW_LAST) PLACEHOLDER_END;

literal:
    STRING_LIT
    | TRIPLE_BACKTICK_STRING
    | NUMBER_LIT
    | list_literal
    | map_literal
    | boolean_literal
    | nil_literal
    ;

nil_literal: KW_NIL;
boolean_literal: KW_TRUE | KW_FALSE;

list_literal: LBRACK expression_list_opt RBRACK;
map_literal: LBRACE map_entry_list_opt RBRACE;

expression_list_opt: expression_list?;
expression_list: expression (COMMA expression)*;

map_entry_list_opt: map_entry_list?;
map_entry_list: map_entry (COMMA map_entry)*;
map_entry: STRING_LIT COLON expression;