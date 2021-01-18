GO ?= go

# Stub for plain 'make'
all:

include scripts/migrate.mk

run:
	go run ./cmd/ledger -config config.dev.yaml

e2e:
	go test -v -count=1 ./e2e/...

