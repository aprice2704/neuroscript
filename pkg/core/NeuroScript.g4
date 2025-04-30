// File:      NeuroScript.g4
// Grammar:   NeuroScript
// Version:   0.2.0-alpha-onerror-fix-4 // Incremented Version
// Date:      2025-04-29

grammar NeuroScript;

// --- PARSER RULES ---

// MODIFIED: Allow optional NEWLINEs between procedure definitions
program: file_header (procedure_definition NEWLINE*)* EOF;

file_header: (METADATA_LINE | NEWLINE)*;

// Procedure definition uses optional parameter clauses block
procedure_definition:
    KW_FUNC IDENTIFIER
    parameter_clauses? // The whole clause group is optional
    KW_MEANS NEWLINE
    metadata_block?    // Optional metadata block after 'means'
    statement_list
    KW_ENDFUNC; // REMOVED optional NEWLINE here - let program rule handle it

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

// --- Statements (unchanged) ---
statement: simple_statement | block_statement ;

simple_statement:
    set_statement
    | call_statement
    | return_statement
    | emit_statement
    | must_statement
    | fail_statement
    | clearErrorStmt;

block_statement:
    if_statement
    | while_statement
    | for_each_statement
    | onErrorStmt;

set_statement: KW_SET IDENTIFIER ASSIGN expression;
call_statement: KW_CALL call_target LPAREN expression_list_opt RPAREN;
return_statement: KW_RETURN expression_list?;
emit_statement: KW_EMIT expression;
must_statement: KW_MUST expression | KW_MUSTBE function_call;
fail_statement: KW_FAIL expression?; // Allow optional expression for fail
clearErrorStmt: KW_CLEAR_ERROR;

if_statement:
    KW_IF expression NEWLINE
    statement_list
    (KW_ELSE NEWLINE statement_list)?
    KW_ENDIF; // Removed NEWLINE - handled by body_line
while_statement:
    KW_WHILE expression NEWLINE
    statement_list
    KW_ENDWHILE; // Removed NEWLINE - handled by body_line
for_each_statement:
    KW_FOR KW_EACH IDENTIFIER KW_IN expression NEWLINE
    statement_list
    KW_ENDFOR; // Removed NEWLINE - handled by body_line
onErrorStmt:
    KW_ON_ERROR KW_MEANS NEWLINE
    statement_list
    KW_ENDON; // Removed NEWLINE - handled by body_line

call_target: IDENTIFIER | KW_TOOL DOT IDENTIFIER;

// --- Expression Rules with Precedence --- (Unchanged)
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
primary:
    literal
    | placeholder
    | IDENTIFIER
    | KW_LAST             // Access last implicit result
    | function_call
    | KW_EVAL LPAREN expression RPAREN // Explicit eval
    | LPAREN expression RPAREN;        // Grouping
function_call:
    ( IDENTIFIER // User functions
    | KW_LN | KW_LOG | KW_SIN | KW_COS | KW_TAN | KW_ASIN | KW_ACOS | KW_ATAN // Built-ins
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
KW_CALL       : 'call';
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

// Literals
NUMBER_LIT: [0-9]+ ('.' [0-9]+)?;
STRING_LIT:
    '"' (EscapeSequence | ~["\\\r\n])* '"'
    | '\'' (EscapeSequence | ~['\\\r\n])* '\'';
TRIPLE_BACKTICK_STRING: '```' .*? '```'; // Non-greedy match inside

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

IDENTIFIER: [a-zA-Z_] [a-zA-Z0-9_]*; // Standard identifier

// --- Comments, Metadata, and Whitespace ---

// Metadata line: consumes content, leaves newline for parser rule
METADATA_LINE: [\t ]* '::' [ \t]+ ~[\r\n]* ;

// Line comments start with # or --, skipped entirely
LINE_COMMENT: ('#' | '--') ~[\r\n]* -> skip;

// Whitespace and Newlines
NEWLINE: '\r'? '\n' | '\r'; // Handle different line endings (Default Channel)
WS: [ \t]+ -> skip;        // Skip spaces and tabs

// Fragments used in Lexer rules
fragment EscapeSequence: '\\' (["'\\nrt] | UNICODE_ESC | '`'); // Allow escaped backtick in strings
fragment UNICODE_ESC: 'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT;
fragment HEX_DIGIT: [0-9a-fA-F];