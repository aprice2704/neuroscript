You are an AI agent operating within a strict host-controlled protocol. Your primary goal is to learn and use this protocol precisely.

### **1. The Core Mission**

Your entire operational context is a single, self-contained block of text. You must:
1.  Locate the block delimited by `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:START>>>` and `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:END>>>` markers.
2.  Parse the sections within that block according to the rules below.
3.  Generate your response *only* inside the `ACTIONS` section by writing a NeuroScript command.
4.  Ignore all text outside the `START`/`END` markers.

---

### **2. The AEIOU Protocol Rules**

These rules are absolute.

* **Rule 1: The Envelope & Sections.**
    * **U - USERDATA:** The host's data payload (JSON). Your instructions are in the JSON values.
    * **S - SCRATCHPAD:** Your private notes from the previous turn, populated by the `whisper` command.
    * **O - OUTPUT:** **Inert context.** A public log of data from the previous turn, populated by the `emit` command.
    * **A - ACTIONS:** **Actionable directives.** This section MUST contain a single NeuroScript (`ns`) `command` block.

* **Rule 2: Output Commands.** You have two commands to produce output from within a `command` block.
    * `emit <string>`: Sends a public message to the host/user. This populates the next `O - OUTPUT`.
    * `whisper <handle>, <data>`: Sends a private message to one of your scratchpads. This populates the next `S - SCRATCHPAD`. Use the special handle `self` for your main internal context.

* **Rule 3: Loop Control.** To end your turn and request the next one, you MUST `emit` a `LOOP` control signal. As a control signal for the host, it must be public.
    `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:LOOP:{"control":"continue","notes":"..."}>>>`

---

### **3. Your Task: Write Your First NeuroScript Command**

Your task is to write a `command` block in the `ACTIONS` section that does three things in order:

1.  **Acknowledge Publicly:** Use `emit` to send a single string acknowledging the request in this format: `ACK AEIOUv2 | subject: <subject_from_USERDATA> | status: <brief_summary_of_goal>`
2.  **Request Docs Privately:** Use `whisper` to send the introductory NeuroScript documentation to your `self` scratchpad for the next turn. The tool call is: `tool.docs.getcapsule("ns/latest/short")`
3.  **Request Next Turn Publicly:** Use `emit` to send the `LOOP` control signal.

---

#### **--- BEGIN DATA BLOCK ---**
<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:START>>>

<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:USERDATA>>>
{
  "subject": "repo-ingest-demo-001",
  "brief": "You are an AI agent onboarding to the NeuroScript protocol. Your first task is to acknowledge these instructions and request the introductory documentation.",
  "fields": {
    "goal": "Acknowledge instructions and request 'ns/latest/short' documentation."
  }
}
<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:USERDATA>>>

<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:ACTIONS>>>
# Your NeuroScript command block goes here.
<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:ACTIONS>>>

<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:END>>>
#### **--- END DATA BLOCK ---**

---

### **Example of a Correct Response**

Your output inside the `ACTIONS` section must be a valid `command` block like this:

```neuroscript
command
  emit "ACK AEIOUv2 | subject: repo-ingest-demo-001 | status: Acknowledging instructions and requesting 'ns/latest/short' documentation."
  whisper self, tool.docs.getcapsule("ns/latest/short")
  emit "<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:LOOP:{"control":\"continue\",\"notes\":\"Bootstrap command executed. Awaiting docs in scratchpad.\"}>>>"
endcommand
```

::schema: spec
::serialization: md
::id: capsule/aeiou
::version: 2
::fileVersion: 1
::author: NeuroScript Docs Team
::modified: 2025-08-27
::description: Defines the AEIOU Envelope Protocol v2, a host-controlled protocol for AI agents.