GOLANGCI_LINT ?= golangci-lint

.PHONY:check-golangci-lint
check-golangci-lint:
	@if ! command -v $(GOLANGCI_LINT) >/dev/null 2>&1 ; then\
		echo "'$(GOLANGCI_LINT)' binary not found. Install golang-migrate or specify binary name with 'GOLANGCI_LINT' parameter" && \
		exit 1; \
	fi;

.PHONY:lint
lint: check-golangci-lint
	$(GOLANGCI_LINT) run
