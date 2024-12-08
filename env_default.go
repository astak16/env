package env

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"time"
)

func (opts *Options) getRawEnv(s string) string {
	val := opts.rawEnvVars[s]
	if val == "" {
		val = opts.Environment[s]
	}
	return os.Expand(val, opts.getRawEnv)
}

func defaultOptions() Options {
	return Options{
		TagName:             "env",
		DefaultValueTagName: "envDefault",
		PrefixTagName:       "envPrefix",
		Environment:         toMap(os.Environ()),
		FuncMap:             defaultTypeParsers(),
		rawEnvVars:          make(map[string]string),
	}
}

func mergeOptions(target, source *Options) {
	targetPtr := reflect.ValueOf(target).Elem()
	sourcePtr := reflect.ValueOf(source).Elem()

	targetType := targetPtr.Type()
	for i := 0; i < targetPtr.NumField(); i++ {
		targetField := targetPtr.Field(i)
		sourceField := sourcePtr.FieldByName(targetType.Field(i).Name)
		// 如果 targetField 可以设置，并且 sourceField 不是零值，就把 sourceField 的值更新到 targetField
		if targetField.CanSet() && !isZero(sourceField) {
			switch targetField.Kind() {
			case reflect.Map:
				// 遍历 sourceFiled 的 map，将 sourceFiled 的每一项设置到 targetField
				iter := sourceField.MapRange()
				for iter.Next() {
					targetField.SetMapIndex(iter.Key(), iter.Value())
				}
			default:
				targetField.Set(sourceField)
			}
		}
	}
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	default:
		zero := reflect.Zero(v.Type())
		return v.Interface() == zero.Interface()
	}
}

// opt 是自定义的 options，如果自定义的 opt 没有对应的属性，就用默认的 defOptions
func customOptions(opts Options) Options {
	defOpts := defaultOptions()
	mergeOptions(&defOpts, &opts)
	return defOpts
}

func optionsWithEnvPrefix(field reflect.StructField, opts Options) Options {
	return Options{
		Environment:           opts.Environment,
		TagName:               opts.TagName,
		PrefixTagName:         opts.PrefixTagName,
		Prefix:                opts.Prefix + field.Tag.Get(opts.PrefixTagName),
		DefaultValueTagName:   opts.DefaultValueTagName,
		FuncMap:               opts.FuncMap,
		rawEnvVars:            opts.rawEnvVars,
		RequiredIfNoDef:       opts.RequiredIfNoDef,
		UseFieldNameByDefault: opts.UseFieldNameByDefault,
		OnSet:                 opts.OnSet,
	}
}

func defaultTypeParsers() map[reflect.Type]ParserFunc {
	return map[reflect.Type]ParserFunc{
		reflect.TypeOf(url.URL{}):       parseURL,
		reflect.TypeOf(time.Nanosecond): parseDuration,
		reflect.TypeOf(time.Location{}): parseLocation,
	}
}

func parseURL(v string) (interface{}, error) {
	u, err := url.Parse(v)
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL: %w", err)
	}
	return *u, nil
}

func parseDuration(v string) (interface{}, error) {
	d, err := time.ParseDuration(v)
	if err != nil {
		return nil, fmt.Errorf("unable to parse duration: %w", err)
	}
	return d, err
}

func parseLocation(v string) (interface{}, error) {
	loc, err := time.LoadLocation(v)
	if err != nil {
		return nil, fmt.Errorf("unable to parse location: %w", err)
	}
	return *loc, nil
}

var defaultBuiltInParsers = map[reflect.Kind]ParserFunc{
	reflect.Bool: func(v string) (interface{}, error) {
		return strconv.ParseBool(v)
	},
	reflect.String: func(v string) (interface{}, error) {
		return v, nil
	},
	reflect.Int: func(v string) (interface{}, error) {
		i, err := strconv.ParseInt(v, 10, 32)
		return int(i), err
	},
	reflect.Int16: func(v string) (interface{}, error) {
		i, err := strconv.ParseInt(v, 10, 16)
		return int16(i), err
	},
	reflect.Int32: func(v string) (interface{}, error) {
		i, err := strconv.ParseInt(v, 10, 32)
		return int32(i), err
	},
	reflect.Int64: func(v string) (interface{}, error) {
		return strconv.ParseInt(v, 10, 64)
	},
	reflect.Int8: func(v string) (interface{}, error) {
		i, err := strconv.ParseInt(v, 10, 8)
		return int8(i), err
	},
	reflect.Uint: func(v string) (interface{}, error) {
		i, err := strconv.ParseUint(v, 10, 32)
		return uint(i), err
	},
	reflect.Uint16: func(v string) (interface{}, error) {
		i, err := strconv.ParseUint(v, 10, 16)
		return uint16(i), err
	},
	reflect.Uint32: func(v string) (interface{}, error) {
		i, err := strconv.ParseUint(v, 10, 32)
		return uint32(i), err
	},
	reflect.Uint64: func(v string) (interface{}, error) {
		i, err := strconv.ParseUint(v, 10, 64)
		return i, err
	},
	reflect.Uint8: func(v string) (interface{}, error) {
		i, err := strconv.ParseUint(v, 10, 8)
		return uint8(i), err
	},
	reflect.Float64: func(v string) (interface{}, error) {
		return strconv.ParseFloat(v, 64)
	},
	reflect.Float32: func(v string) (interface{}, error) {
		f, err := strconv.ParseFloat(v, 32)
		return float32(f), err
	},
}
