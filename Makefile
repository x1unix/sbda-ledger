GO ?= go

# Stub for plain 'make'
all:

include scripts/migrate.mk

run:
	go run ./cmd/ledger -config config.dev.yaml
