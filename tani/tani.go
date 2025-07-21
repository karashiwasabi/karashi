// File: tani/tani.go
package tani

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// internalMap にコード→名称を保持します
var internalMap map[string]string

// ★ 新規追加: 名称→コードの逆引きマップ
var reverseMap map[string]string

// LoadTANIFile は ShiftJIS で保存された単位マスターCSVを読み込みます。
func LoadTANIFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("LoadTANIFile: open %s: %w", path, err)
	}
	defer file.Close()

	decoder := japanese.ShiftJIS.NewDecoder()
	reader := csv.NewReader(transform.NewReader(file, decoder))
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	m := make(map[string]string)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("LoadTANIFile: read %s: %w", path, err)
		}
		if len(record) < 2 {
			continue
		}
		code := record[0]
		name := record[1]
		m[code] = name
	}
	internalMap = m

	// ★ 新規追加: 読み込み時に逆引きマップも生成する
	reverseMap = make(map[string]string)
	for code, name := range internalMap {
		reverseMap[name] = code
	}

	return m, nil
}

// ResolveName は与えられたコードの名称を返します。
func ResolveName(code string) string {
	if internalMap == nil {
		return code
	}
	if name, ok := internalMap[code]; ok {
		return name
	}
	return code
}

// ★ 新規追加: 単位名をコードに変換する関数
func ResolveCode(name string) string {
	if reverseMap == nil {
		return "" // マップがなければ空文字を返す
	}
	if code, ok := reverseMap[name]; ok {
		return code
	}
	return "" // 見つからなければ空文字を返す
}
