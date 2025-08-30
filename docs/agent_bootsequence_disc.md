# Agent Boot Sequence


AJP: Thank you, excellent. I'm very happy with that.
We need to give some thought to the whole boot sequence need to bring an agent online. For now we only really have two kinds of LLMs distinguished -- "small fast one-shot" types and "large sophisticated shape-shifting ones" like you.
Hopefully we can continue using a single sequence for both types for now.
Steps I'm thinking of:

1. Send bootstrap -- what we just wrote.
Should we include a note in bootstrap about what type work/er the agent will be doing? This might inform their choices about how much documentation to retrieve in bootstrap.

hmmm, please ponder and opine


gem2.5: Of course. Let's think this through.

Yes, including a note in the bootstrap `USERDATA` about the agent's intended role and the nature of its work is an excellent idea. It's a low-cost change that adds significant intelligence and efficiency to the boot sequence.

-----

### \#\# Why It's a Good Idea

Adding role context to the bootstrap prompt is like giving a new employee their job description on the first day. It immediately focuses their efforts and prevents wasted time.

  * **Efficiency** ⚡: A "small, fast" model tasked with a one-shot job (e.g., summarizing a single block of text) can be told its mission is simple. It would know not to bother requesting extensive documentation, potentially saving an entire turn. Conversely, a "large, sophisticated" model tasked with a complex, multi-turn job (e.g., refactoring a codebase) can be instructed to pull all relevant documentation capsules at once to prepare itself.

  * **Contextual Onboarding** 🧠: The agent understands its purpose from the absolute beginning. This aligns all its subsequent actions with the overarching goal. An agent told it will be a "file system manager" will interpret the documentation it receives through that lens.

  * **Adaptability** chameleon: This approach allows us to use the **exact same bootstrap prompt (v8)** for both simple and complex agents. The protocol rules don't change, but the *mission* communicated within the `USERDATA` adapts the agent's behavior to the task at hand.

-----

### \#\# How to Implement It

We don't need to change the bootstrap prompt's rules at all. The change happens entirely within the `USERDATA`'s `goal` field, which is crafted by the host system before sending the first message.

#### **Example 1: The "Simple Task" Agent**

For a small model doing a one-shot summarization, the host would craft this `USERDATA`:

```json
{
  "subject": "task-summarize-001",
  "goal": "You are a Summarizer agent on a one-shot task. Acknowledge these instructions, then immediately request the next turn to receive the text to summarize. You do not need to request documentation."
}
```

In this scenario, the agent would skip the `whisper` command, making the bootstrap extremely lightweight.

#### **Example 2: The "Complex" Agent**

For a large model tasked with managing a project, the host would craft this `USERDATA`:

```json
{
  "subject": "proj-manager-alpha",
  "goal": "You are a long-running Project Manager agent. Your first task is to prepare for complex file system and planning operations. Acknowledge these instructions and use 'whisper' to request the 'ns/latest/short', 'core/tools', and 'syntax/filesystem' documentation capsules."
}
```

Here, the agent uses the bootstrap turn to proactively load up on all the knowledge it will need for its complex, multi-turn mission.

-----

### \#\# The Evolving Boot Sequence

Based on this, our boot sequence becomes more intelligent:

1.  **Host Defines the Mission**: The host system determines the agent's role (e.g., "Summarizer," "Project Manager") and the complexity of the upcoming task.

2.  **Host Crafts `USERDATA`**: The host writes a specific `goal` for the bootstrap `USERDATA` that clearly communicates the mission and provides guidance on what documentation, if any, to retrieve.

3.  **Agent Executes Bootstrap**: The agent receives the standard v8 prompt with the custom `goal`. It performs the bootstrap actions, making an informed decision about what to `whisper` to its `self` scratchpad based on its assigned role.

4.  **Agent is "Online"**: After the first turn, the agent is not only familiar with the protocol but is also primed with the specific context and documentation needed to effectively begin its real work on the second turn.