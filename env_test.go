package env

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Config struct {
	String     string    `env:"STRING"`
	StringPtr  *string   `env:"STRING"`
	Strings    []string  `env:"STRINGS"`
	StringPtrs []*string `env:"STRINGS"`

	Bool     bool    `env:"BOOL"`
	BoolPtr  *bool   `env:"BOOL"`
	Bools    []bool  `env:"BOOLS"`
	BoolPtrs []*bool `env:"BOOLS"`

	Int     int    `env:"INT"`
	IntPtr  *int   `env:"INT"`
	Ints    []int  `env:"INTS"`
	IntPtrs []*int `env:"INTS"`

	Int8     int8    `env:"INT8"`
	Int8Ptr  *int8   `env:"INT8"`
	Int8s    []int8  `env:"INT8S"`
	Int8Ptrs []*int8 `env:"INT8S"`

	Int16     int16    `env:"INT16"`
	Int16Ptr  *int16   `env:"INT16"`
	Int16s    []int16  `env:"INT16S"`
	Int16Ptrs []*int16 `env:"INT16S"`

	Int32     int32    `env:"INT32"`
	Int32Ptr  *int32   `env:"INT32"`
	Int32s    []int32  `env:"INT32S"`
	Int32Ptrs []*int32 `env:"INT32S"`

	Int64     int64    `env:"INT64"`
	Int64Ptr  *int64   `env:"INT64"`
	Int64s    []int64  `env:"INT64S"`
	Int64Ptrs []*int64 `env:"INT64S"`

	Uint     uint    `env:"UINT"`
	UintPtr  *uint   `env:"UINT"`
	Uints    []uint  `env:"UINTS"`
	UintPtrs []*uint `env:"UINTS"`

	Uint8     uint8    `env:"UINT8"`
	Uint8Ptr  *uint8   `env:"UINT8"`
	Uint8s    []uint8  `env:"UINT8S"`
	Uint8Ptrs []*uint8 `env:"UINT8S"`

	Uint16     uint16    `env:"UINT16"`
	Uint16Ptr  *uint16   `env:"UINT16"`
	Uint16s    []uint16  `env:"UINT16S"`
	Uint16Ptrs []*uint16 `env:"UINT16S"`

	Uint32     uint32    `env:"UINT32"`
	Uint32Ptr  *uint32   `env:"UINT32"`
	Uint32s    []uint32  `env:"UINT32S"`
	Uint32Ptrs []*uint32 `env:"UINT32S"`

	Uint64     uint64    `env:"UINT64"`
	Uint64Ptr  *uint64   `env:"UINT64"`
	Uint64s    []uint64  `env:"UINT64S"`
	Uint64Ptrs []*uint64 `env:"UINT64S"`

	Float32     float32    `env:"FLOAT32"`
	Float32Ptr  *float32   `env:"FLOAT32"`
	Float32s    []float32  `env:"FLOAT32S"`
	Float32Ptrs []*float32 `env:"FLOAT32S"`

	Float64     float64    `env:"FLOAT64"`
	Float64Ptr  *float64   `env:"FLOAT64"`
	Float64s    []float64  `env:"FLOAT64S"`
	Float64Ptrs []*float64 `env:"FLOAT64S"`

	Duration     time.Duration    `env:"DURATION"`
	Durations    []time.Duration  `env:"DURATIONS"`
	DurationPtr  *time.Duration   `env:"DURATION"`
	DurationPtrs []*time.Duration `env:"DURATIONS"`

	Location     time.Location    `env:"LOCATION"`
	Locations    []time.Location  `env:"LOCATIONS"`
	LocationPtr  *time.Location   `env:"LOCATION"`
	LocationPtrs []*time.Location `env:"LOCATIONS"`

	URL     url.URL    `env:"URL"`
	URLPtr  *url.URL   `env:"URL"`
	URLs    []url.URL  `env:"URLS"`
	URLPtrs []*url.URL `env:"URLS"`

	StringWithDefault string `env:"DATABASE_URL" envDefault:"postgres://localhost:5432/db"`

	CustomSeparator []string `env:"SEPSTRINGS" envSeparator:":"`

	NonDefined struct {
		String string `env:"NONDEFINED_STR"`
	}

	NestedNonDefined struct {
		NonDefined struct {
			String string `env:"STR"`
		} `envPrefix:"NONDEFINED_"`
	} `envPrefix:"PRF_"`

	NotAnEnv   string
	unexported string `env:"FOO"`
}

type ParentStruct struct {
	InnerStruct    *InnerStruct `env:",init"`
	NilInnerStruct *InnerStruct
	unexported     *InnerStruct
	Ignored        *http.Client
}

type InnerStruct struct {
	Inner  string `env:"innervar"`
	Number uint   `env:"innernum"`
}

func TestParsesEnv(t *testing.T) {
	tos := func(v interface{}) string {
		return fmt.Sprintf("%v", v)
	}

	toss := func(v ...interface{}) string {
		ss := []string{}
		for _, s := range v {
			ss = append(ss, tos(s))
		}
		return strings.Join(ss, ",")
	}

	str1 := "str1"
	str2 := "str2"
	t.Setenv("STRING", str1)
	t.Setenv("STRINGS", toss(str1, str2))

	bool1 := true
	bool2 := false
	t.Setenv("BOOL", tos(bool1))
	t.Setenv("BOOLS", toss(bool1, bool2))

	int1 := -1
	int2 := 2
	t.Setenv("INT", tos(int1))
	t.Setenv("INTS", toss(int1, int2))

	var int81 int8 = -2
	var int82 int8 = 5
	t.Setenv("INT8", tos(int81))
	t.Setenv("INT8S", toss(int81, int82))

	var int161 int16 = -24
	var int162 int16 = 15
	t.Setenv("INT16", tos(int161))
	t.Setenv("INT16S", toss(int161, int162))

	var int321 int32 = -14
	var int322 int32 = 154
	t.Setenv("INT32", tos(int321))
	t.Setenv("INT32S", toss(int321, int322))

	var int641 int64 = -12
	var int642 int64 = 150
	t.Setenv("INT64", tos(int641))
	t.Setenv("INT64S", toss(int641, int642))

	var uint1 uint = 1
	var uint2 uint = 2
	t.Setenv("UINT", tos(uint1))
	t.Setenv("UINTS", toss(uint1, uint2))

	var uint81 uint8 = 15
	var uint82 uint8 = 51
	t.Setenv("UINT8", tos(uint81))
	t.Setenv("UINT8S", toss(uint81, uint82))

	var uint161 uint16 = 532
	var uint162 uint16 = 123
	t.Setenv("UINT16", tos(uint161))
	t.Setenv("UINT16S", toss(uint161, uint162))

	var uint321 uint32 = 93
	var uint322 uint32 = 14
	t.Setenv("UINT32", tos(uint321))
	t.Setenv("UINT32S", toss(uint321, uint322))

	var uint641 uint64 = 5
	var uint642 uint64 = 43
	t.Setenv("UINT64", tos(uint641))
	t.Setenv("UINT64S", toss(uint641, uint642))

	var float321 float32 = 9.3
	var float322 float32 = 1.1
	t.Setenv("FLOAT32", tos(float321))
	t.Setenv("FLOAT32S", toss(float321, float322))

	float641 := 1.53
	float642 := 0.5
	t.Setenv("FLOAT64", tos(float641))
	t.Setenv("FLOAT64S", toss(float641, float642))

	duration1 := time.Second
	duration2 := time.Second * 4
	t.Setenv("DURATION", tos(duration1))
	t.Setenv("DURATIONS", toss(duration1, duration2))

	location1 := time.UTC
	location2, errLoadLocation := time.LoadLocation("Europe/Berlin")
	isNoErr(t, errLoadLocation)
	t.Setenv("LOCATION", tos(location1))
	t.Setenv("LOCATIONS", toss(location1, location2))

	url1 := "https://goreleaser.com"
	url2 := "https://caarlos0.dev"
	t.Setenv("URL", tos(url1))
	t.Setenv("URLS", toss(url1, url2))

	t.Setenv("SEPSTRINGS", str1+":"+str2)

	nonDefinedStr := "nonDefinedStr"
	t.Setenv("NONDEFINED_STR", nonDefinedStr)
	t.Setenv("PRF_NONDEFINED_STR", nonDefinedStr)

	t.Setenv("FOO", str1)

	cfg := Config{}
	isNoErr(t, Parse(&cfg))

	isEqual(t, str1, cfg.String)
	isEqual(t, &str1, cfg.StringPtr)
	isEqual(t, str1, cfg.Strings[0])
	isEqual(t, str2, cfg.Strings[1])
	isEqual(t, &str1, cfg.StringPtrs[0])
	isEqual(t, &str2, cfg.StringPtrs[1])

	isEqual(t, bool1, cfg.Bool)
	isEqual(t, &bool1, cfg.BoolPtr)
	isEqual(t, bool1, cfg.Bools[0])
	isEqual(t, bool2, cfg.Bools[1])
	isEqual(t, &bool1, cfg.BoolPtrs[0])
	isEqual(t, &bool2, cfg.BoolPtrs[1])

	isEqual(t, int1, cfg.Int)
	isEqual(t, &int1, cfg.IntPtr)
	isEqual(t, int1, cfg.Ints[0])
	isEqual(t, int2, cfg.Ints[1])
	isEqual(t, &int1, cfg.IntPtrs[0])
	isEqual(t, &int2, cfg.IntPtrs[1])

	isEqual(t, int81, cfg.Int8)
	isEqual(t, &int81, cfg.Int8Ptr)
	isEqual(t, int81, cfg.Int8s[0])
	isEqual(t, int82, cfg.Int8s[1])
	isEqual(t, &int81, cfg.Int8Ptrs[0])
	isEqual(t, &int82, cfg.Int8Ptrs[1])

	isEqual(t, int161, cfg.Int16)
	isEqual(t, &int161, cfg.Int16Ptr)
	isEqual(t, int161, cfg.Int16s[0])
	isEqual(t, int162, cfg.Int16s[1])
	isEqual(t, &int161, cfg.Int16Ptrs[0])
	isEqual(t, &int162, cfg.Int16Ptrs[1])

	isEqual(t, int321, cfg.Int32)
	isEqual(t, &int321, cfg.Int32Ptr)
	isEqual(t, int321, cfg.Int32s[0])
	isEqual(t, int322, cfg.Int32s[1])
	isEqual(t, &int321, cfg.Int32Ptrs[0])
	isEqual(t, &int322, cfg.Int32Ptrs[1])

	isEqual(t, int641, cfg.Int64)
	isEqual(t, &int641, cfg.Int64Ptr)
	isEqual(t, int641, cfg.Int64s[0])
	isEqual(t, int642, cfg.Int64s[1])
	isEqual(t, &int641, cfg.Int64Ptrs[0])
	isEqual(t, &int642, cfg.Int64Ptrs[1])

	isEqual(t, uint1, cfg.Uint)
	isEqual(t, &uint1, cfg.UintPtr)
	isEqual(t, uint1, cfg.Uints[0])
	isEqual(t, uint2, cfg.Uints[1])
	isEqual(t, &uint1, cfg.UintPtrs[0])
	isEqual(t, &uint2, cfg.UintPtrs[1])

	isEqual(t, uint81, cfg.Uint8)
	isEqual(t, &uint81, cfg.Uint8Ptr)
	isEqual(t, uint81, cfg.Uint8s[0])
	isEqual(t, uint82, cfg.Uint8s[1])
	isEqual(t, &uint81, cfg.Uint8Ptrs[0])
	isEqual(t, &uint82, cfg.Uint8Ptrs[1])

	isEqual(t, uint161, cfg.Uint16)
	isEqual(t, &uint161, cfg.Uint16Ptr)
	isEqual(t, uint161, cfg.Uint16s[0])
	isEqual(t, uint162, cfg.Uint16s[1])
	isEqual(t, &uint161, cfg.Uint16Ptrs[0])
	isEqual(t, &uint162, cfg.Uint16Ptrs[1])

	isEqual(t, uint321, cfg.Uint32)
	isEqual(t, &uint321, cfg.Uint32Ptr)
	isEqual(t, uint321, cfg.Uint32s[0])
	isEqual(t, uint322, cfg.Uint32s[1])
	isEqual(t, &uint321, cfg.Uint32Ptrs[0])
	isEqual(t, &uint322, cfg.Uint32Ptrs[1])

	isEqual(t, uint641, cfg.Uint64)
	isEqual(t, &uint641, cfg.Uint64Ptr)
	isEqual(t, uint641, cfg.Uint64s[0])
	isEqual(t, uint642, cfg.Uint64s[1])
	isEqual(t, &uint641, cfg.Uint64Ptrs[0])
	isEqual(t, &uint642, cfg.Uint64Ptrs[1])

	isEqual(t, float321, cfg.Float32)
	isEqual(t, &float321, cfg.Float32Ptr)
	isEqual(t, float321, cfg.Float32s[0])
	isEqual(t, float322, cfg.Float32s[1])
	isEqual(t, &float321, cfg.Float32Ptrs[0])

	isEqual(t, float641, cfg.Float64)
	isEqual(t, &float641, cfg.Float64Ptr)
	isEqual(t, float641, cfg.Float64s[0])
	isEqual(t, float642, cfg.Float64s[1])
	isEqual(t, &float641, cfg.Float64Ptrs[0])
	isEqual(t, &float642, cfg.Float64Ptrs[1])

	isEqual(t, duration1, cfg.Duration)
	isEqual(t, &duration1, cfg.DurationPtr)
	isEqual(t, duration1, cfg.Durations[0])
	isEqual(t, duration2, cfg.Durations[1])
	isEqual(t, &duration1, cfg.DurationPtrs[0])
	isEqual(t, &duration2, cfg.DurationPtrs[1])

	isEqual(t, *location1, cfg.Location)
	isEqual(t, location1, cfg.LocationPtr)
	isEqual(t, *location1, cfg.Locations[0])
	isEqual(t, *location2, cfg.Locations[1])
	isEqual(t, location1, cfg.LocationPtrs[0])
	isEqual(t, location2, cfg.LocationPtrs[1])

	isEqual(t, url1, cfg.URL.String())
	isEqual(t, url1, cfg.URLPtr.String())
	isEqual(t, url1, cfg.URLs[0].String())
	isEqual(t, url2, cfg.URLs[1].String())
	isEqual(t, url1, cfg.URLPtrs[0].String())
	isEqual(t, url2, cfg.URLPtrs[1].String())

	isEqual(t, "postgres://localhost:5432/db", cfg.StringWithDefault)
	isEqual(t, nonDefinedStr, cfg.NonDefined.String)
	isEqual(t, nonDefinedStr, cfg.NestedNonDefined.NonDefined.String)

	isEqual(t, str1, cfg.CustomSeparator[0])
	isEqual(t, str2, cfg.CustomSeparator[1])

	isEqual(t, cfg.NotAnEnv, "")

	isEqual(t, cfg.unexported, "")
}

func TestInvalidBool(t *testing.T) {
	t.Setenv("BOOL", "should-be-a-bool")
	err := Parse(&Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Bool" of type "bool": strconv.ParseBool: parsing "should-be-a-bool": invalid syntax; parse error on field "BoolPtr" of type "*bool": strconv.ParseBool: parsing "should-be-a-bool": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidInt(t *testing.T) {
	t.Setenv("INT", "should-be-an-int")
	err := Parse(&Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Int" of type "int": strconv.ParseInt: parsing "should-be-an-int": invalid syntax; parse error on field "IntPtr" of type "*int": strconv.ParseInt: parsing "should-be-an-int": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidUint(t *testing.T) {
	t.Setenv("UINT", "-44")
	err := Parse(&Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Uint" of type "uint": strconv.ParseUint: parsing "-44": invalid syntax; parse error on field "UintPtr" of type "*uint": strconv.ParseUint: parsing "-44": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidFloat32(t *testing.T) {
	t.Setenv("FLOAT32", "AAA")
	err := Parse(&Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Float32" of type "float32": strconv.ParseFloat: parsing "AAA": invalid syntax; parse error on field "Float32Ptr" of type "*float32": strconv.ParseFloat: parsing "AAA": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidFloat64(t *testing.T) {
	t.Setenv("FLOAT64", "AAA")
	err := Parse(&Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Float64" of type "float64": strconv.ParseFloat: parsing "AAA": invalid syntax; parse error on field "Float64Ptr" of type "*float64": strconv.ParseFloat: parsing "AAA": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidUint64(t *testing.T) {
	t.Setenv("UINT64", "AAA")
	err := Parse(&Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Uint64" of type "uint64": strconv.ParseUint: parsing "AAA": invalid syntax; parse error on field "Uint64Ptr" of type "*uint64": strconv.ParseUint: parsing "AAA": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidInt64(t *testing.T) {
	t.Setenv("INT64", "AAA")
	err := Parse(&Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Int64" of type "int64": strconv.ParseInt: parsing "AAA": invalid syntax; parse error on field "Int64Ptr" of type "*int64": strconv.ParseInt: parsing "AAA": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidInt64Slice(t *testing.T) {
	t.Setenv("BADINTS", "A,2,3")
	type config struct {
		BadFloats []int64 `env:"BADINTS"`
	}
	err := Parse(&config{})
	isErrorWithMessage(t, err, `env: parse error on field "BadFloats" of type "[]int64": strconv.ParseInt: parsing "A": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidUInt64Slice(t *testing.T) {
	t.Setenv("BADINTS", "A,2,3")
	type config struct {
		BadFloats []uint64 `env:"BADINTS"`
	}
	err := Parse(&config{})
	isErrorWithMessage(t, err, `env: parse error on field "BadFloats" of type "[]uint64": strconv.ParseUint: parsing "A": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidFloat32Slice(t *testing.T) {
	t.Setenv("BADFLOATS", "A,2.0,3.0")
	type config struct {
		BadFloats []float32 `env:"BADFLOATS"`
	}
	err := Parse(&config{})
	isErrorWithMessage(t, err, `env: parse error on field "BadFloats" of type "[]float32": strconv.ParseFloat: parsing "A": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidFloat64Slice(t *testing.T) {
	t.Setenv("BADFLOATS", "A,2.0,3.0")
	type config struct {
		BadFloats []float64 `env:"BADFLOATS"`
	}
	err := Parse(&config{})
	isErrorWithMessage(t, err, `env: parse error on field "BadFloats" of type "[]float64": strconv.ParseFloat: parsing "A": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidBoolsSlice(t *testing.T) {
	t.Setenv("BADBOOLS", "t,f,TRUE,faaaalse")
	type config struct {
		BadBools []bool `env:"BADBOOLS"`
	}
	err := Parse(&config{})
	isErrorWithMessage(t, err, `env: parse error on field "BadBools" of type "[]bool": strconv.ParseBool: parsing "faaaalse": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidDuration(t *testing.T) {
	t.Setenv("DURATION", "should-be-a-valid-duration")
	err := Parse(&Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Duration" of type "time.Duration": unable to parse duration: time: invalid duration "should-be-a-valid-duration"; parse error on field "DurationPtr" of type "*time.Duration": unable to parse duration: time: invalid duration "should-be-a-valid-duration"`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidDurations(t *testing.T) {
	t.Setenv("DURATIONS", "1s,contains-an-invalid-duration,3s")
	err := Parse(&Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Durations" of type "[]time.Duration": unable to parse duration: time: invalid duration "contains-an-invalid-duration"; parse error on field "DurationPtrs" of type "[]*time.Duration": unable to parse duration: time: invalid duration "contains-an-invalid-duration"`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidLocation(t *testing.T) {
	t.Setenv("LOCATION", "should-be-a-valid-location")
	err := Parse(&Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Location" of type "time.Location": unable to parse location: unknown time zone should-be-a-valid-location; parse error on field "LocationPtr" of type "*time.Location": unable to parse location: unknown time zone should-be-a-valid-location`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidLocations(t *testing.T) {
	t.Setenv("LOCATIONS", "should-be-a-valid-location,UTC,Europe/Berlin")
	err := Parse(&Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Locations" of type "[]time.Location": unable to parse location: unknown time zone should-be-a-valid-location; parse error on field "LocationPtrs" of type "[]*time.Location": unable to parse location: unknown time zone should-be-a-valid-location`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestParseStructWithoutEnvTag(t *testing.T) {
	cfg := Config{}
	isNoErr(t, Parse(&cfg))
	isEqual(t, cfg.NotAnEnv, "")
}

func TestParseStructWithInvalidFieldKind(t *testing.T) {
	type config struct {
		WontWorkByte byte `env:"BLAH"`
	}
	t.Setenv("BLAH", "a")
	err := Parse(&config{})
	isErrorWithMessage(t, err, `env: parse error on field "WontWorkByte" of type "uint8": strconv.ParseUint: parsing "a": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestParsesEnv_Map(t *testing.T) {
	type config struct {
		MapStringString                map[string]string `env:"MAP_STRING_STRING" envSeparator:","`
		MapStringInt64                 map[string]int64  `env:"MAP_STRING_INT64"`
		MapStringBool                  map[string]bool   `env:"MAP_STRING_BOOL" envSeparator:";"`
		CustomSeparatorMapStringString map[string]string `env:"CUSTOM_SEPARATOR_MAP_STRING_STRING" envSeparator:"," envKeyValSeparator:"|"`
	}

	mss := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}
	t.Setenv("MAP_STRING_STRING", "k1:v1,k2:v2")

	msi := map[string]int64{
		"k1": 1,
		"k2": 2,
	}
	t.Setenv("MAP_STRING_INT64", "k1:1,k2:2")

	msb := map[string]bool{
		"k1": true,
		"k2": false,
	}
	t.Setenv("MAP_STRING_BOOL", "k1:true;k2:false")

	withCustomSeparator := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}
	t.Setenv("CUSTOM_SEPARATOR_MAP_STRING_STRING", "k1|v1,k2|v2")

	var cfg config
	isNoErr(t, Parse(&cfg))

	isEqual(t, mss, cfg.MapStringString)
	isEqual(t, msi, cfg.MapStringInt64)
	isEqual(t, msb, cfg.MapStringBool)
	isEqual(t, withCustomSeparator, cfg.CustomSeparatorMapStringString)
}

func TestParsesEnvInvalidMap(t *testing.T) {
	type config struct {
		MapStringString map[string]string `env:"MAP_STRING_STRING" envSeparator:","`
	}

	t.Setenv("MAP_STRING_STRING", "k1,k2:v2")

	var cfg config
	err := Parse(&cfg)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestErrorRequiredNotSet(t *testing.T) {
	type config struct {
		IsRequired string `env:"IS_REQUIRED,required"`
	}
	err := Parse(&config{})
	isErrorWithMessage(t, err, `env: required environment variable "IS_REQUIRED" is not set`)
	isTrue(t, errors.Is(err, VarIsNotSetError{}))
}

func TestNoErrorRequiredSet(t *testing.T) {
	type config struct {
		IsRequired string `env:"IS_REQUIRED,required"`
	}

	cfg := &config{}

	t.Setenv("IS_REQUIRED", "")
	isNoErr(t, Parse(cfg))
	isEqual(t, "", cfg.IsRequired)
}

func TestNoErrorRequiredAndNotEmptySet(t *testing.T) {
	t.Setenv("IS_REQUIRED", "1")
	type config struct {
		IsRequired string `env:"IS_REQUIRED,required,notEmpty"`
	}
	isNoErr(t, Parse(&config{}))
}

func TestNoErrorNotEmptySet(t *testing.T) {
	t.Setenv("IS_REQUIRED", "1")
	type config struct {
		IsRequired string `env:"IS_REQUIRED,notEmpty"`
	}
	isNoErr(t, Parse(&config{}))
}

func TestErrorNotEmptySet(t *testing.T) {
	t.Setenv("IS_REQUIRED", "")
	type config struct {
		IsRequired string `env:"IS_REQUIRED,notEmpty"`
	}
	err := Parse(&config{})
	isErrorWithMessage(t, err, `env: environment variable "IS_REQUIRED" should not be empty`)
	isTrue(t, errors.Is(err, EmptyVarError{}))
}

func TestErrorRequiredAndNotEmptySet(t *testing.T) {
	t.Setenv("IS_REQUIRED", "")
	type config struct {
		IsRequired string `env:"IS_REQUIRED,required,notEmpty"`
	}
	err := Parse(&config{})
	isErrorWithMessage(t, err, `env: environment variable "IS_REQUIRED" should not be empty`)
	isTrue(t, errors.Is(err, EmptyVarError{}))
}

func TestParsesEnvInnerFailsMultipleErrors(t *testing.T) {
	type config struct {
		Foo struct {
			Name   string `env:"NAME,required"`
			Number int    `env:"NUMBER"`
			Bar    struct {
				Age int `env:"AGE,required"`
			}
		}
	}
	t.Setenv("NUMBER", "not-a-number")
	c := &config{}
	err := Parse(c)
	isErrorWithMessage(t, err, `env: required environment variable "NAME" is not set; parse error on field "Number" of type "int": strconv.ParseInt: parsing "not-a-number": invalid syntax; required environment variable "AGE" is not set`)
	isTrue(t, errors.Is(err, ParseError{}))
	isTrue(t, errors.Is(err, VarIsNotSetError{}))
}

func TestParsesEnvInner_WhenInnerStructPointerIsNil(t *testing.T) {
	t.Setenv("innervar", "someinnervalue")
	t.Setenv("innernum", "8")
	cfg := ParentStruct{}
	isNoErr(t, Parse(&cfg))
	isEqual(t, "someinnervalue", cfg.InnerStruct.Inner)
	isEqual(t, uint(8), cfg.InnerStruct.Number)
}

func TestParsesEnvInner(t *testing.T) {
	type InnerStruct struct {
		Inner  string `env:"innervar"`
		Number uint   `env:"innernum"`
	}
	type ParentStruct struct {
		InnerStruct *InnerStruct `env:",init"`
	}

	t.Setenv("innervar", "someinnervalue")
	t.Setenv("innernum", "8")
	cfg := ParentStruct{
		InnerStruct: &InnerStruct{},
	}
	isNoErr(t, Parse(&cfg))
	isEqual(t, "someinnervalue", cfg.InnerStruct.Inner)
	isEqual(t, uint(8), cfg.InnerStruct.Number)
}

func TestParseExpandOption(t *testing.T) {
	type config struct {
		Host        string `env:"HOST" envDefault:"localhost"`
		Port        int    `env:"PORT,expand" envDefault:"3000"`
		SecretKey   string `env:"SECRET_KEY,expand"`
		ExpandKey   string `env:"EXPAND_KEY"`
		CompoundKey string `env:"HOST_PORT,expand" envDefault:"${HOST}:${PORT}"`
		Default     string `env:"DEFAULT,expand" envDefault:"def1"`
	}

	t.Setenv("HOST", "localhost")
	t.Setenv("PORT", "3000")
	t.Setenv("EXPAND_KEY", "qwerty12345")
	t.Setenv("SECRET_KEY", "${EXPAND_KEY}")

	cfg := config{}
	err := Parse(&cfg)

	isNoErr(t, err)
	isEqual(t, "localhost", cfg.Host)
	isEqual(t, 3000, cfg.Port)
	isEqual(t, "qwerty12345", cfg.SecretKey)
	isEqual(t, "qwerty12345", cfg.ExpandKey)
	isEqual(t, "localhost:3000", cfg.CompoundKey)
	isEqual(t, "def1", cfg.Default)
}

func TestParseExpandWithDefaultOption(t *testing.T) {
	type config struct {
		Host            string `env:"HOST" envDefault:"localhost"`
		Port            int    `env:"PORT,expand" envDefault:"3000"`
		OtherPort       int    `env:"OTHER_PORT" envDefault:"4000"`
		CompoundDefault string `env:"HOST_PORT,expand" envDefault:"${HOST}:${PORT}"`
		SimpleDefault   string `env:"DEFAULT,expand" envDefault:"def1"`
		MixedDefault    string `env:"MIXED_DEFAULT,expand" envDefault:"$USER@${HOST}:${OTHER_PORT}"`
		OverrideDefault string `env:"OVERRIDE_DEFAULT,expand"`
		DefaultIsExpand string `env:"DEFAULT_IS_EXPAND,expand" envDefault:"$THIS_IS_EXPAND"`
		NoDefault       string `env:"NO_DEFAULT,expand"`
	}

	t.Setenv("OTHER_PORT", "5000")
	t.Setenv("USER", "jhon")
	t.Setenv("THIS_IS_USED", "this is used instead")
	t.Setenv("OVERRIDE_DEFAULT", "msg: ${THIS_IS_USED}")
	t.Setenv("THIS_IS_EXPAND", "msg: ${THIS_IS_USED}")
	t.Setenv("NO_DEFAULT", "$PORT:$OTHER_PORT")

	cfg := config{}
	err := Parse(&cfg)

	isNoErr(t, err)
	isEqual(t, "localhost", cfg.Host)
	isEqual(t, 3000, cfg.Port)
	isEqual(t, 5000, cfg.OtherPort)
	isEqual(t, "localhost:3000", cfg.CompoundDefault)
	isEqual(t, "def1", cfg.SimpleDefault)
	isEqual(t, "jhon@localhost:5000", cfg.MixedDefault)
	isEqual(t, "msg: this is used instead", cfg.OverrideDefault)
	isEqual(t, "3000:5000", cfg.NoDefault)
}

func TestParseUnsetRequireOptions(t *testing.T) {
	type config struct {
		Password string `env:"PASSWORD,unset,required"`
	}
	cfg := config{}

	err := Parse(&cfg)
	isErrorWithMessage(t, err, `env: required environment variable "PASSWORD" is not set`)
	isTrue(t, errors.Is(err, VarIsNotSetError{}))
	t.Setenv("PASSWORD", "superSecret")
	isNoErr(t, Parse(&cfg))

	isEqual(t, "superSecret", cfg.Password)
	unset, exists := os.LookupEnv("PASSWORD")
	isEqual(t, "", unset)
	isEqual(t, false, exists)
}

func TestFile(t *testing.T) {
	type config struct {
		SecretKey string `env:"SECRET_KEY,file"`
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "sec_key")
	isNoErr(t, os.WriteFile(file, []byte("secret"), 0o660))

	t.Setenv("SECRET_KEY", file)

	cfg := config{}
	isNoErr(t, Parse(&cfg))
	isEqual(t, "secret", cfg.SecretKey)
}

func TestFileNoParam(t *testing.T) {
	type config struct {
		SecretKey string `env:"SECRET_KEY,file"`
	}

	cfg := config{}
	err := Parse(&cfg)
	isNoErr(t, err)
}

func TestFileNoParamRequired(t *testing.T) {
	type config struct {
		SecretKey string `env:"SECRET_KEY,file,required"`
	}

	err := Parse(&config{})
	isErrorWithMessage(t, err, `env: required environment variable "SECRET_KEY" is not set`)
	isTrue(t, errors.Is(err, VarIsNotSetError{}))
}

func TestFileBadFile(t *testing.T) {
	type config struct {
		SecretKey string `env:"SECRET_KEY,file"`
	}

	filename := "not-a-real-file"
	t.Setenv("SECRET_KEY", filename)

	oserr := "no such file or directory"
	if runtime.GOOS == "windows" {
		oserr = "The system cannot find the file specified."
	}

	err := Parse(&config{})
	isErrorWithMessage(t, err, fmt.Sprintf("env: could not load content of file %q from variable SECRET_KEY: open %s: %s", filename, filename, oserr))
	isTrue(t, errors.Is(err, LoadFileContentError{}))
}

func TestFileWithDefault(t *testing.T) {
	type config struct {
		SecretKey string `env:"SECRET_KEY,file,expand" envDefault:"${FILE}"`
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "sec_key")
	isNoErr(t, os.WriteFile(file, []byte("secret"), 0o660))

	t.Setenv("FILE", file)

	cfg := config{}
	isNoErr(t, Parse(&cfg))
	isEqual(t, "secret", cfg.SecretKey)
}

func TestParsesEnvInnerNil(t *testing.T) {
	t.Setenv("innervar", "someinnervalue")
	cfg := ParentStruct{}
	isNoErr(t, Parse(&cfg))
}

func TestParsesEnvInnerInvalid(t *testing.T) {
	t.Setenv("innernum", "-547")
	cfg := ParentStruct{
		InnerStruct: &InnerStruct{},
	}
	err := Parse(&cfg)
	isErrorWithMessage(t, err, `env: parse error on field "Number" of type "uint": strconv.ParseUint: parsing "-547": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestParsesEnvNested(t *testing.T) {
	type NestedStruct struct {
		NestedVar string `env:"nestedvar"`
	}
	type ForNestedStruct struct {
		NestedStruct
	}

	t.Setenv("nestedvar", "somenestedvalue")
	var cfg ForNestedStruct
	isNoErr(t, Parse(&cfg))
	isEqual(t, "somenestedvalue", cfg.NestedVar)
}

func TestEmptyVars(t *testing.T) {
	cfg := Config{}
	isNoErr(t, Parse(&cfg))
	isEqual(t, "", cfg.String)
	isEqual(t, false, cfg.Bool)
	isEqual(t, 0, cfg.Int)
	isEqual(t, uint(0), cfg.Uint)
	isEqual(t, uint64(0), cfg.Uint64)
	isEqual(t, int64(0), cfg.Int64)
	isEqual(t, 0, len(cfg.Strings))
	isEqual(t, 0, len(cfg.CustomSeparator))
	isEqual(t, 0, len(cfg.Ints))
	isEqual(t, 0, len(cfg.Bools))
}

func TestPassAnInvalidPtr(t *testing.T) {
	var thisShouldBreak int
	err := Parse(&thisShouldBreak)
	isErrorWithMessage(t, err, "env: expected a pointer to a Struct")
	isTrue(t, errors.Is(err, NotStructPtrError{}))
}

func TestPassReference(t *testing.T) {
	cfg := Config{}
	err := Parse(cfg)
	isErrorWithMessage(t, err, "env: expected a pointer to a Struct")
	isTrue(t, errors.Is(err, NotStructPtrError{}))
}

func TestEmptyOption(t *testing.T) {
	type config struct {
		Var string `env:"VAR,"`
	}

	cfg := &config{}

	t.Setenv("VAR", "")
	isNoErr(t, Parse(cfg))
	isEqual(t, "", cfg.Var)
}

func TestErrorOptionNotRecognized(t *testing.T) {
	type config struct {
		Var string `env:"VAR,not_supported!"`
	}
	err := Parse(&config{})
	isErrorWithMessage(t, err, `env: tag option "not_supported!" not supported`)
	isTrue(t, errors.Is(err, NoSupportedTagOptionError{}))
}

func TestParseURL(t *testing.T) {
	type config struct {
		ExampleURL url.URL `env:"EXAMPLE_URL" envDefault:"https://google.com"`
	}
	var cfg config
	isNoErr(t, Parse(&cfg))
	isEqual(t, "https://google.com", cfg.ExampleURL.String())
}

func TestParseInvalidURL(t *testing.T) {
	type config struct {
		ExampleURL url.URL `env:"EXAMPLE_URL_2"`
	}
	t.Setenv("EXAMPLE_URL_2", "nope://s s/")

	err := Parse(&config{})
	isErrorWithMessage(t, err, `env: parse error on field "ExampleURL" of type "url.URL": unable to parse URL: parse "nope://s s/": invalid character " " in host name`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestIgnoresUnexported(t *testing.T) {
	type unexportedConfig struct {
		home  string `env:"HOME"`
		Home2 string `env:"HOME"`
	}
	cfg := unexportedConfig{}

	t.Setenv("HOME", "/tmp/fakehome")
	isNoErr(t, Parse(&cfg))
	isEqual(t, cfg.home, "")
	isEqual(t, "/tmp/fakehome", cfg.Home2)
}

func TestErrorIs(t *testing.T) {
	err := newAggregateError(newParseError(reflect.StructField{}, nil))
	t.Run("is", func(t *testing.T) {
		isTrue(t, errors.Is(err, ParseError{}))
	})
	t.Run("is not", func(t *testing.T) {
		isFalse(t, errors.Is(err, NoParserError{}))
	})
}

func TestIssue245(t *testing.T) {
	t.Setenv("NAME_NOT_SET", "")
	type user struct {
		Name string `env:"NAME_NOT_SET" envDefault:"abcd"`
	}
	cfg := user{}
	isNoErr(t, Parse(&cfg))
	isEqual(t, cfg.Name, "abcd")
}

func TestIssue339(t *testing.T) {
	t.Run("Should parse with bool ptr set and env undefined", func(t *testing.T) {
		existingValue := true
		cfg := Config{
			BoolPtr: &existingValue,
		}

		isNoErr(t, Parse(&cfg))

		isEqual(t, &existingValue, cfg.BoolPtr)
	})

	t.Run("Should parse with bool ptr set and env defined", func(t *testing.T) {
		existingValue := true
		cfg := Config{
			BoolPtr: &existingValue,
		}

		newValue := false
		t.Setenv("BOOL", strconv.FormatBool(newValue))

		isNoErr(t, Parse(&cfg))

		isEqual(t, &newValue, cfg.BoolPtr)
	})

	t.Run("Should parse with string ptr set and env undefined", func(t *testing.T) {
		existingValue := "one"
		cfg := Config{
			StringPtr: &existingValue,
		}

		isNoErr(t, Parse(&cfg))

		isEqual(t, &existingValue, cfg.StringPtr)
	})

	t.Run("Should parse with string ptr set and env defined", func(t *testing.T) {
		existingValue := "one"
		cfg := Config{
			StringPtr: &existingValue,
		}

		newValue := "two"
		t.Setenv("STRING", newValue)

		isNoErr(t, Parse(&cfg))

		isEqual(t, &newValue, cfg.StringPtr)
	})
}

func isEqual(tb testing.TB, a, b interface{}) {
	tb.Helper()

	if areEqual(a, b) {
		return
	}

	tb.Fatalf("expected %#v (type %T) == %#v (type %T)", a, a, b, b)
}

func areEqual(a, b interface{}) bool {
	if isNil(a) && isNil(b) {
		return true
	}
	if isNil(a) || isNil(b) {
		return false
	}
	if reflect.DeepEqual(a, b) {
		return true
	}
	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)
	return aValue == bValue
}

// copied from https://github.com/matryer/is
func isNil(object interface{}) bool {
	if object == nil {
		return true
	}
	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}
	return false
}

func isNoErr(tb testing.TB, err error) {
	tb.Helper()

	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}

func isTrue(tb testing.TB, b bool) {
	tb.Helper()

	if !b {
		tb.Fatalf("expected true, got false")
	}
}

func isFalse(tb testing.TB, b bool) {
	tb.Helper()

	if b {
		tb.Fatalf("expected false, got true")
	}
}

func isErrorWithMessage(tb testing.TB, err error, msg string) {
	tb.Helper()

	if err == nil {
		tb.Fatalf("expected error, got nil")
	}

	if msg != err.Error() {
		tb.Fatalf("expected error message %q, got %q", msg, err.Error())
	}
}
