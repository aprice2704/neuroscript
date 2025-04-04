// pkg/core/NeuroScript.g4
grammar NeuroScript;

// --- PARSER RULES ---

program: optional_newlines file_version_decl? optional_newlines procedure_definition* optional_newlines EOF;

optional_newlines: NEWLINE*;

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
    KW_IF expression KW_THEN NEWLINE // Use full expression for condition now
    if_body=statement_list
    (KW_ELSE NEWLINE else_body=statement_list)?
    KW_ENDBLOCK;

while_statement:
    KW_WHILE expression KW_DO NEWLINE // Use full expression for condition now
    statement_list
    KW_ENDBLOCK;

for_each_statement:
    KW_FOR KW_EACH IDENTIFIER KW_IN expression KW_DO NEWLINE
    statement_list
    KW_ENDBLOCK;

call_target: IDENTIFIER | KW_TOOL DOT IDENTIFIER | KW_LLM;

// --- Expression Rules with Precedence ---
// Lowest precedence first (parsed last)
expression: logical_or_expr; // Entry point for expressions

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
    (MINUS | KW_NOT) unary_expr // Add unary minus and NOT
    | power_expr;

power_expr: // Exponentiation (right-associative)
    // *** MODIFIED: Point to accessor_expr instead of primary ***
    accessor_expr (STAR_STAR power_expr)?;

// *** NEW RULE: Handles primary followed by zero or more [...] accessors ***
accessor_expr:
    primary ( LBRACK expression RBRACK )* ;

// Primary expression contains the base cases
primary:
    literal
    | placeholder
    | IDENTIFIER
    | KW_LAST
    | function_call // Added function call
    | KW_EVAL LPAREN expression RPAREN
    | LPAREN expression RPAREN;

function_call: // New rule for function calls
    ( KW_LN | KW_LOG | KW_SIN | KW_COS | KW_TAN | KW_ASIN | KW_ACOS | KW_ATAN )
    LPAREN expression_list_opt RPAREN;

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

// Keywords
KW_FILE_VERSION: 'FILE_VERSION';
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
// Boolean/Logical Keywords
KW_AND : 'AND';
KW_OR  : 'OR';
KW_NOT : 'NOT';
// Math Function Keywords
KW_LN   : 'LN';
KW_LOG  : 'LOG';
KW_SIN  : 'SIN';
KW_COS  : 'COS';
KW_TAN  : 'TAN';
KW_ASIN : 'ASIN';
KW_ACOS : 'ACOS';
KW_ATAN : 'ATAN';

// COMMENT_BLOCK handling corrected - capture within lexer rule and skip
COMMENT_BLOCK: KW_COMMENT_START .*? KW_ENDCOMMENT -> skip;

// Literals
NUMBER_LIT: [0-9]+ ('.' [0-9]+)?;
STRING_LIT:
    '"' (EscapeSequence | ~["\\\r\n])* '"'
    | '\'' (EscapeSequence | ~['\\\r\n])* '\'';

// Operators
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

// Punctuation
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

// Comparison
EQ: '==';
NEQ: '!=';
GT: '>';
LT: '<';
GTE: '>=';
LTE: '<=';

IDENTIFIER: [a-zA-Z_] [a-zA-Z0-9_]*;

// --- Comments and Whitespace ---
LINE_COMMENT: ('#' ~[!]? | '--') ~[\r\n]* -> skip; // Skip line comments
HASH_BANG: '#!' ~[\r\n]* -> skip; // Skip hashbang lines
NEWLINE: '\r'? '\n' | '\r'; // Handle different newline conventions
WS: [ \t]+ -> skip; // Skip whitespace

// Fragments (used within other lexer rules)
fragment EscapeSequence: '\\' (["'\\nrt] | UNICODE_ESC);
fragment UNICODE_ESC: 'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT;
fragment HEX_DIGIT: [0-9a-fA-F];