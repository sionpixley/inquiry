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
	_NO_FIELDS_ERROR                string = "inquiry error: struct has no fields"
	_NOT_A_STRUCT_ERROR             string = "inquiry error: generic type provided is not a struct"
	_UNSUPPORTED_FIELD_TYPE_ERROR   string = "inquiry error: unsupported field type"
)

func buildCreateTableStatement(t reflect.Type) (string, error) {
	if t.Kind() != reflect.Struct {
		return "", errors.New(_NOT_A_STRUCT_ERROR)
	}

	builder := strings.Builder{}
	builder.WriteString("CREATE TABLE '")
	builder.WriteString(t.Name())
	builder.WriteString("'('")
	for i := range t.NumField() {
		field := t.Field(i)
		switch {
		case field.Type.Kind() == reflect.Bool:
			builder.WriteString(field.Name)
			builder.WriteString("' INTEGER NOT NULL CHECK('")
			builder.WriteString(field.Name)
			builder.WriteString("' IN (0,1)),'")
		case field.Type.Kind() == reflect.Float32:
			fallthrough
		case field.Type.Kind() == reflect.Float64:
			builder.WriteString(field.Name)
			builder.WriteString("' REAL NOT NULL,'")
		case field.Type.Kind() == reflect.Int:
			fallthrough
		case field.Type.Kind() == reflect.Int8:
			fallthrough
		case field.Type.Kind() == reflect.Int16:
			fallthrough
		case field.Type.Kind() == reflect.Int32:
			fallthrough
		case field.Type.Kind() == reflect.Int64:
			builder.WriteString(field.Name)
			builder.WriteString("' INTEGER NOT NULL,'")
		case field.Type.Kind() == reflect.Pointer:
			f := field.Type.Elem()
			switch {
			case f.Kind() == reflect.Bool:
				builder.WriteString(field.Name)
				builder.WriteString("' INTEGER NULL CHECK('")
				builder.WriteString(field.Name)
				builder.WriteString("' IN (0,1)),'")
			case f.Kind() == reflect.Float32:
				fallthrough
			case f.Kind() == reflect.Float64:
				builder.WriteString(field.Name)
				builder.WriteString("' REAL NULL,'")
			case f.Kind() == reflect.Int:
				fallthrough
			case f.Kind() == reflect.Int8:
				fallthrough
			case f.Kind() == reflect.Int16:
				fallthrough
			case f.Kind() == reflect.Int32:
				fallthrough
			case f.Kind() == reflect.Int64:
				builder.WriteString(field.Name)
				builder.WriteString("' INTEGER NULL,'")
			case f.Kind() == reflect.String:
				builder.WriteString(field.Name)
				builder.WriteString("' TEXT NULL,'")
			default:
				return "", errors.New(_UNSUPPORTED_FIELD_TYPE_ERROR)
			}
		case field.Type.Kind() == reflect.String:
			builder.WriteString(field.Name)
			builder.WriteString("' TEXT NOT NULL,'")
		default:
			return "", errors.New(_UNSUPPORTED_FIELD_TYPE_ERROR)
		}
	}

	statement := builder.String()
	if strings.HasSuffix(statement, "('") {
		return "", errors.New(_NO_FIELDS_ERROR)
	} else {
		statement = strings.TrimSuffix(statement, ",'")
	}

	statement += ");"
	return statement, nil
}

func createTable[T any](db *sql.DB) (reflect.Type, error) {
	var zeroValue T
	t := reflect.TypeOf(zeroValue)

	createStatement, err := buildCreateTableStatement(t)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(createStatement)
	return t, err
}

func insert(tx *sql.Tx, statement string, row []string) error {
	args := []any{}
	for _, item := range row {
		if trimmedStr := strings.TrimSpace(item); trimmedStr == "" || trimmedStr == "null" || trimmedStr == "NULL" {
			args = append(args, nil)
		} else {
			args = append(args, any(item))
		}
	}

	_, err := tx.Exec(statement, args...)
	return err
}

func insertRows(db *sql.DB, csvFilePath string, t reflect.Type, options CsvOptions) (*sql.DB, error) {
	if _, err := os.Stat(csvFilePath); os.IsNotExist(err) {
		return nil, errors.New(_FILE_PATH_DOES_NOT_EXIST_ERROR)
	} else if err != nil {
		return nil, err
	}

	file, err := os.Open(csvFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	statement, err := prepareStatement(t)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

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
				return nil, err
			}
			options.HasHeaderRow = false
		} else {
			row, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}

			err = insert(tx, statement, row)
			if err != nil {
				return nil, err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func prepareStatement(t reflect.Type) (string, error) {
	if t.Kind() != reflect.Struct {
		return "", errors.New(_NOT_A_STRUCT_ERROR)
	} else if t.NumField() == 0 {
		return "", errors.New(_NO_FIELDS_ERROR)
	}

	builder := strings.Builder{}
	builder.WriteString("INSERT INTO '")
	builder.WriteString(t.Name())
	builder.WriteString("' VALUES (")
	for range t.NumField() {
		builder.WriteString("?,")
	}

	statement := builder.String()
	statement = strings.TrimSuffix(statement, ",")
	statement += ");"

	return statement, nil
}
