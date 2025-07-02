# Design Proposal: A Consistent on...do Syntax

The core of the inconsistency is that on error uses means and endon, while on event uses different keywords and a different structure. We can unify them under a single, more intuitive syntax.

Proposed New Syntax:

For Error Handling:

Old: on_error means ... endon
New: on error do ... endon
For Event Handling:

Old: on event <expression> [as <var>] ... endevent
New: on event <expression> [as <var>] do ... endon
Benefits of this Change:

Consistency: Both constructs now follow the same on...do...endon pattern, making the language easier to learn and read.
Clarity: do is a more active and clearer word than means for introducing a block of code that executes in response to something.
Bug Resolution: Critically, this removes means from the on error statement, which I believe is the root cause of the parser ambiguity with func...means.
Event Handler Lifecycle and De-registration
You asked how to de-register an event handler. This requires a way to name or identify handlers so they can be targeted. Here is a proposal for that lifecycle.

1. Named Event Handlers

I propose adding an optional named clause to the on event statement. This allows you to assign a simple string literal as a unique ID to a handler.

New on event syntax with named clause:
on event <expression> named <string_name> [as <var>] do ... endon

Example:
on event "file.updated" named "ui-refresher" as event_data do ... endon

2. New clear event Statement

To de-register handlers, I propose a new clear event statement with two forms:

Clear by Name: De-registers a single, specific handler.

clear event named "ui-refresher"
Clear by Event Type: De-registers all anonymous (un-named) handlers for a specific event type.

clear event "file.updated"
Lifecycle Semantics:

on error: This handler does not need a name or a clear statement. As you noted, its lifecycle is simple: defining a new on error block implicitly replaces the previous one for the current scope.
Anonymous on event handlers: An on event block without a named clause creates an anonymous handler. It lives until the end of the script execution or until it's cleared in bulk via clear event "event-type".
Named on event handlers: A handler defined with named "my-handler-name" can be specifically removed at any time with clear event named "my-handler-name", giving you full control over its lifecycle.