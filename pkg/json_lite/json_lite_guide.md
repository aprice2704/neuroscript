# json_lite: Path-Lite & Shape-Lite

This document defines the **Path-Lite** and **Shape-Lite** mini-specs used by `pkg/json-lite`.
They are deliberately small, explicit, and easy to test.

- **Path-Lite**: how we address values inside nested JSON-like data.
- **Shape-Lite**: how we describe/validate the expected structure of that data.

The APIs used below are from this package unless stated otherwise:
- `ParsePath(string) (Path, error)`
- `Select(value any, path Path, options *SelectOptions) (any, error)`
- `ParseShape(map[string]any) (*Shape, error)`
- `(*Shape).Validate(value any, options *ValidateOptions) error`

---

## Path-Lite

A **path** is a sequence of segments. Each segment is either:
- a **map key** (string), or
- a **list index** (non-negative integer, zero-based).

We support two ways to express the same path:

### 1) String form (parsed by `ParsePath`)

- **Dot** between keys, e.g. `a.b.c`
- **Brackets** for list index, e.g. `items[0]`
- Combine freely, e.g. `items[1].id`

```go
p, err := ParsePath("items[1].id")
if err != nil { /* handle */ }
got, err := Select(data, p, nil) // No special options
````

**Rules**

* Index must be digits (`0`, `1`, …). No signs, spaces, or letters.
* Keys must not be empty.
* No leading/trailing/duplicate dots.
* No stray brackets.
* **Literal** dots/brackets inside keys are **not** supported here
  (use array form for that — see below).

**Typical errors** (use `errors.Is`):

* `ErrInvalidPath` — bad syntax (e.g., `a..b`, `a[`, `a[1a]`)
* `ErrInvalidArgument` — overlong segment, empty key, etc.
* During selection:
  * `ErrMapKeyNotFound`
  * `ErrListIndexOutOfBounds`
  * `ErrCannotAccessType` (e.g., `items.key` when `items` is a list)
  * `ErrCollectionIsNil`

### 2) Array form (programmatic)

Build a `Path` directly with key/index segments; no parsing/regex involved.
Use this when keys contain **`.`** or **`[]`** literally, or when generating paths in code.

```go
// Example data containing “weird” keys:
data := map[string]any{
  "a.b":  1,
  "c[0]": 2,
}

// Build a path that treats "a.b" as a single key:
path := Path{
  {Key: "a.b", IsKey: true}, // map key
}
v, _ := Select(data, path, nil) // -> 1
```

Notes

* Each element of `Path` is either `{Key: string, IsKey: true}` or `{Index: int, IsKey: false}`.
* Array form bypasses string parsing limits on allowed characters in keys.
* Depth and segment-length limits **still apply** (see below).

### Path limits

* **Max segments**: `maxPathSegments` (e.g. 128). Exceed → `ErrNestingDepthExceeded`.
* **Max segment length** (keys and bracketed index strings): `maxPathSegmentLen`.
  Exceed → `ErrInvalidArgument`.

---

## Shape-Lite

A **shape** is a minimal, readable type declaration for JSON-like maps.
It is a Go `map[string]any` with:

* **Keys**: field names, optionally suffixed by:

  * `?` → **optional** field
  * `[]` → **list** of the given type/shape
  * Suffix **order is irrelevant**: `items[]?` ≡ `items?[]`
* **Values**: either:

  * a **primitive type name** (string), or
  * a **nested shape** (`map[string]any`)

Example shapes

```go
// Flat
Person := map[string]any{
  "name":   "string",
  "email":  "email",
  "company?": "string",
}

// List of maps
Cart := map[string]any{
  "items[]": map[string]any{
    "sku": "string",
    "qty": "int",
  },
}
```

Parse & validate

```go
s, err := ParseShape(Person)
if err != nil { /* handle */ }

data := map[string]any{"NAME": "Ada", "EMAIL": "ada@example.com"}

// Validate with case-insensitivity
err = s.Validate(data, &ValidateOptions{
  CaseInsensitive: true,
  AllowExtra:      false,
})
if err != nil {
  // Validation fails fast with a precise path.
}
```

### Primitive types

* `string`, `int`, `float`, `bool`, `any`
* **Special string types**: `email`, `url`, `isoDatetime`
  (Accepted iff the underlying value is a **string**; otherwise type-mismatch.)

### Optional and list semantics

* `field?` → the key may be **absent**. If present with value **nil**, it’s a **type mismatch** unless the type is `any`.
* `field[]` with a **primitive** → the value must be a list of that primitive.
* `field[]` with a **nested shape** → the value must be a list of maps validating against that nested shape.

### Validation Options

The `Validate` method accepts a `ValidateOptions` struct:

```go
type ValidateOptions struct {
    AllowExtra      bool
    CaseInsensitive bool
}
```

* **`AllowExtra`**: If `false` (default), **unexpected** keys at any level cause `ErrInvalidArgument`. If `true`, extra keys are ignored.
* **`CaseInsensitive`**: If `true`, map keys in the data are matched against the shape definition without regard to case (e.g., data with key `"NAME"` will match a shape with key `"name"`). The default is `false`.

### Nil handling

* A present key with `nil` value is a **type mismatch** (`ErrValidationTypeMismatch`)
  unless the declared type is `any`.

### Type Coercion

* If a shape expects type `int`, the validator will also accept a `float32` or `float64`
  value, **if and only if** that float has no fractional part (i.e., it is a
  whole number). This handles type-system mismatches, such as data unwrapped
  from JSON or NeuroScript where all numbers may be floats.

### Depth limits

* **Parsing** enforces a maximum **shape definition** depth: `maxShapeDepth`.
  Exceeding this during `ParseShape` → `ErrNestingDepthExceeded`.
* **Validation** also enforces `maxShapeDepth` on the **data structure**:
  Excessively deep nesting during `Validate` → `ErrNestingDepthExceeded`.

---

## Quick reference

### Path-Lite

* **Select Options**: `Select(..., &SelectOptions{CaseInsensitive: true})`.
* Prefer **string form** (`"a.b[0].c"`) for normal keys.
* Use **array form** (`[{Key:"a"},{Index:0},{Key:"c"}]`) for literal `.` or `[]` in keys.
Example
* Errors: `ErrInvalidPath`, `ErrMapKeyNotFound`, etc.
* Limits: `maxPathSegments`, `maxPathSegmentLen`.

### Shape-Lite

* **Validate Options**: `Validate(..., &ValidateOptions{AllowExtra: true, CaseInsensitive: true})`.
* Keys: `field`, `field?`, `field[]`.
* Values: primitive type name **or** nested shape map.
* Primitives: `string`, `int`, `float`, `bool`, `any`; special string types `email`, `url`, `isoDatetime`.
* `nil` is only valid for type `any`.
* A `float` value is accepted for type `int` if it's a whole number.
* Limits: `maxShapeDepth`.
* Errors: `ErrValidationRequiredArgMissing`, `ErrValidationTypeMismatch`, etc.