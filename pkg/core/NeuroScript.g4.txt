// File:      NeuroScript.g4
// Grammar:   NeuroScript
// Version:   0.2.0-alpha-func-returns-1 // Incremented Version for function returns
// Date:      2025-04-29 // << Date adjusted based on context >>
// NOTE: Added ask_stmt rule Apr 30, 2025

grammar NeuroScript;

// --- PARSER RULES ---

// Allow optional NEWLINEs between procedure definitions
program: file_header (procedure_definition NEWLINE*)* EOF;

file_header: (METADATA_LINE | NEWLINE)*;

// Procedure definition uses optional parameter clauses block
procedure_definition:
    KW_FUNC IDENTIFIER
    parameter_clauses? // The whole clause group is optional
    KW_MEANS NEWLINE
    metadata_block?    // Optional metadata block after 'means'
    statement_list
    KW_ENDFUNC; // NEWLINE handled by program rule

// Groups parameter clauses and handles optional parens
parameter_clauses:
    // Option 1: Parens are present (clauses inside are still optional)
    LPAREN needs_clause? optional_clause? returns_clause? RPAREN
    // Option 2: Parens are absent - AT LEAST ONE clause must exist
    | needs_clause optional_clause? returns_clause?
    | optional_clause returns_clause?
    | returns_clause
    ;

// Parameter clause rules (unchanged)
needs_clause: KW_NEEDS param_list;
optional_clause: KW_OPTIONAL param_list;
returns_clause: KW_RETURNS param_list;

param_list: IDENTIFIER (COMMA IDENTIFIER)*;

// Metadata block rule: Use '*' and consume NEWLINE after each line
metadata_block: (METADATA_LINE NEWLINE)*;

// Statement list and body line allow blank lines
statement_list: body_line*;
body_line: statement NEWLINE | NEWLINE; // Each statement ends with NL, or just NL

// --- Statements ---
statement: simple_statement | block_statement ;

// MODIFIED: Added ask_stmt
simple_statement:
    set_statement
    // | call_statement // REMOVED
    | return_statement
    | emit_statement
    | must_statement
    | fail_statement
    | clearErrorStmt
    | ask_stmt         // <<< ADDED ask_stmt ALTERNATIVE
    ;

block_statement:
    if_statement
    | while_statement
    | for_each_statement
    | onErrorStmt;

// --- Simple Statements Details ---
set_statement: KW_SET IDENTIFIER ASSIGN expression;
// call_statement rule REMOVED
return_statement: KW_RETURN expression_list?;
emit_statement: KW_EMIT expression;
must_statement: KW_MUST expression | KW_MUSTBE callable_expr;
fail_statement: KW_FAIL expression?;
clearErrorStmt: KW_CLEAR_ERROR;
ask_stmt: KW_ASK expression (KW_INTO IDENTIFIER)? ; // <<< ADDED ask_stmt RULE DEFINITION

// --- Block Statements Details (Unchanged structurally, END terminators handled) ---
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

// --- Call Target (Used within callable_expr now) ---
call_target: IDENTIFIER | KW_TOOL DOT IDENTIFIER;

// --- Expression Rules with Precedence --- (Largely Unchanged down to primary)
expression: logical_or_expr;
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
    (MINUS | KW_NOT | KW_NO | KW_SOME) unary_expr
    | power_expr;
power_expr:
    accessor_expr (STAR_STAR power_expr)?;
accessor_expr:
    primary ( LBRACK expression RBRACK )* ; // Map/List access

// MODIFIED: Primary now directly includes callable expressions
primary:
    literal
    | placeholder
    | IDENTIFIER // Simple variable access
    | KW_LAST
    | callable_expr // <<< ADDED: Function/Tool calls are now primary expressions
    | KW_EVAL LPAREN expression RPAREN
    | LPAREN expression RPAREN;

// NEW/REVISED: Rule for callable expressions (replaces old function_call)
callable_expr:
    ( call_target // User proc, tool call (e.g., myFunc, tool.ReadFile)
    | KW_LN | KW_LOG | KW_SIN | KW_COS | KW_TAN | KW_ASIN | KW_ACOS | KW_ATAN // Built-ins
    )
    LPAREN expression_list_opt RPAREN;

// function_call rule REMOVED (subsumed by callable_expr in primary)

placeholder: PLACEHOLDER_START (IDENTIFIER | KW_LAST) PLACEHOLDER_END;
literal:
    STRING_LIT
    | TRIPLE_BACKTICK_STRING
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
map_entry: STRING_LIT COLON expression; // Map keys must be string literals

// --- LEXER RULES --- (Unchanged)

// Keywords (All Lowercase)
KW_FUNC       : 'func';
KW_NEEDS      : 'needs';
KW_OPTIONAL   : 'optional';
KW_RETURNS    : 'returns';
KW_MEANS      : 'means';
KW_ENDFUNC    : 'endfunc';
KW_SET        : 'set';
KW_RETURN     : 'return';
KW_EMIT       : 'emit';
KW_IF         : 'if';
KW_ELSE       : 'else';
KW_ENDIF      : 'endif';
KW_WHILE      : 'while';
KW_ENDWHILE   : 'endwhile';
KW_FOR        : 'for';
KW_EACH       : 'each';
KW_IN         : 'in';
KW_ENDFOR     : 'endfor';
KW_ON_ERROR   : 'on_error';
KW_ENDON      : 'endon';
KW_CLEAR_ERROR: 'clear_error';
KW_MUST       : 'must';
KW_MUSTBE     : 'mustbe';
KW_FAIL       : 'fail';
KW_NO         : 'no';
KW_SOME       : 'some';
KW_TOOL       : 'tool';
KW_LAST       : 'last';
KW_EVAL       : 'eval';
KW_TRUE       : 'true';
KW_FALSE      : 'false';
KW_AND        : 'and';
KW_OR         : 'or';
KW_NOT        : 'not';
KW_LN         : 'ln';
KW_LOG        : 'log';
KW_SIN        : 'sin';
KW_COS        : 'cos';
KW_TAN        : 'tan';
KW_ASIN       : 'asin';
KW_ACOS       : 'acos';
KW_ATAN       : 'atan';
KW_ASK        : 'ask';   // <<< ADDED KEYWORD
KW_INTO       : 'into';  // <<< ADDED KEYWORD

// Literals (Unchanged)
NUMBER_LIT: [0-9]+ ('.' [0-9]+)?;
STRING_LIT:
    '"' (EscapeSequence | ~["\\\r\n])* '"'
    | '\'' (EscapeSequence | ~['\\\r\n])* '\'';
TRIPLE_BACKTICK_STRING: '```' .*? '```';

// Operators (Unchanged)
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

// Punctuation (Unchanged)
LPAREN: '(';
RPAREN: ')';
COMMA: ',';
LBRACK: '[';
RBRACK: ']';
LBRACE: '{';
RBRACE: '}';
COLON: ':';
DOT: '.';
PLACEHOLDER_START: '{{';
PLACEHOLDER_END: '}}';

// Comparison (Unchanged)
EQ: '==';
NEQ: '!=';
GT: '>';
LT: '<';
GTE: '>=';
LTE: '<=';

IDENTIFIER: [a-zA-Z_] [a-zA-Z0-9_]*;

// --- Comments, Metadata, and Whitespace --- (Unchanged)
METADATA_LINE: [\t ]* '::' [ \t]+ ~[\r\n]* ;
LINE_COMMENT: ('#' | '--') ~[\r\n]* -> skip;
NEWLINE: '\r'? '\n' | '\r';
WS: [ \t]+ -> skip;
fragment EscapeSequence: '\\' (["'\\nrt] | UNICODE_ESC | '`');
fragment UNICODE_ESC: 'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT;
fragment HEX_DIGIT: [0-9a-fA-F];