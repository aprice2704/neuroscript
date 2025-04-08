// pkg/neurodata/blocks/generated/FencedBlockExtractor.g4
grammar FencedBlockExtractor;

// --- Tokens ---
FENCE_MARKER: '```' ;
LANG_ID:      [a-zA-Z0-9_.-]+ ;

// *** Tokens now on DEFAULT channel (no -> channel(HIDDEN)) ***
METADATA_LINE: '::' [ \t]+ ~[\r\n]* ;
LINE_COMMENT:  ('#' | '--') ~[\r\n]* ;
WS:            [ \t]+ ; // Keep whitespace visible too

// Match newlines explicitly
NEWLINE:       ( '\r'? '\n' | '\r' )+ ;

// Capture any other non-whitespace, non-special character
// Adjusted to avoid consuming parts of other tokens accidentally
OTHER_TEXT:    ~[`#:\r\n\t ]+ ;


// --- Parser Rules (Minimal) ---
document: token* EOF;

// Include all relevant tokens
token: FENCE_MARKER | LANG_ID | NEWLINE | OTHER_TEXT | METADATA_LINE | LINE_COMMENT | WS ;