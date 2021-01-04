# Ledger

> Backend for a social expenses ledger for groups of friends.

In this system, users are part of groups; a user can add an expense to the group, specifying how the expense has to be split. At any time, a user can request its balance, that will be the outstanding credits and debits for that users inside the group.

API documentation is available as [OpenAPI/Swagger spec](api/swagger.json).

## Prerequisites

* Docker
* docker-compose
* **GNU** Make
* Go 1.12+
* [golang-migrate](https://github.com/golang-migrate/migrate)

## Usage

### Development

* Start DB and Redis containers with `docker-compose start`
  * Pre-create containers *before* start using `docker-compose up -d` (one time operation)
* `make gen`
* `make run`

### Production

Use `make` to build the project.
Output binary will be located at `target` directory.

#### Configuration

Configuration file is required in order to start the service.

Default configuration file is available as `config.template.yml`.

Use `sbda-ledger -c [config-file]` to start service.
