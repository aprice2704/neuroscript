You've correctly identified the source of the confusion. The documents don't disagree; they simply describe different layers of abstraction. The `aeiou-v3-spec.md` is indeed the canonical source of truth for the **protocol**, while the simplified `ask` syntax is the correct, user-facing implementation in the **language**.

Here is a detailed comparison and report.

---

### Protocol and Capsule Analysis Report

After a careful review, the files are in agreement, but they operate at two different levels: the strict **protocol level** and the user-friendly **language level**. The confusion arises from how the NeuroScript `ask` statement bridges these two layers.

**Finding:** There is no drift between the V3 spec and the capsules. The drift was in the test scripts, which were manually performing a task that the `ask` statement handles automatically.

---

### On-the-Wire Protocol vs. Language Syntax

#### 1. The Protocol Layer (Strict and Explicit)

* **Canonical Source**: `aeiou-v3-spec.md`
* **Rule**: The spec is unambiguous. The `USERDATA` section of an AEIOU v3 envelope **MUST** contain a JSON object. It even provides a minimal schema: `{"subject":"string","brief":"string (optional)","fields":{ }}`.
* **Agreement in Capsules**: The `bootstrap_oneshot.md` and `bootstrap_agentic.md` capsules **correctly reflect this rule**. Their examples show the AI how to parse a JSON object from the `USERDATA` section, which is critical for training the model to follow the protocol.

This layer defines the literal, on-the-wire format that is sent to the AI model. It is strict to ensure determinism and security.

#### 2. The Language Layer (User-Friendly Abstraction)

* **Canonical Syntax**: As you noted, the correct way to write a simple query in a NeuroScript file is:
    `ask "live_agent", "What were the astronauts' names?" into result`
* **The "Magic" of `ask`**: The `ask` statement itself is the abstraction layer. When the NeuroScript interpreter executes this line, its internal logic performs the following steps:
    1.  It takes the simple string prompt ("What were the astronauts' names?").
    2.  It automatically constructs the required protocol-compliant JSON object, effectively doing this for you: `{"subject":"ask", "fields":{"prompt": "What were the astronauts' names?"}}`.
    3.  It then assembles the full AEIOU v3 envelope with this JSON in the `USERDATA` section and sends it to the model.

This is why you don't need to write JSON in your scripts for simple queries. The language abstracts the protocol's complexity away.

---

### Oneshot vs. Agentic Operation

The core mechanism is the same for both, but the expected `USERDATA` complexity and agent behavior differ slightly.

* **Oneshot**:
    * **How it Works**: The user provides a simple string prompt. The `ask` statement wraps it in the basic `{"subject":"ask", "fields":{"prompt":...}}` JSON object.
    * **Agreement**: The `bootstrap_oneshot.md` capsule correctly prepares the AI to receive this JSON and to reply with a single `done` or `abort` token. The files are in complete agreement here.

* **Agentic**:
    * **How it Works**: For more complex, multi-turn tasks, you might need to provide a more structured `USERDATA` payload. The `agentic.txt` test script demonstrates this by manually creating a more complex JSON string: `set userdata_json = '{"subject":"' + subject + '","goal":"' + goal_text + '"}'`. The `ask` statement is smart enough to recognize when it's given a string that is already a JSON object and will use it directly as the `USERDATA` content.
    * **Agreement**: The `bootstrap_agentic.md` capsule prepares the AI for this more structured input and instructs it on how to use the `continue` token to proceed with a multi-step plan. This also aligns perfectly with the V3 spec.

In summary, the system is working as designed. The V3 specification is canonical, and the capsules correctly reflect its strict rules to the AI. The `ask` statement provides a convenient abstraction so that for the most common use case—a simple question—you don't have to deal with the underlying protocol's complexity.