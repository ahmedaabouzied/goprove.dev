BINARY     = goprove-site
REMOTE     ?= aabouzied@praha.aabouzied.com
REMOTE_DIR ?= /home/aabouzied/goprove.dev
VERSION    := $(shell git rev-parse --short HEAD)
LDFLAGS    := -ldflags "-X main.Version=$(VERSION)"

.PHONY: build build-linux build-mac run deploy minify clean

minify:
	go run tools/minify.go

build: minify
	go build $(LDFLAGS) -o $(BINARY) .

build-linux: minify
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY) .

build-mac: minify
	go build $(LDFLAGS) -o $(BINARY) .

run: build
	./$(BINARY)

deploy: build-linux
	ssh $(REMOTE) "mkdir -p $(REMOTE_DIR)"
	ssh $(REMOTE) "sudo systemctl stop goprove-site || true"
	scp $(BINARY) $(REMOTE):$(REMOTE_DIR)/$(BINARY)
	scp deploy/goprove-site.service $(REMOTE):$(REMOTE_DIR)/goprove-site.service
	ssh $(REMOTE) "sudo systemctl start goprove-site"
	@echo "Deployed to $(REMOTE):$(REMOTE_DIR)"

clean:
	rm -f $(BINARY) static/style.min.css static/app.min.js
