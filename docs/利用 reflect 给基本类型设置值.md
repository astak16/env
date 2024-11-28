在学习 `go` 反射时，发现了一个用于解析 `env` 的库： `github.com/caarlos0/env/v11`，其内部是用反射实现的

它使用是比较简单的，调用 `env.Parse` 就能拿到你想要的环境变量了

```go
type config struct {
  Home string `env:"HOME"`
}

// parse
var cfg config
err := env.Parse(&cfg)
```

学习这个库的源码，如果直接上手看的话，很难知道他里面每个参数的作用是啥

比较好的一种学习方法是找到这个库的入口函数，然后根据测试用来一步步调试源码，就能够明白作者为什么要这么写

这个库的入口函数是 `Parse`，我们找到它测试用例 `TestParsesEnv`

## TestParsesEnv

这个测试用例定义了一个 `Config` 结构体，这个结构体包含了 `go` 的大部分类型 (有些类型先暂时忽略)

比如：

- `string`
- `bool`
- `int`
- `int8`
- `int16`
- `int32`
- `int64`
- `uint`
- `uint8`
- `uint16`
- `uint32`
- `uint64`
- `float32`
- `float64`
- `time.Duration`
- `time.Location`
- `url.URL`

以及这些类型的指针和切片，比如： `*string` ，`[]string` ，`[]*string`

通过 `json tag` 去环境变量中读取相应的 `key`，将读取到的值设置到对应的字段中，并且类型要正确

```go
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

    NotAnEnv   string
    unexported string `env:"FOO"`
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
    location2, _ := time.LoadLocation("Europe/Berlin")

    t.Setenv("LOCATION", tos(location1))
    t.Setenv("LOCATIONS", toss(location1, location2))

    url1 := "https://goreleaser.com"
    url2 := "https://caarlos0.dev"
    t.Setenv("URL", tos(url1))
    t.Setenv("URLS", toss(url1, url2))

    t.Setenv("SEPSTRINGS", strings.Join([]string{str1, str2}, ":"))

    nonDefinedStr := "nonDefinedStr"
    t.Setenv("NONDEFINED_STR", nonDefinedStr)
    t.Setenv("PRF_NONDEFINED_STR", nonDefinedStr)

    cfg := Config{}
    _ = Parse(&cfg)

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

    isEqual(t, cfg.NotAnEnv, "")
    isEqual(t, cfg.unexported, "")
}
```

## 类型转换函数

环境变量中的值本质上是字符串，就需要将字符串转换成相应的类型，比如 `"true" -> true`

我们看下 `defaultTypeParsers` 方法中的对于 `url.URL`、`time.Duration`、`time.Location` 的 `key` 使用

为什么 `url.URL` 和 `time.Location` 都是结构体，而 `time.Duration` 对应的 `key` 使用的是 `time.Nanosecond`

1. `reflect.TypeOf` 接收的是值，不是类型， `time.Duration` 是类型，不是值和结构体
2. `time.Nanosecond` 是 `time.Duration` 的最小单位，当然用 `time.Microsecond`、`time.Millisecond` 等，其他值也可以

```go
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
    // 系统默认就是 64 伟
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
```

## Parse

`Parse()` 方法是对外提供的 `api`，接收任意类型的参数，其内部调用 `parseInternal` ，这个方法才是具体的实现

```go
func Parse(v interface{}) error {
    return parseInternal(v, setField, defaultOptions())
}
```

`parseInternal()` 函数内部主要完成的工作有两个：

1. 解析用户传入的结构体的字段的类型
2. 从环境变量中获取到用户需要的值
3. 将环境变量中的值转换成用户需要的类型

### 解析结构体

通过反射获取结构体的类型

- `reflect.ValueOf(c)`：拿到结构体中的值
- `reflect.ValueOf(c).Kind()`：得到结构体的类型
- `reflect.ValueOf(c).Elem()`：结构体是指针的话，获取指针中的内容

```go
type Config struct {}

func main(){
	c := &Config{}
	ptrRef := reflect.ValueOf(c)  // 获取结构体的值
	// Kind() 是获取类型
	fmt.Printf("ptrRef Value %v; Type: %v\n", ptrRef, ptrRef.Kind())  // Kind 得到结构体的类型
	// Elem() 是获取指针中的内容
	ref := ptrRef.Elem() // 如果是指针，通过 Elem 方法拿到内容
	fmt.Printf("ref Value: %v, Type: %v\n", ref, ref.Kind())
}
```

获取结构体中的字段和字段类型详细信息

- `ref.Type()`：拿到结构体的类型信息
- `ref.Type().NumField()`：拿到结构体中字段的数量
- `ref.Field(i)`：拿到结构体中的某一个字段
- `ref.Type().Field(i)`：拿到结构体中某一个字段的详细信息

```go
refType := ref.Type()  // 拿到结构体的类型信息

for i := 0; i < refType.NumField(); i++ {
    refField := ref.Field(i)
    refTypeField := refType.Field(i)
	fmt.Printf("refField %T, refField %+v\n", refField, refField)
	fmt.Printf("refTypeField %T, refTypeField %+v\n", refTypeField, refTypeField)
}
```

`field.Tag` 可以读到 `json tag`，这里 `TagName` 默认是 `env`

```go
field.Tag.Get(opts.TagName)
```

### 读取环境变量中的值

`os.Environ()` 这个 `api` 可以读取所有的环境变量

读取出来的环境变量通过 `toMap()` 处理，最终输出的结果是个 `map` 的数据类型

```go
func toMap(env []string) map[string]string {
    r := map[string]string{}
    for _, e := range env {
       p := strings.SplitN(e, "=", 2)
       if len(p) == 2 {
          r[p[0]] = p[1]
       }
    }
    return r
}
```

然后通过 `json tag env` 指定的类型去找到环境变量中对应的值就可以了

### 转换类型并设置到对应的属性上

从上面 `Config` 结构体中取出下面这四种类型，分别讲解怎么设置到对应的属性上

```go
type Config {
	String     string    `env:"STRING"`
	StringPtr  *string   `env:"STRING"`
	Strings    []string  `env:"STRINGS"`
	StringPtrs []*string `env:"STRINGS"`
}
```

首先类型转换比较好做，使用上面定义的 `defaultBuiltInParsers` 这个 `map` 对象就可以进行对应的类型转换

`defaultBuiltInParsers` 这个 `map` 的 `key` 是 `reflect.Kind` 类型，那么我们可以通过 `reflect.ValueOf(&Config{}).Elem().Field(0).Type()` 获得字段的

```go
reflect.ValueOf(&Config{}).Elem().Field(0).Type() // 可以获取 Config 结构体中第一个字段 String 的类型 string
```

然后调用对应的函数就可以将环境变量中取到的字符串转成结构体定义的类型了

#### `string` 类型设置

基本类型是比较好设置值的

```go
reflect.ValueOf(c).Elem().Field(0).Set(reflect.ValueOf("我是基本类型字符串"))
```

#### \*string 类型设置

指针类型设置值需要先初始化一个地址

```go
field := reflect.ValueOf(c).Elem().Field(0)

field.Set(reflect.New(field.Type().Elem()))
field.Elem().Set(reflect.ValueOf("我是指针类型字符串"))
```

#### []string 类型设置

切片类型的 `string`

```go
ref := reflect.ValueOf(c).Elem()
result := reflect.MakeSlice(ref.Type().Field(0).Type, 0, 1)
result = reflect.Append(result, reflect.ValueOf("我是 string 切片类型"))
result = reflect.Append(result, reflect.ValueOf("我是 string 切片类型2"))
ref.Field(0).Set(result)
fmt.Println(c)
```

#### []\*string 类型设置

切片类型 `*string` 类型设置

```go
ref := reflect.ValueOf(c).Elem()
sf := ref.Type().Field(0)
result := reflect.MakeSlice(sf.Type, 0, 2)
if sf.Type.Elem().Kind() == reflect.Ptr {
    v1 := reflect.New(sf.Type.Elem().Elem())
    v1.Elem().Set(reflect.ValueOf("我是 string 切片类型"))
    result = reflect.Append(result, v1)

    v2 := reflect.New(sf.Type.Elem().Elem())
    v2.Elem().Set(reflect.ValueOf("我是 string 切片类型2"))
    result = reflect.Append(result, v2)
}
ref.Field(0).Set(result)
fmt.Println(*c.StringPtrs[0])
fmt.Println(*c.StringPtrs[1])
```

这里详细看下 `sf.Type.Elem().Elem()` 的意思，下面是一步步打印出来的结果

```go
fmt.Printf("%+v\n", sf)  // {Name:StringPtrs PkgPath: Type:[]*string Tag: Offset:0 Index:[0] Anonymous:false}
fmt.Printf("%+v\n", sf.Type)  // []*string
fmt.Printf("%+v\n", sf.Type.Elem())  // *string
fmt.Printf("%+v\n", sf.Type.Elem().Elem())  // string
```

所以 `v1` 是个 `reflect.Ptr`，在设置值时需要调用 `Elem()` 方法

## 总结

1. 通过反射可以给结构体中的字段设置值
2. 指针类型的字段，在取值时要使用 `Elem()` 方法
3. 指针类型的字段，在设置值时要初始化一个地址

源码：[`Parse`](https://github.com/astak16/env/blob/b08bd76fa5099cfb7545821f058794b52e19e70c/env.go#L8)
