# Ledger

> Backend for a social expenses ledger for groups of friends.

In this system, users are part of groups; a user can add an expense to the group, specifying how the expense has to be split. At any time, a user can request its balance, that will be the outstanding credits and debits for that users inside the group.

API documentation is available as [OpenAPI/Swagger spec](api/swagger.json).

## Prerequisites

* Docker
* docker-compose
* **GNU** Make
* Go 1.12+
* [golang-migrate](https://github.com/golang-migrate/migrate) (for manual sql migration)

## Usage

### Development

* Start DB and Redis containers with `docker-compose start`
  * Pre-create containers *before* start using `docker-compose up -d` (one time operation)
* `make gen`
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
