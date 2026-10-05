LANGUAGES_OUT := internal/domain/value/languages_gen.go
TEST_TAGS := feature,e2e,platform
FUZZTIME ?= 30s
PKG ?= .
PROPERTY_DEEP_CHECKS := 10000
# Each mutant recompiles under parallel workers, which the coverage-run baseline doesn't account for.
# At 20 or below, tui/overlay times out nearly every mutant; 30 is the edge, 50 leaves headroom.
MUTATION_TIMEOUT_COEFFICIENT := 50
MUTATION_TAGS := feature
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
	@$(SUBMAKE) generate-agent-rules

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

.PHONY: generate-agent-rules
generate-agent-rules: ## Generate OpenCode's edit permissions from scripts/agent/protected-paths
	scripts/agent/gen-opencode-rules.sh

##@ Code quality

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: arch-lint
arch-lint: ## Check the layer rules with go-arch-lint
	go-arch-lint check

.PHONY: lint-shell
lint-shell: ## Run shellcheck on every tracked shell script
	git ls-files -z '*.sh' | xargs -0 shellcheck

.PHONY: vulncheck
vulncheck: ## Check dependencies for known vulnerabilities
	govulncheck ./...

.PHONY: fmt
fmt: ## Format with golangci-lint
	golangci-lint fmt

.PHONY: fix
fix: ## Apply pending go fix rewrites
	go fix -tags $(TEST_TAGS) ./...

.PHONY: fix-check
fix-check: ## Fail if go fix has rewrites pending
	go fix -diff -tags $(TEST_TAGS) ./...

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

.PHONY: test-platform
test-platform: ## Run the unit tests plus those against the real clipboard, which they overwrite and restore
	go test -race -tags platform ./...

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
test-mutation: ## Write the gremlins mutation report (PKG=./internal/... for one package)
	gremlins unleash --tags $(MUTATION_TAGS) --exclude-files '$(GENERATED_FILES)' \
		--timeout-coefficient $(MUTATION_TIMEOUT_COEFFICIENT) --output mutation-report.json $(PKG)

.PHONY: test-mutation-changed
test-mutation-changed: ## Mutation-test the packages changed since origin/main, skipping those a parent package's run covers
	MUTATION_TAGS=$(MUTATION_TAGS) scripts/test-mutation-changed.sh

##@ Build

.PHONY: build
build: ## Build bin/tuisnip
	go build -o bin/tuisnip ./cmd/tuisnip

.PHONY: run
run: ## Start tuisnip on your own config and data (ARGS=--paths to pass flags)
	go run ./cmd/tuisnip $(ARGS)

.PHONY: snapshot
snapshot: ## Build the release archives into dist/ without publishing
	goreleaser release --snapshot --clean

.PHONY: release-check
release-check: ## Check the goreleaser config
	goreleaser check
