BINARY=chromedriver_wsl
BINDIR:=bin
GO_FILES:=$(shell find . -type f -name '*.go' -print)
BIN_FILE:=${BINDIR}/${BINARY}

.PHONY: build
build: clean fmt $(BIN_FILE)
$(BIN_FILE): $(GO_FILES)
	@go build -o ${BIN_FILE} .
fmt: $(GO_FILES)
	@go fmt ./...
run: $(GO_FILES)
	@go run .
clean:
	@rm -rf bin
