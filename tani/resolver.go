// File: tani/resolver.go
package tani

import "strings"

// 内部マップ
var (
	codeToName map[string]string
	nameToCode map[string]string
)

// SetMaps は LoadTANIFile の結果を設定し、逆引きマップも構築します。
func SetMaps(m map[string]string) {
	codeToName = make(map[string]string, len(m))
	nameToCode = make(map[string]string, len(m))
	for code, name := range m {
		codeToName[code] = name
		nameToCode[name] = code
	}
}

// ResolveName は単位コード→名称を返します。
// マップに存在しなければそのままコードを返します。
func ResolveName(code string) string {
	code = strings.TrimSpace(code)
	if n, ok := codeToName[code]; ok {
		return n
	}
	return code
}

// ResolveCode は単位名称→コードを返します。
// マップに存在しなければ空文字を返します。
func ResolveCode(name string) string {
	name = strings.TrimSpace(name)
	if c, ok := nameToCode[name]; ok {
		return c
	}
	return ""
}
