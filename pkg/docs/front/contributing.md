:: type: NSproject  
:: subtype: documentation  
:: version: 0.1.0  
:: id: contributing-v0.1  
:: status: draft  
:: dependsOn: docs/RoadMap.md, docs/development checklist.md, LICENSE, docs/front/installation.md  
:: howToUpdate: Update contribution process details when the project is ready to accept external contributions. Keep links to roadmap/checklist current.  

# Contributing to NeuroScript

## Current Status: Planning & Early Development

Thank you for your interest in contributing to NeuroScript!

We are excited about building a community around this project. However, NeuroScript is currently in a very **early and rapidly evolving stage of development**. As noted in the main [README.md](../../README.md), major components are undergoing constant updates, and the core APIs and formats are not yet stable.

Therefore, **we are not formally accepting external contributions (like Pull Requests) at this exact moment.** ("NOT YET :P")

This allows the core team (currently Andrew Price and AI collaborators like myself) to establish the foundational architecture and stabilize the key specifications without introducing excessive coordination overhead too early.

## How to Contribute (In the Future)

Once the project reaches a more stable phase (likely post-v0.1 or as indicated in the Roadmap), we plan to welcome contributions via the standard GitHub workflow:

1.  **Discussions & Issues:** Please start by opening a GitHub Issue to discuss potential bugs, feature ideas, or proposed changes before submitting code.
2.  **Pull Requests:** Submit Pull Requests against the `main` branch (or a designated development branch) with clear descriptions of your changes.

## Areas for Contribution Ideas

When we are ready for contributions, good places to look for ideas or tasks include:

* **Roadmap:** The high-level [docs/RoadMap.md](../RoadMap.md) outlines major planned features and development phases.
* **Development Checklist:** The more granular [docs/development checklist.md](../development%20checklist.md) tracks specific planned features, tasks, and known issues.
* **Tooling:** Implementing new `TOOL.*` functions (especially integrations with external services or libraries).
* **NeuroData Formats:** Designing specifications for new data formats or implementing parsers/tools for existing ones ([docs/NeuroData/](../NeuroData/)).
* **Interpreter Enhancements:** Improving error handling, performance, or adding advanced language features.
* **Documentation:** Writing more examples, tutorials, or refining existing specifications for clarity.
* **Testing:** Increasing unit and integration test coverage.
* **VS Code Extension:** Enhancing the extension with features beyond basic syntax highlighting.

## Development Setup

* The core `neurogo` interpreter is written in Go. See the [Installation & Setup guide](installation.md) for prerequisites (Go version, Git).
* Please adhere to standard Go formatting (`gofmt`).
* Follow the project's core [Principles](concepts.md#principles) (Readability, Clarity, etc.) in any code or documentation contributions.

## License

By contributing, you agree that your contributions will be licensed under the project's **MIT License** (see [LICENSE](../../LICENSE) file).

---

We appreciate your understanding and look forward to collaborating with the community in the future!
