// Package inquiry converts CSV files into a SQLite database, allowing you to run SQL statements on them.
package inquiry

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// CsvOptions is a struct that allows you to configure information about your CSV file.
type CsvOptions struct {
	CommentCharacter rune `json:"commentCharacter"`
	Delimiter        rune `json:"delimiter"`
	HasHeaderRow     bool `json:"hasHeaderRow"`
	TrimLeadingSpace bool `json:"trimLeadingSpace"`
	UseLazyQuotes    bool `json:"useLazyQuotes"`
}

/*
Connect is a function that creates an in-memory SQLite database from a CSV file.
It takes the CSV file path as a parameter and returns two things: A pointer to the in-memory SQLite database and an error.
If no errors occur, the returned error will be nil.

It assumes that the CSV file doesn't have a header row and that the file's delimiter is a comma.
If you need to customize these, please use the function ConnectWithOptions.
*/
func Connect[T any](csvFilePath string) (*sql.DB, error) {
	options := CsvOptions{
		Delimiter:        ',',
		HasHeaderRow:     false,
		TrimLeadingSpace: false,
		UseLazyQuotes:    false,
	}
	return ConnectWithOptions[T](csvFilePath, options)
}

// ConnectWithOptions is a function that creates an in-memory SQLite database from a CSV file.
// It takes two parameters: the CSV file path and a CsvOptions struct.
// It returns two things: A pointer to the in-memory SQLite database and an error.
// If no errors occur, the returned error will be nil.
func ConnectWithOptions[T any](csvFilePath string, options CsvOptions) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	t, err := createTable[T](tx)
	if err != nil {
		return nil, err
	}

	err = insertRows(tx, csvFilePath, t, options)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	return db, err
}

/*
CreateTable is a function that creates a new table from a CSV file and adds it to an existing SQLite database.
It takes two parameters: a pointer to the SQLite database and a CSV file path.
It returns an error. If no errors occur, the returned error will be nil.

It assumes that the CSV file doesn't have a header row and that the file's delimiter is a comma.
If you need to customize these, please use the function CreateTableWithOptions.
*/
func CreateTable[T any](db *sql.DB, csvFilePath string) error {
	options := CsvOptions{
		Delimiter:        ',',
		HasHeaderRow:     false,
		TrimLeadingSpace: false,
		UseLazyQuotes:    false,
	}
	return CreateTableWithOptions[T](db, csvFilePath, options)
}

// CreateTableWithOptions is a function that creates a new table from a CSV file and adds it to an existing SQLite database.
// It takes three parameters: a pointer to the SQLite database, a CSV file path, and a CsvOptions struct.
// It returns an error. If no errors occur, the returned error will be nil.
func CreateTableWithOptions[T any](db *sql.DB, csvFilePath string, options CsvOptions) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	t, err := createTable[T](tx)
	if err != nil {
		return err
	}

	err = insertRows(tx, csvFilePath, t, options)
	if err != nil {
		return err
	}

	return tx.Commit()
}
