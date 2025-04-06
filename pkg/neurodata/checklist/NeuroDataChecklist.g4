// Revised: pkg/neurodata/checklist/NeuroDataChecklist.g4
grammar NeuroDataChecklist;

// --- PARSER RULES ---

// A checklist file consists of zero or more item lines followed by EOF.
// Metadata/comments/blank lines are skipped by the lexer.
checklistFile: itemLine* EOF;

// An item line specifically matches the checklist item structure.
itemLine: HYPHEN WS LBRACK MARK RBRACK WS TEXT NEWLINE?;

// --- LEXER RULES ---

// Item Markers (Parser uses these)
LBRACK : '[';
RBRACK : ']';
MARK   : [xX ] ; // Matches 'x', 'X', or space
HYPHEN : '-';

// TEXT: Capture the content after the checkbox marker until the newline
// Make it non-greedy to stop before the NEWLINE if present
TEXT   : ~[\r\n]+? ;

// Skipped Tokens (Send to hidden channel or skip entirely)
METADATA_LINE : ('#'|'--') WS? ('i' 'd' | 'v' 'e' 'r' 's' 'i' 'o' 'n' | 'r' 'e' 'n' 'd' 'e' 'r' 'i' 'n' 'g' '_' 'h' 'i' 'n' 't') ':' ~[\r\n]* -> skip ;
COMMENT_LINE  : ('#'|'--') ~[\r\n]* -> skip ; // Catch-all for other comment lines
NEWLINE       : ('\r'? '\n' | '\r');         // Don't skip newlines by default, parser needs them sometimes
WS            : [ \t]+ -> skip;              // Skip whitespace

// Fragment for keywords (optional, can simplify METADATA_LINE if preferred)
// fragment META_KEY: 'id' | 'version' | 'rendering_hint';