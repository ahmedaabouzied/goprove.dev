BINARY     = goprove-site
REMOTE     ?= aabouzied@praha.aabouzied.com
REMOTE_DIR ?= /home/aabouzied/goprove.dev

.PHONY: build build-linux build-mac run deploy

build:
	go build -o $(BINARY) .

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY) .

build-mac:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

deploy: build-linux
	ssh $(REMOTE) "mkdir -p $(REMOTE_DIR)"
	ssh $(REMOTE) "sudo systemctl stop goprove-site || true"
	scp $(BINARY) $(REMOTE):$(REMOTE_DIR)/$(BINARY)
	scp deploy/goprove-site.service $(REMOTE):$(REMOTE_DIR)/goprove-site.service
	ssh $(REMOTE) "sudo systemctl start goprove-site"
	@echo "Deployed to $(REMOTE):$(REMOTE_DIR)"
