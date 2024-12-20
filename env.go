package env

import (
	"encoding"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
)

func Parse(v interface{}) error {
	return parseInternal(v, setField, defaultOptions())
}

func ParseWithOptions(v interface{}, opts Options) error {
	return parseInternal(v, setField, customOptions(opts))
}

func ParseAs[T any]() (T, error) {
	var t T
	err := Parse(&t)
	return t, err
}

func ParseAsWithOptions[T any](opts Options) (T, error) {
	var t T
	err := ParseWithOptions(&t, opts)
	return t, err
}

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func GetFieldParams(v interface{}) ([]FieldParams, error) {
	return GetFieldParamsWithOptions(v, defaultOptions())
}

func GetFieldParamsWithOptions(v interface{}, opts Options) ([]FieldParams, error) {
	var result []FieldParams
	err := parseInternal(
		v,
		func(_ reflect.Value, _ reflect.StructField, _ Options, fieldParams FieldParams) error {
			if fieldParams.OwnKey != "" {
				result = append(result, fieldParams)
			}
			return nil
		},
		customOptions(opts),
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func parseInternal(v interface{}, processField processFieldFn, opts Options) error {
	ptrRef := reflect.ValueOf(v)
	if ptrRef.Kind() != reflect.Ptr {
		return newAggregateError(NotStructPtrError{})
	}
	ref := ptrRef.Elem()
	if ref.Kind() != reflect.Struct {
		return newAggregateError(NotStructPtrError{})
	}
	return doParse(ref, processField, opts)
}

func doParse(ref reflect.Value, processField processFieldFn, opts Options) error {
	refType := ref.Type()
	var agrErr AggregateError
	for i := 0; i < refType.NumField(); i++ {
		refField := ref.Field(i)
		refTypeField := refType.Field(i)
		if err := doParseField(refField, refTypeField, processField, opts); err != nil {
			var val AggregateError
			if errors.As(err, &val) {
				agrErr.Errors = append(agrErr.Errors, val.Errors...)
			} else {
				agrErr.Errors = append(agrErr.Errors, err)
			}
		}
	}
	if len(agrErr.Errors) == 0 {
		return nil
	}

	return agrErr
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

	// 对 isInvalidPtr 说明
	// type Internal struct{}
	// type Config struct{ Internal *Internal }
	// config1 := &Config{}                       isInvalidPtr(refField) == true
	// config2 := &Config{Internal: &Internal{}}  isInvalidPtr(refField) == false
	if params.Init && isInvalidPtr(refField) {
		refField.Set(reflect.New(refField.Type().Elem()))
		refField = refField.Elem()
	}

	if refField.Kind() == reflect.Ptr && refField.Elem().Kind() == reflect.Struct {
		return doParse(refField.Elem(), processField, optionsWithEnvPrefix(refTypeField, opts))
	}

	if refField.Kind() == reflect.Struct {
		return doParse(refField, processField, optionsWithEnvPrefix(refTypeField, opts))
	}

	return nil
}

func parseFieldParams(field reflect.StructField, opts Options) (FieldParams, error) {
	ownKey, tags := parseKeyForOption(field.Tag.Get(opts.TagName))

	if ownKey == "" && opts.UseFieldNameByDefault {
		ownKey = toEnvName(field.Name)
	}

	defaultValue, hasDefaultValue := field.Tag.Lookup(opts.DefaultValueTagName)
	result := FieldParams{
		OwnKey:          ownKey,
		Key:             opts.Prefix + ownKey,
		DefaultValue:    defaultValue,
		Required:        opts.RequiredIfNoDef,
		HasDefaultValue: hasDefaultValue,
	}

	for _, tag := range tags {
		switch tag {
		case "":
			continue
		case "required":
			result.Required = true
		case "notEmpty":
			result.NotEmpty = true
		case "init":
			result.Init = true
		case "expand":
			result.Expand = true
		case "unset":
			result.Unset = true
		case "file":
			result.LoadFile = true
		default:
			return FieldParams{}, newNoSupportedTagOptionError(tag)
		}
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
	val, exists, isDefault := getOr(fieldParams.Key, fieldParams.DefaultValue, fieldParams.HasDefaultValue, opts.Environment)

	if fieldParams.Expand {
		val = os.Expand(val, opts.getRawEnv)
	}

	opts.rawEnvVars[fieldParams.OwnKey] = val

	if fieldParams.Unset {
		defer os.Unsetenv(fieldParams.Key)
	}

	if fieldParams.Required && !exists && fieldParams.OwnKey != "" {
		return "", newVarIsNotSetError(fieldParams.Key)
	}
	if fieldParams.NotEmpty && val == "" {
		return "", newEmptyVarError(fieldParams.Key)
	}

	if fieldParams.LoadFile && val != "" {
		filename := val
		val, err = getFromFile(filename)
		if err != nil {
			return "", newLoadFileContentError(filename, fieldParams.Key, err)
		}
	}

	if opts.OnSet != nil {
		if fieldParams.OwnKey != "" {
			opts.OnSet(fieldParams.Key, val, isDefault)
		}
	}

	return val, nil
}

func getOr(key, defaultValue string, defExists bool, env map[string]string) (val string, exists bool, isDefault bool) {
	value, exists := env[key]
	switch {
	case exists && value == "" && defExists:
		return defaultValue, true, true
	case (!exists || key == "") && defExists:
		return defaultValue, true, true
	case !exists:
		return "", false, false
	}
	return value, true, false
}

func set(field reflect.Value, sf reflect.StructField, value string, funcMap map[reflect.Type]ParserFunc) error {
	if tm := asTextUnmarshaler(field); tm != nil {
		if err := tm.UnmarshalText([]byte(value)); err != nil {
			return newParseError(sf, err)
		}
		return nil
	}

	typee := sf.Type
	fieldee := field
	if typee.Kind() == reflect.Ptr {
		typee = typee.Elem()
		fieldee.Set(reflect.New(field.Type().Elem()))
		fieldee = field.Elem()
	}

	parserFunc, ok := getParserFunc(funcMap, typee)
	if ok {
		val, err := parserFunc(value)
		if err != nil {
			return newParseError(sf, err)
		}
		fieldee.Set(reflect.ValueOf(val).Convert(typee))
		return nil
	}

	switch field.Kind() {
	case reflect.Slice:
		return handleSlice(field, value, sf, funcMap)
	case reflect.Map:
		return handleMap(field, value, sf, funcMap)
	}

	return newNoParserError(sf)
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

	if _, ok := reflect.New(typee).Interface().(encoding.TextUnmarshaler); ok {
		return parseTextUnmarshalers(field, parts, sf)
	}

	parserFunc, ok := getParserFunc(funcMap, typee)
	if !ok {
		return newNoParserError(sf)
	}

	result := reflect.MakeSlice(sf.Type, 0, len(parts))
	for _, part := range parts {
		r, err := parserFunc(part)
		if err != nil {
			return newParseError(sf, err)
		}
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

func handleMap(field reflect.Value, value string, sf reflect.StructField, funcMap map[reflect.Type]ParserFunc) error {
	// 获取 key 的解析函数
	keyType := sf.Type.Key()
	keyParserFunc, ok := getParserFunc(funcMap, keyType)
	if !ok {
		return newNoParserError(sf)
	}

	// 获取 value 的解析函数
	elemType := sf.Type.Elem()
	elemParserFunc, ok := getParserFunc(funcMap, elemType)
	if !ok {
		return newNoParserError(sf)
	}

	separator := sf.Tag.Get("envSeparator")
	if separator == "" {
		separator = ","
	}

	keyValSeparator := sf.Tag.Get("envKeyValSeparator")
	if keyValSeparator == "" {
		keyValSeparator = ":"
	}

	// 初始化 reflect.map
	result := reflect.MakeMap(sf.Type)
	//"k1:v1,k2:v2" => ["k1:v1", "k2,v2"]
	for _, part := range strings.Split(value, separator) {
		//"k1:v1" => ["k1", "v1"]
		pairs := strings.Split(part, keyValSeparator)
		if len(pairs) != 2 {
			return newParseError(sf, fmt.Errorf(`%q should be in "key%svalue" format`, part, keyValSeparator))
		}

		// 对 key 做转换
		key, err := keyParserFunc(pairs[0])
		if err != nil {
			return newParseError(sf, err)
		}

		// 对 value 做转换
		elem, err := elemParserFunc(pairs[1])
		if err != nil {
			return newParseError(sf, err)
		}

		// 设置 map 值
		result.SetMapIndex(reflect.ValueOf(key).Convert(keyType), reflect.ValueOf(elem).Convert(elemType))
	}

	// 设置字段值
	field.Set(result)
	return nil
}

func getParserFunc(funcMap map[reflect.Type]ParserFunc, typee reflect.Type) (ParserFunc, bool) {
	parserFunc, ok := funcMap[typee]
	if !ok {
		parserFunc, ok = defaultBuiltInParsers[typee.Kind()]
		if !ok {
			return nil, false
		}
	}
	return parserFunc, ok
}

func asTextUnmarshaler(field reflect.Value) encoding.TextUnmarshaler {
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
	} else if field.CanAddr() {
		field = field.Addr()
	}

	tm, ok := field.Interface().(encoding.TextUnmarshaler)
	if !ok {
		return nil
	}
	return tm
}

func parseTextUnmarshalers(field reflect.Value, data []string, sf reflect.StructField) error {
	s := len(data)
	elemType := field.Type().Elem()
	slice := reflect.MakeSlice(reflect.SliceOf(elemType), s, s)
	for i, v := range data {
		sv := slice.Index(i)
		tm := asTextUnmarshaler(sv)
		if err := tm.UnmarshalText([]byte(v)); err != nil {
			return newParseError(sf, err)
		}
		if sv.Kind() == reflect.Ptr {
			slice.Index(i).Set(sv)
		}
	}

	field.Set(slice)
	return nil
}
