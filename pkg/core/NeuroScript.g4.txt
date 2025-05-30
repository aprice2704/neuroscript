// File:       NeuroScript.g4
// Grammar:    NeuroScript
// Version:    0.4.0 (Added nil, typeof, // comments, C-style string escapes)
// Date:       2025-05-29
// NOTES:
// - Added KW_NIL, nil_literal, and integrated into 'literal'.
// - Added KW_TYPEOF and integrated into 'unary_expr'.
// - Modified LINE_COMMENT to include '//'.
// - Modified STRING_LIT to correctly use EscapeSequence for C-style escapes.
// - Based on original v0.3.0 with qualified tool names.

grammar NeuroScript;

// --- PARSER RULES ---
program: file_header (procedure_definition NEWLINE*)* EOF;

file_header: (METADATA_LINE | NEWLINE)*;

procedure_definition: KW_FUNC IDENTIFIER signature_part KW_MEANS NEWLINE metadata_block? statement_list KW_ENDFUNC;

signature_part: LPAREN needs_clause? optional_clause? returns_clause? RPAREN | needs_clause optional_clause? returns_clause? | optional_clause returns_clause? | returns_clause | ;

needs_clause: KW_NEEDS param_list;
optional_clause: KW_OPTIONAL param_list;
returns_clause: KW_RETURNS param_list;
param_list: IDENTIFIER (COMMA IDENTIFIER)*;

metadata_block: (METADATA_LINE NEWLINE)*;

statement_list: body_line*;
body_line: statement NEWLINE | NEWLINE;

// --- STATEMENTS (MODIFIED from v0.3.0 base) ---
statement:
      simple_statement
    | block_statement
    ;

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

block_statement: if_statement | while_statement | for_each_statement | onErrorStmt;

// expressionStatement rule still exists but isn't directly used by 'statement'
expressionStatement: expression ;

// --- Simple Statements Details (MODIFIED from v0.3.0 base) ---
set_statement: KW_SET IDENTIFIER ASSIGN expression;
call_statement: KW_CALL callable_expr;
return_statement: KW_RETURN expression_list?;
emit_statement: KW_EMIT expression;
must_statement: KW_MUST expression | KW_MUSTBE callable_expr;
fail_statement: KW_FAIL expression?;
clearErrorStmt: KW_CLEAR_ERROR;
ask_stmt: KW_ASK expression (KW_INTO IDENTIFIER)? ;
break_statement: KW_BREAK;
continue_statement: KW_CONTINUE;

// (Rest of parser rules unchanged from v0.3.0 base, adapted for context)
if_statement: KW_IF expression NEWLINE statement_list (KW_ELSE NEWLINE statement_list)? KW_ENDIF;
while_statement: KW_WHILE expression NEWLINE statement_list KW_ENDWHILE;
for_each_statement: KW_FOR KW_EACH IDENTIFIER KW_IN expression NEWLINE statement_list KW_ENDFOR;
onErrorStmt: KW_ON_ERROR KW_MEANS NEWLINE statement_list KW_ENDON;

// --- MODIFIED FOR QUALIFIED TOOL NAMES (from v0.3.0 base) ---
qualified_identifier: IDENTIFIER (DOT IDENTIFIER)*;

call_target: IDENTIFIER // For user-defined functions
           | KW_TOOL DOT qualified_identifier // For tools, now using qualified_identifier
           ;
// --- END MODIFICATION for qualified tool names ---

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

// MODIFIED unary_expr to include KW_TYPEOF
unary_expr:
    (MINUS | KW_NOT | KW_NO | KW_SOME | TILDE) unary_expr
    | KW_TYPEOF unary_expr // <<< ADDED typeof operator
    | power_expr
    ;

power_expr: accessor_expr (STAR_STAR power_expr)?;
accessor_expr: primary ( LBRACK expression RBRACK )* ;

primary: literal | placeholder | IDENTIFIER | KW_LAST | callable_expr | KW_EVAL LPAREN expression RPAREN | LPAREN expression RPAREN;

callable_expr: ( call_target | KW_LN | KW_LOG | KW_SIN | KW_COS | KW_TAN | KW_ASIN | KW_ACOS | KW_ATAN ) LPAREN expression_list_opt RPAREN;
placeholder: PLACEHOLDER_START (IDENTIFIER | KW_LAST) PLACEHOLDER_END;

// MODIFIED literal to include nil_literal
literal: STRING_LIT | TRIPLE_BACKTICK_STRING | NUMBER_LIT | list_literal | map_literal | boolean_literal | nil_literal;

nil_literal: KW_NIL; // <<< ADDED nil literal rule

boolean_literal: KW_TRUE | KW_FALSE;
list_literal: LBRACK expression_list_opt RBRACK;
map_literal: LBRACE map_entry_list_opt RBRACE;
expression_list_opt: expression_list?;
expression_list: expression (COMMA expression)*;
map_entry_list_opt: map_entry_list?;
map_entry_list: map_entry (COMMA map_entry)*;
map_entry: STRING_LIT COLON expression;

// --- LEXER RULES ---
KW_CALL       : 'call';
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
KW_ASK        : 'ask';
KW_INTO       : 'into';
KW_BREAK      : 'break';
KW_CONTINUE   : 'continue';
KW_NIL        : 'nil';    // <<< ADDED KEYWORD
KW_TYPEOF     : 'typeof'; // <<< ADDED KEYWORD

NUMBER_LIT: [0-9]+ ('.' [0-9]+)?;

// MODIFIED STRING_LIT to use EscapeSequence for C-style escapes
STRING_LIT:
    '"' ( EscapeSequence | ~["\\\r\n] )*? '"'
    | '\'' ( EscapeSequence | ~['\\\r\n] )*? '\''
    ;

TRIPLE_BACKTICK_STRING: '```' .*? '```'; // Raw string, escapes not processed here

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
LBRACK: '[';
RBRACK: ']';
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

IDENTIFIER: [a-zA-Z_] [a-zA-Z0-9_]*;
METADATA_LINE: [\t ]* '::' [ \t]+ ~[\r\n]* ;

// MODIFIED LINE_COMMENT to include '//'
LINE_COMMENT: ('#' | '--' | '//') ~[\r\n]* -> channel(HIDDEN);
NEWLINE: ('\r'? '\n' | '\r');
WS: [ \t]+ -> channel(HIDDEN);

// EscapeSequence fragment: Used by STRING_LIT
// Supports: \", \', \\, \n, \r, \t, \~, \` and \uXXXX
// Note: \b, \f, \v, octal, and other hex escapes are not currently included.
fragment EscapeSequence: '\\' ( UNICODE_ESC | HEX_ESC | OCTAL_ESC | CHAR_ESC );

fragment CHAR_ESC: ["'\\btnfrv~`]; // Common character escapes including backslash itself

// Standard UNICODE_ESC as before
fragment UNICODE_ESC: 'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT;

// Optional: Add Hex and Octal escapes if desired, for now keeping it simpler.
// Example for basic hex (like \xHH), not fully implemented here to keep it concise.
fragment HEX_ESC: 'x' HEX_DIGIT HEX_DIGIT; // Example, adjust as needed
fragment OCTAL_ESC: [0-3]? [0-7] [0-7];   // Example, adjust as needed

fragment HEX_DIGIT: [0-9a-fA-F];