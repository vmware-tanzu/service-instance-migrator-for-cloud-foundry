SI_MIGRATOR_NAME ?= service-instance-migrator
SI_CRYPTO_NAME ?= service-instance-crypto
SI_MIGRATOR_OUTPUT = ./bin/$(SI_MIGRATOR_NAME)
SI_CRYPTO_OUTPUT = ./bin/$(SI_CRYPTO_NAME)
SI_MIGRATOR_SOURCES = $(shell find ./pkg ./cmd/si-migrator -type f -name '*.go')
SI_CRYPTO_SOURCES = $(shell find ./pkg ./cmd/si-crypto -type f -name '*.go')
GOBIN ?= $(shell go env GOPATH)/bin
VERSION ?= $(shell ./hack/next-version)
GITSHA = $(shell git rev-parse HEAD)
GITDIRTY = $(shell git diff --quiet HEAD || echo "dirty")
LDFLAGS_VERSION = -X github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cli.cliName=$(SI_MIGRATOR_NAME) \
				  -X github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cli.cliVersion=$(VERSION) \
				  -X github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cli.cliGitSHA=$(GITSHA) \
				  -X github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cli.cliGitDirty=$(GITDIRTY)
.DEFAULT_GOAL := help

.PHONY: all
all: install test lint ## Run 'install' 'test' and 'lint'

.PHONY: clean
clean: ## Clean testcache and delete build output
	go clean -testcache
	@rm -rf bin/
	@rm -rf dist/

$(SI_MIGRATOR_OUTPUT): $(SI_MIGRATOR_SOURCES)
	@echo "Building $(VERSION)"
	go build -o $(SI_MIGRATOR_OUTPUT) -ldflags "$(LDFLAGS_VERSION)" ./cmd/si-migrator

$(SI_CRYPTO_OUTPUT): $(SI_CRYPTO_SOURCES)
	@echo "Building $(VERSION)"
	go build -o $(SI_CRYPTO_OUTPUT) -ldflags "$(LDFLAGS_VERSION)" ./cmd/si-crypto

.PHONY: build
build: $(SI_MIGRATOR_OUTPUT) ## Build the main binary
build: $(SI_CRYPTO_OUTPUT)

.PHONY: test
test: ## Run the unit tests
	go test -short  ./...

.PHONY: ginkgo
ginkgo: ## Run the unit tests with ginkgo
	ginkgo -r ./pkg

.PHONY: test-main
test-main: ## Run main tests
	@rm -rf export/
	go test -timeout 30m -v -tags integration ./main_test.go

.PHONY: test-integration
test-integration: test-main

.PHONY: test-export
test-export: ## Run the 'export' main tests
	@rm -rf export/
	go test -timeout 15m -v -tags integration -run ExportOrgSpaceCommand ./main_test.go

.PHONY: test-import
test-import: ## Run the 'import' main tests
	go test -timeout 15m -v -tags integration -run ImportOrgSpaceCommand ./main_test.go

.PHONY: test-export-org
test-export-org: ## Run export org e2e tests only
	@rm -rf ./test/export-org-tests
	go test -timeout 15m -v -tags integration ./test/e2e/export_org_test.go

.PHONY: test-export-space
test-export-space: ## Run export space e2e tests only
	@rm -rf ./test/export-space-tests
	go test -timeout 15m -v -tags integration ./test/e2e/export_space_test.go

.PHONY: test-import-org
test-import-org: ## Run import org e2e tests only
	go test -timeout 15m -v -tags integration ./test/e2e/import_org_test.go

.PHONY: test-import-space
test-import-space: ## Run import space e2e tests only
	go test -timeout 15m -v -tags integration ./test/e2e/import_space_test.go

.PHONY: test-e2e
test-e2e: test-export-org test-import-org test-export-space test-import-space ## Run all e2e tests under test folder

.PHONY: test-all
test-all: test test-main test-e2e ## Run the unit tests and the e2e tests

test-bench: ## Run all the benchmark tests
	go test -bench=. -benchmem ./...

.PHONY: install
install: build ## Copy build to GOPATH/bin
	@cp $(SI_MIGRATOR_OUTPUT) $(GOBIN)
	@cp $(SI_CRYPTO_OUTPUT) $(GOBIN)
	@echo "[OK] CLI binary installed under $(GOBIN)"

.PHONY: coverage
coverage: ## Run the tests with coverage and race detection
	go test -v --race -coverprofile=c.out -covermode=atomic ./...

.PHONY: report
report: ## Show coverage in an html report
	go tool cover -html=c.out -o coverage.html

.PHONY: generate
generate: ## Generate fakes
	go generate ./...

.PHONY: clean-docs
clean-docs: ## Delete the generated docs
	mkdir -p docs/
	rm -f docs/*.md

.PHONY: docs
docs: clean-docs ## Generate documentation
	go run cmd/generate_docs/main.go

.PHONY: release
release: $(SI_MIGRATOR_SOURCES) ## Cross-compile binary for various operating systems
	@mkdir -p dist
	GOOS=darwin   GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(SI_MIGRATOR_OUTPUT)     ./cmd/si-migrator && tar -czf dist/$(SI_MIGRATOR_NAME)-darwin-amd64.tgz -C bin . && rm -f $(SI_MIGRATOR_OUTPUT)
	GOOS=linux    GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(SI_MIGRATOR_OUTPUT)     ./cmd/si-migrator && tar -czf dist/$(SI_MIGRATOR_NAME)-linux-amd64.tgz  -C bin . && rm -f $(SI_MIGRATOR_OUTPUT)
	GOOS=windows  GOARCH=amd64 go build -ldflags "$(LDFLAGS_VERSION)" -o $(SI_MIGRATOR_OUTPUT).exe ./cmd/si-migrator && zip -rj  dist/$(SI_MIGRATOR_NAME)-windows-amd64.zip   bin   && rm -f $(SI_MIGRATOR_OUTPUT).exe

.PHONY: lint-prepare
lint-prepare:
	@echo "Installing latest golangci-lint"
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.47.1
	@echo "[OK] golangci-lint installed"

.PHONY: lint
lint: lint-prepare ## Run the golangci linter
	./bin/golangci-lint run

.PHONY: tidy
tidy: ## Remove unused dependencies
	go mod tidy

.PHONY: list
list: ## Print the current module's dependencies.
	go list -m all

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print help for each make target
	@grep -E '^[a-zA-Z_2-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
