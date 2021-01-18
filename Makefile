GO ?= go

# Stub for plain 'make'
all:

include scripts/migrate.mk
include scripts/golangci.mk

.PHONY: run
run:
	go run ./cmd/ledger -config config.dev.yaml

.PHONY: e2e
e2e:
	go test -v -count=1 ./e2e/...

