package env

import (
	"fmt"
	"reflect"
	"strings"
)

type AggregateError struct {
	Errors []error
}

func newAggregateError(initErr error) error {
	return AggregateError{
		[]error{
			initErr,
		},
	}
}

func (e AggregateError) Error() string {
	var sb strings.Builder

	sb.WriteString("env:")

	for _, err := range e.Errors {
		sb.WriteString(fmt.Sprintf(" %v;", err.Error()))
	}

	return strings.TrimRight(sb.String(), ";")
}

// Is conforms with errors.Is.
func (e AggregateError) Is(err error) bool {
	for _, ie := range e.Errors {
		if reflect.TypeOf(ie) == reflect.TypeOf(err) {
			return true
		}
	}
	return false
}

type ParseError struct {
	Name string
	Type reflect.Type
	Err  error
}

func newParseError(sf reflect.StructField, err error) error {
	return ParseError{sf.Name, sf.Type, err}
}

func (e ParseError) Error() string {
	//parse error on field "MapStringString" of type "map[string]string": "k1" should be in "key:value" format
	return fmt.Sprintf("parse error on field %q of type %q: %v", e.Name, e.Type, e.Err)
}

// This error occurs when the required variable is not set.
type VarIsNotSetError struct {
	Key string
}

func newVarIsNotSetError(key string) error {
	return VarIsNotSetError{key}
}

func (e VarIsNotSetError) Error() string {
	return fmt.Sprintf(`required environment variable %q is not set`, e.Key)
}
