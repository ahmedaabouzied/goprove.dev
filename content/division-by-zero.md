---
title: Division by Zero
description: How GoProve proves the absence of division-by-zero panics using interval analysis.
section: docs
order: 3
---

## Overview

GoProve tracks integer ranges using **interval analysis** — every integer variable gets an abstract value `[lo, hi]` representing all possible values it can hold. At every division operation, GoProve checks whether the denominator's interval includes zero.

## Examples

### Proven safe

```go
func safe(x, y int) int {
    if y != 0 {
        return x / y // GoProve: Safe — y is [min, -1] ∪ [1, max]
    }
    return 0
}
```

The `y != 0` check narrows y's interval to exclude zero on the true branch.

### Proven bug

```go
func unsafe(x int) int {
    zero := 0
    return x / zero // GoProve: Bug — denominator is [0, 0]
}
```

When the denominator is a constant zero, GoProve proves the division will always panic.

### Warning

```go
func maybe(x, y int) int {
    return x / y // GoProve: Warning — y might be zero
}
```

When GoProve can't determine whether y includes zero, it issues a warning.

## How interval analysis works

The interval domain tracks a `[lo, hi]` pair for each integer variable:

- Constants get exact intervals: `x := 5` → `[5, 5]`
- Arithmetic widens intervals: `x + y` where `x ∈ [1, 10]` and `y ∈ [2, 5]` → `[3, 15]`
- Branch conditions narrow intervals: `if x < 10` narrows x to `[lo, 9]` on the true branch
- Loop widening prevents infinite iteration while maintaining soundness

## Branch refinement

GoProve understands all comparison operators:

```go
func guarded(x, y int) int {
    if y > 0 {
        return x / y  // Safe — y ∈ [1, max]
    }
    if y < 0 {
        return x / y  // Safe — y ∈ [min, -1]
    }
    return 0
}
```

Each branch narrows the interval for the compared variable.

## What triggers detection

| Denominator interval | Finding |
|---|---|
| `[0, 0]` | **Bug** — always divides by zero |
| Includes 0 (e.g., `[-5, 5]`) | **Warning** — might divide by zero |
| Excludes 0 (e.g., `[1, 100]`) | **Safe** — no output |

## Modulo operations

The same analysis applies to the `%` (remainder) operator, which also panics on zero denominator in Go.

```go
func mod(x, y int) int {
    if y != 0 {
        return x % y // Safe
    }
    return 0
}
```
