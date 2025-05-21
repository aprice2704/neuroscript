# TUI Design for NeuroScript (ng)

Principally for ng

## Screen Areas

1.  "A" Local Output (tall, 1/2 width of screen, top left)
2.  "B" AI Output (tall, 1/2 width of screen, top right)
3.  "C" Local Input (below A)
4.  "D" AI Input (below B)
5.  Status bar (bottom row) - *Primarily for displaying active screen names in panes.*

## Focus Cycling

-   **Tab**: Cycles focus: Left Input (C) -> Right Input (D) -> Right Pane (B, for scrolling/interaction) -> Left Pane (A, for scrolling/interaction) -> (loop).
-   **Shift+Tab**: Cycles focus in reverse: Left Input (C) -> Left Pane (A) -> Right Pane (B) -> Right Input (D) -> (loop).

## Pane Content Cycling (Screen Switching)

-   **Ctrl+B**: Cycles through the available `Screener` implementations for the Left Pane (A). (e.g., Next/Previous screen in Pane A)
-   **Ctrl+N**: Cycles through the available `Screener` implementations for the Right Pane (B). (e.g., Next/Previous screen in Pane B, including Chat screens)
    *Note: Specific keys for next/previous like Ctrl+P, Ctrl+F might also be implemented for more granular control.*

## Available Screens (Initial Set & Planned)

### Left Pane (A) Screens:

-   Script Output Display (default): Shows output from EMIT statements in NeuroScript.
-   Worker Manager Status Display: Lists AI Worker Definitions, their status, etc.
-   Help Screen: Displays help text.
-   (Future: Local files in sandbox & sync status)

### Right Pane (B) Screens:

-   AI Reply Display (default): Shows general, non-chat AI responses.
-   Chat Session Screen: Dedicated screen for an active conversation. Multiple instances can exist.
-   DebugLog Screen: Shows internal TUI debug messages.
-   Help Screen: Displays help text.
-   (Future: AI Items In Flight - tasks, worker status)
-   (Future: Function calls from AI)
-   (Future: Files in File API)

## Command Input Conventions

Input submitted in either Left (C) or Right (D) input area:

-   `//system_command [args]`: Always interpreted as a system-level command.
-   `/screen_command [args]`: Interpreted as a command specific to the currently active Screener (if supported by the screener).
-   `regular text input`: Interpreted as general input for the focused input area (C or D) or active Screener.

## Multiple Chat screens

Each chat session will have its own screen which will be added to the right panel (B) list of screens.
Chats are typically initiated from the AIWM screen in Pane A when a worker definition is selected.

## Other Controls

-   **?** (Question Mark): Toggles Help Display in the Left Pane (A) when an input field is not focused, or when Pane A or B is focused.
-   **Ctrl+C**: Copies the text content of the currently focused pane (Input Areas C/D, or the visible text view in Panes A/B) to the system clipboard.
-   **Ctrl+Q**: Quits the application.

---
*Screener Interface section (which you had in your design doc) can remain as is, as it describes the architectural pattern.*