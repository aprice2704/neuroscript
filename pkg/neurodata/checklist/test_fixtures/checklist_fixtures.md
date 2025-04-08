# Checklist Test Fixtures

## Simple Valid Items

- [ ] Task 1
- [x] Task 2
- [X] Task 3 (Upper X)

## Whitespace Variations

  - [ ] Leading space text 
- [x]  Trailing space text  
-    [ ]   Spaces before marker

## Mixed Valid and Invalid Lines

# Comment line

- [ ] Valid Item 1
Just some random text
- [x] Valid Item 2
-- Another comment

## Empty Input

(This section intentionally left blank for testing empty strings)


## No Valid Items

This has no checklist items.
- Invalid item format
[ ] Another invalid line

## CRLF Line Endings

(Content below should use CRLF if possible in your editor, otherwise simulates it)
- [ ] Task A
- [x] Task B

## Example from project_plan.md (MVP block)

# id: kronos-mvp-reqs
# version: 0.1.1
# rendering_hint: markdown-list
# status: draft

- [ ] Allow user to start a timer for a task (e.g., `kronos start "Coding feature X"`).
- [ ] Allow user to stop the current timer (e.g., `kronos stop`).
- [ ] Store time entries locally (format TBD - potentially simple CSV or JSON).
- [ ] Allow tagging entries with a project name (e.g., `kronos start "Review PR" --project "NeuroScript"`).
- [x] Basic command-line argument parsing for `start` and `stop`.
- [ ] Generate a simple summary report for today's entries (e.g., `kronos report today`).
- [ ] Ensure basic persistence across app restarts.

## Edge Cases

- [ ] Item with brackets [] in text
- [x] Item with hyphen - in text
- [ ] Item ending in space 
- [ ]