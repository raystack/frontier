GOVERSION := $(shell go version | cut -d ' ' -f 3 | cut -d '.' -f 2)

.PHONY: build check fmt lint test test-race vet test-cover-html help install proto ui
.DEFAULT_GOAL := build
PROTON_COMMIT := "99aa2b37aa3ade1ab99a9da101c1db38b559afe5"

ui:
	@echo " > generating ui build"
	@cd ui && $(MAKE) 

install:
	@echo "Clean up imports..."
	@go mod download
	@go get -d github.com/vektra/mockery/v2@v2.13.1

build:
	CGO_ENABLED=0 go build -o shield .

generate: ## run all go generate in the code base (including generating mock files)
	go generate ./...

lint: ## Run linters
	golangci-lint run

lint-fix:
	golangci-lint run --fix

# TODO: create separate command for integration tests
test: ## Run tests
	@go test -race $(shell go list ./... | grep -v /ui | grep -v /vendor/ | grep -v /test/) -coverprofile=coverage.out

test-all: test e2e-smoke-test e2e-regression-test integration-test lint ## Run all tests

e2e-test: ## Run all e2e tests
	## run `docker network prune` if docker fails to find non-overlapping ipv4 address pool
	go test -v -race ./test/e2e/...

e2e-smoke-test: ## Run smoke tests
	go test -v -race ./test/e2e/smoke

e2e-regression-test: ## Run regression tests
	go test -v -race ./test/e2e/regression

integration-test:
	go test -v -race ./test/integration

benchmark: ## Run benchmarks
	go test -run=XX -bench=Benchmark. -count 3 -benchtime=1s github.com/odpf/shield/test/integration

coverage: ## print code coverage
	go test -race -coverprofile coverage.out -covermode=atomic ./... -tags=unit_test && go tool cover -html=coverage.txt

clean :
	rm -rf dist

proto: ## Generate the protobuf files
	@echo " > generating protobuf from odpf/proton"
	@echo " > [info] make sure correct version of dependencies are installed using 'make install'"
	@buf generate https://github.com/odpf/proton/archive/${PROTON_COMMIT}.zip#strip_components=1 --template buf.gen.yaml --path odpf/shield
	@cp -R proto/odpf/shield/* proto/ && rm -Rf proto/odpf
	@echo " > protobuf compilation finished"

update-swagger-md:
	npx swagger-markdown -i proto/apidocs.swagger.json -o docs/docs/reference/api.md

clean-doc:
	@echo "> cleaning up auto-generated docs"
	@rm -rf ./docs/docs/reference/cli.md
	@rm -f ./docs/docs/reference/api.md

doc: clean-doc update-swagger-md ## Generate api and cli documentation
	@echo "> generate cli docs"
	@go run . reference --plain | sed '1 s,.*,# CLI,' > ./docs/docs/reference/cli.md

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

