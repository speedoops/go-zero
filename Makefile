# NOTICE: commands
.PHONY: all build run test lint

all: gen build test lint

build:
	@echo ":: `date -Iseconds` build executable target"
	go build

run:
	@echo ":: `date -Iseconds` run application with live reloading"
	#@test -f graph/generated/generated.go || go run github.com/speedoops/go-gqlrest
	#@command -v air &> /dev/null || go get github.com/cosmtrek/air
	# air
	go run main.go

test:
	@echo ":: `date -Iseconds` run unit testing"
	@# go test ./... -v
	@#go.exe test -timeout 30s -run ^TestTodo ./... -short -v
	@#go.exe test -timeout 30s ./... -short -v
	go test -timeout 30s -coverprofile=C:\Users\sangfor\AppData\Local\Temp\vscode-goUquo5g\go-code-cover github.com/tal-tech/go-zero/core/logx -v

lint:
	@echo ":: `date -Iseconds` run static code analysis"
	golangci-lint run --timeout=5m

clean:
	@echo ":: `date -Iseconds` clean temporary building files"
	go clean

help:
	@echo "Welcome to graphql api-gateway project."
	@echo ""
	@echo "Available commands:"
	@echo "    make init       - Setup local develop environment."
	@echo "    make all        - Do this every time before you make a git commit."
	@echo "    make gen        - Do this to apply your modifications of graphql schema."
	@echo "    make build      - Do this to build release binary."
	@echo "    make run        - Run application to play with it. (with Hot-Reloading support)"
	@echo "    make lint       - Do static code analysis. (i.e. golangci-lint)"
	@echo "    make test       - Run developer tests."
	@echo "    make clean      - Clean temporary build files."
	@echo ""
	@echo "Have no idea where to go? Try run 'make init all'."