# Ledger

> Backend for a social expenses ledger for groups of friends.

In this system, users are part of groups; a user can add an expense to the group, specifying how the expense has to be split. At any time, a user can request its balance, that will be the outstanding credits and debits for that users inside the group.

API documentation is available as [OpenAPI/Swagger spec](https://editor.swagger.io/?url=https://raw.githack.com/x1unix/sbda-ledger/master/api/swagger.yaml).

## Prerequisites

* Docker
* docker-compose
* **GNU** Make
* Go 1.12+
    * *Optionals*:
    * [golang-migrate](https://github.com/golang-migrate/migrate) (for manual sql migration)
    * [golangci-lint](https://golangci-lint.run/) as linter

## Tests

### End-to-end

Integration tests (e2e) are described in [e2e](e2e) directory.
Tests cover all cases, including data structure and routes validation.

* Start environment with `docker-compose start`
* Start back-end API with `make run`
* Run tests with `make e2e`

### Unit

I didn't have much time to cover code with unit tests, so most effort was done on integration testing.
Integration tests cover all cases that could be covered by unit tests, so I don't think that it's critical.

## Usage

### Development

* Start DB and Redis containers with `docker-compose start`
  * Pre-create containers *before* start using `docker-compose up -d` (one time operation)
* `make run`

#### Migrations

Default location for migrations is `db/migrations`. Use `make new-migration` to create a new migration.

Use `LGR_NO_MIGRATION` environment variable to omit on-start migration.

### Production

Use `make` to build the project.
Output binary will be located at `target` directory.

#### Configuration

The service can be configured using environment variables, or a [config file](config.example.yaml).

Use `-c` flag to provide path to a config file.

#### Environment variables

See [config.go](internal/config/config.go) for more options.

| Name                 | Type   | Defaults                           | Description                                      |
|----------------------|--------|------------------------------------|--------------------------------------------------|
| `LGR_HTTP_ADDR`      | string | `:8800`                            | Interface to listen by HTTP server               |
| `LGR_DB_ADDRESS`     | string | `postgres://localhost:5432/ledger` | Postgres DB address (URL or DSN)                 |
| `LGR_REDIS_ADDRESS`  | string | `localhost:6379`                   | Redis server address                             |
| `LGR_REDIS_USER`     | string | -                                  | Redis username                                   |
| `LGR_REDIS_PASSWORD` | string | -                                  | Redis password                                   |
| `LGR_REDIS_DB`       | int    | -                                  | Redis database number                            |
| `LGR_MIGRATIONS_DIR` | string | `db/migrations`                    | Path to directory containing migration scripts   |
| `LGR_VERSION_TABLE`  | string | `schema_migrations`                | Name of a table, which contains database version |
| `LGR_SCHEMA_VERSION` | int    | -                                  | Force set schema version (dangerous)             |
| `LGR_NO_MIGRATION`   | bool   | `false`                            | Skip database migration                          |
