# NeuroScript: A Pseudocode Framework for AI Reasoning

by ChatGPT-o1 and Andrew Price

NeuroScript is a structured, human-readable language that provides a *procedural scaffolding* for large language models (LLMs). It is designed to store, discover, and reuse **"skills"** (procedures) with clear docstrings and robust metadata, enabling AI systems to build up a library of **reusable, well-documented knowledge**.

## 1. Goals and Principles

1. **Explicit Reasoning**: Rather than relying on hidden chain-of-thought, NeuroScript encourages step-by-step logic in a code-like format that is both *executable* and *self-documenting*.

2. **Reusable Skills**: Each procedure is stored and can be retrieved via a standard interface. LLMs or humans can then call, refine, or extend these procedures without re-implementing from scratch.

3. **Self-Documenting**: NeuroScript procedures must include docstrings that clarify *purpose*, *inputs*, *outputs*, *algorithmic rationale*, and *edge cases*—mirroring how humans comment code.

4. **LLM Integration**: NeuroScript natively supports calling LLMs for tasks that benefit from free-form generation, pattern matching, or advanced “human-like” reasoning. 

5. **Multi-Modal Reasoning**: The language can incorporate constructs for deductive logic (assertions), inductive inference (via LLM calls), reflection, and more.

---

## 2. Language Constructs

### 2.1 High-Level Structure

A NeuroScript file (or “skill” definition) typically contains:
1. **DEFINE PROCEDURE** *Name*( *Arguments* )
2. **COMMENT** block (Docstring)
3. **Statements** (the pseudocode body)
4. **END** to close out definitions or blocks

**Example**:
```
DEFINE PROCEDURE WeightedAverage(ListOfNumbers)
COMMENT:
    PURPOSE: Compute weighted average of list items {value, weight}.
    INPUTS: ListOfNumbers -> Each item has .value and .weight
    OUTPUT: A single numeric average
    ALGORITHM:
        1. Sum up (value * weight) for each item.
        2. Divide total by sum of weights.
    CAVEATS: Returns 0 if weight sum is zero.
    EXAMPLES:
        WeightedAverage([{value:3, weight:2}, {value:5, weight:6}]) => 4.5
END

SET total = 0
SET weightSum = 0

FOR EACH item IN ListOfNumbers DO
    SET total = total + item.value * item.weight  # Accumulate weighted sum
    SET weightSum = weightSum + item.weight       # Accumulate total weight
END

IF weightSum = 0 THEN
    RETURN 0
ELSE
    RETURN total / weightSum
END
```

### 2.2 Built-In Statements

- **SET var = expr**  
  Assigns the result of an expression to a variable.

- **IF condition THEN ... END**  
  Basic conditional control flow.

- **WHILE condition DO ... END** / **FOR EACH x IN collection DO ... END**  
  Iteration constructs.

- **CALL**  
  - `CALL AnotherProcedure(args...)` – calls another NeuroScript procedure.
  - `CALL LLM("prompt")` – delegates a subtask or prompt to the LLM and captures its response.
  - `CALL TOOL.SomeExternalFunction(...)` – calls an external tool/function if integrated.

- **RETURN value**  
  Returns a value from the current procedure.

- **COMMENT**  
  A block or inline annotation to clarify “why” behind each step.

### 2.3 Docstrings (Structured Comments)

NeuroScript **requires** a docstring block at the top of each procedure. It can be free-form, but strongly recommended to include:

- `PURPOSE:` Short statement of what this procedure does.  
- `INPUTS:` Parameter list with explanations.  
- `OUTPUT:` Return value or result.  
- `ALGORITHM:` High-level summary of the logic or approach.  
- `CAVEATS:` Edge cases, performance notes, limitations.  
- `EXAMPLES:` At least one example input and output.

Example:
```
COMMENT:
    PURPOSE: ...
    INPUTS: ...
    OUTPUT: ...
    ALGORITHM: ...
    CAVEATS: ...
    EXAMPLES: ...
END
```

**In-line Comments** are also encouraged to explain the rationale behind specific code blocks.

---

## 3. Storing and Discovering Procedures

### 3.1 Skill Registry Schema

You need a repository or database where each NeuroScript procedure (“skill”) is stored, typically with:

- **name** (unique identifier)  
- **docstring** (text metadata)  
- **neuroscript_code** (the body of the pseudocode)  
- **version** or **timestamp**  
- Possibly **embeddings** (for semantic search)

### 3.2 Retrieval & Discovery

- **Vector Search**: The docstring (and possibly the code) is embedded and stored in a vector DB or in a Postgres extension (e.g., pgvector).  
- **Keyword/Full-Text Search**: You might also use standard text search for simpler queries.  

A query like “Need a function for advanced sentiment analysis” can match docstrings that mention “sentiment analysis.”

### 3.3 API or Functions

If you have a service-based approach:
- `search_procedures(query) -> list of matches`
- `get_procedure(name) -> returns docstring + code`
- `save_procedure(name, docstring, code) -> updates or creates skill`

If you use a local approach, these can be simple library calls or direct SQL statements.

---

## 4. Interfacing with LLMs

### 4.1 `CALL LLM("prompt")`

A built-in statement that:

1. Takes a string prompt.  
2. Sends it to an LLM gateway or API endpoint.  
3. Returns the raw text response, which can then be parsed or stored by NeuroScript.

**Example**:
```
SET analysis = CALL LLM("Analyze the following text for sentiment: {{text}}")
IF analysis CONTAINS "negative" THEN
    EMIT "The text is negative in tone."
END
```

### 4.2 Variation: Provide Context or Additional Instructions

NeuroScript might support `CALL LLM_WITH_CONTEXT(contextData, "prompt")`. The interpreter can embed `contextData` into the LLM prompt (for more advanced usage).

---

## 5. Built-In Reasoning Constructs

NeuroScript aims to support multiple forms of reasoning, akin to human cognition:

1. **Deductive** – Use `ASSERT`, `VERIFY`, or explicit logic checks:
   ```
   ASSERT user_age > 0 => OnFailure: EMIT "Invalid age!"
   ```

2. **Inductive/Abductive** – Typically requires free-form pattern recognition:
   - `CALL LLM(...)` to interpret data, propose hypotheses, or generalize.

3. **Heuristic** – For quick guesses or fallback solutions. Possibly stored as:
   ```
   DEFINE PROCEDURE HeuristicGuess( situation )
       # Some simpler approach, or calls to LLM with "Provide your best guess"
   END
   ```

4. **REFLECT** – A special block that can re-check or refine the code’s logic, possibly re-calling the LLM for a meta-analysis. (Optional advanced feature.)

---

## 6. Example Workflow

1. **User / System**: Needs a skill to “Classify text by emotion.”  
2. **LLM**:
   - **Search** the registry for anything referencing “emotion classification.”  
   - If found, calls the existing procedure. If not, defines a new one, e.g.:
     ```
     DEFINE PROCEDURE ClassifyEmotion(TextInput)
     COMMENT:
         PURPOSE: ...
         INPUTS: TextInput -> the text to be classified
         OUTPUT: label (string) among [happy, sad, fear, etc.]
         ALGORITHM:
             1. Use LLM to parse the emotion in text.
             2. Return the best match.
         EXAMPLES: ...
     END

     SET rawResponse = CALL LLM("Given this text, what emotion is it primarily expressing: {{TextInput}}")
     # parse rawResponse for best label
     RETURN label
     ```
   - **Store** the new procedure (docstring + code) in the registry for future use.

3. **Runtime** executes the procedure: It calls the LLM, obtains a label, and returns it.

4. **Feedback / Revision**: If it’s inaccurate, the LLM or user updates the procedure docstring or logic. Over time, the library evolves.

---

## 7. Implementation and Architecture

### 7.1 NeuroScript Interpreter

- **Parsing**: A minimal grammar to convert NeuroScript text into an internal AST or direct evaluation.  
- **Execution**: Step-by-step runs statements (SET, IF, CALL, etc.).  
- **Error Handling**: Logs or throws errors if a procedure references undefined variables or tools.

### 7.2 Database / Store

- **Relational** (e.g., PostgreSQL + pgvector) or
- **Dedicated Vector DB** (like Pinecone, Qdrant, Weaviate)  
- **Version Control** (optional): Keep each procedure in Git if you want code-like merges and diffs.

### 7.3 LLM Gateway

- A simple API or library call that NeuroScript uses to send prompts and receive responses.  
- Could incorporate specialized prompts or instruction presets for different tasks.

---

## 8. Summary and Future Directions

- **NeuroScript** is a **structured pseudocode** layer that fosters explicit, procedural reasoning and skill accumulation.  
- **Docstrings** are central—procedures are self-documenting, enabling better discoverability and maintenance.  
- **Store/Discover/Retrieve** is crucial: we keep all procedures in a robust repository and let LLMs “look them up” via semantic or keyword queries.  
- **LLM Integration** is first-class: NeuroScript seamlessly delegates “complex text reasoning” tasks back to the LLM.  
- **Reasoning Modes** can be expanded as needed, from deduce-and-verify to reflection-based self-checking.  

Future expansions might include concurrency models, advanced error handling (TRY/CATCH), refined reflection blocks, and specialized data structures for domain-specific tasks. But the **core** is consistent: encourage **structured, documented, reusable** AI reasoning code.

---

### That’s the NeuroScript Spec v0.1!

Use it as a foundation to build prototypes—test out storing procedures, calling an LLM in pseudocode, and refining those skills over time. Then iterate on the design as you gather real-world feedback.