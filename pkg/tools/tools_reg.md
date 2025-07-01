:: version: 1.0.0
:: title: NeuroScript Core Tool Registration Checklist
:: generated: 2025-05-06
:: updated: 2025-05-06

# Core Tool Registration Status (for init-in-own-file pattern)

## File System Tools (`pkg/core/tools_fs_*.go`) in tooldefs_fs.go
- [x] ReadFile
- [x] WriteFile
- [x] DeleteFile
- [x] ListDirectory
- [x] Mkdir
- [x] FileHash
- [x] MoveFile
- [x] FileStat
- [ ] WalkFiles
- [x] WalkDir
- [ ] GetOS
- [ ] GetCWD
- [ ] PathExists
- [ ] IsDir
- [ ] IsFile

## Git Tools (`pkg/core/tools_git.go`, `pkg/core/tools_git_status.go`)
- [x] GitAdd # Has init in tools_git.go
- [x] GitCommit # Has init in tools_git.go
- [x] GitNewBranch # Has init in tools_git.go
- [x] GitCheckout # Has init in tools_git.go
- [x] GitDiff # Has init in tools_git.go
- [x] GitMerge # Has init in tools_git.go
- [x] GitPull # Has init in tools_git.go
- [x] GitPush # Has init in tools_git.go
- [x] GitStatus # In tooldefs_git.go

## Go Language Tools
- | | **Module Tools (`pkg/core/tools_go_mod.go`)**
  - [x] GoGetModuleInfo # Has its own init
  - [ ] GoModTidy # (Definition needed/unconfirmed in fetched file)
  - [ ] GoGetPackageInfo # (Definition needed/unconfirmed in fetched file)
- | | **Execution & Formatting (`pkg/core/tools_go_*.go`)**
  - [ ] GoRun
  - [ ] GoTest
  - [ ] GoFormat
  - [ ] GoDiagnostics
- | | **Code Analysis (`pkg/core/tools_go_find_*.go`, `tools_go_semantic.go`)**
  - [ ] GoFindDeclarations
  - [ ] GoFindUsages
  - [ ] GoIndexCode # (from toolGoIndexCode in tools_go_semantic.go)
  - [ ] GoRelateSymbol # (from toolGoRelateSymbol in tools_go_semantic.go)
  - [ ] GetGoPathInfo # (Example, if exists in tools_go.go)

## File API Tools (`pkg/core/tools_file_api.go`, `tools_file_api_sync.go`)
- [x] ListAPIFiles # Has its own init in tools_file_api.go
- [x] DeleteAPIFile # Has its own init in tools_file_api.go
- [x] UploadFile # Has its own init in tools_file_api.go
- [x] SyncFiles # (from toolSyncFiles, assumed registered by tools_file_api.go init)

## IO Tools (`pkg/core/tools_io.go`)
- [ ] Print
- [ ] Input

## Shell Tools (`pkg/core/tools_shell.go`)
- [ ] ShellExec
- [ ] ExecOutputToFile

## Metadata Tools (`pkg/core/tools_metadata.go`)
- [ ] GetScriptMetadata
- [ ] SetScriptMetadata

## Vector Tools (`pkg/core/tools_vector.go`)
- [ ] VectorCreate
- [ ] VectorDotProduct
- [ ] VectorMagnitude
- [ ] VectorAdd
- [ ] VectorSubtract
- [ ] VectorMultiplyScalar
- [ ] VectorDistance
- [ ] VectorCosineSimilarity

## Group-Registered Tools (Status indicates the group registrar exists with an init)
- |x| **Tree Tools** (`pkg/core/tools_tree_register.go`)
  - [x] TreeFindNodes
  - [x] TreeLoad
  - [x] TreeGetNodeMetadata
  - [x] TreeSetNodeMetadata
  - [x] TreeAddNode
  - [x] TreeRemoveNode
  - [x] TreeModifyNode
  - [x] TreeNavigate
  - [x] TreeRender
  - # (Individual tools within this group inherit status from the main registrar)
- |x| **List Tools** (`pkg/core/tools_list_register.go`)
  - [x] ListCreate
  - [x] ListLength
  - [x] ListAppend
  - [x] ListGet
  - [x] ListSet
  # (And many more...)
- |x| **String Tools** (`pkg/core/tools_string.go`)
  - [x] StringLength
  - [x] StringConcat
  # (And many more...)
- |x| **Math Tools** (`pkg/core/tools_math.go`)
  - [x] Add
  - [x] Subtract
  # (And many more...)

## AI/WM Tools (Various `ai_wm_*.go` files - registration needs review)
- [ ] WM_Query # Example from ai_wm_tools.go -> toolWMQuery
- [ ] WM_Store # Example from ai_wm_tools.go -> toolWMStore
- [ ] WM_Delete # Example from ai_wm_tools.go -> toolWMDelete
- [ ] WM_ListKeys # Example from ai_wm_tools.go -> toolWMListKeys
- [ ] PerformTaskWithRetries # Example from ai_wm_tools_execution.go
- [ ] InstanceAdmin # Example from ai_wm_tools_admin.go
# (Many others in ai_wm_tools_definitions.go, ai_wm_tools_instances.go, etc.)

## Semantic Tools (Sub-package `pkg/core/tools/gosemantic/register.go` - has its own init)
- |x| **GoSemantic Sub-package Tools**
  - [x] FindDeclarations # (Advanced version in sub-package)
  - [x] FindUsages # (Advanced version in sub-package)
  - [x] RenameSymbol # (Advanced version in sub-package)
  # (Other tools specific to this sub-package)