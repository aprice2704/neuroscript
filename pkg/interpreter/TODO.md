# Interpreter TODO

- check we are short-circuit evaluating -- it seems not and its a pain in the neck.

- need to make all fn and variable lookups case insensitive -- restricted char set?

- should add guard or filter to event handlers **eval in golang** to prevent excessive activations e.g. project: {"myproject", queue: "workq1","allcome"} etc.