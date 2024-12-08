package env

import "reflect"

type OnSetFn func(tag string, value interface{}, isDefault bool)

type ParserFunc func(v string) (interface{}, error)

type processFieldFn func(refField reflect.Value, refTypeField reflect.StructField, opts Options, fieldParams FieldParams) error

type Options struct {
	Environment           map[string]string
	TagName               string
	DefaultValueTagName   string
	PrefixTagName         string
	Prefix                string
	FuncMap               map[reflect.Type]ParserFunc
	rawEnvVars            map[string]string
	UseFieldNameByDefault bool
	RequiredIfNoDef       bool
	OnSet                 OnSetFn
}

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
