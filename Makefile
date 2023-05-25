GOVERSION := $(shell go version | cut -d ' ' -f 3 | cut -d '.' -f 2)

.PHONY: build check fmt lint test test-race vet test-cover-html help install proto
.DEFAULT_GOAL := build
PROTON_COMMIT := "a1795affeb3f5a6bd8431234e7eb69df253a72c9"

install:
	@echo "Clean up imports..."
	@go mod download
	@go get -d github.com/vektra/mockery/v2@v2.13.1

build: ## build all
	CGO_ENABLED=0 go build -o shield .

generate: ## run all go generate in the code base (including generating mock files)
	go generate ./...

lint: ## Run linters
	golangci-lint run

# TODO: create separate command for integration tests
test: ## Run tests
	go test -race $(shell go list ./... | grep -v /vendor/ | grep -v /test/) -coverprofile=coverage.out

e2e-test: ## Run all e2e tests
	go test -v -race ./test/e2e_test/... -coverprofile=coverage.out

e2e-smoke-test: ## Run smoke tests
	go test -v -race ./test/e2e_test/smoke -coverprofile=coverage.out

e2e-regression-test: ## Run regression tests
	go test -v -race ./test/e2e_test/regression  -coverprofile=coverage.out

benchmark: ## Run benchmarks
	go test -run=XX -bench=Benchmark. -count 3 -benchtime=1s github.com/goto/shield/integration

coverage: ## print code coverage
	go test -race -coverprofile coverage.out -covermode=atomic ./... -tags=unit_test && go tool cover -html=coverage.txt

clean :
	rm -rf dist

proto: ## Generate the protobuf files
	@echo " > generating protobuf from goto/proton"
	@echo " > [info] make sure correct version of dependencies are installed using 'make install'"
	@buf generate https://github.com/goto/proton/archive/${PROTON_COMMIT}.zip#strip_components=1 --template buf.gen.yaml --path gotocompany/shield
	@cp -R proto/gotocompany/shield/* proto/ && rm -Rf proto/gotocompany
	@echo " > protobuf compilation finished"

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

