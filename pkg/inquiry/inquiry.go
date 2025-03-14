// Package inquiry converts CSV files into a SQLite database, allowing you to run SQL statements on them.
package inquiry

import (
	"database/sql"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// CsvOptions is a struct that allows you to configure information about your CSV file.
type CsvOptions struct {
	Delimiter    rune `json:"delimiter"`
	HasHeaderRow bool `json:"hasHeaderRow"`
}

/*
Connect is a function that creates an in-memory SQLite database from a CSV file.
It takes the CSV file path as a parameter and returns two things: A pointer to the in-memory SQLite database
and an error slice with the errors that happened during the creation of the database (cap of 25).
If no errors occur, the returned error slice will be nil.

It assumes that the CSV file doesn't have a header row and that the file's delimiter is a comma.
If you need to customize these, please use the function ConnectWithOptions.
*/
func Connect[T any](csvFilePath string) (*sql.DB, []error) {
	return ConnectWithOptions[T](csvFilePath, CsvOptions{Delimiter: ',', HasHeaderRow: false})
}

// ConnectWithOptions is a function that creates an in-memory SQLite database from a CSV file.
// It takes two parameters: the CSV file path and a CsvOptions struct.
// It returns two things: A pointer to the in-memory SQLite database
// and an error slice with the errors that happened during the creation of the database (cap of 25).
// If no errors occur, the returned error slice will be nil.
func ConnectWithOptions[T any](csvFilePath string, options CsvOptions) (*sql.DB, []error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, []error{err}
	}

	t, err := createTable[T](db)
	if err != nil {
		return nil, []error{err}
	}

	return insertRows(db, csvFilePath, options, t)
}

/*
CreateTable is a function that creates a new table from a CSV file and adds it to an existing SQLite database.
It takes two parameters: a pointer to the SQLite database and a CSV file path.
It returns an error slice with the errors that happened during the creation of the table (cap of 25).
If no errors occur, the returned error slice will be nil.

It assumes that the CSV file doesn't have a header row and that the file's delimiter is a comma.
If you need to customize these, please use the function CreateTableWithOptions.
*/
func CreateTable[T any](db *sql.DB, csvFilePath string) []error {
	return CreateTableWithOptions[T](db, csvFilePath, CsvOptions{Delimiter: ',', HasHeaderRow: false})
}

// CreateTableWithOptions is a function that creates a new table from a CSV file and adds it to an existing SQLite database.
// It takes three parameters: a pointer to the SQLite database, a CSV file path, and a CsvOptions struct.
// It returns an error slice with the errors that happened during the creation of the table (cap of 25).
// If no errors occur, the returned error slice will be nil.
func CreateTableWithOptions[T any](db *sql.DB, csvFilePath string, options CsvOptions) []error {
	t, err := createTable[T](db)
	if err != nil {
		return []error{err}
	}

	_, errs := insertRows(db, csvFilePath, options, t)
	return errs
}
