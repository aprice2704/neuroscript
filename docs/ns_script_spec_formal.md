# NeuroScript: Formal specification

Version: 0.4.0 
DependsOn: pkg/core/NeuroScript.g4.txt (Version ~0.6.0)
HowToUpdate: Review pkg/core/NeuroScript.g4.txt and update this EBNF accordingly.

(* EBNF-like notation for NeuroScript Language *)

:: Version: 0.6.0 (* Reflects grammar version for must enhancements *)
```ebnf

(* --- Top Level --- *)
program               ::= file_header ( procedure_definition newline* )* EOF ;
file_header           ::= ( metadata_line | newline )* ;

procedure_definition  ::= kw_func identifier signature_part kw_means newline metadata_block? statement_list kw_endfunc ;

signature_part        ::= ( lparen needs_clause? optional_clause? returns_clause? rparen )
                      | ( needs_clause optional_clause? returns_clause? )
                      | ( optional_clause returns_clause? )
                      | returns_clause
                      | (* empty *) ;

needs_clause          ::= kw_needs param_list ;
optional_clause       ::= kw_optional param_list ;
returns_clause        ::= kw_returns param_list ;
param_list            ::= identifier ( comma identifier )* ;

metadata_block        ::= ( metadata_line newline )* ;
metadata_line         ::= '::' metadata_value_content ;

(* --- Statements --- *)
statement_list        ::= body_line* ;
body_line             ::= ( statement newline ) | newline ;

statement             ::= simple_statement | block_statement ;

simple_statement      ::= set_statement
                      | call_statement
                      | return_statement
                      | emit_statement
                      | must_statement
                      | fail_statement
                      | clear_error_statement
                      | ask_statement
                      | break_statement
                      | continue_statement ;

block_statement       ::= if_statement
                      | while_statement
                      | for_each_statement
                      | on_error_statement ;

(* --- Simple Statements Details --- *)

(* UPDATE: set_statement now supports multiple assignment targets and must expressions. *)
set_statement         ::= kw_set identifier ( comma identifier )* assign expression ;

call_statement        ::= kw_call callable_expr ;
return_statement      ::= kw_return expression_list? ;
emit_statement        ::= kw_emit expression ;

(* UPDATE: must_statement now primarily refers to boolean assertions.
   Mandatory assignments are handled by 'must' as an expression prefix. See expression rules. *)
must_statement        ::= kw_must expression | kw_mustbe callable_expr ;

fail_statement        ::= kw_fail expression? ;
clear_error_statement ::= kw_clear_error ;
ask_statement         ::= kw_ask expression ( kw_into identifier )? ;
break_statement       ::= kw_break ;
continue_statement    ::= kw_continue ;

(* --- Block Statements Details --- *)
if_statement          ::= kw_if expression newline statement_list ( kw_else newline statement_list )? kw_endif ;
while_statement       ::= kw_while expression newline statement_list kw_endwhile ;
for_each_statement    ::= kw_for kw_each identifier kw_in expression newline statement_list kw_endfor ;
on_error_statement    ::= kw_on_error kw_means newline statement_list kw_endon ;

(* --- Expressions (Simplified Precedence) --- *)
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

(* UPDATE: unary_expr now includes the must_expression for mandatory assignments. *)
unary_expr            ::= ( minus | kw_not | kw_no | kw_some | tilde ) unary_expr
                      | must_expression
                      | power_expr ;

(* UPDATE: New expression for handling mandatory success checks on assignments. *)
must_expression       ::= kw_must ( accessor_expr | callable_expr ) ( must_type_assertion )? ;
must_type_assertion ::= kw_as type_name ( comma type_name )* ;
type_name             ::= identifier ;

power_expr            ::= accessor_expr ( star_star power_expr )? ;
accessor_expr         ::= primary ( lbrack expression rbrack )* ;

primary               ::= literal
                      | placeholder
                      | identifier
                      | kw_last
                      | callable_expr
                      | ( kw_eval lparen expression rparen )
                      | ( lparen expression rparen ) ;

callable_expr         ::= qualified_identifier lparen expression_list_opt rparen ;
qualified_identifier  ::= ( kw_tool dot )? identifier ( dot identifier )* ;

(* ... other definitions like literals, placeholders, etc. remain similar ... *)

expression_list_opt   ::= expression_list? ;
expression_list       ::= expression ( comma expression )* ;

(* --- Terminals (Keywords & Operators) --- *)
(* ... existing keywords ... *)
kw_as                 ::= "as" ; (* NEW *)
kw_must               ::= "must" ;
kw_set                ::= "set" ;
(* ... etc ... *)

```

:: language: md  
:: lang_version: neuroscript@0.4.0  
:: file_version: 2  
:: type: NSproject  
:: subtype: spec  
:: author: Gemini  
:: created: 2025-06-16  
:: modified: 2025-06-16  
:: dependsOn: ns_script_spec_formal.md (original), must_enhancements.md  
:: howToUpdate: Update the EBNF to reflect changes in the core ANTLR G4 grammar file, particularly for new keywords and statement structures.