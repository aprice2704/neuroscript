# json_lite: Path-Lite & Shape-Lite

This document defines the **Path-Lite** and **Shape-Lite** mini-specs used by `pkg/json-lite`.  
They are deliberately small, explicit, and easy to test.

- **Path-Lite**: how we address values inside nested JSON-like data.
- **Shape-Lite**: how we describe/validate the expected structure of that data.

The APIs used below are from this package unless stated otherwise:
- `ParsePath(string) (Path, error)`
- `Select(value any, path Path) (any, error)`
- `ParseShape(map[string]any) (*Shape, error)`
- `(*Shape).Validate(value any, allowExtra bool) error`

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
got, err := Select(data, p)
````

**Rules**

* Index must be digits (`0`, `1`, …). No signs, spaces, or letters.
* Keys must not be empty.
* No leading/trailing/duplicate dots.
* No stray brackets.
* **Literal** dots/brackets inside keys are **not** supported here
  (use array form for that — see below).

**Typical errors** (use `errors.Is`):

* `lang.ErrInvalidPath` — bad syntax (e.g., `a..b`, `a[`, `a[1a]`)
* `lang.ErrInvalidArgument` — overlong segment, empty key, etc.
* During selection:

  * `lang.ErrMapKeyNotFound`
  * `lang.ErrListIndexOutOfBounds`
  * `lang.ErrCannotAccessType` (e.g., `items.key` when `items` is a list)
  * `lang.ErrCollectionIsNil`

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
v, _ := Select(data, path) // -> 1
```

Notes

* Each element of `Path` is either `{Key: string, IsKey: true}` or `{Index: int, IsKey: false}`.
* Array form bypasses string parsing limits on allowed characters in keys.
* Depth and segment-length limits **still apply** (see below).

### Path limits

* **Max segments**: `maxPathSegments` (e.g. 128). Exceed → `lang.ErrNestingDepthExceeded`.
* **Max segment length** (keys and bracketed index strings): `maxPathSegmentLen`.
  Exceed → `lang.ErrInvalidArgument`.

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

// Nested map
ContactCard := map[string]any{
  "name": "string",
  "contact": map[string]any{
    "email":  "email",
    "phone?": "string",
  },
}

// List of primitives
Tags := map[string]any{
  "tags[]": "string",
}

// List of maps
Cart := map[string]any{
  "items[]": map[string]any{
    "sku": "string",
    "qty": "int",
    "price": "float",
  },
}
```

Parse & validate

```go
s, err := ParseShape(Person)
if err != nil { /* handle */ }

data := map[string]any{"name": "Ada", "email": "ada@example.com"}
if err := s.Validate(data, /*allowExtra=*/false); err != nil {
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

### Unknown/extra keys

* By default (**`allowExtra=false`**), **unexpected** keys at any level cause `lang.ErrInvalidArgument`.
* Pass **`allowExtra=true`** to tolerate extra keys everywhere.

### Nil handling

* A present key with `nil` value is a **type mismatch** (`lang.ErrValidationTypeMismatch`)
  unless the declared type is `any`.

### Depth limits

* **Parsing** enforces a maximum **shape definition** depth: `maxShapeDepth`.
  Exceeding this during `ParseShape` → `lang.ErrNestingDepthExceeded`.
* **Validation** also enforces `maxShapeDepth` on the **data structure**:
  Excessively deep nesting during `Validate` → `lang.ErrNestingDepthExceeded`.

### Normalization

During `ParseShape`, keys are **normalized**:

* The stored field name is the **base name** with suffixes removed (e.g., `"items[]?"` → `"items"`).
* Flags `IsList` and `IsOptional` are recorded on the field spec.
* You should therefore read/inspect fields **without** suffixes: `s.Fields["items"]`.

### Validation errors (examples)

All errors are standard sentinels wrapped with context; use `errors.Is`:

* `lang.ErrValidationRequiredArgMissing`: `missing required key 'email' at path 'contact'`
* `lang.ErrValidationTypeMismatch`: `expected type 'string' but got 'int' (path: items[0].sku)`
* `lang.ErrInvalidArgument`: `unexpected key 'notes' at path 'contact.notes'`
* `lang.ErrNestingDepthExceeded`: depth limit exceeded either in parsing or validation

---

## End-to-end examples

### A) Validate then select (string path)

```go
shape := map[string]any{
  "user": map[string]any{
    "name": "string",
    "email": "email",
  },
  "items[]": map[string]any{
    "id": "int",
  },
}

s, _ := ParseShape(shape)
data := map[string]any{
  "user": map[string]any{"name": "Ada", "email": "ada@example.com"},
  "items": []any{map[string]any{"id": 100}},
}
_ = s.Validate(data, false)

p, _ := ParsePath("user.email")
v, _ := Select(data, p) // "ada@example.com"
```

### B) Validate then select (array form, weird keys)

```go
s, _ := ParseShape(map[string]any{
  "weird key?": "string",
  "a[]": map[string]any{"b.c": "int"},
})

data := map[string]any{
  "weird key": "ok",              // normalized base key is "weird key"
  "a": []any{ map[string]any{"b.c": 42} },
}
_ = s.Validate(data, true) // allow extra or weirds if needed

path := Path{
  {Key: "a", IsKey: true},
  {Index: 0, IsKey: false},
  {Key: "b.c", IsKey: true}, // literal dotted key
}
v, _ := Select(data, path) // 42
```

### C) Using from ns scripts (typical flow)

```ns
# Ask a model for JSON, then validate against a shape and select values.

set Person = {"name":"string","email":"email","company?":"string"}

ask "Return JSON with {name,email,company?}" with {"json": true} into raw

must tool.ai.Validate(raw, Person, /*allow_extra=*/false)

set email = tool.ai.Select(raw, "email")              # string form
set name  = tool.ai.Select(raw, ["weird.key"])        # array form for literal weird key
```

---

## Quick reference

### Path-Lite

* Prefer **string form** (`"a.b[0].c"`) for normal keys.
* Use **array form** (`[{Key:"a"},{Index:0},{Key:"c"}]`) for literal `.` or `[]` in keys, or code-generated paths.
* Errors: `ErrInvalidPath`, `ErrInvalidArgument`, `ErrMapKeyNotFound`, `ErrListIndexOutOfBounds`, `ErrCannotAccessType`, `ErrCollectionIsNil`.
* Limits: `maxPathSegments`, `maxPathSegmentLen`.

### Shape-Lite

* Keys: `field`, `field?`, `field[]` (order of `[]` and `?` does not matter).
* Values: primitive type name **or** nested shape map.
* Primitives: `string`, `int`, `float`, `bool`, `any`; special string types `email`, `url`, `isoDatetime`.
* `allowExtra=false` by default — unexpected keys are errors.
* `nil` is only valid for type `any`.
* Limits: `maxShapeDepth` (parse + validate).
* Errors: `ErrValidationRequiredArgMissing`, `ErrValidationTypeMismatch`, `ErrInvalidArgument`, `ErrNestingDepthExceeded`.

---

## Rationale

* Keep simple things simple (readable paths and tiny type specs).
* Make failure modes **obvious** and **testable** with sentinel errors.
* Avoid heavy schema languages; use plain maps and short type names.
* Separate **addressing** (Path-Lite) from **shape** concerns (Shape-Lite).

