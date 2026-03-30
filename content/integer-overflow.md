---
title: Integer Overflow
description: How GoProve detects arithmetic and narrowing conversion overflow.
section: docs
order: 4
---

## Overview

Go silently wraps on integer overflow — there's no panic, no error, just wrong results. A `uint8` addition of `200 + 100` silently becomes `44`. GoProve uses interval analysis to detect these cases.

## Types of overflow

### Arithmetic overflow

```go
func add(a, b int8) int8 {
    return a + b // GoProve: Warning — result may exceed int8 range
}
```

When the result interval of an arithmetic operation exceeds the type's bounds, GoProve flags it.

### Narrowing conversion overflow

```go
func narrow(x int16) int8 {
    return int8(x) // GoProve: Warning — x may not fit in int8
}
```

Converting a wider integer to a narrower one can silently truncate.

### Proven safe narrowing

```go
func safeNarrow(x int16) int8 {
    if x < 100 && x > -100 {
        return int8(x) // GoProve: Safe — x ∈ [-99, 99] fits in int8
    }
    return 0
}
```

When bounds checks constrain the value to the target type's range, GoProve proves the conversion is safe.

### Negation overflow

```go
func negate(x int8) int8 {
    return -x // GoProve: Warning — if x == -128, result overflows
}
```

`-math.MinInt8` overflows because `128` doesn't fit in `int8`.

## What gets checked

| Operation | Overflow condition |
|---|---|
| `a + b` | Result exceeds type bounds |
| `a - b` | Result exceeds type bounds |
| `a * b` | Result exceeds type bounds |
| `int8(x)` | x outside `[-128, 127]` |
| `uint8(x)` | x outside `[0, 255]` |
| `-x` | x is the minimum negative value |

## Type bounds

| Type | Range |
|---|---|
| `int8` | `[-128, 127]` |
| `int16` | `[-32768, 32767]` |
| `int32` | `[-2147483648, 2147483647]` |
| `uint8` | `[0, 255]` |
| `uint16` | `[0, 65535]` |

## How it works with branch refinement

GoProve narrows intervals through conditionals:

```go
func safeMul(x, y int8) int8 {
    if x > -10 && x < 10 && y > -10 && y < 10 {
        return x * y // Safe — product ∈ [-81, 81] fits in int8
    }
    return 0
}
```

The bounds checks narrow x and y to `[-9, 9]`, and the multiplication result is proven to fit in `int8`.
