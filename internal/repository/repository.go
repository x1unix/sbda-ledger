package repository

import "github.com/Masterminds/squirrel"

var (
	// psql is query builder configured for PostgreSQL
	psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
)
