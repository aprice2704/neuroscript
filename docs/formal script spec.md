
 # NeuroScript: Formal specification
 
 Version: 0.2.0-alpha-func-returns-2
 DependsOn: pkg/core/NeuroScript.g4.txt
 HowToUpdate: Review pkg/core/NeuroScript.g4.txt and update this EBNF accordingly.
 
 The language is currently defined by pkg/core/NeuroScript.g4.txt. This document provides an EBNF-like representation based on that grammar.
 
 (* EBNF-like notation for NeuroScript Language *)
 
 :: Version: 0.2.0-alpha-func-returns-2
 ```ebnf
 
 (* --- Top Level --- *)
 program               ::= file_header ( procedure_definition newline* )* EOF ;
 file_header           ::= ( metadata_line | newline )* ;
 
 procedure_definition  ::= kw_func identifier signature_part kw_means newline metadata_block? statement_list kw_endfunc ;
 
 signature_part        ::= ( lparen needs_clause? optional_clause? returns_clause? rparen )
                       | ( needs_clause optional_clause? returns_clause? )
                       | ( optional_clause returns_clause? )
                       | returns_clause 
                       | (* empty *) ; (* Allows func MyProc() means ... or func MyProc means ... *)
 
 needs_clause          ::= kw_needs param_list ;
 optional_clause       ::= kw_optional param_list ;
 returns_clause        ::= kw_returns param_list ;
 param_list            ::= identifier ( comma identifier )* ;
 
 metadata_block        ::= ( metadata_line newline )* ;
 metadata_line         ::= '::' space+ text_until_newline ; (* Simplified - see lexer for exact parsing *)
 
 (* --- Statements --- *)
 statement_list        ::= body_line* ;
 body_line             ::= ( statement newline ) | newline ; (* Statement or blank line *)
 
 statement             ::= simple_statement | block_statement ;
 
 simple_statement      ::= set_statement
                       | return_statement
                       | emit_statement
                       | must_statement
                       | fail_statement
                       | clear_error_statement
                       | ask_statement ;
 
 block_statement       ::= if_statement
                       | while_statement
                       | for_each_statement
                       | on_error_statement ;
 
 (* --- Simple Statements Details --- *)
 set_statement         ::= kw_set identifier assign expression ;
 return_statement      ::= kw_return expression_list? ;
 emit_statement        ::= kw_emit expression ;
 must_statement        ::= kw_must expression | kw_mustbe callable_expr ;
 fail_statement        ::= kw_fail expression? ;
 clear_error_statement ::= kw_clear_error ;
 ask_statement         ::= kw_ask expression ( kw_into identifier )? ;
 
 (* --- Block Statements Details --- *)
 if_statement          ::= kw_if expression newline statement_list ( kw_else newline statement_list )? kw_endif ;
 while_statement       ::= kw_while expression newline statement_list kw_endwhile ;
 for_each_statement    ::= kw_for kw_each identifier kw_in expression newline statement_list kw_endfor ;
 on_error_statement    ::= kw_on_error kw_means newline statement_list kw_endon ;
 
 (* --- Expressions (Simplified Precedence - Follow G4 for exact rules) --- *)
 expression            ::= logical_or_expr ;
 logical_or_expr       ::= logical_and_expr ( kw_or logical_and_expr )* ;
 logical_and_expr      ::= bitwise_or_expr ( kw_and bitwise_or_expr )* ;
 bitwise_or_expr       ::= bitwise_xor_expr ( pipe bitwise_xor_expr )* ;
 bitwise_xor_expr      ::= bitwise_and_expr ( caret bitwise_and_expr )* ;
 bitwise_and_expr      ::= equality_expr ( ampersand equality_expr )* ;
 equality_expr         ::= relational_expr ( ( eq | neq ) relational_expr )* ;
 relational_expr       ::= additive_expr ( ( gt | lt | gte | lte ) additive_expr )* ;
 additive_expr         ::= multiplicative_expr ( ( plus | minus ) multiplicative_expr )* ;
 multiplicative_expr   ::= unary_expr ( ( star | slash | percent ) unary_expr )* ;
 
 unary_expr            ::= ( minus | kw_not | kw_no | kw_some | tilde ) unary_expr 
                       | power_expr ;
 
 power_expr            ::= accessor_expr ( star_star power_expr )? ;
 accessor_expr         ::= primary ( lbrack expression rbrack )* ;
 
 primary               ::= literal
                       | placeholder
                       | identifier
                       | kw_last
                       | callable_expr
                       | ( kw_eval lparen expression rparen )
                       | ( lparen expression rparen ) ;
 
 callable_expr         ::= ( call_target | built_in_function_keyword ) lparen expression_list_opt rparen ;
 
 call_target           ::= identifier | ( kw_tool dot identifier ) ;
 
 built_in_function_keyword ::= kw_ln | kw_log | kw_sin | kw_cos | kw_tan | kw_asin | kw_acos | kw_atan ;
 
 placeholder           ::= placeholder_start ( identifier | kw_last ) placeholder_end ;
 
 literal               ::= string_literal | triple_backtick_string | number_literal | list_literal | map_literal | boolean_literal ;
 boolean_literal       ::= kw_true | kw_false ;
 list_literal          ::= lbrack expression_list_opt rbrack ;
 map_literal           ::= lbrace map_entry_list_opt rbrace ;
 
 expression_list_opt   ::= expression_list? ;
 expression_list       ::= expression ( comma expression )* ;
 map_entry_list_opt    ::= map_entry_list? ;
 map_entry_list        ::= map_entry ( comma map_entry )* ;
 map_entry             ::= string_literal colon expression ;
 
 
 (* --- Terminals (Keywords & Operators) --- *)
 kw_func               ::= "func" ;
 kw_needs              ::= "needs" ;
 kw_optional           ::= "optional" ;
 kw_returns            ::= "returns" ;
 kw_means              ::= "means" ;
 kw_endfunc            ::= "endfunc" ;
 kw_set                ::= "set" ;
 kw_return             ::= "return" ;
 kw_emit               ::= "emit" ;
 kw_if                 ::= "if" ;
 kw_else               ::= "else" ;
 kw_endif              ::= "endif" ;
 kw_while              ::= "while" ;
 kw_endwhile           ::= "endwhile" ;
 kw_for                ::= "for" ;
 kw_each               ::= "each" ;
 kw_in                 ::= "in" ;
 kw_endfor             ::= "endfor" ;
 kw_on_error           ::= "on_error" ;
 kw_endon              ::= "endon" ;
 kw_clear_error        ::= "clear_error" ;
 kw_must               ::= "must" ;
 kw_mustbe             ::= "mustbe" ;
 kw_fail               ::= "fail" ;
 kw_no                 ::= "no" ;
 kw_some               ::= "some" ;
 kw_tool               ::= "tool" ;
 kw_last               ::= "last" ;
 kw_eval               ::= "eval" ;
 kw_true               ::= "true" ;
 kw_false              ::= "false" ;
 kw_and                ::= "and" ;
 kw_or                 ::= "or" ;
 kw_not                ::= "not" ;
 kw_ln                 ::= "ln" ;
 kw_log                ::= "log" ;
 kw_sin                ::= "sin" ;
 kw_cos                ::= "cos" ;
 kw_tan                ::= "tan" ;
 kw_asin               ::= "asin" ;
 kw_acos               ::= "acos" ;
 kw_atan               ::= "atan" ;
 kw_ask                ::= "ask" ;
 kw_into               ::= "into" ;
 
 number_literal        ::= digit+ ( "." digit+ )? ; (* Simplified *)
 string_literal        ::= '"' ( character | escape_sequence )* '"' 
                       | "'" ( character | escape_sequence )* "'" ; (* Simplified *)
 triple_backtick_string::= "```" .*? "```" ; (* Regex - not pure EBNF *)
 
 assign                ::= "=" ;
 plus                  ::= "+" ;
 minus                 ::= "-" ;
 star                  ::= "*" ;
 slash                 ::= "/" ;
 percent               ::= "%" ;
 star_star             ::= "**" ;
 ampersand             ::= "&" ;
 pipe                  ::= "|" ;
 caret                 ::= "^" ;
 tilde                 ::= "~" ; (* Added *)
 
 lparen                ::= "(" ;
 rparen                ::= ")" ;
 comma                 ::= "," ;
 lbrack                ::= "[" ;
 rbrack                ::= "]" ;
 lbrace                ::= "{" ;
 rbrace                ::= "}" ;
 colon                 ::= ":" ;
 dot                   ::= "." ;
 placeholder_start     ::= "{{" ;
 placeholder_end       ::= "}}" ;
 
 eq                    ::= "==" ;
 neq                   ::= "!=" ;
 gt                    ::= ">" ;
 lt                    ::= "<" ;
 gte                   ::= ">=" ;
 lte                   ::= "<=" ;
 
 identifier            ::= letter ( letter | digit | "_" )* ;
 
 (* --- Implicit Lexer Rules --- *)
 (* line_comment          ::= ( "#" | "--" ) text_until_newline ; (* Skipped *) *)
 newline               ::= "\\r"? "\\n" | "\\r" ;
 (* whitespace            ::= ( " " | "\\t" )+ ; (* Skipped *) *)
 (* letter                ::= "a".."z" | "A".."Z" | "_" ; *)
 (* digit                 ::= "0".."9" ; *)
 (* character             ::= (* Any character except quotes or backslash *) ; *)
 (* escape_sequence       ::= "\\" ( '"' | "'" | "\\" | "n" | "r" | "t" | unicode_escape ) ; *)
 (* unicode_escape        ::= "u" hex_digit hex_digit hex_digit hex_digit ; *)
 (* hex_digit             ::= digit | "a".."f" | "A".."F" ; *)
 (* text_until_newline    ::= [^\\n\\r]* ; *)
 
 ```