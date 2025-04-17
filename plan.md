Okay, drawing on our discussions and the guidelines in AI_README.md, here is the concise plan to get the necessary filesystem utilities ready in pkg/core before we tackle the nssync refactor:

Finalize Stat Tool: We have already created toolStat (in pkg/core/tools_fs_stat.go), registered it (in pkg/core/tools_fs_utils.go), and created tests (pkg/core/tools_fs_stat_test.go) following NeuroScript conventions (standard testing, helpers, errors.Is). We should consider this step complete unless further issues arise with those files.

Defer OpenFile (Streaming Read): We will not implement a streaming OpenFile tool or helper at this time. When the nssync package (or its adapter) needs file content (e.g., for hashing), it will rely on reading the entire file via the existing ReadFile tool's capabilities. We accept the current limitation for very large files.

Write a WalkDir tool that returns the file hierarchy in some form the list and tree fns in ns can use

