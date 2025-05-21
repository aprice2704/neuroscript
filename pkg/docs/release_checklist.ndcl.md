# NeuroScript 0.5.0 Release Checklist

## Core Functionality & Stability

### Interpreter
- [ ] Verify all language constructs (variables, control flow (`if`/`else`, `loop`), procedures, operators) work as documented (`docs/formal script spec.md`).
- [ ] Test error handling for syntax errors and runtime errors. Are error messages clear?
- [ ] Confirm evaluation logic (`pkg/core/evaluation_*.go`) is correct, especially for complex expressions and type handling.
- [ ] Test `last` keyword functionality (`library/TestLastKeyword.ns.txt`).
- [ ] Check list/map operations (`library/test_listmap.ns.txt`).

### Tools
- [ ] **General:** Test all registered tools (`pkg/core/tools_list_register.go`) for basic functionality and edge cases.
- [ ] **`fs` tool:** Verify read, write, list, delete operations (`pkg/core/tools_fs_*.go`). Pay special attention to path handling and permissions.
- [ ] **`go_ast` tool:** Confirm find and modify operations work correctly (`pkg/core/tools_go_ast_*.go`). Test with various Go code structures.
- [ ] **`shell` tool:** Thoroughly test execution (`pkg/core/tools_shell.go`) and ensure security implications are understood/mitigated (see Restricted Mode).
- [ ] **`git` tool:** Test basic git operations if included (`pkg/core/tools_git.go`).
- [ ] **`llm` tool:** Ensure interaction with the LLM API (`pkg/core/llm.go`, `llm_tools.go`) is functional and handles responses correctly.
- [ ] **`nspatch` tool:** Verify patching works as expected (`pkg/nspatch/nspatch.go`, `cmd/nspatch/nspatch.go`).
- [ ] **NeuroData Tools:** (`pkg/neurodata/*`): Ensure tools for specific data formats (checklist, blocks) are stable.
- [ ] **Other Tools:** Verify `io`, `math`, `string`, `vector`, etc. tools.

### Restricted Mode
- [ ] Thoroughly test restricted mode (`pkg/core/security*.go`, `docs/restricted_mode.md`).
- [ ] Confirm tool restrictions (especially `fs`, `shell`, `git`) are enforced correctly.
- [ ] Validate allowlist mechanisms (`cmd/neurogo/agent_allowlist.txt`).

## Build & Distribution

### Build Process
- [ ] Document clear build instructions (`docs/build.md` seems to exist, ensure it's up-to-date).
- [ ] Ensure the project builds cleanly (`go build ./...`).
- [ ] Test cross-compilation if supporting multiple OS/architectures.

### Dependencies
- [ ] Check `go.mod` and `go.sum` are up-to-date and committed.
- [ ] Minimize external dependencies where possible.

### Packaging
- [ ] Decide on distribution format (e.g., source code release on GitHub, pre-compiled binaries).
- [ ] Create release artifacts (e.g., zip/tarball of source, binaries).

### Tagging
- [ ] Tag the release commit in Git (e.g., `v0.1.0`).

## Documentation

### README (`README.md`)
- [ ] Clear project description and purpose.
- [ ] Quick start guide (installation, simple example).
- [ ] Link to more detailed documentation.
- [ ] Installation instructions.
- [ ] Basic usage examples for `neurogo` CLI.

### Language Specification (`docs/formal script spec.md`)
- [ ] Ensure it accurately reflects the current language implementation.
- [ ] Cover syntax, data types, operators, control flow, procedures, built-in tools.

### Tool Documentation (`docs/ns/tools/*`?)
- [ ] Document *all* available tools, their parameters, return values, and examples. (Consider generating this automatically if possible).
- [ ] Clearly state potential side effects or security considerations (e.g., `shell`, `fs.write`, `fs.delete`).

### NeuroData Formats (`docs/NeuroData/*`, `docs/neurodata_and_composite_file_spec.md`)
- [ ] Clearly document the specifications for checklist (`.ndcl`), patch (`.ndpatch`), blocks (`.ndblk`), and any other custom formats.

### Examples (`library/examples/*`, other `.ns.txt` files)
- [ ] Provide clear, working examples demonstrating key features and common use cases.
- [ ] Ensure examples run correctly with the release version.

### Agent/LLM Features (`docs/llm_agent_facilities.md`)
- [ ] Explain how to use NeuroScript with LLMs, including tool integration.

### Roadmap (`docs/RoadMap.md`)
- [ ] Update with completed items and future plans (optional but helpful).

### Contribution Guide
- [ ] Basic instructions on how others can contribute (bug reports, features) (Optional, but good for future).

## Testing

### Unit Tests (`pkg/.../*_test.go`)
- [ ] Ensure all existing tests pass (`go test ./...`).
- [ ] Review test coverage. Add tests for any critical, untested areas.
- [ ] Test edge cases and error conditions.

### Integration Tests
- [ ] Test scripts that use multiple tools together (`library/*.ns.txt` files likely serve this purpose).
- [ ] Test `neurogo` command-line execution with various flags and scripts.

### Example Tests
- [ ] Manually run or automate the execution of all provided examples to ensure they work.

## Usability & DX (Developer Experience)

### Command Line Interface (`cmd/neurogo/main.go`)
- [ ] Test command-line arguments (flags for script execution, agent mode, restricted mode, config).
- [ ] Ensure help messages (`-h`, `--help`) are informative.
- [ ] Provide useful output/logging during script execution.

### Configuration (`pkg/neurogo/config.go`)
- [ ] Verify configuration loading (e.g., API keys for LLM) works correctly.
- [ ] Document configuration options.

### Error Messages
- [ ] Review error messages across the interpreter and tools for clarity and helpfulness.

## Legal & Administrative

### License
- [ ] Ensure the MIT license file (`LICENSE`) is present in the root directory.
- [ ] Ensure source files have appropriate license headers (if desired).

### Repository
- [ ] Clean up unnecessary files or branches.
- [ ] Ensure the default branch (e.g., `main` or `master`) is stable.

## Publicity (Post-Release)

### GitHub Release
- [ ] Create a release on GitHub with the tag, including release notes summarizing changes.

### Announcement
- [ ] Plan how/where to announce the release (e.g., blog post, social media, relevant forums).