:: type: Checklist
:: version: 0.1.1  // Updated version
:: id: autonomous-refactor-tools-todo-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/index.md, pkg/core/tools_register.go
:: howToUpdate: Update status ([ ], [-], [x]) as tools are specified and implemented. Add new tools if requirements change.

# Checklist: Tools Needed for Autonomous Refactoring

- | | Autonomous Refactoring Tools [ Overall Status ]
  - |-| Filesystem Operations
    - [x] TOOL.ReadFile(path)
    - [x] TOOL.WriteFile(path, content)
    - [-] TOOL.ListDirectory(path, [recursive], [pattern]) // Base implemented
      - [ ] Add recursive option implementation
      - [ ] Add pattern filtering implementation
    - [x] TOOL.Mkdir(path)
    - [x] TOOL.DeleteFile(path)
    - [ ] TOOL.MoveFile(source, destination)
      - [x] Specification Exists ([docs/ns/tools/move_file.md](../ns/tools/move_file.md))
      - [ ] Go Implementation
  - |x| Go Code Analysis & Manipulation (AST Tools)
    - [x] TOOL.GoParseFile(path or content)
    - [x] TOOL.GoFindIdentifiers(ast_handle, pkg_name, identifier)
    - [x] TOOL.GoModifyAST(ast_handle, modifications) // Core and listed sub-ops implemented
      - [x] - Change Package Declaration
      - [x] - Add/Remove/Replace Import Paths
      - [x] - Replace Qualified Identifiers
    - [x] TOOL.GoFormatASTNode(ast_handle)
    - [ ] TOOL.GoUpdateImportsForMovedPackage(...)
      - [x] Specification Exists ([docs/ns/tools/go_update_imports_for_moved_package.md](../ns/tools/go_update_imports_for_moved_package.md))
      - [ ] Go Implementation
  - |x| Build & Verification Tools
    - [x] TOOL.GoBuild([target])
    - [x] TOOL.GoTest()
    - [x] TOOL.GoCheck([target])
    - [x] TOOL.GoModTidy()
  - |-| Version Control Tools (Git) // Updated based on code review
    - [x] TOOL.GitAdd(path)
    - [x] TOOL.GitCommit(message)
    - [x] TOOL.GitNewBranch(branch_name) // Was [ ], now [x]
    - [x] TOOL.GitCheckout(branch_name)  // Was [ ], now [x]
    - [ ] TOOL.GitStatus()
    - [ ] TOOL.GitPull()
    - [ ] TOOL.GitPush()
    - [ ] TOOL.GitDiff()
    - [x] TOOL.GitRm(path) // Was [ ], now [x]
  - |-| File Synchronization Tools (Gemini File API)
    - [ ] TOOL.SyncFiles(direction, localDir, [filterPattern])
    - [ ] TOOL.UploadFile(localPath, [displayName])
    - [ ] TOOL.ListAPIFiles()
    - [ ] TOOL.DeleteAPIFile(apiFileName)
  - |x| User Interaction / Control
    - [x] IO.Input(prompt)