# Checks

# Check for go
.PHONY: check-go
check-go:
	@which go > /dev/null 2>&1 || (echo "Error: go is not installed" && exit 1)

# Check for golangci-lint
.PHONY: check-golangci-lint
check-golangci-lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "Error: golangci-lint is not installed" && exit 1)

# Targets that require the checks
update: check-go
generate: check-go
vet: check-go
lint: check-golangci-lint
test: check-go
testacc: check-go
lint-fix: check-golangci-lint
build: check-go
install: check-go

.PHONY: update
update: ## Update go.mod
	go get -u -v
	go mod tidy -v

generate: ## Generate docs and copywrite headers
	go generate ./...
	cd tools; go generate ./...

.PHONY: vet
vet: ## Run vet
	go vet -race ./...

.PHONY: lint
lint: ## Run test
	golangci-lint run ./...

.PHONY: test
test: ## Run test
	go test -v -cover -timeout=120s -parallel=10 ./...

.PHONY: testacc
testacc: ## Run testacc
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: lint-fix
lint-fix: ## Auto lint fix
	golangci-lint run --fix ./...

.PHONY: ci
ci: vet lint test testacc ## Run CI (vet, lint, test)

.PHONY: build
build: ## Build the binary
	go build -v .

.PHONY: install
install: ## Install the binary
	go install -v ./...

## Help display.
## Pulls comments from beside commands and prints a nicely formatted
## display with the commands and their usage information.
.DEFAULT_GOAL := help

.PHONY: help
help: ## Prints this help
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| sort \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
