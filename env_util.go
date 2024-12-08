package env

import (
	"os"
	"reflect"
	"strings"
	"unicode"
)

func parseKeyForOption(key string) (string, []string) {
	opts := strings.Split(key, ",")
	return opts[0], opts[1:]
}

func isInvalidPtr(v reflect.Value) bool {
	return reflect.Ptr == v.Kind() && v.Elem().Kind() == reflect.Invalid
}

func getFromFile(filename string) (value string, err error) {
	b, err := os.ReadFile(filename)
	return string(b), err
}

const underscore rune = '_'

func toEnvName(input string) string {
	var output []rune
	for i, c := range input {
		if c == underscore {
			continue
		}
		// 当前字符是不是大写
		if len(output) > 0 && unicode.IsUpper(c) {
			// i+1 表示当前索引的下一个索引
			if len(input) > i+1 {
				// 当前字符的的下一个字符
				peek := rune(input[i+1])
				// 当前字符的下一个字符是不是小写 或者 当前字符的上一个字符是不是小写
				if unicode.IsLower(peek) || unicode.IsLower(rune(input[i-1])) {
					// 如果是，那么就在他们之间添加 "_"
					output = append(output, underscore)
				}
			}
		}
		// 否则直接连接起来
		output = append(output, unicode.ToUpper(c))
	}
	return string(output)
}
