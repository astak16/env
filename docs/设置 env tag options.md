`env` 这个库提供了一些 `tag options`

- `,expand`: 从环境变量中插入进来，例如  `FOO_${BAR}`
- `,file`: 环境变量应该是一个文件的路径
- `,init`: 给指针类型的结构体参数设置 `nil`
- `,notEmpty`：如果环境变量为空会报错
- `,required`：如果环境变量没有设置会报错
- `,unset`: 环境变量在使用会删除掉

## required

`requried` 这个参数使用是方式

第二种使用方式为什么要写一个逗号，是因为 `tag` 第一个参数是 `env` 的 `key`

```go
type Config struct {
    IsRequired string `env:"IS_REQUIRED,required"`
}
// 或者
type config struct {
    IsRequired string `env:",required"`
}
```

在解析 `json tag env` 时，第一个参数是环境变量的 `key` ，后面的参数作为 `env` 的配置参数

```go
func parseKeyForOption(key string) (string, []string) {
    opts := strings.Split(field.Tag.Get("env"), ",")
    return opts[0], opts[1:]
}
```

在参数解析时设置判断有没有 `required` 配置参数

```go
func parseFieldParams(field reflect.StructField, opts Options) (FieldParams, error) {
    for _, tag := range tags {
       switch tag {
       case "":
          continue
       case "required":
          result.Required = true
       }
    }
    return result, nil
}
```

在获取环境变量时，如果传了 `required` 参数，但是在环境变量中没有找到，就需要抛出 `VarIsNotSetError` 的错误

```go
func get(fieldParams FieldParams, opts Options) (val string, err error) {
    if fieldParams.Required && !exists {
       return "", newVarIsNotSetError(fieldParams.Key)
    }
    return value, nil
}
```

下面两个测试用例就是用来测试 `required` 参数的

```go
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
```

## notEmpty

它的配置和 `required` 参不多

```go
func parseFieldParams(field reflect.StructField, opts Options) (FieldParams, error) {
    for _, tag := range tags {
       switch tag {
       case "":
          continue
       case "notEmpty":
          result.NotEmpty = true
       }
    }
    return result, nil
}
```

在获取环境变量时，如果传了 `notEmpty` 参数，但是在环境变量中发现值是空的，就需要抛出 `newEmptyVarError` 的错误

```go
func get(fieldParams FieldParams, opts Options) (val string, err error) {
    if fieldParams.NotEmpty && value == "" {
	    return "", newEmptyVarError(fieldParams.Key)
	}
    return value, nil
}
```

下面四个测试用例就是用来测试 `required` 参数的

```go
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
```

## init

`init` 参数的作用是给指针类型的结构体设置 `nil`，如果不是设置 `nil` 的话，在嵌套结构体中会报错

下面代码中在 `cfg.InnerStruct` 取值 `Inner` 会报错

```go
func TestParsesEnvInner_WhenInnerStructPointerIsNil(t *testing.T) {
    type InnerStruct struct {
       Inner  string `env:"innervar"`
       Number uint   `env:"innernum"`
    }
    type ParentStruct struct {
       InnerStruct *InnerStruct
    }

    t.Setenv("innervar", "someinnervalue")
    t.Setenv("innernum", "8")
    cfg := ParentStruct{}
    isNoErr(t, Parse(&cfg))
    isEqual(t, "someinnervalue", cfg.InnerStruct.Inner)
    isEqual(t, uint(8), cfg.InnerStruct.Number)
}
```

为了解决这个问题，提供了 `init` 这个 `env tag options` 参数

这个参数的作用是给 `InnerStruct` 这个属性进行初始化

```go
func TestParsesEnvInner_WhenInnerStructPointerIsNil(t *testing.T) {
    type InnerStruct struct {
       Inner  string `env:"innervar"`
       Number uint   `env:"innernum"`
    }
    type ParentStruct struct {
       InnerStruct *InnerStruct `env:",init"`
    }

    t.Setenv("innervar", "someinnervalue")
    t.Setenv("innernum", "8")
    cfg := ParentStruct{}
    isNoErr(t, Parse(&cfg))
    isEqual(t, "someinnervalue", cfg.InnerStruct.Inner)
    isEqual(t, uint(8), cfg.InnerStruct.Number)
}
```

因为 `cfg1` 这个结构体没有初始化 `InnerStruct` ，所以会报错，需要用 `init` 参数初始化 `InnerStruct`

`cfg2` 结构体是显示初始化了 `InnerStruct` 结构体

```go
 type ParentStruct struct {
    InnerStruct *InnerStruct
}

cfg1 := &ParentStruct{}
cfg2 := &ParentStruct{
	InnerStruct: &InnerStruct{}
}
```

使用 `init` 初始化结构体地址

```go
if params.Init && isStructPtr(refField) && refField.IsNil() {
    refField.Set(reflect.New(refField.Type().Elem()))
    refField = refField.Elem()
}
```

## expand

`expand` 这个参数是利用 `os.Expand()` 这个 `api` 实现的

原理是将 `${HOST}` 或者 `$HOME` 传递给 `os.Expand()` ，会有一个解析参数将 `HOST` 成想要的值

```go
mapping := func(varName string) string {
    switch varName {
    case "HOST":
       return "localhost"
    case "PORT":
       return "3000"
    default:
       return ""
    }
}
result := os.Expand("${HOST}:${PORT}", mapping)
fmt.Println(result) // localhost:3000
```

如果不是 `${HOST}` 格式的值，会被原封不动的返回回去

```go
result2 := os.Expand("HOST", mapping)
fmt.Println(result2) // HOST
```

所以 `expand` 的具体实现方法

```go
if fieldParams.Expand {
    val = os.Expand(val, opts.getRawEnv)
}
func (opts *Options) getRawEnv(s string) string {
    val := opts.rawEnvVars[s]
    if val == "" {
       val = opts.Environment[s]
    }
    return os.Expand(val, opts.getRawEnv)
}
```

它的测试用例

```go
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
    t.Setenv("HOST", "localhost")
    t.Setenv("PORT", "3000")
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
```

## unset

`unset` 这个参数的作用是将已经使用的环境变量删除

```go
os.Setenv("PASSWORD", "superSecret")
os.Unsetenv("PASSWORD")
unset, exists := os.LookupEnv("PASSWORD")
fmt.Println(unset, exists)  // "", false
```

实现方式

```go
if fieldParams.Unset {
    defer os.Unsetenv(fieldParams.Key)
}
```

测试用例

```go
func TestParseUnsetRequireOptions(t *testing.T) {
    type config struct {
       Password string `env:"PASSWORD,unset,required"`
    }
    cfg := config{}

    err := Parse(&cfg)
    isErrorWithMessage(t, err, `env: required environment variable "PASSWORD" is not set`)
    isTrue(t, errors.Is(err, VarIsNotSetError{}))
    os.Setenv("PASSWORD", "superSecret")
    isNoErr(t, Parse(&cfg))

    isEqual(t, "superSecret", cfg.Password)
    unset, exists := os.LookupEnv("PASSWORD")
    isEqual(t, "", unset)
    isEqual(t, false, exists)
}
```

## file

`file` 参数用来设置环境变量的值是某个文件的路径

```go
dir, _ := os.Getwd()
file := filepath.Join(dir, "sec_key")   // file 是文件路径
isNoErr(t, os.WriteFile(file, []byte("secret"), 0o660))

t.Setenv("SECRET_KEY", file)
```

实现：

```go
if fieldParams.LoadFile && val != "" {
    filename := val
    val, err = getFromFile(filename)
    if err != nil {
       return "", newLoadFileContentError(filename, fieldParams.Key, err)
    }
}
func getFromFile(filename string) (value string, err error) {
    b, err := os.ReadFile(filename)
    return string(b), err
}
```

测试用例

```go
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
```

## 源码：

1. [required](https://github.com/astak16/env/blob/6e8324b86d3faf8ef85ccf1c3e40b295786b07e0/env.go#L96)
2. [notEmpty](https://github.com/astak16/env/blob/7b5735f06cb08d7dbff95168370b0f9bcc188392/env.go#L107)
3. [init](https://github.com/astak16/env/blob/320b3acd34448bca13bdd0f0c842f9511208fb4b/env.go#L62)
4. [expand](https://github.com/astak16/env/blob/dafbae6934148a57b380c5045b9b7ec05efc2704/env.go#L120)
5. [unset](https://github.com/astak16/env/blob/0c6da651051a898ab0b39b1f505f2c31f35684af/env.go#L128)
6. [file](https://github.com/astak16/env/blob/b76caf5661d3be827c91e3c24581911c61cbfe1a/env.go#L141)
