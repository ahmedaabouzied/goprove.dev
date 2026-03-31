---
title: Upgrading
description: How to upgrade GoProve to the latest version.
section: docs
order: 6
---

## Built-in upgrade command

Starting from v0.2.0, GoProve includes a built-in upgrade command:

```bash
goprove upgrade
```

This fetches the latest release from GitHub and reinstalls using `go install`. It is the recommended way to stay up to date.

## Upgrade notifications

GoProve automatically checks for new versions when you run an analysis. If a newer release is available, you will see a message like:

```
A new version of goprove is available: v0.3.0 (current: v0.2.4)
Run 'goprove upgrade' to update.
```

Version information is cached locally in your home directory to avoid repeated network calls.

## Manual upgrade

You can always upgrade manually with `go install`:

```bash
go install github.com/ahmedaabouzied/goprove/cmd/goprove@latest
```

Or pin a specific version:

```bash
go install github.com/ahmedaabouzied/goprove/cmd/goprove@v0.3.0
```

## Checking your version

```bash
goprove version
```

This prints the installed version and build metadata.

## Upgrading from v0.2.x to v0.3.0

v0.3.0 is a drop-in upgrade. No configuration changes are needed. The analysis is significantly more accurate with fewer false positives, and large projects will see faster analysis times thanks to the new caching layer.

If you use the stdlib cache feature, regenerate it after upgrading:

```bash
goprove cache stdlib
```
