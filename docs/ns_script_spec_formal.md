 # NeuroScript — Formal Language Specification (v0.4.3)
 
 > **Status:** DRAFT
 > **Source:** Derived directly from the `NeuroScript.g4` (v0.4.2) grammar file.
 > **Note:** This document defines the formal structure of NeuroScript using a human-readable EBNF notation that mirrors the official ANTLR grammar.
 
 ---
 
 ## 1 · Notation
 
 This document uses a simplified EBNF-like notation:
 
 * `::=`  means "is defined as".
 * `|`    means "OR".
 * `{...}` means "zero or more" repetitions (ANTLR `*`).
 * `[...]` means "optional" (zero or one) (ANTLR `?`).
 * `<NonTerminal>` refers to another rule (lowercase_with_underscores).
 * `'TOKEN'` refers to a literal keyword or symbol (UPPERCASE).
 * `ε` represents an empty production.
 
 ---
 
 ## 2 · Script Structure
 
 A NeuroScript file consists of a header containing metadata, followed by a series of code blocks.
 
 ```ebnf
 <program> ::= <file_header> { <code_block> } 'EOF'
 
 <file_header> ::= { 'METADATA_LINE' | 'NEWLINE' } 
 
 <code_block> ::= ( <procedure_definition> | <on_statement> ) { 'NEWLINE' } 
 ```
 
 ---
 
 ## 3 · Procedures
 
 A procedure is a named, callable block of code defined with `'func'`.
 
 ```ebnf
 <procedure_definition> ::= 'KW_FUNC' 'IDENTIFIER' <signature_part> 'KW_MEANS' 'NEWLINE' <metadata_block> <statement_list> 'KW_ENDFUNC' 
 
 <signature_part> ::= 'LPAREN' { <needs_clause> | <optional_clause> | <returns_clause> } 'RPAREN'
                    | ( <needs_clause> | <optional_clause> | <returns_clause> )+
                    | ε 
 
 <needs_clause> ::= 'KW_NEEDS' <param_list> 
 <optional_clause> ::= 'KW_OPTIONAL' <param_list> 
 <returns_clause> ::= 'KW_RETURNS' <param_list> 
 
 <param_list> ::= 'IDENTIFIER' { 'COMMA' 'IDENTIFIER' } 
 
 <metadata_block> ::= { 'METADATA_LINE' 'NEWLINE' } 
 ```
 
 ---
 
 ## 4 · Statements
 
 A statement list is a sequence of statements, each on its own line.
 
 ```ebnf
 <statement_list> ::= { <body_line> } 
 
 <body_line> ::= <statement> 'NEWLINE' | 'NEWLINE' 
 
 <statement> ::= <simple_statement> | <block_statement> | <on_statement> 
 ```
 
 ### 4.1 Simple Statements
 Simple statements are single-line executable constructs.
 
 ```ebnf
 <simple_statement> ::= <set_statement> | <call_statement> | <return_statement> 
                      | <emit_statement> | <must_statement> | <fail_statement> 
                      | <clear_error_statement> | <clear_event_statement> | <ask_statement> 
                      | <break_statement> | <continue_statement> 
 
 <lvalue> ::= 'IDENTIFIER' { 'LBRACK' <expression> 'RBRACK' | 'DOT' 'IDENTIFIER' } 
 <lvalue_list> ::= <lvalue> { 'COMMA' <lvalue> } 
 
 <set_statement> ::= 'KW_SET' <lvalue_list> 'ASSIGN' <expression> 
 <call_statement> ::= 'KW_CALL' <callable_expr> 
 <return_statement> ::= 'KW_RETURN' [ <expression_list> ] 
 <emit_statement> ::= 'KW_EMIT' <expression> 
 <must_statement> ::= 'KW_MUST' <expression> | 'KW_MUSTBE' <callable_expr> 
 <fail_statement> ::= 'KW_FAIL' [ <expression> ] 
 <clear_error_statement> ::= 'KW_CLEAR_ERROR' 
 <clear_event_statement> ::= 'KW_CLEAR' 'KW_EVENT' ( <expression> | 'KW_NAMED' 'STRING_LIT' ) 
 <ask_statement> ::= 'KW_ASK' <expression> [ 'KW_INTO' 'IDENTIFIER' ] 
 <break_statement> ::= 'KW_BREAK' 
 <continue_statement> ::= 'KW_CONTINUE' 
 ```
 
 ### 4.2 Block Statements
 Block statements contain a nested list of statements.
 
 ```ebnf
 <block_statement> ::= <if_statement> | <while_statement> | <for_each_statement> 
 
 <if_statement> ::= 'KW_IF' <expression> 'NEWLINE' <statement_list>
                    [ 'KW_ELSE' 'NEWLINE' <statement_list> ]
                    'KW_ENDIF' 
 
 <while_statement> ::= 'KW_WHILE' <expression> 'NEWLINE' <statement_list> 'KW_ENDWHILE' 
 
 <for_each_statement> ::= 'KW_FOR' 'KW_EACH' 'IDENTIFIER' 'KW_IN' <expression> 'NEWLINE' <statement_list> 'KW_ENDFOR' 
 ```
 
 ### 4.3 On Statements (Handlers)
 Handlers for events and errors are defined with `on`.
 
 ```ebnf
 <on_statement> ::= 'KW_ON' ( <error_handler> | <event_handler> ) 
 
 <error_handler> ::= 'KW_ERROR' 'KW_DO' 'NEWLINE' <statement_list> 'KW_ENDON' 
 
 <event_handler> ::= 'KW_EVENT' <expression> [ 'KW_NAMED' 'STRING_LIT' ] [ 'KW_AS' 'IDENTIFIER' ] 'KW_DO' 'NEWLINE' <statement_list> 'KW_ENDON' 
 ```
 ---
 ## 5 · Expressions
 
 Expressions are defined with a specific operator precedence, from lowest to highest.
 
 ```ebnf
 <expression> ::= <logical_or_expr> 
 
 <logical_or_expr> ::= <logical_and_expr> { 'KW_OR' <logical_and_expr> } 
 
 <logical_and_expr> ::= <bitwise_or_expr> { 'KW_AND' <bitwise_or_expr> } 
 
 <bitwise_or_expr> ::= <bitwise_xor_expr> { 'PIPE' <bitwise_xor_expr> } 
 
 <bitwise_xor_expr> ::= <bitwise_and_expr> { 'CARET' <bitwise_and_expr> } 
 
 <bitwise_and_expr> ::= <equality_expr> { 'AMPERSAND' <equality_expr> } 
 
 <equality_expr> ::= <relational_expr> { ( 'EQ' | 'NEQ' ) <relational_expr> } 
 
 <relational_expr> ::= <additive_expr> { ( 'GT' | 'LT' | 'GTE' | 'LTE' ) <additive_expr> } 
 
 <additive_expr> ::= <multiplicative_expr> { ( 'PLUS' | 'MINUS' ) <multiplicative_expr> } 
 
 <multiplicative_expr> ::= <unary_expr> { ( 'STAR' | 'SLASH' | 'PERCENT' ) <unary_expr> } 
 
 <unary_expr> ::= ( 'MINUS' | 'KW_NOT' | 'KW_NO' | 'KW_SOME' | 'TILDE' | 'KW_MUST' ) <unary_expr>
                | 'KW_TYPEOF' <unary_expr>
                | <power_expr> 
 
 <power_expr> ::= <accessor_expr> [ 'STAR_STAR' <power_expr> ] 
 
 <accessor_expr> ::= <primary> { 'LBRACK' <expression> 'RBRACK' } 
 ```
 
 ### 5.1 Primary Expressions
 Primary expressions are the atomic operands of the language.
 
 ```ebnf
 <primary> ::= <literal>
             | <placeholder>
             | 'IDENTIFIER'
             | 'KW_LAST'
             | <callable_expr>
             | 'KW_EVAL' 'LPAREN' <expression> 'RPAREN'
             | 'LPAREN' <expression> 'RPAREN' 
 
 <callable_expr> ::= ( <call_target> | 'KW_LN' | 'KW_LOG' | 'KW_SIN' | 'KW_COS' | 'KW_TAN' | 'KW_ASIN' | 'KW_ACOS' | 'KW_ATAN' | 'KW_LEN' ) 
                     'LPAREN' [ <expression_list> ] 'RPAREN' 
 
 <call_target> ::= 'IDENTIFIER' | 'KW_TOOL' 'DOT' <qualified_identifier> 
 
 <qualified_identifier> ::= 'IDENTIFIER' { 'DOT' 'IDENTIFIER' } 
 
 <placeholder> ::= 'PLACEHOLDER_START' ( 'IDENTIFIER' | 'KW_LAST' ) 'PLACEHOLDER_END' 
 ```
 
 ---
 
 ## 6 · Literals & Lists
 
 ```ebnf
 <literal> ::= 'STRING_LIT'
             | 'TRIPLE_BACKTICK_STRING'
             | 'NUMBER_LIT'
             | <list_literal>
             | <map_literal>
             | <boolean_literal>
             | <nil_literal> 
 
 <nil_literal> ::= 'KW_NIL' 
 <boolean_literal> ::= 'KW_TRUE' | 'KW_FALSE' 
 
 <list_literal> ::= 'LBRACK' [ <expression_list> ] 'RBRACK' 
 <expression_list> ::= <expression> { 'COMMA' <expression> } 
 
 <map_literal> ::= 'LBRACE' [ <map_entry_list> ] 'RBRACE' 
 <map_entry_list> ::= <map_entry> { 'COMMA' <map_entry> } 
 <map_entry> ::= 'STRING_LIT' 'COLON' <expression> 
 ```