---
title: Comparison
description: How GoProve compares to NilAway, staticcheck, and go vet.
section: learn
order: 7
---

## Overview

GoProve isn't the first Go static analysis tool, and it won't be the last. Here's how it compares to the tools you're probably already using.

## Feature comparison

| | **GoProve** | **NilAway (Uber)** | **staticcheck** | **go vet** |
|---|---|---|---|---|
| Technique | Abstract interpretation | Constraint-based 2-SAT | Pattern matching | Pattern matching |
| Nil detection | Lattice-based with proof | Constraint inference | Limited patterns | Very limited |
| Division by zero | Yes (interval domain) | No | No | No |
| Integer overflow | Yes (interval domain) | No | No | No |
| Soundness | **Sound for bugs** | Neither sound nor complete | Not sound | Not sound |
| Interprocedural | Return summaries + whole-program params | go/analysis Facts | No | No |
| Parameter tracking | **Whole-program dataflow** | Constraint propagation | No | No |
| Address model | **Memory-address based** | Assertion trees | No | No |

## GoProve's advantages

### Mathematical guarantees

When GoProve says **Error**, it's proven. When it produces no output, it's mathematically guaranteed safe for tracked patterns. No other open-source Go tool offers this level of certainty.

### Broader detection

GoProve is the only open-source Go tool that detects **nil dereferences**, **division by zero**, and **integer overflow** in a single analysis pass.

### Whole-program analysis

GoProve analyzes your entire program at once, tracking values across function boundaries. If all callers pass non-nil, the parameter is proven non-nil — not inferred, proven.

### Address-based memory model

GoProve tracks nil state per memory address. If you check `obj.Field != nil` and then access `obj.Field`, GoProve knows it's safe. Most tools lose track after a field reload.

## NilAway's advantages

NilAway uses constraint inference which can be faster than abstract interpretation's fixed-point iteration. It's battle-tested at Uber and integrates directly with golangci-lint.

## When to use what

| Scenario | Recommendation |
|---|---|
| You want mathematical proof | **GoProve** |
| You need div-by-zero or overflow detection | **GoProve** (only option) |
| You want linting + nil checks in one | **NilAway + staticcheck** |
| CI gate with zero tolerance | **GoProve CLI** |

## Complementary, not competing

GoProve and existing tools serve different purposes. You can run GoProve alongside staticcheck and go vet — they check different things with different guarantees. GoProve adds the mathematical proof layer that no other tool provides.
