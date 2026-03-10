# Interpreter TODO

- check we are short-circuit evaluating -- it seems not and its a pain in the neck.

- need to make all fn and variable lookups case insensitive -- restricted char set?

- should add guard or filter to event handlers **evaluated in golang** to prevent excessive activations e.g. {project: "myproject", queue: ["workq1","allcome"]} etc.

- fix AST builder bug for empty map literals `{}`. The visitor likely fails to return an empty map node when `map_entry_list_opt` is empty, requiring the `tool.str.ParseJsonString("{}")` workaround.

- add `self` and `system_error_message` to the predefined variables list for nslsp (Language Server Protocol).

:: id: capsule/interpreter_todo
:: version: 3
:: description: Interpreter TODO list tracking parser and evaluation improvements.
:: serialization: md
:: filename: TODO.md