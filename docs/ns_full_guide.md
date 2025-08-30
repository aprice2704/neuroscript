# A Comprehensive Guide to NeuroScript

### [1. Introduction to NeuroScript](#1-introduction-to-neuroscript)
* [1.1. What is NeuroScript?](#11-what-is-neuroscript)
* [1.2. Key Language Philosophies](#12-key-language-philosophies)
* [1.3. Your First NeuroScript File: A "Hello, World" Example](#13-your-first-neuroscript-file-a-hello-world-example)

### [2. Lexical Structure & Core Syntax](#2-lexical-structure--core-syntax)
* [2.1. Comments](#21-comments)
* [2.2. Whitespace, Newlines, and Line Continuations](#22-whitespace-newlines-and-line-continuations)
* [2.3. Identifiers](#23-identifiers)
* [2.4. Keywords](#24-keywords)
* [2.5. File Header and Metadata](#25-file-header-and-metadata)

### [3. Data Types and Literals](#3-data-types-and-literals)
* [3.1. Overview of NeuroScript's Dynamic Types](#31-overview-of-neuroscripts-dynamic-types)
* [3.2. Primitive Types](#32-primitive-types)
* [3.3. Composite Types](#33-composite-types)
* [3.4. Special-Purpose Types](#34-special-purpose-types)

### [4. Variables, State, and Expressions](#4-variables-state-and-expressions)
* [4.1. Variable Scope and Lifetime](#41-variable-scope-and-lifetime)
* [4.2. Predefined Variables: `self`](#42-predefined-variables-self)
* [4.3. L-Values: The Targets of Assignment](#43-l-values-the-targets-of-assignment)
* [4.4. Assigning State with the `set` Statement](#44-assigning-state-with-the-set-statement)
* [4.5. Placeholders for String Interpolation](#45-placeholders-for-string-interpolation)
* [4.6. Operator Precedence](#46-operator-precedence)
* [4.7. Detailed Operator Guide](#47-detailed-operator-guide)

### [5. Fundamental Statements](#5-fundamental-statements)
* [5.1. The `must` Statement: Asserting Truth](#51-the-must-statement-asserting-truth)
* [5.2. The `call` Statement: Invoking Functions and Tools](#52-the-call-statement-invoking-functions-and-tools)
* [5.3. The `ask` Statement: Querying AI Models](#53-the-ask-statement-querying-ai-models)
* [5.4. The `promptuser` Statement: Getting User Input](#54-the-promptuser-statement-getting-user-input)
* [5.5. The `emit` Statement: Firing Events](#55-the-emit-statement-firing-events)
* [5.6. The `whisper` Statement: Providing Context](#56-the-whisper-statement-providing-context)
* [5.7. The `fail` Statement: Halting with an Error](#57-the-fail-statement-halting-with-an-error)
* [5.8. State Clearing Statements](#58-state-clearing-statements)

### [6. Control Flow Structures](#6-control-flow-structures)
* [6.1. Conditional Logic: `if`/`else`/`endif`](#61-conditional-logic-ifelseendif)
* [6.2. Looping](#62-looping)
* [6.3. Modifying Loop Behavior: `break` and `continue`](#63-modifying-loop-behavior-break-and-continue)
* [6.4. Limitations](#64-limitations)

### [7. Scripting Models](#7-scripting-models)
* [7.1. Command Scripts: The Top-Level Execution Block](#71-command-scripts-the-top-level-execution-block)
* [7.2. Library Scripts](#72-library-scripts)

### [8. Procedures, Tools, and Calls](#8-procedures-tools-and-calls)
* [8.1. Defining a Function (`func`)](#81-defining-a-function-func)
* [8.2. Defining a Signature](#82-defining-a-signature)
* [8.3. The `return` Statement](#83-the-return-statement)
* [8.4. `callable_expr`: How Functions are Called](#84-callable_expr-how-functions-are-called)
* [8.5. Built-in Functions](#85-built-in-functions)
* [8.6. External Logic: The `tool` Keyword](#86-external-logic-the-tool-keyword)

### [9. Event and Error Handling](#9-event-and-error-handling)
* [9.1. The Event Model](#91-the-event-model)
* [9.2. The Error Model](#92-the-error-model)

### [10. Advanced Operators and Reserved Keywords](#10-advanced-operators-and-reserved-keywords)
* [10.1. Type Introspection with `typeof`](#101-type-introspection-with-typeof)
* [10.2. The `some` and `no` Operators](#102-the-some-and-no-operators)
* [10.3. Reserved Keywords for Future Use](#103-reserved-keywords-for-future-use)



## 1. Introduction to NeuroScript

Welcome to NeuroScript, a language designed from the ground up to be a clear, expressive, and powerful bridge between human developers and artificial intelligence. It provides a shared, structured medium for defining tasks, logic, and data in a way that is both human-readable and machine-executable.

NeuroScript (ns) is **not** designed to be: subtle, refined, highly-expressive, elegant, rich, strongly-typed, particularly fast, or enjoyable for human developers to write. We expect that most ns will be written by AIs.

---

### 1.1. What is NeuroScript?

At its core, **NeuroScript is a high-level, statement-driven scripting language**. Its primary purpose is to orchestrate complex workflows, manage state, and interact with external systems, or **"tools"**.

Think of it as a set of precise instructions that can be written by an AI, understood by a person, and executed by a computer or AI. This makes it ideal for:

- **Automating complex tasks:** Define multi-step processes that involve logic, data manipulation, and calls to external APIs or functions.
- **AI and Agentic Workflows:** Provide a clear, unambiguous language for AI agents to understand goals, execute tasks, and report results.
- **Configuration and Logic Definition:** Create readable and maintainable files that define behavior for systems and applications.

---

### 1.2. Key Language Philosophies

NeuroScript's design is guided by a few core principles that make it robust and easy to use.

#### 1.2.1. Explicit Statement-Driven Logic

Every action in NeuroScript is initiated by an explicit keyword statement (like `set`, `call`, `emit`, or `if`). There are no "bare" expressions that execute on their own. This design choice eliminates ambiguity and makes the code's intent incredibly clear. You always know what action a line of code is intended to perform.

```neuroscript
# This is a statement. It explicitly assigns a value.
set my_variable = "hello"

# This is also a statement. It explicitly calls a function.
call tool.print(my_variable)

# This would be an error. An expression alone does nothing.
# "hello" 
```

#### 1.2.2. Dual Scripting Models

NeuroScript files can operate in one of two primary modes, allowing you to either define reusable logic or execute an immediate task.

1.  **Library Scripts:** These files act as collections of reusable procedures (`func`) and event handlers (`on event`). They are designed to be loaded by other scripts or an execution engine to provide a library of functions and behaviors.
2.  **Command Scripts:** These files define a single, top-level `command` block that is executed immediately. This is useful for one-off tasks or for serving as the main entry point of a program.

A script can contain either library blocks or command blocks, but not both.

#### 1.2.3. Metadata-Rich Files

NeuroScript encourages the embedding of metadata directly into the source code. Lines beginning with `::` are treated as metadata, allowing you to document the file's purpose, version, author, or any other relevant information directly within the script itself.

This metadata can be used by the interpreter or other tools for documentation, version control, or conditional execution.

```neuroscript
:: title: My First Script
:: version: 1.0
:: author: A. Developer

func main() means
  :: description: This is the main entry point.
  emit "Hello from a well-documented script!"
endfunc
```

---

### 1.3. Your First NeuroScript File: A "Hello, World" Example

The best way to understand NeuroScript is to see it in action. Here is a complete, simple "Hello, World" script.

This example uses the **command script** model for direct execution.

```neuroscript
:: title: Hello World Example
:: version: 1.0
:: purpose: A minimal script to demonstrate basic syntax.

command
  # The 'set' keyword assigns a value to a variable.
  set message = "Hello, NeuroScript!"

  # The 'emit' keyword outputs a value, typically to the console
  # or as an event within the host system.
  emit message
endcommand
```

When executed, this script will:
1.  Define a variable named `message` and assign it the string `"Hello, NeuroScript!"`.
2.  `emit` the content of the `message` variable.
---
# 2. Lexical Structure & Core Syntax

The lexical structure of a language defines its most basic rules: how comments are written, how words and symbols are interpreted, and what constitutes the fundamental building blocks of the code.

---

### 2.1. Comments

Comments are used to leave explanatory notes in the code that are ignored by the interpreter. NeuroScript supports three styles of single-line comments. There are no multi-line comment blocks.

```neuroscript
# A hash symbol starts a comment.
-- Two dashes also start a comment.
// Two slashes are also valid.

set x = 1 # A comment can also appear after a statement.
```

---

### 2.2. Whitespace, Newlines, and Line Continuations

**Whitespace** (spaces and tabs) is used to separate elements in the code to improve readability. It is generally ignored by the interpreter except where it is needed to distinguish one keyword or identifier from another.

**Newlines** are significant. They are used to terminate statements. Every simple statement and most block headers must be followed by a newline.

**Line Continuations** allow you to break a long statement across multiple physical lines for readability. A backslash (`\`) at the very end of a line tells the interpreter to treat the next line as part of the current one.

```neuroscript
# A long statement can be broken up using a backslash.
set a_very_long_variable_name = "part 1" + \
                              "part 2" + \
                              "part 3"
```

---

### 2.3. Identifiers

An **identifier** is the name given to a variable or a function (`func`).

- Identifiers must start with a letter (`a-z`, `A-Z`) or an underscore (`_`).
- After the first character, they can contain letters, numbers (`0-9`), or underscores.
- Identifiers are **case-sensitive**. `myVar` is a different identifier from `myvar`.

```neuroscript
# Valid Identifiers
set my_variable = 1
set _internal_var = 2
set FunctionName = 3

# Invalid Identifier (starts with a number)
# set 1_bad_name = 4
```

---

### 2.4. Keywords

Keywords are reserved words that have special meaning in NeuroScript and cannot be used as identifiers. All keywords are lowercase.

The full list of keywords includes: `acos`, `and`, `as`, `asin`, `ask`, `atan`, `break`, `call`, `clear`, `clear_error`, `command`, `continue`, `cos`, `do`, `each`, `else`, `emit`, `endcommand`, `endfor`, `endfunc`, `endif`, `endon`, `endwhile`, `error`, `eval`, `event`, `fail`, `false`, `for`, `func`, `fuzzy`, `if`, `in`, `into`, `last`, `len`, `ln`, `log`, `means`, `must`, `mustbe`, `named`, `needs`, `nil`, `no`, `not`, `on`, `optional`, `or`, `promptuser`, `return`, `returns`, `set`, `sin`, `some`, `tan`, `timedate`, `tool`, `true`, `typeof`, `while`, `whisper`.

---

### 2.5. File Header and Metadata

A NeuroScript file can begin with a **header** section that contains file-level metadata and blank lines. This header is processed before the main script logic.

**Metadata lines** start with `::` and provide key-value information about the script. The key is separated from the value by the first colon (`:`).

```neuroscript
:: title: Script with a Header
:: version: 1.0
:: author: AI Agent 42

# The header ends at the first line of code.
func doSomething() means
  :: description: This metadata belongs to the function, not the file.
  emit "Script started"
endfunc
```

The `file_header` can contain any number of metadata lines or newlines before the first `command` or `func` block begins.

---
# 3. Data Types and Literals

A **literal** is a value that is written directly into the source code, like the number `123` or the string `"hello"`. NeuroScript has a rich set of built-in data types, and their literal forms are the foundation of working with data in the language.

---

### 3.1. Overview of NeuroScript's Dynamic Types

NeuroScript is a **dynamically-typed** language. This means you do not need to declare the type of a variable before you use it. The type is determined at runtime based on the value assigned to it.

```neuroscript
# 'my_var' is a number here...
set my_var = 10

# ...and now it's a string. This is perfectly valid.
set my_var = "ten"
```

The built-in types can be categorized into three groups: **Primitive**, **Composite**, and **Special-Purpose**.

---

### 3.2. Primitive Types

Primitive types are the simplest, most fundamental data types.

#### 3.2.1. String
A sequence of characters to represent text. NeuroScript supports single-quoted (`'...'`), double-quoted (`"..."`), and triple-backtick (`` `...` ``) raw string literals.

#### 3.2.2. Number
Represents both integers (`100`) and floating-point (`3.14`) numbers. The language does not have a `complex` type.

#### 3.2.3. Boolean
A truth value, which can only be `true` or `false`.

#### 3.2.4. Bytes
Represents a sequence of raw bytes. There is no literal for bytes; they are typically created by tools (e.g., reading a file). This is useful for file I/O and networking contexts.

#### 3.2.5. Nil
The `nil` type represents the intentional absence of any value.

---

### 3.3. Composite Types

Composite types are complex types built up from simpler ones.

#### 3.3.1. List
An ordered collection of values, enclosed in square brackets `[]`. A list can contain any mix of data types.

```neuroscript
set my_list = [1, "hello", true, nil]
```

Lists are **mutable**, meaning they can be changed in-place. Appending elements is handled by external tools (e.g., `tool.List.Append()`), not a built-in operator.

> **Note on Passing Lists to Functions:** Composite types like Lists and Maps behave as if they are **passed by reference**. Modifications made to a list inside a function will affect the original variable in the caller's scope.

#### 3.3.2. Map
A collection of key-value pairs, enclosed in curly braces `{}`. Keys must be string literals, and values can be any data type.

```neuroscript
set my_map = {"name": "Agent Smith", "id": 101}
```

##### Using Maps as Structs or Objects

NeuroScript does not have a formal `struct` or `class` construct. Instead, it uses **Maps** to create structured data. The dot notation for access (`my_map.key`) is syntactic sugar for bracket notation (`my_map["key"]`), allowing maps to be used in a way that feels like accessing properties on an object. The `set` statement will even automatically create nested maps as needed ("auto-vivification").

```neuroscript
# This single line creates a nested structure of maps.
set user.address.city = "Zion"

# It is equivalent to:
# set user = {"address": {"city": "Zion"}}
```

---

### 3.4. Special-Purpose Types

These types have specific roles within the NeuroScript ecosystem. They generally do not have a direct literal representation and are instead returned from functions, tools, or specific language constructs.

#### 3.4.1. Function
A reference to a `func` defined within a script. Functions are **first-class citizens**, meaning they can be assigned to variables and passed as arguments to other functions or tools. This enables higher-order programming patterns.

#### 3.4.2. Tool
A reference to an external capability provided by the host environment.

#### 3.4.3. Error
A special type that holds information about a runtime error, such as an error code and message.

#### 3.4.4. Event
Represents an event that can be emitted (`emit`) or handled (`on event`).

#### 3.4.5. Timedate
Represents a specific point in time, often with nanosecond precision. The `timedate` keyword can be used to get the current time. Arithmetic and comparisons (e.g., before/after) are handled by tools.

```neuroscript
set now = timedate
```

#### 3.4.6. Fuzzy
A reserved type for fuzzy logic operations, which allow for degrees of truth rather than simple true/false values.

#### 3.4.7. Unknown
An internal type representing an indeterminate or unrecognized type, often as the result of an error.
---

# 4. Variables, State, and Expressions

The core of any NeuroScript program involves managing state through variables and manipulating that data using expressions. This section covers how to assign values, how to access them, and the rules that govern the operations you can perform.

---

### 4.1. Variable Scope and Lifetime

NeuroScript's variable scope is simple and predictable, designed to be easily understood by both humans and AIs.

- **Function-Level Scope:** The scope of a variable is the entire `func` or `command` block in which it is first defined. There is no block-level scope; a variable defined inside an `if` or `while` block is accessible for the rest of the function.
- **Function Call Isolation:** When a function is called, it executes in a **new, isolated memory space**. It cannot directly access or modify the local variables of its caller. Data is passed explicitly through parameters and `return` statements.
- **Sandboxed Event Handlers:** `on event` handlers also execute in a sandboxed, isolated scope. They have read-only access to any global variables but cannot modify them, preventing unintended side effects on the main program state.

```neuroscript
func scope_example() means
  set x = 10 # x is visible throughout the function
  if x > 5
    set y = 20 # y is also visible throughout the function
  endif
  emit y # This is valid and will emit 20
endfunc
```

---

### 4.2. Predefined Variables: `self`

NeuroScript provides a small number of predefined global variables. The most important is `self`.

- **`self`**: This variable holds the handle to the interpreter's default internal text buffer. Its primary purpose is to serve as a channel for providing contextual information to the `ask` statement via the `whisper` command.

---

### 4.3. L-Values: The Targets of Assignment

An **l-value** (short for "left-hand-side value") is anything that can appear on the left side of an assignment (`=`) statement. It represents a memory location where a value can be stored. In NeuroScript, an l-value can be a simple variable or a more complex path that accesses elements within lists or maps.

The basic syntax for an l-value is an `IDENTIFIER`, which can be followed by any number of accessors:

- **Dot Access (`.key`):** Accesses a value in a map using a static key.
- **Bracket Access (`[expression]`):** Accesses an element in a list using a number or a value in a map using a string. The expression inside the brackets is fully evaluated to determine the key.

```neuroscript
# Simple L-Values (just identifiers)
set my_variable = 1

# Complex L-Values
set my_map.user.name = "Alice"           # Dot access
set my_list[0] = "first_item"           # Bracket access with number
set key_name = "user"
set my_map[key_name] = "bob"            # Bracket access with variable
set data.users[1].email = "a@b.com"   # Mixed access
```

---

### 4.4. Assigning State with the `set` Statement

The `set` statement is the fundamental way to create or modify variables in NeuroScript. It evaluates the expression on the right side of the `=` and assigns the result to the l-value on the left.

```neuroscript
# Assign a literal value
set count = 0

# Assign the result of an expression
set total = count + 10 * 5

# Assign the result of a function call
set current_time = tool.time.now()

# You can assign to multiple l-values from a list.
# The number of variables must exactly match the number of items.
set a, b, c = [1, 2, 3]
# `a` is now 1, `b` is 2, `c` is 3
```

---

### 4.5. Placeholders for String Interpolation

Placeholders allow you to embed the value of an identifier directly inside a string literal. This is a convenient alternative to string concatenation. A placeholder is an identifier enclosed in double curly braces `{{...}}`.

Placeholders are only evaluated within triple-backtick ```` ```...``` ```` strings.

```neuroscript
set name = "World"
set raw_string = ```Hello, {{name}}! The value of 2+2 is {{2+2}}.```

# After interpolation, raw_string would be "Hello, World! The value of 2+2 is 4."
emit raw_string
```

---

### 4.6. Operator Precedence

NeuroScript has a well-defined order of operations to ensure that complex expressions are evaluated predictably. Operators with higher precedence are evaluated before operators with lower precedence.

The order from **highest to lowest** precedence is:

1.  **Accessor** (`[]`) and Function Call (`()`)
2.  **Power** (`**`)
3.  **Unary** (`-`, `not`, `no`, `some`, `~`, `typeof`)
4.  **Multiplicative** (`*`, `/`, `%`)
5.  **Additive** (`+`, `-`)
6.  **Relational** (`>`, `<`, `>=`, `<=`)
7.  **Equality** (`==`, `!=`)
8.  **Bitwise AND** (`&`)
9.  **Bitwise XOR** (`^`)
10. **Bitwise OR** (`|`)
11. **Logical AND** (`and`)
12. **Logical OR** (`or`)

You can use parentheses `()` to override the default precedence and force an expression to be evaluated first.

---

### 4.7. Detailed Operator Guide

#### 4.7.1. Arithmetic Operators

| Operator | Description      | Example         |
| :------- | :--------------- | :-------------- |
| `+`      | Addition         | `5 + 3`         |
| `-`      | Subtraction      | `5 - 3`         |
| `*`      | Multiplication   | `5 * 3`         |
| `/`      | Division         | `6 / 3`         |
| `%`      | Modulo/Remainder | `5 % 3`         |
| `**`     | Power/Exponent   | `5 ** 3`        |

#### 4.7.2. Comparison Operators

| Operator | Description          | Example      |
| :------- | :------------------- | :----------- |
| `==`     | Equal to             | `a == b`     |
| `!=`     | Not equal to         | `a != b`     |
| `>`      | Greater than         | `a > b`      |
| `<`      | Less than            | `a < b`      |
| `>=`     | Greater than or equal to | `a >= b`     |
| `<=`     | Less than or equal to    | `a <= b`     |

#### 4.7.3. Logical Operators

| Operator | Description                                       | Example        |
| :------- | :------------------------------------------------ | :------------- |
| `and`    | Returns `true` if both operands are true          | `a and b`      |
| `or`     | Returns `true` if at least one operand is true    | `a or b`       |

#### 4.7.4. Bitwise Operators

| Operator | Description      |
| :------- | :--------------- |
| `&`      | Bitwise AND      |
| `|`      | Bitwise OR       |
| `^`      | Bitwise XOR      |
| `~`      | Bitwise NOT      |

> The language does not currently support bit-shifting operators (`<<`, `>>`).

#### 4.7.5. Unary Operators

| Operator | Description                                   | Example         |
| :------- | :-------------------------------------------- | :-------------- |
| `-`      | Negates a number                              | `-5`            |
| `not`    | Inverts a boolean value                       | `not true`      |
| `no`     | Checks if a list is empty or `nil`           | `no my_list`    |
| `some`   | Checks if a list contains at least one element | `some my_list`  |
| `typeof` | Returns the data type of its operand as a string | `typeof 123`  |

#### 4.7.6. Fuzzy Operators

NeuroScript supports fuzzy logic, which deals with reasoning that is approximate rather than precise. A `fuzzy` value represents a degree of truth, typically between 0.0 (completely false) and 1.0 (completely true).

##### Creating a Fuzzy Value

Currently, there is **no literal syntax** for creating a fuzzy value directly. This is a planned future addition. For now, fuzzy values can only be created as the result of specific tool calls that are designed to measure degrees of similarity, confidence, or other non-binary states.

```neuroscript
# A hypothetical tool that compares two strings and returns a
# fuzzy value representing their similarity.
set similarity_score = tool.text.get_similarity("apple", "apples")

# similarity_score is now a `fuzzy` type.
```

##### Using Fuzzy Values

Fuzzy operations are **automatically triggered** when one or both operands of a standard operator is a `fuzzy` type.

- **Fuzzy Equality and Relational (`==`, `!=`, `>`, etc.):** When comparing a `fuzzy` value to another value, the result is a standard boolean (`true` or `false`) based on whether the fuzzy value is above or below a certain threshold (typically 0.5).

- **Fuzzy Logical (`and`, `or`):** When logical operators are used with `fuzzy` values, they perform fuzzy logic calculations (like taking the minimum value for `and` and the maximum for `or`) and return a new `fuzzy` value.
- 
---

# 5. Fundamental Statements

Fundamental statements are the explicit, keyword-driven commands that form the building blocks of any NeuroScript program. Unlike expressions, which are evaluated to produce a value, statements are executed to perform an action, such as assigning a variable, controlling program flow, or interacting with the host system. Every line of executable code in NeuroScript is a statement.

---

### 5.1. The `must` Statement: Asserting Truth

The `must` statement is a powerful assertion tool. It evaluates an expression, and if the result is not `true`, it halts the program with an error. It is the primary mechanism for enforcing preconditions and validating state within a script.

**Syntax:** `must <expression>`

```neuroscript
func process_data(needs data_map) means
  # Ensure the input map has the required key before proceeding.
  must data_map["status"] == "ready"

  # ... continue processing, confident that status is 'ready'
endfunc
```

---

### 5.2. The `call` Statement: Invoking Functions and Tools

The `call` statement is used to execute a function or a tool without assigning its return value to a variable. This is useful when you are interested in the *side effects* of the call, such as printing to the console or modifying a file.

**Syntax:** `call <callable_expression>`

```neuroscript
# Execute a tool that performs an action but returns nothing.
call tool.log.info("Processing has started.")

# Call a user-defined function for its side effects.
call cleanup_temporary_files()
```

---

### 5.3. The `ask` Statement: Querying AI Models

The `ask` statement is the primary interface for interacting with pre-configured AI models. It sends a prompt to a specified model and can store the model's response in a variable. This statement is dedicated to AI interaction, not general user input.

**Syntax:** `ask <model_expression>, <prompt_expression> [into <l-value>]`

Before using `ask`, models must be configured and registered with the host environment, typically using a tool. This involves defining the model's properties (like provider and name) in a map and then registering it with a human-friendly name.

```neuroscript
command
  # 1. Define the configuration for an AI model as a map.
  # The 'api_key_ref' points to the name of an environment variable.
  set codey_config = {\
    "provider": "google",\
    "model": "gemini-2.5-flash",\
    "api_key_ref": "GOOGLE_API_KEY"\
  }

  # 2. Register the model with the system using a tool,
  # giving it a friendly name to use with 'ask'.
  must tool.agentmodel.Register("code_assistant", codey_config)

  # 3. Use 'ask' to send a prompt to the registered model
  # and store the response in the 'answer' variable.
  ask "code_assistant", "please write the sieve of Eratosthenes in forth" into answer

  emit "AI Response:"
  emit answer
endcommand
```

---

### 5.4. The `promptuser` Statement: Getting User Input

The `promptuser` statement is used to prompt a human user for text input. It sends an expression (typically a question string) to the host environment and stores the user's response in a variable.

**Syntax:** `promptuser <expression> into <l-value>`

```neuroscript
command
  # Ask a question and store the answer in the 'user_name' variable.
  promptuser "What is your name?" into user_name

  emit "Hello, " + user_name
endcommand
```

---

### 5.5. The `emit` Statement: Firing Events

The `emit` statement is the primary way for a script to output data or signal that something has happened. The host system determines how to handle an emitted value—it could be printed to the console, logged to a file, or broadcast as an event to other parts of an application. For sending data to multiple specific outputs, it is recommended to use tools (e.g., `tool.log.info`, `tool.network.send`).

**Syntax:** `emit <expression>`

```neuroscript
set user_name = "Alex"

# Emit a simple string.
emit "User logged in."

# Emit the contents of a variable.
emit user_name

# Emit a complex data structure.
emit {"user": user_name, "status": "active"}
```

---

### 5.6. The `whisper` Statement: Providing Context

The `whisper` statement is a specialized way to send information to a specific "handle" without it being part of the main `emit` stream. Its primary purpose is to provide contextual "observations" to the `ask` statement.

**Syntax:** `whisper <handle_expression>, <value_expression>`

- `<handle_expression>`: An expression that resolves to a handle identifying the destination. The host application determines what each handle means.
- `<value_expression>`: The data to be sent.

While applications can define many custom handles, NeuroScript provides one special, predefined handle through the global variable `self`. Anything whispered to `self` is automatically collected and included as context in the next `ask` call.

```neuroscript
# Get the current state of a file
set file_state = tool.fs.stat("config.txt")

# Whisper this state to 'self' so the AI knows about it.
# This text will not appear in the normal 'emit' stream.
whisper self, "The current state of config.txt is: " + file_state

# Now, ask the AI a question. The whispered text will be
# included in the prompt context automatically.
ask "default_agent", "Should I update config.txt?" into decision
```

---

### 5.7. The `fail` Statement: Halting with an Error

The `fail` statement immediately stops the execution of the script and raises an error. You can optionally provide an expression (like an error message) to give more context about the failure. This is often used in `else` blocks or when validation checks do not pass.

**Syntax:** `fail [<expression>]`

```neuroscript
func get_user_data(needs user_id) means
  set user = tool.db.find_user(user_id)
  if user == nil
    # Stop execution if the user could not be found.
    fail "User with ID " + user_id + " not found."
  endif
  return user
endfunc
```

---

### 5.8. State Clearing Statements

These statements are used to reset specific states within the interpreter.

#### 5.8.1. `clear_error`

Used inside an `on error` block, `clear_error` resets the script's error state, allowing execution to continue normally after the error handler finishes. If the error is not cleared, the script will terminate after the handler completes.

**Syntax:** `clear_error`

#### 5.8.2. `clear event`

The `clear event` statement removes event handlers that were previously registered with `on event`. You can clear a handler by its event name expression or by a specific handler name.

**Syntax:** `clear event <expression>` or `clear event named <string_literal>`

```neuroscript
# Remove the handler for a specific event
clear event "user.login"

# Remove a handler that was given a specific name
clear event named "MyLoginHandler"
```

---

# 6. Control Flow Structures

Control flow structures allow you to direct the execution of your script, enabling it to make decisions, repeat actions, and handle different situations in a structured way. In NeuroScript, all control flow is managed with explicit blocks that have clear start and end keywords.

---

### 6.1. Conditional Logic: `if`/`else`/`endif`

The `if` statement is the primary tool for making decisions. It executes a block of code only if a specified condition is `true`.

**Syntax:**
```neuroscript
if <expression>
  ... body ...
else
  ... optional else body ...
endif
```

- The expression after `if` is evaluated.
- If it is `true`, the statements in the main body are executed.
- If it is `false`, the statements in the optional `else` block are executed.
- Every `if` statement must be closed with `endif`.
- **Else-if chains** are created by nesting a new `if` statement directly inside an `else` block.

**Examples:**

```neuroscript
# Simple if statement
func check_value(needs x) means
  if x > 100
    emit "The value is large."
  endif
endfunc

# If-else statement
func check_sign(needs x) means
  if x >= 0
    emit "The number is non-negative."
  else
    emit "The number is negative."
  endif
endfunc

# Nested if to create an "else-if" chain
func check_grade(needs score) means
  if score >= 90
    emit "Grade: A"
  else
    if score >= 80
      emit "Grade: B"
    else
      emit "Grade: C or lower"
    endif
  endif
endfunc
```

---

### 6.2. Looping

Loops are used to execute a block of code multiple times. NeuroScript provides two types of loops.

#### 6.2.1. `while` / `endwhile` Loops

A `while` loop repeats a block of statements as long as its condition remains `true`. The condition is checked *before* each iteration.

**Syntax:**
```neuroscript
while <expression>
  ... loop body ...
endwhile
```

```neuroscript
# A loop that counts down from 5
func countdown() means
  set counter = 5
  while counter > 0
    emit counter
    set counter = counter - 1
  endwhile
  emit "Liftoff!"
endfunc
```

#### 6.2.2. `for each` / `in` / `endfor` Loops

A `for each` loop iterates over the elements of a list or other iterable collection. For each element in the collection, it executes the loop body, assigning the current element's value to a temporary variable.

**Syntax:**
```neuroscript
for each <identifier> in <expression>
  ... loop body ...
endfor
```

```neuroscript
# A loop that processes each item in a list
func process_items(needs item_list) means
  for each item in item_list
    call tool.process(item)
    emit "Processed: " + item
  endfor
endfunc
```

---

### 6.3. Modifying Loop Behavior: `break` and `continue`

You can alter the flow of a `while` or `for each` loop using the `break` and `continue` statements.

- **`break`:** Immediately terminates the innermost loop it is in. Execution resumes at the first statement *after* the `endwhile` or `endfor`.
- **`continue`:** Immediately stops the current iteration of the innermost loop and proceeds to the next one. For a `while` loop, it re-evaluates the condition. For a `for each` loop, it moves to the next item in the collection.

```neuroscript
func find_first_admin(needs user_list) means
  set found_admin = nil
  for each user in user_list
    # Skip any users that are not maps
    if typeof(user) != "map"
      continue
    endif

    # If we find an admin, store it and exit the loop
    if user["is_admin"] == true
      set found_admin = user
      break
    endif
  endfor
  return found_admin
endfunc
```

---

### 6.4. Limitations

NeuroScript does not currently implement a `switch` statement for multi-way branching. Complex conditional logic should be handled with `if`/`else` chains.

---

# 7. Scripting Models

NeuroScript files are structured in one of two distinct models, depending on their intended purpose. A single script file must exclusively use one model; they cannot be mixed. This design enforces a clear separation between scripts that execute a direct sequence of commands and those that define reusable libraries of logic.

---

### 7.1. Command Scripts: The Top-Level Execution Block

A **Command Script** is used for direct, immediate execution of a task. It consists of one or more `command` blocks that are executed sequentially by the interpreter. This model is ideal for single-purpose scripts, application entry points, or simple automation tasks.

#### Structure: `command` / `endcommand`

The script's logic is placed inside a block that starts with the `command` keyword and ends with `endcommand`. All statements inside this block are executed from top to bottom.

```neuroscript
:: title: A Simple Command Script

command
  :: description: This command will run immediately.

  set user_name = "guest"
  emit "Command started for user: " + user_name
  call tool.perform_task(user_name)
  emit "Command finished."
endcommand
```

#### Allowed Statements

Command blocks can contain most simple statements (`set`, `call`, `emit`, `fail`, `ask`, `must`) and all control flow blocks (`if`, `while`, `for each`).

Key differences from library functions are:
- You **cannot** use the `return` statement.
- You **cannot** define a `func` within a `command` block.
- Error handling is done via `on error do` blocks.

---

### 7.2. Library Scripts

A **Library Script** is not executed directly. Instead, it serves as a collection of reusable procedures and event handlers that are loaded into an environment. Another script or a host application can then call the functions or trigger the events defined in the library. This model is used to build up a repository of common logic.

A library script consists of any number of `func` blocks and top-level `on event` blocks.

#### Defining Procedures with `func` / `endfunc`

The most common part of a library is the procedure, defined with `func` and `endfunc`. These are named blocks of code that can accept parameters and return values, making them the primary unit of reusable logic.

#### Handling Global Events with `on event`

A library can also define handlers for global events. These `on event` blocks are not attached to a specific function but exist at the top level of the script. When the host system emits a matching event, the code inside the handler is executed.

```neuroscript
:: title: User Utility Library
:: version: 1.2

# A reusable function to format a user's name.
func format_user_name(needs user_map returns formatted_name) means
  set formatted_name = user_map["last_name"] + ", " + user_map["first_name"]
  return formatted_name
endfunc

# A handler that listens for a system-wide user login event.
on event "user.login" as event_data do
  call tool.log.info("User logged in: " + event_data["user_id"])
endon
```
---

# 8. Procedures, Tools, and Calls

While fundamental statements and control flow direct the immediate execution of a script, the true power of NeuroScript comes from creating reusable logic with **procedures** (functions) and interacting with the outside world through **tools**. This section details how to define, call, and structure these powerful constructs.

---

### 8.1. Defining a Function (`func`)

A procedure, or function, is a named block of reusable code defined in a library script. Functions are the primary way to organize and modularize your logic. Every function is defined using the `func` keyword and must end with `endfunc`. The terms "function" and "procedure" are used interchangeably.

**Syntax:**
```neuroscript
func <identifier>(<signature>) means
  ... function body ...
endfunc
```

---

### 8.2. Defining a Signature

A function's **signature** declares the parameters it accepts and the values it returns. The signature is defined within the parentheses `()` following the function's name. It consists of three optional clauses: `needs`, `optional`, and `returns`.

#### 8.2.1. Required Parameters: `needs`

The `needs` clause lists the parameters that are **required** for the function to execute. The caller must provide a value for every parameter in this list.

```neuroscript
func calculate_area(needs width, height) means
  set area = width * height
  return area
endfunc
```

#### 8.2.2. Optional Parameters: `optional`

The `optional` clause lists parameters that are not required. If the caller does not provide a value for an optional parameter, it will default to `nil` inside the function.

```neuroscript
func create_greeting(needs name optional title) means
  if title != nil
    return "Hello, " + title + " " + name
  else
    return "Hello, " + name
  endif
endfunc
```

#### 8.2.3. Return Values: `returns`

The `returns` clause declares the names of the variables that the function will output. These names are used within the function body to assign the results that will be sent back to the caller.

```neuroscript
# This function returns two values, assigned to 'quotient' and 'remainder'.
func divide(needs dividend, divisor returns quotient, remainder) means
  set quotient = dividend / divisor
  set remainder = dividend % divisor
  return quotient, remainder
endfunc
```

---

### 8.3. The `return` Statement

The `return` statement immediately exits the current function and passes values back to the caller. You can return zero or more values. If you return multiple values, they should correspond to the variables declared in the `returns` clause.

**Syntax:** `return [<expression1>, <expression2>, ...]`

---

### 8.4. `callable_expr`: How Functions are Called

A `callable_expr` (callable expression) is the grammar rule for anything that can be invoked with parentheses `()`. This includes user-defined functions, built-in functions, and external tools. The result of a callable expression can be assigned to a variable using `set` or executed for its side effects using `call`.

```neuroscript
# A callable expression whose result is assigned to a variable.
set my_area = calculate_area(10, 5)

# A callable expression executed for its side effects.
call tool.log.info("Calculation complete.")
```

---

### 8.5. Built-in Functions

NeuroScript provides a small set of built-in functions for common operations. These are invoked just like user-defined functions. The list includes:

* `len()`: Returns the length of a string, list, or map.
* `ln()`, `log()`: Mathematical logarithm functions.
* `sin()`, `cos()`, `tan()`: Trigonometric functions.
* `asin()`, `acos()`, `atan()`: Inverse trigonometric functions.

---

### 8.6. External Logic: The `tool` Keyword

The `tool` keyword is the gateway to interacting with the host environment. It allows a script to call external functions, APIs, or any other capability registered with the NeuroScript interpreter.

> **Note on Naming:** The `tool` keyword is a special part of the language for accessing external functions and is not a variable. The dot notation that follows it is a required hierarchical naming convention, not a nested data access path on a variable.

#### 8.6.1. Tool Naming Convention: `tool.<group>.<name>`

Tools are organized into a two-level namespace to prevent collisions and improve clarity. Every tool call must follow this structure:

`tool.<group>.<name>(<arguments>)`

- `tool`: The required keyword.
- `<group>`: A logical grouping for a set of related tools (e.g., `fs` for filesystem, `db` for database). Group names can themselves contain dots for further organization (e.g., `my.corp.utils`).
- `<name>`: The specific name of the tool to be executed.

```neuroscript
# Call a tool in the 'fs' group named 'ReadFile'.
set file_content = tool.fs.ReadFile("/path/to/my/file.txt")

# Call a tool to load another script into the environment.
call tool.script.LoadScript("::title: My Other Lib\nfunc helper() means\n emit 'ok'\nendfunc")
```

---

# 9. Event and Error Handling

A robust script must be able to react to significant occurrences and gracefully manage unexpected problems. NeuroScript provides two distinct, powerful mechanisms for this: a declarative **Event Model** for responding to signals, and a structured **Error Model** for handling runtime failures.

---

### 9.1. The Event Model

The event model allows scripts to react to signals, or "events," that can be triggered by the host system, external tools, or even the script itself using the `emit` statement. This creates a loosely coupled way for different parts of a system to communicate. The event model is **synchronous**; when an `emit` statement is executed, all corresponding `on event` handlers are run to completion before the script continues.

#### 9.1.1. `on event ... do ... endon`

The `on event` block is the core of event handling. It registers a block of code to be executed whenever a matching event occurs. These handlers can be defined at the top level of a library script.

**Syntax:**
`on event <expression> [named <string>] [as <identifier>] do`
  `... handler body ...`
`endon`

- **`<expression>`**: An expression that identifies the event to listen for. This is often a string literal (e.g., `"user.login"`) or a tool call that resolves to an event name.
- **`named <string>`**: (Optional) Assigns a unique name to the handler, allowing it to be specifically cleared later.
- **`as <identifier>`**: (Optional) Captures the payload of the emitted event into a variable that can be used within the handler body.

```neuroscript
# A simple event handler
on event "system.startup" do
  call tool.log.info("System has started.")
endon

# A handler that captures the event's data payload
on event "user.created" as new_user do
  emit "A new user was created with ID: " + new_user["id"]
endon
```

#### 9.1.2. `emit`: Triggering an Event

The `emit` statement, as seen previously, is used to fire an event. It takes an expression as its argument, which becomes the data payload of the event. Any `on event` handler registered for that event name will then be triggered.

#### 9.1.3. `clear event`: Removing Listeners

You can unregister an event handler using the `clear event` statement. This is useful for dynamically managing which events your script should respond to. You can clear by the event name or by the specific name you gave the handler.

```neuroscript
# Clear all handlers for the "system.startup" event
clear event "system.startup"

# Clear a specific handler by its assigned name
on event "user.login" named "MyLoginHandler" do
  # ...
endon

clear event named "MyLoginHandler"
```

---

### 9.2. The Error Model

The error model provides a structured way to handle runtime failures, such as a tool failing or a `must` statement evaluating to false.

#### 9.2.1. `on error ... do ... endon`

The `on error` block defines a set of statements to execute when a failure occurs within its scope. This block can be defined inside any `func` or `command` block.

**Syntax:**
`on error do`
  `... error handler body ...`
`endon`

When an error occurs, the interpreter immediately stops normal execution and jumps to the nearest enclosing `on error` block.

#### 9.2.2. `fail`: Triggering an Error

The `fail` statement explicitly triggers an error, immediately halting the current flow and activating the error handling mechanism.

#### 9.2.3. `clear_error`: Resetting the Error State

Inside an `on error` block, you have two choices:
1.  **Let the script terminate:** If you do nothing, the script will exit after the `on error` block finishes.
2.  **Handle the error and continue:** If you want the script to recover and continue execution from the statement following the handler, you must call `clear_error`. This statement resets the interpreter's error state.

```neuroscript
func process_file(needs file_path) means
  on error do
    emit "Failed to process file: " + file_path
    # We don't clear the error, so the function will exit after this.
  endon

  set content = tool.fs.read(file_path) # This tool might fail
  emit "File read successfully."
  # ... more processing ...
endfunc

func resilient_task() means
  on error do
    emit "A recoverable error occurred. Continuing."
    clear_error  # The error is handled; the script will not terminate.
  endon

  fail "This is a test failure."

  emit "This line will be executed because the error was cleared."
endfunc
```
---

# 10. Advanced Operators and Reserved Keywords

Beyond the fundamental statements and expressions, NeuroScript includes a few specialized operators for introspection and convenience, as well as keywords that are reserved for future language features. This section covers these more nuanced parts of the language.

---

### 10.1. Type Introspection with `typeof`

The `typeof` operator is a unary operator that returns the data type of its operand as a string. This is useful for performing checks and validations at runtime, allowing your script to handle different types of data dynamically.

**Syntax:** `typeof <expression>`

The strings returned by `typeof` include: `"string"`, `"number"`, `"boolean"`, `"list"`, `"map"`, `"nil"`, `"function"`, `"tool"`, and more.

```neuroscript
func process_value(needs val) means
  set value_type = typeof val

  if value_type == "string"
    emit "The value is a string with length: " + len(val)
  else
    if value_type == "list"
      emit "The value is a list with " + len(val) + " items."
    else
      emit "The value is of an unhandled type: " + value_type
    endif
  endif
endfunc
```

---

### 10.2. The `some` and `no` Operators

The `some` and `no` operators provide a highly readable, "syntactic sugar" way to check if a list is empty or `nil`.

- **`some`**: Returns `true` if the list exists and contains at least one element. It is equivalent to `(my_list != nil and len(my_list) > 0)`.
- **`no`**: Returns `true` if the list is `nil` or has a length of zero. It is the direct opposite of `some`.

**Syntax:** `some <expression>` or `no <expression>`

```neuroscript
func check_list(needs items) means
  if some items
    emit "The list contains items."
    # Proceed to iterate or process the list
    for each i in items
      # ...
    endfor
  endif

  if no items
    emit "The list is empty or nil."
  endif
endfunc
```

---

### 10.3. Reserved Keywords for Future Use

Some words are reserved as keywords in the NeuroScript grammar to ensure they are available for planned future features. While they will be parsed correctly, they currently have no implemented functionality in the interpreter.

#### 10.3.1. Type Assertion: `mustbe`

The `mustbe` keyword is reserved for a future type assertion system. The intended functionality is to provide a way to enforce that a variable is of a certain type, similar to how `must` enforces a boolean condition. This keyword is defined in the grammar but is not used in any parser rules.

----
Metadata

::schema: spec
::serialization: md
::langVersion: neuroscript@0.4.6
::fileVersion: 3
::description: A comprehensive guide to the NeuroScript language, covering syntax, data types, control flow, and advanced features.
::author: Gemini and Andrew Price
::created: 2025-07-13
::modified: 2025-08-23
::license: Proprietary
::tags: neuroscript,guide,documentation,spec,language
::type: documentation
::subtype: language_guide
::dependsOn: NeuroScript.g4.txt
::howToUpdate: Update sections to match new interpreter features or grammar changes. Increment fileVersion.