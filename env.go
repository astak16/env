package env

import (
	"reflect"
	"strings"
)

func Parse(v interface{}) error {
	return parseInternal(v, setField, defaultOptions())
}

func parseInternal(v interface{}, processField processFieldFn, opts Options) error {
	ptrRef := reflect.ValueOf(v)
	if ptrRef.Kind() != reflect.Ptr {
		return nil
	}
	ref := ptrRef.Elem()
	if ref.Kind() != reflect.Struct {
		return nil
	}
	return doParse(ref, processField, opts)
}

func doParse(ref reflect.Value, processField processFieldFn, opts Options) error {
	refType := ref.Type()

	for i := 0; i < refType.NumField(); i++ {
		refField := ref.Field(i)
		refTypeField := refType.Field(i)
		if err := doParseField(refField, refTypeField, processField, opts); err != nil {
			return nil
		}
	}
	return nil
}

func doParseField(refField reflect.Value, refTypeField reflect.StructField, processField processFieldFn, opts Options) error {
	params, err := parseFieldParams(refTypeField, opts)
	if err != nil {
		return err
	}

	return processField(refField, refTypeField, opts, params)
}

func parseFieldParams(field reflect.StructField, opts Options) (FieldParams, error) {
	ownKey, _ := parseKeyForOption(field.Tag.Get(opts.TagName))
	result := FieldParams{
		Key: ownKey,
	}

	return result, nil
}

func setField(refField reflect.Value, refTypeField reflect.StructField, opts Options, fieldParams FieldParams) error {
	value, ok := opts.Environment[fieldParams.Key]
	if ok && value != "" {
		return set(refField, refTypeField, value, opts.FuncMap)
	}
	return nil
}

func set(field reflect.Value, sf reflect.StructField, value string, funcMap map[reflect.Type]ParserFunc) error {
	typee := sf.Type
	fieldee := field
	if typee.Kind() == reflect.Ptr {
		typee = typee.Elem()
		fieldee.Set(reflect.New(field.Type().Elem()))
		fieldee = field.Elem()
	}
	parserFunc, ok := funcMap[typee]
	if !ok {
		parserFunc, ok = defaultBuiltInParsers[typee.Kind()]
	}
	if ok {
		val, err := parserFunc(value)
		if err != nil {
			return nil
		}
		fieldee.Set(reflect.ValueOf(val).Convert(typee))
		return nil
	}
	if field.Kind() == reflect.Slice {
		return handleSlice(field, value, sf, funcMap)
	}
	return nil
}

func handleSlice(field reflect.Value, value string, sf reflect.StructField, funcMap map[reflect.Type]ParserFunc) error {
	parts := strings.Split(value, ",")
	typee := sf.Type.Elem()
	if typee.Kind() == reflect.Ptr {
		typee = typee.Elem()
	}

	parserFunc, ok := funcMap[typee]
	if !ok {
		parserFunc, ok = defaultBuiltInParsers[typee.Kind()]
	}

	result := reflect.MakeSlice(sf.Type, 0, len(parts))
	for _, part := range parts {
		r, _ := parserFunc(part)
		v := reflect.ValueOf(r)
		if sf.Type.Elem().Kind() == reflect.Ptr {
			v = reflect.New(typee)
			v.Elem().Set(reflect.ValueOf(r))
		}
		result = reflect.Append(result, v)
	}
	field.Set(result)
	return nil
}
