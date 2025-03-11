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
			return "", errors.New(_UNSUPPORTED_FIELD_TYPE_ERROR)
		}
	}
	if strings.HasSuffix(statement, "(") {
		return "", errors.New(_NO_FIELDS_ERROR)
	} else {
		statement = strings.TrimSuffix(statement, ",")
	}

	statement += ");"
	return statement, nil
}

// func convertToStruct[T any](row []string, zeroValue T) (T, error) {
// 	t := reflect.TypeOf(zeroValue)
// 	v := reflect.ValueOf(zeroValue)

// 	if t.Kind() != reflect.Struct {
// 		return zeroValue, errors.New(_NOT_A_STRUCT_ERROR)
// 	} else if t.NumField() != len(row) {
// 		return zeroValue, errors.New(_MISMATCH_NUM_OF_COLUMNS_ERROR)
// 	}

// 	for i := range t.NumField() {
// 		field := t.Field(i)
// 		f := v.FieldByName(field.Name)
// 		switch {
// 		case field.Type.Kind() == reflect.Bool:
// 			b, err := strconv.ParseBool(row[i])
// 			if err != nil {
// 				return zeroValue, err
// 			}
// 			f.SetBool(b)
// 		case field.Type.Kind() == reflect.Float32:
// 			float, err := strconv.ParseFloat(row[i], 32)
// 			if err != nil {
// 				return zeroValue, err
// 			}
// 			f.SetFloat(float)
// 		case field.Type.Kind() == reflect.Float64:
// 			float, err := strconv.ParseFloat(row[i], 64)
// 			if err != nil {
// 				return zeroValue, err
// 			}
// 			f.SetFloat(float)
// 		case field.Type.Kind() == reflect.Int:
// 			integer, err := strconv.Atoi(row[i])
// 			if err != nil {
// 				return zeroValue, err
// 			}
// 			f.SetInt(int64(integer))
// 		case field.Type.Kind() == reflect.Int8:
// 			integer, err := strconv.ParseInt(row[i], 10, 8)
// 			if err != nil {
// 				return zeroValue, err
// 			}
// 			f.SetInt(integer)
// 		case field.Type.Kind() == reflect.Int16:
// 			integer, err := strconv.ParseInt(row[i], 10, 16)
// 			if err != nil {
// 				return zeroValue, err
// 			}
// 			f.SetInt(integer)
// 		case field.Type.Kind() == reflect.Int32:
// 			integer, err := strconv.ParseInt(row[i], 10, 32)
// 			if err != nil {
// 				return zeroValue, err
// 			}
// 			f.SetInt(integer)
// 		case field.Type.Kind() == reflect.Int64:
// 			integer, err := strconv.ParseInt(row[i], 10, 64)
// 			if err != nil {
// 				return zeroValue, err
// 			}
// 			f.SetInt(integer)
// 		// case field.Type.Kind() == reflect.Pointer:
// 		case field.Type.Kind() == reflect.String:
// 			f.SetString(row[i])
// 		default:
// 			return zeroValue, errors.New(_UNSUPPORTED_FIELD_TYPE_ERROR)
// 		}
// 	}

// 	return zeroValue, nil
// }

func createTable[T any]() error {
	createStatement, err := buildCreateTableStatement[T]()
	if err != nil {
		return err
	}

	_, err = Database.Exec(createStatement)
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

func insertRows[T any](csvFilePath string, hasHeaderRow bool) []error {
	if _, err := os.Stat(csvFilePath); os.IsNotExist(err) {
		return []error{errors.New(_FILE_PATH_DOES_NOT_EXIST_ERROR)}
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
				err = insert[T](row, tx)
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

	wg.Wait()
	close(errs)

	if len(errs) > 0 {
		es := []error{}
		for e := range errs {
			es = append(es, e)
		}
		return es
	}

	err = tx.Commit()
	if err != nil {
		return []error{err}
	}

	return nil
}
