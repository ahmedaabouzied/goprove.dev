---
title: Getting Started
description: Install GoProve and run your first analysis in under a minute.
section: docs
order: 1
---

## Install

```bash
go install github.com/ahmedaabouzied/goprove/cmd/goprove@latest
```

GoProve analyzes your entire program at once. It uses whole-program parameter tracking with fixed-point iteration — if all callers of a function pass non-nil after nil guards, the parameter is **proven** non-nil.

## First run

```bash
# Analyze a single package
goprove ./pkg/server

# Analyze all packages (recommended)
goprove ./...

# Use in CI — exits with code 1 if any findings
goprove ./... || exit 1
```

## Understanding the output

GoProve classifies every finding:

- **Error** — Proven to crash at runtime. Mathematical guarantee.
- **Warning** — Could not prove safe or unsafe. Worth investigating.
- No output means the operation is proven safe for tracked patterns.
