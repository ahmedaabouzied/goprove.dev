# goprove.dev

The website for [GoProve](https://github.com/ahmedaabouzied/goprove) — a static analysis tool for Go that uses abstract interpretation to mathematically prove properties about your code.

Live at **[goprove.dev](https://goprove.dev)**

## Overview

This is a single-binary Go web server that serves the GoProve documentation site. Content is written in Markdown with YAML frontmatter and rendered at startup. All assets (templates, static files, content) are embedded into the binary via `go:embed`.

## Prerequisites

- Go 1.25+

## Running locally

```sh
make run
```

The site will be available at `http://localhost:8080`. Set the `PORT` environment variable to change the port.

## Project structure

```
.
├── content/          # Markdown documentation pages
├── deploy/           # systemd service file
├── static/           # CSS
├── templates/        # Go HTML templates
├── main.go           # Server entrypoint
└── Makefile
```

## Deployment

The Makefile includes a `deploy` target that cross-compiles for Linux, copies the binary and service file to a remote server, and prints systemd setup instructions.

Configure your server details:

```sh
export REMOTE=user@your-server.example.com
export REMOTE_DIR=/home/user/goprove.dev
make deploy
```

### First-time server setup

```sh
sudo cp /path/to/goprove-site.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable goprove-site
sudo systemctl start goprove-site
```

### Subsequent deploys

```sh
make deploy
# then on the server:
sudo systemctl restart goprove-site
```

## Adding content

Create a new Markdown file in `content/` with YAML frontmatter:

```markdown
---
title: Page Title
description: Short description for SEO and llms.txt
section: Section Name
order: 5
---

Your content here.
```

The page will be served at `/{filename-without-extension}` and automatically included in the navigation, sitemap, and LLM-readable endpoints (`/llms.txt`, `/llms-full.txt`).

## License

MIT License. See [LICENSE](LICENSE).
