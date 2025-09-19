# Capsule Metadata Requirements

**Version:** 1.0  
**Status:** Stable

---

All `capsule` files, whether in Markdown (`.md`) or NeuroScript (`.ns`) format, must include a block of metadata that follows these strict rules. The capsule loader will reject any file that does not conform.

## 1. Required Fields

Every capsule **must** provide the following three metadata keys:

- `::id`: The unique, stable name for the capsule.
- `::version`: The version number for this iteration of the capsule.
- `::description`: A brief, one-line summary of the capsule's purpose.

Additionally, every capsule file must declare its own format with `::serialization: md` or `::serialization: ns`.

## 2. ID (Name) Formatting Rules

The `::id` field is strictly validated:

- It **must** begin with the prefix `capsule/`.
- The portion after the prefix may **only** contain lowercase letters (`a-z`), numbers (`0-9`), underscores (`_`), and hyphens (`-`).
- Dots, spaces, uppercase letters, and `@` symbols are **not** permitted.

**✓ Valid Examples:**
`::id: capsule/aeiou`  
`::id: capsule/bootstrap-agentic`  
`::id: capsule/tool-spec_v2`

**✗ Invalid Examples:**
`::id: capsule/bootstrapAgentic` (contains uppercase)  
`::id: capsule/aeiou.spec` (contains a dot)

## 3. Version Formatting Rules

The `::version` **must be a whole integer**.

**✓ Valid Examples:**
`::version: 1`  
`::version: 5`

**✗ Invalid Examples:**
`::version: 1.2.3` (semver is not allowed)  
`::version: v1` (prefix is not allowed)

## 4. Metadata Location

- For **Markdown** files (`.md`), the metadata block must be at the very **end** of the file.
- For **NeuroScript** files (`.ns`), the metadata block must be at the very **beginning** of the file.