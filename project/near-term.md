:: title: Near-Term TODO Checklist
:: status: draft
:: version: 0.1.3
:: id: near-term-checklist-v0.1
:: derivedFrom: project/near-term.md (v unspecified)
:: description: Checklist derived from the near-term TODO markdown file, tracking tool implementation status based on Gemini's awareness during documentation session on Apr 22, 2025.

# Near-Term TODO Checklist

## 1. Filesystem Operations (Core Primitives & Refactoring Support)
  [x] FS.ReadFile(path): Read content for analysis/modification.
  [x] FS.WriteFile(path, content): Write modified code back.
  [x] FS.ListDirectory(path, [recursive], [pattern]): List files/directories. (Needs enhancement for pattern?)
  [x] FS.Mkdir(path): Create new directories.
  [x] FS.DeleteFile(path): Remove original files after moving.
  [x] FS.MoveFile(source, destination): Crucial for renaming/moving files or directories.
  [x] FS.LineCountFile(path): Count lines in a file. (Implied existence via documentation)
  [x] FS.SanitizeFilename(name): Clean string for filename use. (Implied existence via documentation)
  [x] FS.WalkDir(path): Recursively walk directory tree. (Implied existence via documentation)
  [x] FS.FileHash(path): Calculate SHA256 hash of file. (Implied existence via documentation)

## 2. Go Code Analysis & Manipulation (AST Tools)
  [x] GoAST.GoParseFile(path or content): Parse Go code into an AST representation.
  [x] GoAST.GoFindIdentifiers(ast_handle, pkg_name, identifier): Find where specific functions/types are used.
  [x] GoAST.GoModifyAST(ast_handle, modifications): Modify the AST. (Marked Exists, assuming core functionality is present).
    [ ] Sub-op: Change Package Declaration
    [ ] Sub-op: Add/Remove/Replace Import Paths
    [ ] Sub-op: Replace Qualified Identifiers
  [x] GoAST.GoFormatASTNode(ast_handle): Format the modified AST back into Go code string.
  [ ] GoAST.GoUpdateImportsForMovedPackage(...): Intelligent tool for fixing imports after moves. (Marked Specced)

## 3. Build & Verification Tools
  [x] Shell.GoBuild([target]): Compile the code.
  [x] Shell.GoTest(): Run unit tests.
  [x] Shell.GoCheck([target]): Run static analysis checks.
  [x] Shell.GoModTidy(): Tidy the go.mod file.

## 4. Version Control Tools (Git)
  [x] git.Add(paths): Stage changes.
  [x] git.Commit(message): Commit changes.
  [x] git.NewBranch(branch_name): Create branches for refactoring work.
  [x] git.Checkout(branch_name): Switch branches.
  [x] git.Status(): Check the working tree status.
  [x] git.Pull(): Update from remote.
  [x] git.Push(): Push changes.
  [x] git.Diff(): Check differences.
  [x] git.Rm(path): Remove files from Git tracking.

## 5. File Synchronization Tools (Interacting with Gemini File API)
  [ ] FileAPI.SyncFiles(direction, localDir, [filterPattern]): High-level sync tool. (Marked New)
  [ ] FileAPI.UploadFile(localPath, [displayName]): Lower-level single file upload. (Marked New)
  [ ] FileAPI.ListAPIFiles(): List files in API. (Marked New)
  [ ] FileAPI.DeleteAPIFile(apiFileName): Delete file from API. (Marked New)

## 6. User Interaction / Control
  [x] IO.Input(prompt): Allows script to pause and ask for human confirmation.

## 7. Agent Architecture & Core Enhancements
  [ ] Agent Startup Script (`agent_startup.ns`): Replace flags with scriptable config.
    [ ] Requires new `TOOL.Agent*` config tools (AgentSetSandbox, AgentPinFile, AgentSetModel, etc.).
  [ ] `AgentContext` Object (`pkg/neurogo`): Central struct for agent state.
  [ ] Typed Handles (`category::uuid`): Implement prefix system for runtime handle type safety.
  [ ] Dual Context Management Strategy:
    [ ] Pinning (`TOOL.AgentPinFile`) + Temp Request (`TOOL.RequestFileContext`) implementation.
    [ ] AI Forgetting (`TOOL.Forget`/`TOOL.ForgetAll`) implementation.