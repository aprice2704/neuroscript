:: type: NSproject
:: subtype: tool_spec_summary
:: version: 0.1.0
:: id: tool-spec-summary-math-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_math.go, docs/script_spec.md
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Summary: Math Tools (v0.1)

This document provides an abbreviated overview of the built-in Math tools. Note that number arguments can often accept strings that are convertible to numbers.

---

### Tool: `Math.Add`
* **Purpose:** Calculates the sum of two numbers.
* **Syntax:** `CALL Math.Add(num1: <Number>, num2: <Number>)`
* **Arguments:**
    * `num1` (Number): Required. The first number to add.
    * `num2` (Number): Required. The second number to add.
* **Returns:** (Number) The sum (`num1 + num2`). Result is float64.
* **Example:** `CALL Math.Add(5, 3.5)` -> `LAST` will be `8.5`.

---

### Tool: `Math.Subtract`
* **Purpose:** Calculates the difference between two numbers.
* **Syntax:** `CALL Math.Subtract(num1: <Number>, num2: <Number>)`
* **Arguments:**
    * `num1` (Number): Required. The number to subtract from.
    * `num2` (Number): Required. The number to subtract.
* **Returns:** (Number) The difference (`num1 - num2`). Result is float64.
* **Example:** `CALL Math.Subtract(10, 4)` -> `LAST` will be `6.0`.

---

### Tool: `Math.Multiply`
* **Purpose:** Calculates the product of two numbers.
* **Syntax:** `CALL Math.Multiply(num1: <Number>, num2: <Number>)`
* **Arguments:**
    * `num1` (Number): Required. The first number.
    * `num2` (Number): Required. The second number.
* **Returns:** (Number) The product (`num1 * num2`). Result is float64.
* **Example:** `CALL Math.Multiply(6, 7)` -> `LAST` will be `42.0`.

---

### Tool: `Math.Divide`
* **Purpose:** Calculates the division of two numbers, returning a float. Handles division by zero.
* **Syntax:** `CALL Math.Divide(num1: <Number>, num2: <Number>)`
* **Arguments:**
    * `num1` (Number): Required. The dividend.
    * `num2` (Number): Required. The divisor.
* **Returns:** (Number | Error) The quotient (`num1 / num2`). Result is float64. Returns an error if `num2` is zero.
* **Example:** `CALL Math.Divide(10, 4)` -> `LAST` will be `2.5`. `CALL Math.Divide(5, 0)` -> Returns an error.

---

### Tool: `Math.Modulo`
* **Purpose:** Calculates the modulo (remainder) of the division of two *integers*. Handles division by zero.
* **Syntax:** `CALL Math.Modulo(num1: <Integer>, num2: <Integer>)`
* **Arguments:**
    * `num1` (Integer): Required. The dividend (must be an integer).
    * `num2` (Integer): Required. The divisor (must be an integer).
* **Returns:** (Number | Error) The integer remainder (`num1 % num2`). Returns an error if `num2` is zero.
* **Example:** `CALL Math.Modulo(10, 3)` -> `LAST` will be `1`. `CALL Math.Modulo(10, 0)` -> Returns an error.