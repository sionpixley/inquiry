package inquiry

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"os"
	"reflect"
	"strings"

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

	var zeroValue T
	err = createTable(zeroValue)
	if err != nil {
		return err
	}

	return insertRows(csvFilePath, hasHeaderRow, zeroValue)
}

func buildCreateTableStatement[T any](zeroValue T) (string, error) {
	t := reflect.TypeOf(zeroValue)

	if t.Kind() != reflect.Struct {
		return "", errors.New("inquiry error: generic type provided is not a struct")
	}

	statement := "CREATE TABLE " + t.Name() + "("
	for i := range t.NumField() {
		field := t.Field(i)
		switch {
		case field.Type.Kind() == reflect.Bool:
			statement += field.Name + " INTEGER NOT NULL CHECK(" + field.Name + " IN (0,1)),"
		case field.Type.Kind() == reflect.Float32:
			fallthrough
		case field.Type.Kind() == reflect.Float64:
			statement += field.Name + " REAL NOT NULL,"
		case field.Type.Kind() == reflect.Int:
			fallthrough
		case field.Type.Kind() == reflect.Int8:
			fallthrough
		case field.Type.Kind() == reflect.Int16:
			fallthrough
		case field.Type.Kind() == reflect.Int32:
			fallthrough
		case field.Type.Kind() == reflect.Int64:
			statement += field.Name + " INTEGER NOT NULL,"
		// case field.Type.Kind() == reflect.Pointer:
		case field.Type.Kind() == reflect.String:
			statement += field.Name + " TEXT NOT NULL,"
		default:
			return "", errors.New("inquiry error: unsupported field type")
		}
	}
	if strings.HasSuffix(statement, "(") {
		return "", errors.New("inquiry error: struct has no fields")
	} else {
		statement = strings.TrimSuffix(statement, ",")
	}

	statement += ");"
	return statement, nil
}

func createTable[T any](zeroValue T) error {
	createStatement, err := buildCreateTableStatement(zeroValue)
	if err != nil {
		return err
	}

	_, err = Database.Exec(createStatement)
	return err
}

func insertRows[T any](csvFilePath string, hasHeaderRow bool, zeroValue T) error {
	if _, err := os.Stat(csvFilePath); os.IsNotExist(err) {
		return errors.New("inquiry error: file path does not exist")
	} else if err != nil {
		return err
	}

	file, err := os.Open(csvFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	data, err := reader.ReadAll()

	tx, err := Database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if hasHeaderRow {
		_, err = tx.Exec(".import -skip 1 " + csvFilePath + " " + reflect.TypeOf(zeroValue).Name())
		if err != nil {
			return err
		}
	} else {
		_, err = tx.Exec(".import " + csvFilePath + " " + reflect.TypeOf(zeroValue).Name())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
