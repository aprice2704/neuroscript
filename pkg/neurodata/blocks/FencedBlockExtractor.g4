// pkg/neurodata/blocks2/FencedBlockExtractor.g4
grammar FencedBlockExtractor;

/*
 Purpose: This grammar is designed to TOKENIZE a text document,
          identifying potential fenced code block markers ('```'),
          language identifiers immediately following an opening fence,
          newlines, line comments, whitespace, and any other text content.

 Usage:   The Go code using this parser's output (the token stream)
          will be responsible for:
          1. Reconstructing the raw content within blocks (including whitespace).
          2. Handling nested fences using a counter.
          3. Detecting ambiguous fences (e.g., ``` immediately after a closing ```).
          4. Associating LANG_ID tokens with the correct FENCE_MARKER.
*/

// --- Parser Rule (Minimal) ---
// The primary goal is lexing. This rule just ensures all tokens are consumed.
document: token* EOF;

// Defines the possible tokens the parser might see.
token: FENCE_MARKER | LANG_ID | NEWLINE | OTHER_TEXT | LINE_COMMENT | WS ;


// --- Lexer Rules (Order is important!) ---

// Matches the triple backtick sequence used for fences.
FENCE_MARKER     : '```' ;

// Matches typical language identifiers that might follow an opening fence.
// Allows letters, numbers, underscore, plus, minus, dot, hash.
LANG_ID          : [a-zA-Z0-9_+.#-]+ ;

// Matches common line comment styles (# or --) including the trailing newline if present.
// Using .*? makes it non-greedy so it stops before \r or \n
LINE_COMMENT     : ('#'|'--') .*? NEWLINE? ; // Consume comment and optional newline

// Matches one or more newline characters. Consolidates consecutive newlines.
NEWLINE          : [\r\n]+ ;

// Captures whitespace (spaces, tabs) - NO LONGER SKIPPED.
WS               : [ \t]+ ; // Capture whitespace on default channel

// Matches any other single character not captured by the rules above.
// This acts as a fallback to capture all content text.
OTHER_TEXT       : . ;