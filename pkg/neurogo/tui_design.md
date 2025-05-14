# TUI Design for NeuroScript (ng)

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


 ## Core Architecture: Screen-Based Panes

 The TUI is divided into two main vertical panes (Left and Right), each capable of displaying different "Screens". Below these panes are dedicated input areas. A status bar resides at the bottom.

 ### Panes & Associated Input Areas:
 1.  **Left Pane (A)**: Displays content from the active "Left Screen".
 2.  **Right Pane (B)**: Displays content from the active "Right Screen".
 3.  **Left Input Area (C)**: Positioned below Pane A. Typically used for system-level commands (`//command`) or commands/input for the active Left Screen.
 4.  **Right Input Area (D)**: Positioned below Pane B. Typically used for commands/input for the active Right Screen (e.g., chat messages).
 5.  **Status Bar**: Bottom row for messages, activity indicators.

 ### The `Screen` Interface
 Each logical view (e.g., Script Output, WM Status, Chat Session) is an implementation of the `Screen` interface. This interface defines methods for:
 -   `Init(*App) tea.Cmd`: Initialization when the screen becomes active.
 -   `Update(tea.Msg, *App) (Screen, tea.Cmd)`: Handling TUI messages and events.
 -   `View(width, height int) string`: Rendering the screen's content.
 -   `Name() string`: Providing a unique name.
 -   `SetSize(width, height int)`: Adjusting to new dimensions.
 -   `GetInputBubble() *textarea.Model`: Providing its dedicated input area (if any).
 -   `HandleSubmit(*App) tea.Cmd`: Processing input submitted from its bubble.
 -   `Focus(*App) tea.Cmd`: Handling gain of focus.
 -   `Blur(*App) tea.Cmd`: Handling loss of focus.

 The main TUI model (`pkg/neurogo/model.go`) manages collections of these screens and routes updates and view calls to the currently active ones for each pane. The global input areas (C and D) will reflect the input bubble provided by the focused screen in the corresponding pane.

 ## Focus Cycling
 -   `Tab`: Cycles focus: Left Input (C) -> Right Input (D) -> Right Pane (B, for scrolling) -> Left Pane (A, for scrolling) -> (loop).
 -   `Shift+Tab`: Cycles focus in reverse.

 ## Pane Content Cycling (Screen Switching)
 -   `Ctrl+B`: Cycles through the available `Screen` implementations for the Left Pane (A).
 -   `Ctrl+N`: Cycles through the available `Screen` implementations for the Right Pane (B). This will include dynamically added Chat screens.

 ## Available Screens (Initial Set & Planned)

 ### Left Pane (A) Screens:
 -   **Script Output Display** (default): Shows output from `EMIT` statements in NeuroScript. (Corresponds to old "Local Output")
 -   **Worker Manager Status Display**: Lists AI Worker Definitions, their status, numbers for chat selection, etc.
 -   *(Future: Local files in sandbox & sync status)*

 ### Right Pane (B) Screens:
 -   **AI Reply Display** (default): Shows general, non-chat AI responses or system messages. (Corresponds to old "AI Output")
 -   **Chat Session Screen**: Dedicated screen for an active conversation with a selected AI Worker. Multiple instances can exist, cycled with `Ctrl+N`.
 -   *(Future: AI Items In Flight - tasks, worker status)*
 -   *(Future: Function calls from AI)*
 -   *(Future: Files in File API)*

 ## Command Input Conventions
 Input submitted in either Left (C) or Right (D) input area:
 -   **`//system_command [args]`**: Always interpreted as a system-level command by the main TUI model (e.g., `//chat <num>`, `//sync`, `//run <script>`).
 -   **`/screen_command [args]`**: Interpreted as a command specific to the currently active `Screen` that "owns" the input area. The `Screen`'s `HandleSubmit` method processes it (e.g., `/end` in a ChatScreen).
 -   **`regular text input`**: Interpreted as general input for the active `Screen` (e.g., a message to be sent in a ChatScreen).

 ## Other Parts
 -   **Help Display**: Toggled by `?`.
 -   **Status Bar**: Shows current activity, spinner, errors, patch status.
