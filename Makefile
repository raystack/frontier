GOVERSION := $(shell go version | cut -d ' ' -f 3 | cut -d '.' -f 2)

.PHONY: build check fmt lint test test-race vet test-cover-html help install proto
.DEFAULT_GOAL := build

install:
	@echo "Clean up imports..."
	@go mod download

build: ## build all
	CGO_ENABLED=0 go build -o shield .

lint: ## Run linters
	golangci-lint run

# TODO: create seperate command for integration tests
test: ## Run tests
	go test -race ./...

benchmark: ## Run benchmarks
	go test -run=XX -bench=Benchmark. -count 3 -benchtime=1s github.com/odpf/shield/integration

coverage: ## print code coverage
	go test -race -coverprofile coverage.txt -covermode=atomic ./... -tags=unit_test && go tool cover -html=coverage.txt

clean :
	rm -rf dist

proto:
	./buf.gen.yaml && cp proto/odpf/shield/v1/* proto/v1 && rm -Rf proto/odpf

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

