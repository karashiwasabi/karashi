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

// LoadTANIFile は ShiftJIS で保存された単位マスターCSVを読み込み、
// code→名称のマップを返します。
func LoadTANIFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("LoadTANIFile: open %s: %w", path, err)
	}
	defer file.Close()

	// ShiftJIS を UTF-8 に変換しつつ読み込む
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
	return m, nil
}

// ResolveName は与えられたコードの名称を返します。
// マップに存在しない場合はコード自身を返します。
func ResolveName(code string) string {
	if internalMap == nil {
		return code
	}
	if name, ok := internalMap[code]; ok {
		return name
	}
	return code
}
