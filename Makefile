
.PHONY: help
help:
	@echo "make [TARGETS...]"
	@echo
	@echo 'Targets:'
	@awk 'match($$0, /^([a-zA-Z_\/-]+):.*? ## (.*)$$/, m) {printf "  \033[36m%-30s\033[0m %s\n", m[1], m[2]}' $(MAKEFILE_LIST) | sort
	@echo
	@echo 'Internal Targets:'
	@awk 'match($$0, /^([a-zA-Z_\/-]+):.*? ### (.*)$$/, m) {printf "  \033[36m%-30s\033[0m %s\n", m[1], m[2]}' $(MAKEFILE_LIST) | sort

.PHONY: unit-tests
unit-tests:  ## Run all tests
	go test -race -v -covermode=atomic -coverprofile=coverage.txt ./...

.PHONY: lint
lint:  ## run linter / static checker
	staticcheck ./...

.PHONY: coverage-report
coverage-report: unit-tests  ## Run unit tests and generate an HTML coverage report
	go tool cover -o coverage.html -html coverage.txt

.PHONY: run-examples
run-examples:  ## Run examples tests
	for d in internal/example_*; do go run github.com/osbuild/logging/$$d || exit 1; done
