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
	if !refField.CanSet() {
		return nil
	}

	params, err := parseFieldParams(refTypeField, opts)
	if err != nil {
		return err
	}

	if err := processField(refField, refTypeField, opts, params); err != nil {
		return err
	}
	if refField.Kind() == reflect.Struct {
		return doParse(refField, processField, optionsWithEnvPrefix(refTypeField, opts))
	}

	return nil
}

func parseFieldParams(field reflect.StructField, opts Options) (FieldParams, error) {
	ownKey, _ := parseKeyForOption(field.Tag.Get(opts.TagName))
	defaultValue, hasDefaultValue := field.Tag.Lookup(opts.DefaultValueTagName)
	result := FieldParams{
		Key:             opts.Prefix + ownKey,
		DefaultValue:    defaultValue,
		HasDefaultValue: hasDefaultValue,
	}

	return result, nil
}

func setField(refField reflect.Value, refTypeField reflect.StructField, opts Options, fieldParams FieldParams) error {
	value, err := get(fieldParams, opts)
	if err != nil {
		return err
	}
	if value != "" {
		return set(refField, refTypeField, value, opts.FuncMap)
	}
	return nil
}

func get(fieldParams FieldParams, opts Options) (val string, err error) {
	value, exists, isDefault := getOr(fieldParams.Key, fieldParams.DefaultValue, fieldParams.HasDefaultValue, opts.Environment)
	if isDefault {
	}
	if exists {
	}
	return value, nil
}

func getOr(key, defaultValue string, defExists bool, env map[string]string) (val string, exists bool, isDefault bool) {
	value, exists := env[key]
	switch {
	case (!exists || key == "") && defExists:
		return defaultValue, true, true
	case !exists:
		return "", false, false
	}
	return value, true, false
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
	separator := sf.Tag.Get("envSeparator")
	if separator == "" {
		separator = ","
	}
	parts := strings.Split(value, separator)
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
