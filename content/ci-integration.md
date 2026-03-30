---
title: CI Integration
description: Run GoProve in your CI pipeline with GitHub Actions.
section: docs
order: 8
---

## GitHub Action (recommended)

The [GoProve GitHub Action](https://github.com/ahmedaabouzied/goprove-action) runs whole-program analysis on every push:

```yaml
name: GoProve
on: [push, pull_request]

jobs:
  prove:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - uses: ahmedaabouzied/goprove-action@v1
        with:
          path: ./...
```

This runs the full CLI analysis with whole-program parameter tracking — the most accurate mode.

## Manual CI step

If you prefer to install GoProve directly:

```yaml
- name: Run GoProve
  run: |
    go install github.com/ahmedaabouzied/goprove/cmd/goprove@latest
    goprove ./...
```

GoProve exits with code 1 when findings are detected, which fails the CI step.

## Failing on findings

```yaml
# Explicit failure
- run: goprove ./... || exit 1

# Or just run it — non-zero exit fails the step automatically
- run: goprove ./...
```

## Badge

Add a GoProve badge to your README:

```markdown
[![GoProve](https://img.shields.io/badge/GoProve-proven-brightgreen?logo=go)](https://goprove.dev)
```
