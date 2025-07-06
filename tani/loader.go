// File: tani/loader.go
package tani

import (
	"encoding/csv"
	"os"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// LoadTANIFile は Shift-JIS 形式の TANI.CSV を読み込み
// code→name マップを返します。
func LoadTANIFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(transform.NewReader(f, japanese.ShiftJIS.NewDecoder()))
	r.LazyQuotes = true
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	m := make(map[string]string, len(rows))
	for _, rec := range rows {
		if len(rec) >= 2 {
			code := strings.TrimSpace(rec[0])
			name := strings.TrimSpace(rec[1])
			m[code] = name
		}
	}
	return m, nil
}
