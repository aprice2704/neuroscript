// Grammar: NeuroScript Version: 0.4.5
// File version: 77
// Purpose: Fixes 'must' statement bug by allowing expressions as statements within function/command bodies, while respecting all architectural constraints.
grammar NeuroScript;

// --- LEXER RULES ---
LINE_ESCAPE_GLOBAL:
	'\\' ('\r'? '\n' | '\r') -> channel(HIDDEN);

// Keywords
KW_ACOS: 'acos';
KW_AND: 'and';
KW_AS: 'as';
KW_ASIN: 'asin';
KW_ASK: 'ask';
KW_ATAN: 'atan';
KW_BREAK: 'break';
KW_CALL: 'call';
KW_CLEAR: 'clear';
KW_CLEAR_ERROR: 'clear_error';
KW_COMMAND: 'command';
KW_CONTINUE: 'continue';
KW_COS: 'cos';
KW_DO: 'do';
KW_EACH: 'each';
KW_ELSE: 'else';
KW_EMIT: 'emit';
KW_ENDCOMMAND: 'endcommand';
KW_ENDFOR: 'endfor';
KW_ENDFUNC: 'endfunc';
KW_ENDIF: 'endif';
KW_ENDON: 'endon';
KW_ENDWHILE: 'endwhile';
KW_ERROR: 'error';
KW_EVAL: 'eval';
KW_EVENT: 'event';
KW_FAIL: 'fail';
KW_FALSE: 'false';
KW_FOR: 'for';
KW_FUNC: 'func';
KW_FUZZY: 'fuzzy';
KW_IF: 'if';
KW_IN: 'in';
KW_INTO: 'into';
KW_LAST: 'last';
KW_LEN: 'len';
KW_LN: 'ln';
KW_LOG: 'log';
KW_MEANS: 'means';
KW_MUST: 'must';
KW_MUSTBE: 'mustbe';
KW_NAMED: 'named';
KW_NEEDS: 'needs';
KW_NIL: 'nil';
KW_NO: 'no';
KW_NOT: 'not';
KW_ON: 'on';
KW_OPTIONAL: 'optional';
KW_OR: 'or';
KW_RETURN: 'return';
KW_RETURNS: 'returns';
KW_SET: 'set';
KW_SIN: 'sin';
KW_SOME: 'some';
KW_TAN: 'tan';
KW_TIMEDATE: 'timedate';
KW_TOOL: 'tool';
KW_TRUE: 'true';
KW_TYPEOF: 'typeof';
KW_WHILE: 'while';

// --- Other Tokens ---
fragment CONTINUED_LINE: '\\' ('\r'? '\n' | '\r');
fragment STRING_DQ_ATOM:
	EscapeSequence
	| CONTINUED_LINE
	| ~["\\\r\n];
fragment STRING_SQ_ATOM:
	EscapeSequence
	| CONTINUED_LINE
	| ~['\\\r\n];
STRING_LIT:
	'"' (STRING_DQ_ATOM)*? '"'
	| '\'' (STRING_SQ_ATOM)*? '\'';
TRIPLE_BACKTICK_STRING: '```' .*? '```';
fragment METADATA_CONTENT_ATOM:
	EscapeSequence
	| CONTINUED_LINE
	| ~[\\\r\n];
METADATA_LINE: [\t ]* '::' [ \t]+ (METADATA_CONTENT_ATOM)*;
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
LINE_COMMENT: ('#' | '--' | '//') ~[\r\n]* -> channel(HIDDEN);
NEWLINE: ('\r'? '\n' | '\r');
WS: [ \t]+ -> channel(HIDDEN);
fragment EscapeSequence:
	'\\' (UNICODE_ESC | HEX_ESC | OCTAL_ESC | CHAR_ESC);
fragment CHAR_ESC: ["'\\btnfrv~`];
fragment UNICODE_ESC:
	'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT;
fragment HEX_ESC: 'x' HEX_DIGIT HEX_DIGIT;
fragment OCTAL_ESC: [0-3]? [0-7] [0-7];
fragment HEX_DIGIT: [0-9a-fA-F];

// --- PARSER RULES ---

// Top-level rule enforces that a script is EITHER for libraries OR for commands.
program: file_header (library_script | command_script)? EOF;

file_header: (METADATA_LINE | NEWLINE)*;

// Script types - No change here, separation is preserved.
library_script: library_block+;
command_script: command_block+;

library_block: (procedure_definition | KW_ON event_handler) NEWLINE*;

// --- Command Block Definition ---
command_block:
	KW_COMMAND NEWLINE metadata_block command_statement_list KW_ENDCOMMAND NEWLINE*;

// A list that must contain at least one command_statement.
command_statement_list:
	(NEWLINE)* command_statement NEWLINE (command_body_line)*;

// Defines the restricted set of statements allowed inside a 'command' block.
command_body_line: command_statement NEWLINE | NEWLINE;

command_statement:
	simple_command_statement
	| block_statement
	| on_error_only_stmt;

// A restricted 'on' statement that only permits the error handler clause.
on_error_only_stmt: KW_ON error_handler;

// A restricted simple_statement that disallows 'return'.
simple_command_statement:
	set_statement
	| call_statement
	| emit_statement
	// FIX: 'must_statement' is removed. It's now handled by the 'expression_statement'.
	| fail_statement
	| clearEventStmt
	| ask_stmt
	| break_statement
	| continue_statement
	| expression_statement; // FIX: Added to handle 'must' and other expressions.

// --- Library (Function) Definitions ---
procedure_definition:
	KW_FUNC IDENTIFIER signature_part KW_MEANS NEWLINE metadata_block non_empty_statement_list
		KW_ENDFUNC;
signature_part:
	LPAREN (needs_clause | optional_clause | returns_clause)* RPAREN
	| (needs_clause | optional_clause | returns_clause)+
	| /* empty */;

needs_clause: KW_NEEDS param_list;
optional_clause: KW_OPTIONAL param_list;
returns_clause: KW_RETURNS param_list;
param_list: IDENTIFIER (COMMA IDENTIFIER)*;
metadata_block: (METADATA_LINE NEWLINE)*;

// --- General Statements & Lists ---

// A list that must contain at least one statement. For funcs and handlers.
non_empty_statement_list:
	(NEWLINE)* statement NEWLINE (body_line)*;

statement_list: body_line*;
body_line: statement NEWLINE | NEWLINE;

statement: simple_statement | block_statement | on_stmt;

// FIX: 'simple_statement' is now a list of true statements plus the new 'expression_statement'.
simple_statement:
	set_statement
	| call_statement
	| return_statement
	| emit_statement
	// FIX: 'must_statement' is removed. It's now handled by the 'expression_statement'.
	| fail_statement
	| clearErrorStmt
	| clearEventStmt
	| ask_stmt
	| break_statement
	| continue_statement
	| expression_statement; // FIX: Added to handle 'must' and other expressions.

// FIX: New rule for a statement that is just an expression.
// This rule ensures that a standalone 'must' or function call is a valid statement.
expression_statement: expression;

block_statement:
	if_statement
	| while_statement
	| for_each_statement;

on_stmt: KW_ON ( error_handler | event_handler);
error_handler:
	KW_ERROR KW_DO NEWLINE non_empty_statement_list KW_ENDON;
event_handler:
	KW_EVENT expression (KW_NAMED STRING_LIT)? (KW_AS IDENTIFIER)? KW_DO NEWLINE
		non_empty_statement_list KW_ENDON;

clearEventStmt:
	KW_CLEAR KW_EVENT (expression | KW_NAMED STRING_LIT);
lvalue: IDENTIFIER ( LBRACK expression RBRACK | DOT IDENTIFIER)*;
lvalue_list: lvalue (COMMA lvalue)*;
set_statement: KW_SET lvalue_list ASSIGN expression;
call_statement: KW_CALL callable_expr;
return_statement: KW_RETURN expression_list?;
emit_statement: KW_EMIT expression;
// 'must_statement' rule is now completely removed.
fail_statement: KW_FAIL expression?;
clearErrorStmt: KW_CLEAR_ERROR;
ask_stmt: KW_ASK expression (KW_INTO IDENTIFIER)?;
break_statement: KW_BREAK;
continue_statement: KW_CONTINUE;

if_statement:
	KW_IF expression NEWLINE non_empty_statement_list (
		KW_ELSE NEWLINE non_empty_statement_list
	)? KW_ENDIF;
while_statement:
	KW_WHILE expression NEWLINE non_empty_statement_list KW_ENDWHILE;
for_each_statement:
	KW_FOR KW_EACH IDENTIFIER KW_IN expression NEWLINE non_empty_statement_list KW_ENDFOR;

// --- Expression Rules ---
qualified_identifier: IDENTIFIER (DOT IDENTIFIER)*;
call_target: IDENTIFIER | KW_TOOL DOT qualified_identifier;
expression: logical_or_expr;
logical_or_expr: logical_and_expr (KW_OR logical_and_expr)*;
logical_and_expr: bitwise_or_expr (KW_AND bitwise_or_expr)*;
bitwise_or_expr: bitwise_xor_expr (PIPE bitwise_xor_expr)*;
bitwise_xor_expr: bitwise_and_expr (CARET bitwise_and_expr)*;
bitwise_and_expr: equality_expr (AMPERSAND equality_expr)*;
equality_expr: relational_expr ((EQ | NEQ) relational_expr)*;
relational_expr:
	additive_expr ((GT | LT | GTE | LTE) additive_expr)*;
additive_expr:
	multiplicative_expr ((PLUS | MINUS) multiplicative_expr)*;
multiplicative_expr:
	unary_expr ((STAR | SLASH | PERCENT) unary_expr)*;
unary_expr: (MINUS | KW_NOT | KW_NO | KW_SOME | TILDE | KW_MUST) unary_expr
	| KW_TYPEOF unary_expr
	| power_expr;
power_expr: accessor_expr (STAR_STAR power_expr)?;
accessor_expr: primary ( LBRACK expression RBRACK)*;
primary:
	literal
	| placeholder
	| IDENTIFIER
	| KW_LAST
	| callable_expr
	| KW_EVAL LPAREN expression RPAREN
	| LPAREN expression RPAREN;
callable_expr: (
		call_target
		| KW_LN
		| KW_LOG
		| KW_SIN
		| KW_COS
		| KW_TAN
		| KW_ASIN
		| KW_ACOS
		| KW_ATAN
		| KW_LEN
	) LPAREN expression_list_opt RPAREN;
placeholder:
	PLACEHOLDER_START (IDENTIFIER | KW_LAST) PLACEHOLDER_END;
literal:
	STRING_LIT
	| TRIPLE_BACKTICK_STRING
	| NUMBER_LIT
	| list_literal
	| map_literal
	| boolean_literal
	| nil_literal;
nil_literal: KW_NIL;
boolean_literal: KW_TRUE | KW_FALSE;
list_literal: LBRACK expression_list_opt RBRACK;
map_literal: LBRACE map_entry_list_opt RBRACE;
expression_list_opt: expression_list?;
expression_list: expression (COMMA expression)*;
map_entry_list_opt: map_entry_list?;
map_entry_list: map_entry (COMMA map_entry)*;
map_entry: STRING_LIT COLON expression;