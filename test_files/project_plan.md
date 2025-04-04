# Project Kronos: Time Tracking App - Plan

Version: 0.1.0
DependsOn: docs/conventions.md

## 1. Overview

Project Kronos aims to create a simple, cross-platform command-line time tracking application. Users should be able to start/stop timers for specific tasks, tag tasks with projects, and generate basic reports.

## 2. Core Requirements

The following checklist outlines the Minimum Viable Product (MVP) requirements.

```neurodata-checklist
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
3. Core Timer Logic (NeuroScript Sketch)
The core logic for managing the timer state can be sketched out in NeuroScript. This procedure would handle starting/stopping and recording entries.

```neuroscript
# id: kronos-timer-logic
# version: 0.0.1
# lang_version: 1.1.0

DEFINE PROCEDURE ManageTimer(action, task_name, project_tag)
COMMENT:
    PURPOSE: Handles starting or stopping the timer, recording entries.
    INPUTS:
        - action: "start" or "stop"
        - task_name: Description of the task (string, required for start)
        - project_tag: Optional project tag (string)
    OUTPUT: Status message (string).
    ALGORITHM:
        1. Get current time.
        2. If action is "start":
            a. Check if a timer is already running. If yes, return error.
            b. Store start time, task name, project tag as 'current_entry'.
            c. Return "Timer started for 'task_name'".
        3. If action is "stop":
            a. Check if a timer is running. If no, return error.
            b. Calculate duration from stored start time.
            c. Create final entry record (task, project, start, end, duration).
            d. Append record to persistent storage (e.g., CALL TOOL.AppendCsvRow).
            e. Clear 'current_entry'.
            f. Return "Timer stopped. Duration: ...".
        4. Handle invalid action.
    CAVEATS: Needs time functions, persistent storage tools (CSV/JSON), and robust state management.
ENDCOMMENT

EMIT "Executing ManageTimer (Placeholder)..." # Placeholder

IF action == "start" THEN
    # Placeholder logic
    EMIT "Placeholder: Starting timer for task: " + task_name
    SET status = "Timer started (placeholder)"
ELSE
    IF action == "stop" THEN
        # Placeholder logic
        EMIT "Placeholder: Stopping timer."
        SET status = "Timer stopped (placeholder)"
    ELSE
        SET status = "Error: Invalid action '" + action + "'"
    ENDBLOCK
ENDBLOCK

RETURN status

END
```

4. Deployment Checklist
Once the MVP is ready, the following steps are needed for an initial internal release.

```neurodata-checklis
# id: kronos-deployment-v1
# version: 0.1.0
# rendering_hint: markdown-list

- [ ] Finalize local storage format (CSV decided).
- [ ] Build executable binaries for Linux, macOS, Windows.
- [ ] Create basic usage instructions (`README.md`).
- [ ] Package binaries and instructions into a zip archive.
- [ ] Announce availability to internal testers.
