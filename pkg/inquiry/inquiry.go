package inquiry

import (
	"database/sql"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

var Database *sql.DB

func Close() error {
	return Database.Close()
}

func Connect[T any](csvFilePath string) []error {
	var err error
	Database, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		return []error{err}
	}

	err = createTable[T]()
	if err != nil {
		return []error{err}
	}

	return insertRows[T](csvFilePath, false)
}
