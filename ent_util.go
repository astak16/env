package env

import (
	"reflect"
	"strings"
)

func parseKeyForOption(key string) (string, []string) {
	opts := strings.Split(key, ",")
	return opts[0], opts[1:]
}

func isInvalidPtr(v reflect.Value) bool {
	return reflect.Ptr == v.Kind() && v.Elem().Kind() == reflect.Invalid
}
