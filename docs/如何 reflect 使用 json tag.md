## 默认值设置

`env` 这个库中内置 `env` 这个 `json tag`，去环境变量中去读 `env` 指定的 `key`

如果环境变量中没有这个 `key` 需要一个默认值，怎么办呢？

它提供了 `envDefault` 这个 `json tag`

```go
type Config struct {
    StringWithDefault string   `env:"DATABASE_URL" envDefault:"postgres://localhost:5432/db"`
}
```

通过 `filed.Tag.Lookup("envDefault")` 就可以读到 `envDefault` 这个 `json tag` 指定的内容了

```go
field := refType.Field(i)
defaultValue, hasDefaultValue := field.Tag.Lookup(opts.DefaultValueTagName)
```

当然 `Field.Tag.Get()` 也可以

## 字符串切割

如果某个环境变量有多个值，比如 `STRINGS=str1,str2` 解析出来是个后应该要转成切片的形式：`[str, str2]`

`env` 这个库默认是用 `","` 做分割的

如果想要自己指定分割符，它提供了一个 `envSeparator` 的 `json tag`

```go
type Config struct {
    CustomSeparator   []string `env:"SEPSTRINGS" envSeparator:":"`
}
```

通过 `field.Tag.Get("envSeparator")` 获取分隔符

```go
separator := field.Tag.Get("envSeparator")
if separator == "" {
    separator = ","
}
parts := strings.Split("str1:str2", separator)
```

这样就可以自定义分割符了

## 字段小写开头，默认忽略

```go
type Config struct {
    unexported string
    Exported   string
}
```

判断结构体中的字段是不是未导出的，有两种方法

1. `fieldType.CanSet()`，如果这个字段不能不是设置，说明这个字段是未导出的字段

```go
c := &Config{}

ptrRef := reflect.ValueOf(c)
ref := ptrRef.Elem()
refType := ref.Type()

for i := 0; i < ref.NumField(); i++ {
    field := refType.Field(i)
    fieldType := ref.Field(i)
    if !fieldType.CanSet() {
    fmt.Printf("私有字段: %s\n", field.Name)
    } else {
    fmt.Printf("公开字段: %s\n", field.Name)
    }
}
```

2. `field.PkgPath != ""`，如果 `PkgPath` 不能 `""` 字符串，说明是未导出的字段

```go
c := &Config{}

ptrRef := reflect.ValueOf(c)
ref := ptrRef.Elem()
refType := ref.Type()

for i := 0; i < ref.NumField(); i++ {
    field := refType.Field(i)
    if field.PkgPath != "" {
    fmt.Printf("私有字段: %s\n", field.Name)
    } else {
    fmt.Printf("公开字段: %s\n", field.Name)
    }
}
```

## 解析嵌套结构体

如果结构体中的字段是个结构体类型，该如何处理呢？

```go
type Config struct {
    NonDefined struct {
        String string `env:"NONDEFINED_STR"`
    }
}
```

通过 `Kind()` 函数可以判断当前的字段的类型，如果是结构体的话，就递归调用 `doParse()` ，继续解析结构体

```go
func doParseField(refField reflect.Value, refTypeField reflect.StructField, processField processFieldFn, opts Options) error {
    if refField.Kind() == reflect.Struct {
        return doParse(refField, processField, opts)
    }
}
```

## 添加字段前缀

如果是嵌套结构体，那么我们希望的是每层有自己的 `key`，最终的 `key` 应该是每一层的名字拼接

比如下面的结构体，我们希望的环境变量的 `key` 是 `PRF_NONDEFINED_STR`

```go
type Config struct {
    NestedNonDefined struct {
        NonDefined struct {
        String string `env:"STR"`
        } `env:"NONDEFINED_"`
    } `env:"PRF_"`
}
```

`env` 这个库提供了 `envPrefix` 的 `json tag`

```go
type Config struct {
    NestedNonDefined struct {
        NonDefined struct {
        String string `env:"STR"`
        } `envPrefix:"NONDEFINED_"`
    } `envPrefix:"PRF_"`
}
```

在上面解析嵌套结构体时，我们传给 `doParse` 是默认的 `options`，在这里就需要合并每一层的 `envPrefix`

```go
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
```

在解析结构体时调用 `optionsWithEnvPrefix()` 合并上一层的 `options`

```go
func doParseField(refField reflect.Value, refTypeField reflect.StructField, processField processFieldFn, opts Options) error {
    if refField.Kind() == reflect.Struct {
        return doParse(refField, processField, optionsWithEnvPrefix(refTypeField, opts))
    }
}
```

## time.Location

`time.Location` 是用于表示特定时区的类型，它是一个结构体

```go
type Config struct {
    Location     time.Location    `env:"LOCATION"`
    Locations    []time.Location  `env:"LOCATIONS"`
    LocationPtr  *time.Location   `env:"LOCATION"`
    LocationPtrs []*time.Location `env:"LOCATIONS"`
}
```

在解析 `time.Location` 时，`refField.Kind() == reflect.Struct` 这个判断应该放在 `processField()` 函数下面

```go
func doParseField(refField reflect.Value, refTypeField reflect.StructField, processField processFieldFn, opts Options) error {
    if !refField.CanSet() {
    return nil
    }
    //if refField.Kind() == reflect.Struct {
    // return doParse(refField, processField, optionsWithEnvPrefix(refTypeField, opts))
    //}
    params, err := parseFieldParams(refTypeField, opts)
    if err != nil {
    return err
    }

    if err := processField(refField, refTypeField, opts, params); err != nil {
    return err
    }
    // 这段代码不能放在上面
    if refField.Kind() == reflect.Struct {
    return doParse(refField, processField, optionsWithEnvPrefix(refTypeField, opts))
    }

    return nil
}
```

因为 `processField()` 函数会对当前的字段解析并设置值，如果这段代码放在上面的话，`processsField()` 函数就不会执行，也就不会解析当前字段了

在解析 `time.Location` 类型解析是

```go
location1 := time.UTC
t.Setenv("LOCATION", fmt.Sprintf("%v", location1)) //  fmt.Sprintf("%v", location1) => UTC
location, err := time.LoadLocation(v)
```

源码：[`Parse`](https://github.com/astak16/env/blob/96b714beaa49c12b58e42414cc94cf01de0a6455/env.go#L8)
