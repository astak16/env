package env

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"time"
)

type Options struct {
	Environment         map[string]string
	TagName             string
	DefaultValueTagName string
	PrefixTagName       string
	Prefix              string
	FuncMap             map[reflect.Type]ParserFunc
}

func defaultOptions() Options {
	return Options{
		TagName:             "env",
		DefaultValueTagName: "envDefault",
		PrefixTagName:       "envPrefix",
		Environment:         toMap(os.Environ()),
		FuncMap:             defaultTypeParsers(),
	}
}

func optionsWithEnvPrefix(field reflect.StructField, opts Options) Options {
	return Options{
		Environment:         opts.Environment,
		TagName:             opts.TagName,
		PrefixTagName:       opts.PrefixTagName,
		Prefix:              opts.Prefix + field.Tag.Get(opts.PrefixTagName),
		DefaultValueTagName: opts.DefaultValueTagName,
		FuncMap:             opts.FuncMap,
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
