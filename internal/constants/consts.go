package constants

import "github.com/sionpixley/inquiry/internal/models"

const (
	FILE_PATH_DOES_NOT_EXIST_ERROR string = "inquiry error: file path does not exist"
	NO_FIELDS_ERROR                string = "inquiry error: struct has no fields"
	NOT_A_STRUCT_ERROR             string = "inquiry error: generic type provided is not a struct"
	UNSUPPORTED_FIELD_TYPE_ERROR   string = "inquiry error: unsupported field type"
)

const (
	NA_TAG models.Tag = iota
	INDEX_TAG
	UNIQUE_TAG
)
