package inquiry

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
)

const (
	_FILE_PATH_DOES_NOT_EXIST_ERROR string = "inquiry error: file path does not exist"
	_MISMATCH_NUM_OF_COLUMNS_ERROR  string = "inquiry error: number of columns in the row does not match the number of fields in the struct"
	_NO_FIELDS_ERROR                string = "inquiry error: struct has no fields"
	_NOT_A_STRUCT_ERROR             string = "inquiry error: generic type provided is not a struct"
	_UNSUPPORTED_FIELD_TYPE_ERROR   string = "inquiry error: unsupported field type"
)

func buildCreateTableStatement[T any]() (string, error) {
	var zeroValue T
	t := reflect.TypeOf(zeroValue)

	if t.Kind() != reflect.Struct {
		return "", errors.New(_NOT_A_STRUCT_ERROR)
	}

	statement := "CREATE TABLE '" + t.Name() + "'('"
	for i := range t.NumField() {
		field := t.Field(i)
		switch {
		case field.Type.Kind() == reflect.Bool:
			statement += field.Name + "' INTEGER NOT NULL CHECK('" + field.Name + "' IN (0,1)),'"
		case field.Type.Kind() == reflect.Float32:
			fallthrough
		case field.Type.Kind() == reflect.Float64:
			statement += field.Name + "' REAL NOT NULL,'"
		case field.Type.Kind() == reflect.Int:
			fallthrough
		case field.Type.Kind() == reflect.Int8:
			fallthrough
		case field.Type.Kind() == reflect.Int16:
			fallthrough
		case field.Type.Kind() == reflect.Int32:
			fallthrough
		case field.Type.Kind() == reflect.Int64:
			statement += field.Name + "' INTEGER NOT NULL,'"
		// case field.Type.Kind() == reflect.Pointer:
		case field.Type.Kind() == reflect.String:
			statement += field.Name + "' TEXT NOT NULL,'"
		default:
			return "", errors.New(_UNSUPPORTED_FIELD_TYPE_ERROR)
		}
	}
	if strings.HasSuffix(statement, "('") {
		return "", errors.New(_NO_FIELDS_ERROR)
	} else {
		statement = strings.TrimSuffix(statement, ",'")
	}

	statement += ");"
	return statement, nil
}

func createTable[T any](db *sql.DB) error {
	createStatement, err := buildCreateTableStatement[T]()
	if err != nil {
		return err
	}

	_, err = db.Exec(createStatement)
	return err
}

func insert[T any](row []string, tx *sql.Tx) error {
	var zeroValue T
	t := reflect.TypeOf(zeroValue)

	if t.Kind() != reflect.Struct {
		return errors.New(_NOT_A_STRUCT_ERROR)
	}

	statement := "INSERT INTO " + t.Name() + " VALUES ("
	args := []any{}
	for i := range t.NumField() {
		statement += "?,"
		args = append(args, any(row[i]))
	}
	if strings.HasSuffix(statement, "(") {
		return errors.New(_NO_FIELDS_ERROR)
	} else {
		statement = strings.TrimSuffix(statement, ",")
	}
	statement += ");"

	_, err := tx.Exec(statement, args...)
	return err
}

func insertRows[T any](db *sql.DB, csvFilePath string, options CsvOptions) (*sql.DB, []error) {
	if _, err := os.Stat(csvFilePath); os.IsNotExist(err) {
		return nil, []error{errors.New(_FILE_PATH_DOES_NOT_EXIST_ERROR)}
	} else if err != nil {
		return nil, []error{err}
	}

	file, err := os.Open(csvFilePath)
	if err != nil {
		return nil, []error{err}
	}
	defer file.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, []error{err}
	}
	defer tx.Rollback()

	errs := []error{}
	reader := csv.NewReader(file)
	if int(options.Delimiter) != 0 {
		reader.Comma = options.Delimiter
	}
	for {
		// Skip first loop if there's a header row.
		if options.HasHeaderRow {
			_, err = reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, []error{err}
			}
			options.HasHeaderRow = false
		} else {
			row, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, []error{err}
			}

			err = insert[T](row, tx)
			if err != nil {
				if len(errs) < 25 {
					errs = append(errs, err)
				}
			}
		}
	}

	if len(errs) > 0 {
		return nil, errs
	}

	err = tx.Commit()
	if err != nil {
		return nil, []error{err}
	}

	return db, nil
}
