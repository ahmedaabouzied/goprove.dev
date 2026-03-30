BINARY     = goprove-site
REMOTE     ?= user@your-server.example.com
REMOTE_DIR ?= /home/user/goprove.dev

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
	scp $(BINARY) $(REMOTE):$(REMOTE_DIR)/$(BINARY)
	scp deploy/goprove-site.service $(REMOTE):$(REMOTE_DIR)/goprove-site.service
	@echo "Deployed to $(REMOTE):$(REMOTE_DIR)"
	@echo ""
	@echo "First time setup on server:"
	@echo "  sudo cp $(REMOTE_DIR)/goprove-site.service /etc/systemd/system/"
	@echo "  sudo systemctl daemon-reload"
	@echo "  sudo systemctl enable goprove-site"
	@echo "  sudo systemctl start goprove-site"
	@echo ""
	@echo "Subsequent deploys:"
	@echo "  sudo systemctl restart goprove-site"
