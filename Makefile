VERSION=$(shell git rev-parse HEAD)
BINARY=move-gtasks
RELEASE_TAG ?= "unknown"

# build for your system
.PHONY: build
build:
	go build -o $(BINARY) -ldflags "-X github.com/practice-room/move-gtasks/cmd.version=$(RELEASE_TAG)" main.go
	@ls | grep move-gtasks

.PHONY: tidy
tidy:
	go mod tidy -compat=1.17

.PHONY: build-cross
build-cross:
	make build-linux
	make build-darwin

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)_linux_amd64 -ldflags "-X github.com/practice-room/move-gtasks/cmd.version=$(RELEASE_TAG)" main.go
	@ls | grep $(BINARY)_linux_amd64

.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY)_darwin_amd64 -ldflags "-X github.com/practice-room/move-gtasks/cmd.version=$(RELEASE_TAG)" main.go
	@ls | grep $(BINARY)_darwin_amd64