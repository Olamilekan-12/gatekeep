BINARY := bin/gatekeep
PKG    := ./...

.DEFAULT_GOAL := help
.PHONY: help run build test race cover fmt vet tidy check clean

help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-8s\033[0m %s\n", $$1, $$2}'

run: ## Run the server locally
	go run ./cmd/gatekeep

build: ## Compile the binary into bin/
	go build -o $(BINARY) ./cmd/gatekeep

test: ## Run the test suite
	go test $(PKG)

race: ## Run the test suite with the race detector
	go test -race $(PKG)

cover: ## Run tests and open an HTML coverage report
	go test -coverprofile=coverage.out $(PKG)
	go tool cover -html=coverage.out

fmt: ## Format all Go source
	gofmt -w .

vet: ## Report suspicious constructs
	go vet $(PKG)

tidy: ## Add missing and remove unused modules
	go mod tidy

check: fmt vet race ## Everything CI runs, in order

clean: ## Remove build and coverage artefacts
	rm -rf bin coverage.out
