package env

import (
	"errors"
	"fmt"
	"os"
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
	defaultValue, hasDefaultValue := field.Tag.Lookup(opts.DefaultValueTagName)
	result := FieldParams{
		OwnKey:          ownKey,
		Key:             opts.Prefix + ownKey,
		DefaultValue:    defaultValue,
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

	if fieldParams.Required && !exists {
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

	if isDefault {
	}

	return val, nil
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

func handleMap(field reflect.Value, value string, sf reflect.StructField, funcMap map[reflect.Type]ParserFunc) error {
	// 获取 key 的解析函数
	keyType := sf.Type.Key()
	keyParserFunc, ok := getParserFunc(funcMap, keyType)
	if !ok {
		return nil
	}

	// 获取 value 的解析函数
	elemType := sf.Type.Elem()
	elemParserFunc, ok := getParserFunc(funcMap, elemType)
	if !ok {
		return nil
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
	return parserFunc, true
}
