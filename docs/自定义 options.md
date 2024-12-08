`env` 这个库提供的 `Parse()` 使用的是默认参数

如果你需要自定义参数的话，`env` 这个库提供了一个 `ParseWithOptions()`，可以传入自定义的 `options`

`Options` 的参数有：

- `Environment`：传入环境变量的 `map`，用来代替 `os.Environment()`
- `TagName`：用来替代 `env`，`json tag` 默认是 `env`
- `PrefixTagName`: 用来替代 `envPrefix`
- `DefaultValueTagName`：用来替代 `envDefault`
- `RequiredIfNoDef`：如果未声明 `envDefault`，则将所有 `env` 字段设置为必填
- `OnSet`：允许在解析过程中插入钩子，并在设置值时执行某些操作
- `Prefix`：被用于环境变量的前面
- `UseFieldNameByDefault`：当 `env` 字段缺失是，是否应默认使用字段名称
- `FuncMap`：自定义类型转换函数

如果传入自定义 `options`，`ParseWithOptions()` 函数需要完成 `customOptions` 和 `defaultOptions` 的合并

如果有传入的属性，需要使用传入的属性，如果没有传入的属性，就用默认的属性

`options` 合并交给了 `mergeOptions` 函数

```go
func customOptions(opts Options) Options {
    defOpts := defaultOptions()
    mergeOptions[Options](&defOpts, &opts)
    return defOpts
}
```

## mergeOptions

`mergeOptions` 函数的作用是用来合并 `defaultOptions` 和 `customOptions`

接收两个参数 `target`、`source`，其中 `target` 是 `defaultOptions`，`source` 是 `customOptions`

把 `defaultOptions` 作为 `target` 是因为 `defaultOptions` 有所有的属性，如果 `source` 中某些属性有值，只需要把 `target` 中对应的值给更新了

主要的逻辑是：

1. 如果 `targetField` 可以设置，并且 `sourceField` 不是零值，就把 `sourceField` 的值更新到 `targetField`
2. 如果是 `map` 类型的字段，我们应该是合并 `map`，而不是替换 `map`
   1. 遍历 `sourceFiled` 的 `map`，将 `sourceFiled` 的每一项设置到 `targetField`

```go
func mergeOptions(target, source *Option) {
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
             if !sourceField.IsZero() {
                iter := sourceField.MapRange()
                for iter.Next() {
                   targetField.SetMapIndex(iter.Key(), iter.Value())
                }
             }
          default:
             targetField.Set(sourceField)
          }
       }
    }
}
```

#### 零值判断

零值判断分为基本类型和引用类型

引用类型判断是不是 `nil`，基本类型用 `reflect.Zero()` 判断

```go
func isZero(v reflect.Value) bool {
    switch v.Kind() {
    case reflect.Func, reflect.Map, reflect.Slice:
       return v.IsNil()
    default:
       zero := reflect.Zero(v.Type())
       return v.Interface() == zero.Interface()
    }
}
```

## Environment 和 TagName

`Environment` 的作用是可以自己传入环境变量，会覆盖 `os.Environment()`

`tagName` 默认是 `env`，可以指定自己想要的 `tag`

```go
func TestSetenvAndTagOptsChain(t *testing.T) {
    type config struct {
       Key1 string `mytag:"KEY1,required"`
       Key2 int    `mytag:"KEY2,required"`
    }
    envs := map[string]string{
       "KEY1": "VALUE1",
       "KEY2": "3",
    }

    cfg := config{}
    isNoErr(t, ParseWithOptions(&cfg, Options{TagName: "mytag", Environment: envs}))
    isEqual(t, "VALUE1", cfg.Key1)
    isEqual(t, 3, cfg.Key2)
}
```

## RequiredIfNoDef

`RequiredIfNoDef` 默认是 `false`，如果设置为 `true`，那么设置了 `env tag` 的字段是必传

```go
func TestRequiredIfNoDefOption(t *testing.T) {
    type Tree struct {
       Fruit string `env:"FRUIT"`
    }
    type config struct {
       Name  string `env:"NAME"`
       Genre string `env:"GENRE" envDefault:"Unknown"`
       Tree
    }
    var cfg config

    t.Run("missing", func(t *testing.T) {
       err := ParseWithOptions(&cfg, Options{RequiredIfNoDef: true})
       isErrorWithMessage(t, err, `env: required environment variable "NAME" is not set; required environment variable "FRUIT" is not set`)
       isTrue(t, errors.Is(err, VarIsNotSetError{}))
       t.Setenv("NAME", "John")
       err = ParseWithOptions(&cfg, Options{RequiredIfNoDef: true})
       isErrorWithMessage(t, err, `env: required environment variable "FRUIT" is not set`)
       isTrue(t, errors.Is(err, VarIsNotSetError{}))
    })

    t.Run("all set", func(t *testing.T) {
       t.Setenv("NAME", "John")
       t.Setenv("FRUIT", "Apple")

       // should not trigger an error for the missing 'GENRE' env because it has a default value.
       isNoErr(t, ParseWithOptions(&cfg, Options{RequiredIfNoDef: true}))
    })
}
```

## Prefix

`prefix` 指定的名字将用于环境变量的前面

```go
func TestComplePrefix(t *testing.T) {
    type Config struct {
       Home string `env:"HOME"`
    }
    type ComplexConfig struct {
       Foo   Config `envPrefix:"FOO_"`
       Clean Config
       Bar   Config `envPrefix:"BAR_"`
       Blah  string `env:"BLAH"`
    }
    cfg := ComplexConfig{}
    isNoErr(t, ParseWithOptions(&cfg, Options{
       Prefix: "T_",
       Environment: map[string]string{
          "T_FOO_HOME": "/foo",
          "T_BAR_HOME": "/bar",
          "T_BLAH":     "blahhh",
          "T_HOME":     "/clean",
       },
    }))
    isEqual(t, "/foo", cfg.Foo.Home)
    isEqual(t, "/bar", cfg.Bar.Home)
    isEqual(t, "/clean", cfg.Clean.Home)
    isEqual(t, "blahhh", cfg.Blah)
}
```

## FuncMap

`FuncMap` 的作用是用于转换自定义类型

```go
func TestParseCustomMapType(t *testing.T) {
    type custommap map[string]bool

    type config struct {
       SecretKey custommap `env:"SECRET_KEY"`
    }

    t.Setenv("SECRET_KEY", "somesecretkey:1")

    var cfg config
    isNoErr(t, ParseWithOptions(&cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
       reflect.TypeOf(custommap{}): func(_ string) (interface{}, error) {
          return custommap(map[string]bool{}), nil
       },
    }}))
}

func TestParseMapCustomKeyType(t *testing.T) {
    type CustomKey string

    type config struct {
       SecretKey map[CustomKey]bool `env:"SECRET"`
    }

    t.Setenv("SECRET", "somesecretkey:1")

    var cfg config
    isNoErr(t, ParseWithOptions(&cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
       reflect.TypeOf(CustomKey("")): func(value string) (interface{}, error) {
          return CustomKey(value), nil
       },
    }}))
}
```

## DefaultValueTagName

`DefaultValueTagName`：如果指定了 `DefaultValueTagName` 默认值将会从 `DefaultValueTagName` 指定的 `json tag` 中取，如果没有指定从 `envDefault` 中取

```go
func TestParseWithOptionsRenamedDefault(t *testing.T) {
    type config struct {
       Str string `env:"STR" envDefault:"foo" myDefault:"bar"`
    }

    cfg := &config{}
    isNoErr(t, ParseWithOptions(cfg, Options{DefaultValueTagName: "myDefault"}))
    isEqual(t, "bar", cfg.Str)

    isNoErr(t, Parse(cfg))
    isEqual(t, "foo", cfg.Str)
}
```

## PrefixTagName

`PrefixTagName`：如果指定了 `PrefixTagName` 用来替代 `envPrefix`

```go
func TestParseWithOptionsRenamedPrefix(t *testing.T) {
    type Config struct {
       Str string `env:"STR"`
    }
    type ComplexConfig struct {
       Foo Config `envPrefix:"FOO_" myPrefix:"BAR_"`
    }

    t.Setenv("FOO_STR", "101")
    t.Setenv("BAR_STR", "202")
    t.Setenv("APP_BAR_STR", "303")

    cfg := &ComplexConfig{}
    isNoErr(t, ParseWithOptions(cfg, Options{PrefixTagName: "myPrefix"}))
    isEqual(t, "202", cfg.Foo.Str)

    isNoErr(t, ParseWithOptions(cfg, Options{PrefixTagName: "myPrefix", Prefix: "APP_"}))
    isEqual(t, "303", cfg.Foo.Str)

    isNoErr(t, Parse(cfg))
    isEqual(t, "101", cfg.Foo.Str)
}
```

## UseFieldNameByDefault

`UseFieldNameByDefault` 如果没有指定 `env`，则默认使用字段名

```go
func TestNoEnvKey(t *testing.T) {
    type Config struct {
       Foo      string
       FooBar   string
       HTTPPort int
       bar      string
    }
    var cfg Config
    isNoErr(t, ParseWithOptions(&cfg, Options{
       UseFieldNameByDefault: true,
       Environment: map[string]string{
          "FOO":       "fooval",
          "FOO_BAR":   "foobarval",
          "HTTP_PORT": "10",
       },
    }))
    isEqual(t, "fooval", cfg.Foo)
    isEqual(t, "foobarval", cfg.FooBar)
    isEqual(t, 10, cfg.HTTPPort)
    isEqual(t, "", cfg.bar)
}
```

#### 组合字段名

将字段名拼接成 `HTTP_PORT` 这样的格式

1. `unicode.IsUpper(c)` 检查当前字符是不是大写
2. `rune(input[i+1])` 取出当前字符的下一个，`rune(input[i-1])` 取出当前字符的上一个
3. 如果当前字符是大写，并且前一个或者后一个字符是小写，那么应该用 `_` 连接

```go
const underscore rune = '_'

func toEnvName(input string) string {
    var output []rune
    for i, c := range input {
       if c == underscore {
          continue
       }
       //
       if len(output) > 0 && unicode.IsUpper(c) {
          if len(input) > i+1 {
             peek := rune(input[i+1])
             if unicode.IsLower(peek) || unicode.IsLower(rune(input[i-1])) {
                output = append(output, underscore)
             }
          }
       }
       output = append(output, unicode.ToUpper(c))
    }
    return string(output)
}
```

## OnSet

允许在解析过程中插入钩子，并在设置值时执行某些操作

```go
func TestHook(t *testing.T) {
    type config struct {
       Something string `env:"SOMETHING" envDefault:"important"`
       Another   string `env:"ANOTHER"`
       Nope      string
       Inner     struct{} `envPrefix:"FOO_"`
    }

    cfg := &config{}
    t.Setenv("ANOTHER", "1")

    type onSetArgs struct {
       tag       string
       key       interface{}
       isDefault bool
    }

    var onSetCalled []onSetArgs

    isNoErr(t, ParseWithOptions(cfg, Options{
       OnSet: func(tag string, value interface{}, isDefault bool) {
          onSetCalled = append(onSetCalled, onSetArgs{tag, value, isDefault})
       },
    }))
    isEqual(t, "important", cfg.Something)
    isEqual(t, "1", cfg.Another)
    isEqual(t, 2, len(onSetCalled))
    isEqual(t, onSetArgs{"SOMETHING", "important", true}, onSetCalled[0])
    isEqual(t, onSetArgs{"ANOTHER", "1", false}, onSetCalled[1])
}
```

## 源码：

[options](https://github.com/astak16/env/blob/2f04b5a5f127cabbc2d92594fe7e7caab5cc5919/env_type.go#L11)
