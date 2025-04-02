package models

import "reflect"

type FieldTagMap struct {
	Field reflect.StructField
	Tag   Tag
}

type Tag int
