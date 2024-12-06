package env

import "reflect"

type ParserFunc func(v string) (interface{}, error)

type processFieldFn func(refField reflect.Value, refTypeField reflect.StructField, opts Options, fieldParams FieldParams) error

type FieldParams struct {
	OwnKey          string
	Key             string
	DefaultValue    string
	HasDefaultValue bool
	Required        bool
	NotEmpty        bool
	Init            bool
	Expand          bool
	Unset           bool
	LoadFile        bool
}
