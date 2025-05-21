:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-vector-searchskills-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_vector.go, docs/script_spec.md
:: relatedTo: Vector.VectorUpdate
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `Vector.SearchSkills` (v0.1)

* **Tool Name:** `Vector.SearchSkills` (v0.1)
* **Purpose:** Performs a similarity search against a (mock) vector index of indexed skill files. It takes a natural language query, generates an embedding for it, compares it to stored embeddings, and returns a ranked list of matching skill files based on cosine similarity.
* **NeuroScript Syntax:** `CALL Vector.SearchSkills(query: <String>)`
* **Arguments:**
    * `query` (String): Required. The natural language text query used to find relevant skills.
* **Return Value:** (String)
    * On success: A JSON-formatted string representing a list of matching skills. (Accessible via `LAST` after the `CALL`).
        * Each item in the list is an object (Map) with two keys:
            * `path` (String): The relative path (within the sandbox) of the skill file that matched.
            * `score` (Number): The cosine similarity score between the query and the skill file (typically between 0.0 and 1.0, higher means more similar).
        * The list is sorted in descending order by `score`.
        * Returns an empty JSON list (`"[]"`) if no skills meet the similarity threshold (currently 0.5 in the implementation).
    * On failure (e.g., error generating query embedding): Returns an error message string (e.g., "SearchSkills embedding generation failed: ..."). Internal errors (like JSON marshalling) might cause a script execution error. (Accessible via `LAST` after the `CALL`).
* **Behavior:**
    1.  Validates that exactly one argument (`query`) of type String is provided.
    2.  Initializes the interpreter's internal `vectorIndex` (a map of absolute file paths to embeddings) if it's not already initialized.
    3.  Calls the interpreter's `GenerateEmbedding` function to get a vector embedding for the input `query`. If this fails, returns an error message string.
    4.  Iterates through each entry (absolute path -> stored embedding) in the `vectorIndex`.
    5.  Calculates the cosine similarity score between the query embedding and the stored embedding. Skips the entry if calculation fails.
    6.  If the calculated score meets or exceeds a predefined threshold (0.5), it proceeds.
    7.  Converts the absolute path (from the index key) back to a relative path based on the interpreter's `sandboxDir`.
    8.  Creates a result Map: `{"path": relative_path, "score": score}`.
    9.  Appends this Map to a list of results.
    10. After checking all entries, sorts the list of result Maps in descending order based on the `score`.
    11. Marshals the sorted list into a JSON string. If marshalling fails, an internal error occurs.
    12. Returns the JSON string representing the sorted list of results.
* **Security Considerations:**
    * Relies on `interpreter.GenerateEmbedding`, which typically involves sending the query text to an external service (like an LLM API). Consider the privacy implications of the query content being sent externally.
    * The tool itself reads from an in-memory index (`interpreter.vectorIndex`) and does not directly perform file I/O during the search operation.
    * The conversion from absolute paths (stored in the index) back to relative paths depends on the correct configuration of the interpreter's `sandboxDir`.
* **Examples:**
    ```neuroscript
    # Prerequisite: Assume VectorUpdate has been called previously for some files
    # CALL Vector.VectorUpdate("library/math_helpers.ns.txt")
    # CALL Vector.VectorUpdate("library/string_utils.ns.txt")

    # Example 1: Search for math-related skills
    SET search_query = "How do I add two numbers?"
    CALL Vector.SearchSkills(search_query)
    SET search_results_json = LAST

    EMIT "Search Results for '" + search_query + "':"
    EMIT search_results_json
    # Example Output (JSON string): "[{\"path\":\"library/math_helpers.ns.txt\",\"score\":0.82}]"
    # Note: Actual score and results depend heavily on the embedding model and indexed content.

    # Example 2: Search for something unlikely to match
    CALL Vector.SearchSkills("Tell me about ancient history")
    SET no_match_json = LAST
    EMIT "Search results for unlikely query: " + no_match_json # Expect "[]"

    # Example 3: (Illustrative) Process results if JSON parsing were available
    # SET results_list = JSON.Parse(search_results_json) # Assumes JSON.Parse tool
    # IF List.Length(results_list) > 0 THEN
    #   SET top_result = results_list[0]
    #   EMIT "Top result path: " + top_result["path"] + " with score: " + top_result["score"]
    # ELSE
    #   EMIT "No relevant skills found."
    # ENDBLOCK
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_vector.go`
    * Function: `toolSearchSkills`
    * Spec Name: `SearchSkills`
    * Key Go Packages: `encoding/json`, `fmt`, `path/filepath`, `sort`
    * Relies heavily on `interpreter.GenerateEmbedding` (external call likely) and `interpreter.vectorIndex` (internal state). Also uses an internal `cosineSimilarity` helper.
    * Registration: Registered by `registerVectorTools` within `pkg/core/tools_register.go`. Returns `string, error`.