// pkg/core/NeuroScript.g4
grammar NeuroScript;

// --- Parser Rules (Adapted from .y file) --- version 2 

// Start rule: program consists of optional newlines, procedures, optional newlines
program: optional_newlines non_empty_procedure_list? optional_newlines EOF; // Use optional '?' for list, require EOF

// Procedure list structure
non_empty_procedure_list: procedure_definition (required_newlines procedure_definition)* ; // Use repetition '*'

// Helper for optional/required newlines
optional_newlines: NEWLINE* ; // Zero or more
required_newlines: NEWLINE+ ; // One or more

// Procedure definition - Note: Actions ({...}) are removed, handled later by Listener/Visitor
procedure_definition:
    KW_DEFINE KW_PROCEDURE IDENTIFIER LPAREN param_list_opt RPAREN NEWLINE comment_block statement_list KW_END NEWLINE;

param_list_opt: param_list? ; // Optional parameter list using '?'
param_list: IDENTIFIER (COMMA IDENTIFIER)* ; // List structure

// Comment block treated as a single token for now by the lexer rule
// *** NOTE: This now expects a COMMENT_BLOCK token from the lexer ***
comment_block: COMMENT_BLOCK;

// Statement list structure
statement_list: (statement)* ; // Zero or more statements

// Statement requires NEWLINE termination
statement: (simple_statement | block_statement) NEWLINE ; // Grouped alternatives

// Simple statements
simple_statement: set_statement | call_statement | return_statement ;

// Block statements
block_statement: if_statement | while_statement | for_each_statement ;

// Statement definitions
set_statement: KW_SET IDENTIFIER ASSIGN expression ;
call_statement: KW_CALL call_target LPAREN expression_list_opt RPAREN ;
return_statement: KW_RETURN expression? ; // Optional expression

// Block statement definitions
if_statement: KW_IF condition KW_THEN NEWLINE statement_list KW_END ;
// TODO: Add IF/ELSE rule later
while_statement: KW_WHILE condition KW_DO NEWLINE statement_list KW_END ;
for_each_statement: KW_FOR KW_EACH IDENTIFIER KW_IN expression KW_DO NEWLINE statement_list KW_END ;

// Call target variations
call_target: IDENTIFIER | KW_TOOL DOT IDENTIFIER | KW_LLM ;

// Condition structure
condition: expression ((EQ | NEQ | GT | LT | GTE | LTE) expression)? ; // Optional comparison

// Expression structure (Simplified - ANTLR handles precedence better if defined)
expression: term (PLUS term)* ; // Only string concat for now

term: literal
    | placeholder
    | IDENTIFIER          // Variable access
    | KW_LAST_CALL_RESULT
    | LPAREN expression RPAREN
    ;

placeholder: PLACEHOLDER_START IDENTIFIER PLACEHOLDER_END ;

literal: STRING_LIT
       | NUMBER_LIT
       | list_literal
       | map_literal
       ;

list_literal: LBRACK expression_list_opt RBRACK ;
map_literal: LBRACE map_entry_list_opt RBRACE ;

expression_list_opt: expression_list? ;
expression_list: expression (COMMA expression)* ;

map_entry_list_opt: map_entry_list? ;
map_entry_list: map_entry (COMMA map_entry)* ;

map_entry: STRING_LIT COLON expression ; // Key must be string literal


// --- Lexer Rules (Adapted from lexer.go and .y tokens) ---

// Keywords (Order matters - place before IDENTIFIER)
KW_DEFINE: 'DEFINE';
KW_PROCEDURE: 'PROCEDURE';
// *** Changed KW_COMMENT_START to KW_COMMENT ***
KW_COMMENT: 'COMMENT:';
KW_END: 'END';
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
KW_LAST_CALL_RESULT: '__last_call_result'; // Treat as keyword/reserved

// Literals
NUMBER_LIT : [0-9]+ ; // Simple integer for now
STRING_LIT : '"' (~["\\\r\n] | EscapeSequence)* '"'
           | '\'' (~['\\\r\n] | EscapeSequence)* '\''
           ;

// Operators and Delimiters
ASSIGN : '=';
PLUS   : '+';
LPAREN : '(';
RPAREN : ')';
COMMA  : ',';
LBRACK : '[';
RBRACK : ']';
LBRACE : '{';
RBRACE : '}';
COLON  : ':';
DOT    : '.';
PLACEHOLDER_START : '{{';
PLACEHOLDER_END   : '}}';
EQ     : '==';
NEQ    : '!=';
GT     : '>';
LT     : '<';
GTE    : '>=';
LTE    : '<=';

// Identifiers
IDENTIFIER : [a-zA-Z_] [a-zA-Z0-9_]* ; // Standard identifier pattern

// *** MODIFIED Comment Block Handling (No Modes) ***
// This matches the whole structure KW_COMMENT ... KW_END as one token.
// Content extraction needs to happen later. Non-greedy .*? is important.
COMMENT_BLOCK : KW_COMMENT .*? KW_END ; // Capture the whole block as one token
// Alternatively, capture it:
// COMMENT_BLOCK : KW_COMMENT .*? KW_END;
// Or more precisely (handling nested comments is hard without modes):
// COMMENT_BLOCK : KW_COMMENT (~[E] | 'E' ~[N] | 'EN' ~[D] )*? KW_END; // Tries to avoid matching inner END

// Single-line comments (Skipped)
LINE_COMMENT : ('#' | '--') ~[\r\n]* -> skip ;

// Newlines (Significant for parser)
NEWLINE : ( '\r'? '\n' | '\r' ) ; // Handle different line endings

// Whitespace (Skipped)
WS : [ \t]+ -> skip ;

// Line Continuation (Handled via lexer mode or pre-processing - simpler to skip for now)
// LINE_CONTINUATION: '\\' [\t ]* ('\r'? '\n' | '\r') -> skip; // Or potentially handle differently

// Error character
// INVALID : . ; // Catch any other character - ANTLR handles this implicitly


// Fragment for escape sequences used in STRING_LIT
fragment EscapeSequence : '\\' (["'\\trn] | UNICODE_ESC) ; // Basic escapes
fragment UNICODE_ESC : 'u' HEX_DIGIT HEX_DIGIT HEX_DIGIT HEX_DIGIT ; // Example
fragment HEX_DIGIT : [0-9a-fA-F] ;

// *** REMOVED Lexer Modes Section ***