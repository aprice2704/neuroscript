# NeuroScript: Formal specification

Version: 0.1.0
DependsOn: pkg/core/NeuroScript.g4
HowToUpdate: Read the file pkg/core/NeuroScript.g4 and deduce the correct EBNF from it.

The language is currently defined by pkg/core/NeuroScript.g4. At some point, this file will come to be the offical definition, but not yet.


```ebnf
(* NeuroScript Grammar (EBNF-like) *)
(* Version reflecting blocks, line continuation, list/map literals *)

(* Top Level *)
script_file        ::= { procedure_definition | comment_line | blank_line } ;

procedure_definition ::= "DEFINE" "PROCEDURE" identifier "(" [ param_list ] ")" newline
                         comment_block
                         statement_list
                         "END" newline ;

param_list         ::= identifier { "," identifier } ;

comment_block      ::= "COMMENT:" newline { text_line } "END" newline ; (* Content structure (PURPOSE:, etc.) is semantic *)

statement_list     ::= { statement } ;

(* Statements *)
(* Note: Line continuation '\' handled by pre-processor joining lines before parsing *)
statement          ::= ( set_statement | call_statement | return_statement | if_statement | while_statement | for_each_statement | comment_line | blank_line ) newline ;

set_statement      ::= "SET" identifier "=" expression ;

call_statement     ::= "CALL" call_target "(" [ expression_list ] ")" ;

return_statement   ::= "RETURN" [ expression ] ;

if_statement       ::= "IF" condition "THEN" newline statement_list "END" ;
(* TODO: Add ELSE block support *)

while_statement    ::= "WHILE" condition "DO" newline statement_list "END" ;

for_each_statement ::= "FOR" "EACH" identifier "IN" expression "DO" newline statement_list "END" ;
(* Interpreter TODO: Iterate lists, maps, string characters *)
(* TODO: Define map iteration specifics (keys? key-value?) *)

(* Expressions and Conditions *)
expression         ::= term { "+" term } ; (* TODO: Implement full arithmetic/boolean operators; Define list/map concat *)

term               ::= literal | placeholder | identifier | last_call_result | "(" expression ")" ;

condition          ::= expression [ ("==" | "!=") expression ] ; (* TODO: Implement >, <, >=, <= *)

expression_list    ::= expression { "," expression } ; (* Used in CALL and list_literal *)

(* Literals *)
literal            ::= string_literal | list_literal | map_literal
                   | numeric_literal (* TODO *) | boolean_literal (* TODO *) ;

string_literal     ::= '"' { character | escape_sequence } '"'
                   | "'" { character | escape_sequence } "'" ;

list_literal       ::= "[" [ expression_list ] "]" ; (* NEW *)

map_literal        ::= "{" [ map_entry_list ] "}" ; (* NEW *)

map_entry_list     ::= map_entry { "," map_entry } ; (* NEW *)

map_entry          ::= string_literal ":" expression ; (* NEW - Keys must be string literals *)

numeric_literal    ::= (* TODO: Define digits, optional decimal point *) ;
boolean_literal    ::= "true" | "false" (* TODO: Define as```

**Notes on this EBNF:**

* **Line Continuation:** The grammar assumes line continuation (`\`) is handled *before* this syntax is applied, joining physical lines into logical lines.
* **Whitespace/Newlines:** Whitespace is generally implied between tokens. `newline` is explicitly mentioned where significant (e.g., after `THEN`/`DO`, terminating lines). `blank_line` allows empty lines.
* **Comments:** Comments (`#` or `--`) are treated like whitespace – they can generally appear anywhere whitespace is allowed and are ignored by the structural parsing.
* **Docstring Content:** The `comment_block` syntax doesn't enforce the internal structure (`PURPOSE:`, etc.). This is treated as a semantic requirement on the `text_line` content.
* **Expressions/Conditions:** These are simplified. A real implementation would need operator precedence, more operators (arithmetic, boolean, comparison), and potentially function calls within expressions if desired.
* **Literals:** `numeric_literal` and `boolean_literal` definitions are placeholders. Escape sequences in strings are basic.
* **Keywords:** Identifiers cannot be the same as reserved keywords (case-insensitive check recommended).
* **TODOs:** Explicitly marked features from the spec that aren't fully defined in this grammar or implemented yet.

This EBNF provides a more formal definition of the NeuroScript syntax based on our progress.