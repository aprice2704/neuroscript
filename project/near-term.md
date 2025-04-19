# Near-Term TODO

Based on the refactoring checklist and the requirements for autonomous operation on potentially large codebases (synced via the File API), here's a list of necessary tools, categorized for clarity:

1. Filesystem Operations (Core Primitives & Refactoring Support)

These are fundamental for manipulating the codebase structure.

TOOL.ReadFile(path): (Exists) Read content for analysis/modification.
TOOL.WriteFile(path, content): (Exists) Write modified code back.
TOOL.ListDirectory(path, [recursive], [pattern]): (Exists, but needs enhancement) List files/directories to operate on. Needs recursive (boolean) and pattern (string, e.g., *.go) options to effectively handle large trees en masse.
TOOL.Mkdir(path): (Exists) Create new directories for moved packages/files.
TOOL.DeleteFile(path): (Exists) Remove original files after moving.
TOOL.MoveFile(source, destination): (Specced) Crucial for renaming/moving files or directories during refactoring. Needs to be implemented.
2. Go Code Analysis & Manipulation (AST Tools)

These are essential for understanding and modifying Go code structure safely.

TOOL.GoParseFile(path or content): (Exists ) Parse Go code into an Abstract Syntax Tree (AST) representation.
TOOL.GoFindIdentifiers(ast_handle, pkg_name, identifier): (Exists ) Find where specific functions/types are used.
TOOL.GoModifyAST(ast_handle, modifications): (Exists, needs robust sub-operations ) Modify the AST based on refactoring rules. Key modifications needed include:
Change Package Declaration
Add/Remove/Replace Import Paths
Replace Qualified Identifiers (e.g., change oldpkg.MyType to newpkg.MyType)
TOOL.GoFormatASTNode(ast_handle): (Exists ) Format the modified AST back into Go code string.
TOOL.GoUpdateImportsForMovedPackage(...): (Specced) An intelligent, higher-level tool specifically for fixing imports across the codebase after packages/symbols have been moved. Needs implementation.
3. Build & Verification Tools

To ensure the refactoring didn't break anything.

TOOL.GoBuild([target]): (Exists via Shell ) Compile the code.
TOOL.GoTest(): (Exists via Shell ) Run unit tests.
TOOL.GoCheck([target]): (Exists via Shell ) Run static analysis checks.
TOOL.GoModTidy(): (Exists via Shell ) Tidy the go.mod file.
4. Version Control Tools (Git)

Essential for managing changes safely during autonomous refactoring.

TOOL.GitAdd(path): (Exists ) Stage changes.
TOOL.GitCommit(message): (Exists ) Commit changes.
TOOL.GitNewBranch(branch_name): (Needed ) Create branches for refactoring work.
TOOL.GitCheckout(branch_name): (Needed ) Switch branches.
TOOL.GitStatus(): (Needed) Check the working tree status.
TOOL.GitPull(): (Needed) Update from remote.
TOOL.GitPush(): (Needed) Push changes.
TOOL.GitDiff(): (Needed) Check differences.
TOOL.GitRm(path): (Needed ) Remove files from Git tracking (after moving/deleting).
5. File Synchronization Tools (Interacting with Gemini File API)

These support the "en masse" aspect by ensuring the local workspace reflects the intended state and can interact with files uploaded via the API.

TOOL.SyncFiles(direction, localDir, [filterPattern]): (New) A high-level tool encapsulating the logic from the gensync program we built.
direction (enum: "up", "down"): "up" syncs local changes to the API; "down" syncs API changes to local (more complex).
localDir (string): The local directory root to sync.
filterPattern (string, optional): Glob pattern to include/exclude files.
This would handle reading the local index, comparing state (hash/mod time), and performing necessary API uploads/downloads/deletes.
TOOL.UploadFile(localPath, [displayName]): (New, mirrors helper) A lower-level tool to upload a single file and return its API Name and URI.
TOOL.ListAPIFiles(): (New, mirrors original script) List files currently stored in the API associated with the key.
TOOL.DeleteAPIFile(apiFileName): (New, mirrors API call) Delete a specific file from the API using its files/... name.
6. User Interaction / Control (Optional but recommended)

IO.Input(prompt): (Exists, though potentially under TOOL.) Allows the script to pause and ask for human confirmation before potentially destructive steps (like large refactors or Git operations).
This list provides a comprehensive set of capabilities needed for the kind of autonomous refactoring you're envisioning, covering file manipulation, code understanding, verification, versioning, interaction with the File API en masse, and optional human oversight. Many primitives exist, but key refactoring helpers (MoveFile, GoUpdateImports...), enhanced Git tools, and dedicated file sync tools would need to be developed.

