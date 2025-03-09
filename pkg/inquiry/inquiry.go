package idb

import (
	"database/sql"
	"errors"
	"reflect"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

var Database *sql.DB

func Close() error {
	return Database.Close()
}

func Connect[T any](csvFilePath string, hasHeaderRow bool) error {
	var err error
	Database, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		return err
	}

	return createTable[T]()
}

func buildCreateTableStatement[T any](zeroValueStruct T) (string, error) {
	t := reflect.TypeOf(zeroValueStruct)

	if t.Kind() != reflect.Struct {
		return "", errors.New("inquiry error: generic type provided is not a struct")
	}
}

func createTable[T any]() error {
	var zeroValueStruct T
	createStatement, err := buildCreateTableStatement[T](zeroValueStruct)
}
