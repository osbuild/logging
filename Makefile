
.PHONY: help
help:
	@echo "make [TARGETS...]"
	@echo
	@echo 'Targets:'
	@awk 'match($$0, /^([a-zA-Z_\/-]+):.*? ## (.*)$$/, m) {printf "  \033[36m%-30s\033[0m %s\n", m[1], m[2]}' $(MAKEFILE_LIST) | sort
	@echo
	@echo 'Internal Targets:'
	@awk 'match($$0, /^([a-zA-Z_\/-]+):.*? ### (.*)$$/, m) {printf "  \033[36m%-30s\033[0m %s\n", m[1], m[2]}' $(MAKEFILE_LIST) | sort

coverage_data_examples coverage_data_unittests:
	mkdir $@

.PHONY: clean
clean: ## clean all build and test artifacts
	rm -rf coverage_data_examples coverage_data_unittests
	rm -rf coverage.{txt,html}

.PHONY: unit-tests
unit-tests: coverage_data_unittests ## Run all tests with coverage
	go test -race -covermode=atomic ./... -args -test.gocoverdir="$(shell pwd)/coverage_data_unittests"

.PHONY: test
test: ## Run tests without coverage
	go test -race ./...

.PHONY: lint
lint:  ## run linter / static checker
	staticcheck ./...

.PHONY: coverage-report
coverage-report: coverage_data_examples unit-tests run-examples-cov  ## Run unit tests and examples. Then generate an HTML coverage report
	go tool covdata textfmt -i=./coverage_data_examples/,./coverage_data_unittests/ -o coverage.txt
	go tool cover -o coverage.html -html coverage.txt

.PHONY: run-examples-cov
run-examples-cov: coverage_data_examples  ### Run examples tests with coverage
	for d in internal/example_*; do \
	  GOCOVERDIR=coverage_data_examples go run -coverpkg=github.com/osbuild/logging/... -covermode=atomic -cover github.com/osbuild/logging/$$d || exit 1; \
	done

.PHONY: run-examples
run-examples:  ## Run examples tests
	for d in internal/example_*; do go run github.com/osbuild/logging/$$d || exit 1; done

.PHONY: release
release: ## Inform GOPROXY about a new release
	@GOPROXY=proxy.golang.org go list -m github.com/osbuild/logging@latest
