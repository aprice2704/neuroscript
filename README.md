# NeuroScript: A Toolkit for AI Communication

## Foundation

The world utterly depends on collaboration between humans, AI agents, and traditional computer programs. NeuroScript **cuts through** the friction in these hybrid systems, bridging the gap between ambiguous natural language and rigid code.

Instead of relying on hidden 'chain-of-thought' or ad-hoc prompts, NeuroScript aims to:

* Provide a **shared 'language'** – a toolkit of simple, readable, and executable scripts and data formats.
* Capture **procedures and information clearly**, making complex workflows explicit and repeatable.
* Foster **reliable collaboration** between all participants (human, AI, or computer).
* Uplift **lesser AI models** allowing them to do the work of greater ones.

<p align="center"><img src="docs/sparking_AI_med.jpg" alt="humans uplift machines" width="320" height="200"></p>


## Quick Examples

**NeuroScript (`.ns.txt` - Defining a simple action):**

```neuroscript
-- Defines a reusable "skill"
DEFINE PROCEDURE GreetUser(user_name)
COMMENT:
    PURPOSE: Greets the user.
    INPUTS: - user_name (string).
    OUTPUT: None.
ENDCOMMENT
  -- Basic variable assignment and string operation
  SET message = "Hello, " + user_name + "!"
  -- Output the result
  EMIT message
END
```

**NeuroData Checklist (`.ndcl` - Tracking tasks):**

```plaintext
:: type: Checklist
:: version: 0.1.0
:: description: Simple project task list.

# Main Goals
- [x] Define Core Problem
- [ ] Design Solution
  - [ ] Phase 1: Core language
  - [ ] Phase 2: Basic tools
- [ ] Implement & Test
```

## Dive Deeper

* **Why NeuroScript?** Read the [motivation and benefits](docs/front/why-ns.md).
* **Core Architecture:** Understand the main [components (script, data, go)](docs/front/architecture.md).
* **Key Concepts & Features:** Explore the underlying [principles and features](docs/front/concepts.md).
* **Language & Data Specs:**
    * **For Senior Tech Staff / Architects:** Review the [NeuroScript Language Specification](docs/script%20spec.md), the [Formal Grammar Ideas](docs/formal%20script%20spec.md), the [NeuroData Overview](docs/neurodata_and_composite_file_spec.md), and the [Agent Facilities Design](docs/llm_agent_facilities.md).
    * **For All:** Browse the specific [NeuroData format specifications](docs/NeuroData/).
* **Using `neurogo`:** See the [Installation & Setup guide](docs/front/installation.md).
* **Code & Examples:**
    * **For Developers:** Browse the available [Built-in Tools source](pkg/core/tools_register.go), look at the [examples in the library](library/), and dive into the main [Go source code](pkg/).
* **Project Status & Roadmap:** Check the [Development Checklist](docs/development%20checklist.md) and the high-level [Roadmap](docs/RoadMap.md).
* **Frequently Asked Questions:** See the [FAQ](docs/front/faq.md).
* **Contributing:** Read the [contribution guidelines](docs/front/contributing.md).

## License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.


## Authors

Authors:  Andrew Price (www.eggstremestructures.com),  
          Gemini 2.5 Pro (Exp) (gemini.google.com)  

:: version: 0.1.6  
:: dependsOn: docs/script spec.md, docs/development checklist.md  
:: Authors: Andrew Price, Gemini 2.5 Pro (Exp)  