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
 1.  Left Pane (A): Displays content from the active "Left Screen".
 2.  Right Pane (B): Displays content from the active "Right Screen".
 3.  Left Input Area (C): Positioned below Pane A. Typically used for system-level commands (//command) or commands/input for the active Left Screen.
 4.  Right Input Area (D): Positioned below Pane B. Typically used for commands/input for the active Right Screen (e.g., chat messages).
 5.  Status Bar: Bottom row for messages, activity indicators.

### The Screener Interface (for tview Panes)

To manage the different types of content that can be displayed within the Left Pane (A) and Right Pane (B), a Go interface named Screener is defined. This interface provides a contract for any struct that wants to act as a displayable screen within these panes.

The Screener interface is defined in pkg/neurogo/tui_screens.go as follows:

go type Screener interface {  Name() string // Returns a short name or identifier for the screen.  Title() string // Returns the title to be displayed for the screen, typically at the top of the pane.  Contents() string // Returns the main text content to be displayed in the pane. } 

Purpose and Usage:

The primary purpose of the Screener interface is to allow for polymorphic handling of different content views. The main TUI logic, managed in pkg/neurogo/tview_tui.go, can treat various screen types uniformly as long as they satisfy this interface.

Key aspects of its usage include:

* Storage: The tviewAppPointers struct holds two slices of Screener implementations: leftScreens []Screener and rightScreens []Screener. These slices store the available screens for the Left Pane (A) and Right Pane (B), respectively.
* Adding Screens: The addScreen(s Screener, onLeft bool) method allows new screens (any type that implements Screener) to be dynamically added to either the left or right pane's list.
* Displaying Screens:
    * The setScreen(s int, onLeft bool) method is responsible for displaying a particular screen. It accesses the Screener from the appropriate slice using the index s.
    * It then calls Title() on the Screener instance to set the title of the tview.TextView representing the pane (e.g., tvP.localOutputView.SetTitle(t.leftScreens[s].Title())).
    * Similarly, it calls Contents() to fetch the main content and display it in the pane's tview.TextView (e.g., tvP.localOutputView.SetText(t.leftScreens[s].Contents())).
* Cycling Screens: The nextScreen(d int, onLeft bool) method allows cycling through the available screens in the leftScreens or rightScreens slices by updating the current index and calling setScreen.
* Status Bar Information: The updateStatusText() method now uses the Name() method of the currently active screens to display their names in the status bar, providing context to the user.

Example Implementation: StaticScreen

A concrete implementation of the Screener interface is the StaticScreen struct, also found in pkg/neurogo/tui_screens.go. This struct is designed for screens that have predefined, static content.

go package neurogo  import "fmt"  // Screener interface (as defined above)  type StaticScreen struct {  title string  contents string  name string }  // Methods implementing Screener interface for StaticScreen func (ss *StaticScreen) Title() string {  return ss.title }  func (ss *StaticScreen) Name() string {  return ss.name }  func (ss *StaticScreen) Contents() string {  return ss.contents }  // Example of creating StaticScreen instances (from tview_tui.go): // hs := &StaticScreen{title: "Help", contents: helpText, name: "Help"} // blnk := &StaticScreen{title: "Blank", contents: " ", name: "Blank"} 
This StaticScreen type allows you to easily create new views with fixed titles, names, and textual content, which can then be managed and displayed by the TUI framework through the Screener interface. For more dynamic content, other structs can be defined that also implement Screener but generate their Contents() or Title() dynamically.

 ## Focus Cycling
 -   Tab: Cycles focus: Left Input (C) -> Right Input (D) -> Right Pane (B, for scrolling) -> Left Pane (A, for scrolling) -> (loop).
 -   Shift+Tab: Cycles focus in reverse.

 ## Pane Content Cycling (Screen Switching)
 -   Ctrl+B: Cycles through the available Screener implementations for the Left Pane (A).
 -   Ctrl+N: Cycles through the available Screener implementations for the Right Pane (B). This will include dynamically added Chat screens.

 ## Available Screens (Initial Set & Planned)

 ### Left Pane (A) Screens:
 -   Script Output Display (default): Shows output from EMIT statements in NeuroScript. (Corresponds to old "Local Output")
 -   Worker Manager Status Display: Lists AI Worker Definitions, their status, numbers for chat selection, etc. Implemented as a StaticScreen or a more dynamic Screener.
 -   Help Screen: Displays help text. Implemented as a StaticScreen.
 -   Blank Screen: A blank screen, useful for clearing a pane. Implemented as a StaticScreen.
 -   (Future: Local files in sandbox & sync status)

 ### Right Pane (B) Screens:
 -   AI Reply Display (default): Shows general, non-chat AI responses or system messages. (Corresponds to old "AI Output")
 -   Help Screen: Displays help text. Implemented as a StaticScreen.
 -   (Future: Chat Session Screen - Dedicated screen for an active conversation with a selected AI Worker. Multiple instances can exist, cycled with Ctrl+N.)
 -   (Future: AI Items In Flight - tasks, worker status)
 -   (Future: Function calls from AI)
 -   (Future: Files in File API)

 ## Command Input Conventions
 Input submitted in either Left (C) or Right (D) input area:
 -   //system_command [args]: Always interpreted as a system-level command by the main TUI model (e.g., //chat <num>, //sync, //run <script>).
 -   /screen_command [args]: Interpreted as a command specific to the currently active Screener that "owns" the input area (this would require extending Screener or type-asserting to a more capable interface if screens are to handle commands directly).
 -   regular text input: Interpreted as general input. Its handling depends on the focused input area (C or D) and potentially the active Screener if custom input handling is implemented for specific screens.

 ## Other Parts
 -   Help Display: Toggled by ? (currently shows the "Help" StaticScreen in the left pane).
 -   Status Bar: Shows current focus, screen counts, and active screen names.