---
title: Roadmap
description: What's done, what's next, and where GoProve is headed.
section: project
order: 9
---

## Current status

GoProve is in active development. Phases 1 and 2 are complete. Phase 2.5 (false positive reduction) is in progress.

## Phase overview

| Phase | Focus | Status |
|---|---|---|
| **1** | Integer interval analysis (div-by-zero, overflow) | **Done** |
| **2** | Nil pointer analysis (address model, interprocedural, whole-program params) | **Done** |
| **2.5** | False positive reduction | **In progress** |
| **3** | Slice bounds analysis | Planned |
| **4** | Whole-program integer range tracking | Planned |
| **5** | GC pressure analysis | Planned |
| **6** | Concurrency analysis | Planned |
| **7** | SARIF output, golangci-lint plugin, GitHub Action improvements | Planned |

## Phase 1: Integer interval analysis (Done)

Tracks integer ranges through your program using the interval abstract domain `[lo, hi]`. Detects:

- Division by zero — proven when denominator interval is `[0, 0]`
- Integer overflow — arithmetic and narrowing conversion overflow
- Modulo by zero

Includes worklist algorithm with widening for loop termination, branch refinement for all comparison operators, and constant propagation.

## Phase 2: Nil pointer analysis (Done)

Tracks nil state for every pointer using a lattice. Features:

- Address-based memory model — field reloads share nil state
- Whole-program parameter tracking — fixed-point iteration across all callers
- Interprocedural return summaries via CHA call graph
- Branch refinement for nil checks
- Map ok pattern, type assertion ok pattern
- Interface method call nil detection

## Phase 2.5: False positive reduction (In progress)

Seed analysis across 20 open-source Go modules identified areas for improvement. Current work focuses on reducing false positives through better handling of:

- callsite collection and Extract instruction handling
- TypeAssert and Lookup instruction patterns
- Standard library return value guarantees
- Type switch narrowing

## Phase 3: Slice bounds analysis (Planned)

Track slice/array length as an interval. At every index operation, check if the index interval fits within `[0, len-1]`. Handle `range` loops, `append`, `copy`, and slicing.

## Phase 4: Whole-program integer range tracking (Planned)

Extend the interval domain with the same whole-program tracking used for nil analysis. Track what callers actually pass as integer arguments.

## Phase 5: GC pressure analysis (Planned)

Classify functions by allocation behavior: GC-transparent (no allocations), GC-bounded (bounded allocations), GC-unbounded (unbounded allocations).

## Phase 6: Concurrency analysis (Planned)

Static data race and deadlock detection.

## Phase 7: Production hardening (Planned)

SARIF output format, golangci-lint plugin submission, and GitHub Action improvements.

## Contributing

Contributions welcome.

```bash
git clone https://github.com/ahmedaabouzied/goprove
cd goprove
go test ./...
```
