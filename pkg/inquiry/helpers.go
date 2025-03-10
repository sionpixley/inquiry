package inquiry

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
)

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

func convertToStruct[T any](row []string, zeroValue T) ([]T, error) {

}

func createTable[T any](zeroValue T) error {
	createStatement, err := buildCreateTableStatement(zeroValue)
	if err != nil {
		return err
	}

	_, err = Database.Exec(createStatement)
	return err
}

func insert[T any](obj T, tx *sql.Tx) error {

}

func insertRows[T any](csvFilePath string, hasHeaderRow bool, zeroValue T) []error {
	if _, err := os.Stat(csvFilePath); os.IsNotExist(err) {
		return []error{errors.New("inquiry error: file path does not exist")}
	} else if err != nil {
		return []error{err}
	}

	file, err := os.Open(csvFilePath)
	if err != nil {
		return []error{err}
	}
	defer file.Close()

	tx, err := Database.Begin()
	if err != nil {
		return []error{err}
	}
	defer tx.Rollback()

	wg := sync.WaitGroup{}
	errs := make(chan error, 25)
	reader := csv.NewReader(file)
	for {
		// Skip first loop if there's a header row.
		if hasHeaderRow {
			_, err = reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return []error{err}
			}
			hasHeaderRow = false
		} else {
			row, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return []error{err}
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				obj, err := convertToStruct(row, zeroValue)
				if err != nil {
					select {
					// Only puts error in the channel if it can (not full).
					case errs <- err:
						return
					default:
						// Do nothing with error because channel is full.
						return
					}
				}

				err = insert(obj, tx)
				if err != nil {
					select {
					// Only puts error in the channel if it can (not full).
					case errs <- err:
						return
					default:
						// Do nothing with error because channel is full.
						return
					}
				}
			}()
		}
	}

	// rows, err := readCsv(csvFilePath)
	// if err != nil {
	// 	return err
	// }

	// structs, err := mapToStruct[T](rows)
	// if err != nil {
	// 	return err
	// }

	// tx, err := Database.Begin()
	// if err != nil {
	// 	return err
	// }
	// defer tx.Rollback()

	// if hasHeaderRow {
	// 	_, err = tx.Exec(".import -skip 1 " + csvFilePath + " " + reflect.TypeOf(zeroValue).Name())
	// 	if err != nil {
	// 		return err
	// 	}
	// } else {
	// 	_, err = tx.Exec(".import " + csvFilePath + " " + reflect.TypeOf(zeroValue).Name())
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// return tx.Commit()
}
