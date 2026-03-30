---
title: How It Works
description: How GoProve analyzes your Go code to prove safety properties.
section: learn
order: 5
---

## Abstract interpretation

GoProve is based on a technique called **abstract interpretation**. Instead of running your program with actual values, GoProve runs it with **abstract values** that represent every possible value a variable could hold.

Here's the intuition. Consider this function:

```go
func divide(x, y int) int {
    return x / y
}
```

A test might call `divide(10, 2)` and verify the result is `5`. But that only proves it works for those specific inputs. What about `divide(10, 0)`?

GoProve doesn't test specific inputs. It tracks that `y` could be **any integer** — represented as the range `[-∞, +∞]`. Since that range includes `0`, GoProve flags the division as a warning: the denominator might be zero.

Now add a guard:

```go
func divide(x, y int) int {
    if y == 0 {
        return 0
    }
    return x / y
}
```

GoProve sees the `if y == 0` branch. On the false branch (where the division happens), it **narrows** `y`'s range to exclude zero: `[-∞, -1] ∪ [1, +∞]`. The division is now proven safe for all possible inputs.

This is the core idea: track abstract values through every branch, narrow them at conditionals, and check safety at every operation.

## How it works for nil pointers

The same approach works for pointers. Instead of tracking integer ranges, GoProve tracks whether a pointer is **definitely nil**, **definitely not nil**, or **maybe nil**.

```go
func process(config *Config) {
    config.Validate() // Warning: config might be nil
}
```

GoProve sees that `config` is a parameter — any caller could pass nil. So `config` starts as **maybe-nil**, and the method call is flagged.

Add a nil check:

```go
func process(config *Config) {
    if config == nil {
        return
    }
    config.Validate() // Safe: config is proven non-nil here
}
```

After the nil check with early return, GoProve narrows `config` to **definitely not nil** on all remaining paths. The method call is proven safe.

## Whole-program reasoning

GoProve goes further than single-function analysis. It tracks what **every caller** actually passes to each parameter:

```go
func main() {
    cfg := loadConfig() // returns non-nil Config
    process(cfg)
}

func process(config *Config) {
    config.Validate() // Safe: all callers pass non-nil
}
```

The CLI mode uses fixed-point iteration across the entire call graph. If every caller of `process` passes a non-nil value, the parameter is **proven** non-nil — no nil check needed.

## Address-based tracking

GoProve tracks values per **memory address**, not per variable name. This handles a common pattern that trips up other tools:

```go
if obj.Field != nil {
    obj.Field.Use() // Safe: same address, same nil state
}
```

Two loads from the same field share nil state because they reference the same address. GoProve knows the second access is safe.

## How the analysis runs

1. **Parse and type-check** your Go source
2. **Build SSA form** — a compiler intermediate representation where every variable is assigned once, making dataflow explicit
3. **Walk the control flow graph** — visit each block, updating abstract values
4. **Refine on branches** — narrow values at every conditional
5. **Iterate to a fixed point** — re-analyze loops until values stabilize
6. **Check operations** — at every dereference, division, or conversion, check the abstract value and classify the result
