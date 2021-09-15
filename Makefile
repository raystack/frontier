GOVERSION := $(shell go version | cut -d ' ' -f 3 | cut -d '.' -f 2)

.PHONY: build check fmt lint test test-race vet test-cover-html help install proto
.DEFAULT_GOAL := build

build: ## build all
	CGO_ENABLED=0 go build -o shield .

check: fmt vet lint ## Run tests and linters

test: ## Run tests
	go test -race ./...

coverage: ## print code coverage
	go test -race -coverprofile coverage.txt -covermode=atomic ./... -tags=unit_test && go tool cover -html=coverage.txt

clean :
	rm -rf dist

proto:
	@buf generate --template buf.gen.yaml

fmt: ## Run gofmt linter
ifeq "$(GOVERSION)" "12"
	@for d in `go list` ; do \
		if [ "`gofmt -l -s $$GOPATH/src/$$d | tee /dev/stderr`" ]; then \
			echo "^ improperly formatted go files" && echo && exit 1; \
		fi \
	done
endif

lint: ## Run golint linter
	@for d in `go list` ; do \
		if [ "`golint $$d | tee /dev/stderr`" ]; then \
			echo "^ golint errors!" && echo && exit 1; \
		fi \
	done

vet: ## Run go vet linter
	@if [ "`go vet | tee /dev/stderr`" ]; then \
		echo "^ go vet errors!" && echo && exit 1; \
	fi

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

cert: ## generate tls self signed cert
	@echo "> cleaning old certs"
	mkdir -p ./certs
	cd ./certs; rm -f ./ca-key.pem ./ca-cert.pem ./ca-cert.srl ./service-key.pem ./service-cert.pem
	@echo "> Generate CA's private key and self-signed certificate"
	cd ./certs; openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=IN/ST=KA/L=Bangalore/O=Kush/OU=KushTech/CN=*.example.io/emailAddress=kush@example.com"
	@echo "> Generate server's private key and certificate signing request (CSR)"
	cd ./certs; openssl genrsa -out service-key.pem 4096
	cd ./certs; openssl req -new -key service-key.pem -out service.csr -config ./certificate.conf
	@echo "Use CA's private key to sign web server's CSR and get back the signed certificate"
	cd ./certs; openssl x509 -req -in service.csr -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial \
		-out service-cert.pem -days 365 -sha256 -extfile ./certificate.conf -extensions req_ext
	cd ./certs; rm service.csr
	cd ./certs; openssl x509 -in service-cert.pem -noout -text
	@echo "> certs generated successfully"


install: ## install required dependencies
	@echo "> installing dependencies"
	go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1
	go get github.com/golang/protobuf/proto@v1.5.2
	go get github.com/golang/protobuf/protoc-gen-go@v1.5.2
	go get google.golang.org/grpc@v1.40.0
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.5.0
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.5.0
	go get github.com/bufbuild/buf/cmd/buf@v0.54.1
