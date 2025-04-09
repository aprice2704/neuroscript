# NeuroScript: Formal specification

Version: 0.2.0
DependsOn: pkg/core/NeuroScript.g4
HowToUpdate: Review pkg/core/NeuroScript.g4 and update this EBNF accordingly.

The language is currently defined by pkg/core/NeuroScript.g4. At some point, this file will come to be the offical definition, but not yet.

(* EBNF-like notation for NeuroScript Language )


:: Version: 0.2.0
```ebnf

( --- Top Level --- )
program               ::= { newline } [ file_version_decl ] { newline } { procedure_definition } { newline } EOF ;
optional_newlines     ::= { newline } ; ( Used implicitly by parser/lexer where appropriate )
file_version_decl   ::= "FILE_VERSION" string_literal newline ;

procedure_definition ::= "DEFINE" "PROCEDURE" identifier
                         "(" [ param_list ] ")" newline
                         [ comment_block_marker ] ( Handled by lexer skip rule )
                         statement_list
                         "END" [ newline ] ;

param_list_opt        ::= [ param_list ] ; ( Optional parameter list )
param_list            ::= identifier { "," identifier } ;

( --- Statements --- )
statement_list        ::= { body_line } ;
body_line             ::= ( statement newline ) | newline ; ( Statement or blank line )

statement             ::= simple_statement | block_statement ;

simple_statement      ::= set_statement
                      | call_statement
                      | return_statement
                      | emit_statement ;

block_statement       ::= if_statement
                      | while_statement
                      | for_each_statement ;

set_statement         ::= "SET" identifier "=" expression ;
call_statement        ::= "CALL" call_target "(" [ expression_list ] ")" ;
return_statement      ::= "RETURN" [ expression ] ;
emit_statement        ::= "EMIT" expression ;

( --- Block Statements --- )
if_statement          ::= "IF" expression "THEN" newline
                          if_body=statement_list
                          [ "ELSE" newline else_body=statement_list ]
                          "ENDBLOCK" ;

while_statement       ::= "WHILE" expression "DO" newline
                          statement_list
                          "ENDBLOCK" ;

for_each_statement    ::= "FOR" "EACH" identifier "IN" expression "DO" newline
                          statement_list
                          "ENDBLOCK" ;

call_target           ::= identifier | ( "TOOL" "." identifier ) | "LLM" ;

( --- Expressions (Operator Precedence Defined by Order) --- )
expression            ::= logical_or_expr ; ( Entry point )

logical_or_expr       ::= logical_and_expr { "OR" logical_and_expr } ;
logical_and_expr      ::= bitwise_or_expr { "AND" bitwise_or_expr } ;
bitwise_or_expr       ::= bitwise_xor_expr { "|" bitwise_xor_expr } ;
bitwise_xor_expr      ::= bitwise_and_expr { "^" bitwise_and_expr } ;
bitwise_and_expr      ::= equality_expr { "&" equality_expr } ;
equality_expr         ::= relational_expr { ( "==" | "!=" ) relational_expr } ;
relational_expr       ::= additive_expr { ( ">" | "<" | ">=" | "<=" ) additive_expr } ;
additive_expr         ::= multiplicative_expr { ( "+" | "-" ) multiplicative_expr } ;
multiplicative_expr   ::= unary_expr { ( "" | "/" | "%" ) unary_expr } ;

unary_expr            ::= ( "-" | "NOT" ) unary_expr
                      | power_expr ;

power_expr            ::= accessor_expr [ "**" power_expr ] ; (* Right-associative via recursion )

accessor_expr         ::= primary { "[" expression "]" } ; ( Handles element access )

primary               ::= literal
                      | placeholder
                      | identifier
                      | "LAST"
                      | function_call
                      | "EVAL" "(" expression ")"
                      | "(" expression ")" ;

function_call         ::= ( "LN" | "LOG" | "SIN" | "COS" | "TAN" | "ASIN" | "ACOS" | "ATAN" )
                          "(" [ expression_list ] ")" ;

placeholder           ::= "{{" ( identifier | "LAST" ) "}}" ;

( --- Literals --- )
literal               ::= string_literal
                      | number_literal
                      | list_literal
                      | map_literal
                      | boolean_literal ;

boolean_literal       ::= "true" | "false" ;
list_literal          ::= "[" [ expression_list ] "]" ;
map_literal           ::= "{" [ map_entry_list ] "}" ;

expression_list_opt   ::= [ expression_list ] ;
expression_list       ::= expression { "," expression } ;

map_entry_list_opt    ::= [ map_entry_list ] ;
map_entry_list        ::= map_entry { "," map_entry } ;
map_entry             ::= string_literal ":" expression ;

( --- Lexer Tokens (Conceptual EBNF) --- )
file_version_keyword  ::= "FILE_VERSION" ;
define_keyword        ::= "DEFINE" ;
procedure_keyword     ::= "PROCEDURE" ;
end_keyword           ::= "END" ;
endblock_keyword      ::= "ENDBLOCK" ;
comment_start_keyword ::= "COMMENT:" ; ( Start of COMMENT_BLOCK marker )
endcomment_keyword    ::= "ENDCOMMENT" ; ( End of COMMENT_BLOCK marker )
set_keyword           ::= "SET" ;
call_keyword          ::= "CALL" ;
return_keyword        ::= "RETURN" ;
if_keyword            ::= "IF" ;
then_keyword          ::= "THEN" ;
else_keyword          ::= "ELSE" ;
while_keyword         ::= "WHILE" ;
do_keyword            ::= "DO" ;
for_keyword           ::= "FOR" ;
each_keyword          ::= "EACH" ;
in_keyword            ::= "IN" ;
tool_keyword          ::= "TOOL" ;
llm_keyword           ::= "LLM" ;
last_keyword          ::= "LAST" ;
eval_keyword          ::= "EVAL" ;
emit_keyword          ::= "EMIT" ;
true_keyword          ::= "true" ;
false_keyword         ::= "false" ;
and_keyword           ::= "AND" ;
or_keyword            ::= "OR" ;
not_keyword           ::= "NOT" ;
ln_keyword            ::= "LN" ;
log_keyword           ::= "LOG" ;
sin_keyword           ::= "SIN" ;
cos_keyword           ::= "COS" ;
tan_keyword           ::= "TAN" ;
asin_keyword          ::= "ASIN" ;
acos_keyword          ::= "ACOS" ;
atan_keyword          ::= "ATAN" ;

comment_block_marker  ::= "COMMENT:" ... "ENDCOMMENT" ; ( Handled/skipped by lexer )
number_literal        ::= digit+ [ "." digit+ ] ;
string_literal        ::= '"' { character | escape_sequence } '"'
                      | "'" { character | escape_sequence } "'" ;
assign_operator       ::= "=" ;
plus_operator         ::= "+" ;
minus_operator        ::= "-" ;
star_operator         ::= "" ;
slash_operator        ::= "/" ;
percent_operator      ::= "%" ;
star_star_operator    ::= "**" ;
ampersand_operator    ::= "&" ;
pipe_operator         ::= "|" ;
caret_operator        ::= "^" ;
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
eq_operator           ::= "==" ;
neq_operator          ::= "!=" ;
gt_operator           ::= ">" ;
lt_operator           ::= "<" ;
gte_operator          ::= ">=" ;
lte_operator          ::= "<=" ;

identifier            ::= letter { letter | digit | "" } ;

(* --- Implicit Lexer Rules --- )
line_comment          ::= ( "#" [^"!"\n\r] | "--" ) [^\n\r] ; (* Skipped )
hash_bang             ::= "#!" [^\n\r] ; (* Skipped )
newline               ::= "\r"? "\n" | "\r" ;
whitespace            ::= ( " " | "\t" )+ ; ( Skipped *)
letter                ::= "a".."z" | "A".."Z" | "" ;
digit                 ::= "0".."9" ;
character             ::= (* Any character except quotes or backslash or newline *) ;
escape_sequence       ::= "\" ( '"' | "'" | "\" | "n" | "r" | "t" | unicode_escape ) ;
unicode_escape        ::= "u" hex_digit hex_digit hex_digit hex_digit ;
hex_digit             ::= digit | "a".."f" | "A".."F" ;
