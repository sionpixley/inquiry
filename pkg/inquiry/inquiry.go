// Package inquiry converts a CSV file into an in-memory SQLite database, allowing you to run SQL statements on the CSV file.
package inquiry

import (
	"database/sql"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// InquiryOptions is a struct that allows you to configure how you connect to your CSV file.
type InquiryOptions struct {
	Delimiter    rune `json:"delimiter"`
	HasHeaderRow bool `json:"hasHeaderRow"`
}

/*
Connect is a function that takes the CSV file path in as a parameter and returns two things:
A pointer to the in-memory SQLite database
and an error slice with the errors that happened during the creation of the database (cap of 25).
If no errors occur, the returned error slice will be nil.

It assumes that there is no header row and that your file delimiter is a comma.
If you need to customize these, please use the function ConnectWithOptions.
*/
func Connect[T any](csvFilePath string) (*sql.DB, []error) {
	return ConnectWithOptions[T](csvFilePath, InquiryOptions{Delimiter: ',', HasHeaderRow: false})
}

// ConnectWithOptions is a function that takes the CSV file path and InquiryOptions in as parameters and returns two things:
// A pointer to the in-memory SQLite database
// and an error slice with the errors that happened during the creation of the database (cap of 25).
// If no errors occur, the returned error slice will be nil.
func ConnectWithOptions[T any](csvFilePath string, options InquiryOptions) (*sql.DB, []error) {
	csvFile, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, []error{err}
	}

	err = createTable[T](csvFile)
	if err != nil {
		return nil, []error{err}
	}

	return insertRows[T](csvFile, csvFilePath, options)
}
