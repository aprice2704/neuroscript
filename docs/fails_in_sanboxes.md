# Directive: Core Interpreter Updates for Robustness and Usability
Date: 2025-09-08
To: NeuroScript Interpreter Team
From: NeuroScript API Team
Subject: Plan for Default Tool Loading & Event Handler Error Propagation

---

## 1. Overview

Two critical issues have been identified that require coordinated changes in the interpreter and api packages. The first is a usability bug where newly created interpreters do not load the standard tool library by default. The second is a serious debuggability issue where runtime errors inside on event handlers fail silently.

This document outlines the required changes for the internal interpreter package. The API team will concurrently implement the corresponding public-facing facades.

---

## 2. Part 1: Default Tool Loading in New Interpreters

### 2.1. The Problem

The public api.New() function creates an interpreter that does not have the standard tool library (str, math, fs, etc.) loaded. This violates the principle of least surprise, as users expect a new interpreter to be functional out-of-the-box. The current behavior forces users to manually register every standard tool, which is cumbersome and error-prone.

### 2.2. Required Changes in pkg/interpreter

The fix is to make the internal interpreter.NewInterpreter() load the standard tools by default, and to provide an option to disable this for specialized use cases.

#### 2.2.1. Modify interpreter.NewInterpreter()

The constructor in pkg/interpreter/interpreter.go must be updated.

-   It should add a new internal flag, skipStdTools bool, to the Interpreter struct.
-   By default, after applying all options, it should check if !i.skipStdTools. If true, it must iterate over the globally registered standard tools (available via tool.ListStandardTools()) and register them into the new interpreter's local tool registry (i.tools).

#### 2.2.2. Create interpreter.WithoutStandardTools() Option

In a file like pkg/interpreter/options.go, please add a new InterpreterOption:

go // WithoutStandardTools prevents the default loading of the standard tool library. func WithoutStandardTools() InterpreterOption {     return func(i *Interpreter) {         i.skipStdTools = true     } } 

This option will be re-exported by the api package to give users control.

---

## 3. Part 2: Eliminating Silent Failures in Event Handlers

### 3.1. The Problem

Currently, when a runtime error (e.g., ToolNotFound, ArgumentMismatch) occurs inside an on event handler, the error is logged internally, but the handler's execution simply stops. The error does not propagate out to the host application, and no in-script signal is generated. This "silent fail" behavior makes debugging nearly impossible.

### 3.2. The Solution: Dual-Channel Error Propagation

We will implement two new, parallel mechanisms for reporting these errors. This work will be centered in the interpreter.Interpreter.EmitEvent method in pkg/interpreter/interpreter.go, which is responsible for cloning the interpreter and executing the handlers.

The logic within EmitEvent should be modified to wrap the execution of each handler in a recover() block to catch panics, which is how lang.RuntimeErrors are often surfaced. If an error is caught, it should trigger the following two channels.

#### 3.2.1. Channel A: Host Callback Function (Go-Idiomatic)

A new InterpreterOption is required to allow the host application to register a Go-level callback for these specific errors.

-   Add to Interpreter struct:
    eventHandlerErrorCallback func(eventName, source string, err *lang.RuntimeError)

-   Create new InterpreterOption in pkg/interpreter/options.go:
    go     // WithEventHandlerErrorCallback registers a function to be called when a runtime     // error occurs during the execution of an 'on event' handler.     func WithEventHandlerErrorCallback(f func(eventName, source string, err *lang.RuntimeError)) InterpreterOption {         return func(i *Interpreter) {             i.eventHandlerErrorCallback = f         }     }     

-   Invoke the Callback: Inside EmitEvent, if an error is recovered, the interpreter must check if i.eventHandlerErrorCallback != nil and, if so, call it with the relevant details.

#### 3.2.2. Channel B: System Error Event (NeuroScript-Idiomatic)

In addition to the callback, the interpreter must also emit a new, standardized system event to make the error visible within the NeuroScript environment itself.

-   Event Name: "system:handler:error"
-   Event Source: Should be a constant, e.g., "NeuroScriptInterpreter".
-   Payload Structure: The payload must be a lang.MapValue with the following key-value pairs:
    -   "source_event_name": (string) The name of the event whose handler failed (e.g., "user:login").
    -   "source_event_source": (string) The original source of the event that failed.
    -   "error_code": (number) The integer code of the lang.RuntimeError.
    -   "error_message": (string) The message from the lang.RuntimeError.

-   Invoke the Event: Inside EmitEvent, immediately after invoking the host callback (if any), the interpreter should construct this payload and call itself (i.EmitEvent(...)) to fire this new system event. This must be done carefully to avoid infinite loops if the error handler for "system:handler:error" itself fails.

---

## 4. Summary of Interpreter Team Deliverables

1.  Tool Loading:
    -   [ ] Modify interpreter.NewInterpreter() to load standard tools by default.
    -   [ ] Add skipStdTools flag to Interpreter struct.
    -   [ ] Implement and export interpreter.WithoutStandardTools() option.

2.  Error Handling:
    -   [ ] Add eventHandlerErrorCallback function field to Interpreter struct.
    -   [ ] Implement and export interpreter.WithEventHandlerErrorCallback() option.
    -   [ ] Modify interpreter.EmitEvent to catch runtime errors from handler execution.
    -   [ ] Inside the error-catching logic in EmitEvent, implement the call to the host callback.
    -   [ ] Inside the same logic, implement the composition and emission of the "system:handler:error" event.

Please coordinate with the API team as we will be building the public facades for these options in parallel. Thank you.