// File:       NeuroScript.g4
// Grammar:    NeuroScript
// Version:    0.3.8 (Updated for qualified tool names)
// Date:       2025-05-09 // Assuming today's date for the change
// NOTE: Removed expressionStatement from statement rule.
//       Added explicit call_statement requiring 'call' keyword.
//       Lexer uses channel(HIDDEN) for WS/LINE_COMMENT.
//       MODIFIED: call_target for tools to use qualified_identifier.

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

// --- STATEMENTS (MODIFIED) ---
statement:
      simple_statement
    | block_statement
   // | expressionStatement // <<< REMOVED expressionStatement as a standalone statement
    ;

simple_statement:
    set_statement
    | call_statement // <<< ADDED BACK
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

// --- Simple Statements Details (MODIFIED) ---
set_statement: KW_SET IDENTIFIER ASSIGN expression;
call_statement: KW_CALL callable_expr; // <<< Requires 'call' keyword now
return_statement: KW_RETURN expression_list?;
emit_statement: KW_EMIT expression;
must_statement: KW_MUST expression | KW_MUSTBE callable_expr;
fail_statement: KW_FAIL expression?;
clearErrorStmt: KW_CLEAR_ERROR;
ask_stmt: KW_ASK expression (KW_INTO IDENTIFIER)? ;
break_statement: KW_BREAK;
continue_statement: KW_CONTINUE;

// (Rest of parser rules unchanged from v0.3.6, adapted for v0.3.7 base)
if_statement: KW_IF expression NEWLINE statement_list (KW_ELSE NEWLINE statement_list)? KW_ENDIF;
while_statement: KW_WHILE expression NEWLINE statement_list KW_ENDWHILE;
for_each_statement: KW_FOR KW_EACH IDENTIFIER KW_IN expression NEWLINE statement_list KW_ENDFOR;
onErrorStmt: KW_ON_ERROR KW_MEANS NEWLINE statement_list KW_ENDON;

// --- MODIFIED FOR QUALIFIED TOOL NAMES ---
// Rule for a multi-part identifier, e.g., Group.Name or Group.SubGroup.Name
qualified_identifier: IDENTIFIER (DOT IDENTIFIER)*;

// Updated call_target
call_target: IDENTIFIER // For user-defined functions
           | KW_TOOL DOT qualified_identifier // For tools, now using qualified_identifier
           ;
// --- END MODIFICATION ---

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
unary_expr: (MINUS | KW_NOT | KW_NO | KW_SOME | TILDE) unary_expr | power_expr;
power_expr: accessor_expr (STAR_STAR power_expr)?;
accessor_expr: primary ( LBRACK expression RBRACK )* ;

// NOTE: callable_expr is still allowed within primary, so calls *within* expressions like set x = myFunc() + 1 still work fine.
primary: literal | placeholder | IDENTIFIER | KW_LAST | callable_expr | KW_EVAL LPAREN expression RPAREN | LPAREN expression RPAREN;
callable_expr: ( call_target | KW_LN | KW_LOG | KW_SIN | KW_COS | KW_TAN | KW_ASIN | KW_ACOS | KW_ATAN ) LPAREN expression_list_opt RPAREN;
placeholder: PLACEHOLDER_START (IDENTIFIER | KW_LAST) PLACEHOLDER_END;
literal: STRING_LIT | TRIPLE_BACKTICK_STRING | NUMBER_LIT | list_literal | map_literal | boolean_literal;
boolean_literal: KW_TRUE | KW_FALSE;
list_literal: LBRACK expression_list_opt RBRACK;
map_literal: LBRACE map_entry_list_opt RBRACE;
expression_list_opt: expression_list?;
expression_list: expression (COMMA expression)*;
map_entry_list_opt: map_entry_list?;
map_entry_list: map_entry (COMMA map_entry)*;
map_entry: STRING_LIT COLON expression;

// --- LEXER RULES ---
KW_CALL       : 'call'; // <<< ADDED KEYWORD
KW_FUNC : 'func'; KW_NEEDS : 'needs'; KW_OPTIONAL : 'optional'; KW_RETURNS : 'returns'; KW_MEANS : 'means';
KW_ENDFUNC : 'endfunc'; KW_SET : 'set'; KW_RETURN : 'return'; KW_EMIT : 'emit'; KW_IF : 'if'; KW_ELSE : 'else';
KW_ENDIF : 'endif'; KW_WHILE : 'while'; KW_ENDWHILE : 'endwhile'; KW_FOR : 'for'; KW_EACH : 'each'; KW_IN : 'in';
KW_ENDFOR : 'endfor'; KW_ON_ERROR : 'on_error'; KW_ENDON : 'endon'; KW_CLEAR_ERROR: 'clear_error'; KW_MUST : 'must'; KW_MUSTBE : 'mustbe'; KW_FAIL : 'fail';
KW_NO : 'no'; KW_SOME : 'some'; KW_TOOL : 'tool'; KW_LAST : 'last'; KW_EVAL : 'eval'; KW_TRUE : 'true';
KW_FALSE : 'false'; KW_AND : 'and'; KW_OR : 'or'; KW_NOT : 'not'; KW_LN : 'ln'; KW_LOG : 'log';
KW_SIN : 'sin'; KW_COS : 'cos'; KW_TAN : 'tan'; KW_ASIN : 'asin'; KW_ACOS : 'acos'; KW_ATAN : 'atan';
KW_ASK : 'ask'; KW_INTO : 'into'; KW_BREAK : 'break'; KW_CONTINUE : 'continue';

NUMBER_LIT: [0-9]+ ('.' [0-9]+)?;
STRING_LIT: '"' ( '\\"' | ~["\\] )*? '"' | '\'' ( '\\\'' | ~['\\] )*? '\'' ;
TRIPLE_BACKTICK_STRING: '```' .*? '```';

ASSIGN: '='; PLUS: '+'; MINUS: '-'; STAR: '*'; SLASH: '/'; PERCENT: '%'; STAR_STAR: '**'; AMPERSAND: '&'; PIPE: '|';
CARET: '^'; TILDE: '~';

LPAREN: '('; RPAREN: ')'; COMMA: ','; LBRACK: '['; RBRACK: ']'; LBRACE: '{'; RBRACE: '}'; COLON: ':';
DOT: '.'; PLACEHOLDER_START: '{{'; PLACEHOLDER_END: '}}';

EQ: '=='; NEQ: '!='; GT: '>'; LT: '<'; GTE: '>='; LTE: '<=';

IDENTIFIER: [a-zA-Z_] [a-zA-Z0-9_]*;
METADATA_LINE: [\t ]* '::' [ \t]+ ~[\r\n]* ; // Updated to allow leading tabs or spaces
LINE_COMMENT: ('#' | '--') ~[\r\n]* -> channel(HIDDEN); // Keep HIDDEN channel
NEWLINE: ('\r'? '\n' | '\r');
WS: [ \t]+ -> channel(HIDDEN); // Keep HIDDEN channel

fragment EscapeSequence: '\\' (["'\\nrt~] | UNICODE_ESC | '`');
fragment UNICODE_ESC: 'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT;
fragment HEX_DIGIT: [0-9a-fA-F];