package repository

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/x1unix/sbda-ledger/internal/web"
)

var (
	// psql is query builder configured for PostgreSQL
	psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
)

var returnIDSuffix = returningSuffix(colID)

func returningSuffix(colName string) string {
	return "RETURNING " + colName
}

func checkAffectedRows(r sql.Result) error {
	affected, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf("cannot check affected rows: %w", err)
	}

	if affected == 0 {
		return web.NewErrNotFound("item not found")
	}

	return nil
}
