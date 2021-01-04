MIGRATIONS_DIR ?= db/migrations
TERN ?= tern

# see: https://github.com/golang-migrate/migrate/issues/96
ifeq (, $(shell which realpath))
MIGRATIONS_PREFIX?=$(shell ruby -e 'puts File.expand_path("$(MIGRATIONS_DIR)")')
else
MIGRATIONS_PREFIX?=$(shell realpath $(MIGRATIONS_DIR))
endif

.PHONY:check-tern
check-tern:
	@if ! command -v $(TERN) >/dev/null 2>&1 ; then\
		echo "'$(TERN)' binary not found. Install golang-migrate or specify binary name with 'TERN' parameter" && \
		exit 1; \
	fi;

.PHONY:new-migration
new-migration: check-tern
	@if [ -z $(name) ]; then echo "usage: 'make new-migration name=migration-name'" && exit 1; fi; \
	$(TERN) new -m $(MIGRATIONS_DIR) $(name) && \
	echo "Migration '$(name)' has been created"
