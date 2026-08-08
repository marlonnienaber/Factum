package shared

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type QueryExecutor interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Rebind(query string) string
}

func GetQueryExecutor(transaction *sqlx.Tx, db *sqlx.DB) QueryExecutor {
	if transaction != nil {
		return transaction
	}
	return db
}
