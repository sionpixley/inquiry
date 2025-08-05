package constants

import "github.com/sionpixley/inquiry/internal/models"

const (
	FilePathDoesNotExistError string = "inquiry error: file path does not exist"
	NoFieldsError             string = "inquiry error: struct has no fields"
	NotAStructError           string = "inquiry error: generic type provided is not a struct"
	UnsupportedFileTypeError  string = "inquiry error: unsupported field type"
)

const (
	OtherTag models.Tag = iota
	IndexTag
	PrimaryKeyTag
	UniqueTag
)
