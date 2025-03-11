package inquiry

import (
	"database/sql"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type InquiryOptions struct {
	Delimiter    rune `json:"delimiter"`
	HasHeaderRow bool `json:"hasHeaderRow"`
}

func Connect[T any](csvFilePath string) (*sql.DB, []error) {
	return ConnectWithOptions[T](csvFilePath, InquiryOptions{Delimiter: ',', HasHeaderRow: false})
}

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
