Of course. Here is a brief guide for accessing the newly preserved source code information from the NeuroScript AST.

-----

### Guide to Accessing Source Code Formatting from the NeuroScript AST

This guide explains how to access metadata, comments, and blank line information that is now preserved in the NeuroScript Abstract Syntax Tree (AST). This data is crucial for reconstructing the original source code with perfect fidelity.

-----

#### **1. File-Level Metadata and Comments**

File-level metadata and any comments that appear before the first code block are attached directly to the root `ast.Program` node.

  * **File Metadata**: Stored in the `Metadata` map on the `Program` struct.
  * **Header Comments**: Stored in the `Comments` slice on the `Program` struct.

**Example:**

```go
// Assuming 'program' is your *ast.Program instance
fileAuthor := program.Metadata["author"]
fileVersion := program.Metadata["version"]

fmt.Println("File-level comments:")
for _, comment := range program.Comments {
    fmt.Println(comment.Text)
}
```

-----

#### **2. Block-Level Metadata, Comments, and Blank Lines**

Declarative blocks like **Procedures**, **Commands**, and **Events** also store their own metadata, comments, and preceding blank lines.

  * **`ast.Procedure`**: Contains `Metadata`, `Comments`, and `BlankLinesBefore` fields.
  * **`ast.CommandNode`**: Contains `Metadata`, `Comments`, and `BlankLinesBefore` fields.
  * **`ast.OnEventDecl`**: Contains a `BlankLinesBefore` field.

**Example:**

```go
// Assuming 'proc' is an *ast.Procedure instance
fmt.Printf("Blank lines before procedure: %d\n", proc.BlankLinesBefore)
purpose := proc.Metadata["purpose"]

fmt.Println("Procedure-level comments:")
for _, comment := range proc.Comments {
    fmt.Println(comment.Text)
}
```

-----

#### **3. Statement-Level Comments and Blank Lines**

Every statement within a block is represented by an `ast.Step` struct. Each `Step` node captures any comments and blank lines that appeared immediately before it in the source code.

  * **`ast.Step`**: Contains `Comments` and `BlankLinesBefore` fields.

**Example:**

```go
// Assuming 'proc' is an *ast.Procedure instance
for _, step := range proc.Steps {
    fmt.Printf("Blank lines before this step: %d\n", step.BlankLinesBefore)
    if len(step.Comments) > 0 {
        fmt.Println("Comments for this step:")
        for _, comment := range step.Comments {
            fmt.Println(comment.Text)
        }
    }
}
```