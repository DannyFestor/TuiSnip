LANGUAGES_OUT := internal/domain/value/languages_gen.go
TEST_TAGS := feature,e2e
FUZZTIME ?= 30s
PROPERTY_DEEP_CHECKS := 10000
GENERATED_FILES := sqlcgen/|_gen\.go$$|_enum\.go$$|mocks_test\.go$$
SUBMAKE := $(MAKE) --no-print-directory

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@scripts/make-help.sh $(MAKEFILE_LIST)

##@ Generate

# The steps run from the recipe, not as prerequisites, because make -j would run prerequisites out of order.
.PHONY: generate
generate: ## Run every generator, in order
	@$(SUBMAKE) generate-sql
	@$(SUBMAKE) generate-mocks
	@$(SUBMAKE) generate-enums
	@$(SUBMAKE) generate-languages

.PHONY: generate-sql
generate-sql: ## Generate the sqlc query code
	sqlc generate

.PHONY: generate-mocks
generate-mocks: ## Generate the mockery mocks
	mockery

.PHONY: generate-enums
generate-enums: ## Generate the go-enum types
	go generate ./...

.PHONY: generate-languages
generate-languages: ## Generate the Language list from chroma
	go run ./internal/tools/languagegen -out $(LANGUAGES_OUT)

##@ Code quality

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: arch-lint
arch-lint: ## Check the layer rules with go-arch-lint
	go-arch-lint check

.PHONY: fmt
fmt: ## Format with golangci-lint
	golangci-lint fmt

.PHONY: fix
fix: ## Apply pending go fix rewrites
	go fix -tags $(TEST_TAGS) ./...

##@ Test

.PHONY: test
test: test-unit test-feature test-e2e ## Run the unit, feature, and e2e tiers

.PHONY: test-unit
test-unit: ## Run the unit tests
	go test -race ./...

.PHONY: test-feature
test-feature: ## Run the feature tests
	go test -race -tags feature ./test/feature/...

.PHONY: test-e2e
test-e2e: ## Run the e2e tests
	go test -race -tags e2e ./test/e2e/...

##@ Deep test

.PHONY: test-fuzz
test-fuzz: ## Fuzz every target for FUZZTIME each
	FUZZTIME=$(FUZZTIME) scripts/test-fuzz.sh

# rapid reads RAPID_CHECKS from the environment. The -rapid.checks flag would fail every
# test binary that doesn't import rapid.
.PHONY: test-property-deep
test-property-deep: ## Run the property tests with 10000 checks
	RAPID_CHECKS=$(PROPERTY_DEEP_CHECKS) go test -tags feature -run 'Property' ./...

.PHONY: test-mutation
test-mutation: ## Write the gremlins mutation report
	gremlins unleash --tags feature --exclude-files '$(GENERATED_FILES)' --output mutation-report.json .

##@ Build

.PHONY: build
build: ## Build bin/tuisnip
	go build -o bin/tuisnip ./cmd/tuisnip
