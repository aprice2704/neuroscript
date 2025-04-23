:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-vector-vectorupdate-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_vector.go, docs/script_spec.md
:: relatedTo: Vector.SearchSkills
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `Vector.VectorUpdate` (v0.1)

* **Tool Name:** `Vector.VectorUpdate` (v0.1)
* **Purpose:** Reads the content of a specified file (presumably a skill file) within the sandbox, generates a vector embedding for that content, and then adds or updates this embedding in the interpreter's internal (mock) vector index. This makes the file discoverable via `Vector.SearchSkills`.
* **NeuroScript Syntax:** `CALL Vector.VectorUpdate(filepath: <String>)`
* **Arguments:**
    * `filepath` (String): Required. The relative path (within the sandbox) of the file whose content should be indexed.
* **Return Value:** (String)
    * On success: Returns the literal string `"OK"`. (Accessible via `LAST` after the `CALL`).
    * On error (e.g., path validation fails, file read error, embedding generation fails): Returns an error message string describing the failure. (Accessible via `LAST` after the `CALL`).
* **Behavior:**
    1.  Validates that exactly one argument (`filepath`) of type String is provided.
    2.  Retrieves the interpreter's `sandboxDir`.
    3.  Uses `SecureFilePath` to validate the `filepath` argument against the `sandboxDir`. If validation fails, returns an error message string.
    4.  Reads the entire content of the file at the validated absolute path using `os.ReadFile`. If reading fails, returns an error message string.
    5.  Calls the interpreter's `GenerateEmbedding` function to get a vector embedding for the file's content string. If this fails, returns an error message string.
    6.  Initializes the interpreter's internal `vectorIndex` (a map of absolute file paths to embeddings) if it's not already initialized.
    7.  Stores the generated embedding in the `vectorIndex` map, using the *absolute path* of the file as the key. If an entry for this absolute path already exists, it is overwritten with the new embedding.
    8.  Returns the string `"OK"` to indicate successful indexing/updating.
* **Security Considerations:**
    * Uses `SecureFilePath` to ensure the file being read and indexed is within the sandbox.
    * Relies on `interpreter.GenerateEmbedding`, which typically involves sending the *entire file content* to an external service (like an LLM API). This has significant privacy implications depending on the file content and the embedding service used. Ensure the embedding provider's privacy policy is acceptable.
    * Modifies the interpreter's internal state (`vectorIndex`).
* **Examples:**
    ```neuroscript
    # Example 1: Index a newly created skill file
    SET skill_content = """
    DEFINE PROCEDURE CalculateArea(width, height)
    COMMENT:
        PURPOSE: Calculates the area of a rectangle.
        INPUTS: - width: number
                - height: number
        OUTPUT: number (the calculated area)
        ALGORITHM: Multiply width by height.
    ENDCOMMENT
    # SET area = width * height # Requires Math tool
    # RETURN area
    RETURN "Calculation placeholder" # Placeholder
    END
    """
    CALL FS.WriteFile("library/geometry/area.ns.txt", skill_content)
    SET write_status = LAST

    IF write_status == "OK" THEN
      EMIT "Indexing skill file: library/geometry/area.ns.txt"
      CALL Vector.VectorUpdate("library/geometry/area.ns.txt")
      SET index_status = LAST
      EMIT "Indexing status: " + index_status # Expect "OK"
    ELSE
      EMIT "Failed to write skill file, cannot index."
    ENDBLOCK

    # Example 2: Attempt to index a non-existent file
    CALL Vector.VectorUpdate("non_existent_skill.ns.txt")
    SET non_exist_status = LAST
    IF non_exist_status != "OK" THEN
      EMIT "Indexing failed for non-existent file as expected: " + non_exist_status
    ELSE
      EMIT "Indexing unexpectedly succeeded for non-existent file."
    ENDBLOCK

    # Example 3: Search after indexing (using example from SearchSkills)
    CALL Vector.SearchSkills("calculate rectangle area")
    SET search_results_json = LAST
    EMIT "Search results after indexing: " + search_results_json
    # Should hopefully now include 'library/geometry/area.ns.txt' if embedding works well
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_vector.go`
    * Function: `toolVectorUpdate`
    * Spec Name: `VectorUpdate`
    * Key Go Packages: `fmt`, `os`, `path/filepath`
    * Relies heavily on `interpreter.GenerateEmbedding` (external call likely) and modifies `interpreter.vectorIndex` (internal state). Uses `core.SecureFilePath`.
    * Registration: Registered by `registerVectorTools` within `pkg/core/tools_register.go`. Returns `string, error`.