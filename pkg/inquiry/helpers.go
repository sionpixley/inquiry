package inquiry

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/sionpixley/inquiry/internal/constants"
	"github.com/sionpixley/inquiry/internal/models"
)

func buildCreateTableStatement(t reflect.Type) (string, []models.FieldTagMap, error) {
	if t.Kind() != reflect.Struct {
		return "", nil, errors.New(constants.NotAStructError)
	} else if t.NumField() == 0 {
		return "", nil, errors.New(constants.NoFieldsError)
	}

	indexes := []models.FieldTagMap{}
	constraints := []models.FieldTagMap{}

	var builder strings.Builder
	builder.WriteString("CREATE TABLE '")
	builder.WriteString(t.Name())
	builder.WriteString("'(")
	for i := range t.NumField() {
		field := t.Field(i)

		tags := convertToTags(strings.Split(trimAndToLowerStr(field.Tag.Get("inquiry")), ","))
		for _, tag := range tags {
			switch tag {
			case constants.IndexTag:
				indexes = append(indexes, models.FieldTagMap{Field: field, Tag: tag})
			case constants.PrimaryKeyTag:
				constraints = append(constraints, models.FieldTagMap{Field: field, Tag: tag})
			case constants.UniqueTag:
				if field.Type.Kind() == reflect.Pointer {
					indexes = append(indexes, models.FieldTagMap{Field: field, Tag: tag})
				} else {
					constraints = append(constraints, models.FieldTagMap{Field: field, Tag: tag})
				}
			default:
				// Do nothing.
			}
		}

		switch field.Type.Kind() {
		case reflect.Bool:
			builder.WriteString("'")
			builder.WriteString(field.Name)
			builder.WriteString("' INTEGER NOT NULL CHECK('")
			builder.WriteString(field.Name)
			builder.WriteString("' IN (0,1)),")
		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			builder.WriteString("'")
			builder.WriteString(field.Name)
			builder.WriteString("' REAL NOT NULL,")
		case reflect.Int:
			fallthrough
		case reflect.Int8:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			builder.WriteString("'")
			builder.WriteString(field.Name)
			builder.WriteString("' INTEGER NOT NULL,")
		case reflect.Pointer:
			f := field.Type.Elem()
			switch f.Kind() {
			case reflect.Bool:
				builder.WriteString("'")
				builder.WriteString(field.Name)
				builder.WriteString("' INTEGER NULL CHECK('")
				builder.WriteString(field.Name)
				builder.WriteString("' IN (0,1)),")
			case reflect.Float32:
				fallthrough
			case reflect.Float64:
				builder.WriteString("'")
				builder.WriteString(field.Name)
				builder.WriteString("' REAL NULL,")
			case reflect.Int:
				fallthrough
			case reflect.Int8:
				fallthrough
			case reflect.Int16:
				fallthrough
			case reflect.Int32:
				fallthrough
			case reflect.Int64:
				builder.WriteString("'")
				builder.WriteString(field.Name)
				builder.WriteString("' INTEGER NULL,")
			case reflect.String:
				builder.WriteString("'")
				builder.WriteString(field.Name)
				builder.WriteString("' TEXT NULL,")
			default:
				return "", nil, errors.New(constants.UnsupportedFileTypeError)
			}
		case reflect.String:
			builder.WriteString("'")
			builder.WriteString(field.Name)
			builder.WriteString("' TEXT NOT NULL,")
		default:
			return "", nil, errors.New(constants.UnsupportedFileTypeError)
		}
	}

	for _, constraint := range constraints {
		if constraint.Tag == constants.PrimaryKeyTag {
			builder.WriteString("CONSTRAINT ")
			builder.WriteString("PK_")
			builder.WriteString(t.Name())
			builder.WriteString("_")
			builder.WriteString(constraint.Field.Name)
			builder.WriteString(" PRIMARY KEY('")
			builder.WriteString(constraint.Field.Name)
			builder.WriteString("'),")
		} else {
			builder.WriteString("CONSTRAINT ")
			builder.WriteString("Unique_")
			builder.WriteString(t.Name())
			builder.WriteString("_")
			builder.WriteString(constraint.Field.Name)
			builder.WriteString(" UNIQUE('")
			builder.WriteString(constraint.Field.Name)
			builder.WriteString("'),")
		}
	}

	statement := builder.String()
	statement = strings.TrimSuffix(statement, ",")
	statement += ");"
	return statement, indexes, nil
}

func convertToTags(t []string) []models.Tag {
	tags := make([]models.Tag, len(t))
	for i, strT := range t {
		switch strT {
		case "index":
			tags[i] = constants.IndexTag
		case "primarykey":
			tags[i] = constants.PrimaryKeyTag
		case "unique":
			tags[i] = constants.UniqueTag
		default:
			tags[i] = constants.OtherTag
		}
	}
	return tags
}

func createTable[T any](tx *sql.Tx) (reflect.Type, error) {
	var zeroValue T
	t := reflect.TypeOf(zeroValue)

	createStatement, indexes, err := buildCreateTableStatement(t)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(createStatement)
	if err != nil {
		return nil, err
	}

	for _, index := range indexes {
		var builder strings.Builder
		builder.WriteString("CREATE ")
		if index.Tag == constants.IndexTag {
			builder.WriteString("INDEX NonClustered_")
			builder.WriteString(t.Name())
			builder.WriteString("_")
			builder.WriteString(index.Field.Name)
			builder.WriteString(" ON '")
			builder.WriteString(t.Name())
			builder.WriteString("'('")
			builder.WriteString(index.Field.Name)
			builder.WriteString("');")
		} else {
			builder.WriteString("UNIQUE INDEX Unique_")
			builder.WriteString(t.Name())
			builder.WriteString("_")
			builder.WriteString(index.Field.Name)
			builder.WriteString(" ON '")
			builder.WriteString(t.Name())
			builder.WriteString("'('")
			builder.WriteString(index.Field.Name)
			builder.WriteString("') WHERE '")
			builder.WriteString(index.Field.Name)
			builder.WriteString("' IS NOT NULL;")
		}

		_, err = tx.Exec(builder.String())
		if err != nil {
			return nil, err
		}
	}

	return t, err
}

func insert(tx *sql.Tx, statement string, row []string, t reflect.Type) error {
	args := make([]any, t.NumField())
	for i := range t.NumField() {
		if trimmedStr := strings.TrimSpace(row[i]); (trimmedStr == "" || trimmedStr == "null" || trimmedStr == "NULL") && t.Field(i).Type.Kind() == reflect.Pointer {
			args[i] = nil
		} else {
			args[i] = any(row[i])
		}
	}

	_, err := tx.Exec(statement, args...)
	return err
}

func insertRows(tx *sql.Tx, csvFilePath string, t reflect.Type, options CsvOptions) error {
	if _, err := os.Stat(csvFilePath); os.IsNotExist(err) {
		return errors.New(constants.FilePathDoesNotExistError)
	} else if err != nil {
		return err
	}

	file, err := os.Open(csvFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	statement, err := prepareStatement(t)
	if err != nil {
		return err
	}

	reader := csv.NewReader(file)
	reader.LazyQuotes = options.UseLazyQuotes
	reader.TrimLeadingSpace = options.TrimLeadingSpace
	if int(options.Delimiter) != 0 {
		reader.Comma = options.Delimiter
	}
	if int(options.CommentCharacter) != 0 {
		reader.Comment = options.CommentCharacter
	}

	for {
		// Skip first loop if there's a header row.
		if options.HasHeaderRow {
			_, err = reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			options.HasHeaderRow = false
		} else {
			row, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			err = insert(tx, statement, row, t)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func prepareStatement(t reflect.Type) (string, error) {
	if t.Kind() != reflect.Struct {
		return "", errors.New(constants.NotAStructError)
	} else if t.NumField() == 0 {
		return "", errors.New(constants.NoFieldsError)
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

func trimAndToLowerStr(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
