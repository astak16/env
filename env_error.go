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

type EmptyVarError struct {
	Key string
}

func newEmptyVarError(key string) error {
	return EmptyVarError{key}
}

func (e EmptyVarError) Error() string {
	return fmt.Sprintf("environment variable %q should not be empty", e.Key)
}

type LoadFileContentError struct {
	Filename string
	Key      string
	Err      error
}

func newLoadFileContentError(filename, key string, err error) error {
	return LoadFileContentError{filename, key, err}
}

func (e LoadFileContentError) Error() string {
	return fmt.Sprintf("could not load content of file %q from variable %s: %v", e.Filename, e.Key, e.Err)
}

type NotStructPtrError struct{}

func (e NotStructPtrError) Error() string {
	return "expected a pointer to a Struct"
}

type NoSupportedTagOptionError struct {
	Tag string
}

func newNoSupportedTagOptionError(tag string) error {
	return NoSupportedTagOptionError{tag}
}

func (e NoSupportedTagOptionError) Error() string {
	return fmt.Sprintf("tag option %q not supported", e.Tag)
}

type NoParserError struct {
	Name string
	Type reflect.Type
}

func newNoParserError(sf reflect.StructField) error {
	return NoParserError{sf.Name, sf.Type}
}

func (e NoParserError) Error() string {
	return fmt.Sprintf("no parser found for field %q of type %q", e.Name, e.Type)
}
