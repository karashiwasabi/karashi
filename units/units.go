// File: units/units.go (修正版)
package units

import (
	"encoding/csv"
	"fmt"
	"io"
	"karashi/model"
	"os"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var internalMap map[string]string
var reverseMap map[string]string

// ★★★ ここからが新しく追加・修正された部分 ★★★

// resolveInnerUnitはJA007のコードから内包装の単位を解決するヘルパー関数です。
// 0や空の場合は空文字を返します。
func resolveInnerUnit(unitCode string) string {
	if unitCode != "0" && unitCode != "" {
		return ResolveName(unitCode)
	}
	return ""
}

// FormatPackageSpecは、JCSHMSのデータから仕様通りの包装文字列を生成します。
func FormatPackageSpec(jcshms *model.JCShms) string {
	if jcshms == nil {
		return ""
	}

	yjUnitName := ResolveName(jcshms.JC039)
	pkg := fmt.Sprintf("%s %g%s", jcshms.JC037, jcshms.JC044, yjUnitName)

	if jcshms.JA006.Valid && jcshms.JA008.Valid && jcshms.JA008.Float64 != 0 {
		// 新しいヘルパー関数を呼び出して単位を取得
		innerUnit := resolveInnerUnit(jcshms.JA007.String)

		pkg += fmt.Sprintf(" (%g%s×%g%s)",
			jcshms.JA006.Float64,
			yjUnitName,
			jcshms.JA008.Float64,
			innerUnit,
		)
	}
	return pkg
}

// ★★★ ここまで ★★★

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

	reverseMap = make(map[string]string)
	for code, name := range internalMap {
		reverseMap[name] = code
	}

	return m, nil
}

func ResolveName(code string) string {
	if internalMap == nil {
		return code
	}
	if name, ok := internalMap[code]; ok {
		return name
	}
	return code
}

func ResolveCode(name string) string {
	if reverseMap == nil {
		return ""
	}
	if code, ok := reverseMap[name]; ok {
		return code
	}
	return ""
}
