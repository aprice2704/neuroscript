* TUI design in pkg/neurogo

Principally for ng

## Screen Areas

1. "A" Local Output (tall, 1/2 width of screen, top left)
2.    "B" AI Output (tall, 1/2 width of screen, top right)
3. "C" Local Input (below A)
4.    "D" AI Input (below B)
5. Status bar (bottom row)

## Focus Cycling

Tab cycles input focus C-D-B-A (clockwise)
Shift-Tab cycles C-A-B-D (anti-clockwise)

## Alternate Displays

Ctrl-B cycles through the displays available for A
Ctrl-N cycles through those available for B

A:  Script Output display (default)
    Worker Manager status display (worker defs etc.)
    Local files in sandbox & sync status - not yet


B:  AI Reply display (default)
    AI Items In Flight (tasks, workers status etc) - not yet
    Function calls from AI - not yet
    Files in File API - not yet

# Other Parts

Help display