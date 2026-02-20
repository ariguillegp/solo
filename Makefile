.PHONY: help validate validate-full fmt compile lint test-short test test-race build modcheck covercheck vulncheck deploy install canary

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

GO ?= go
GOLANGCI_LINT ?= golangci-lint
COVERAGE_THRESHOLD ?= 60

PRINT = @printf "\n==> %s\n" "$(1)"

# ---------------------------------------------------------------------------
# Dependency edges (enable make -j parallelism)
# ---------------------------------------------------------------------------
#   fmt → compile ──→ test-short
#     └→ lint     └→ build
# ---------------------------------------------------------------------------

compile: fmt
lint: fmt
test-short: compile
build: compile

# validate-full extras: same root, wider fan-out
modcheck: fmt
test: compile
test-race: compile
covercheck: compile
vulncheck: compile

# ---------------------------------------------------------------------------
# Top-level gates
# ---------------------------------------------------------------------------

help: ## Show available targets
	@awk 'BEGIN {FS=":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

validate: fix lint test-short build ## Fast local gate (make -j validate)

validate-full: modcheck lint test test-race covercheck vulncheck build ## Full CI gate (make -j validate-full)

# ---------------------------------------------------------------------------
# Formatting
# ---------------------------------------------------------------------------

fmt: ## Auto-format code (check-only in CI)
ifdef CI
	$(call PRINT,Format check (gofmt -l))
	@test -z "$$(gofmt -l ./cmd ./internal)" || { gofmt -l ./cmd ./internal; echo "FAIL: files need formatting"; exit 1; }
else
	$(call PRINT,Format (gofmt -w))
	@gofmt -w ./cmd ./internal
endif

# ---------------------------------------------------------------------------
# Static analysis
# ---------------------------------------------------------------------------

compile: ## Compile-only including test files (catch build errors fast)
	$(call PRINT,Compile check)
	@$(GO) test -run=^$$ -bench=^$$ ./...

lint: ## golangci-lint (govet + staticcheck + errcheck + unused + gofmt + …)
	$(call PRINT,Lint)
	@$(GOLANGCI_LINT) run --config=.golangci.yml ./...

fix: ## go fix (code modernizer)
	$(call PRINT,Code modernizer)
	@$(GO) fix ./...
# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------

test-short: ## Fast tests for iteration
	$(call PRINT,Tests (short))
	@$(GO) test -short ./...

test: ## Thorough tests (no cache, shuffled)
	$(call PRINT,Tests (no cache, shuffle))
	@$(GO) test -count=1 -shuffle=on ./...

test-race: ## Tests with race detector
	$(call PRINT,Tests (race detector))
	@$(GO) test -race -count=1 ./...

# ---------------------------------------------------------------------------
# Module & coverage checks
# ---------------------------------------------------------------------------

modcheck: ## Verify modules are tidy and unchanged
	$(call PRINT,Module verify)
	@$(GO) mod verify
	@$(GO) mod tidy
	@git diff --exit-code go.mod go.sum || { echo "FAIL: go.mod/go.sum not tidy"; exit 1; }

covercheck: ## Enforce minimum coverage threshold
	$(call PRINT,Coverage (threshold=$(COVERAGE_THRESHOLD)%))
	@$(GO) test -coverprofile=.cover.out ./... > /dev/null 2>&1
	@$(GO) tool cover -func=.cover.out | awk '/^total:/ {gsub(/%/,"",$$3); if ($$3+0 < $(COVERAGE_THRESHOLD)) { printf "FAIL: coverage %.1f%% < %d%%\n", $$3, $(COVERAGE_THRESHOLD); exit 1 } else { printf "OK: coverage %.1f%%\n", $$3 }}'
	@rm -f .cover.out

vulncheck: ## Check for known vulnerabilities
	$(call PRINT,Vulnerability check (govulncheck))
	@govulncheck ./...

# ---------------------------------------------------------------------------
# Build & install
# ---------------------------------------------------------------------------

build: ## Build CLI binary (verify linking)
	$(call PRINT,Build rv)
	@tmpdir=$$(mktemp -d); trap 'rm -rf $$tmpdir' EXIT; $(GO) build -o $$tmpdir/rv ./cmd/rv

deploy: ## Install binary to ~/.local/bin
	mkdir -p ~/.local/bin
	$(GO) build -o ~/.local/bin/rv ./cmd/rv

install: ## Full setup + deploy
	mkdir -p ~/Projects
	mkdir -p ~/.rivet/worktrees
	$(MAKE) deploy

canary: ## Trigger CI workflow on current branch and open in browser
	gh workflow run ci.yml --ref $$(git branch --show-current)
	@sleep 3
	@printf "\nPipeline is running, check it in your browser:\n"
	@gh run list --workflow=ci.yml --branch=$$(git branch --show-current) --limit=1 --json url --jq '.[0].url'
