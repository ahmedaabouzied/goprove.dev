---
title: "Release: v0.3.0"
description: "v0.3.0 brings global variable tracking, analysis caching, and a major reduction in false positives."
section: project
order: 10
---

## v0.3.0 — March 31, 2026

This is the largest release since the initial nil analysis landed in v0.2.0. The focus is on **accuracy** and **performance** — fewer false positives, faster analysis, and deeper tracking of nil state across your program.

## What's new

### Global variable tracking

GoProve now tracks nil state for **package-level global variables** across function boundaries using fixed-point iteration. If a global is set to a non-nil value at package init time, all subsequent reads see that state. If multiple functions write to the same global, the analysis joins their effects.

This eliminates a class of false positives where globals initialized in `init()` or `main()` were incorrectly flagged as possibly nil.

### Address-taken local variables

Variables whose address is taken (`&x`) and later accessed through pointers now have their nil state tracked correctly. This includes nested field lookups like `obj.Inner.Field` — the analysis resolves the full chain of address operations.

### Analysis caching

Two new caching layers significantly speed up analysis on large codebases:

- **Stdlib nil analysis cache** — pre-computed summaries for Go standard library functions. Generate once with `goprove cache stdlib`, reuse across all projects.
- **Per-project summary cache** — function summaries are cached and reused during fixed-point iteration, avoiding redundant re-analysis.

### False positive reduction

Substantial work went into reducing false positives:

- **Type switch narrowing** — branches inside a type switch correctly narrow the nil state of the switched value.
- **Comma-ok type assertions** — `v, ok := x.(T); if ok { v.Use() }` is now proven safe in both simple and nested forms.
- **Map lookup nil checks** — `v, ok := m[key]` patterns are handled correctly.
- **Function pointer and closure tracking** — passing function values and closures to other functions no longer causes spurious warnings.
- **CHA call graph resolution** — the analysis uses Class Hierarchy Analysis to resolve virtual calls more precisely, reducing false positives from imprecise call graph edges.
- **Multi-predecessor join fix** — a bug where branch refinement was overwritten when a block had multiple predecessors has been fixed.

## Commits in this release

| Area | Change |
|---|---|
| Global tracking | Track global nil state across functions via fixed-point iteration |
| Global tracking | Track global address state across functions in fixed-point loop |
| Global tracking | Set package globals as non-nil variables |
| Global tracking | Handle passing a global var to a function |
| Address model | Track nil state through address-taken local variables |
| Address model | Resolve nested field address lookups |
| Caching | Add a stdlib nil analysis cache |
| Caching | Add a cache for nil pointer analysis summaries |
| False positives | Refine type switch and comma-ok type assertions in true branches |
| False positives | Handle comma-ok type assertion of a value to interface |
| False positives | Fix false positives in type assertion statements |
| False positives | Fix false positives on extract stmt of func call |
| False positives | Handle map lookup nil checks |
| False positives | Handle passing a function pointer to a function |
| False positives | Handle make closure SSA statements |
| False positives | Fix multi-predecessor refinement overwriting joined state |
| Cleanup | Remove the targetPkgs maybe-nil optimization |

## Upgrading

```bash
goprove upgrade
```

Or manually:

```bash
go install github.com/ahmedaabouzied/goprove/cmd/goprove@v0.3.0
```

See the [Upgrading guide](/upgrading) for more details.
