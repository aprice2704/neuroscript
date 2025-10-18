# String tools todo

1. If you pass a map or some rando thing to a string tool that only wants a string, it should produce a useful error, not 21 at unknown position -OR- render to string then operate on it


2. TestFDM_NeuroScript_Integration Failure: The error Inspect: max_length must be an integer, got <nil> happens because the script calls tool.str.inspect(event_to_send) with only one argument. The inspect tool has optional arguments (max_length, max_depth). While NeuroScript allows omitting optional args (treating them as nil), the Go implementation or the argument coercion layer seems to be mishandling this nil value for the max_length integer argument. The simplest fix is to explicitly pass nil for the optional arguments in the script itself.