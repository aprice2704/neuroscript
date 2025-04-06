// Corrected v4: pkg/neurodata/checklist/NeuroDataChecklist.g4
grammar NeuroDataChecklist;

// --- PARSER RULES ---

// Allows optional newlines before the first item, and between items.
// Consumes itemLines and the NEWLINEs that might follow them.
// Skipped comments/metadata are handled implicitly by the lexer.
checklistFile : NEWLINE* (itemLine NEWLINE*)* EOF;

// Define itemLine with explicit, flexible whitespace handling.
// Allows optional leading whitespace (WS*), requires space after HYPHEN (WS+)
// and RBRACK (WS+), allows optional space around MARK (WS*).
itemLine
    : WS* HYPHEN WS+ LBRACK WS* MARK WS* RBRACK WS+ TEXT
    ;
    // Optional NEWLINE removed here, handled by checklistFile rule


// --- LEXER RULES (Order Matters!) ---

// Skipped constructs first (comments, metadata)
METADATA_LINE : '#' ~[\r\n]* -> skip ;
COMMENT_LINE  : '--' ~[\r\n]* -> skip ;

// Tokens needed by parser ON DEFAULT CHANNEL
HYPHEN : '-';
LBRACK : '[';
RBRACK : ']';
MARK   : [xX ] ;         // Matches 'x', 'X', or space
TEXT   : ~[\r\n]+ ;      // Greedy match for rest of line content
NEWLINE: ( '\r'? '\n' | '\r' )+ ; // Group consecutive newlines
WS     : [ \t]+ ;         // CORRECTED: Capture whitespace, DO NOT SKIP