---
title: Nil Pointer Analysis
description: How GoProve tracks nil state through your entire program and proves nil safety.
section: docs
order: 2
---

## Overview

GoProve tracks nil state for every pointer-typed value in your program using abstract interpretation. It doesn't just pattern-match common mistakes — it builds a mathematical model of which pointers can be nil at every point in your code.

## How it works

The nil analysis uses a three-valued lattice:

- **DefinitelyNil** — the pointer is always nil at this point
- **DefinitelyNotNil** — the pointer is never nil at this point
- **MaybeNil** — the pointer might or might not be nil

These states are computed through fixed-point iteration over your program's control flow graph.

## What gets tracked

### Nil checks and early returns

```go
func process(config *Config) {
    if config == nil {
        return
    }
    helper(config) // GoProve knows config is non-nil here
}
```

After a nil check with an early return, GoProve proves the pointer is non-nil on all subsequent paths.

### Allocation functions

```go
buf := new(bytes.Buffer)   // DefinitelyNotNil
buf2 := &bytes.Buffer{}    // DefinitelyNotNil
m := make(map[string]int)  // DefinitelyNotNil
```

Values from `new()`, address-of (`&`), and `make()` are proven non-nil.

### Field reloads after nil checks

```go
if obj.Field != nil {
    obj.Field.Use() // Safe — GoProve uses address-based tracking
}
```

GoProve tracks nil state per **memory address**, not per SSA register. Two loads from the same field share nil state.

### Whole-program parameter tracking

```go
func helper(config *Config) {
    config.Validate() // No warning — all callers pass non-nil
}
```

The CLI mode uses fixed-point iteration across the entire call graph. If **all callers** of a function pass non-nil after nil guards, the parameter is **proven** non-nil.

### Interface method calls

```go
var s fmt.Stringer
s.String() // Bug — nil interface method call
```

GoProve detects nil interface method calls.

### Map ok pattern

```go
v, ok := m[key]
if ok && v != nil {
    *v // Safe — guarded by ok check and nil check
}
```

### Type assertion ok pattern

```go
v, ok := x.(T)
if ok {
    v.Do() // Safe — type assertion succeeded
}
```

## Classification

| State at dereference | Finding |
|---|---|
| DefinitelyNil | **Bug** — proven nil dereference |
| MaybeNil | **Warning** — could not prove safe |
| DefinitelyNotNil | **Safe** — no output |

## Intentional pragmatic choices

- **Method receivers are assumed non-nil.** This eliminates noise from `(*T)(nil).Method()` patterns which are rare in practice.
- **Slice MaybeNil is suppressed.** Nil slice indexing warnings are deferred to the future bounds checker.

## Patterns correctly handled

- `v, ok := m[key]; if ok && v != nil { *v }` — map ok pattern
- `v, ok := x.(T); if ok { v.Do() }` — type assertion ok pattern
- `time.NewTimer()`, `bytes.NewBuffer()` — stdlib return values (via interprocedural analysis)
- `if x.F != nil { x.F.Use() }` — field reload after nil check (via address model)
- Global variable nil checks propagate across subsequent reads
