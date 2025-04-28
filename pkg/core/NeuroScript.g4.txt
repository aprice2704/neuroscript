// File:     NeuroScript.g4
// Grammar: NeuroScript
// Version: 0.2.0
// Date:    2025-04-27 // Updated date

grammar NeuroScript;

// --- PARSER RULES ---

program: optional_newlines file_version_decl? optional_newlines procedure_definition* optional_newlines EOF;

optional_newlines: NEWLINE*;

file_version_decl: KW_FILE_VERSION STRING_LIT NEWLINE;

// Procedure Definition (v0.2.0)
procedure_definition:
    KW_FUNC IDENTIFIER
    needs_clause?
    optional_clause?
    returns_clause?
    KW_MEANS NEWLINE
    metadata_block? // Use the metadata_block rule here
    statement_list
    KW_ENDFUNC NEWLINE?;

needs_clause: KW_NEEDS param_list;
optional_clause: KW_OPTIONAL param_list;
returns_clause: KW_RETURNS param_list;

param_list: IDENTIFIER (COMMA IDENTIFIER)*;

// Metadata block collects one or more metadata lines
metadata_block: METADATA_LINE+;

statement_list: body_line*;

// FIX: Removed reference to metadata_line_inline from body_line
body_line: statement NEWLINE | NEWLINE; // Allow statements or blank lines

// FIX: Removed metadata_line_inline from statement alternatives
statement: simple_statement | block_statement ;

// FIX: Removed metadata_line_inline rule completely for now
// metadata_line_inline : METADATA_LINE;

simple_statement:
    set_statement
    | call_statement
    | return_statement // Modified rule referenced below
    | emit_statement
    | must_statement
    | fail_statement;

block_statement:
    if_statement
    | while_statement
    | for_each_statement
    | try_statement;

// Simple Statements (Keywords Lowercase)
set_statement: KW_SET IDENTIFIER ASSIGN expression;
call_statement: KW_CALL call_target LPAREN expression_list_opt RPAREN;
// FIX: Use expression_list? to allow zero, one, or more return expressions
return_statement: KW_RETURN expression_list?;
emit_statement: KW_EMIT expression;
must_statement: KW_MUST expression | KW_MUSTBE function_call;
fail_statement: KW_FAIL expression?;

// Block Statements (Keywords Lowercase, specific terminators)
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
try_statement:
    KW_TRY NEWLINE
    try_body=statement_list
    (KW_CATCH catch_param=IDENTIFIER? NEWLINE catch_body=statement_list)*
    (KW_FINALLY NEWLINE finally_body=statement_list)?
    KW_ENDTRY;

// Call Target
call_target: IDENTIFIER | KW_TOOL DOT IDENTIFIER;

// --- Expression Rules with Precedence ---
expression: logical_or_expr;
// ... (Rest of expression rules remain the same) ...
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
    primary ( LBRACK expression RBRACK )* ;
primary:
    literal
    | placeholder
    | IDENTIFIER
    | KW_LAST
    | function_call
    | KW_EVAL LPAREN expression RPAREN
    | LPAREN expression RPAREN;
function_call:
    ( IDENTIFIER
    | KW_LN | KW_LOG | KW_SIN | KW_COS | KW_TAN | KW_ASIN | KW_ACOS | KW_ATAN
    )
    LPAREN expression_list_opt RPAREN;
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
map_entry: STRING_LIT COLON expression;

// --- LEXER RULES ---

// Keywords (All Lowercase)
KW_FILE_VERSION: 'file_version';
KW_FUNC        : 'func';
KW_NEEDS       : 'needs';
KW_OPTIONAL    : 'optional';
KW_RETURNS     : 'returns';
KW_MEANS       : 'means';
KW_ENDFUNC     : 'endfunc';
KW_SET         : 'set';
KW_CALL        : 'call';
KW_RETURN      : 'return';
KW_EMIT        : 'emit';
KW_IF          : 'if';
KW_ELSE        : 'else';
KW_ENDIF       : 'endif';
KW_WHILE       : 'while';
KW_ENDWHILE    : 'endwhile';
KW_FOR         : 'for';
KW_EACH        : 'each';
KW_IN          : 'in';
KW_ENDFOR      : 'endfor';
KW_TRY         : 'try';
KW_CATCH       : 'catch';
KW_FINALLY     : 'finally';
KW_ENDTRY      : 'endtry';
KW_MUST        : 'must';
KW_MUSTBE      : 'mustbe';
KW_FAIL        : 'fail';
KW_NO          : 'no';
KW_SOME        : 'some';
KW_TOOL        : 'tool';
KW_LAST        : 'last';
KW_EVAL        : 'eval';
KW_TRUE        : 'true';
KW_FALSE       : 'false';
KW_AND         : 'and';
KW_OR          : 'or';
KW_NOT         : 'not';
KW_LN          : 'ln';
KW_LOG         : 'log';
KW_SIN         : 'sin';
KW_COS         : 'cos';
KW_TAN         : 'tan';
KW_ASIN        : 'asin';
KW_ACOS        : 'acos';
KW_ATAN        : 'atan';

// Literals
NUMBER_LIT: [0-9]+ ('.' [0-9]+)?;
STRING_LIT:
    '"' (EscapeSequence | ~["\\\r\n])* '"'
    | '\'' (EscapeSequence | ~['\\\r\n])* '\'';
TRIPLE_BACKTICK_STRING: '```' .*? '```';

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

// --- Comments, Metadata, and Whitespace ---

// Send whole content after ':: ' to parser
METADATA_LINE: [\t ]* '::' [ \t]+ .*? ;

// Line comments start with # or --
LINE_COMMENT: ('#' | '--') ~[\r\n]* -> skip;

// Whitespace and Newlines
NEWLINE: '\r'? '\n' | '\r'; // Default Channel
WS: [ \t]+ -> skip;       // Skip whitespace

// Fragments
fragment EscapeSequence: '\\' (["'\\nrt] | UNICODE_ESC | '`');
fragment UNICODE_ESC: 'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT;
fragment HEX_DIGIT: [0-9a-fA-F];