MIGRATIONS_DIR ?= db/migrations
GOMIGRATE ?= migrate

# see: https://github.com/golang-migrate/migrate/issues/96
ifeq (, $(shell which realpath))
MIGRATIONS_PREFIX?=$(shell ruby -e 'puts File.expand_path("$(MIGRATIONS_DIR)")')
else
MIGRATIONS_PREFIX?=$(shell realpath $(MIGRATIONS_DIR))
endif

.PHONY:check-gomigrate
check-gomigrate:
	@if ! command -v $(GOMIGRATE) >/dev/null 2>&1 ; then\
		echo "'$(GOMIGRATE)' binary not found. Install golang-migrate or specify binary name with 'GOMIGRATE' parameter" && \
		exit 1; \
	fi;

.PHONY:new-migration
new-migration: check-gomigrate
	@if [ -z $(name) ]; then echo "usage: 'make new-migration name=migration-name'" && exit 1; fi; \
	$(GOMIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name) && \
	echo "Migration '$(name)' has been created"
